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
	apigeeClient *GatewayClient
	apiChan      chan *apiDocData
	stopChan     chan interface{}
	isRunning    bool
	runningChan  chan bool
}

func newPortalAPIHandlerJob(apigeeClient *GatewayClient, apiChan chan *apiDocData) *newPortalAPIHandler {
	job := &newPortalAPIHandler{
		apigeeClient: apigeeClient,
		stopChan:     make(chan interface{}),
		isRunning:    false,
		apiChan:      apiChan,
		runningChan:  make(chan bool),
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
		case newAPI, ok := <-j.apiChan:
			if !ok {
				err := fmt.Errorf("new api channel was closed")
				return err
			}
			j.handleAPI(newAPI)
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

	// Check DiscoveryCache for API
	if !agent.IsAPIPublishedByID(fmt.Sprint(newAPI.ID)) {
		// call new API
		j.handleNewAPI(newAPI)
		return
	}

	//
}

func (j *newPortalAPIHandler) getAPISpec(contentID string) []byte {
	specData := []byte{}
	if contentID != "" {
		// Get API Spec
		specData = j.apigeeClient.getSpecContent(contentID)
	}
	return specData
}

func (j *newPortalAPIHandler) handleNewAPI(newAPI *apiDocData) {
	spec := j.getAPISpec(newAPI.SpecContent)

	apiTitle := fmt.Sprintf("%s (%s)", newAPI.Title, newAPI.PortalTitle)
	serviceBody, _ := apic.NewServiceBodyBuilder().
		SetID(fmt.Sprint(newAPI.ID)).
		SetAPIName(fmt.Sprintf("%s-%s", newAPI.PortalID, newAPI.APIID)).
		SetDescription(newAPI.Description).
		SetAPISpec(spec).
		SetTitle(apiTitle).
		Build()

	serviceBodyHash, _ := coreutil.ComputeHash(serviceBody)

	log.Infof("Published API %s to AMPLIFY Central", apiTitle)
	serviceBody.ServiceAttributes[hashKey] = util.ConvertUnitToString(serviceBodyHash)
	serviceBody.ServiceAttributes["GatewayType"] = gatewayType
	agent.PublishAPI(serviceBody)
	currentHash, _ := coreutil.ComputeHash(serviceBody)
	cache.GetCache().Set(fmt.Sprint(newAPI.ID), currentHash)
}
