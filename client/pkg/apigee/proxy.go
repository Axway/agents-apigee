package apigee

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/Axway/agents-apigee/client/pkg/apigee/models"
)

// Products
type Proxies []string

// GetAllProxies - get all proxies
func (a *ApigeeClient) GetAllProxies() Proxies {
	response, err := a.newRequest(http.MethodGet, fmt.Sprintf(orgURL+"/apis", a.cfg.Organization),
		WithDefaultHeaders(),
	).Execute()

	proxies := Proxies{}
	if err == nil {
		json.Unmarshal(response.Body, &proxies)
	}

	return proxies
}

// GetProxy - get a proxy with a name
func (a *ApigeeClient) GetProxy(apiName string) models.ApiProxy {
	response, err := a.newRequest(http.MethodGet, fmt.Sprintf(orgURL+"/apis/"+apiName, a.cfg.Organization),
		WithDefaultHeaders(),
	).Execute()

	proxy := models.ApiProxy{}
	if err == nil {
		json.Unmarshal(response.Body, &proxy)
	}

	return proxy
}

// GetDeployments - get a deployments for a proxy
func (a *ApigeeClient) GetDeployments(apiName string) models.ApiProxy {
	response, err := a.newRequest(http.MethodGet, fmt.Sprintf(orgURL+"/apis/"+apiName+"/deployments", a.cfg.Organization),
		WithDefaultHeaders(),
	).Execute()

	proxy := models.ApiProxy{}
	if err == nil {
		json.Unmarshal(response.Body, &proxy)
	}

	return proxy
}
