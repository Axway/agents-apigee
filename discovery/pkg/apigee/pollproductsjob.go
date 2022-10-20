package apigee

import (
	"context"
	"fmt"
	"sync"

	"github.com/Axway/agent-sdk/pkg/agent"
	"github.com/Axway/agent-sdk/pkg/apic"
	"github.com/Axway/agent-sdk/pkg/jobs"
	coreutil "github.com/Axway/agent-sdk/pkg/util"
	"github.com/Axway/agent-sdk/pkg/util/log"

	"github.com/Axway/agents-apigee/client/pkg/apigee"
	"github.com/Axway/agents-apigee/client/pkg/apigee/models"
	"github.com/Axway/agents-apigee/discovery/pkg/util"
)

const (
	productNameField    ctxKeys = "product"
	productDetailsField ctxKeys = "productDetails"
)

type productClient interface {
	GetProducts() (apigee.Products, error)
	GetProduct(productName string) (*models.ApiProduct, error)
	GetSpecFile(specPath string) ([]byte, error)
	IsReady() bool
}

type productCache interface {
	GetSpecWithName(name string) (*specCacheItem, error)
	AddPublishedServiceToCache(cacheKey string, serviceBody *apic.ServiceBody)
}

// job that will poll for any new portals on APIGEE Edge
type pollProductsJob struct {
	jobs.Job
	client      productClient
	cache       productCache
	firstRun    bool
	specsReady  jobFirstRunDone
	pubLock     sync.Mutex
	publishFunc agent.PublishAPIFunc
	logger      log.FieldLogger
	workers     int
	running     bool
	runningLock sync.Mutex
}

func newPollProductsJob(client productClient, cache productCache, specsReady jobFirstRunDone, workers int) *pollProductsJob {
	job := &pollProductsJob{
		client:      client,
		cache:       cache,
		firstRun:    true,
		specsReady:  specsReady,
		logger:      log.NewFieldLogger().WithComponent("pollProducts").WithPackage("apigee"),
		publishFunc: agent.PublishAPI,
		workers:     workers,
		runningLock: sync.Mutex{},
	}
	return job
}

func (j *pollProductsJob) Ready() bool {
	j.logger.Trace("checking if the apigee client is ready for calls")
	if !j.client.IsReady() {
		return false
	}

	j.logger.Trace("checking if specs have been cached")
	return j.specsReady()
}

func (j *pollProductsJob) Status() error {
	return nil
}

func (j *pollProductsJob) updateRunning(running bool) {
	j.runningLock.Lock()
	defer j.runningLock.Unlock()
	j.running = running
}

func (j *pollProductsJob) isRunning() bool {
	j.runningLock.Lock()
	defer j.runningLock.Unlock()
	return j.running
}

func (j *pollProductsJob) Execute() error {
	j.logger.Trace("executing")

	if j.isRunning() {
		j.logger.Warn("previous spec poll job run has not completed, will run again on next interval")
		return nil
	}
	j.updateRunning(true)
	defer j.updateRunning(false)

	products, err := j.client.GetProducts()
	if err != nil {
		j.logger.WithError(err).Error("getting products")
		return err
	}

	limiter := make(chan string, j.workers)

	wg := sync.WaitGroup{}
	wg.Add(len(products))
	for _, p := range products {
		go func() {
			defer wg.Done()
			name := <-limiter
			j.handleProduct(name)
		}()
		limiter <- p
	}

	wg.Wait()
	close(limiter)

	return nil
}

func (j *pollProductsJob) FirstRunComplete() bool {
	return !j.firstRun
}

func (j *pollProductsJob) handleProduct(productName string) {
	logger := j.logger.WithField(productNameField.String(), productName)
	logger.Trace("handling product")

	// get product full details
	ctx := addLoggerToContext(context.Background(), logger)
	ctx = context.WithValue(ctx, productNameField, productName)

	// try to get spec by using the name of the product
	specDetails, err := j.cache.GetSpecWithName(productName)
	if err != nil {
		logger.WithError(err).Trace("could not find spec for product by name")
		return
	}
	ctx = context.WithValue(ctx, specPathField, specDetails.ContentPath)

	// get the full product details
	productDetails, err := j.client.GetProduct(productName)
	if err != nil {
		logger.WithError(err).Trace("could not retrieve product details")
		return
	}
	ctx = context.WithValue(ctx, productDetailsField, productDetails)

	// create service
	serviceBody, err := j.buildServiceBody(ctx)
	if err != nil {
		logger.WithError(err).Error("building service body")
		return
	}
	serviceBodyHash, _ := coreutil.ComputeHash(*serviceBody)
	hashString := util.ConvertUnitToString(serviceBodyHash)
	cacheKey := createProductCacheKey(productName)

	// Check DiscoveryCache for API
	j.pubLock.Lock() // only publish one at a time
	defer j.pubLock.Unlock()
	value := agent.GetAttributeOnPublishedAPIByID(productName, "hash")

	err = nil
	if !agent.IsAPIPublishedByID(productName) {
		// call new API
		err = j.publishAPI(*serviceBody, hashString, cacheKey)
	} else if value != hashString {
		// handle update
		log.Tracef("%s has been updated, push new revision", productName)
		serviceBody.APIUpdateSeverity = "Major"
		serviceBody.SpecDefinition = []byte{}
		log.Tracef("%+v", serviceBody)
		err = j.publishAPI(*serviceBody, hashString, cacheKey)
	}

	if err == nil {
		j.cache.AddPublishedServiceToCache(cacheKey, serviceBody)
	}
}

func (j *pollProductsJob) buildServiceBody(ctx context.Context) (*apic.ServiceBody, error) {
	logger := getLoggerFromContext(ctx)
	product := ctx.Value(productDetailsField).(*models.ApiProduct)
	specPath := getStringFromContext(ctx, specPathField)

	// get the spec to build the service body
	spec, err := j.client.GetSpecFile(specPath)
	if err != nil {
		logger.WithError(err).Error("could not download spec")
		return nil, err
	}

	if len(spec) == 0 {
		return nil, fmt.Errorf("spec had no content")
	}

	logger.Debug("creating service body")

	sb, err := apic.NewServiceBodyBuilder().
		SetID(product.Name).
		SetAPIName(product.Name).
		SetDescription(product.Description).
		SetAPISpec(spec).
		SetTitle(product.DisplayName).
		Build()
	return &sb, err
}

func (j *pollProductsJob) publishAPI(serviceBody apic.ServiceBody, hashString, cacheKey string) error {
	// Add a few more attributes to the service body
	serviceBody.ServiceAttributes["GatewayType"] = gatewayType
	serviceBody.ServiceAgentDetails["hash"] = hashString
	serviceBody.InstanceAgentDetails[cacheKeyAttribute] = cacheKey

	err := j.publishFunc(serviceBody)
	if err == nil {
		log.Infof("Published API %s to AMPLIFY Central", serviceBody.NameToPush)
		return err
	}
	return nil
}
