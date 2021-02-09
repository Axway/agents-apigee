package cmd

import (
	"time"

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
	rootProps.AddStringProperty("apigee.organization", "", "APIGEE Organization")
	rootProps.AddStringProperty("apigee.auth.username", "", "Username to use to authenticate to APIGEE")
	rootProps.AddStringProperty("apigee.auth.password", "", "Password for the user to authenticate to APIGEE")
	rootProps.AddDurationProperty("apigee.pollInterval", 30*time.Second, "The time interval between checking for new APIGEE resources")
	rootProps.AddStringProperty("apigee.loggly.customertoken", "", "The Loggly Customer Token for sending log events")
	rootProps.AddStringProperty("apigee.loggly.apitoken", "", "The Loggly API Token for retrieving log events")
	rootProps.AddStringProperty("apigee.loggly.subdomain", "", "The Loggly subdomain")
	rootProps.AddStringProperty("apigee.loggly.host", "logs-01.loggly.com", "The Loggly Host URL")
	rootProps.AddStringProperty("apigee.loggly.port", "514", "The Loggly Port")

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
		Organization: rootProps.StringPropertyValue("apigee.organization"),
		PollInterval: rootProps.DurationPropertyValue("apigee.pollInterval"),
		Auth: &config.AuthConfig{
			Username: rootProps.StringPropertyValue("apigee.auth.username"),
			Password: rootProps.StringPropertyValue("apigee.auth.password"),
		},
		Loggly: &config.LogglyConfig{
			Subdomain:     rootProps.StringPropertyValue("apigee.loggly.subdomain"),
			CustomerToken: rootProps.StringPropertyValue("apigee.loggly.customertoken"),
			APIToken:      rootProps.StringPropertyValue("apigee.loggly.apitoken"),
			Host:          rootProps.StringPropertyValue("apigee.loggly.host"),
			Port:          rootProps.StringPropertyValue("apigee.loggly.port"),
		},
	}

	agentConfig := &config.AgentConfig{
		CentralCfg: centralConfig,
		GatewayCfg: apigeeConfig,
	}

	beater.SetLogglyConfig(apigeeConfig.GetLoggly())

	return agentConfig, nil
}
