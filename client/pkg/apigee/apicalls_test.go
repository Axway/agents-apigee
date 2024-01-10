package apigee

import (
	"encoding/json"
	"net/http"
	"testing"
	"time"

	"github.com/Axway/agent-sdk/pkg/api"
	"github.com/Axway/agents-apigee/client/pkg/apigee/models"
	"github.com/stretchr/testify/assert"
)

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

func TestUpdateDeveloperApp(t *testing.T) {
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
					RespCode: http.StatusOK,
					RespData: `"data":"aaaa"`,
				},
			},
			expectErr: true,
		},
		"success developer app updated": {
			responses: []api.MockResponse{
				{
					RespCode: http.StatusOK,
					RespData: string(appInData),
				},
			},
			appIn: appIn,
		},
	}
	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			c := createTestClient(t, &api.MockHTTPClient{Responses: tc.responses})

			appOut, err := c.UpdateDeveloperApp(tc.appIn)
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

func TestGetDeveloperApp(t *testing.T) {
	expectedApp := models.DeveloperApp{
		Name: "my-app",
	}
	expectedAppData, _ := json.Marshal(expectedApp)
	cases := map[string]struct {
		responses []api.MockResponse
		appName   string
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
					RespCode: http.StatusOK,
					RespData: `"data":"aaaa"`,
				},
			},
			expectErr: true,
		},
		"success getting developer app": {
			responses: []api.MockResponse{
				{
					RespCode: http.StatusOK,
					RespData: string(expectedAppData),
				},
			},
			appName: expectedApp.Name,
		},
	}
	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			c := createTestClient(t, &api.MockHTTPClient{Responses: tc.responses})

			appOut, err := c.GetDeveloperApp(tc.appName)
			if tc.expectErr {
				assert.NotNil(t, err)
				return
			}
			assert.Nil(t, err)
			assert.NotNil(t, appOut)
			assert.Equal(t, expectedApp.Name, appOut.Name)
		})
	}
}

func TestRemoveDeveloperApp(t *testing.T) {
	cases := map[string]struct {
		responses []api.MockResponse
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
		"success deleting developer app": {
			responses: []api.MockResponse{
				{
					RespCode: http.StatusOK,
				},
			},
		},
	}
	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			c := createTestClient(t, &api.MockHTTPClient{Responses: tc.responses})

			err := c.RemoveDeveloperApp("app", "dev")
			if tc.expectErr {
				assert.NotNil(t, err)
				return
			}
			assert.Nil(t, err)
		})
	}
}

func TestGetProducts(t *testing.T) {
	cases := map[string]struct {
		responses    []api.MockResponse
		expectedEnvs int
		expectErr    bool
	}{
		"products returned": {
			responses: []api.MockResponse{
				{
					RespData: `["prod1","prod2"]`,
					RespCode: http.StatusOK,
				},
			},
			expectedEnvs: 2,
		},
		"error getting products": {
			responses: []api.MockResponse{
				{
					ErrString: "error",
				},
			},
			expectErr: true,
		},
	}
	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			c := createTestClient(t, &api.MockHTTPClient{Responses: tc.responses})

			data, err := c.GetProducts()
			if tc.expectErr {
				assert.NotNil(t, err)
			}
			assert.Len(t, data, tc.expectedEnvs)
		})
	}
}

func TestGetProduct(t *testing.T) {
	expectedProd := models.ApiProduct{
		Name: "my-app",
	}
	expectedAppData, _ := json.Marshal(expectedProd)
	cases := map[string]struct {
		responses []api.MockResponse
		prodName  string
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
		"error, data returned not a product": {
			responses: []api.MockResponse{
				{
					RespCode: http.StatusOK,
					RespData: `"data":"aaaa"`,
				},
			},
			expectErr: true,
		},
		"success getting product": {
			responses: []api.MockResponse{
				{
					RespCode: http.StatusOK,
					RespData: string(expectedAppData),
				},
			},
			prodName: expectedProd.Name,
		},
	}
	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			c := createTestClient(t, &api.MockHTTPClient{Responses: tc.responses})

			prodOut, err := c.GetProduct(tc.prodName)
			if tc.expectErr {
				assert.NotNil(t, err)
				return
			}
			assert.Nil(t, err)
			assert.NotNil(t, prodOut)
			assert.Equal(t, expectedProd.Name, prodOut.Name)
		})
	}
}

func TestGetRevisionSpec(t *testing.T) {
	expectedSpec := []byte("spec data")
	cases := map[string]struct {
		responses    []api.MockResponse
		expectedSpec []byte
	}{
		"spec data returned": {
			responses: []api.MockResponse{
				{
					RespData: string(expectedSpec),
					RespCode: http.StatusOK,
				},
			},
			expectedSpec: expectedSpec,
		},
		"error getting environments": {
			responses: []api.MockResponse{
				{
					ErrString: "error",
				},
			},
			expectedSpec: []byte{},
		},
	}
	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			c := createTestClient(t, &api.MockHTTPClient{Responses: tc.responses})

			data := c.GetRevisionSpec("api", "rev", "spec")
			assert.Equal(t, tc.expectedSpec, data)
		})
	}
}

func TestGetStats(t *testing.T) {
	expectedStats := &models.Metrics{
		Environments: []models.MetricsEnvironments{{Name: "env1"}},
	}
	expectedStatsData, _ := json.Marshal(expectedStats)
	cases := map[string]struct {
		responses     []api.MockResponse
		expectedStats *models.Metrics
		expectErr     bool
	}{
		"stats data returned": {
			responses: []api.MockResponse{
				{
					RespData: string(expectedStatsData),
					RespCode: http.StatusOK,
				},
			},
			expectedStats: expectedStats,
		},
		"error, data returned not stats": {
			responses: []api.MockResponse{
				{
					RespCode: http.StatusOK,
					RespData: `"data":"aaaa"`,
				},
			},
			expectErr: true,
		},
		"error getting environments": {
			responses: []api.MockResponse{
				{
					ErrString: "error",
				},
			},
			expectErr: true,
		},
	}
	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			c := createTestClient(t, &api.MockHTTPClient{Responses: tc.responses})

			data, err := c.GetStats("env", "dim", "met", time.Now(), time.Now())
			if tc.expectErr {
				assert.NotNil(t, err)
				return
			}
			assert.Nil(t, err)
			assert.Equal(t, tc.expectedStats, data)
		})
	}
}

func TestCreateAPIProduct(t *testing.T) {
	prodIn := models.ApiProduct{
		Name: "my-prod",
	}
	prodInData, _ := json.Marshal(prodIn)
	cases := map[string]struct {
		responses []api.MockResponse
		prodIn    models.ApiProduct
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
		"error, data returned not a product": {
			responses: []api.MockResponse{
				{
					RespCode: http.StatusCreated,
					RespData: `"data":"aaaa"`,
				},
			},
			expectErr: true,
		},
		"success api product created": {
			responses: []api.MockResponse{
				{
					RespCode: http.StatusCreated,
					RespData: string(prodInData),
				},
			},
			prodIn: prodIn,
		},
	}
	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			c := createTestClient(t, &api.MockHTTPClient{Responses: tc.responses})

			appOut, err := c.CreateAPIProduct(&tc.prodIn)
			if tc.expectErr {
				assert.NotNil(t, err)
				return
			}
			assert.Nil(t, err)
			assert.NotNil(t, appOut)
			assert.Equal(t, prodIn.Name, appOut.Name)
		})
	}
}
