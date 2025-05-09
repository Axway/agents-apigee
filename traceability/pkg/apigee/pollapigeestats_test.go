package apigee

import (
	"encoding/json"
	"os"
	"testing"
	"time"

	"github.com/Axway/agent-sdk/pkg/cache"
	"github.com/Axway/agent-sdk/pkg/transaction"
	"github.com/Axway/agents-apigee/client/pkg/apigee"
	"github.com/Axway/agents-apigee/client/pkg/apigee/models"
	"github.com/stretchr/testify/assert"
)

const testdata = "testdata/"

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
		skipNotSet    bool
	}{
		{
			name:      "Only Success",
			responses: []string{"only_success.json"},
			total:     7,
			successes: 7,
			errors:    0,
			apiCalls: map[string][]int{
				"Petstore": {7, 7, 0},
			},
		},
		{
			name:      "Only Errors",
			responses: []string{"only_errors.json"},
			total:     7,
			successes: 0,
			errors:    7,
			apiCalls: map[string][]int{
				"Petstore": {7, 0, 7},
			},
		},
		{
			name:      "Multiple Calls",
			responses: []string{"multiple_calls_1.json", "multiple_calls_2.json"},
			total:     28,
			successes: 14,
			errors:    14,
			apiCalls: map[string][]int{
				"Petstore": {28, 14, 14},
			},
		},
		{
			name:      "Multiple APIs",
			responses: []string{"multiple_apis.json"},
			total:     45,
			successes: 27,
			errors:    18,
			apiCalls: map[string][]int{
				"Petstore":     {19, 11, 8},
				"Practitioner": {26, 16, 10},
			},
		},
		{
			name:      "Real Data",
			responses: []string{"real_data.json"},
			total:     1788,
			successes: 894,
			errors:    894,
			apiCalls: map[string][]int{
				"Swagger-Petstore": {1788, 894, 894},
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
		{
			name:      "Real Data - Product Mode - No Not Set",
			responses: []string{"real_data_2.json"},
			total:     24,
			successes: 24,
			errors:    0,
			apiCalls: map[string][]int{
				"Test": {24, 24, 0},
			},
			isProductMode: true,
			productsMap: map[string]string{
				"Test-planname": "Test",
			},
			skipNotSet: true,
		},
	}

	for _, test := range testCases {
		t.Run(test.name, func(t *testing.T) {
			eventReport, err := transaction.NewEventReportBuilder().
				SetOnlyTrackMetrics(true).
				Build()
			assert.Nil(t, err)
			opts := []func(*pollApigeeStats){
				withStatsCache(cache.New()),
				withStatsClient(&mockClient{
					statResponses: test.responses,
					envs:          []string{"test"},
					productsMap:   test.productsMap,
				}),
				withAllTraffic(true),
				withNotSetTraffic(!test.skipNotSet),
				withEventReport(eventReport),
			}
			if test.isProductMode {
				opts = append(opts, withProductMode())
			}
			job := newPollStatsJob(opts...)

			// send all metrics through the processor
			for range test.responses {
				err := job.Execute()
				assert.Nil(t, err)
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
			assert.Equal(t, test.cachePath, job.cachePath)
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
