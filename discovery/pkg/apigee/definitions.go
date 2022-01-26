package apigee

import "github.com/Axway/agents-apigee/client/pkg/apigee"

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

type agentChannels struct {
	wgActionChan      chan wgAction
	newPortalChan     chan string
	removedPortalChan chan string
	removedAPIChan    chan string
	processAPIChan    chan *apigee.APIDocData
	productChan       chan productRequest
}
