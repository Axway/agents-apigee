package apigee

import (
	"fmt"

	"github.com/Axway/agents-apigee/discovery/pkg/apigee/apigeebundle"
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
	Bundle      *apigeebundle.APIGEEBundle
	Environment string
}

func (a *apigeeProxyDetails) GetVersion() string {
	return fmt.Sprintf("%v", a.APIRevision.Revision)
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

//Products
type products []string

//EnvironmentRevisions
type environmentRevision struct {
	EnvironmentName string
	Revisions       []models.DeploymentDetailsRevision
}

//specAssociationFile -
type specAssociationFile struct {
	URL string `json:"url"`
}

// portalResponse
type portalResponse struct {
	Status    string       `json:"status"`
	Message   string       `json:"message"`
	Code      string       `json:"code"`
	ErrorCode string       `json:"error_code"`
	RequestID string       `json:"request_id"`
	Data      []portalData `json:"data"`
}

type portalData struct {
	ID                   string `json:"id"`
	Name                 string `json:"name"`
	Description          string `json:"description"`
	CustomDomain         string `json:"customDomain"`
	OrgName              string `json:"orgName"`
	Status               string `json:"status"`
	VisibleToCustomers   bool   `json:"visibleToCustomers"`
	HTTPS                bool   `json:"https"`
	DefaultDomain        string `json:"defaultDomain"`
	CustomeDomainEnabled bool   `json:"customDomainEnabled"`
	DefaultURL           string `json:"defaultURL"`
	CurrentURL           string `json:"currentURL"`
	CurrentDomain        string `json:"currentDomain"`
}

// apiDocDataResponse
type apiDocDataResponse struct {
	Status    string       `json:"status"`
	Message   string       `json:"message"`
	Code      string       `json:"code"`
	ErrorCode string       `json:"error_code"`
	RequestID string       `json:"request_id"`
	Data      []apiDocData `json:"data"`
}

type apiDocData struct {
	ID            string `json:"id"`
	PortalID      string `json:"siteId"`
	Title         string `json:"title"`
	Description   string `json:"description"`
	APIID         string `json:"apiId"`
	ProductName   string `json:"edgeAPIProductName"`
	SpecContent   string `json:"specContent"`
	SpecTitle     string `json:"specTitle"`
	SpecID        string `json:"specId"`
	ProductExists bool   `json:"productExists"`
}
