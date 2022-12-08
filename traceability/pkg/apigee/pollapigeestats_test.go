package apigee

import (
	"encoding/json"
	"os"
	"strconv"
	"sync"
	"testing"
	"time"

	"github.com/Axway/agent-sdk/pkg/cache"
	"github.com/Axway/agent-sdk/pkg/transaction/metric"
	"github.com/Axway/agents-apigee/client/pkg/apigee"
	"github.com/Axway/agents-apigee/client/pkg/apigee/models"
	"github.com/stretchr/testify/assert"
)

const testdata = "testdata/"

type mockCollector struct {
	metric.Collector
	apiCounts map[string][]int
	total     int
	successes int
	errors    int
	mutex     *sync.Mutex
}

func newMockCollector() *mockCollector {
	return &mockCollector{
		apiCounts: make(map[string][]int),
		mutex:     &sync.Mutex{},
	}
}

func (m *mockCollector) AddMetricDetail(metricDetail metric.Detail) {
	m.AddMetric(metricDetail.APIDetails, metricDetail.StatusCode, metricDetail.Duration, metricDetail.Bytes, metricDetail.AppDetails.Name)
}

func (m *mockCollector) AddMetric(apiDetails metric.APIDetails, statusCode string, duration, bytes int64, appName string) {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	apiCount := make([]int, 3)
	if c, ok := m.apiCounts[apiDetails.Name]; ok {
		apiCount = c
	}
	m.total++
	apiCount[0]++
	if code, _ := strconv.Atoi(statusCode); code < 400 {
		m.successes++
		apiCount[1]++
	} else {
		m.errors++
		apiCount[2]++
	}
	m.apiCounts[apiDetails.Name] = apiCount
}

type mockClient struct {
	envs          []string
	responseCount int
	statResponses []string
	productsMap   map[string]string
}

func (m *mockClient) GetEnvironments() []string {
	return m.envs
}

func (m *mockClient) GetStats(env, dimension, metricSelect string, start, end time.Time) (*models.Metrics, error) {
	content, _ := os.ReadFile(testdata + m.statResponses[m.responseCount])
	metrics := &models.Metrics{}
	json.Unmarshal(content, metrics)
	m.responseCount++
	return metrics, nil
}

func (m *mockClient) GetProduct(productName string) (*models.ApiProduct, error) {
	if m.productsMap == nil {
		// so empty
	} else if p, ok := m.productsMap[productName]; ok {
		return &models.ApiProduct{
			Attributes: []models.Attribute{
				{
					Name:  "Attribute1",
					Value: "Value1",
				},
				{
					Name:  "Attribute2",
					Value: "Value2",
				},
				{
					Name:  apigee.ClonedProdAttribute,
					Value: p,
				},
			},
		}, nil
	}
	return nil, nil
}

func TestProcessMetric(t *testing.T) {
	testCases := []struct {
		name          string
		responses     []string
		total         int
		successes     int
		errors        int
		apiCalls      map[string][]int
		isProductMode bool
		productsMap   map[string]string
	}{
		{
			name:      "Only Success",
			responses: []string{"only_success.json"},
			total:     7,
			successes: 7,
			errors:    0,
			apiCalls: map[string][]int{
				"Petstore (prod)": {7, 7, 0},
			},
		},
		{
			name:      "Only Errors",
			responses: []string{"only_errors.json"},
			total:     7,
			successes: 0,
			errors:    7,
			apiCalls: map[string][]int{
				"Petstore (prod)": {7, 0, 7},
			},
		},
		{
			name:      "Multiple Calls",
			responses: []string{"multiple_calls_1.json", "multiple_calls_2.json"},
			total:     21,
			successes: 7,
			errors:    14,
			apiCalls: map[string][]int{
				"Petstore (prod)": {21, 7, 14},
			},
		},
		{
			name:      "Multiple APIs",
			responses: []string{"multiple_apis.json"},
			total:     45,
			successes: 27,
			errors:    18,
			apiCalls: map[string][]int{
				"Petstore (prod)":     {19, 11, 8},
				"Practitioner (prod)": {26, 16, 10},
			},
		},
		{
			name:      "Real Data",
			responses: []string{"real_data.json"},
			total:     1788,
			successes: 894,
			errors:    894,
			apiCalls: map[string][]int{
				"Swagger-Petstore (prod)": {1788, 894, 894},
			},
		},
		{
			name:      "Real Data - Product Mode",
			responses: []string{"real_data_2.json"},
			total:     47,
			successes: 47,
			errors:    0,
			apiCalls: map[string][]int{
				"Test": {24, 24, 0},
			},
			isProductMode: true,
			productsMap: map[string]string{
				"Test-planname": "Test",
			},
		},
	}

	for _, test := range testCases {
		t.Run(test.name, func(t *testing.T) {
			opts := []func(*pollApigeeStats){
				withStatsCache(cache.New()),
				withStatsClient(&mockClient{
					statResponses: test.responses,
					envs:          []string{"test"},
					productsMap:   test.productsMap,
				}),
			}
			if test.isProductMode {
				opts = append(opts, withProductMode())
			}
			job := newPollStatsJob(opts...)

			mCollector := newMockCollector()
			job.collector = mCollector

			// send all metrics through the processor
			for range test.responses {
				job.Execute()
			}

			// check the totals
			mCollector.mutex.Lock()
			defer mCollector.mutex.Unlock()
			assert.Equal(t, test.total, mCollector.total)
			assert.Equal(t, test.successes, mCollector.successes)
			assert.Equal(t, test.errors, mCollector.errors)

			// check the counts for each api
			for proxy, expectedCounts := range test.apiCalls {
				assert.Contains(t, mCollector.apiCounts, proxy)
				assert.Equal(t, expectedCounts, mCollector.apiCounts[proxy])
			}
		})
	}
}

func TestCleanCache(t *testing.T) {
	testCases := []struct {
		name        string
		inputs      [][]string
		cleanedKeys []string
	}{
		{
			name: "Create Keys",
			inputs: [][]string{
				{"key3", "key2", "key1"},
			},
			cleanedKeys: []string{},
		},
		{
			name: "Same Keys",
			inputs: [][]string{
				{"key3", "key2", "key1"},
				{"key3", "key2", "key1"},
			},
			cleanedKeys: []string{},
		},
		{
			name: "Clean Keys",
			inputs: [][]string{
				{"key3", "key2", "key1"},
				{"key3", "key2", "key1"},
				{"key4", "key3", "key2"},
				{"key5", "key4", "key3"},
				{"key6", "key5", "key4"},
				{"key6", "key5", "key4"},
				{"key7", "key6", "key5"},
			},
			cleanedKeys: []string{"key4", "key3", "key2", "key1"},
		},
	}

	for _, test := range testCases {
		t.Run(test.name, func(t *testing.T) {
			job := newPollStatsJob(
				withStatsCache(cache.New()),
			)

			expected := test.inputs[len(test.inputs)-1 : len(test.inputs)][0]
			// load the cache with keys from all inputs and send inputs to cleanCache
			for _, in := range test.inputs {
				for _, key := range in {
					job.statCache.Set(key, nil)
				}
				job.cacheKeys = in
				job.cleanCache()
			}

			// check that all expected keys still in cache
			for _, key := range expected {
				_, err := job.statCache.Get(key)
				assert.Nil(t, err)
			}

			// check that all cleaned keys not in cache
			for _, key := range test.cleanedKeys {
				_, err := job.statCache.Get(key)
				assert.NotNil(t, err)
			}
		})
	}
}

func TestNewPollStatsJob(t *testing.T) {
	testCases := []struct {
		name          string
		startTime     time.Time
		increment     time.Duration
		cacheClean    bool
		productMode   bool
		isReady       bool
		cachePath     string
		withStatCache bool
	}{
		{
			name: "No Options",
		},
		{
			name:          "All Options",
			startTime:     time.Now().Add(time.Hour * -1),
			increment:     time.Hour,
			cacheClean:    true,
			productMode:   true,
			isReady:       true,
			cachePath:     "/path/to/cache",
			withStatCache: true,
		},
	}

	for _, test := range testCases {
		t.Run(test.name, func(t *testing.T) {
			opts := make([]func(*pollApigeeStats), 0)
			opts = append(opts, withStatsCache(cache.New()))

			if !test.startTime.IsZero() {
				opts = append(opts, withStartTime(test.startTime))
			}
			if test.increment > 0 {
				opts = append(opts, withIncrement(test.increment))
			}
			if test.cacheClean {
				opts = append(opts, withCacheClean())
			}
			if test.productMode {
				opts = append(opts, withProductMode())
			}
			if test.isReady {
				opts = append(opts, withIsReady(func() bool { return true }))
			}
			if test.cachePath != "" {
				opts = append(opts, withCachePath(test.cachePath))
			}
			if test.withStatCache {
				opts = append(opts, withStatsCache(cache.New()))
			}

			job := newPollStatsJob(opts...)

			assert.NotNil(t, job)
			assert.Equal(t, test.startTime, job.startTime)
			assert.Equal(t, test.increment, job.increment)
			assert.Equal(t, test.cachePath, job.cachePath)
			assert.Equal(t, []string{}, job.cacheKeys)
			assert.NotNil(t, job.mutex)
			if test.cacheClean {
				assert.True(t, job.cacheClean)
			} else {
				assert.False(t, job.cacheClean)
			}
			if test.productMode {
				assert.True(t, job.isProduct)
			} else {
				assert.False(t, job.isProduct)
			}
			if test.isReady {
				assert.NotNil(t, job.ready)
			} else {
				assert.Nil(t, job.ready)
			}
		})
	}
}
