package apigee

import (
	"fmt"

	"github.com/Axway/agent-sdk/pkg/agent"
	"github.com/Axway/agent-sdk/pkg/apic"
	"github.com/Axway/agent-sdk/pkg/cache"
	"github.com/Axway/agent-sdk/pkg/jobs"
	coreutil "github.com/Axway/agent-sdk/pkg/util"
	"github.com/Axway/agent-sdk/pkg/util/log"
	"github.com/Axway/agents-apigee/discovery/pkg/util"
)

const (
	hashKey = "APIGEEHash"
)

//newPortalAPIHandler - job that waits for
type newPortalAPIHandler struct {
	jobs.Job
	apigeeClient   *GatewayClient
	newAPIChan     chan *apiDocData
	removedAPIChan chan string
	stopChan       chan interface{}
	isRunning      bool
	runningChan    chan bool
}

func newPortalAPIHandlerJob(apigeeClient *GatewayClient, newAPIChan chan *apiDocData, removedAPIChan chan string) *newPortalAPIHandler {
	job := &newPortalAPIHandler{
		apigeeClient:   apigeeClient,
		stopChan:       make(chan interface{}),
		isRunning:      false,
		newAPIChan:     newAPIChan,
		removedAPIChan: removedAPIChan,
		runningChan:    make(chan bool),
	}
	go job.statusUpdate()
	return job
}

func (j *newPortalAPIHandler) Ready() bool {
	if j.apigeeClient.accessToken == "" {
		return false
	}
	return true
}

func (j *newPortalAPIHandler) Status() error {
	if !j.isRunning {
		return fmt.Errorf("new api handler not running")
	}
	return nil
}

func (j *newPortalAPIHandler) Execute() error {
	j.started()
	defer j.stopped()
	for {
		select {
		case newAPI, ok := <-j.newAPIChan:
			if !ok {
				err := fmt.Errorf("new api channel was closed")
				return err
			}
			j.handleAPI(newAPI)
		case removedAPI, ok := <-j.removedAPIChan:
			if !ok {
				err := fmt.Errorf("removed api channel was closed")
				return err
			}
			j.handleRemovedAPI(removedAPI)
		case <-j.stopChan:
			log.Info("Stopping the api handler")
			return nil
		}
	}
}

func (j *newPortalAPIHandler) statusUpdate() {
	for {
		select {
		case update := <-j.runningChan:
			j.isRunning = update
		}
	}
}

func (j *newPortalAPIHandler) started() {
	j.runningChan <- true
}

func (j *newPortalAPIHandler) stopped() {
	j.runningChan <- false
}

func (j *newPortalAPIHandler) handleAPI(newAPI *apiDocData) {
	log.Tracef("Handling new api %+v", newAPI)

	// get the spec to build the service body
	spec := j.getAPISpec(newAPI.SpecContent)

	// create the service body to use for update or create
	apiID := fmt.Sprint(newAPI.ID)
	serviceBody, _ := apic.NewServiceBodyBuilder().
		SetID(apiID).
		SetAPIName(fmt.Sprintf("%s-%s", newAPI.PortalID, newAPI.APIID)).
		SetDescription(newAPI.Description).
		SetAPISpec(spec).
		SetTitle(fmt.Sprintf("%s (%s)", newAPI.Title, newAPI.PortalTitle)).
		Build()

	serviceBodyHash, _ := coreutil.ComputeHash(serviceBody)

	// Check DiscoveryCache for API
	if !agent.IsAPIPublishedByID(apiID) {
		// call new API
		j.publishAPI(newAPI, serviceBody, serviceBodyHash)
		return
	}

	// Check to see if the API has changed
	if value := agent.GetAttributeOnPublishedAPIByID(apiID, hashKey); value != fmt.Sprint(serviceBodyHash) {
		// handle update
		log.Tracef("%s has been updated, push new revision", newAPI.ProductName)
		serviceBody.APIUpdateSeverity = "Major"
		j.publishAPI(newAPI, serviceBody, serviceBodyHash)
	}
}

func (j *newPortalAPIHandler) getAPISpec(contentID string) []byte {
	specData := []byte{}
	if contentID != "" {
		// Get API Spec
		specData = j.apigeeClient.getSpecContent(contentID)
	}
	// handle products without a spec
	return specData
}

func (j *newPortalAPIHandler) publishAPI(newAPI *apiDocData, serviceBody apic.ServiceBody, serviceBodyHash uint64) {

	// Add a few more attributes to the service body
	serviceBody.ServiceAttributes[hashKey] = util.ConvertUnitToString(serviceBodyHash)
	serviceBody.ServiceAttributes["GatewayType"] = gatewayType

	err := agent.PublishAPI(serviceBody)
	if err == nil {
		log.Infof("Published API %s to AMPLIFY Central", serviceBody.NameToPush)
		currentHash, _ := coreutil.ComputeHash(serviceBody)
		cache.GetCache().Set(fmt.Sprint(newAPI.ID), currentHash)
	}
}

func (j *newPortalAPIHandler) handleRemovedAPI(removedAPIID string) {
	log.Tracef("Handling removed api %s", removedAPIID)

	// TODO - handle removed API
}
