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
func (a *ApigeeClient) GetAllProxies() (Proxies, error) {
	response, err := a.newRequest(http.MethodGet, fmt.Sprintf("%s/apis", a.orgURL),
		WithDefaultHeaders(),
	).Execute()
	if err != nil {
		return nil, err
	}

	proxies := Proxies{}
	err = json.Unmarshal(response.Body, &proxies)
	if err != nil {
		return nil, err
	}

	return proxies, nil
}

// GetProxy - get a proxy with a name
func (a *ApigeeClient) GetProxy(proxyName string) (*models.ApiProxy, error) {
	response, err := a.newRequest(http.MethodGet, fmt.Sprintf("%s/apis/%s", a.orgURL, proxyName),
		WithDefaultHeaders(),
	).Execute()
	if err != nil {
		return nil, err
	}

	proxy := &models.ApiProxy{}
	err = json.Unmarshal(response.Body, proxy)
	if err != nil {
		return nil, err
	}

	return proxy, nil
}

// GetProxy - get a proxy with a name
func (a *ApigeeClient) GetRevision(proxyName, revision string) (*models.ApiProxyRevision, error) {
	response, err := a.newRequest(http.MethodGet, fmt.Sprintf("%s/apis/%s/revisions/%s", a.orgURL, proxyName, revision),
		WithDefaultHeaders(),
	).Execute()
	if err != nil {
		return nil, err
	}

	proxyRevision := &models.ApiProxyRevision{}
	json.Unmarshal(response.Body, proxyRevision)
	if err != nil {
		return nil, err
	}

	return proxyRevision, nil
}

// GetProxy - get a proxy with a name
func (a *ApigeeClient) GetRevisionResourceFile(proxyName, revision, resourceType, resourceName string) ([]byte, error) {
	response, err := a.newRequest(http.MethodGet, fmt.Sprintf("%s/apis/%s/revisions/%s/resourcefiles/%s/%s", a.orgURL, proxyName, revision, resourceType, resourceName),
		WithDefaultHeaders(),
	).Execute()
	if err != nil {
		return nil, err
	}

	return response.Body, nil
}

// GetRevisionPolicyByName - get the details about a named policy on a revision
func (a *ApigeeClient) GetRevisionPolicyByName(proxyName, revision, policyName string) (*PolicyDetail, error) {
	response, err := a.newRequest(http.MethodGet, fmt.Sprintf("%s/apis/%s/revisions/%s/policies/%s", a.orgURL, proxyName, revision, policyName),
		WithDefaultHeaders(),
	).Execute()
	if err != nil {
		return nil, err
	}

	policyDetails := &PolicyDetail{}
	json.Unmarshal(response.Body, policyDetails)
	if err != nil {
		return nil, err
	}

	return policyDetails, nil
}
