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
