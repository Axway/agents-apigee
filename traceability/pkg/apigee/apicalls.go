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
	// return the api response
	return a.getRequestWithQuery(url, map[string]string{})
}

func (a *GatewayClient) getRequestWithQuery(url string, queryParams map[string]string) (*coreapi.Response, error) {
	request := coreapi.Request{
		Method: coreapi.GET,
		URL:    url,
		Headers: map[string]string{
			"Accept":        "application/json",
			"Authorization": "Bearer " + a.accessToken,
		},
		QueryParams: queryParams,
	}

	// return the api response
	return a.apiClient.Send(request)
}

// environments: (token) => ({
// 	method: 'GET',
// 	url: `https://api.enterprise.apigee.com/v1/organizations/${organizationId}/environments`,
// 	headers: { Authorization: `Basic ${token}`, Accept: 'application/json' }
// }),
func (a *GatewayClient) getEnvironments() environments {

	// Get the initial authentication token
	response, _ := a.getRequest(fmt.Sprintf(discoURL+"environments", a.cfg.Organization))
	log.Debugf("getEnvironments: %s", string(response.Body))
	environments := environments{}
	json.Unmarshal(response.Body, &environments)

	return environments
}

// logs: (token) => ({
// 	method: 'GET',
//	url: `https://apimonitoring.enterprise.apigee.com/logs?org={organizationId}&env={environment}`,
// 	headers: { Authorization: `Basic ${token}`, Accept: 'application/json' }
// }),
func (a *GatewayClient) getApigeeLogs(environment string) apigeeLogs {

	queryParams := map[string]string{
		"org": a.cfg.Organization,
		"env": environment,
	}
	// Get the initial authentication token
	response, _ := a.getRequestWithQuery(fmt.Sprintf(traceURL+"logs"), queryParams)
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
	response, _ := a.getRequest(fmt.Sprintf(traceURL+"logs/apiproxies?org=%s", a.cfg.Organization))
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
func (a *GatewayClient) getEvents(environment string) apigeeEvents {

	queryParams := map[string]string{
		"org": a.cfg.Organization,
		"env": environment,
	}
	// Get the initial authentication token
	response, _ := a.getRequestWithQuery(fmt.Sprintf(traceURL+"metrics/events"), queryParams)

	log.Debugf("getEvents: %s", string(response.Body))
	apigeeEvents := apigeeEvents{}
	json.Unmarshal(response.Body, &apigeeEvents)

	return apigeeEvents
}
