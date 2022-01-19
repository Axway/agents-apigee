package apigee

import (
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

// Agent - Represents the Gateway client
type Agent struct {
	cfg           *AgentConfig
	apigeeClient  *apigee.ApigeeClient
	statCache     cache.Cache
	statChannel   chan interface{}
	cacheFilePath string
}

// NewAgent - Creates a new Agent
func NewAgent(agentCfg *AgentConfig) (*Agent, error) {
	apigeeClient, err := apigee.NewClient(agentCfg.ApigeeCfg)
	if err != nil {
		return nil, err
	}

	agent := &Agent{
		apigeeClient:  apigeeClient,
		cfg:           agentCfg,
		statCache:     cache.New(),
		statChannel:   make(chan interface{}),
		cacheFilePath: traceability.GetDataDirPath() + "/" + apiStatCacheFile,
	}
	agent.statCache.Load(agent.cacheFilePath)

	// Start the poll api stats job
	// statPollJob := &pollApigeeStats{
	// 	apigeeClient: apigeeClient,
	// 	statChannel:  agent.statChannel,
	// }
	// jobs.RegisterIntervalJobWithName(statPollJob, 30*time.Second, "Apigee API Stats")
	registerPollStatsJob(agent)

	return agent, nil
}
