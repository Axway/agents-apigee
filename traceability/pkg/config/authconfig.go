package config

// AuthConfig - represents the config for gateway
type AuthConfig struct {
	Username string `config:"username"`
	Password string `config:"password"`
}

// GetUsername - Returns the APIGEE username
func (a *AuthConfig) GetUsername() string {
	return a.Username
}

// GetPassword - Returns the APIGEE password
func (a *AuthConfig) GetPassword() string {
	return a.Password
}
