package cmd

import (
	corecmd "github.com/Axway/agent-sdk/pkg/cmd"
	corecfg "github.com/Axway/agent-sdk/pkg/config"

	"github.com/Axway/agents-apigee/discovery/pkg/apigee"
	"github.com/Axway/agents-apigee/discovery/pkg/config"
)

// RootCmd - Agent root command
var RootCmd corecmd.AgentRootCmd
var apigeeConfig *config.ApigeeConfig

func init() {
	// Create new root command with callbacks to initialize the agent config and command execution.
	// The first parameter identifies the name of the yaml file that agent will look for to load the config
	RootCmd = corecmd.NewRootCmd(
		"apigee_discovery_agent", // Name of the yaml file
		"APIGEE Discovery Agent", // Agent description
		initConfig,               // Callback for initializing the agent config
		run,                      // Callback for executing the agent
		corecfg.DiscoveryAgent,   // Agent Type (Discovery or Traceability)
	)

	// Get the root command properties and bind the config property in YAML definition
	rootProps := RootCmd.GetProperties()
	rootProps.AddStringProperty("apigee.organization", "", "APIGEE Organization")
	rootProps.AddStringProperty("apigee.auth.username", "", "Username to use to authenticate to APIGEE")
	rootProps.AddStringProperty("apigee.auth.password", "", "Password for the user to authenticate to APIGEE")
}

// Callback that agent will call to process the execution
func run() error {
	apigeeClient, err := apigee.NewClient(apigeeConfig)
	if err == nil {
		apigeeClient.DiscoverAPIs()
	}
	return err
}

// Callback that agent will call to initialize the config. CentralConfig is parsed by Agent SDK
// and passed to the callback allowing the agent code to access the central config
func initConfig(centralConfig corecfg.CentralConfig) (interface{}, error) {
	rootProps := RootCmd.GetProperties()
	// Parse the config from bound properties and setup gateway config
	apigeeConfig = &config.ApigeeConfig{
		Organization: rootProps.StringPropertyValue("apigee.organization"),
		Auth: &config.AuthConfig{
			Username: rootProps.StringPropertyValue("apigee.auth.username"),
			Password: rootProps.StringPropertyValue("apigee.auth.password"),
		},
	}

	agentConfig := config.AgentConfig{
		CentralCfg: centralConfig,
		GatewayCfg: apigeeConfig,
	}
	return agentConfig, nil
}

// GetAgentConfig - Returns the agent config
func GetAgentConfig() *config.ApigeeConfig {
	return apigeeConfig
}
