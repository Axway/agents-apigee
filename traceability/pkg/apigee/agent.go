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
	}
	if a.apigeeClient.GetConfig().IsProductMode() {
		baseOpts = append(baseOpts, withProductMode())
	}

	job := newPollStatsJob(append(baseOpts, withCacheClean())...)
	lastStatTimeIface, err := a.statCache.Get(lastStartTimeKey)
	if err == nil {
		//"2022-01-21T11:31:32.079962632-07:00"
		// there was a last time in the cache
		lastStatTime, _ := time.Parse(time.RFC3339Nano, lastStatTimeIface.(string))
		if time.Now().Add(time.Hour*-1).Sub(lastStatTime) > 0 {
			// last start time not within an hour
			catchUpJob := newPollStatsJob(
				append(baseOpts,
					[]func(*pollApigeeStats){
						withStartTime(lastStatTime),
						withIncrement(time.Hour),
					}...,
				)...,
			) // create the job that catches up and stops
			catchUpJob.Execute()
		}
	}

	jobs.RegisterIntervalJobWithName(job, a.cfg.ApigeeCfg.Intervals.Stats, "Apigee Stats")
	return "", nil
}
