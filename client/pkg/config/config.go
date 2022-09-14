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
	Organization string           `config:"organization"`
	Auth         *AuthConfig      `config:"auth"`
	Intervals    *ApigeeIntervals `config:"intervals"`
	Filter       string           `config:"filter"`
	DeveloperID  string           `config:"developerID"`
}

// ApigeeIntervals - intervals for the apigee agent to use
type ApigeeIntervals struct {
	Product time.Duration `config:"product"`
	Portal  time.Duration `config:"portal"`
	API     time.Duration `config:"api"`
}

const (
	pathOrganization    = "apigee.organization"
	pathAuthUsername    = "apigee.auth.username"
	pathAuthPassword    = "apigee.auth.password"
	pathProductInterval = "apigee.interval.product"
	pathPortalInterval  = "apigee.interval.portal"
	pathAPIInterval     = "apigee.interval.api"
	pathFilter          = "apigee.filter"
	pathDeveloper       = "apigee.developerID"
)

// AddProperties - adds config needed for apigee client
func AddProperties(rootProps properties.Properties) {
	rootProps.AddStringProperty(pathOrganization, "", "APIGEE Organization")
	rootProps.AddStringProperty(pathAuthUsername, "", "Username to use to authenticate to APIGEE")
	rootProps.AddStringProperty(pathAuthPassword, "", "Password for the user to authenticate to APIGEE")
	rootProps.AddDurationProperty(pathProductInterval, 5*time.Minute, "The time interval between updating a products attributes")
	rootProps.AddDurationProperty(pathPortalInterval, 1*time.Minute, "The time interval between checking for new Apigee portals")
	rootProps.AddDurationProperty(pathAPIInterval, 30*time.Second, "The time interval between checking for new APIs in an Apigee portal")
	rootProps.AddStringProperty(pathFilter, "", "Filter used on discovering Apigee products")
	rootProps.AddStringProperty(pathDeveloper, "", "Developer ID used to create applications")
}

// ParseConfig - parse the config on startup
func ParseConfig(rootProps properties.Properties) *ApigeeConfig {
	return &ApigeeConfig{
		Organization: rootProps.StringPropertyValue(pathOrganization),
		Filter:       rootProps.StringPropertyValue(pathFilter),
		DeveloperID:  rootProps.StringPropertyValue(pathDeveloper),
		Intervals: &ApigeeIntervals{
			Product: rootProps.DurationPropertyValue(pathProductInterval),
			Portal:  rootProps.DurationPropertyValue(pathPortalInterval),
			API:     rootProps.DurationPropertyValue(pathAPIInterval),
		},
		Auth: &AuthConfig{
			Username: rootProps.StringPropertyValue(pathAuthUsername),
			Password: rootProps.StringPropertyValue(pathAuthPassword),
		},
	}
}

// ValidateCfg - Validates the gateway config
func (a *ApigeeConfig) ValidateCfg() (err error) {
	if a.Auth.Username == "" {
		return errors.New("invalid APIGEE configuration: username is not configured")
	}

	if a.Auth.Password == "" {
		return errors.New("invalid APIGEE configuration: password is not configured")
	}

	if a.DeveloperID == "" {
		return errors.New("invalid APIGEE configuration: developer ID must be configured")
	}

	return
}

// GetAuth - Returns the Auth Config
func (a *ApigeeConfig) GetAuth() *AuthConfig {
	return a.Auth
}

// GetPollInterval - Returns the Poll Interval
func (a *ApigeeConfig) GetIntervals() *ApigeeIntervals {
	return a.Intervals
}
