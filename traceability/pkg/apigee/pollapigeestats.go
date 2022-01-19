package apigee

import (
	"time"

	"github.com/Axway/agent-sdk/pkg/jobs"
)

const (
	lastStartTimeKey = "lastStartTime"
	savedStatCacheKey
)

type pollApigeeStats struct {
	jobs.Job
	id        string
	agent     *Agent
	endTime   time.Time
	startTime time.Time
	lastTime  time.Time
	increment time.Duration // increment the end and start times by this amount
}

func newPollStatsJob(agent *Agent, options ...func(*pollApigeeStats)) *pollApigeeStats {
	job := &pollApigeeStats{
		agent: agent,
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

func registerPollStatsJob(agent *Agent) (string, error) {
	job := newPollStatsJob(agent) // create the job that runs every minute
	lastStatTimeIface, err := agent.statCache.Get(lastStartTimeKey)
	if err == nil {
		// there was a last time in the cache
		lastStatTime := lastStatTimeIface.(time.Time)
		if time.Now().Add(time.Hour*-1).Sub(lastStatTime) > 0 {
			// last start time not within an hour
			catchUpJob := newPollStatsJob(agent,
				withStartTime(lastStatTime),
				withEndTime(lastStatTime),
				withIncrement(time.Hour),
			) // create the job that catches up and stops
			jobID, _ := jobs.RegisterIntervalJobWithName(catchUpJob, time.Second*10, "Apigee Stats Catch Up")
			catchUpJob.setJobID(jobID) // add hte job id
		}
	}

	jobID, _ := jobs.RegisterIntervalJobWithName(job, time.Minute, "Apigee Stats")
	job.setJobID(jobID)
	return "", nil
}

func (j *pollApigeeStats) setJobID(id string) {
	j.id = id
}

func (j *pollApigeeStats) Ready() bool {
	return j.agent.apigeeClient.IsReady()
}

func (j *pollApigeeStats) Status() error {
	return nil
}

func (j *pollApigeeStats) Execute() error {
	lastTime := j.lastTime
	startTime := j.startTime
	if j.increment == 0 {
		lastTime = time.Now().Add(time.Minute * -10)  // go back 10 minutes
		startTime = time.Now().Add(time.Minute * -70) // go back 70 minutes
	}

	metrics, err := j.agent.apigeeClient.GetStats("prod", startTime, lastTime)
	if err != nil {
		return err
	}

	// TODO send to processor
	_ = metrics

	if j.increment != 0 {
		// update the start and lastTime times
		j.startTime = j.startTime.Add(j.increment)
		j.lastTime = j.lastTime.Add(j.increment)
		if j.startTime.Sub(j.endTime) > 0 {
			// all caught up
			jobs.UnregisterJob(j.id)
		}
	} else {
		j.agent.statCache.Set(lastStartTimeKey, startTime)
		j.agent.statCache.Save(j.agent.cacheFilePath)
	}

	return nil
}
