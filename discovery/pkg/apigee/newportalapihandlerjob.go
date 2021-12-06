package apigee

import (
	"fmt"

	"github.com/Axway/agent-sdk/pkg/jobs"
	"github.com/Axway/agent-sdk/pkg/util/log"
)

//newPortalAPIHandler - job that waits for
type newPortalAPIHandler struct {
	jobs.Job
	apigeeClient *GatewayClient
	apiChan      chan *apiDocData
	stopChan     chan interface{}
	isRunning    bool
	statusChan   chan bool
}

func newPortalAPIHandlerJob(apigeeClient *GatewayClient, apiChan chan *apiDocData) *newPortalAPIHandler {
	job := &newPortalAPIHandler{
		apigeeClient: apigeeClient,
		stopChan:     make(chan interface{}),
		isRunning:    false,
		apiChan:      apiChan,
		statusChan:   make(chan bool),
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

func (j *newPortalAPIHandler) statusUpdate() {
	j.isRunning = <-j.statusChan
	j.statusUpdate()
}

func (j *newPortalAPIHandler) started() {
	j.statusChan <- true
}

func (j *newPortalAPIHandler) stopped() {
	j.statusChan <- false
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

func (j *newPortalAPIHandler) handleAPI(newAPI *apiDocData) {
	log.Tracef("Handling new api %+v", newAPI)

	// Get API Spec

	// Check DiscoveryCache for API

	//
}
