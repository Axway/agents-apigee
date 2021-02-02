package apigee

import (
	"encoding/json"
	"fmt"
	"net/url"

	coreapi "github.com/Axway/agent-sdk/pkg/api"
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

//getEnvironments - get the list of environments for the org
func (a *GatewayClient) getEnvironments() environments {

	// Get the environments
	response, _ := a.getRequest(fmt.Sprintf(orgURL+"environments", a.cfg.Organization))
	environments := environments{}
	json.Unmarshal(response.Body, &environments)

	return environments
}

//getAPIs - get the list of apis for the org
func (a *GatewayClient) getAPIs() apis {

	// Get the apis
	response, _ := a.getRequest(fmt.Sprintf(orgURL+"apis", a.cfg.Organization))
	apiProxies := apis{}
	json.Unmarshal(response.Body, &apiProxies)

	return apiProxies
}

//getAPIsWithData - get the list of apis for the org
func (a *GatewayClient) getAPIsWithData() []models.ApiProxy {
	queryParams := map[string]string{
		"includeRevisions": "true",
		"includeMetaData":  "true",
	}

	// Get the apis
	response, _ := a.getRequestWithQuery(fmt.Sprintf(orgURL+"apis", a.cfg.Organization), queryParams)
	apiProxies := []models.ApiProxy{}
	json.Unmarshal(response.Body, &apiProxies)

	return apiProxies
}

//getAPI - get details of the api
func (a *GatewayClient) getAPI(apiName string) models.ApiProxy {

	// Get the apis
	response, _ := a.getRequest(fmt.Sprintf(orgURL+"apis/%s", a.cfg.Organization, apiName))
	apiProxy := models.ApiProxy{}
	json.Unmarshal(response.Body, &apiProxy)

	return apiProxy
}

//getRevisionsDetails - get the revision details for a specific org, api, revision combo
func (a *GatewayClient) getRevisionsDetails(apiName, revisionNumber string) models.ApiProxyRevision {

	// Get the revision details
	response, _ := a.getRequest(fmt.Sprintf(orgURL+"apis/%s/revisions/%s", a.cfg.Organization, apiName, revisionNumber))
	apiRevision := models.ApiProxyRevision{}
	json.Unmarshal(response.Body, &apiRevision)

	return apiRevision
}

//getRevisionDefinitionBundle - get the revision defintion bundle for a specific org, api, revision combo
func (a *GatewayClient) getRevisionDefinitionBundle(apiName, revisionNumber string) []byte {
	queryParams := map[string]string{
		"format": "bundle",
	}

	// Get the revision bundle
	response, _ := a.getRequestWithQuery(fmt.Sprintf(orgURL+"apis/%s/revisions/%s", a.cfg.Organization, apiName, revisionNumber), queryParams)

	return response.Body
}

//getResourceFiles - get the revision resource files list for the org, api, revision combo
func (a *GatewayClient) getResourceFiles(apiName, revisionNumber string) models.ApiProxyRevisionResourceFiles {

	// Get the revision resource files
	response, _ := a.getRequest(fmt.Sprintf(orgURL+"apis/%s/revisions/%s/resourcefiles", a.cfg.Organization, apiName, revisionNumber))
	apiResourceFiles := models.ApiProxyRevisionResourceFiles{}
	json.Unmarshal(response.Body, &apiResourceFiles)

	return apiResourceFiles
}

//getRevisionSpec - gets the resource file of type openapi for  the org, api, revision, and spec file specified
func (a *GatewayClient) getRevisionSpec(apiName, revisionNumber, specFile string) []byte {

	// Get the openapi resource file
	response, _ := a.getRequest(fmt.Sprintf(orgURL+"apis/%s/revisions/%s/resourcefiles/openapi/%s", a.cfg.Organization, apiName, revisionNumber, specFile))

	return response.Body
}

//getDeployments - gets all deployments of an api in the org
func (a *GatewayClient) getDeployments(apiName string) models.DeploymentDetails {

	// Get the deployments
	response, _ := a.getRequest(fmt.Sprintf(orgURL+"apis/%s/deployments", a.cfg.Organization, apiName))
	deployments := models.DeploymentDetails{}
	json.Unmarshal(response.Body, &deployments)

	return deployments
}

//getVirtualHosts - gets all virtual hosts for an environment in the org
func (a *GatewayClient) getVirtualHosts(environment string) virtualHosts {

	// Get the virtual hosts
	response, _ := a.getRequest(fmt.Sprintf(orgURL+"/environments/%s/virtualhosts", a.cfg.Organization, environment))
	hosts := virtualHosts{}
	json.Unmarshal(response.Body, &hosts)

	return hosts
}

//getVirtualHost - gets the details on a virtual host for an environment, hostname combo in the org
func (a *GatewayClient) getVirtualHost(environment, hostName string) models.VirtualHost {

	// Get the virtual host details
	response, _ := a.getRequest(fmt.Sprintf(orgURL+"/environments/%s/virtualhosts/%s", a.cfg.Organization, environment, hostName))
	host := models.VirtualHost{}
	json.Unmarshal(response.Body, &host)

	return host
}

//getSwagger - downloads the specfile from apigee given the url path of its location
func (a *GatewayClient) getSwagger(specPath string) []byte {

	// Get the spec file
	response, _ := a.getRequest(fmt.Sprintf("https://apigee.com%s", specPath))

	return response.Body
}
