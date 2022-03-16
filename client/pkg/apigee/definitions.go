package apigee

import (
	"github.com/Axway/agents-apigee/client/pkg/apigee/models"
)

const (
	defaultSubscriptionSchema = "apigee-subscription-schema"
	appNameKey                = "appName"
)

// grantType values
type grantType int

const (
	password grantType = iota
	refresh
)

var ApigeeAgentAttribute = models.Attribute{
	Name:  "createdBy",
	Value: "apigee-agent",
}

func (g grantType) String() string {
	return [...]string{"password", "refresh_token"}[g]
}

// AuthResponse - response struct from APIGEE auth call
type AuthResponse struct {
	AccessToken  string `json:"access_token"`
	TokenType    string `json:"token_type"`
	RefreshToken string `json:"refresh_token"`
	ExpiresIn    int    `json:"expires_in"`
	Scope        string `json:"scope"`
	JTI          string `json:"jti"`
}

// Products
type Products []string

// PortalResponse
type PortalResponse struct {
	Status    string     `json:"status"`
	Message   string     `json:"message"`
	Code      string     `json:"code"`
	ErrorCode string     `json:"error_code"`
	RequestID string     `json:"request_id"`
	Data      PortalData `json:"data"`
}

// PortalsResponse
type PortalsResponse struct {
	Status    string       `json:"status"`
	Message   string       `json:"message"`
	Code      string       `json:"code"`
	ErrorCode string       `json:"error_code"`
	RequestID string       `json:"request_id"`
	Data      []PortalData `json:"data"`
}

type PortalData struct {
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

// CredentialProvisionRequest represents the request body needed to add an api product to a credential in an app.
type CredentialProvisionRequest struct {
	ApiProducts []string           `json:"apiProducts"`
	Attributes  []models.Attribute `json:"attributes,omitempty"`
}
