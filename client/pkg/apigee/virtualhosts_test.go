package apigee

import (
	"encoding/json"
	"fmt"
	"net/http"
	"testing"

	"github.com/Axway/agent-sdk/pkg/api"
	"github.com/Axway/agents-apigee/client/pkg/apigee/models"
	"github.com/stretchr/testify/assert"
)

func TestGetVirtualHost(t *testing.T) {
	expectedVHost := &models.VirtualHost{
		Name: "host",
	}
	expectedVHostData, _ := json.Marshal(expectedVHost)
	cases := map[string]struct {
		responses       []api.MockResponse
		expectedDetails *models.VirtualHost
		expectErr       bool
	}{
		"virtual host returned": {
			responses: []api.MockResponse{
				{
					RespData: string(expectedVHostData),
					RespCode: http.StatusOK,
				},
			},
			expectedDetails: expectedVHost,
		},
		"error, data returned not virtual host": {
			responses: []api.MockResponse{
				{
					RespCode: http.StatusOK,
					RespData: `"data":"aaaa"`,
				},
			},
			expectErr: true,
		},
		"error getting virtual host": {
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

			data, err := c.GetVirtualHost("env", "vhost")
			if tc.expectErr {
				assert.NotNil(t, err)
				return
			}
			assert.Nil(t, err)
			assert.Equal(t, data, tc.expectedDetails)
		})
	}
}

func TestGetAllEnvironmentVirtualHosts(t *testing.T) {
	cases := map[string]struct {
		responses   []api.MockResponse
		expectHosts []string
		expectErr   bool
	}{
		"error, data returned not virtual host list": {
			responses: []api.MockResponse{
				{
					RespData: `{"name":"host1"}`,
					RespCode: http.StatusOK,
				},
			},
			expectErr: true,
		},
		"success getting all but one virtual host, error on the one": {
			responses: []api.MockResponse{
				{
					RespData: `["host1","host2","host3"]`,
					RespCode: http.StatusOK,
				},
				{
					RespData: `{"name":"host1"}`,
					RespCode: http.StatusOK,
				},
				{
					RespData: `["host1","host2","host3"]`,
					RespCode: http.StatusOK,
				},
				{
					RespData: `{"name":"host3"}`,
					RespCode: http.StatusOK,
				},
			},
			expectHosts: []string{"host1", "host3"},
		},
		"virtual host returned": {
			responses: []api.MockResponse{
				{
					RespData: `["host1","host2","host3"]`,
					RespCode: http.StatusOK,
				},
				{
					RespData: `{"name":"host1"}`,
					RespCode: http.StatusOK,
				},
				{
					RespData: `{"name":"host2"}`,
					RespCode: http.StatusOK,
				},
				{
					RespData: `{"name":"host3"}`,
					RespCode: http.StatusOK,
				},
			},
			expectHosts: []string{"host1", "host2", "host3"},
		},
	}
	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			c := createTestClient(t, &api.MockHTTPClient{Responses: tc.responses})

			data, err := c.GetAllEnvironmentVirtualHosts("env")
			if tc.expectErr {
				assert.NotNil(t, err)
				return
			}
			assert.Nil(t, err)
			assert.NotNil(t, data)
			for _, host := range tc.expectHosts {
				found := false
				for _, vh := range data {
					if vh.Name == host {
						found = true
						break
					}
				}
				if !found {
					assert.Fail(t, fmt.Sprintf("vhost with name %s not found", host))
				}
			}
		})
	}
}
