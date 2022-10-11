package config

import (
	"errors"
	"strings"
	"time"

	"github.com/Axway/agent-sdk/pkg/cmd/properties"
	corecfg "github.com/Axway/agent-sdk/pkg/config"
)

// ApigeeConfig - represents the config for gateway
type ApigeeConfig struct {
	corecfg.IConfigValidator
	Organization string           `config:"organization"`
	URL          string           `config:"url"`
	DataURL      string           `config:"dataURL"`
	APIVersion   string           `config:"apiVersion"`
	Auth         *AuthConfig      `config:"auth"`
	Intervals    *ApigeeIntervals `config:"intervals"`
	Filter       string           `config:"filter"`
	DeveloperID  string           `config:"developerID"`
}

// ApigeeIntervals - intervals for the apigee agent to use
type ApigeeIntervals struct {
	Proxy time.Duration `config:"proxy"`
	Spec  time.Duration `config:"spec"`
}

const (
	pathURL                = "apigee.url"
	pathDataURL            = "apigee.dataURL"
	pathAPIVersion         = "apigee.apiVersion"
	pathOrganization       = "apigee.organization"
	pathAuthURL            = "apigee.auth.url"
	pathAuthServerUsername = "apigee.auth.serverUsername"
	pathAuthServerPassword = "apigee.auth.serverPassword"
	pathAuthUsername       = "apigee.auth.username"
	pathAuthPassword       = "apigee.auth.password"
	pathSpecInterval       = "apigee.interval.spec"
	pathProxyInterval      = "apigee.interval.proxy"
	pathDeveloper          = "apigee.developerID"
)

// AddProperties - adds config needed for apigee client
func AddProperties(rootProps properties.Properties) {
	rootProps.AddStringProperty(pathOrganization, "", "APIGEE Organization")
	rootProps.AddStringProperty(pathURL, "https://api.enterprise.apigee.com", "APIGEE Base URL")
	rootProps.AddStringProperty(pathAPIVersion, "v1", "APIGEE API Version")
	rootProps.AddStringProperty(pathDataURL, "https://apigee.com/dapi/api", "APIGEE Data API URL")
	rootProps.AddStringProperty(pathAuthURL, "https://login.apigee.com", "URL to use when authenticting to APIGEE")
	rootProps.AddStringProperty(pathAuthServerUsername, "edgecli", "Username to use to when requesting APIGEE token")
	rootProps.AddStringProperty(pathAuthServerPassword, "edgeclisecret", "Password to use to when requesting APIGEE token")
	rootProps.AddStringProperty(pathAuthUsername, "", "Username to use to authenticate to APIGEE")
	rootProps.AddStringProperty(pathAuthPassword, "", "Password for the user to authenticate to APIGEE")
	rootProps.AddDurationProperty(pathSpecInterval, 30*time.Minute, "The time interval between checking for updated specs")
	rootProps.AddDurationProperty(pathProxyInterval, 30*time.Second, "The time interval between checking for updated proxies")
	rootProps.AddStringProperty(pathDeveloper, "", "Developer ID used to create applications")
}

// ParseConfig - parse the config on startup
func ParseConfig(rootProps properties.Properties) *ApigeeConfig {

	return &ApigeeConfig{
		Organization: rootProps.StringPropertyValue(pathOrganization),
		URL:          strings.TrimSuffix(rootProps.StringPropertyValue(pathURL), "/"),
		APIVersion:   rootProps.StringPropertyValue(pathAPIVersion),
		DataURL:      strings.TrimSuffix(rootProps.StringPropertyValue(pathDataURL), "/"),
		DeveloperID:  rootProps.StringPropertyValue(pathDeveloper),
		Intervals: &ApigeeIntervals{
			Proxy: rootProps.DurationPropertyValue(pathProxyInterval),
			Spec:  rootProps.DurationPropertyValue(pathSpecInterval),
		},
		Auth: &AuthConfig{
			Username:       rootProps.StringPropertyValue(pathAuthUsername),
			Password:       rootProps.StringPropertyValue(pathAuthPassword),
			ServerUsername: rootProps.StringPropertyValue(pathAuthServerUsername),
			ServerPassword: rootProps.StringPropertyValue(pathAuthServerPassword),
			URL:            rootProps.StringPropertyValue(pathAuthURL),
		},
	}
}

// ValidateCfg - Validates the gateway config
func (a *ApigeeConfig) ValidateCfg() (err error) {
	if a.URL == "" {
		return errors.New("invalid APIGEE configuration: url is not configured")
	}

	if a.APIVersion == "" {
		return errors.New("invalid APIGEE configuration: api version is not configured")
	}

	if a.DataURL == "" {
		return errors.New("invalid APIGEE configuration: data url is not configured")
	}

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
