package apigee

import (
	"encoding/json"
	"fmt"
	"net/http"
)

// GetSpecFile - downloads the specfile from apigee given the path of its location
func (a *ApigeeClient) GetSpecFile(specPath string) ([]byte, error) {
	// Get the spec file
	response, err := a.newRequest(http.MethodGet, fmt.Sprintf("%s%s", a.dataURL, specPath),
		WithDefaultHeaders(),
	).Execute()

	if err != nil {
		return nil, err
	}

	return response.Body, nil
}

// GetSpecFromURL - downloads the specfile from a URL outside of APIGEE
func (a *ApigeeClient) GetSpecFromURL(url string, options ...RequestOption) ([]byte, error) {
	// Get the spec file
	response, err := a.newRequest(http.MethodGet, url, options...).Execute()

	if err != nil {
		return nil, err
	}

	return response.Body, nil
}

// GetAllSpecs - downloads the specfile from apigee given the path of its location
func (a *ApigeeClient) GetAllSpecs() ([]SpecDetails, error) {
	// Get the spec file
	response, err := a.newRequest(http.MethodGet, fmt.Sprintf("%s/organizations/%s/specs/folder/home", a.dataURL, a.cfg.Organization),
		WithDefaultHeaders(),
	).Execute()

	if err != nil {
		return nil, err
	}

	details := SpecDetails{}
	err = json.Unmarshal(response.Body, &details)
	if err != nil {
		return nil, err
	}

	return details.Contents, nil
}
