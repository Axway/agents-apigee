package apigee

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sync"

	"github.com/Axway/agents-apigee/client/pkg/apigee/models"
)

// GetAllVirtualHosts - returns an array of all virtual hosts defined
func (a *ApigeeClient) GetAllEnvironmentVirtualHosts(envName string) ([]*models.VirtualHost, error) {
	// Get the spec file
	response, err := a.newRequest(http.MethodGet, fmt.Sprintf("%s/environments/%s", a.orgURL, envName),
		WithDefaultHeaders(),
	).Execute()
	if err != nil {
		return nil, err
	}

	hosts := VirtualHosts{}
	err = json.Unmarshal(response.Body, &hosts)
	if err != nil {
		return nil, err
	}

	virtualHostDetails := []*models.VirtualHost{}
	virtualHostLock := sync.Mutex{}
	wg := sync.WaitGroup{}
	for _, h := range hosts {
		wg.Add(1)
		go func(host string) {
			defer wg.Done()
			vh, err := a.GetVirtualHost(envName, host)
			if err != nil {
				return
			}
			virtualHostLock.Lock()
			defer virtualHostLock.Unlock()
			virtualHostDetails = append(virtualHostDetails, vh)
		}(h)
	}
	wg.Wait()

	return virtualHostDetails, nil
}

// GetAllVirtualHosts - returns an array of all virtual hosts defined
func (a *ApigeeClient) GetVirtualHost(envName, virtualHostName string) (*models.VirtualHost, error) {
	// Get the spec file
	response, err := a.newRequest(http.MethodGet, fmt.Sprintf("%s/environments/%s/virtualhosts/%s", a.orgURL, envName, virtualHostName),
		WithDefaultHeaders(),
	).Execute()
	if err != nil {
		return nil, err
	}

	virtualHost := &models.VirtualHost{}
	err = json.Unmarshal(response.Body, &virtualHost)
	if err != nil {
		return nil, err
	}
	return virtualHost, err
}
