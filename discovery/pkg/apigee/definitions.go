package apigee

const (
	defaultSubscriptionSchema = "apigee-subscription-schema"
	appDisplayNameKey         = "appDisplayName"
)

// grantType values
type grantType int

const (
	password grantType = iota
	refresh
)

type wgAction int

const (
	wgAdd wgAction = iota
	wgDone
)

func (g grantType) String() string {
	return [...]string{"password", "refresh_token"}[g]
}

type productRequest struct {
	name     string
	response chan map[string]string
}
