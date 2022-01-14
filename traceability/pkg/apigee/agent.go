package apigee

import (
	corecfg "github.com/Axway/agent-sdk/pkg/config"

	"github.com/Axway/agents-apigee/client/pkg/apigee"
	"github.com/Axway/agents-apigee/client/pkg/config"
)

// AgentConfig - represents the config for agent
type AgentConfig struct {
	CentralCfg corecfg.CentralConfig `config:"central"`
	ApigeeCfg  *config.ApigeeConfig  `config:"apigee"`
	// LogglyCfg  *logglycfg.LogglyConfig `config:"loggly"`
}

// Agent - Represents the Gateway client
type Agent struct {
	cfg          *AgentConfig
	apigeeClient *apigee.ApigeeClient
}

// NewAgent - Creates a new Agent
func NewAgent(agentCfg *AgentConfig) (*Agent, error) {
	apigeeClient, err := apigee.NewClient(agentCfg.ApigeeCfg)
	if err != nil {
		return nil, err
	}

	agent := &Agent{
		apigeeClient: apigeeClient,
		cfg:          agentCfg,
	}

	// create job that registers the shared flow
	// _, err = jobs.RegisterSingleRunJobWithName(newRegisterSharedFlowJob(agent.apigeeClient, agentCfg.LogglyCfg), "Register Shared Flow")

	return agent, nil
}
