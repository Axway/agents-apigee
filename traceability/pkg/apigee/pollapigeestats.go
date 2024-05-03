package apigee

import (
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/Axway/agent-sdk/pkg/agent"
	management "github.com/Axway/agent-sdk/pkg/apic/apiserver/models/management/v1alpha1"
	sdkDef "github.com/Axway/agent-sdk/pkg/apic/definitions"
	"github.com/Axway/agent-sdk/pkg/cache"
	"github.com/Axway/agent-sdk/pkg/jobs"
	"github.com/Axway/agent-sdk/pkg/transaction/metric"
	metricModels "github.com/Axway/agent-sdk/pkg/transaction/models"
	"github.com/Axway/agent-sdk/pkg/transaction/util"
	sdkUtil "github.com/Axway/agent-sdk/pkg/util"
	"github.com/Axway/agent-sdk/pkg/util/log"
	"github.com/Axway/agents-apigee/client/pkg/apigee"
	"github.com/Axway/agents-apigee/client/pkg/apigee/models"
	"github.com/Axway/agents-apigee/traceability/pkg/apigee/definitions"
	"github.com/gofrs/uuid"
)

const (
	lastStartTimeKey  = "lastStartTime"
	countMetric       = "sum(message_count)"
	policyErrMetric   = "sum(policy_error)"
	serverErrMetric   = "sum(target_error)"
	avgResponseMetric = "avg(total_response_time)"
	maxResponseMetric = "max(total_response_time)"
	minResponseMetric = "min(total_response_time)"
)

type isReady func() bool

type metricData struct {
	environment     string
	baseName        string
	name            string
	timestamp       time.Time
	count           int64
	policyErrCount  int64
	serverErrCount  int64
	avgResponseTime float64
	minResponseTime int64
	maxResponseTime int64
}

type pollApigeeStats struct {
	jobs.Job
	startTime           time.Time
	endTime             time.Time
	envs                []string
	mutex               *sync.Mutex
	cacheClean          bool
	reportAllTraffic    bool
	reportNotSetTraffic bool
	collector           metric.Collector
	ready               isReady
	client              definitions.StatsClient
	statCache           cache.Cache
	cachePath           string
	clonedProduct       map[string]string
	dimension           string
	environment         string
	isProduct           bool
	logger              log.FieldLogger
	filteredAPIs        map[string]struct{}
	filterMetrics       bool
}

func newPollStatsJob(options ...func(*pollApigeeStats)) *pollApigeeStats {
	job := &pollApigeeStats{
		collector:     metric.GetMetricCollector(),
		mutex:         &sync.Mutex{},
		clonedProduct: make(map[string]string),
		dimension:     "apiproxy",
		logger:        log.NewFieldLogger().WithComponent("pollStatsJob").WithPackage("apigee"),
		filteredAPIs:  make(map[string]struct{}),
	}

	for _, o := range options {
		o(job)
	}
	return job
}

func withEnvironment(env string) func(p *pollApigeeStats) {
	return func(p *pollApigeeStats) {
		p.environment = env
	}
}

func withStartTime(startTime time.Time) func(p *pollApigeeStats) {
	return func(p *pollApigeeStats) {
		p.startTime = startTime
	}
}

func withCacheClean() func(p *pollApigeeStats) {
	return func(p *pollApigeeStats) {
		p.cacheClean = true
	}
}

func withStatsClient(client definitions.StatsClient) func(p *pollApigeeStats) {
	return func(p *pollApigeeStats) {
		p.client = client
	}
}

func withStatsCache(cache cache.Cache) func(p *pollApigeeStats) {
	return func(p *pollApigeeStats) {
		p.statCache = cache
	}
}

func withCachePath(path string) func(p *pollApigeeStats) {
	return func(p *pollApigeeStats) {
		p.cachePath = path
	}
}

func withIsReady(ready isReady) func(p *pollApigeeStats) {
	return func(p *pollApigeeStats) {
		p.ready = ready
	}
}

func withAllTraffic(allTraffic bool) func(p *pollApigeeStats) {
	return func(p *pollApigeeStats) {
		p.reportAllTraffic = allTraffic
	}
}

func withNotSetTraffic(notSetTraffic bool) func(p *pollApigeeStats) {
	return func(p *pollApigeeStats) {
		p.reportNotSetTraffic = notSetTraffic
	}
}

func withProductMode() func(p *pollApigeeStats) {
	return func(p *pollApigeeStats) {
		p.dimension = "api_product"
		p.isProduct = true
	}
}

func withFilteredAPIs(filteredAPIs []string, filterMetrics bool) func(p *pollApigeeStats) {
	return func(p *pollApigeeStats) {
		p.filterMetrics = filterMetrics
		if filterMetrics {
			for _, filteredAPI := range filteredAPIs {
				p.filteredAPIs[filteredAPI] = struct{}{}
			}
		}
	}
}

func (j *pollApigeeStats) Ready() bool {
	return j.ready()
}

func (j *pollApigeeStats) Status() error {
	return nil
}

func (j *pollApigeeStats) Execute() error {
	id, _ := uuid.NewV4()
	logger := j.logger.WithField("executionID", id)

	logger.Trace("starting execution")
	j.collector.InitializeBatch()
	j.envs = j.client.GetEnvironments()

	// when start time is 0 we are in our regular execution loop
	j.endTime = time.Now().Add(time.Minute * -10).Truncate(time.Minute)

	metricSelect := strings.Join([]string{countMetric, policyErrMetric, serverErrMetric, avgResponseMetric, minResponseMetric, maxResponseMetric}, ",")
	wg := &sync.WaitGroup{}

	if j.filterMetrics && len(j.filteredAPIs) == 0 {
		apiService := management.NewAPIService("", GetAgent().cfg.CentralCfg.GetEnvironmentName())
		res, err := agent.GetCentralClient().GetResources(apiService)
		if len(res) > 0 && err == nil {
			for _, apiSvc := range res {
				ri, _ := apiSvc.AsInstance()
				apiID, _ := sdkUtil.GetAgentDetailsValue(ri, sdkDef.AttrExternalAPIID)
				j.filteredAPIs[apiID] = struct{}{}
			}
		}
	}
	for _, e := range j.envs {
		if !(j.environment == "" || j.environment == e) {
			continue
		}

		wg.Add(1)
		go func(logger log.FieldLogger, envName string) {
			defer wg.Done()
			logger = logger.WithField("env", envName)
			metrics, err := j.client.GetStats(envName, j.dimension, metricSelect, j.startTime, j.endTime)
			if err != nil {
				return
			}

			j.processMetricResponse(logger, metrics)
		}(logger, e)
	}
	wg.Wait()

	logger.Trace("finished execution")
	if j.cacheClean {
		j.cleanCache()
	}

	// set startTime for the next api call
	j.startTime = j.endTime

	// only update the lastStartTime when it is not zero
	if !j.startTime.IsZero() {
		j.statCache.Set(lastStartTimeKey, j.startTime.String())
		j.statCache.Save(j.cachePath)
	}
	j.collector.Publish()

	return nil
}

func (j *pollApigeeStats) processMetricResponse(logger log.FieldLogger, metrics *models.Metrics) {
	logger.Trace("start processing env")
	if len(metrics.Environments) != 1 {
		logger.Error("exactly 1 environment should be returned")
		return
	}

	if len(metrics.Environments[0].Dimensions) == 0 {
		logger.Trace("At least one proxy is needed to process response data")
		return
	}

	// get the index of each metric
	var metricsIndex = map[string]int{
		countMetric:       -1,
		policyErrMetric:   -1,
		serverErrMetric:   -1,
		avgResponseMetric: -1,
		maxResponseMetric: -1,
		minResponseMetric: -1,
	}

	// initialize the metrics index map for each proxy
	for i, m := range metrics.Environments[0].Dimensions[0].Metrics { // api_proxies or api_product
		if _, found := metricsIndex[m.Name]; !found {
			logger.Tracef("skipping metric, %s, in return data", m.Name)
		}
		metricsIndex[m.Name] = i
	}

	// check for -1 index in metricsIndex
	for key, index := range metricsIndex {
		if key != "" && index < 0 {
			logger.Errorf("did not find the %s metric in the returned data", key)
			return
		}
	}

	dimensions := []string{}
	for _, d := range metrics.Environments[0].Dimensions { // api_proxies
		dimensions = append(dimensions, d.Name)
	}

	logger.WithField("value", dimensions).Trace("dimensions")

	metricGroups := map[string][]*metricData{}
	for _, d := range metrics.Environments[0].Dimensions {
		if j.filterMetrics {
			if _, filteredAPI := j.filteredAPIs[d.Name]; !filteredAPI {
				continue
			}
		}
		serviceName := j.getBaseProduct(d.Name)
		logger := logger.WithField("name", d.Name).WithField("serviceName", serviceName)
		logger.Trace("processing metric for dimension")
		if !j.reportNotSetTraffic && serviceName == "(not set)" {
			continue
		}
		if !j.reportAllTraffic && !agent.IsAPIPublishedByID(serviceName) {
			logger.Trace("skipping as its not discovered")
			continue
		}
		for i := range d.Metrics[0].MetricValues {
			if d.Metrics[metricsIndex[countMetric]].MetricValues[i].Value == "0.0" {
				continue
			}
			key := fmt.Sprint(metrics.Environments[0].Name, "_", serviceName)
			if _, ok := metricGroups[key]; !ok {
				metricGroups[key] = make([]*metricData, 0)
			}

			metricGroups[key] = append(metricGroups[key], &metricData{
				environment:     metrics.Environments[0].Name,
				name:            serviceName,
				baseName:        d.Name,
				timestamp:       time.UnixMilli(d.Metrics[metricsIndex[countMetric]].MetricValues[i].Timestamp),
				count:           parseFloatToInt64(d.Metrics[metricsIndex[countMetric]].MetricValues[i].Value),
				policyErrCount:  parseFloatToInt64(d.Metrics[metricsIndex[policyErrMetric]].MetricValues[i].Value),
				serverErrCount:  parseFloatToInt64(d.Metrics[metricsIndex[serverErrMetric]].MetricValues[i].Value),
				avgResponseTime: parseFloatToFloat64(d.Metrics[metricsIndex[avgResponseMetric]].MetricValues[i].Value),
				minResponseTime: parseFloatToInt64(d.Metrics[metricsIndex[minResponseMetric]].MetricValues[i].Value),
				maxResponseTime: parseFloatToInt64(d.Metrics[metricsIndex[maxResponseMetric]].MetricValues[i].Value),
			})
		}
	}

	j.createMetricEvents(logger, metricGroups)

	logger.Trace("finished processing env")
}
func (j *pollApigeeStats) createMetricEvents(logger log.FieldLogger, metricGroups map[string][]*metricData) {
	wg := sync.WaitGroup{}
	wg.Add(len(metricGroups))
	for key := range metricGroups {
		go func(group []*metricData) {
			defer wg.Done()
			m := &metricData{
				environment:     group[0].environment,
				name:            group[0].name,
				baseName:        group[0].baseName,
				timestamp:       group[0].timestamp,
				minResponseTime: group[0].minResponseTime,
				maxResponseTime: group[0].maxResponseTime,
			}

			for _, d := range group {
				if d.timestamp.After(m.timestamp) {
					m.timestamp = d.timestamp
				}
				if d.minResponseTime < m.minResponseTime {
					m.minResponseTime = d.minResponseTime
				}
				if d.maxResponseTime > m.maxResponseTime {
					m.maxResponseTime = d.maxResponseTime
				}
				m.avgResponseTime = ((m.avgResponseTime * float64(m.count)) + (d.avgResponseTime * float64(d.count))) / float64(m.count+d.count)
				m.count += d.count
				m.policyErrCount += d.policyErrCount
				m.serverErrCount += d.serverErrCount
			}
			j.processMetric(logger, m)
		}(metricGroups[key])
	}
	wg.Wait()
}

func (j *pollApigeeStats) getBaseProduct(name string) string {
	if !j.isProduct {
		// the dimension being queried is not api_product, return the name back
		return name
	}

	if p, found := j.clonedProduct[name]; found {
		return p
	}

	prod, err := j.client.GetProduct(name)
	if err != nil || prod == nil {
		return name
	}
	for _, att := range prod.Attributes {
		if att.Name == apigee.ClonedProdAttribute {
			j.clonedProduct[name] = att.Value
			return att.Value
		}
	}
	return name
}

func (j *pollApigeeStats) processMetric(logger log.FieldLogger, metData *metricData) {
	// calculate the successes
	success := metData.count - metData.policyErrCount - metData.serverErrCount
	if success < 0 {
		success = 0
	}

	logger = logger.WithField("success", success).
		WithField("policyErr", metData.policyErrCount).
		WithField("serverErr", metData.serverErrCount).
		WithField("time", j.endTime.Format(time.RFC822))

	logger.Debug("reporting metrics")

	reportMetric := func(count int64, code string) {
		if count == 0 {
			return
		}
		j.collector.AddAPIMetric(&metric.APIMetric{
			API: metricModels.APIDetails{
				ID:       util.FormatProxyID(metData.name),
				Name:     metData.name,
				Revision: 1,
			},
			StatusCode: code,
			Count:      count,
			StartTime:  metData.timestamp,
			Observation: metric.ObservationDetails{
				Start: metData.timestamp.Unix(),
				End:   time.Minute.Milliseconds(),
			},
		})
	}
	reportMetric(metData.policyErrCount, "400")
	reportMetric(metData.serverErrCount, "500")
	reportMetric(success, "200")

	logger.Info("finished processing metric")
}

func (j *pollApigeeStats) cleanCache() {
	// clean the cache, only need lastStarTtime
	for _, key := range j.statCache.GetKeys() {
		if key == lastStartTimeKey {
			continue
		}
		j.statCache.Delete(key)
	}
}
