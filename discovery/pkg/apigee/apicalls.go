package apigee

import (
	"encoding/base64"
	"encoding/json"
	"fmt"

	coreapi "github.com/Axway/agent-sdk/pkg/api"
	"github.com/Axway/agents-apigee/discovery/pkg/apigee/models"
)

const (
	orgURL        = "https://api.enterprise.apigee.com/v1/organizations/%s/"
	portalsURL    = "https://apigee.com/portals/api/sites"
	orgDataAPIURL = "https://apigee.com/dapi/api/organizations/%s"
)

func (a *GatewayClient) defaultHeaders() map[string]string {
	// return the default headers
	return map[string]string{
		"Content-Type":  "application/json",
		"Accept":        "application/json",
		"Authorization": "Bearer " + a.accessToken,
	}
}

func (a *GatewayClient) getRequest(url string) (*coreapi.Response, error) {
	// return the api response
	return a.getRequestWithQuery(url, map[string]string{})
}

func (a *GatewayClient) getRequestWithQuery(url string, queryParams map[string]string) (*coreapi.Response, error) {
	// create the get request
	request := coreapi.Request{
		Method:      coreapi.GET,
		URL:         url,
		Headers:     a.defaultHeaders(),
		QueryParams: queryParams,
	}

	// return the api response
	return a.apiClient.Send(request)
}

func (a *GatewayClient) postRequestWithQuery(url string, queryParams map[string]string, data []byte) (*coreapi.Response, error) {
	// create the post request
	request := coreapi.Request{
		Method:      coreapi.POST,
		URL:         url,
		Headers:     a.defaultHeaders(),
		QueryParams: queryParams,
	}

	// return the api response
	return a.apiClient.Send(request)
}

func (a *GatewayClient) putRequest(url string, data []byte) (*coreapi.Response, error) {
	// create the put request
	request := coreapi.Request{
		Method:  coreapi.PUT,
		URL:     url,
		Headers: a.defaultHeaders(),
		Body:    data,
	}

	// return the api response
	return a.apiClient.Send(request)
}

//getProducts - get the list of products for the org
func (a *GatewayClient) getProducts() products {
	// Get the products
	response, _ := a.getRequest(fmt.Sprintf(orgURL+"apiproducts", a.cfg.Organization))
	products := products{}
	json.Unmarshal(response.Body, &products)

	return products
}

//getPortals - get the list of portals for the org
func (a *GatewayClient) getPortals() []portalData {
	// Get the portals
	response, _ := a.getRequestWithQuery(portalsURL, map[string]string{"orgname": a.cfg.Organization})
	portalRes := portalsResponse{}
	json.Unmarshal(response.Body, &portalRes)

	return portalRes.Data
}

//getPortals - get the list of portals for the org
func (a *GatewayClient) getPortal(portalID string) portalData {
	// Get the portals
	response, _ := a.getRequest(fmt.Sprintf("%s/%s/portal", portalsURL, portalID))
	portalRes := portalResponse{}
	json.Unmarshal(response.Body, &portalRes)

	return portalRes.Data
}

//getPortalAPIs - get the list of portals for the org
func (a *GatewayClient) getPortalAPIs(portalID string) []*apiDocData {
	// Get the apidocs
	response, _ := a.getRequest(fmt.Sprintf("%s/%s/apidocs", portalsURL, portalID))
	apiDocRes := apiDocDataResponse{}
	json.Unmarshal(response.Body, &apiDocRes)

	return apiDocRes.Data
}

//getProduct - get details of the product
func (a *GatewayClient) getProduct(productName string) models.ApiProduct {
	// Get the product
	response, _ := a.getRequest(fmt.Sprintf(orgURL+"apiproducts/%s", a.cfg.Organization, productName))
	product := models.ApiProduct{}
	json.Unmarshal(response.Body, &product)

	return product
}

//getImageWithURL - get the list of portals for the org
func (a *GatewayClient) getImageWithURL(imageURL, portalURL string) (string, string) {
	// Get the portal
	request := coreapi.Request{
		Method: coreapi.GET,
		URL:    fmt.Sprintf("%s%s", portalURL, imageURL),
	}
	response, _ := a.apiClient.Send(request)

	contentType := ""
	if contentTypeArray, ok := response.Headers["Content-Type"]; ok {
		contentType = contentTypeArray[0]
		// assuming an octet stream type is actually an image/png
		if contentType == "application/octet-stream" {
			contentType = "image/png"
		}
	}

	if response.Code != 200 || string(response.Body) == "" || contentType == "" {
		return "", ""
	}

	return base64.StdEncoding.EncodeToString(response.Body), contentType
}

//getSpecContent - get the spec content for an api product
func (a *GatewayClient) getSpecContent(contentID string) []byte {
	// Get the spec content file
	response, _ := a.getRequest(fmt.Sprintf(orgDataAPIURL+"/specs/doc/%s/content", a.cfg.Organization, contentID))

	return response.Body
}

//getRevisionSpec - gets the resource file of type openapi for the org, api, revision, and spec file specified
func (a *GatewayClient) getRevisionSpec(apiName, revisionNumber, specFile string) []byte {
	// Get the openapi resource file
	response, _ := a.getRequest(fmt.Sprintf(orgURL+"apis/%s/revisions/%s/resourcefiles/openapi/%s", a.cfg.Organization, apiName, revisionNumber, specFile))

	return response.Body
}

//getSwagger - downloads the specfile from apigee given the url path of its location
func (a *GatewayClient) getSwagger(specPath string) []byte {
	// Get the spec file
	response, _ := a.getRequest(fmt.Sprintf("https://apigee.com%s", specPath))

	return response.Body
}
