package apigee

import (
	"fmt"
	"testing"

	defs "github.com/Axway/agent-sdk/pkg/apic/definitions"
	"github.com/Axway/agent-sdk/pkg/apic/provisioning"
	prov "github.com/Axway/agent-sdk/pkg/apic/provisioning"
	"github.com/Axway/agent-sdk/pkg/apic/provisioning/mock"
	"github.com/Axway/agents-apigee/client/pkg/apigee"
	"github.com/Axway/agents-apigee/client/pkg/apigee/models"
	"github.com/stretchr/testify/assert"
)

func TestAccessRequestDeprovision(t *testing.T) {
	tests := []struct {
		name        string
		status      prov.Status
		appName     string
		apiID       string
		getAppErr   error
		rmCredErr   error
		missingCred bool
	}{
		{
			name:    "should deprovision an access request",
			appName: "app-one",
			apiID:   "abc-123",
			status:  prov.Success,
		},
		{
			name:      "should return success when the developer app is already removed",
			appName:   "app-one",
			apiID:     "abc-123",
			status:    prov.Success,
			getAppErr: fmt.Errorf("404"),
		},
		{
			name:      "should fail to deprovision an access request when retrieving the app, and the error is not a 404",
			appName:   "app-one",
			apiID:     "abc-123",
			getAppErr: fmt.Errorf("error"),
			status:    prov.Error,
		},
		{
			name:      "should fail to deprovision an access request when removing the credential",
			appName:   "app-one",
			apiID:     "abc-123",
			status:    prov.Error,
			rmCredErr: fmt.Errorf("error"),
		},
		{
			name:    "should return an error when the appName is not found",
			appName: "",
			apiID:   "abc-123",
			status:  prov.Error,
		},
		{
			name:    "should return an error when the apiID is not found",
			appName: "app-one",
			apiID:   "",
			status:  prov.Error,
		},
		{
			name:        "should return an error when the apiID is not found",
			appName:     "app-one",
			apiID:       "api-123",
			status:      prov.Error,
			missingCred: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			app := newApp(tc.apiID)

			p := NewProvisioner(&mockClient{
				t:           t,
				devID:       "dev-id-123",
				getAppErr:   tc.getAppErr,
				rmCredErr:   tc.rmCredErr,
				app:         app,
				appName:     tc.appName,
				key:         app.Credentials[0].ConsumerKey,
				productName: tc.apiID,
			})

			if tc.missingCred {
				app.Credentials = nil
			}

			mar := mock.MockAccessRequest{
				InstanceDetails: map[string]interface{}{
					defs.AttrExternalAPIID:    tc.apiID,
				},
				AppDetails: nil,
				AppName:    tc.appName,
			}

			status := p.AccessRequestDeprovision(&mar)
			assert.Equal(t, tc.status.String(), status.GetStatus().String())
			assert.Equal(t, 0, len(status.GetProperties()))
		})
	}
}

func TestAccessRequestProvision(t *testing.T) {
	tests := []struct {
		name        string
		status      prov.Status
		appName     string
		apiID       string
		getAppErr   error
		addCredErr  error
		noCreds     bool
		isApiLinked bool
	}{
		{
			name:    "should provision an access request",
			appName: "app-one",
			apiID:   "abc-123",
			status:  prov.Success,
		},
		{
			name:       "should fail to deprovision an access request",
			appName:    "app-one",
			apiID:      "abc-123",
			status:     prov.Error,
			addCredErr: fmt.Errorf("error"),
		},
		{
			name:      "should fail to deprovision when unable to retrieve the app",
			appName:   "app-one",
			apiID:     "abc-123",
			status:    prov.Error,
			getAppErr: fmt.Errorf("error"),
		},
		{
			name:    "should return an error when the apiID is not found",
			appName: "app-one",
			apiID:   "",
			status:  prov.Error,
		},
		{
			name:    "should return an error when the appName is not found",
			appName: "",
			apiID:   "abc-123",
			status:  prov.Error,
		},
		{
			name:    "should return an error when there are no credentials on the app",
			appName: "app-one",
			apiID:   "abc-123",
			status:  prov.Error,
			noCreds: true,
		},
		{
			name:        "should return an error when the api is already linked to a credential",
			appName:     "app-one",
			apiID:       "abc-123",
			status:      prov.Error,
			isApiLinked: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			apiID := ""
			if tc.isApiLinked {
				apiID = tc.apiID
			}
			app := newApp(apiID)

			key := app.Credentials[0].ConsumerKey

			if tc.noCreds {
				app.Credentials = nil
			}

			p := NewProvisioner(&mockClient{
				addCredErr:  tc.addCredErr,
				app:         app,
				appName:     tc.appName,
				devID:       "dev-id-123",
				key:         key,
				getAppErr:   tc.getAppErr,
				productName: tc.apiID,
				t:           t,
			})

			mar := mock.MockAccessRequest{
				InstanceDetails: map[string]interface{}{
					defs.AttrExternalAPIID:    tc.apiID,
				},
				AppDetails: nil,
				AppName:    tc.appName,
			}

			status := p.AccessRequestProvision(&mar)
			assert.Equal(t, tc.status.String(), status.GetStatus().String())
			assert.Equal(t, 0, len(status.GetProperties()))
		})
	}
}

func TestApplicationRequestDeprovision(t *testing.T) {
	tests := []struct {
		name     string
		status   prov.Status
		appName  string
		apiID    string
		rmAppErr error
	}{
		{
			name:    "should deprovision an application",
			status:  prov.Success,
			appName: "app-one",
			apiID:   "api-123",
		},
		{
			name:    "should return an error when the app name is not found",
			status:  prov.Error,
			appName: "",
			apiID:   "api-123",
		},
		{
			name:    "should return an error when the app name is not found",
			status:  prov.Error,
			appName: "",
			apiID:   "api-123",
		},
		{
			name:     "should return an error failing to remove the app",
			status:   prov.Error,
			appName:  "app-one",
			apiID:    "api-123",
			rmAppErr: fmt.Errorf("err"),
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			app := newApp(tc.apiID)

			p := NewProvisioner(&mockClient{
				app:         app,
				appName:     tc.appName,
				devID:       "dev-id-123",
				productName: tc.apiID,
				t:           t,
				rmAppErr:    tc.rmAppErr,
			})

			mar := mock.MockApplicationRequest{
				AppName:  tc.appName,
				TeamName: "team-one",
			}

			status := p.ApplicationRequestDeprovision(&mar)
			assert.Equal(t, tc.status.String(), status.GetStatus().String())
			assert.Equal(t, 0, len(status.GetProperties()))
		})
	}
}

func TestApplicationRequestProvision(t *testing.T) {
	tests := []struct {
		name         string
		status       prov.Status
		appName      string
		apiID        string
		createAppErr error
	}{
		{
			name:    "should provision an application",
			status:  prov.Success,
			appName: "app-one",
			apiID:   "api-123",
		},
		{
			name:         "should return an error when creating the app",
			status:       prov.Error,
			appName:      "app-one",
			apiID:        "api-123",
			createAppErr: fmt.Errorf("err"),
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			app := newApp(tc.apiID)

			p := NewProvisioner(&mockClient{
				app:          app,
				appName:      tc.appName,
				devID:        "dev-id-123",
				productName:  tc.apiID,
				t:            t,
				createAppErr: tc.createAppErr,
			})

			mar := mock.MockApplicationRequest{
				AppName:  tc.appName,
				TeamName: "team-one",
			}

			status := p.ApplicationRequestProvision(&mar)
			assert.Equal(t, tc.status.String(), status.GetStatus().String())
			assert.Equal(t, 0, len(status.GetProperties()))
		})
	}
}

func TestCredentialDeprovision(t *testing.T) {
	p := NewProvisioner(&mockClient{})
	mcr := mock.MockCredentialRequest{}
	status := p.CredentialDeprovision(mcr)
	assert.Equal(t, prov.Success, status.GetStatus())
	assert.NotEmpty(t, status.GetMessage())
}

func TestCredentialProvision(t *testing.T) {
	tests := []struct {
		name      string
		status    prov.Status
		appName   string
		apiID     string
		getAppErr error
		credType  string
	}{
		{
			name:     "should provision an api-key credential",
			status:   prov.Success,
			appName:  "app-one",
			apiID:    "api-123",
			credType: "api-key",
		},
		{
			name:     "should provision an oauth credential",
			status:   prov.Success,
			appName:  "app-one",
			apiID:    "api-123",
			credType: "oauth",
		},
		{
			name:     "should return an error when the app name is not found",
			status:   prov.Error,
			appName:  "",
			apiID:    "api-123",
			credType: "oauth",
		},
		{
			name:      "should return an error when unable to retrieve the app",
			status:    prov.Error,
			appName:   "app-one",
			apiID:     "api-123",
			getAppErr: fmt.Errorf("err"),
			credType:  "oauth",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			app := newApp(tc.apiID)

			p := NewProvisioner(&mockClient{
				app:         app,
				appName:     tc.appName,
				devID:       "dev-id-123",
				productName: tc.apiID,
				t:           t,
				getAppErr:   tc.getAppErr,
			})

			mcr := mock.MockCredentialRequest{
				AppName:     tc.appName,
				CredDefName: tc.credType,
			}

			status, cred := p.CredentialProvision(&mcr)
			if tc.status == prov.Error {
				assert.Nil(t, cred)
			} else {
				assert.NotNil(t, cred)
				if tc.credType == "oauth" {
					assert.Contains(t, cred.GetData(), provisioning.OauthClientID)
					assert.Contains(t, cred.GetData(), provisioning.OauthClientSecret)
					assert.NotContains(t, cred.GetData(), provisioning.APIKey)
				} else {
					assert.NotContains(t, cred.GetData(), provisioning.OauthClientID)
					assert.NotContains(t, cred.GetData(), provisioning.OauthClientSecret)
					assert.Contains(t, cred.GetData(), provisioning.APIKey)
				}
			}
			assert.Equal(t, tc.status.String(), status.GetStatus().String())
			assert.Equal(t, 0, len(status.GetProperties()))
		})
	}
}

type mockClient struct {
	addCredErr   error
	app          *models.DeveloperApp
	appName      string
	createAppErr error
	devID        string
	getAppErr    error
	key          string
	productName  string
	rmAppErr     error
	rmCredErr    error
	t            *testing.T
}

func (m mockClient) CreateDeveloperApp(newApp models.DeveloperApp) (*models.DeveloperApp, error) {
	assert.Equal(m.t, m.appName, newApp.Name)
	assert.Equal(m.t, m.devID, newApp.DeveloperId)
	return nil, m.createAppErr
}

func (m mockClient) RemoveDeveloperApp(appName, developerID string) error {
	assert.Equal(m.t, m.appName, appName)
	assert.Equal(m.t, m.devID, developerID)
	return m.rmAppErr
}

func (m mockClient) GetDeveloperID() string {
	return m.devID
}

func (m mockClient) GetDeveloperApp(name string) (*models.DeveloperApp, error) {
	assert.Equal(m.t, m.appName, name)
	return m.app, m.getAppErr
}

func (m mockClient) GetAppCredential(appName, devID, key string) (*models.DeveloperAppCredentials, error) {
	assert.Equal(m.t, m.appName, appName)
	assert.Equal(m.t, m.devID, devID)
	assert.Equal(m.t, m.key, key)
	return nil, nil
}

func (m mockClient) AddProductCredential(appName, devID, key string, cpr apigee.CredentialProvisionRequest) (*models.DeveloperAppCredentials, error) {
	assert.Equal(m.t, m.appName, appName)
	assert.Equal(m.t, m.devID, devID)
	assert.Equal(m.t, m.key, key)
	return nil, m.addCredErr
}

func (m mockClient) RemoveProductCredential(appName, devID, key, productName string) error {
	assert.Equal(m.t, m.appName, appName)
	assert.Equal(m.t, m.devID, devID)
	assert.Equal(m.t, m.key, key)
	assert.Equal(m.t, m.productName, productName)
	return m.rmCredErr
}

func newApp(apiID string) *models.DeveloperApp {
	cred := &models.DeveloperApp{
		Credentials: []models.DeveloperAppCredentials{
			{
				ApiProducts:    nil,
				ConsumerKey:    "consumer-key",
				ConsumerSecret: "consumer-secret",
			},
		},
	}

	if apiID != "" {
		cred.Credentials[0].ApiProducts = []models.ApiProductRef{
			{
				Apiproduct: apiID,
			},
		}
	}

	return cred
}
