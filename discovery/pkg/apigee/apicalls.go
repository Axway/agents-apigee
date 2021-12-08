package apigee

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"

	coreapi "github.com/Axway/agent-sdk/pkg/api"
	"github.com/Axway/agents-apigee/discovery/pkg/apigee/models"
)

const (
	orgURL        = "https://api.enterprise.apigee.com/v1/organizations/%s/"
	portalsURL    = "https://apigee.com/portals/api/sites"
	orgDataAPIURL = "https://apigee.com/dapi/api/organizations/%s"
)

func (a *GatewayClient) defaultHeaders() map[string]string {
	// return the default headers
	return map[string]string{
		"Content-Type":  "application/json",
		"Accept":        "application/json",
		"Authorization": "Bearer " + a.accessToken,
	}
}

func (a *GatewayClient) getRequest(url string) (*coreapi.Response, error) {
	// return the api response
	return a.getRequestWithQuery(url, map[string]string{})
}

func (a *GatewayClient) getRequestWithQuery(url string, queryParams map[string]string) (*coreapi.Response, error) {
	// create the get request
	request := coreapi.Request{
		Method:      coreapi.GET,
		URL:         url,
		Headers:     a.defaultHeaders(),
		QueryParams: queryParams,
	}

	// return the api response
	return a.apiClient.Send(request)
}

func (a *GatewayClient) postRequestWithQuery(url string, queryParams map[string]string, data []byte) (*coreapi.Response, error) {
	// create the post request
	request := coreapi.Request{
		Method:      coreapi.POST,
		URL:         url,
		Headers:     a.defaultHeaders(),
		QueryParams: queryParams,
	}

	// return the api response
	return a.apiClient.Send(request)
}

func (a *GatewayClient) putRequest(url string, data []byte) (*coreapi.Response, error) {
	// create the put request
	request := coreapi.Request{
		Method:  coreapi.PUT,
		URL:     url,
		Headers: a.defaultHeaders(),
		Body:    data,
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

//getProducts - get the list of products for the org
func (a *GatewayClient) getProducts() products {
	// Get the products
	response, _ := a.getRequest(fmt.Sprintf(orgURL+"apiproducts", a.cfg.Organization))
	products := products{}
	json.Unmarshal(response.Body, &products)

	return products
}

//getPortals - get the list of portals for the org
func (a *GatewayClient) getPortals() []portalData {
	// Get the portals
	response, _ := a.getRequestWithQuery(portalsURL, map[string]string{"orgname": a.cfg.Organization})
	portalRes := portalsResponse{}
	json.Unmarshal(response.Body, &portalRes)

	return portalRes.Data
}

//getPortals - get the list of portals for the org
func (a *GatewayClient) getPortal(portalID string) portalData {
	// Get the portals
	response, _ := a.getRequest(fmt.Sprintf("%s/%s/portal", portalsURL, portalID))
	portalRes := portalResponse{}
	json.Unmarshal(response.Body, &portalRes)

	return portalRes.Data
}

//getPortalAPIs - get the list of portals for the org
func (a *GatewayClient) getPortalAPIs(portalID string) []*apiDocData {
	// Get the apidocs
	response, _ := a.getRequest(fmt.Sprintf("%s/%s/apidocs", portalsURL, portalID))
	apiDocRes := apiDocDataResponse{}
	json.Unmarshal(response.Body, &apiDocRes)

	return apiDocRes.Data
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

//getProduct - get details of the product
func (a *GatewayClient) getProduct(productName string) models.ApiProduct {
	// Get the product
	response, _ := a.getRequest(fmt.Sprintf(orgURL+"apiproducts/%s", a.cfg.Organization, productName))
	product := models.ApiProduct{}
	json.Unmarshal(response.Body, &product)

	return product
}

//getImageWithURL - get the list of portals for the org
func (a *GatewayClient) getImageWithURL(imageURL, portalURL string) (string, string) {
	// Get the portal
	request := coreapi.Request{
		Method: coreapi.GET,
		URL:    fmt.Sprintf("%s%s", portalURL, imageURL),
	}
	response, _ := a.apiClient.Send(request)

	contentType := ""
	if contentTypeArray, ok := response.Headers["Content-Type"]; ok {
		contentType = contentTypeArray[0]
		// assuming an octet stream type is actually an image/png
		if contentType == "application/octet-stream" {
			contentType = "image/png"
		}
	}

	if response.Code != 200 || string(response.Body) == "" || contentType == "" {
		return "", ""
	}

	return base64.StdEncoding.EncodeToString(response.Body), contentType
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

//getSpecContent - get the spec content for an api product
func (a *GatewayClient) getSpecContent(contentID string) []byte {
	// Get the spec content file
	response, _ := a.getRequest(fmt.Sprintf(orgDataAPIURL+"/specs/doc/%s/content", a.cfg.Organization, contentID))

	return response.Body
}

//getResourceFiles - get the revision resource files list for the org, api, revision combo
func (a *GatewayClient) getResourceFiles() models.ApiProxyRevisionResourceFiles {
	// Get the revision resource files
	response, _ := a.getRequest(fmt.Sprintf(orgURL+"/resourcefiles", a.cfg.Organization))
	apiResourceFiles := models.ApiProxyRevisionResourceFiles{}
	json.Unmarshal(response.Body, &apiResourceFiles)

	return apiResourceFiles
}

//getRevisionSpec - gets the resource file of type openapi for the org, api, revision, and spec file specified
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

//getSharedFlows - gets the list of shared flows
func (a *GatewayClient) getSharedFlow(name string) (*models.SharedFlowRevisionDeploymentDetails, error) {
	// Get the shared flows list
	response, err := a.getRequest(fmt.Sprintf(orgURL+"/sharedflows/%v", a.cfg.Organization, name))
	if err != nil {
		return nil, err
	}

	if response.Code == http.StatusNotFound {
		return nil, fmt.Errorf("could not find shared flow named %v", name)
	}

	flow := models.SharedFlowRevisionDeploymentDetails{}
	json.Unmarshal(response.Body, &flow)

	return &flow, nil
}

//createSharedFlow - uploads an apigee bundle as a shared flow
func (a *GatewayClient) createSharedFlow(data []byte, name string) error {
	var buffer bytes.Buffer
	writer := multipart.NewWriter(&buffer)

	// Create the flow zip part
	flow, _ := writer.CreateFormFile("file", name+".zip")
	io.Copy(flow, bytes.NewReader(data))
	writer.Close()

	// assemble the request with the writer content type and buffer data
	req, _ := http.NewRequest(http.MethodPost, fmt.Sprintf(orgURL+"/sharedflows?action=import&name=%s", a.cfg.Organization, name), &buffer)
	req.Header.Add("Content-Type", writer.FormDataContentType())
	req.Header.Add("Accept", "application/json")
	req.Header.Add("Authorization", "Bearer "+a.accessToken)
	client := &http.Client{}

	// submit the request
	_, err := client.Do(req)

	return err
}

//deploySharedFlow - deploye the shared flow and revision to the environmnet
func (a *GatewayClient) deploySharedFlow(env, name, revision string) error {
	queryParams := map[string]string{
		"override": "true",
	}

	// deploy the shared flow to the environment
	_, err := a.postRequestWithQuery(fmt.Sprintf(orgURL+"/environments/%v/sharedflows/%v/revisions/%v/deployments", a.cfg.Organization, env, name, revision), queryParams, []byte{})

	if err != nil {
		return err
	}

	return nil
}

//createSharedFlow - gets the list of shared flows
func (a *GatewayClient) publishSharedFlowToEnvironment(env, name string) error {
	// This is the structure that is expected for adding a shared flow as a flow hook
	type flowhook struct {
		ContinueOnError bool   `json:"continueOnError"`
		SharedFlow      string `json:"sharedFlow"`
		State           string `json:"state"`
	}

	// create the data for the put request
	hook := flowhook{
		ContinueOnError: true,
		SharedFlow:      name,
		State:           "deployed",
	}
	data, _ := json.Marshal(hook)

	// Add the flow to the post proxy flow hook
	_, err := a.putRequest(fmt.Sprintf(orgURL+"/environments/%v/flowhooks/PostProxyFlowHook", a.cfg.Organization, env), data)
	return err
}
