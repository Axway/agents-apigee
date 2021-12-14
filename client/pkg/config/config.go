package config

import (
	"errors"
	"time"

	"github.com/Axway/agent-sdk/pkg/cmd/properties"
	corecfg "github.com/Axway/agent-sdk/pkg/config"
)

// ApigeeConfig - represents the config for gateway
type ApigeeConfig struct {
	corecfg.IConfigValidator
	Organization string        `config:"organization"`
	Auth         *AuthConfig   `config:"auth"`
	PollInterval time.Duration `config:"pollInterval"`
}

// AddProperties - adds config needed for apigee client
func AddProperties(rootProps properties.Properties) {
	rootProps.AddStringProperty("apigee.organization", "", "APIGEE Organization")
	rootProps.AddStringProperty("apigee.auth.username", "", "Username to use to authenticate to APIGEE")
	rootProps.AddStringProperty("apigee.auth.password", "", "Password for the user to authenticate to APIGEE")
	rootProps.AddDurationProperty("apigee.pollInterval", 30*time.Second, "The time interval between checking for new APIGEE resources")
}

// ParseConfig - parse the config on startup
func ParseConfig(rootProps properties.Properties) *ApigeeConfig {
	return &ApigeeConfig{
		Organization: rootProps.StringPropertyValue("apigee.organization"),
		PollInterval: rootProps.DurationPropertyValue("apigee.pollInterval"),
		Auth: &AuthConfig{
			Username: rootProps.StringPropertyValue("apigee.auth.username"),
			Password: rootProps.StringPropertyValue("apigee.auth.password"),
		},
	}
}

// ValidateCfg - Validates the gateway config
func (a *ApigeeConfig) ValidateCfg() (err error) {
	if a.Auth.Username == "" {
		return errors.New("Invalid APIGEE configuration: username is not configured")
	}

	if a.Auth.Password == "" {
		return errors.New("Invalid APIGEE configuration: password is not configured")
	}

	return
}

// GetAuth - Returns the Auth Config
func (a *ApigeeConfig) GetAuth() *AuthConfig {
	return a.Auth
}

// GetPollInterval - Returns the Poll Interval
func (a *ApigeeConfig) GetPollInterval() time.Duration {
	return a.PollInterval
}
