package apigee

import (
	"encoding/json"
	"fmt"
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

// logs: (token) => ({
// 	method: 'GET',
//	url: `https://apimonitoring.enterprise.apigee.com/logs?org={organizationId}`,
// 	headers: { Authorization: `Basic ${token}`, Accept: 'application/json' }
// }),
func (a *GatewayClient) getLogs() apigeeLogs {

	// Get the initial authentication token
	response, _ := a.getRequest(fmt.Sprintf(orgURL+"logs?org=%s", a.cfg.Organization))
	log.Debugf("getLogs: %s", string(response.Body))
	apigeeLogs := apigeeLogs{}
	json.Unmarshal(response.Body, &apigeeLogs)

	return apigeeLogs
}

// apiproxies: (token) => ({
// 	method: 'GET',
//	url: `https://apimonitoring.enterprise.apigee.com/logs/apiproxies?org={organizationId}`,
// 	headers: { Authorization: `Basic ${token}`, Accept: 'application/json' }
// }),
func (a *GatewayClient) getAPIProxies() apiProxies {

	// Get the initial authentication token
	response, _ := a.getRequest(fmt.Sprintf(orgURL+"logs/apiproxies?org=%s", a.cfg.Organization))
	log.Debugf("getAPIProxies: %s", string(response.Body))
	apiProxies := apiProxies{}
	json.Unmarshal(response.Body, &apiProxies)

	return apiProxies
}

// apiproxies: (token) => ({
// 	method: 'GET',
//	url: `https://apimonitoring.enterprise.apigee.com/metrics/events?org={organizationId}`,
// 	headers: { Authorization: `Basic ${token}`, Accept: 'application/json' }
// }),
func (a *GatewayClient) getEvents() apigeeEvents {

	// Get the initial authentication token
	response, _ := a.getRequest(fmt.Sprintf(orgURL+"metrics/events?org=%s", a.cfg.Organization))
	log.Debugf("getEvents: %s", string(response.Body))
	apigeeEvents := apigeeEvents{}
	json.Unmarshal(response.Body, &apigeeEvents)

	return apigeeEvents
}
