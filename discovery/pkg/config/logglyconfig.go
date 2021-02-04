package config

// LogglyConfig - represents the config for gateway
type LogglyConfig struct {
	Organization string `config:"organization"`
	APIToken     string `config:"token"`
	Host         string `config:"host"`
	Port         string `config:"port"`
}

// GetOrganization - Returns the Loggly Organization
func (l *LogglyConfig) GetOrganization() string {
	return l.Organization
}

// GetAPIToken - Returns the Loggly GetAPIToken
func (l *LogglyConfig) GetAPIToken() string {
	return l.APIToken
}
