package apigee

import (
	"fmt"
	"testing"
	"time"

	v1 "github.com/Axway/agent-sdk/pkg/apic/apiserver/models/api/v1"
	management "github.com/Axway/agent-sdk/pkg/apic/apiserver/models/management/v1alpha1"
	defs "github.com/Axway/agent-sdk/pkg/apic/definitions"
	"github.com/Axway/agent-sdk/pkg/apic/provisioning"
	"github.com/Axway/agent-sdk/pkg/apic/provisioning/mock"
	"github.com/Axway/agent-sdk/pkg/util"
	"github.com/Axway/agents-apigee/client/pkg/apigee"
	"github.com/Axway/agents-apigee/client/pkg/apigee/models"
	"github.com/stretchr/testify/assert"
)

func TestAccessRequestDeprovision(t *testing.T) {
	tests := []struct {
		name        string
		status      provisioning.Status
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
			status:  provisioning.Success,
		},
		{
			name:      "should return success when the developer app is already removed",
			appName:   "app-one",
			apiID:     "abc-123",
			status:    provisioning.Success,
			getAppErr: fmt.Errorf("404"),
		},
		{
			name:      "should fail to deprovision an access request when retrieving the app, and the error is not a 404",
			appName:   "app-one",
			apiID:     "abc-123",
			getAppErr: fmt.Errorf("error"),
			status:    provisioning.Error,
		},
		{
			name:      "should fail to deprovision an access request when removing the credential",
			appName:   "app-one",
			apiID:     "abc-123",
			status:    provisioning.Error,
			rmCredErr: fmt.Errorf("error"),
		},
		{
			name:    "should return an error when the appName is not found",
			appName: "",
			apiID:   "abc-123",
			status:  provisioning.Error,
		},
		{
			name:    "should return an error when the apiID is not found",
			appName: "app-one",
			apiID:   "",
			status:  provisioning.Error,
		},
		{
			name:        "should return an error when the apiID is not found",
			appName:     "app-one",
			apiID:       "api-123",
			status:      provisioning.Error,
			missingCred: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			app := newApp(tc.apiID, tc.appName)

			p := NewProvisioner(&mockClient{
				t:           t,
				devID:       "dev-id-123",
				getAppErr:   tc.getAppErr,
				rmCredErr:   tc.rmCredErr,
				app:         app,
				appName:     tc.appName,
				key:         app.Credentials[0].ConsumerKey,
				productName: tc.apiID,
			}, 30, &mockCache{t: t})

			if tc.missingCred {
				app.Credentials = nil
			}

			mar := mock.MockAccessRequest{
				InstanceDetails: map[string]interface{}{
					defs.AttrExternalAPIID: tc.apiID,
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
		name         string
		status       provisioning.Status
		appName      string
		apiID        string
		apiStage     string
		getAppErr    error
		addCredErr   error
		addProdErr   error
		existingProd bool
		noCreds      bool
		isApiLinked  bool
	}{
		{
			name:     "should provision an access request",
			appName:  "app-one",
			apiID:    "abc-123",
			apiStage: "prod",
			status:   provisioning.Success,
		},
		{
			name:     "should provision an access request when there are no credentials on the app",
			appName:  "app-one",
			apiID:    "abc-123",
			apiStage: "prod",
			status:   provisioning.Success,
			noCreds:  true,
		},
		{
			name:        "should provision an access request when the api is already linked to a credential",
			appName:     "app-one",
			apiID:       "abc-123",
			apiStage:    "prod",
			status:      provisioning.Success,
			isApiLinked: true,
		},
		{
			name:       "should fail to deprovision an access request",
			appName:    "app-one",
			apiID:      "abc-123",
			apiStage:   "prod",
			status:     provisioning.Error,
			addCredErr: fmt.Errorf("error"),
		},
		{
			name:      "should fail to deprovision when unable to retrieve the app",
			appName:   "app-one",
			apiID:     "abc-123",
			apiStage:  "prod",
			status:    provisioning.Error,
			getAppErr: fmt.Errorf("error"),
		},
		{
			name:     "should return an error when the apiID is not found",
			appName:  "app-one",
			apiID:    "",
			apiStage: "prod",
			status:   provisioning.Error,
		},
		{
			name:     "should return an error when the appName is not found",
			appName:  "",
			apiID:    "abc-123",
			apiStage: "prod",
			status:   provisioning.Error,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			apiID := ""
			if tc.isApiLinked {
				apiID = tc.apiID
			}
			app := newApp(apiID, tc.appName)

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
			}, 30, &mockCache{t: t})

			mar := mock.MockAccessRequest{
				InstanceDetails: map[string]interface{}{
					defs.AttrExternalAPIID:    tc.apiID,
					defs.AttrExternalAPIStage: tc.apiStage,
				},
				AppDetails: nil,
				AppName:    tc.appName,
			}

			status, _ := p.AccessRequestProvision(&mar)
			assert.Equal(t, tc.status.String(), status.GetStatus().String())
			if tc.status == provisioning.Success {
				assert.Equal(t, 1, len(status.GetProperties()))
			} else {
				assert.Equal(t, 0, len(status.GetProperties()))
			}
		})
	}
}

func TestApplicationRequestDeprovision(t *testing.T) {
	tests := []struct {
		name     string
		status   provisioning.Status
		appName  string
		apiID    string
		rmAppErr error
	}{
		{
			name:    "should deprovision an application",
			status:  provisioning.Success,
			appName: "app-one",
			apiID:   "api-123",
		},
		{
			name:    "should return an error when the app name is not found",
			status:  provisioning.Error,
			appName: "",
			apiID:   "api-123",
		},
		{
			name:    "should return an error when the app name is not found",
			status:  provisioning.Error,
			appName: "",
			apiID:   "api-123",
		},
		{
			name:     "should return an error failing to remove the app",
			status:   provisioning.Error,
			appName:  "app-one",
			apiID:    "api-123",
			rmAppErr: fmt.Errorf("err"),
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			app := newApp(tc.apiID, tc.appName)

			p := NewProvisioner(&mockClient{
				app:         app,
				appName:     tc.appName,
				key:         "key",
				devID:       "dev-id-123",
				productName: tc.apiID,
				t:           t,
				rmAppErr:    tc.rmAppErr,
			}, 30, &mockCache{t: t})

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
		status       provisioning.Status
		appName      string
		apiID        string
		createAppErr error
	}{
		{
			name:    "should provision an application",
			status:  provisioning.Success,
			appName: "app-one",
			apiID:   "api-123",
		},
		{
			name:         "should return an error when creating the app",
			status:       provisioning.Error,
			appName:      "app-one",
			apiID:        "api-123",
			createAppErr: fmt.Errorf("err"),
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			app := newApp(tc.apiID, tc.appName)

			p := NewProvisioner(&mockClient{
				app:          app,
				appName:      tc.appName,
				devID:        "dev-id-123",
				productName:  tc.apiID,
				t:            t,
				createAppErr: tc.createAppErr,
			}, 30, &mockCache{t: t})

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
	tests := []struct {
		name      string
		status    provisioning.Status
		appName   string
		apiID     string
		getAppErr error
		credType  string
	}{
		{
			name:     "should deprovision an api-key credential",
			status:   provisioning.Success,
			appName:  "app-one",
			apiID:    "api-123",
			credType: "api-key",
		},
		{
			name:     "should deprovision an oauth credential",
			status:   provisioning.Success,
			appName:  "app-one",
			apiID:    "api-123",
			credType: "oauth",
		},
		{
			name:      "should return success when unable to retrieve the app",
			status:    provisioning.Success,
			appName:   "app-one",
			apiID:     "api-123",
			getAppErr: fmt.Errorf("err"),
			credType:  "oauth",
		},
		{
			name:    "should return success when credential not on app",
			status:  provisioning.Success,
			appName: "app-one",
			apiID:   "api-123",
		},
		{
			name:     "should return an error when the app name is not found",
			status:   provisioning.Error,
			appName:  "",
			apiID:    "api-123",
			credType: "oauth",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			app := newApp(tc.apiID, tc.appName)

			key := "consumer-key"
			if tc.credType == "" {
				key = ""
			}
			p := NewProvisioner(&mockClient{
				app:         app,
				appName:     tc.appName,
				key:         key,
				devID:       "dev-id-123",
				productName: tc.apiID,
				t:           t,
				getAppErr:   tc.getAppErr,
			}, 30, &mockCache{t: t, appName: tc.appName})

			thisHash, _ := util.ComputeHash(key)
			mcr := mock.MockCredentialRequest{
				AppName:     tc.appName,
				CredDefName: tc.credType,
				Details: map[string]string{
					appRefName: tc.appName,
					credRefKey: fmt.Sprintf("%v", thisHash),
				},
			}

			status := p.CredentialDeprovision(&mcr)
			assert.Equal(t, tc.status.String(), status.GetStatus().String())
		})
	}
}

func TestCredentialProvision(t *testing.T) {
	tests := []struct {
		name      string
		status    provisioning.Status
		appName   string
		apiID     string
		getAppErr error
		credType  string
	}{
		{
			name:     "should provision an api-key credential",
			status:   provisioning.Success,
			appName:  "app-one",
			apiID:    "api-123",
			credType: "api-key",
		},
		{
			name:     "should provision an oauth credential",
			status:   provisioning.Success,
			appName:  "app-one",
			apiID:    "api-123",
			credType: "oauth",
		},
		{
			name:     "should return an error when the app name is not found",
			status:   provisioning.Error,
			appName:  "",
			apiID:    "api-123",
			credType: "oauth",
		},
		{
			name:      "should return an error when unable to retrieve the app",
			status:    provisioning.Error,
			appName:   "app-one",
			apiID:     "api-123",
			getAppErr: fmt.Errorf("err"),
			credType:  "oauth",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			app := newApp(tc.apiID, tc.appName)

			p := NewProvisioner(&mockClient{
				app:         app,
				appName:     tc.appName,
				devID:       "dev-id-123",
				productName: tc.apiID,
				t:           t,
				getAppErr:   tc.getAppErr,
			}, 30, &mockCache{t: t, appName: tc.appName})

			mcr := mock.MockCredentialRequest{
				AppName:     tc.appName,
				CredDefName: tc.credType,
			}

			status, cred := p.CredentialProvision(&mcr)
			if tc.status == provisioning.Error {
				assert.Nil(t, cred)
				assert.Equal(t, 0, len(status.GetProperties()))
			} else {
				assert.NotNil(t, cred)
				assert.Equal(t, 2, len(status.GetProperties()))
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
		})
	}
}

func TestCredentialUpdate(t *testing.T) {
	tests := []struct {
		name           string
		status         provisioning.Status
		appName        string
		apiID          string
		getAppErr      error
		credType       string
		skipAppDetail  bool
		skipCredDetail bool
		action         provisioning.CredentialAction
	}{
		{
			name:     "should revoke credential",
			status:   provisioning.Success,
			appName:  "app-one",
			apiID:    "api-123",
			credType: "api-key",
			action:   provisioning.Suspend,
		},
		{
			name:     "should enable credential",
			status:   provisioning.Success,
			appName:  "app-one",
			apiID:    "api-123",
			credType: "api-key",
			action:   provisioning.Enable,
		},
		{
			name:          "should return an error when unable to get appName detail",
			status:        provisioning.Error,
			appName:       "app-one",
			skipAppDetail: true,
			apiID:         "api-123",
		},
		{
			name:           "should return an error when unable to get cred ref detail",
			status:         provisioning.Error,
			appName:        "app-one",
			skipCredDetail: true,
			apiID:          "api-123",
		},
		{
			name:      "should return an error when unable to retrieve the app",
			status:    provisioning.Error,
			appName:   "app-one",
			apiID:     "api-123",
			getAppErr: fmt.Errorf("err"),
		},
		{
			name:     "should return an error when credential action unknown",
			status:   provisioning.Error,
			appName:  "app-one",
			apiID:    "api-123",
			credType: "api-key",
			action:   provisioning.Rotate,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			app := newApp(tc.apiID, tc.appName)

			key := "consumer-key"
			if tc.credType == "" {
				key = ""
			}
			p := NewProvisioner(&mockClient{
				app:         app,
				appName:     tc.appName,
				key:         key,
				devID:       "dev-id-123",
				productName: tc.apiID,
				t:           t,
				getAppErr:   tc.getAppErr,
				enable:      tc.action == provisioning.Enable,
			}, 30, &mockCache{t: t, appName: tc.appName})

			thisHash, _ := util.ComputeHash(key)
			details := map[string]string{
				appRefName: tc.appName,
				credRefKey: fmt.Sprintf("%v", thisHash),
			}
			if tc.skipAppDetail {
				delete(details, appRefName)
			}
			if tc.skipCredDetail {
				delete(details, credRefKey)
			}

			mcr := mock.MockCredentialRequest{
				AppName:     tc.appName,
				CredDefName: tc.credType,
				Details:     details,
				Action:      tc.action,
			}

			status, _ := p.CredentialUpdate(&mcr)
			assert.Equal(t, tc.status.String(), status.GetStatus().String())
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
	enable       bool
	existingProd bool
	t            *testing.T
}

func (m mockClient) CreateDeveloperApp(newApp models.DeveloperApp) (*models.DeveloperApp, error) {
	assert.Equal(m.t, m.appName, newApp.Name)
	assert.Equal(m.t, m.devID, newApp.DeveloperId)
	return &models.DeveloperApp{
		Credentials: []models.DeveloperAppCredentials{
			{
				ConsumerKey:    m.key,
				ConsumerSecret: "secret",
			},
		},
	}, m.createAppErr
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

func (m mockClient) RemoveAppCredential(appName, devID, key string) error {
	assert.Equal(m.t, m.appName, appName)
	assert.Equal(m.t, m.devID, devID)
	assert.Equal(m.t, m.key, key)
	return nil
}

func (m mockClient) CreateAppCredential(appName, devID string, products []string, expDays int) (*models.DeveloperApp, error) {
	return &models.DeveloperApp{
		Credentials: []models.DeveloperAppCredentials{
			{
				ExpiresAt: int(time.Now().Add(time.Duration(int64(time.Hour) * int64(24*expDays))).UnixMilli()),
			},
		},
	}, nil
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

func (m mockClient) UpdateAppCredential(appName, devID, key string, enable bool) error {
	assert.Equal(m.t, m.appName, appName)
	assert.Equal(m.t, m.devID, devID)
	assert.Equal(m.t, m.enable, enable)
	return nil
}

func (m mockClient) CreateAPIProduct(product *models.ApiProduct) (*models.ApiProduct, error) {
	if !m.existingProd {
		assert.Equal(m.t, m.productName, product.Name)
		return &models.ApiProduct{
			Name: m.productName,
		}, nil
	}
	return nil, nil
}

func (m mockClient) GetProduct(productName string) (*models.ApiProduct, error) {
	if m.existingProd {
		return &models.ApiProduct{
			Name: m.productName,
		}, nil
	}
	return nil, fmt.Errorf("error")
}

func (m mockClient) UpdateDeveloperApp(app models.DeveloperApp) (*models.DeveloperApp, error) {
	return nil, nil
}

func newApp(apiID string, appName string) *models.DeveloperApp {
	cred := &models.DeveloperApp{
		Credentials: []models.DeveloperAppCredentials{
			{
				ApiProducts:    nil,
				ConsumerKey:    "consumer-key",
				ConsumerSecret: "consumer-secret",
			},
		},
		Name: appName,
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

type mockCache struct {
	t       *testing.T
	appName string
	apiName string
}

func (m *mockCache) GetAccessRequestsByApp(managedAppName string) []*v1.ResourceInstance {
	assert.Equal(m.t, m.appName, managedAppName)
	ar1 := management.NewAccessRequest("ar1", "env")
	ar1.Spec.ManagedApplication = m.appName
	util.SetAgentDetailsKey(ar1, prodNameRef, "product")
	ri, _ := ar1.AsInstance()
	return []*v1.ResourceInstance{ri}
}

func (m *mockCache) GetAPIServiceInstanceByName(apiName string) (*v1.ResourceInstance, error) {
	assert.Equal(m.t, m.apiName, apiName)
	apisi := management.NewAPIServiceInstance(apiName, "env")
	util.SetAgentDetailsKey(apisi, defs.AttrExternalAPIID, "apiName")
	ri, _ := apisi.AsInstance()
	return ri, nil
}
