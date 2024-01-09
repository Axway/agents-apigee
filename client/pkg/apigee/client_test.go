package apigee

import (
	"encoding/json"
	"net/http"
	"testing"

	"github.com/Axway/agent-sdk/pkg/api"
	"github.com/Axway/agents-apigee/client/pkg/apigee/models"
	"github.com/Axway/agents-apigee/client/pkg/config"
	"github.com/stretchr/testify/assert"
)

func createTestClient(t *testing.T, mockClient *api.MockHTTPClient) *ApigeeClient {
	cfg := &config.ApigeeConfig{
		URL:          "http://test.com",
		APIVersion:   "v1",
		Organization: "org",
		DeveloperID:  "test@dev.id",
		DataURL:      "http://data.com",
		Auth: &config.AuthConfig{
			BasicAuth: true,
			Username:  "user",
			Password:  "pass",
		},
	}

	c, err := NewClient(cfg)
	assert.Nil(t, err)
	assert.NotNil(t, c)
	assert.Equal(t, c.GetConfig(), cfg)
	c.apiClient = mockClient

	return c
}

func TestNewClient(t *testing.T) {
	c := createTestClient(t, nil)

	assert.Equal(t, c.GetDeveloperID(), "test@dev.id")
	assert.True(t, c.IsReady())
}

func TestGetEnvironments(t *testing.T) {
	cases := map[string]struct {
		responses    []api.MockResponse
		expectedEnvs int
	}{
		"environments returned": {
			responses: []api.MockResponse{
				{
					RespData: `["env1","env2"]`,
					RespCode: http.StatusOK,
				},
			},
			expectedEnvs: 2,
		},
		"error getting environments": {
			responses: []api.MockResponse{
				{
					ErrString: "error",
				},
			},
			expectedEnvs: 0,
		},
	}
	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			c := createTestClient(t, &api.MockHTTPClient{Responses: tc.responses})

			data := c.GetEnvironments()
			assert.Len(t, data, tc.expectedEnvs)
		})
	}
}

func TestCreateDeveloperApp(t *testing.T) {
	appIn := models.DeveloperApp{
		Name: "my-app",
	}
	appInData, _ := json.Marshal(appIn)
	cases := map[string]struct {
		responses []api.MockResponse
		appIn     models.DeveloperApp
		expectErr bool
	}{
		"error making http call": {
			responses: []api.MockResponse{
				{
					ErrString: "error",
				},
			},
			expectErr: true,
		},
		"error, unexpected response code": {
			responses: []api.MockResponse{
				{
					RespCode: http.StatusAccepted,
				},
			},
			expectErr: true,
		},
		"error, data returned not a developer app": {
			responses: []api.MockResponse{
				{
					RespCode: http.StatusCreated,
					RespData: `"data":"aaaa"`,
				},
			},
			expectErr: true,
		},
		"success developer app created": {
			responses: []api.MockResponse{
				{
					RespCode: http.StatusCreated,
					RespData: string(appInData),
				},
			},
			appIn: appIn,
		},
	}
	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			c := createTestClient(t, &api.MockHTTPClient{Responses: tc.responses})

			appOut, err := c.CreateDeveloperApp(tc.appIn)
			if tc.expectErr {
				assert.NotNil(t, err)
				return
			}
			assert.Nil(t, err)
			assert.NotNil(t, appOut)
			assert.Equal(t, appIn.Name, appOut.Name)
		})
	}
}
