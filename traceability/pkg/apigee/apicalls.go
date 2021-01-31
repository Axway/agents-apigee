package apigee

import (
	"encoding/json"
	"net/url"

	coreapi "github.com/Axway/agent-sdk/pkg/api"
	"github.com/Axway/agent-sdk/pkg/util/log"
)

func (a *GatewayClient) postAuth(authData url.Values) AuthResponse {
	request := coreapi.Request{
		Method: coreapi.POST,
		URL:    apigeeAuthURL,
		Headers: map[string]string{
			"Content-Type":  "application/x-www-form-urlencoded",
			"Authorization": "Basic " + apigeeAuthToken,
		},
		Body: []byte(authData.Encode()),
	}

	// Get the initial authentication token
	response, _ := a.apiClient.Send(request)
	authResponse := AuthResponse{}
	log.Debugf("postAuth: %s", string(response.Body))
	json.Unmarshal(response.Body, &authResponse)

	a.accessToken = authResponse.AccessToken
	return authResponse
}

func (a *GatewayClient) getRequest(url string) (*coreapi.Response, error) {
	request := coreapi.Request{
		Method: coreapi.GET,
		URL:    url,
		Headers: map[string]string{
			"Accept":        "application/json",
			"Authorization": "Bearer " + a.accessToken,
		},
	}

	// return the api response
	return a.apiClient.Send(request)
}
