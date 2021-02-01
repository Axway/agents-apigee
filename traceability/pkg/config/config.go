package config

import (
	"errors"

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
	LogFile        string      `config:"logFile"`
	ProcessOnInput bool        `config:"processOnInput"`
	Organization   string      `config:"organization"`
	Auth           *AuthConfig `config:"auth"`
}

// ValidateCfg - Validates the gateway config
func (a *ApigeeConfig) ValidateCfg() (err error) {
	if a.LogFile == "" {
		return errors.New("Invalid gateway configuration: logFile is not configured")
	}

	if a.Auth.Username == "" {
		return errors.New("Invalid gateway configuration: username is not configured")
	}

	if a.Auth.Password == "" {
		return errors.New("Invalid gateway configuration: password is not configured")
	}
	return
}

// GetAuth - Returns the Auth Config
func (a *ApigeeConfig) GetAuth() *AuthConfig {
	return a.Auth
}
