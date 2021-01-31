package apigee

import (
	"net/url"
	"time"

	coreapi "github.com/Axway/agent-sdk/pkg/api"
	"github.com/Axway/agent-sdk/pkg/util/log"
	"github.com/Axway/agents-apigee/traceability/pkg/config"
)

const (
	apigeeAuthURL   = "https://login.apigee.com/oauth/token"
	apigeeAuthToken = "ZWRnZWNsaTplZGdlY2xpc2VjcmV0" //hardcoded to edgecli:edgeclisecret
	orgURL          = "hhttps://apimonitoring.enterprise.apigee.com/"
)

// GatewayClient - Represents the Gateway client
type GatewayClient struct {
	cfg          *config.ApigeeConfig
	apiClient    coreapi.Client
	accessToken  string
	eventChannel chan string
	stopChannel  chan bool
}

// NewClient - Creates a new Gateway Client
func NewClient(apigeeCfg *config.ApigeeConfig) (*GatewayClient, error) {
	gatewayClient := &GatewayClient{
		apiClient: coreapi.NewClient(nil, ""),
		cfg:       apigeeCfg,
	}

	// Start the authentication
	gatewayClient.Authenticate()

	return gatewayClient, nil
}

// Authenticate - handles the initial authentication then starts a go routine to refresh the token
func (a *GatewayClient) Authenticate() error {
	authData := url.Values{}
	authData.Set("grant_type", password.String())
	authData.Set("username", a.cfg.GetAuth().GetUsername())
	authData.Set("password", a.cfg.GetAuth().GetPassword())

	authResponse := a.postAuth(authData)

	log.Debugf("APIGEE auth token: %s", authResponse.AccessToken)

	// Continually refresh the token
	go func() {
		for {
			// Refresh the token 5 minutes before expiration
			time.Sleep(time.Duration(authResponse.ExpiresIn-300) * time.Second)

			log.Debug("Refreshing auth token")
			authData := url.Values{}
			authData.Set("grant_type", refresh.String())
			authData.Set("refresh_token", authResponse.RefreshToken)

			authResponse = a.postAuth(authData)
			log.Debugf("APIGEE auth token: %s", authResponse.AccessToken)
		}
	}()

	return nil
}

// Start - Starts reading log file
func (a *GatewayClient) Start() {
	// ctx := context.Background()
	// go func() {
	// 	for {
	// 		select {
	// 		case <-a.stopChannel:
	// 			// a.hub.Close(ctx)
	// 			return
	// 		}
	// 	}
	// }()

	// runtimeInfo, err := a.hub.GetRuntimeInformation(ctx)
	// if err != nil {
	// 	fmt.Println(err)
	// 	return
	// }

	// for _, partitionID := range runtimeInfo.PartitionIDs {
	// 	_, err := c.hub.Receive(ctx, partitionID, c.eventHandler, eventhub.ReceiveWithLatestOffset())
	// 	if err != nil {
	// 		fmt.Println("Error: ", err)
	// 	}
	// }
}

// Stop - Stop processing subscriptions
func (a *GatewayClient) Stop() {
	a.stopChannel <- true
}
