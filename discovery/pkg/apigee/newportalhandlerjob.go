package apigee

import (
	"fmt"

	"github.com/Axway/agent-sdk/pkg/cache"
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
	runningChan  chan bool
}

func newPortalHandlerJob(apigeeClient *GatewayClient, portalChan chan string, apiChan chan *apiDocData) *newPortalHandler {
	job := &newPortalHandler{
		apigeeClient: apigeeClient,
		portalChan:   portalChan,
		stopChan:     make(chan interface{}),
		isRunning:    false,
		apiChan:      apiChan,
		runningChan:  make(chan bool),
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
	for {
		select {
		case update := <-j.runningChan:
			j.isRunning = update
		}
	}
}

func (j *newPortalHandler) started() {
	j.runningChan <- true
}

func (j *newPortalHandler) stopped() {
	j.runningChan <- false
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

	portalName, err := j.getPortalNameByID(newPortal)
	if err != nil {
		log.Errorf("could not start watching for apis on portal %s", newPortal)
	}

	// register a new job to poll for apis in this portal
	portalAPIsJob := newPollPortalAPIsJob(j.apigeeClient, newPortal, portalName, j.apiChan)
	jobs.RegisterIntervalJobWithName(portalAPIsJob, j.apigeeClient.pollInterval, fmt.Sprintf("%s Portal Poller", portalName))
}

func (j *newPortalHandler) getPortalNameByID(newPortal string) (string, error) {
	portalMapInterface, err := cache.GetCache().Get(portalMapCacheKey)
	if err != nil {
		log.Error("error hit getting the portal map from the cache")
		return "", err
	}
	portalMap := portalMapInterface.(map[string]string)
	if portalName, ok := portalMap[newPortal]; ok {
		return portalName, nil
	}
	return "", fmt.Errorf("portal id %s not in cache", newPortal)
}
