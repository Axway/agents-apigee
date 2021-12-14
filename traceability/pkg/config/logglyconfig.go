package config

import "github.com/Axway/agent-sdk/pkg/cmd/properties"

// LogglyConfig - represents the config for gateway
type LogglyConfig struct {
	Subdomain     string `config:"subdomain"`
	CustomerToken string `config:"customerToken"`
	APIToken      string `config:"apiToken"`
	Host          string `config:"host"`
	Port          string `config:"port"`
}

// AddProperties - add the loggly properties
func AddProperties(rootProps properties.Properties) {
	rootProps.AddStringProperty("apigee.loggly.customertoken", "", "The Loggly Customer Token for sending log events")
	rootProps.AddStringProperty("apigee.loggly.apitoken", "", "The Loggly API Token for retrieving log events")
	rootProps.AddStringProperty("apigee.loggly.subdomain", "", "The Loggly subdomain")
	rootProps.AddStringProperty("apigee.loggly.host", "logs-01.loggly.com", "The Loggly Host URL")
	rootProps.AddStringProperty("apigee.loggly.port", "514", "The Loggly Port")
}

// ParseConfig - parse the loggly config from the agent
func ParseConfig(rootProps properties.Properties) *LogglyConfig {
	return &LogglyConfig{
		Subdomain:     rootProps.StringPropertyValue("apigee.loggly.subdomain"),
		CustomerToken: rootProps.StringPropertyValue("apigee.loggly.customertoken"),
		APIToken:      rootProps.StringPropertyValue("apigee.loggly.apitoken"),
		Host:          rootProps.StringPropertyValue("apigee.loggly.host"),
		Port:          rootProps.StringPropertyValue("apigee.loggly.port"),
	}
}

// GetSubdomain - Returns the Loggly Organization
func (l *LogglyConfig) GetSubdomain() string {
	return l.Subdomain
}

// GetAPIToken - Returns the Loggly GetAPIToken
func (l *LogglyConfig) GetAPIToken() string {
	return l.APIToken
}
