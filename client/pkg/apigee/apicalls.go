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
	"time"

	"github.com/Axway/agents-apigee/client/pkg/apigee/models"
)

const (
	orgURL        = "https://api.enterprise.apigee.com/v1/organizations/%s"
	portalsURL    = "https://apigee.com/portals/api/sites"
	orgDataAPIURL = "https://apigee.com/dapi/api/organizations/%s"
)

//GetDevelopers - get the list of developers for the org
func (a *ApigeeClient) GetDevelopers() []string {
	// Get the developers
	response, err := a.newRequest(http.MethodGet, fmt.Sprintf(orgURL+"/developers", a.cfg.Organization),
		WithDefaultHeaders(),
	).Execute()

	developers := []string{}
	if err == nil {
		json.Unmarshal(response.Body, &developers)
	}

	return developers
}

//GetEnvironments - get the list of environments for the org
func (a *ApigeeClient) GetEnvironments() []string {
	// Get the developers
	response, err := a.newRequest(http.MethodGet, fmt.Sprintf(orgURL+"/environments", a.cfg.Organization),
		WithDefaultHeaders(),
	).Execute()

	environments := []string{}
	if err == nil {
		json.Unmarshal(response.Body, &environments)
	}

	return environments
}

//GetDeveloper - get the developer by email
func (a *ApigeeClient) GetDeveloper(devEmail string) (*models.Developer, error) {
	// Get the developers
	response, err := a.newRequest(http.MethodGet, fmt.Sprintf(orgURL+"/developers/%s", a.cfg.Organization, strings.ToLower(devEmail)),
		WithDefaultHeaders(),
	).Execute()
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
	data, err := json.Marshal(newDev)
	if err != nil {
		return nil, err
	}

	response, err := a.newRequest(http.MethodPost, fmt.Sprintf(orgURL+"/developers", a.cfg.Organization),
		WithDefaultHeaders(),
		WithBody(data),
	).Execute()
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
func (a *ApigeeClient) createOrUpdateDeveloperApp(method string, newApp models.DeveloperApp) (*models.DeveloperApp, error) {
	// create a developer app
	data, err := json.Marshal(newApp)
	if err != nil {
		return nil, err
	}

	url := fmt.Sprintf(orgURL+"/developers/%s/apps", a.cfg.Organization, newApp.DeveloperId)
	if method == http.MethodPut {
		url = fmt.Sprintf("%s/%s", url, newApp.Name)
	}
	response, err := a.newRequest(
		method, url,
		WithDefaultHeaders(),
		WithBody(data),
	).Execute()
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

func (a *ApigeeClient) CreateDeveloperApp(newApp models.DeveloperApp) (*models.DeveloperApp, error) {
	return a.createOrUpdateDeveloperApp(http.MethodPost, newApp)
}

func (a *ApigeeClient) UpdateDeveloperApp(app models.DeveloperApp) (*models.DeveloperApp, error) {
	return a.createOrUpdateDeveloperApp(http.MethodPut, app)
}

func (a *ApigeeClient) GetDeveloperApp(name string) (*models.DeveloperApp, error) {
	url := fmt.Sprintf(orgURL+"/developers/%s/apps/%s", a.cfg.Organization, a.GetDeveloperID())
	response, err := a.newRequest(
		http.MethodGet, url,
		WithDefaultHeaders(),
	).Execute()
	if err != nil {
		return nil, err
	}
	if response.Code != http.StatusOK {
		return nil, fmt.Errorf("received an unexpected response code %d from Apigee when retrieving the app", response.Code)
	}

	devApp := models.DeveloperApp{}
	err = json.Unmarshal(response.Body, &devApp)
	return &devApp, err
}

//RemoveDeveloperApp - create an app for the developer
func (a *ApigeeClient) RemoveDeveloperApp(appName, developerID string) error {
	// create a new developer app
	response, err := a.newRequest(http.MethodDelete, fmt.Sprintf(orgURL+"/developers/%s/apps/%s", a.cfg.Organization, developerID, appName),
		WithDefaultHeaders(),
	).Execute()

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
	response, err := a.newRequest(http.MethodGet, fmt.Sprintf(orgURL+"/apiproducts", a.cfg.Organization),
		WithDefaultHeaders(),
	).Execute()

	products := Products{}
	if err == nil {
		json.Unmarshal(response.Body, &products)
	}

	return products
}

//GetPortals - get the list of portals for the org
func (a *ApigeeClient) GetPortals() []PortalData {
	// Get the portals
	response, err := a.newRequest(http.MethodGet, portalsURL,
		WithDefaultHeaders(),
		WithQueryParam("orgname", a.cfg.Organization),
	).Execute()

	portalRes := PortalsResponse{}
	if err == nil {
		json.Unmarshal(response.Body, &portalRes)
	}

	return portalRes.Data
}

//GetPortal - get the list of portals for the org
func (a *ApigeeClient) GetPortal(portalID string) PortalData {
	// Get the portals
	response, err := a.newRequest(http.MethodGet, fmt.Sprintf("%s/%s/portal", portalsURL, portalID),
		WithDefaultHeaders(),
	).Execute()

	portalRes := PortalResponse{}
	if err == nil {
		json.Unmarshal(response.Body, &portalRes)
	}

	return portalRes.Data
}

//GetPortalAPIs - get the list of portals for the org
func (a *ApigeeClient) GetPortalAPIs(portalID string) ([]*APIDocData, error) {
	// Get the apidocs
	response, err := a.newRequest(http.MethodGet, fmt.Sprintf("%s/%s/apidocs", portalsURL, portalID),
		WithDefaultHeaders(),
	).Execute()

	if response.Code != http.StatusOK {
		return nil, err
	}

	apiDocRes := APIDocDataResponse{}
	json.Unmarshal(response.Body, &apiDocRes)

	return apiDocRes.Data, err
}

//GetProduct - get details of the product
func (a *ApigeeClient) GetProduct(productName string) (*models.ApiProduct, error) {
	// Get the product
	response, err := a.newRequest(http.MethodGet, fmt.Sprintf(orgURL+"/apiproducts/%s", a.cfg.Organization, productName),
		WithDefaultHeaders(),
	).Execute()
	if err != nil {
		return nil, err
	}
	product := &models.ApiProduct{}
	json.Unmarshal(response.Body, product)

	return product, nil
}

//GetImageWithURL - get the list of portals for the org
func (a *ApigeeClient) GetImageWithURL(imageURL, portalURL string) (string, string) {
	// Get the portal
	response, err := a.newRequest(http.MethodGet, fmt.Sprintf("%s%s", portalURL, imageURL)).Execute()
	if err != nil {
		return "", ""
	}

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
	response, err := a.newRequest(http.MethodGet, fmt.Sprintf(orgDataAPIURL+"/specs/doc/%s/content", a.cfg.Organization, contentID),
		WithDefaultHeaders(),
	).Execute()

	if err != nil {
		return []byte{}
	}

	return response.Body
}

//GetRevisionSpec - gets the resource file of type openapi for the org, api, revision, and spec file specified
func (a *ApigeeClient) GetRevisionSpec(apiName, revisionNumber, specFile string) []byte {
	// Get the openapi resource file
	response, err := a.newRequest(http.MethodGet, fmt.Sprintf(orgURL+"/apis/%s/revisions/%s/resourcefiles/openapi/%s", a.cfg.Organization, apiName, revisionNumber, specFile),
		WithDefaultHeaders(),
	).Execute()

	if err != nil {
		return []byte{}
	}

	return response.Body
}

//GetSwagger - downloads the specfile from apigee given the url path of its location
func (a *ApigeeClient) GetSwagger(specPath string) []byte {
	// Get the spec file
	response, err := a.newRequest(http.MethodGet, fmt.Sprintf("https://apigee.com%s", specPath),
		WithDefaultHeaders(),
	).Execute()

	if err != nil {
		return []byte{}
	}

	return response.Body
}

//GetSharedFlow - gets the list of shared flows
func (a *ApigeeClient) GetSharedFlow(name string) (*models.SharedFlowRevisionDeploymentDetails, error) {
	// Get the shared flows list
	response, err := a.newRequest(http.MethodGet, fmt.Sprintf(orgURL+"/sharedflows/%v", a.cfg.Organization, name),
		WithDefaultHeaders(),
	).Execute()
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

//DeploySharedFlow - deploy the shared flow and revision to the environment
func (a *ApigeeClient) DeploySharedFlow(env, name, revision string) error {

	// deploy the shared flow to the environment
	_, err := a.newRequest(http.MethodPost, fmt.Sprintf(orgURL+"/environments/%v/sharedflows/%v/revisions/%v/deployments", a.cfg.Organization, env, name, revision),
		WithDefaultHeaders(),
		WithBody([]byte{}),
		WithQueryParam("override", "true"),
	).Execute()

	if err != nil {
		return err
	}

	return nil
}

//PublishSharedFlowToEnvironment - publish the shared flow
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
	_, err := a.newRequest(http.MethodPut, fmt.Sprintf(orgURL+"/environments/%v/flowhooks/PostProxyFlowHook", a.cfg.Organization, env),
		WithDefaultHeaders(),
		WithBody(data),
	).Execute()
	return err
}

//GetStats - get the api stats for a specific environment
func (a *ApigeeClient) GetStats(env, metricSelect string, start, end time.Time) (*models.Metrics, error) {
	// Get the spec content file
	const dimension = "apiproxy" // https://docs.apigee.com/api-platform/analytics/analytics-reference#dimensions
	const format = "01/02/2006 15:04"

	response, err := a.newRequest(http.MethodGet, fmt.Sprintf(orgURL+"/environments/%v/stats/%s", a.cfg.Organization, env, dimension),
		WithQueryParams(map[string]string{
			"select":    metricSelect,
			"timeUnit":  "minute",
			"timeRange": fmt.Sprintf("%s~%s", time.Time.UTC(start).Format(format), time.Time.UTC(end).Format(format)),
			"sortby":    "sum(message_count)",
			"sort":      "ASC",
		}),
		WithDefaultHeaders(),
	).Execute()

	if err != nil {
		return nil, err
	}

	stats := &models.Metrics{}
	err = json.Unmarshal(response.Body, stats)
	if err != nil {
		return nil, err
	}

	return stats, nil
}
