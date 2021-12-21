package apigee

import (
	"fmt"
	"strings"
	"time"

	"github.com/Axway/agent-sdk/pkg/cache"
	"github.com/Axway/agent-sdk/pkg/jobs"
	"github.com/Axway/agent-sdk/pkg/util/log"
	"github.com/Axway/agents-apigee/client/pkg/apigee"
)

//productCacheItem - the item to be stored in the cache
type productCacheItem struct {
	name       string
	attribtues map[string]string
	timestamp  time.Time
}

//productHandler - job that waits for
type productHandler struct {
	jobs.Job
	apigeeClient    *apigee.ApigeeClient
	productChan     chan productRequest
	stopChan        chan interface{}
	isRunning       bool
	runningChan     chan bool
	refreshInterval time.Duration
}

func newProductHandlerJob(apigeeClient *apigee.ApigeeClient, channels *agentChannels, refreshInterval time.Duration) *productHandler {
	job := &productHandler{
		apigeeClient:    apigeeClient,
		productChan:     channels.productChan,
		stopChan:        make(chan interface{}),
		isRunning:       false,
		runningChan:     make(chan bool),
		refreshInterval: refreshInterval,
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
				err := fmt.Errorf("product request channel was closed")
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
	cacheKey := fmt.Sprintf("product-%s-attributes", req.name)

	// check if product attributes are in the cache
	if itemInterface, err := cache.GetCache().Get(cacheKey); err == nil {
		//item existed
		cacheItem := itemInterface.(productCacheItem)
		if time.Now().Sub(cacheItem.timestamp) < j.refreshInterval {
			// return the existing attributes
			req.response <- cacheItem.attribtues
			return
		}
	}

	// product is not in cache or its time to refresh
	prod, err := j.apigeeClient.GetProduct(req.name)
	if err != nil {
		// product not found, return empty map
		req.response <- map[string]string{}
	}

	// get the product attributes in a map
	attributes := make(map[string]string)
	for _, att := range prod.Attributes {
		// ignore access attribute
		if strings.ToLower(att.Name) == "access" {
			continue
		}
		attributes[att.Name] = att.Value
	}

	// update the cache
	item := productCacheItem{
		name:       req.name,
		attribtues: attributes,
		timestamp:  time.Now(),
	}
	cache.GetCache().Set(cacheKey, item)

	// send the map back
	req.response <- attributes
}
