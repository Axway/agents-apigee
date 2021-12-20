package apigee

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"strings"

	coreapi "github.com/Axway/agent-sdk/pkg/api"
	"github.com/Axway/agents-apigee/client/pkg/apigee/models"
)

const (
	orgURL        = "https://api.enterprise.apigee.com/v1/organizations/%s/"
	portalsURL    = "https://apigee.com/portals/api/sites"
	orgDataAPIURL = "https://apigee.com/dapi/api/organizations/%s"
)

func (a *ApigeeClient) defaultHeaders() map[string]string {
	// return the default headers
	return map[string]string{
		"Content-Type":  "application/json",
		"Accept":        "application/json",
		"Authorization": "Bearer " + a.accessToken,
	}
}

func (a *ApigeeClient) getRequest(url string) (*coreapi.Response, error) {
	// return the api response
	return a.getRequestWithQuery(url, map[string]string{})
}

func (a *ApigeeClient) getRequestWithQuery(url string, queryParams map[string]string) (*coreapi.Response, error) {
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

func (a *ApigeeClient) postRequest(url string, data []byte) (*coreapi.Response, error) {
	// return the api response
	return a.postRequestWithQuery(url, map[string]string{}, data)
}

func (a *ApigeeClient) postRequestWithQuery(url string, queryParams map[string]string, data []byte) (*coreapi.Response, error) {
	// create the post request
	request := coreapi.Request{
		Method:      coreapi.POST,
		URL:         url,
		Headers:     a.defaultHeaders(),
		QueryParams: queryParams,
		Body:        data,
	}

	// return the api response
	return a.apiClient.Send(request)
}

func (a *ApigeeClient) putRequest(url string, data []byte) (*coreapi.Response, error) {
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

func (a *ApigeeClient) deleteRequest(url string) (*coreapi.Response, error) {
	// create the put request
	request := coreapi.Request{
		Method:  coreapi.DELETE,
		URL:     url,
		Headers: a.defaultHeaders(),
	}

	// return the api response
	return a.apiClient.Send(request)
}

//GetDevelopers - get the list of developers for the org
func (a *ApigeeClient) GetDevelopers() []string {
	// Get the developers
	response, _ := a.getRequest(fmt.Sprintf(orgURL+"developers", a.cfg.Organization))
	developers := []string{}
	json.Unmarshal(response.Body, &developers)

	return developers
}

//GetDeveloper - get the developer by email
func (a *ApigeeClient) GetDeveloper(devEmail string) (*models.Developer, error) {
	// Get the developers
	response, err := a.getRequest(fmt.Sprintf(orgURL+"developers/%s", a.cfg.Organization, strings.ToLower(devEmail)))
	if err != nil {
		return nil, err
	}
	developer := models.Developer{}
	err = json.Unmarshal(response.Body, &developer)
	if err != nil {
		return nil, err
	}

	return &developer, err
}

//CreateDeveloper - get the list of developers for the org
func (a *ApigeeClient) CreateDeveloper(newDev models.Developer) (*models.Developer, error) {
	// Get the developers
	data, _ := json.Marshal(newDev)
	response, err := a.postRequest(fmt.Sprintf(orgURL+"developers", a.cfg.Organization), data)
	if err != nil {
		return nil, err
	}
	developer := models.Developer{}
	err = json.Unmarshal(response.Body, &developer)
	if err != nil {
		return nil, err
	}

	return &developer, err
}

//CreateDeveloperApp - create an app for the developer
func (a *ApigeeClient) CreateDeveloperApp(newApp models.DeveloperApp) (*models.DeveloperApp, error) {
	// create a new developer app
	data, _ := json.Marshal(newApp)
	response, err := a.postRequest(fmt.Sprintf(orgURL+"developers/%s/apps", a.cfg.Organization, newApp.DeveloperId), data)
	if err != nil {
		return nil, err
	}
	if response.Code != http.StatusCreated {
		return nil, fmt.Errorf("received an unexpected response code %d from Apigee when creating the app", response.Code)
	}

	devApp := models.DeveloperApp{}
	err = json.Unmarshal(response.Body, &devApp)
	if err != nil {
		return nil, err
	}

	return &devApp, err
}

//RemoveDeveloperApp - create an app for the developer
func (a *ApigeeClient) RemoveDeveloperApp(appName, developerID string) error {
	// create a new developer app
	response, err := a.deleteRequest(fmt.Sprintf(orgURL+"developers/%s/apps/%s", a.cfg.Organization, developerID, appName))
	if err != nil {
		return err
	}
	if response.Code != http.StatusOK {
		return fmt.Errorf("received an unexpected response code %d from Apigee when deleting the app", response.Code)
	}

	return nil
}

//GetProducts - get the list of products for the org
func (a *ApigeeClient) GetProducts() Products {
	// Get the products
	response, _ := a.getRequest(fmt.Sprintf(orgURL+"apiproducts", a.cfg.Organization))
	products := Products{}
	json.Unmarshal(response.Body, &products)

	return products
}

//GetPortals - get the list of portals for the org
func (a *ApigeeClient) GetPortals() []PortalData {
	// Get the portals
	response, _ := a.getRequestWithQuery(portalsURL, map[string]string{"orgname": a.cfg.Organization})
	portalRes := PortalsResponse{}
	json.Unmarshal(response.Body, &portalRes)

	return portalRes.Data
}

//getPortals - get the list of portals for the org
func (a *ApigeeClient) GetPortal(portalID string) PortalData {
	// Get the portals
	response, _ := a.getRequest(fmt.Sprintf("%s/%s/portal", portalsURL, portalID))
	portalRes := PortalResponse{}
	json.Unmarshal(response.Body, &portalRes)

	return portalRes.Data
}

//GetPortalAPIs - get the list of portals for the org
func (a *ApigeeClient) GetPortalAPIs(portalID string) []*APIDocData {
	// Get the apidocs
	response, _ := a.getRequest(fmt.Sprintf("%s/%s/apidocs", portalsURL, portalID))
	apiDocRes := APIDocDataResponse{}
	json.Unmarshal(response.Body, &apiDocRes)

	return apiDocRes.Data
}

//GetProduct - get details of the product
func (a *ApigeeClient) GetProduct(productName string) models.ApiProduct {
	// Get the product
	response, _ := a.getRequest(fmt.Sprintf(orgURL+"apiproducts/%s", a.cfg.Organization, productName))
	product := models.ApiProduct{}
	json.Unmarshal(response.Body, &product)

	return product
}

//GetImageWithURL - get the list of portals for the org
func (a *ApigeeClient) GetImageWithURL(imageURL, portalURL string) (string, string) {
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

//GetSpecContent - get the spec content for an api product
func (a *ApigeeClient) GetSpecContent(contentID string) []byte {
	// Get the spec content file
	response, err := a.getRequest(fmt.Sprintf(orgDataAPIURL+"/specs/doc/%s/content", a.cfg.Organization, contentID))

	if err != nil {
		return []byte{}
	}

	return response.Body
}

//GetRevisionSpec - gets the resource file of type openapi for the org, api, revision, and spec file specified
func (a *ApigeeClient) GetRevisionSpec(apiName, revisionNumber, specFile string) []byte {
	// Get the openapi resource file
	response, _ := a.getRequest(fmt.Sprintf(orgURL+"apis/%s/revisions/%s/resourcefiles/openapi/%s", a.cfg.Organization, apiName, revisionNumber, specFile))

	return response.Body
}

//GetSwagger - downloads the specfile from apigee given the url path of its location
func (a *ApigeeClient) GetSwagger(specPath string) []byte {
	// Get the spec file
	response, _ := a.getRequest(fmt.Sprintf("https://apigee.com%s", specPath))

	return response.Body
}

//GetSharedFlow - gets the list of shared flows
func (a *ApigeeClient) GetSharedFlow(name string) (*models.SharedFlowRevisionDeploymentDetails, error) {
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

//CreateSharedFlow - uploads an apigee bundle as a shared flow
func (a *ApigeeClient) CreateSharedFlow(data []byte, name string) error {
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

//DeploySharedFlow - deploye the shared flow and revision to the environmnet
func (a *ApigeeClient) DeploySharedFlow(env, name, revision string) error {
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
func (a *ApigeeClient) PublishSharedFlowToEnvironment(env, name string) error {
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
