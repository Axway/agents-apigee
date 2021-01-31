package apigee

import "github.com/Axway/agents-apigee/discovery/pkg/apigee/models"

// grantType values
type grantType int

const (
	password grantType = iota
	refresh
)

func (g grantType) String() string {
	return [...]string{"password", "refresh_token"}[g]
}

//AuthResponse - response struct from APIGEE auth call
type AuthResponse struct {
	AccessToken  string `json:"access_token"`
	TokenType    string `json:"token_type"`
	RefreshToken string `json:"refresh_token"`
	ExpiresIn    int    `json:"expires_in"`
	Scope        string `json:"scope"`
	JTI          string `json:"jti"`
}

//Environments
type environments []string

//VirtualHosts
type virtualHosts []string

//APIs
type apis []string

//EnvironmentRevisions
type environmentRevision struct {
	Name      string
	Revisions []models.DeploymentDetailsRevision
}

//specAssociationFile -
type specAssociationFile struct {
	URL string `json:"url"`
}
