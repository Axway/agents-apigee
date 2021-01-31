package apigee

import (
	"encoding/json"
	"fmt"
	"net/url"

	coreapi "github.com/Axway/agent-sdk/pkg/api"
	"github.com/Axway/agent-sdk/pkg/util/log"
	"github.com/Axway/agents-apigee/discovery/pkg/apigee/models"
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

// environments: (token) => ({
// 	method: 'GET',
// 	url: `https://api.enterprise.apigee.com/v1/organizations/${organizationId}/environments`,
// 	headers: { Authorization: `Basic ${token}`, Accept: 'application/json' }
// }),
func (a *GatewayClient) getEnvironments() environments {

	// Get the initial authentication token
	response, _ := a.getRequest(fmt.Sprintf(orgURL+"environments", a.cfg.Organization))
	log.Debugf("getEnvironments: %s", string(response.Body))
	environments := environments{}
	json.Unmarshal(response.Body, &environments)

	return environments
}

// APIsOptions: (token) => ({
// 	method: 'GET',
// 	url: `https://api.enterprise.apigee.com/v1/organizations/${organizationId}/apis`,
// 	headers: { Authorization: `Basic ${token}` }
// }),
func (a *GatewayClient) getAPIs() apis {

	// Get the initial authentication token
	response, _ := a.getRequest(fmt.Sprintf(orgURL+"apis", a.cfg.Organization))
	log.Debugf("getAPIs: %s", string(response.Body))
	apiProxies := apis{}
	json.Unmarshal(response.Body, &apiProxies)

	return apiProxies
}

// revisionDetails: (apiName, revisionNumber, token) => ({
// 	method: 'GET',
// 	url: `https://api.enterprise.apigee.com/v1/organizations/${organizationId}/apis/${apiName}/revisions/${revisionNumber}`,
// 	headers: { Authorization: `Basic ${token}` }
// }),
func (a *GatewayClient) getRevisionsDetails(apiName, revisionNumber string) models.ApiProxyRevision {

	// Get the initial authentication token
	response, _ := a.getRequest(fmt.Sprintf(orgURL+"apis/%s/revisions/%s", a.cfg.Organization, apiName, revisionNumber))
	log.Debugf("getRevisionsDetails: %s", string(response.Body))
	apiRevision := models.ApiProxyRevision{}
	json.Unmarshal(response.Body, &apiRevision)

	return apiRevision
}

// resourceFiles: (apiName, revisionNumber, token) => ({
// 	method: 'GET',
// 	url: `https://api.enterprise.apigee.com/v1/organizations/${organizationId}/apis/${apiName}/revisions/${revisionNumber}/resourcefiles`,
// 	headers: { Authorization: `Basic ${token}` }
// }),
func (a *GatewayClient) getResourceFiles(apiName, revisionNumber string) models.ApiProxyRevisionResourceFiles {

	// Get the initial authentication token
	response, _ := a.getRequest(fmt.Sprintf(orgURL+"apis/%s/revisions/%s/resourcefiles", a.cfg.Organization, apiName, revisionNumber))
	log.Debugf("getResourceFiles: %s", string(response.Body))
	apiResourceFiles := models.ApiProxyRevisionResourceFiles{}
	json.Unmarshal(response.Body, &apiResourceFiles)

	return apiResourceFiles
}

// revisionSpec: (apiName, revisionNumber, specFile, token) => ({
// 	method: 'GET',
// 	url: `https://api.enterprise.apigee.com/v1/organizations/${organizationId}/apis/${apiName}/revisions/${revisionNumber}/resourcefiles/openapi/${specFile}`,
// 	headers: { Authorization: `Basic ${token}` }
// }),
func (a *GatewayClient) getRevisionSpec(apiName, revisionNumber, specFile string) []byte {

	// Get the initial authentication token
	response, _ := a.getRequest(fmt.Sprintf(orgURL+"apis/%s/revisions/%s/resourcefiles/openapi/%s", a.cfg.Organization, apiName, revisionNumber, specFile))
	log.Debugf("getRevisionSpec: %s", string(response.Body))

	return response.Body
}

// deployments: (apiName, token) => ({
// 	method: 'GET',
// 	url: `https://api.enterprise.apigee.com/v1/organizations/${organizationId}/apis/${apiName}/deployments`,
// 	headers: { Authorization: `Basic ${token}` }
// }),
func (a *GatewayClient) getDeployments(apiName string) models.DeploymentDetails {

	// Get the initial authentication token
	response, _ := a.getRequest(fmt.Sprintf(orgURL+"apis/%s/deployments", a.cfg.Organization, apiName))
	log.Debugf("getDeployments: %s", string(response.Body))
	deployments := models.DeploymentDetails{}
	json.Unmarshal(response.Body, &deployments)

	return deployments
}

// hosts: (envName, token) => ({
// 	method: 'GET',
// 	url: `https://api.enterprise.apigee.com/v1/organizations/${organizationId}/environments/${envName}/virtualhosts`,
// 	headers: { Authorization: `Basic ${token}` }
// }),
func (a *GatewayClient) getVirtualHosts(environment string) virtualHosts {

	// Get the initial authentication token
	response, _ := a.getRequest(fmt.Sprintf(orgURL+"/environments/%s/virtualhosts", a.cfg.Organization, environment))
	log.Debugf("getVirtualHosts: %s", string(response.Body))
	hosts := virtualHosts{}
	json.Unmarshal(response.Body, &hosts)

	return hosts
}

// virtualHosts: (envName, virtualHostName, token) => ({
// 	method: 'GET',
// 	url: `https://api.enterprise.apigee.com/v1/organizations/${organizationId}/environments/${envName}/virtualhosts/${virtualHostName}`,
// 	headers: { Authorization: `Basic ${token}` }
// }),
func (a *GatewayClient) getVirtualHost(environment, hostName string) models.VirtualHost {

	// Get the initial authentication token
	response, _ := a.getRequest(fmt.Sprintf(orgURL+"/environments/%s/virtualhosts/%s", a.cfg.Organization, environment, hostName))
	log.Debugf("getVirtualHost: %s", string(response.Body))
	host := models.VirtualHost{}
	json.Unmarshal(response.Body, &host)

	return host
}

// swagger: (specPath, token) => ({
// 	method: 'GET',
// 	url: `https://apigee.com${specPath}`,
// 	headers: { Authorization: `Bearer ${token}` }
// })
func (a *GatewayClient) getSwagger(specPath string) string {

	// Get the initial authentication token
	response, _ := a.getRequest(fmt.Sprintf("https://apigee.com%s", specPath))
	log.Debugf("getSwagger: %s", string(response.Body))

	return string(response.Body)
}
