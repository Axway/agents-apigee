package apigee

const (
	openapi     = "openapi"
	association = "association.json"
)

type Association struct {
	URL string `json:"url"`
}

type JobFirstRunDone func() bool

const (
	quotaPolicy  = "Quota"
	apiKeyPolicy = "VerifyAPIKey"
	oauthPolicy  = "Oauthv2"
)

const (
	cacheKeyAttribute = "cacheKey"
)
