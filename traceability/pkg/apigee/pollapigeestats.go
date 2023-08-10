package apigee

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/Axway/agent-sdk/pkg/cache"
	"github.com/Axway/agent-sdk/pkg/jobs"
	"github.com/Axway/agent-sdk/pkg/transaction/metric"
	metricModels "github.com/Axway/agent-sdk/pkg/transaction/models"
	"github.com/Axway/agent-sdk/pkg/transaction/util"
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
)

type isReady func() bool

type metricData struct {
	environment    string
	baseName       string
	name           string
	timestamp      int64
	count          string
	policyErrCount string
	serverErrCount string
	responseTime   string
}

type pollApigeeStats struct {
	jobs.Job
	startTime     time.Time
	lastTime      time.Time
	increment     time.Duration // increment the end and start times by this amount
	cacheKeys     []string
	envs          []string
	mutex         *sync.Mutex
	cacheClean    bool
	collector     metric.Collector
	ready         isReady
	client        definitions.StatsClient
	statCache     cache.Cache
	cachePath     string
	clonedProduct map[string]string
	dimension     string
	isProduct     bool
	logger        log.FieldLogger
	metrics       []metric.Detail
}

func newPollStatsJob(options ...func(*pollApigeeStats)) *pollApigeeStats {
	job := &pollApigeeStats{
		collector:     metric.GetMetricCollector(),
		cacheKeys:     make([]string, 0),
		mutex:         &sync.Mutex{},
		clonedProduct: make(map[string]string),
		dimension:     "apiproxy",
		logger:        log.NewFieldLogger().WithComponent("pollStatsJob").WithPackage("apigee"),
	}

	for _, o := range options {
		o(job)
	}
	return job
}

func withStartTime(startTime time.Time) func(p *pollApigeeStats) {
	return func(p *pollApigeeStats) {
		p.startTime = startTime
	}
}

func withIncrement(increment time.Duration) func(p *pollApigeeStats) {
	return func(p *pollApigeeStats) {
		p.increment = increment
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

func withProductMode() func(p *pollApigeeStats) {
	return func(p *pollApigeeStats) {
		p.dimension = "api_product"
		p.isProduct = true
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

	//start the metric channel
	j.metrics = []metric.Detail{}

	logger.Trace("starting execution")
	j.envs = j.client.GetEnvironments()
	j.cacheKeys = make([]string, 0)
	lastTime := j.lastTime
	startTime := j.startTime
	if j.increment == 0 {
		lastTime = time.Now()
		startTime = time.Now().Add(time.Minute * -30) // go back 30 minutes
	}

	metricSelect := strings.Join([]string{countMetric, policyErrMetric, serverErrMetric, avgResponseMetric}, ",")
	wg := &sync.WaitGroup{}
	for _, e := range j.envs {
		wg.Add(1)
		go func(logger log.FieldLogger, envName string) {
			defer wg.Done()
			logger = logger.WithField("env", envName)
			metrics, err := j.client.GetStats(envName, j.dimension, metricSelect, startTime, lastTime)
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

	if j.increment != 0 {
		// update the start and lastTime times
		j.startTime = j.startTime.Add(j.increment)
		j.lastTime = j.lastTime.Add(j.increment)
	}

	// only update the lastStartTime when it is not zero
	if !startTime.IsZero() {
		j.statCache.Set(lastStartTimeKey, startTime.String())
		j.statCache.Save(j.cachePath)
	}

	j.sendMetrics(logger)
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
	}

	for i, m := range metrics.Environments[0].Dimensions[0].Metrics { // api_proxies or api_product
		if _, found := metricsIndex[m.Name]; !found {
			logger.Warnf("skipping metric, %s, in return data", m.Name)
		}
		metricsIndex[m.Name] = i
	}

	// check for -1 index in metricsIndex
	for key, index := range metricsIndex {
		if index < 0 {
			logger.Errorf("did not find the %s metric in the returned data", key)
			return
		}
	}

	dimensions := []string{}
	for _, d := range metrics.Environments[0].Dimensions { // api_proxies
		dimensions = append(dimensions, d.Name)
	}

	logger.WithField("value", dimensions).Trace("dimensions")
	wg := sync.WaitGroup{}
	for _, d := range metrics.Environments[0].Dimensions { // api_proxies
		logger := logger.WithField("dimension", d.Name)
		logger.Trace("processing metric for dimension")
		for i := range d.Metrics[0].MetricValues {
			if d.Metrics[metricsIndex[countMetric]].MetricValues[i].Value == "0.0" {
				continue
			}

			data := &metricData{
				environment:    metrics.Environments[0].Name,
				name:           j.getBaseProduct(d.Name),
				baseName:       d.Name,
				timestamp:      d.Metrics[metricsIndex[countMetric]].MetricValues[i].Timestamp,
				count:          d.Metrics[metricsIndex[countMetric]].MetricValues[i].Value,
				policyErrCount: d.Metrics[metricsIndex[policyErrMetric]].MetricValues[i].Value,
				serverErrCount: d.Metrics[metricsIndex[serverErrMetric]].MetricValues[i].Value,
				responseTime:   d.Metrics[metricsIndex[avgResponseMetric]].MetricValues[i].Value,
			}
			wg.Add(1)

			go func(md *metricData) {
				j.processMetric(logger, md)
				wg.Done()
			}(data)
		}
	}
	wg.Wait()
	logger.Trace("finished processing env")
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

func (j *pollApigeeStats) sendMetrics(logger log.FieldLogger) {
	logger.Trace("sending all metrics")
	for _, msg := range j.metrics {
		j.collector.AddMetricDetail(msg)
	}
}

func (j *pollApigeeStats) processMetric(logger log.FieldLogger, metData *metricData) {
	metricCacheKey := fmt.Sprintf("%s-%s-%d", metData.environment, metData.baseName, metData.timestamp)
	logger = logger.WithField("metricKey", metricCacheKey)
	logger.Trace("begin processing metric")

	j.mutex.Lock()
	j.cacheKeys = append(j.cacheKeys, metricCacheKey)
	j.mutex.Unlock()

	// get the cached values
	newMetricData := &metricCache{
		ProxyName: metData.name,
		Timestamp: metData.timestamp,
	}
	if data, err := j.statCache.Get(metricCacheKey); err == nil {
		stringData := data.(string)
		err := json.Unmarshal([]byte(stringData), &newMetricData)
		if err != nil {
			return
		}
	}

	// get the average response time
	s, err := strconv.ParseFloat(metData.responseTime, 64)
	if err != nil {
		logger.Errorf("could not read message metric value %s, at timestamp %d, as it was not a number", metData.count, metData.timestamp)
		return
	}
	newMetricData.ResponseTime = int64(s)

	// get the total messages
	s, err = strconv.ParseFloat(metData.count, 64)
	if err != nil {
		logger.Errorf("could not read message metric value %s, at timestamp %d, as it was not a number", metData.count, metData.timestamp)
		return
	}
	newMetricData.Total = int(s)

	// get the policy errors
	s, err = strconv.ParseFloat(metData.policyErrCount, 64)
	if err != nil {
		logger.Errorf("could not read policy error metric value %s, at timestamp %d, as it was not a number", metData.policyErrCount, metData.timestamp)
		return
	}
	newMetricData.PolicyError = int(s)

	// get the server errors
	s, err = strconv.ParseFloat(metData.serverErrCount, 64)
	if err != nil {
		logger.Errorf("could not read server error metric value %s, at timestamp %d, as it was not a number", metData.serverErrCount, metData.timestamp)
		return
	}
	newMetricData.ServerError = int(s)

	// calculate the number of successes
	newMetricData.Success = newMetricData.Total - newMetricData.PolicyError - newMetricData.ServerError

	// create the api details structure for the metric collector
	apiName := fmt.Sprintf("%s (%s)", metData.name, metData.environment)
	apiID := util.FormatProxyID(fmt.Sprintf("%s-%s", metData.name, metData.environment))
	if j.isProduct {
		apiName = metData.name
		apiID = util.FormatProxyID(metData.name)
	}

	apiDetail := metricModels.APIDetails{
		ID:       apiID,
		Name:     apiName,
		Revision: 1,
	}
	logger = logger.WithField("success", newMetricData.Success).WithField("policyErr", newMetricData.PolicyError).WithField("serverErr", newMetricData.ServerError)
	logger.Debug("reporting metrics")

	thisMetric := metric.Detail{
		APIDetails: apiDetail,
		Duration:   newMetricData.ResponseTime,
		Bytes:      0,
		AppDetails: metricModels.AppDetails{},
	}

	appendFunc := func(m metric.Detail) {
		j.mutex.Lock()
		defer j.mutex.Unlock()
		j.metrics = append(j.metrics, m)
	}

	thisMetric.StatusCode = "400"
	for i := newMetricData.PolicyError - newMetricData.ReportedPolicyError; i > 0; i-- {
		appendFunc(thisMetric)
		newMetricData.ReportedPolicyError++
	}
	thisMetric.StatusCode = "500"
	for i := newMetricData.ServerError - newMetricData.ReportedServerError; i > 0; i-- {
		appendFunc(thisMetric)
		newMetricData.ReportedServerError++
	}
	thisMetric.StatusCode = "200"
	for i := newMetricData.Success - newMetricData.ReportedSuccess; i > 0; i-- {
		appendFunc(thisMetric)
		newMetricData.ReportedSuccess++
	}

	// convert the metric data to a json string
	data, err := json.Marshal(newMetricData)
	if err != nil {
		return
	}
	err = j.statCache.Set(metricCacheKey, string(data))
	if err != nil {
		logger.Error(err)
	}
	logger.Trace("finished processing metric")
}

func (j *pollApigeeStats) cleanCache() {
	// get cache keys from cache
	knownKeys := j.statCache.GetKeys()

	// find the keys that can be cleaned
	cleanKeys := make([]string, 0)
	keysMap := make(map[string]struct{})

	// add keys that should be kept
	j.mutex.Lock()
	for _, key := range j.cacheKeys {
		keysMap[key] = struct{}{}
	}
	j.mutex.Unlock()

	// find keys not in the keysMap, these should be cleaned
	for _, key := range knownKeys {
		if key == lastStartTimeKey {
			continue
		}
		if _, found := keysMap[key]; !found {
			cleanKeys = append(cleanKeys, key)
		}
	}

	// clean the cache items with keys from cleanKeys
	for _, key := range cleanKeys {
		j.statCache.Delete(key)
	}
}
