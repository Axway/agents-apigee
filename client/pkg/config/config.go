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
	Filter       string        `config:"filter"`
}

const (
	pathOrganization = "apigee.organization"
	pathAuthUsername = "apigee.auth.username"
	pathAuthPassword = "apigee.auth.password"
	pathInterval     = "apigee.pollInterval"
	pathFilter       = "apigee.filter"
)

// AddProperties - adds config needed for apigee client
func AddProperties(rootProps properties.Properties) {
	rootProps.AddStringProperty(pathOrganization, "", "APIGEE Organization")
	rootProps.AddStringProperty(pathAuthUsername, "", "Username to use to authenticate to APIGEE")
	rootProps.AddStringProperty(pathAuthPassword, "", "Password for the user to authenticate to APIGEE")
	rootProps.AddDurationProperty(pathInterval, 30*time.Second, "The time interval between checking for new APIGEE resources")
	rootProps.AddStringProperty(pathFilter, "", "Filter used on discovering Apigee products")
}

// ParseConfig - parse the config on startup
func ParseConfig(rootProps properties.Properties) *ApigeeConfig {
	return &ApigeeConfig{
		Organization: rootProps.StringPropertyValue(pathOrganization),
		PollInterval: rootProps.DurationPropertyValue(pathInterval),
		Filter:       rootProps.StringPropertyValue(pathFilter),
		Auth: &AuthConfig{
			Username: rootProps.StringPropertyValue(pathAuthUsername),
			Password: rootProps.StringPropertyValue(pathAuthPassword),
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
