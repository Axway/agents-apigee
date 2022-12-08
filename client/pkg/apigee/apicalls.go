package apigee

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/Axway/agents-apigee/client/pkg/apigee/models"
)

const (
	developerAppsURL = "%s/developers/%s/apps/%s"
)

// GetEnvironments - get the list of environments for the org
func (a *ApigeeClient) GetEnvironments() []string {
	// Get the developers
	response, err := a.newRequest(http.MethodGet, fmt.Sprintf("%s/environments", a.orgURL),
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

	response, err := a.newRequest(http.MethodPost, fmt.Sprintf("%s/developers/%s/apps", a.orgURL, newApp.DeveloperId),
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

// UpdateDeveloperApp - update an app for the developer
func (a *ApigeeClient) UpdateDeveloperApp(app models.DeveloperApp) (*models.DeveloperApp, error) {
	data, err := json.Marshal(app)
	if err != nil {
		return nil, err
	}

	response, err := a.newRequest(http.MethodPut, fmt.Sprintf(developerAppsURL, a.orgURL, app.DeveloperId, app.Name),
		WithDefaultHeaders(),
		WithBody(data),
	).Execute()
	if err != nil {
		return nil, err
	}
	if response.Code != http.StatusOK {
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
	url := fmt.Sprintf(developerAppsURL, a.orgURL, a.GetDeveloperID(), name)
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
	response, err := a.newRequest(http.MethodDelete, fmt.Sprintf(developerAppsURL, a.orgURL, developerID, appName),
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
func (a *ApigeeClient) GetProducts() (Products, error) {
	// Get the products
	response, err := a.newRequest(http.MethodGet, fmt.Sprintf("%s/apiproducts", a.orgURL),
		WithDefaultHeaders(),
	).Execute()
	if err != nil {
		return nil, err
	}

	products := Products{}
	if err == nil {
		json.Unmarshal(response.Body, &products)
	}

	return products, nil
}

// GetProduct - get details of the product
func (a *ApigeeClient) GetProduct(productName string) (*models.ApiProduct, error) {
	// Get the product
	response, err := a.newRequest(http.MethodGet, fmt.Sprintf("%s/apiproducts/%s", a.orgURL, productName),
		WithDefaultHeaders(),
	).Execute()
	if err != nil {
		return nil, err
	}

	if response.Code != http.StatusOK {
		return nil, fmt.Errorf("received an unexpected response code %d from Apigee when retrieving the app", response.Code)
	}

	product := &models.ApiProduct{}
	json.Unmarshal(response.Body, product)

	return product, nil
}

// GetRevisionSpec - gets the resource file of type openapi for the org, api, revision, and spec file specified
func (a *ApigeeClient) GetRevisionSpec(apiName, revisionNumber, specFile string) []byte {
	// Get the openapi resource file
	response, err := a.newRequest(http.MethodGet, fmt.Sprintf("%s/apis/%s/revisions/%s/resourcefiles/openapi/%s", a.orgURL, apiName, revisionNumber, specFile),
		WithDefaultHeaders(),
	).Execute()

	if err != nil {
		return []byte{}
	}

	return response.Body
}

// GetStats - get the api stats for a specific environment
func (a *ApigeeClient) GetStats(env, dimension, metricSelect string, start, end time.Time) (*models.Metrics, error) {
	// Get the spec content file
	const format = "01/02/2006 15:04"

	response, err := a.newRequest(http.MethodGet, fmt.Sprintf("%s/environments/%v/stats/%s", a.orgURL, env, dimension),
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

func (a *ApigeeClient) CreateAPIProduct(product *models.ApiProduct) (*models.ApiProduct, error) {
	// create a new developer app
	data, err := json.Marshal(product)
	if err != nil {
		return nil, err
	}

	u := fmt.Sprintf("%s/apiproducts", a.orgURL)
	response, err := a.newRequest(http.MethodPost, u,
		WithDefaultHeaders(),
		WithBody(data),
	).Execute()

	if err != nil {
		return nil, err
	}

	if response.Code != http.StatusCreated {
		return nil, fmt.Errorf("received an unexpected response code %d from Apigee when creating the api product", response.Code)
	}

	newProduct := models.ApiProduct{}
	err = json.Unmarshal(response.Body, &newProduct)
	if err != nil {
		return nil, err
	}

	return &newProduct, err

}
