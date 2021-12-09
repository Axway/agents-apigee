package apigee

import (
	"time"

	coreapi "github.com/Axway/agent-sdk/pkg/api"

	"github.com/Axway/agents-apigee/discovery/pkg/config"
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

	return client, nil
}

func (a *GatewayClient) setAccessToken(token string) {
	a.accessToken = token
}
