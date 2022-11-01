package apigee

const (
	openapi     = "openapi"
	association = "association.json"
)

type Association struct {
	URL string `json:"url"`
}

type jobFirstRunDone func() bool

const (
	quotaPolicy  = "Quota"
	apiKeyPolicy = "VerifyAPIKey"
	oauthPolicy  = "OAuthV2"
)

const (
	cacheKeyAttribute    = "cacheKey"
	agentProductTagName  = "AgentCreated"
	agentProductTagValue = "true"
)
