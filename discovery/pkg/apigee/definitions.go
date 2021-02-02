package apigee

import (
	"fmt"

	"github.com/Axway/agents-apigee/discovery/pkg/apigee/models"
	"github.com/Axway/agents-apigee/discovery/pkg/util"
)

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

// apigeeProxyDetails- APIGEE Proxy Details
type apigeeProxyDetails struct {
	Proxy       models.ApiProxy
	Revision    models.DeploymentDetailsRevision
	APIRevision models.ApiProxyRevision
	Environment string
	Spec        []byte
}

func (a *apigeeProxyDetails) GetVersion() string {
	return fmt.Sprintf("%d.%d", a.APIRevision.ConfigurationVersion.MajorVersion, a.APIRevision.ConfigurationVersion.MinorVersion)
}

func (a *apigeeProxyDetails) GetCacheKey() string {
	return util.FormatRemoteAPIID(a.Proxy.Name, a.Environment, a.Revision.Name)
}

//Environments
type environments []string

//VirtualHosts
type virtualHosts []string

//APIs
type apis []string

//EnvironmentRevisions
type environmentRevision struct {
	EnvironmentName string
	Revisions       []models.DeploymentDetailsRevision
}

//specAssociationFile -
type specAssociationFile struct {
	URL string `json:"url"`
}
