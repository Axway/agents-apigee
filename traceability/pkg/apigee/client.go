package apigee

import (
	"time"

	coreapi "github.com/Axway/agent-sdk/pkg/api"
	"github.com/Axway/agent-sdk/pkg/jobs"

	"github.com/Axway/agents-apigee/traceability/pkg/config"
)

// GatewayClient - Represents the Gateway client
type GatewayClient struct {
	cfg          *config.ApigeeConfig
	apiClient    coreapi.Client
	accessToken  string
	pollInterval time.Duration
	envToURLs    map[string][]string
}

// NewClient - Creates a new Gateway Client
func NewClient(apigeeCfg *config.ApigeeConfig) (*GatewayClient, error) {
	client := &GatewayClient{
		apiClient:    coreapi.NewClient(nil, ""),
		cfg:          apigeeCfg,
		pollInterval: apigeeCfg.GetPollInterval(),
		envToURLs:    make(map[string][]string),
	}

	// create the auth job and register it
	authentication := newAuthJob(client.apiClient, apigeeCfg.Auth.GetUsername(), apigeeCfg.Auth.GetPassword(), client.setAccessToken)
	jobs.RegisterIntervalJobWithName(authentication, 10*time.Minute, "APIGEE Auth Token")

	return client, nil
}

func (a *GatewayClient) setAccessToken(token string) {
	a.accessToken = token
}
