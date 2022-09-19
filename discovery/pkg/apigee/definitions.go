package apigee

const (
	defaultSubscriptionSchema = "apigee-subscription-schema"
	appDisplayNameKey         = "appDisplayName"
)

type wgAction int

const (
	wgAdd wgAction = iota
	wgDone
)

type productRequest struct {
	name     string
	response chan map[string]string
}

const (
	openapi     = "openapi"
	association = "association.json"
)

type Association struct {
	URL string `json:"url"`
}

type JobFirstRunDone func() bool
