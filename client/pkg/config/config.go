package config

import (
	"errors"
	"strings"
	"time"

	"github.com/Axway/agent-sdk/pkg/cmd/properties"
	corecfg "github.com/Axway/agent-sdk/pkg/config"
)

type props interface {
	AddStringProperty(name string, defaultVal string, description string)
	AddIntProperty(name string, defaultVal int, description string)
	AddBoolProperty(name string, defaultVal bool, description string)
	AddDurationProperty(name string, defaultVal time.Duration, description string, opts ...properties.DurationOpt)
	StringPropertyValue(name string) string
	IntPropertyValue(name string) int
	BoolPropertyValue(name string) bool
	DurationPropertyValue(name string) time.Duration
}

func NewApigeeConfig() *ApigeeConfig {
	return &ApigeeConfig{
		Auth:      &AuthConfig{},
		Intervals: &ApigeeIntervals{},
		Workers:   &ApigeeWorkers{},
		Specs:     &ApigeeSpecConfig{},
	}
}

// ApigeeConfig - represents the config for gateway
type ApigeeConfig struct {
	corecfg.IConfigValidator
	Organization    string            `config:"organization"`
	URL             string            `config:"url"`
	DataURL         string            `config:"dataURL"`
	APIVersion      string            `config:"apiVersion"`
	Filter          string            `config:"filter"`
	DeveloperID     string            `config:"developerID"`
	Auth            *AuthConfig       `config:"auth"`
	Intervals       *ApigeeIntervals  `config:"interval"`
	Workers         *ApigeeWorkers    `config:"workers"`
	Specs           *ApigeeSpecConfig `config:"specs"`
	CloneAttributes bool              `config:"cloneAttributes"`
	AllTraffic      bool              `config:"allTraffic"`
	NotSetTraffic   bool              `config:"notSetTraffic"`
	mode            discoveryMode
}

// ApigeeWorkers - number of workers for the apigee agent to use
type ApigeeSpecConfig struct {
	DisablePollForSpecs bool   `config:"disablePollForSpecs"`
	Unstructured        bool   `config:"unstructured"`
	LocalPath           string `config:"localDirectory"`
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
		discoveryModeProxy:   discoveryModeProxyString,
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
	pathURL                     = "apigee.url"
	pathDataURL                 = "apigee.dataURL"
	pathAPIVersion              = "apigee.apiVersion"
	pathOrganization            = "apigee.organization"
	pathMode                    = "apigee.discoveryMode"
	pathFilter                  = "apigee.filter"
	pathCloneAttributes         = "apigee.cloneAttributes"
	pathAllTraffic              = "apigee.allTraffic"
	pathNotSetTraffic           = "apigee.notSetTraffic"
	pathAuthURL                 = "apigee.auth.url"
	pathAuthServerUsername      = "apigee.auth.serverUsername"
	pathAuthServerPassword      = "apigee.auth.serverPassword"
	pathAuthUsername            = "apigee.auth.username"
	pathAuthPassword            = "apigee.auth.password"
	pathAuthBasicAuth           = "apigee.auth.useBasicAuth"
	pathSpecInterval            = "apigee.interval.spec"
	pathProxyInterval           = "apigee.interval.proxy"
	pathProductInterval         = "apigee.interval.product"
	pathStatsInterval           = "apigee.interval.stats"
	pathDeveloper               = "apigee.developerID"
	pathSpecWorkers             = "apigee.workers.spec"
	pathProxyWorkers            = "apigee.workers.proxy"
	pathProductWorkers          = "apigee.workers.product"
	pathSpecLocalPath           = "apigee.specConfig.localPath"
	pathSpecUnstructured        = "apigee.specConfig.unstructured"
	pathSpecDisablePollForSpecs = "apigee.specConfig.disablePollForSpecs"
)

// AddProperties - adds config needed for apigee client
func AddProperties(rootProps props) {
	rootProps.AddStringProperty(pathMode, "proxy", "APIGEE Organization")
	rootProps.AddStringProperty(pathOrganization, "", "APIGEE Organization")
	rootProps.AddStringProperty(pathURL, "https://api.enterprise.apigee.com", "APIGEE Base URL")
	rootProps.AddStringProperty(pathAPIVersion, "v1", "APIGEE API Version")
	rootProps.AddStringProperty(pathFilter, "", "Filter used on discovering Apigee products")
	rootProps.AddStringProperty(pathDataURL, "https://apigee.com/dapi/api", "APIGEE Data API URL")
	rootProps.AddStringProperty(pathAuthURL, "https://login.apigee.com", "URL to use when authenticating to APIGEE")
	rootProps.AddStringProperty(pathAuthServerUsername, "edgecli", "Username to use to when requesting APIGEE token")
	rootProps.AddStringProperty(pathAuthServerPassword, "edgeclisecret", "Password to use to when requesting APIGEE token")
	rootProps.AddStringProperty(pathAuthUsername, "", "Username to use to authenticate to APIGEE")
	rootProps.AddStringProperty(pathAuthPassword, "", "Password for the user to authenticate to APIGEE")
	rootProps.AddBoolProperty(pathAuthBasicAuth, false, "Set to true to use basic authentication to authenticate to APIGEE")
	rootProps.AddBoolProperty(pathCloneAttributes, false, "Set to true to copy the tags when provisioning a Product in product mode")
	rootProps.AddBoolProperty(pathAllTraffic, false, "Set to true to report metrics for all traffic for the selected mode")
	rootProps.AddBoolProperty(pathNotSetTraffic, false, "Set to true to report metrics for values reported with (not set) ast the name")
	rootProps.AddDurationProperty(pathSpecInterval, 30*time.Minute, "The time interval between checking for updated specs", properties.WithLowerLimit(1*time.Minute))
	rootProps.AddDurationProperty(pathProxyInterval, 30*time.Second, "The time interval between checking for updated proxies", properties.WithUpperLimit(5*time.Minute))
	rootProps.AddDurationProperty(pathProductInterval, 30*time.Second, "The time interval between checking for updated products", properties.WithUpperLimit(5*time.Minute))
	rootProps.AddDurationProperty(pathStatsInterval, 5*time.Minute, "The time interval between checking for updated stats", properties.WithLowerLimit(1*time.Minute), properties.WithUpperLimit(15*time.Minute))
	rootProps.AddStringProperty(pathDeveloper, "", "Developer ID used to create applications")
	rootProps.AddIntProperty(pathProxyWorkers, 10, "Max number of workers discovering proxies")
	rootProps.AddIntProperty(pathSpecWorkers, 20, "Max number of workers discovering specs")
	rootProps.AddIntProperty(pathProductWorkers, 10, "Max number of workers discovering products")
	rootProps.AddStringProperty(pathSpecLocalPath, "", "Path to a local directory that contains the spec files")
	rootProps.AddBoolProperty(pathSpecUnstructured, false, "Set to true to enable discovering apis that have no associated spec")
	rootProps.AddBoolProperty(pathSpecDisablePollForSpecs, false, "Set to true to disable polling apigee for specs, rely on the local directory or spec URLs")
}

// ParseConfig - parse the config on startup
func ParseConfig(rootProps props) *ApigeeConfig {
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
		NotSetTraffic:   rootProps.BoolPropertyValue(pathNotSetTraffic),
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
			BasicAuth:      rootProps.BoolPropertyValue(pathAuthBasicAuth),
		},
		Specs: &ApigeeSpecConfig{
			LocalPath:           rootProps.StringPropertyValue(pathSpecLocalPath),
			DisablePollForSpecs: rootProps.BoolPropertyValue(pathSpecDisablePollForSpecs),
			Unstructured:        rootProps.BoolPropertyValue(pathSpecUnstructured),
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

	if a.Auth == nil || a.Auth.Username == "" {
		return errors.New("invalid APIGEE configuration: username is not configured")
	}

	if a.Auth == nil || a.Auth.Password == "" {
		return errors.New("invalid APIGEE configuration: password is not configured")
	}

	if a.DeveloperID == "" {
		return errors.New("invalid APIGEE configuration: developer ID is not configured")
	}

	if a.Workers == nil || a.Workers.Proxy < 1 {
		return errors.New("invalid APIGEE configuration: proxy workers must be greater than 0")
	}

	if a.Workers == nil || a.Workers.Spec < 1 {
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

func (a *ApigeeConfig) ShouldReportNotSetTraffic() bool {
	return a.NotSetTraffic
}
