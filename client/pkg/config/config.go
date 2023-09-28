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
	Organization    string           `config:"organization"`
	URL             string           `config:"url"`
	DataURL         string           `config:"dataURL"`
	APIVersion      string           `config:"apiVersion"`
	Filter          string           `config:"filter"`
	DeveloperID     string           `config:"developerID"`
	Auth            *AuthConfig      `config:"auth"`
	Intervals       *ApigeeIntervals `config:"interval"`
	Workers         *ApigeeWorkers   `config:"workers"`
	CloneAttributes bool             `config:"cloneAttributes"`
	AllTraffic      bool             `config:"allTraffic"`
	mode            discoveryMode
}

// ApigeeIntervals - intervals for the apigee agent to use
type ApigeeIntervals struct {
	Proxy   time.Duration `config:"proxy"`
	Spec    time.Duration `config:"spec"`
	Product time.Duration `config:"product"`
	Stats   time.Duration `config:"stats"`
}

// ApigeeWorkers - number of workers for the apigee agent to use
type ApigeeWorkers struct {
	Proxy   int `config:"proxy"`
	Spec    int `config:"spec"`
	Product int `config:"product"`
}

type discoveryMode int

const (
	discoveryModeProxy = iota + 1
	discoveryModeProduct
)

const (
	discoveryModeProxyString   = "proxy"
	discoveryModeProductString = "product"
)

func (m discoveryMode) String() string {
	return map[discoveryMode]string{
		discoveryModeProxy:   discoveryModeProductString,
		discoveryModeProduct: discoveryModeProductString,
	}[m]
}

func stringToDiscoveryMode(s string) discoveryMode {
	if mode, ok := map[string]discoveryMode{
		discoveryModeProxyString:   discoveryModeProxy,
		discoveryModeProductString: discoveryModeProduct,
	}[strings.ToLower(s)]; ok {
		return mode
	}
	return 0
}

const (
	pathURL                = "apigee.url"
	pathDataURL            = "apigee.dataURL"
	pathAPIVersion         = "apigee.apiVersion"
	pathOrganization       = "apigee.organization"
	pathMode               = "apigee.discoveryMode"
	pathFilter             = "apigee.filter"
	pathCloneAttributes    = "apigee.cloneAttributes"
	pathAllTraffic         = "apigee.allTraffic"
	pathAuthURL            = "apigee.auth.url"
	pathAuthServerUsername = "apigee.auth.serverUsername"
	pathAuthServerPassword = "apigee.auth.serverPassword"
	pathAuthUsername       = "apigee.auth.username"
	pathAuthPassword       = "apigee.auth.password"
	pathSpecInterval       = "apigee.interval.spec"
	pathProxyInterval      = "apigee.interval.proxy"
	pathProductInterval    = "apigee.interval.product"
	pathStatsInterval      = "apigee.interval.stats"
	pathDeveloper          = "apigee.developerID"
	pathSpecWorkers        = "apigee.workers.spec"
	pathProxyWorkers       = "apigee.workers.proxy"
	pathProductWorkers     = "apigee.workers.product"
)

// AddProperties - adds config needed for apigee client
func AddProperties(rootProps properties.Properties) {
	rootProps.AddStringProperty(pathMode, "proxy", "APIGEE Organization")
	rootProps.AddStringProperty(pathOrganization, "", "APIGEE Organization")
	rootProps.AddStringProperty(pathURL, "https://api.enterprise.apigee.com", "APIGEE Base URL")
	rootProps.AddStringProperty(pathAPIVersion, "v1", "APIGEE API Version")
	rootProps.AddStringProperty(pathFilter, "", "Filter used on discovering Apigee products")
	rootProps.AddStringProperty(pathDataURL, "https://apigee.com/dapi/api", "APIGEE Data API URL")
	rootProps.AddStringProperty(pathAuthURL, "https://login.apigee.com", "URL to use when authenticating to APIGEE")
	rootProps.AddStringProperty(pathAuthServerUsername, "edgecli", "Username to use to when requesting APIGEE token")
	rootProps.AddStringProperty(pathAuthServerPassword, "edgeclisecret", "Password to use to when requesting APIGEE token")
	rootProps.AddBoolProperty(pathCloneAttributes, false, "Set to true to copy the tags when provisioning a Product in product mode")
	rootProps.AddBoolProperty(pathAllTraffic, false, "Set to true report metrics for all traffic for the selected mode")
	rootProps.AddStringProperty(pathAuthUsername, "", "Username to use to authenticate to APIGEE")
	rootProps.AddStringProperty(pathAuthPassword, "", "Password for the user to authenticate to APIGEE")
	rootProps.AddDurationProperty(pathSpecInterval, 30*time.Minute, "The time interval between checking for updated specs", properties.WithLowerLimit(1*time.Minute))
	rootProps.AddDurationProperty(pathProxyInterval, 30*time.Second, "The time interval between checking for updated proxies", properties.WithUpperLimit(5*time.Minute))
	rootProps.AddDurationProperty(pathProductInterval, 30*time.Second, "The time interval between checking for updated products", properties.WithUpperLimit(5*time.Minute))
	rootProps.AddDurationProperty(pathStatsInterval, 5*time.Minute, "The time interval between checking for updated stats", properties.WithLowerLimit(1*time.Minute), properties.WithUpperLimit(15*time.Minute))
	rootProps.AddStringProperty(pathDeveloper, "", "Developer ID used to create applications")
	rootProps.AddIntProperty(pathProxyWorkers, 10, "Max number of workers discovering proxies")
	rootProps.AddIntProperty(pathSpecWorkers, 20, "Max number of workers discovering specs")
	rootProps.AddIntProperty(pathProductWorkers, 10, "Max number of workers discovering products")
}

// ParseConfig - parse the config on startup
func ParseConfig(rootProps properties.Properties) *ApigeeConfig {
	return &ApigeeConfig{
		Organization:    rootProps.StringPropertyValue(pathOrganization),
		URL:             strings.TrimSuffix(rootProps.StringPropertyValue(pathURL), "/"),
		APIVersion:      rootProps.StringPropertyValue(pathAPIVersion),
		DataURL:         strings.TrimSuffix(rootProps.StringPropertyValue(pathDataURL), "/"),
		DeveloperID:     rootProps.StringPropertyValue(pathDeveloper),
		mode:            stringToDiscoveryMode(rootProps.StringPropertyValue(pathMode)),
		Filter:          rootProps.StringPropertyValue(pathFilter),
		CloneAttributes: rootProps.BoolPropertyValue(pathCloneAttributes),
		AllTraffic:      rootProps.BoolPropertyValue(pathAllTraffic),
		Intervals: &ApigeeIntervals{
			Stats:   rootProps.DurationPropertyValue(pathStatsInterval),
			Proxy:   rootProps.DurationPropertyValue(pathProxyInterval),
			Spec:    rootProps.DurationPropertyValue(pathSpecInterval),
			Product: rootProps.DurationPropertyValue(pathProductInterval),
		},
		Workers: &ApigeeWorkers{
			Proxy:   rootProps.IntPropertyValue(pathProxyWorkers),
			Spec:    rootProps.IntPropertyValue(pathSpecWorkers),
			Product: rootProps.IntPropertyValue(pathProductWorkers),
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
	if a.mode == 0 {
		return errors.New("invalid APIGEE configuration: discoveryMode must be proxy or product")
	}

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

	if a.Workers.Proxy < 1 {
		return errors.New("invalid APIGEE configuration: proxy workers must be greater than 0")
	}

	if a.Workers.Spec < 1 {
		return errors.New("invalid APIGEE configuration: spec workers must be greater than 0")
	}

	return
}

// GetAuth - Returns the Auth Config
func (a *ApigeeConfig) GetAuth() *AuthConfig {
	return a.Auth
}

// GetIntervals - Returns the Intervals
func (a *ApigeeConfig) GetIntervals() *ApigeeIntervals {
	return a.Intervals
}

// GetWorkers - Returns the number of Workers
func (a *ApigeeConfig) GetWorkers() *ApigeeWorkers {
	return a.Workers
}

func (a *ApigeeConfig) IsProxyMode() bool {
	return a.mode == discoveryModeProxy
}

func (a *ApigeeConfig) IsProductMode() bool {
	return a.mode == discoveryModeProduct
}

func (a *ApigeeConfig) ShouldCloneAttributes() bool {
	return a.CloneAttributes
}

func (a *ApigeeConfig) ShouldReportAllTraffic() bool {
	return a.AllTraffic
}
