package apigee

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/Axway/agent-sdk/pkg/jobs"
	"github.com/Axway/agent-sdk/pkg/transaction/metric"
	"github.com/Axway/agent-sdk/pkg/util/log"
	"github.com/Axway/agents-apigee/client/pkg/apigee/models"
)

const (
	lastStartTimeKey  = "lastStartTime"
	countMetric       = "sum(message_count)"
	policyErrMetric   = "sum(policy_error)"
	serverErrMetric   = "sum(target_error)"
	avgResponseMetric = "avg(total_response_time)"
)

type metricData struct {
	environment    string
	name           string
	timestamp      int64
	count          string
	policyErrCount string
	serverErrCount string
	responseTime   string
}

type pollApigeeStats struct {
	jobs.Job
	id         string
	agent      *Agent
	endTime    time.Time
	startTime  time.Time
	lastTime   time.Time
	increment  time.Duration // increment the end and start times by this amount
	cacheKeys  []string
	cacheClean bool
	collector  metric.Collector
}

func newPollStatsJob(agent *Agent, options ...func(*pollApigeeStats)) *pollApigeeStats {
	job := &pollApigeeStats{
		agent:     agent,
		collector: metric.GetMetricCollector(),
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

func withEndTime(endTime time.Time) func(p *pollApigeeStats) {
	return func(p *pollApigeeStats) {
		p.endTime = endTime
	}
}

func withIncrement(increment time.Duration) func(p *pollApigeeStats) {
	return func(p *pollApigeeStats) {
		p.increment = increment
	}
}

func withCacheClean(increment time.Duration) func(p *pollApigeeStats) {
	return func(p *pollApigeeStats) {
		p.cacheKeys = make([]string, 0)
		p.cacheClean = true
	}
}

func registerPollStatsJob(agent *Agent) (string, error) {
	// create the job that runs every minute
	job := newPollStatsJob(agent,
		withCacheClean(time.Hour),
	)
	lastStatTimeIface, err := agent.statCache.Get(lastStartTimeKey)
	if err == nil {
		//"2022-01-21T11:31:32.079962632-07:00"
		// there was a last time in the cache
		lastStatTime, _ := time.Parse(time.RFC3339Nano, lastStatTimeIface.(string))
		if time.Now().Add(time.Hour*-1).Sub(lastStatTime) > 0 {
			// last start time not within an hour
			catchUpJob := newPollStatsJob(agent,
				withStartTime(lastStatTime),
				withEndTime(lastStatTime),
				withIncrement(time.Hour),
			) // create the job that catches up and stops
			jobID, _ := jobs.RegisterIntervalJobWithName(catchUpJob, time.Second*10, "Apigee Stats Catch Up")
			catchUpJob.setJobID(jobID) // add the job id
		}
	} else {
		agent.CatchUpDone()
	}

	jobID, _ := jobs.RegisterIntervalJobWithName(job, time.Minute, "Apigee Stats")
	job.setJobID(jobID)
	return "", nil
}

func (j *pollApigeeStats) setJobID(id string) {
	j.id = id
}

func (j *pollApigeeStats) Ready() bool {
	return j.agent.ready && j.agent.apigeeClient.IsReady()
}

func (j *pollApigeeStats) Status() error {
	return nil
}

func (j *pollApigeeStats) Execute() error {
	j.agent.getApigeeEnvironments()
	lastTime := j.lastTime
	startTime := j.startTime
	if j.increment == 0 {
		lastTime = time.Now()
		startTime = time.Now().Add(time.Minute * -30) // go back 30 minutes
	}

	metricSelect := strings.Join([]string{countMetric, policyErrMetric, serverErrMetric, avgResponseMetric}, ",")
	wg := &sync.WaitGroup{}
	for _, envName := range j.agent.envs {
		wg.Add(1)
		go func(envName string) {
			defer wg.Done()
			metrics, err := j.agent.apigeeClient.GetStats(envName, metricSelect, startTime, lastTime)
			if err != nil {
				return
			}

			j.processMetricResponse(metrics)
		}(envName)
	}
	wg.Wait()
	if j.cacheClean {
		j.cleanCache()
	}

	if j.increment != 0 {
		// update the start and lastTime times
		j.startTime = j.startTime.Add(j.increment)
		j.lastTime = j.lastTime.Add(j.increment)
		j.agent.statCache.Set(lastStartTimeKey, startTime.String())
		if j.startTime.Sub(j.endTime) > 0 {
			// all caught up
			j.agent.CatchUpDone()
			go jobs.UnregisterJob(j.id)
		}
	} else if j.agent.catchUpDone {
		j.agent.statCache.Set(lastStartTimeKey, startTime.String())
	}
	j.agent.statCache.Save(j.agent.cacheFilePath)

	return nil
}

func (j *pollApigeeStats) processMetricResponse(metrics *models.Metrics) {
	if len(metrics.Environments) != 1 {
		log.Error("exactly 1 environment should be returned")
		return
	}

	if len(metrics.Environments[0].Dimensions) == 0 {
		log.Trace("At least one proxy is needed to process response data")
		return
	}

	wg := &sync.WaitGroup{}
	// get the index of each metric
	var metricsIndex = map[string]int{
		countMetric:       -1,
		policyErrMetric:   -1,
		serverErrMetric:   -1,
		avgResponseMetric: -1,
	}

	for i, m := range metrics.Environments[0].Dimensions[0].Metrics { // api_proxies
		if _, found := metricsIndex[m.Name]; !found {
			log.Warnf("skipping metric, %s, in return data", m.Name)
		}
		metricsIndex[m.Name] = i
	}

	// TODO check for -1 index in metricsIndex

	for _, d := range metrics.Environments[0].Dimensions { // api_proxies
		for i := range d.Metrics[0].MetricValues {
			metData := &metricData{
				environment:    metrics.Environments[0].Name,
				name:           d.Name,
				count:          d.Metrics[metricsIndex[countMetric]].MetricValues[i].Value,
				policyErrCount: d.Metrics[metricsIndex[policyErrMetric]].MetricValues[i].Value,
				serverErrCount: d.Metrics[metricsIndex[serverErrMetric]].MetricValues[i].Value,
				responseTime:   d.Metrics[metricsIndex[avgResponseMetric]].MetricValues[i].Value,
			}
			// get the error count metric
			wg.Add(1)
			go func(metData *metricData) {
				defer wg.Done()
				j.processMetric(metData)
			}(metData)
		}
	}
	wg.Wait()
}

func (j *pollApigeeStats) processMetric(metData *metricData) {
	metricCacheKey := fmt.Sprintf("%s-%s-%d", metData.environment, metData.name, metData.timestamp)
	j.cacheKeys = append(j.cacheKeys, metricCacheKey)

	// get the cached values
	newMetricData := &metricCache{
		ProxyName: metData.name,
		Timestamp: metData.timestamp,
	}
	if data, err := j.agent.statCache.Get(metricCacheKey); err == nil {
		stringData := data.(string)
		err := json.Unmarshal([]byte(stringData), &newMetricData)
		if err != nil {
			return
		}
	}

	// get the average response time
	s, err := strconv.ParseFloat(metData.responseTime, 64)
	if err != nil {
		log.Errorf("could not read message metric value %s, at timestamp %d, as it was not a number", metData.count, metData.timestamp)
		return
	}
	newMetricData.ResponseTime = int64(s)

	// get the total messages
	s, err = strconv.ParseFloat(metData.count, 64)
	if err != nil {
		log.Errorf("could not read message metric value %s, at timestamp %d, as it was not a number", metData.count, metData.timestamp)
		return
	}
	newMetricData.Total = int(s)

	// get the policy errors
	s, err = strconv.ParseFloat(metData.policyErrCount, 64)
	if err != nil {
		log.Errorf("could not read policy error metric value %s, at timestamp %d, as it was not a number", metData.policyErrCount, metData.timestamp)
		return
	}
	newMetricData.PolicyError = int(s)

	// get the server errors
	s, err = strconv.ParseFloat(metData.serverErrCount, 64)
	if err != nil {
		log.Errorf("could not read server error metric value %s, at timestamp %d, as it was not a number", metData.serverErrCount, metData.timestamp)
		return
	}
	newMetricData.ServerError = int(s)

	// calculate the number of successes
	newMetricData.Success = newMetricData.Total - newMetricData.PolicyError - newMetricData.ServerError

	// create teh api details structure for the metric collector
	details := metric.APIDetails{
		ID:       fmt.Sprintf("%s-%s", metData.name, metData.environment),
		Name:     fmt.Sprintf("%s (%s)", metData.name, metData.environment),
		Revision: 1,
	}
	if newMetricData.ReportedPolicyError < newMetricData.PolicyError {
		count := newMetricData.PolicyError - newMetricData.ReportedPolicyError
		for count > 0 {
			j.collector.AddMetric(details, "400", newMetricData.ResponseTime, 0, "", "")
			count--
			newMetricData.ReportedPolicyError++
		}
	}
	if newMetricData.ReportedServerError < newMetricData.ServerError {
		count := newMetricData.ServerError - newMetricData.ReportedServerError
		for count > 0 {
			j.collector.AddMetric(details, "500", newMetricData.ResponseTime, 0, "", "")
			count--
			newMetricData.ReportedServerError++
		}
	}
	if newMetricData.ReportedSuccess < newMetricData.Success {
		count := newMetricData.Success - newMetricData.ReportedSuccess
		for count > 0 {
			j.collector.AddMetric(details, "200", newMetricData.ResponseTime, 0, "", "")
			count--
			newMetricData.ReportedSuccess++
		}
	}
	// convert the metric data to a json string
	data, err := json.Marshal(newMetricData)
	if err != nil {
		return
	}
	err = j.agent.statCache.Set(metricCacheKey, string(data))
	if err != nil {
		log.Error(err)
	}
}

func (j *pollApigeeStats) cleanCache() {
	// get cache keys from cache
	knownKeys := j.agent.statCache.GetKeys()

	// find the keys that can be cleaned
	cleanKeys := make([]string, 0)
	keysMap := make(map[string]struct{})

	// add keys that should be kept
	for _, key := range j.cacheKeys {
		keysMap[key] = struct{}{}
	}

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
		j.agent.statCache.Delete(key)
	}
}
