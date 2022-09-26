package apigee

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"

	"github.com/Axway/agents-apigee/client/pkg/apigee/models"
)

// GetSharedFlow - gets the list of shared flows
func (a *ApigeeClient) GetSharedFlow(name string) (*models.SharedFlowRevisionDeploymentDetails, error) {
	// Get the shared flows list
	response, err := a.newRequest(http.MethodGet, fmt.Sprintf("%s/sharedflows/%v", a.orgURL, name),
		WithDefaultHeaders(),
	).Execute()
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

// CreateSharedFlow - uploads an apigee bundle as a shared flow
func (a *ApigeeClient) CreateSharedFlow(data []byte, name string) error {
	var buffer bytes.Buffer
	writer := multipart.NewWriter(&buffer)

	// Create the flow zip part
	flow, _ := writer.CreateFormFile("file", name+".zip")
	io.Copy(flow, bytes.NewReader(data))
	writer.Close()

	// assemble the request with the writer content type and buffer data
	req, _ := http.NewRequest(http.MethodPost, fmt.Sprintf("%s/sharedflows?action=import&name=%s", a.orgURL, name), &buffer)
	req.Header.Add("Content-Type", writer.FormDataContentType())
	req.Header.Add("Accept", "application/json")
	req.Header.Add("Authorization", "Bearer "+a.accessToken)
	client := &http.Client{}

	// submit the request
	_, err := client.Do(req)

	return err
}

// DeploySharedFlow - deploy the shared flow and revision to the environment
func (a *ApigeeClient) DeploySharedFlow(env, name, revision string) error {

	// deploy the shared flow to the environment
	_, err := a.newRequest(http.MethodPost, fmt.Sprintf("%s/environments/%v/sharedflows/%v/revisions/%v/deployments", a.orgURL, env, name, revision),
		WithDefaultHeaders(),
		WithBody([]byte{}),
		WithQueryParam("override", "true"),
	).Execute()

	if err != nil {
		return err
	}

	return nil
}

// PublishSharedFlowToEnvironment - publish the shared flow
func (a *ApigeeClient) PublishSharedFlowToEnvironment(env, name string) error {
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
	_, err := a.newRequest(http.MethodPut, fmt.Sprintf("%s/environments/%v/flowhooks/PostProxyFlowHook", a.orgURL, env),
		WithDefaultHeaders(),
		WithBody(data),
	).Execute()
	return err
}
