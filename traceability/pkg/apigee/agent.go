package apigee

import (
	"path/filepath"

	"github.com/Axway/agent-sdk/pkg/cache"
	corecfg "github.com/Axway/agent-sdk/pkg/config"
	"github.com/Axway/agent-sdk/pkg/traceability"

	"github.com/Axway/agents-apigee/client/pkg/apigee"
	"github.com/Axway/agents-apigee/client/pkg/config"
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
	statChannel   chan interface{}
	cacheFilePath string
	envs          []string
	catchUpDone   bool
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
		statChannel:  make(chan interface{}),
		catchUpDone:  false,
	}

	return thisAgent, nil
}

//CatchUpDone - signal when the catch up job is complete
func (a *Agent) CatchUpDone() {
	a.catchUpDone = true
}

//BeatsReady - signal that the beats are ready
func (a *Agent) BeatsReady() {
	a.setupCache()
	registerPollStatsJob(a)
	a.ready = true
}

//getApigeeEnvironments -
func (a *Agent) getApigeeEnvironments() {
	a.envs = a.apigeeClient.GetEnvironments()
}

//setupCache -
func (a *Agent) setupCache() {
	a.cacheFilePath = filepath.Join(traceability.GetDataDirPath(), "cache", apiStatCacheFile)
	a.statCache.Load(a.cacheFilePath)
}
