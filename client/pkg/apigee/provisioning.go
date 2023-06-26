package apigee

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/Axway/agents-apigee/client/pkg/apigee/models"
)

const (
	developerAppsKeysURL = "%s/developers/%s/apps/%s/keys/%s"
)

func (a *ApigeeClient) GetAppCredential(appName, devID, key string) (*models.DeveloperAppCredentials, error) {
	url := fmt.Sprintf(developerAppsKeysURL, a.orgURL, devID, appName, key)
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
	url := fmt.Sprintf(developerAppsKeysURL, a.orgURL, devID, appName, key)
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
	url := fmt.Sprintf(developerAppsKeysURL, a.orgURL, devID, appName, key)

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

func (a *ApigeeClient) CreateAppCredential(appName, devID string, products []string, expDays int) (*models.DeveloperApp, error) {
	url := fmt.Sprintf("%s/developers/%s/apps/%s", a.orgURL, devID, appName)

	appCredReq := CredentialProvisionRequest{
		ApiProducts: products,
	}
	if expDays > 0 {
		expTime := time.Duration(int64(time.Hour) * int64(24*expDays))
		appCredReq.KeyExpiresIn = int(expTime.Milliseconds())
	}

	credData, _ := json.Marshal(appCredReq)

	response, err := a.newRequest(
		http.MethodPost, url, WithDefaultHeaders(), WithBody(credData),
	).Execute()

	if err != nil {
		return nil, err
	}

	if response.Code != http.StatusOK {
		return nil, fmt.Errorf(
			"received an unexpected response code %d from Apigee while creating app credentials", response.Code,
		)
	}

	appData := &models.DeveloperApp{}
	err = json.Unmarshal(response.Body, appData)

	return appData, err
}

func (a *ApigeeClient) AddCredentialProduct(appName, devID, key string, cpr CredentialProvisionRequest) (*models.DeveloperAppCredentials, error) {
	data, err := json.Marshal(cpr)
	if err != nil {
		return nil, err
	}

	url := fmt.Sprintf(developerAppsKeysURL, a.orgURL, devID, appName, key)

	response, err := a.newRequest(
		http.MethodPost, url, WithDefaultHeaders(),
		WithBody(data),
	).Execute()
	if err != nil {
		return nil, err
	}

	if response.Code != http.StatusOK {
		return nil, fmt.Errorf(
			"received an unexpected response code %d from Apigee while adding a product to an app credentials: %s", response.Code, response.Body,
		)
	}

	cred := &models.DeveloperAppCredentials{}
	err = json.Unmarshal(response.Body, cred)

	return cred, err
}

func (a *ApigeeClient) RemoveCredentialProduct(appName, devID, key, productName string) error {
	url := fmt.Sprintf("%s/developers/%s/apps/%s/keys/%s/apiproducts/%s", a.orgURL, devID, appName, key, productName)

	response, err := a.newRequest(
		http.MethodDelete, url, WithDefaultHeaders(),
	).Execute()
	if err != nil {
		return err
	}

	if response.Code != http.StatusOK {
		return fmt.Errorf(
			"received an unexpected response code %d from Apigee while removing product from an app credentials", response.Code,
		)
	}

	return err
}

func (a *ApigeeClient) UpdateCredentialProduct(appName, devID, key, productName string, enable bool) error {
	url := fmt.Sprintf("%s/developers/%s/apps/%s/keys/%s/apiproducts/%s", a.orgURL, devID, appName, key, productName)

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
			"received an unexpected response code %d from Apigee while updating a product on an app credentials", response.Code,
		)
	}

	return err
}
