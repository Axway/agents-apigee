package apigee

import (
	"fmt"

	"github.com/Axway/agent-sdk/pkg/jobs"
	"github.com/Axway/agent-sdk/pkg/util/log"
)

//newPortalHandler - job that waits for
type newPortalHandler struct {
	jobs.Job
	apigeeClient *GatewayClient
	portalChan   chan string
	apiChan      chan *apiDocData
	stopChan     chan interface{}
	isRunning    bool
	statusChan   chan bool
}

func newPortalHandlerJob(apigeeClient *GatewayClient, portalChan chan string, apiChan chan *apiDocData) *newPortalHandler {
	job := &newPortalHandler{
		apigeeClient: apigeeClient,
		portalChan:   portalChan,
		stopChan:     make(chan interface{}),
		isRunning:    false,
		apiChan:      apiChan,
		statusChan:   make(chan bool),
	}
	go job.statusUpdate()
	return job
}

func (j *newPortalHandler) Ready() bool {
	if j.apigeeClient.accessToken == "" {
		return false
	}
	return true
}

func (j *newPortalHandler) Status() error {
	if !j.isRunning {
		return fmt.Errorf("new portal handler not running")
	}
	return nil
}

func (j *newPortalHandler) statusUpdate() {
	j.isRunning = <-j.statusChan
	j.statusUpdate()
}

func (j *newPortalHandler) started() {
	j.statusChan <- true
}

func (j *newPortalHandler) stopped() {
	j.statusChan <- false
}

func (j *newPortalHandler) Execute() error {
	j.started()
	defer j.stopped()
	for {
		select {
		case newPortal, ok := <-j.portalChan:
			if !ok {
				err := fmt.Errorf("Portal channel was closed")
				return err
			}
			j.handlePortal(newPortal)
		case <-j.stopChan:
			log.Info("Stopping the portal handler")
			return nil
		}
	}
}

func (j *newPortalHandler) handlePortal(newPortal string) {
	log.Tracef("Handling new portal %s", newPortal)

	// register a new job to poll for apis in this portal
	portalAPIsJob := newPollPortalAPIsJob(j.apigeeClient, newPortal, j.apiChan)
	jobs.RegisterIntervalJobWithName(portalAPIsJob, j.apigeeClient.pollInterval, fmt.Sprintf("%s Portal Poller", newPortal))
}
