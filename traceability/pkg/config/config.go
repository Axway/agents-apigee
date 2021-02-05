package config

import (
	"errors"

	corecfg "github.com/Axway/agent-sdk/pkg/config"
)

// AgentConfig - represents the config for agent
type AgentConfig struct {
	CentralCfg   corecfg.CentralConfig `config:"central"`
	LogglyConfig *LogglyConfig         `config:"loggly"`
}

// LogglyConfig - represents the config for gateway
type LogglyConfig struct {
	corecfg.IConfigValidator
	Organization string `config:"organization"`
	APIToken     string `config:"token"`
}

// ValidateCfg - Validates the gateway config
func (a *LogglyConfig) ValidateCfg() (err error) {
	if a.Organization == "" {
		return errors.New("Invalid gateway configuration: organization is not configured")
	}

	if a.APIToken == "" {
		return errors.New("Invalid gateway configuration: apitoken is not configured")
	}
	return
}
