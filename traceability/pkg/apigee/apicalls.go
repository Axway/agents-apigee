package apigee

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"

	coreapi "github.com/Axway/agent-sdk/pkg/api"
	"github.com/Axway/agents-apigee/traceability/pkg/apigee/models"
)

const (
	orgURL = "https://api.enterprise.apigee.com/v1/organizations/%s/"
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

//getSharedFlow - gets the list of shared flows
func (a *GatewayClient) getSharedFlow(name string) (*models.SharedFlowRevisionDeploymentDetails, error) {
	// Get the shared flows list
	response, err := a.getRequest(fmt.Sprintf(orgURL+"/sharedflows/%v", a.cfg.Organization, name))
	if err != nil {
		return nil, err
	}

	if response.Code == http.StatusNotFound {
		return nil, fmt.Errorf("could not find shared flow named %v", name)
	}

	flow := models.SharedFlowRevisionDeploymentDetails{}
	json.Unmarshal(response.Body, &flow)

	return &flow, nil
}

//createSharedFlow - uploads an apigee bundle as a shared flow
func (a *GatewayClient) createSharedFlow(data []byte, name string) error {
	var buffer bytes.Buffer
	writer := multipart.NewWriter(&buffer)

	// Create the flow zip part
	flow, _ := writer.CreateFormFile("file", name+".zip")
	io.Copy(flow, bytes.NewReader(data))
	writer.Close()

	// assemble the request with the writer content type and buffer data
	req, _ := http.NewRequest(http.MethodPost, fmt.Sprintf(orgURL+"/sharedflows?action=import&name=%s", a.cfg.Organization, name), &buffer)
	req.Header.Add("Content-Type", writer.FormDataContentType())
	req.Header.Add("Accept", "application/json")
	req.Header.Add("Authorization", "Bearer "+a.accessToken)
	client := &http.Client{}

	// submit the request
	_, err := client.Do(req)

	return err
}

//deploySharedFlow - deploye the shared flow and revision to the environmnet
func (a *GatewayClient) deploySharedFlow(env, name, revision string) error {
	queryParams := map[string]string{
		"override": "true",
	}

	// deploy the shared flow to the environment
	_, err := a.postRequestWithQuery(fmt.Sprintf(orgURL+"/environments/%v/sharedflows/%v/revisions/%v/deployments", a.cfg.Organization, env, name, revision), queryParams, []byte{})

	if err != nil {
		return err
	}

	return nil
}

//createSharedFlow - gets the list of shared flows
func (a *GatewayClient) publishSharedFlowToEnvironment(env, name string) error {
	// This is the structure that is expected for adding a shared flow as a flow hook
	type flowhook struct {
		ContinueOnError bool   `json:"continueOnError"`
		SharedFlow      string `json:"sharedFlow"`
		State           string `json:"state"`
	}

	// create the data for the put request
	hook := flowhook{
		ContinueOnError: true,
		SharedFlow:      name,
		State:           "deployed",
	}
	data, _ := json.Marshal(hook)

	// Add the flow to the post proxy flow hook
	_, err := a.putRequest(fmt.Sprintf(orgURL+"/environments/%v/flowhooks/PostProxyFlowHook", a.cfg.Organization, env), data)
	return err
}
