package apigee

import (
	"time"

	coreapi "github.com/Axway/agent-sdk/pkg/api"
	"github.com/Axway/agent-sdk/pkg/jobs"

	"github.com/Axway/agents-apigee/client/pkg/config"
)

// ApigeeClient - Represents the Gateway client
type ApigeeClient struct {
	cfg          *config.ApigeeConfig
	apiClient    coreapi.Client
	accessToken  string
	developerID  string
	pollInterval time.Duration
	envToURLs    map[string][]string
	isReady      bool
}

// NewClient - Creates a new Gateway Client
func NewClient(apigeeCfg *config.ApigeeConfig) (*ApigeeClient, error) {
	client := &ApigeeClient{
		apiClient:   coreapi.NewClient(nil, ""),
		cfg:         apigeeCfg,
		envToURLs:   make(map[string][]string),
		isReady:     false,
		developerID: "",
	}

	// create the auth job and register it
	authentication := newAuthJob(
		withAPIClient(client.apiClient),
		withUsername(apigeeCfg.Auth.GetUsername()),
		withPassword(apigeeCfg.Auth.GetPassword()),
		withURL(apigeeCfg.Auth.GetURL()),
		withAuthServerUsername(apigeeCfg.Auth.GetServerUsername()),
		withAuthServerPassword(apigeeCfg.Auth.GetServerPassword()),
		withTokenSetter(client.setAccessToken),
	)
	jobs.RegisterIntervalJobWithName(authentication, 10*time.Minute, "APIGEE Auth Token")

	return client, nil
}

func (a *ApigeeClient) setAccessToken(token string) {
	a.accessToken = token
	a.isReady = true
}

//SetDeveloperID - set the developer id to be used when creating apps
func (a *ApigeeClient) SetDeveloperID(id string) {
	a.developerID = id
}

//GetDeveloperID - get the developer id to be used when creating apps
func (a *ApigeeClient) GetDeveloperID() string {
	return a.developerID
}

//GetConfig - return the apigee client config
func (a *ApigeeClient) GetConfig() *config.ApigeeConfig {
	return a.cfg
}

//IsReady - returns true when the apigee client authenticates
func (a *ApigeeClient) IsReady() bool {
	return a.isReady
}
