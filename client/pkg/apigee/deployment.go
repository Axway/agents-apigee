package apigee

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/Axway/agents-apigee/client/pkg/apigee/models"
)

// GetDeployments - get a deployments for a proxy
func (a *ApigeeClient) GetDeployments(proxyName string) (*models.DeploymentDetails, error) {
	response, err := a.newRequest(http.MethodGet, fmt.Sprintf("%s/apis/%s/deployments", a.orgURL, proxyName),
		WithDefaultHeaders(),
	).Execute()

	if err != nil {
		return nil, err
	}

	details := &models.DeploymentDetails{}
	json.Unmarshal(response.Body, details)
	if err != nil {
		return nil, err
	}

	return details, nil
}
