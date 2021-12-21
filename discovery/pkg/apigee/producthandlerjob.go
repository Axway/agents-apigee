package apigee

import (
	"fmt"

	"github.com/Axway/agent-sdk/pkg/jobs"
	"github.com/Axway/agent-sdk/pkg/util/log"
	"github.com/Axway/agents-apigee/client/pkg/apigee"
)

//productHandler - job that waits for
type productHandler struct {
	jobs.Job
	apigeeClient *apigee.ApigeeClient
	productChan  chan productRequest
	stopChan     chan interface{}
	isRunning    bool
	runningChan  chan bool
}

func newProductHandlerJob(apigeeClient *apigee.ApigeeClient, channels *agentChannels) *productHandler {
	job := &productHandler{
		apigeeClient: apigeeClient,
		productChan:  channels.productChan,
		stopChan:     make(chan interface{}),
		isRunning:    false,
		runningChan:  make(chan bool),
	}
	go job.statusUpdate()
	return job
}

func (j *productHandler) Ready() bool {
	return j.apigeeClient.IsReady()
}

func (j *productHandler) Status() error {
	if !j.isRunning {
		return fmt.Errorf("portal handler not running")
	}
	return nil
}

func (j *productHandler) statusUpdate() {
	for {
		select {
		case update := <-j.runningChan:
			j.isRunning = update
		}
	}
}

func (j *productHandler) started() {
	j.runningChan <- true
}

func (j *productHandler) stopped() {
	j.runningChan <- false
}

func (j *productHandler) Execute() error {
	j.started()
	defer j.stopped()
	for {
		select {
		case req, ok := <-j.productChan:
			if !ok {
				err := fmt.Errorf("product channel was closed")
				return err
			}
			j.handleProductRequest(req)
		case <-j.stopChan:
			log.Info("Stopping the product handler")
			return nil
		}
	}
}

func (j *productHandler) handleProductRequest(req productRequest) {
	prod, err := j.apigeeClient.GetProduct(req.name)
	if err != nil {
		// product not found, return empty map
		req.response <- map[string]string{}
	}

	// get the product attributes in a map
	attributes := make(map[string]string)
	for _, att := range prod.Attributes {
		attributes[att.Name] = att.Value
	}

	// send the map back
	req.response <- attributes
}
