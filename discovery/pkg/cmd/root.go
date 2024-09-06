package cmd

import (
	"github.com/Axway/agent-sdk/pkg/apic"
	corecmd "github.com/Axway/agent-sdk/pkg/cmd"
	corecfg "github.com/Axway/agent-sdk/pkg/config"
	"github.com/Axway/agent-sdk/pkg/migrate"

	"github.com/Axway/agents-apigee/client/pkg/config"

	"github.com/Axway/agents-apigee/discovery/pkg/apigee"
)

// RootCmd - Agent root command
var RootCmd corecmd.AgentRootCmd
var apigeeClient *apigee.Agent

func init() {
	// Create new root command with callbacks to initialize the agent config and command execution.
	// The first parameter identifies the name of the yaml file that agent will look for to load the config
	RootCmd = corecmd.NewRootCmd(
		"apigee_discovery_agent", // Name of the yaml file
		"Apigee Discovery Agent", // Agent description
		initConfig,               // Callback for initializing the agent config
		run,                      // Callback for executing the agent
		corecfg.DiscoveryAgent,   // Agent Type (Discovery or Traceability)
	)

	// set the dataplane type that will be added to the agent spec
	corecfg.AgentDataPlaneType = apic.Apigee.String()

	// Get the root command properties and bind the config property in YAML definition
	rootProps := RootCmd.GetProperties()
	config.AddProperties(rootProps)

	migrate.MatchAttrPattern("-hash")
}

// Callback that agent will call to process the execution
func run() error {
	return apigeeClient.Run()
}

// Callback that agent will call to initialize the config. CentralConfig is parsed by Agent SDK
// and passed to the callback allowing the agent code to access the central config
func initConfig(centralConfig corecfg.CentralConfig) (interface{}, error) {
	rootProps := RootCmd.GetProperties()
	// Parse the config from bound properties and setup gateway config
	agentConfig := &apigee.AgentConfig{
		CentralCfg: centralConfig,
		ApigeeCfg:  config.ParseConfig(rootProps),
	}

	var err error
	apigeeClient, err = apigee.NewAgent(agentConfig)

	return agentConfig, err
}
