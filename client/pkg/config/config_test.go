package config

import (
	"testing"
	"time"

	"github.com/Axway/agent-sdk/pkg/cmd/properties"
	"github.com/stretchr/testify/assert"
)

func TestValidateConfig(t *testing.T) {
	cfg := &ApigeeConfig{}

	err := cfg.ValidateCfg()
	assert.NotNil(t, err)
	assert.Equal(t, "invalid APIGEE configuration: discoveryMode must be proxy or product", err.Error())
	cfg.mode = discoveryModeProxy

	err = cfg.ValidateCfg()
	assert.NotNil(t, err)
	assert.Equal(t, "invalid APIGEE configuration: url is not configured", err.Error())
	cfg.URL = "http://host.com"

	err = cfg.ValidateCfg()
	assert.NotNil(t, err)
	assert.Equal(t, "invalid APIGEE configuration: api version is not configured", err.Error())
	cfg.APIVersion = "v1"

	err = cfg.ValidateCfg()
	assert.NotNil(t, err)
	assert.Equal(t, "invalid APIGEE configuration: data url is not configured", err.Error())
	cfg.DataURL = "http://host.com"

	err = cfg.ValidateCfg()
	assert.NotNil(t, err)
	assert.Equal(t, "invalid APIGEE configuration: username is not configured", err.Error())
	cfg.Auth = &AuthConfig{Username: "username"}

	err = cfg.ValidateCfg()
	assert.NotNil(t, err)
	assert.Equal(t, "invalid APIGEE configuration: password is not configured", err.Error())
	cfg.Auth.Password = "password"

	err = cfg.ValidateCfg()
	assert.NotNil(t, err)
	assert.Equal(t, "invalid APIGEE configuration: developer ID is not configured", err.Error())
	cfg.DeveloperID = "id@dev.com"

	err = cfg.ValidateCfg()
	assert.NotNil(t, err)
	assert.Equal(t, "invalid APIGEE configuration: proxy workers must be greater than 0", err.Error())
	cfg.Workers = &ApigeeWorkers{Proxy: 1}

	err = cfg.ValidateCfg()
	assert.NotNil(t, err)
	assert.Equal(t, "invalid APIGEE configuration: spec workers must be greater than 0", err.Error())
	cfg.Workers.Spec = 1

	err = cfg.ValidateCfg()
	assert.Nil(t, err)
}

type propData struct {
	pType string
	desc  string
	val   interface{}
	opts  []properties.DurationOpt
}

type fakeProps struct {
	props map[string]propData
}

func (f *fakeProps) AddStringProperty(name string, defaultVal string, description string) {
	f.props[name] = propData{"string", description, defaultVal, []properties.DurationOpt{}}
}

func (f *fakeProps) AddIntProperty(name string, defaultVal int, description string) {
	f.props[name] = propData{"int", description, defaultVal, []properties.DurationOpt{}}
}

func (f *fakeProps) AddBoolProperty(name string, defaultVal bool, description string) {
	f.props[name] = propData{"bool", description, defaultVal, []properties.DurationOpt{}}
}

func (f *fakeProps) AddDurationProperty(name string, defaultVal time.Duration, description string, opts ...properties.DurationOpt) {
	f.props[name] = propData{"duration", description, defaultVal, opts}
}

func (f *fakeProps) StringPropertyValue(name string) string {
	if prop, ok := f.props[name]; ok {
		return prop.val.(string)
	}
	return ""
}

func (f *fakeProps) IntPropertyValue(name string) int {
	if prop, ok := f.props[name]; ok {
		return prop.val.(int)
	}
	return 0
}

func (f *fakeProps) BoolPropertyValue(name string) bool {
	if prop, ok := f.props[name]; ok {
		return prop.val.(bool)
	}
	return false
}

func (f *fakeProps) DurationPropertyValue(name string) time.Duration {
	if prop, ok := f.props[name]; ok {
		return prop.val.(time.Duration)
	}
	return 0
}

func TestApigeeProperties(t *testing.T) {
	newProps := &fakeProps{props: map[string]propData{}}

	// validate adding props
	AddProperties(newProps)
	assert.Contains(t, newProps.props, pathURL)
	assert.Contains(t, newProps.props, pathDataURL)
	assert.Contains(t, newProps.props, pathAPIVersion)
	assert.Contains(t, newProps.props, pathOrganization)
	assert.Contains(t, newProps.props, pathMode)
	assert.Contains(t, newProps.props, pathFilter)
	assert.Contains(t, newProps.props, pathCloneAttributes)
	assert.Contains(t, newProps.props, pathAllTraffic)
	assert.Contains(t, newProps.props, pathNotSetTraffic)
	assert.Contains(t, newProps.props, pathAuthURL)
	assert.Contains(t, newProps.props, pathAuthServerUsername)
	assert.Contains(t, newProps.props, pathAuthServerPassword)
	assert.Contains(t, newProps.props, pathAuthUsername)
	assert.Contains(t, newProps.props, pathAuthPassword)
	assert.Contains(t, newProps.props, pathSpecInterval)
	assert.Contains(t, newProps.props, pathProxyInterval)
	assert.Contains(t, newProps.props, pathProductInterval)
	assert.Contains(t, newProps.props, pathStatsInterval)
	assert.Contains(t, newProps.props, pathDeveloper)
	assert.Contains(t, newProps.props, pathSpecWorkers)
	assert.Contains(t, newProps.props, pathProxyWorkers)
	assert.Contains(t, newProps.props, pathProductWorkers)

	// validate defaults
	cfg := ParseConfig(newProps)
	assert.Equal(t, "proxy", cfg.mode.String())
	assert.True(t, cfg.IsProxyMode())
	assert.False(t, cfg.IsProductMode())
	assert.Equal(t, "", cfg.Organization)
	assert.Equal(t, "https://api.enterprise.apigee.com", cfg.URL)
	assert.Equal(t, "v1", cfg.APIVersion)
	assert.Equal(t, "", cfg.Filter)
	assert.Equal(t, "https://apigee.com/dapi/api", cfg.DataURL)
	assert.Equal(t, "https://login.apigee.com", cfg.GetAuth().GetURL())
	assert.Equal(t, "edgecli", cfg.GetAuth().GetServerUsername())
	assert.Equal(t, "edgeclisecret", cfg.GetAuth().GetServerPassword())
	assert.Equal(t, "", cfg.GetAuth().GetUsername())
	assert.Equal(t, "", cfg.GetAuth().GetPassword())
	assert.Equal(t, false, cfg.GetAuth().UseBasicAuth())
	assert.Equal(t, false, cfg.ShouldCloneAttributes())
	assert.Equal(t, false, cfg.ShouldReportAllTraffic())
	assert.Equal(t, false, cfg.ShouldReportNotSetTraffic())
	assert.Equal(t, 30*time.Minute, cfg.GetIntervals().Spec)
	assert.Equal(t, 30*time.Second, cfg.GetIntervals().Proxy)
	assert.Equal(t, 30*time.Second, cfg.GetIntervals().Product)
	assert.Equal(t, 15*time.Minute, cfg.GetIntervals().Stats)
	assert.Equal(t, "", cfg.DeveloperID)
	assert.Equal(t, 10, cfg.GetWorkers().Proxy)
	assert.Equal(t, 20, cfg.GetWorkers().Spec)
	assert.Equal(t, 10, cfg.GetWorkers().Product)
}
