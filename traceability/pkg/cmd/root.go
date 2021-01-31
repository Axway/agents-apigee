package cmd

import (
	corecmd "github.com/Axway/agent-sdk/pkg/cmd"
	corecfg "github.com/Axway/agent-sdk/pkg/config"

	libcmd "github.com/elastic/beats/v7/libbeat/cmd"
	"github.com/elastic/beats/v7/libbeat/cmd/instance"

	"github.com/Axway/agents-apigee/traceability/pkg/beater"
	"github.com/Axway/agents-apigee/traceability/pkg/config"
)

// RootCmd - Agent root command
var RootCmd corecmd.AgentRootCmd
var beatCmd *libcmd.BeatsRootCmd
var apigeeConfig *config.ApigeeConfig

func init() {
	name := "apigee_traceability_agent"
	settings := instance.Settings{
		Name:          name,
		HasDashboards: true,
	}

	// Initialize the beat command
	beatCmd = libcmd.GenRootCmdWithSettings(beater.New, settings)
	cmd := beatCmd.Command
	// Wrap the beat command with the agent command processor with callbacks to initialize the agent config and command execution.
	// The first parameter identifies the name of the yaml file that agent will look for to load the config
	RootCmd = corecmd.NewCmd(
		&cmd,
		name,                        // Name of the agent and yaml config file
		"Apigee Traceability Agent", // Agent description
		initConfig,                  // Callback for initializing the agent config
		run,                         // Callback for executing the agent
		corecfg.TraceabilityAgent,   // Agent Type (Discovery or Traceability)
	)

	// Get the root command properties and bind the config property in YAML definition
	rootProps := RootCmd.GetProperties()
	rootProps.AddStringProperty("gateway-section.logFile", "./logs/traffic.log", "Sample log file with traffic event from gateway")
	rootProps.AddBoolProperty("gateway-section.processOnInput", true, "Flag to process received event on input or by output before publishing the event by transport")
	rootProps.AddStringProperty("apigee.organization", "", "APIGEE Organization")
	rootProps.AddStringProperty("apigee.auth.username", "", "Username to use to authenticate to APIGEE")
	rootProps.AddStringProperty("apigee.auth.password", "", "Password for the user to authenticate to APIGEE")

}

// Callback that agent will call to process the execution
func run() error {
	return beatCmd.Execute()
}

// Callback that agent will call to initialize the config. CentralConfig is parsed by Agent SDK
// and passed to the callback allowing the agent code to access the central config
func initConfig(centralConfig corecfg.CentralConfig) (interface{}, error) {
	rootProps := RootCmd.GetProperties()
	// Parse the config from bound properties and setup gateway config
	apigeeConfig := &config.ApigeeConfig{
		LogFile:        rootProps.StringPropertyValue("gateway-section.logFile"),
		ProcessOnInput: rootProps.BoolPropertyValue("gateway-section.processOnInput"),
		Organization:   rootProps.StringPropertyValue("apigee.organization"),
		Auth: &config.AuthConfig{
			Username: rootProps.StringPropertyValue("apigee.auth.username"),
			Password: rootProps.StringPropertyValue("apigee.auth.password"),
		},
	}

	agentConfig := &config.AgentConfig{
		CentralCfg: centralConfig,
		GatewayCfg: apigeeConfig,
	}
	beater.SetGatewayConfig(apigeeConfig)

	return agentConfig, nil
}

// GetAgentConfig - Returns the agent config
func GetAgentConfig() *config.ApigeeConfig {
	return apigeeConfig
}
