package config

// LogglyConfig - represents the config for gateway
type LogglyConfig struct {
	Subdomain     string `config:"subdomain"`
	CustomerToken string `config:"customerToken"`
	APIToken      string `config:"apiToken"`
	Host          string `config:"host"`
	Port          string `config:"port"`
}

// GetSubdomain - Returns the Loggly Subdomain
func (l *LogglyConfig) GetSubdomain() string {
	return l.Subdomain
}

// GetAPIToken - Returns the Loggly GetAPIToken
func (l *LogglyConfig) GetAPIToken() string {
	return l.APIToken
}
