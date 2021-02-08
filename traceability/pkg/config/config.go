package config

import (
	"errors"
	"time"

	corecfg "github.com/Axway/agent-sdk/pkg/config"
)

// AgentConfig - represents the config for agent
type AgentConfig struct {
	CentralCfg corecfg.CentralConfig `config:"central"`
	GatewayCfg *ApigeeConfig         `config:"apigee"`
}

// ApigeeConfig - represents the config for gateway
type ApigeeConfig struct {
	corecfg.IConfigValidator
	Organization string        `config:"organization"`
	Auth         *AuthConfig   `config:"auth"`
	Loggly       *LogglyConfig `config:"loggly"`
	PollInterval time.Duration `config:"pollInterval"`
}

// ValidateCfg - Validates the gateway config
func (a *ApigeeConfig) ValidateCfg() (err error) {
	if a.Auth.Username == "" {
		return errors.New("Invalid gateway configuration: username is not configured")
	}

	if a.Auth.Password == "" {
		return errors.New("Invalid gateway configuration: password is not configured")
	}

	if a.Loggly.Organization == "" {
		return errors.New("Invalid gateway configuration: organization is not configured")
	}

	if a.Loggly.APIToken == "" {
		return errors.New("Invalid gateway configuration: apitoken is not configured")
	}

	return
}

// GetAuth - Returns the Auth Config
func (a *ApigeeConfig) GetAuth() *AuthConfig {
	return a.Auth
}

// GetLoggly - Returns the Loggly Config
func (a *ApigeeConfig) GetLoggly() *LogglyConfig {
	return a.Loggly
}

// GetPollInterval - Returns the Poll Interval
func (a *ApigeeConfig) GetPollInterval() time.Duration {
	return a.PollInterval
}
