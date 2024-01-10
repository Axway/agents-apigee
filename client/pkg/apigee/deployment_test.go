package apigee

import (
	"encoding/json"
	"net/http"
	"testing"

	"github.com/Axway/agent-sdk/pkg/api"
	"github.com/Axway/agents-apigee/client/pkg/apigee/models"
	"github.com/stretchr/testify/assert"
)

func TestGetDeployments(t *testing.T) {
	expectedDetails := &models.DeploymentDetails{
		Name: "deployment",
	}
	expectedDetailsData, _ := json.Marshal(expectedDetails)
	cases := map[string]struct {
		responses       []api.MockResponse
		expectedDetails *models.DeploymentDetails
		expectErr       bool
	}{
		"deployment details returned": {
			responses: []api.MockResponse{
				{
					RespData: string(expectedDetailsData),
					RespCode: http.StatusOK,
				},
			},
			expectedDetails: expectedDetails,
		},
		"error, data returned not deployment details": {
			responses: []api.MockResponse{
				{
					RespCode: http.StatusOK,
					RespData: `"data":"aaaa"`,
				},
			},
			expectErr: true,
		},
		"error getting deployment details": {
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

			data, err := c.GetDeployments(expectedDetails.Name)
			if tc.expectErr {
				assert.NotNil(t, err)
				return
			}
			assert.Nil(t, err)
			assert.Equal(t, data, tc.expectedDetails)
		})
	}
}
