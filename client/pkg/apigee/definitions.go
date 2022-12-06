package apigee

import (
	"github.com/Axway/agents-apigee/client/pkg/apigee/models"
)

// grantType values
type grantType int

const (
	password grantType = iota
	refresh
)

const (
	ClonedProdAttribute = "ClonedProduct"
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
	// The number of milliseconds the key will live
	KeyExpiresIn int `json:"keyExpiresIn,omitempty"`
}

type SpecDetails struct {
	ID          string        `json:"id"`
	Kind        string        `json:"kind"`
	Name        string        `json:"name"`
	Created     string        `json:"created"`
	Creator     string        `json:"creator"`
	Modified    string        `json:"modified"`
	IsTrashed   bool          `json:"isTrashed"`
	Permissions *string       `json:"permissions"`
	SelfLink    string        `json:"self"`
	ContentLink string        `json:"content"`
	Contents    []SpecDetails `json:"contents"`
	FolderLink  string        `json:"folder"`
	FolderID    string        `json:"folderId"`
	Body        *string       `json:"body"`
}

// VirtualHosts
type VirtualHosts []string

type PolicyDetail struct {
	models.Policy
	PolicyType string `json:"policyType"`
}
