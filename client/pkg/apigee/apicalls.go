package apigee

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"time"

	"github.com/Axway/agents-apigee/client/pkg/apigee/models"
	"github.com/Axway/agents-apigee/client/pkg/util"
)

const (
	orgURL        = "https://api.enterprise.apigee.com/v1/organizations/%s"
	portalsURL    = "https://apigee.com/portals/api/sites"
	orgDataAPIURL = "https://apigee.com/dapi/api/organizations/%s"
)

// GetEnvironments - get the list of environments for the org
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

// CreateDeveloperApp - create an app for the developer
func (a *ApigeeClient) CreateDeveloperApp(newApp models.DeveloperApp) (*models.DeveloperApp, error) {
	// create a new developer app
	data, err := json.Marshal(newApp)
	if err != nil {
		return nil, err
	}

	response, err := a.newRequest(http.MethodPost, fmt.Sprintf(orgURL+"/developers/%s/apps", a.cfg.Organization, newApp.DeveloperId),
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

// GetDeveloperApp gets an app by name
func (a *ApigeeClient) GetDeveloperApp(name string) (*models.DeveloperApp, error) {
	url := fmt.Sprintf(orgURL+"/developers/%s/apps/%s", a.cfg.Organization, a.GetDeveloperID(), name)
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

// RemoveDeveloperApp - create an app for the developer
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

// GetProducts - get the list of products for the org
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

// GetPortals - get the list of portals for the org
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

// GetPortal - get the list of portals for the org
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

// GetPortalAPIs - get the list of portals for the org
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

// GetProduct - get details of the product
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

// GetImageWithURL - get the list of portals for the org
func (a *ApigeeClient) GetImageWithURL(imageURL, portalURL string) (string, string) {
	retries := 2
	for i := 0; i <= retries; i++ {
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
			continue
		}

		return base64.StdEncoding.EncodeToString(response.Body), contentType
	}
	return "", ""
}

// GetSpecContent - get the spec content for an api product
func (a *ApigeeClient) GetSpecContent(contentID string) []byte {
	retries := 2
	for i := 0; i <= retries; i++ {
		// Get the spec content file
		response, err := a.newRequest(http.MethodGet, fmt.Sprintf(orgDataAPIURL+"/specs/doc/%s/content", a.cfg.Organization, contentID),
			WithDefaultHeaders(),
		).Execute()

		if err != nil || response.Code != 200 {
			continue
		}

		return response.Body
	}
	return []byte{}
}

// GetRevisionSpec - gets the resource file of type openapi for the org, api, revision, and spec file specified
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

// GetSwagger - downloads the specfile from apigee given the url path of its location
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

// GetSharedFlow - gets the list of shared flows
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

// CreateSharedFlow - uploads an apigee bundle as a shared flow
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

// DeploySharedFlow - deploy the shared flow and revision to the environment
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

// PublishSharedFlowToEnvironment - publish the shared flow
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

// GetStats - get the api stats for a specific environment
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

func (a *ApigeeClient) GetAppCredential(appName, devID, key string) (*models.DeveloperAppCredentials, error) {
	url := fmt.Sprintf(orgURL+"/developers/%s/apps/%s/keys/%s", a.cfg.Organization, devID, appName, key)
	response, err := a.newRequest(
		http.MethodGet, url, WithDefaultHeaders(),
	).Execute()

	if err != nil {
		return nil, err
	}

	if response.Code != http.StatusOK {
		return nil, fmt.Errorf(
			"received an unexpected response code %d from Apigee while retrieving app credentials", response.Code,
		)
	}

	creds := &models.DeveloperAppCredentials{}
	err = json.Unmarshal(response.Body, creds)

	return creds, err
}

func (a *ApigeeClient) RemoveAppCredential(appName, devID, key string) error {
	url := fmt.Sprintf(orgURL+"/developers/%s/apps/%s/keys/%s", a.cfg.Organization, devID, appName, key)
	response, err := a.newRequest(
		http.MethodDelete, url, WithDefaultHeaders(),
	).Execute()

	if err != nil {
		return err
	}

	if response.Code != http.StatusOK {
		return fmt.Errorf(
			"received an unexpected response code %d from Apigee while removing app credentials", response.Code,
		)
	}

	return nil
}

func (a *ApigeeClient) UpdateAppCredential(appName, devID, key string, enable bool) error {
	url := fmt.Sprintf(orgURL+"/developers/%s/apps/%s/keys/%s", a.cfg.Organization, devID, appName, key)

	action := "revoke"
	if enable {
		action = "approve"
	}

	response, err := a.newRequest(
		http.MethodPost, url,
		WithDefaultHeaders(), WithQueryParam("action", action),
	).Execute()

	if err != nil {
		return err
	}

	if response.Code != http.StatusNoContent {
		return fmt.Errorf(
			"received an unexpected response code %d from Apigee while revoking/enabling app credentials", response.Code,
		)
	}

	return err
}

func (a *ApigeeClient) CreateAppCredential(appName, devID string, expDays int) (*models.DeveloperAppCredentials, error) {
	url := fmt.Sprintf(orgURL+"/developers/%s/apps/%s/keys/create", a.cfg.Organization, devID, appName)

	cred := &models.DeveloperAppCredentials{
		ConsumerKey:    util.RandString(35),
		ConsumerSecret: util.RandString(19),
	}

	if expDays > 0 {
		expTime := time.Now().Add(time.Duration(int64(time.Hour) * int64(24*expDays)))
		cred.ExpiresAt = int(expTime.UnixMilli())
	}

	credData, _ := json.Marshal(cred)

	response, err := a.newRequest(
		http.MethodPost, url, WithDefaultHeaders(), WithBody(credData),
	).Execute()

	if err != nil {
		return nil, err
	}

	if response.Code != http.StatusCreated {
		return nil, fmt.Errorf(
			"received an unexpected response code %d from Apigee while creating app credentials", response.Code,
		)
	}

	creds := &models.DeveloperAppCredentials{}
	err = json.Unmarshal(response.Body, creds)

	return creds, err
}

func (a *ApigeeClient) AddProductCredential(appName, devID, key string, cpr CredentialProvisionRequest) (*models.DeveloperAppCredentials, error) {
	data, err := json.Marshal(cpr)
	if err != nil {
		return nil, err
	}

	url := fmt.Sprintf(orgURL+"/developers/%s/apps/%s/keys/%s", a.cfg.Organization, devID, appName, key)

	response, err := a.newRequest(
		http.MethodPost, url, WithDefaultHeaders(),
		WithBody(data),
	).Execute()
	if err != nil {
		return nil, err
	}

	if response.Code != http.StatusOK {
		return nil, fmt.Errorf(
			"received an unexpected response code %d from Apigee while updating app credentials", response.Code,
		)
	}

	cred := &models.DeveloperAppCredentials{}
	err = json.Unmarshal(response.Body, cred)

	return cred, err
}

func (a *ApigeeClient) RemoveProductCredential(appName, devID, key, productName string) error {
	url := fmt.Sprintf(orgURL+"/developers/%s/apps/%s/keys/%s/apiproducts/%s", a.cfg.Organization, devID, appName, key, productName)

	response, err := a.newRequest(
		http.MethodDelete, url, WithDefaultHeaders(),
	).Execute()
	if err != nil {
		return err
	}

	if response.Code != http.StatusOK {
		return fmt.Errorf(
			"received an unexpected response code %d from Apigee while updating removing product from app credentials", response.Code,
		)
	}

	return err
}
