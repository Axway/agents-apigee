package apigee

import (
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/Axway/agent-sdk/pkg/cache"
	corecfg "github.com/Axway/agent-sdk/pkg/config"
	"github.com/Axway/agent-sdk/pkg/jobs"
	"github.com/Axway/agent-sdk/pkg/traceability"

	"github.com/Axway/agents-apigee/client/pkg/apigee"
	"github.com/Axway/agents-apigee/client/pkg/config"
	"github.com/Axway/agents-apigee/traceability/pkg/apigee/definitions"
	"github.com/Axway/agents-apigee/traceability/pkg/apigee/statsmock"
)

const (
	apiStatCacheFile = "stat-cache-data.json"
)

// AgentConfig - represents the config for agent
type AgentConfig struct {
	CentralCfg corecfg.CentralConfig `config:"central"`
	ApigeeCfg  *config.ApigeeConfig  `config:"apigee"`
}

var thisAgent *Agent

// GetAgent - returns the agent
func GetAgent() *Agent {
	return thisAgent
}

// Agent - Represents the Gateway client
type Agent struct {
	cfg           *AgentConfig
	apigeeClient  *apigee.ApigeeClient
	statCache     cache.Cache
	cacheFilePath string
	ready         bool
}

// NewAgent - Creates a new Agent
func NewAgent(agentCfg *AgentConfig) (*Agent, error) {
	apigeeClient, err := apigee.NewClient(agentCfg.ApigeeCfg)
	if err != nil {
		return nil, err
	}

	thisAgent = &Agent{
		apigeeClient: apigeeClient,
		cfg:          agentCfg,
		statCache:    cache.New(),
	}

	return thisAgent, nil
}

// BeatsReady - signal that the beats are ready
func (a *Agent) BeatsReady() {
	a.setupCache()
	a.registerPollStatsJob()
	a.ready = true
}

func (a *Agent) IsReady() bool {
	return a.ready && a.apigeeClient.IsReady()
}

// setupCache -
func (a *Agent) setupCache() {
	a.cacheFilePath = filepath.Join(traceability.GetDataDirPath(), "cache", apiStatCacheFile)
	a.statCache.Load(a.cacheFilePath)
}

func (a *Agent) registerPollStatsJob() (string, error) {
	var client definitions.StatsClient = a.apigeeClient

	val := os.Getenv("QA_SIMULATE_APIGEE_STATS")
	if strings.ToLower(val) == "true" {
		products, _ := a.apigeeClient.GetProducts()
		client = statsmock.NewStatsMock(a.apigeeClient, products)
	}

	// create the job that runs every minute
	baseOpts := []func(*pollApigeeStats){
		withStatsClient(client),
		withIsReady(a.IsReady),
		withStatsCache(a.statCache),
		withCachePath(a.cacheFilePath),
		withAllTraffic(a.cfg.ApigeeCfg.ShouldReportAllTraffic()),
		withNotSetTraffic(a.cfg.ApigeeCfg.ShouldReportNotSetTraffic()),
		withFilteredAPIs(a.cfg.ApigeeCfg.FilteredAPIs),
	}
	if a.cfg.ApigeeCfg.IsProductMode() {
		baseOpts = append(baseOpts, withProductMode())
	}

	lastStatTimeIface, err := a.statCache.Get(lastStartTimeKey)
	var lastStartTime time.Time
	if err == nil {
		// there was a last time in the cache
		lastStartTime, _ = time.Parse(time.RFC3339Nano, lastStatTimeIface.(string))
	}

	if !lastStartTime.IsZero() {
		// last start time not zero

		// create the job that executes once to get all the stats that were missed
		catchUpJob := newPollStatsJob(
			append(baseOpts,
				withStartTime(lastStartTime),
			)...,
		)
		catchUpJob.Execute()

		// register the regular running job after one interval has passed
		go func() {
			time.Sleep(a.cfg.ApigeeCfg.Intervals.Stats)
			job := newPollStatsJob(append(baseOpts, withCacheClean(), withStartTime(catchUpJob.startTime))...)
			jobs.RegisterIntervalJobWithName(job, a.cfg.ApigeeCfg.Intervals.Stats, "Apigee Stats")
		}()
	} else {
		// register a regular running job, only grabbing hte last hour of stats
		job := newPollStatsJob(append(baseOpts, withCacheClean(), withStartTime(time.Now().Add(time.Hour*-1).Truncate(time.Minute)))...)
		jobs.RegisterIntervalJobWithName(job, a.cfg.ApigeeCfg.Intervals.Stats, "Apigee Stats")
	}

	return "", nil
}
