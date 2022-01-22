package apigee

import (
	"encoding/json"
	"fmt"
	"strconv"
	"sync"
	"time"

	"github.com/Axway/agent-sdk/pkg/jobs"
	"github.com/Axway/agent-sdk/pkg/transaction/metric"
	"github.com/Axway/agent-sdk/pkg/util/log"
	"github.com/Axway/agents-apigee/client/pkg/apigee/models"
)

const lastStartTimeKey = "lastStartTime"

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

	wg := &sync.WaitGroup{}
	for _, envName := range j.agent.envs {
		wg.Add(1)
		go func(envName string) {
			defer wg.Done()
			metrics, err := j.agent.apigeeClient.GetStats(envName, startTime, lastTime)
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
	wg := &sync.WaitGroup{}
	for _, e := range metrics.Environments { // prod,test
		for _, d := range e.Dimensions { // api_proxies
			if len(d.Metrics) == 0 {
				log.Error("metric data did not have the expected number of metric types, sum(message_count) and sum(is_error)")
				return
			}

			for i, messageCount := range d.Metrics[0].Values { // sum(message_count),sum(is_error)
				// get the error count metric
				wg.Add(1)
				errorCount := d.Metrics[1].Values[i]
				go func(i int, messageCount, errorCount models.MetricsValues, envName, proxyName string) {
					defer wg.Done()
					j.processMetric(i, messageCount, errorCount, envName, proxyName)
				}(i, messageCount, errorCount, e.Name, d.Name)
			}
		}
	}
	wg.Wait()
}

func (j *pollApigeeStats) processMetric(i int, messageCount, errorCount models.MetricsValues, envName, proxyName string) {
	metricCacheKey := fmt.Sprintf("%s-%s-%d", envName, proxyName, messageCount.Timestamp)
	j.cacheKeys = append(j.cacheKeys, metricCacheKey)

	// get the cached values
	metricData := &metricCache{
		ProxyName: proxyName,
		Timestamp: messageCount.Timestamp,
	}
	if data, err := j.agent.statCache.Get(metricCacheKey); err == nil {
		stringData := data.(string)
		err := json.Unmarshal([]byte(stringData), &metricData)
		if err != nil {
			return
		}
	}

	// get the total messages
	s, err := strconv.ParseFloat(messageCount.Value, 64)
	if err != nil {
		log.Errorf("could not read message metric value %s, at timestamp %d, as it was not a number", messageCount.Value, messageCount.Timestamp)
		return
	}
	metricData.Total = int(s)

	// get the total errors
	s, err = strconv.ParseFloat(errorCount.Value, 64)
	if err != nil {
		log.Errorf("could not read error metric value %s, at timestamp %d, as it was not a number", errorCount.Value, errorCount.Timestamp)
		return
	}
	metricData.Error = int(s)

	// calculate the number of successes
	metricData.Success = metricData.Total - metricData.Error

	// create teh api details structure for the metric collector
	details := metric.APIDetails{
		ID:   fmt.Sprintf("%s-%s", proxyName, envName),
		Name: fmt.Sprintf("%s-%s", proxyName, envName),
	}
	if metricData.ReportedError < metricData.Error {
		count := metricData.Error - metricData.ReportedError
		for count > 0 {
			j.collector.AddMetric(details, "400", 0, 0, "", "")
			count--
			metricData.ReportedError++
		}
	}
	if metricData.ReportedSuccess < metricData.Success {
		count := metricData.Success - metricData.ReportedSuccess
		for count > 0 {
			j.collector.AddMetric(details, "200", 0, 0, "", "")
			count--
			metricData.ReportedSuccess++
		}
	}
	// convert the metric data to a json string
	data, err := json.Marshal(metricData)
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
