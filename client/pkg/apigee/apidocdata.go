package apigee

import "github.com/Axway/agent-sdk/pkg/apic"

// APIDocDataResponse -
type APIDocDataResponse struct {
	Status    string        `json:"status"`
	Message   string        `json:"message"`
	Code      string        `json:"code"`
	ErrorCode string        `json:"error_code"`
	RequestID string        `json:"request_id"`
	Data      []*APIDocData `json:"data"`
}

// APIDocData - the data returned from teh call to get portal apis
type APIDocData struct {
	ID               int     `json:"id"`
	PortalID         string  `json:"siteId"`
	Title            string  `json:"title"`
	Description      string  `json:"description"`
	APIID            string  `json:"apiId"`
	ProductName      string  `json:"edgeAPIProductName"`
	SpecContent      string  `json:"specContent"`
	SpecTitle        string  `json:"specTitle"`
	SpecID           string  `json:"specId"`
	ProductExists    bool    `json:"productExists"`
	Modified         int     `json:"modified"`
	SnapshotModified int     `json:"snapshotModified"`
	ImageURL         *string `json:"imageUrl"`
	CategoryIds      []int   `json:"categoryIds"`
	Visibility       bool    `json:"visibility"`
	portalTitle      string
	securityPolicies []string
	apiKeyInfo       []apic.APIKeyInfo
	oauth            bool
	apiKey           bool
}

// SetAPIKeyInfo - set the api key info needed to access this api
func (a *APIDocData) SetAPIKeyInfo(keyInfo []apic.APIKeyInfo) {
	a.apiKeyInfo = keyInfo
}

// GetAPIKeyInfo - get the api key info needed to access this api
func (a *APIDocData) GetAPIKeyInfo() []apic.APIKeyInfo {
	return a.apiKeyInfo
}

// SetPortalTitle - set the portal title in the api doc data
func (a *APIDocData) SetPortalTitle(title string) {
	a.portalTitle = title
}

// GetPortalTitle - get the portal title of the api doc data
func (a *APIDocData) GetPortalTitle() string {
	return a.portalTitle
}

// SetSecurityPolicies - set the security policies in the api doc data
func (a *APIDocData) SetSecurityPolicies(policies []string) {
	for _, policy := range policies {
		a.AddSecurityPolicies(policy)
	}
}

// AddSecurityPolicies - add the security policy to the security policies slice in the api doc data
func (a *APIDocData) AddSecurityPolicies(policy string) {
	if len(a.securityPolicies) == 0 {
		a.securityPolicies = make([]string, 0)
	}
	a.securityPolicies = append(a.securityPolicies, policy)

	// set what security the spec supports
	if policy == apic.Oauth {
		a.oauth = true
	} else if policy == apic.Apikey {
		a.apiKey = true
	}
}

// GetSecurityPolicies - get the security policies of the api doc data
func (a *APIDocData) GetSecurityPolicies() []string {
	if len(a.securityPolicies) == 0 {
		return []string{}
	}
	return a.securityPolicies
}

// HasOauth - returns if the api has oauth authentication
func (a *APIDocData) HasOauth() bool {
	return a.oauth
}

// HasAPIKey - returns if the api has api key authentication
func (a *APIDocData) HasAPIKey() bool {
	return a.apiKey
}
