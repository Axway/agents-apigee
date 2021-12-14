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

func (g grantType) String() string {
	return [...]string{"password", "refresh_token"}[g]
}
