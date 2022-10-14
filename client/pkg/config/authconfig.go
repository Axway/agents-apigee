package config

// AuthConfig - represents the config for gateway
type AuthConfig struct {
	URL            string `config:"url"`
	ServerUsername string `config:"serverUsername"`
	ServerPassword string `config:"serverPassword"`
	Username       string `config:"username"`
	Password       string `config:"password"`
}

// GetServerUsername - Returns the APIGEE auth server username
func (a *AuthConfig) GetServerUsername() string {
	return a.ServerUsername
}

// GetServerPassword - Returns the APIGEE auth server password
func (a *AuthConfig) GetServerPassword() string {
	return a.ServerPassword
}

// GetURL - Returns the APIGEE username
func (a *AuthConfig) GetURL() string {
	return a.URL
}

// GetUsername - Returns the APIGEE username
func (a *AuthConfig) GetUsername() string {
	return a.Username
}

// GetPassword - Returns the APIGEE password
func (a *AuthConfig) GetPassword() string {
	return a.Password
}
