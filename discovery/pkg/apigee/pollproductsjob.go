package apigee

import (
	"context"
	"fmt"
	"path"
	"strings"
	"sync"
	"time"

	"github.com/Axway/agent-sdk/pkg/agent"
	"github.com/Axway/agent-sdk/pkg/apic"
	v1 "github.com/Axway/agent-sdk/pkg/apic/apiserver/models/api/v1"
	"github.com/Axway/agent-sdk/pkg/jobs"
	coreutil "github.com/Axway/agent-sdk/pkg/util"
	"github.com/Axway/agent-sdk/pkg/util/log"

	"github.com/Axway/agents-apigee/client/pkg/apigee"
	"github.com/Axway/agents-apigee/client/pkg/apigee/models"
	"github.com/Axway/agents-apigee/client/pkg/config"
	"github.com/Axway/agents-apigee/discovery/pkg/util"
)

const specLocalTag = "spec_local"

type productClient interface {
	GetConfig() *config.ApigeeConfig
	GetProducts() (apigee.Products, error)
	GetProduct(productName string) (*models.ApiProduct, error)
	GetSpecFile(specPath string) ([]byte, error)
	IsReady() bool
}

type productCacheItem struct {
	Name     string
	SpecHash string
	ModDate  time.Time
}

type productCache interface {
	GetSpecWithName(name string) (*specCacheItem, error)
	AddProductToCache(name string, modDate time.Time, specHash string)
	HasProductChanged(name string, modDate time.Time, specHash string) bool
	GetProductWithName(name string) (*productCacheItem, error)
}

type isPublishedFunc func(string) bool
type getAttributeFunc func(string, string) string

// job that will poll for any new portals on APIGEE Edge
type pollProductsJob struct {
	jobs.Job
	client           productClient
	cache            productCache
	firstRun         bool
	specsReady       jobFirstRunDone
	pubLock          sync.Mutex
	isPublishedFunc  isPublishedFunc
	getAttributeFunc getAttributeFunc
	publishFunc      agent.PublishAPIFunc
	logger           log.FieldLogger
	workers          int
	running          bool
	runningLock      sync.Mutex
	shouldPushAPI    func(map[string]string) bool
}

func newPollProductsJob(client productClient, cache productCache, specsReady jobFirstRunDone, workers int, shouldPushAPI func(map[string]string) bool) *pollProductsJob {
	job := &pollProductsJob{
		client:           client,
		cache:            cache,
		firstRun:         true,
		specsReady:       specsReady,
		logger:           log.NewFieldLogger().WithComponent("pollProducts").WithPackage("apigee"),
		isPublishedFunc:  agent.IsAPIPublishedByID,
		getAttributeFunc: agent.GetAttributeOnPublishedAPIByID,
		publishFunc:      agent.PublishAPI,
		workers:          workers,
		runningLock:      sync.Mutex{},
		shouldPushAPI:    shouldPushAPI,
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

	j.firstRun = false
	return nil
}

func (j *pollProductsJob) FirstRunComplete() bool {
	return !j.firstRun
}

func (j *pollProductsJob) handleProduct(productName string) {
	logger := j.logger.WithField("productName", productName)
	logger.Trace("handling product")

	// get product full details
	ctx := addLoggerToContext(context.Background(), logger)

	// get the full product details
	productDetails, err := j.client.GetProduct(productName)
	if err != nil {
		logger.WithError(err).Trace("could not retrieve product details")
		return
	}
	logger = logger.WithField("productDisplay", productDetails.DisplayName)

	if !j.shouldPublishProduct(logger, productDetails) {
		logger.Trace("product has been filtered out")
		return
	}

	// try to get spec by using the name of the product
	ctx, err = j.getSpecDetails(ctx, productDetails)
	if err != nil {
		logger.Trace("could not find spec for product by name")
		return
	}

	// create service
	serviceBody, specHash, err := j.buildServiceBody(ctx, productDetails)
	if err != nil {
		logger.WithError(err).Error("building service body")
		return
	}

	serviceBodyHash, _ := coreutil.ComputeHash(*serviceBody)
	hashString := util.ConvertUnitToString(serviceBodyHash)
	specHashString := util.ConvertUnitToString(specHash)
	cacheKey := createProductCacheKey(productName)

	// Check DiscoveryCache for API
	j.pubLock.Lock() // only publish one at a time
	defer j.pubLock.Unlock()
	value := agent.GetAttributeOnPublishedAPIByID(productName, "hash")

	err = nil
	if !j.isPublishedFunc(productName) {
		// call new API
		err = j.publishAPI(*serviceBody, hashString, cacheKey)
	} else if value != hashString {
		// handle update
		log.Tracef("%s has been updated, push new revision", productName)
		serviceBody.APIUpdateSeverity = "Major"
		log.Tracef("%+v", serviceBody)
		err = j.publishAPI(*serviceBody, hashString, cacheKey)
	}

	if err == nil {
		j.cache.AddProductToCache(productName, time.UnixMilli(int64(productDetails.LastModifiedAt)), specHashString)
	}
}

func (j *pollProductsJob) shouldPublishProduct(logger log.FieldLogger, product *models.ApiProduct) bool {
	// get the product attributes in a map
	attributes := make(map[string]string)
	for _, att := range product.Attributes {
		// ignore access attribute
		if strings.ToLower(att.Name) == "access" {
			continue
		}
		attributes[att.Name] = att.Value
	}
	logger = logger.WithField("attributes", attributes)

	if val, ok := attributes[agentProductTagName]; ok && val == agentProductTagValue {
		logger.Trace("product was created by agent, skipping")
		return false
	}

	logger.WithField("attributes", attributes).Trace("checking against discovery filter")
	return j.shouldPushAPI(attributes)
}

func (j *pollProductsJob) getSpecDetails(ctx context.Context, product *models.ApiProduct) (context.Context, error) {
	for _, att := range product.Attributes {
		// find the spec_local tag
		if strings.ToLower(att.Name) == specLocalTag {
			ctx = context.WithValue(ctx, specPathField, strings.Join([]string{specLocalTag, att.Value}, "_"))
			return ctx, nil
		}
	}

	specDetails, err := j.cache.GetSpecWithName(product.Name)
	if err != nil {
		// try to find the spec details with the display name before giving up
		specDetails, err = j.cache.GetSpecWithName(product.DisplayName)
		if err != nil {
			return ctx, err
		}
	}
	ctx = context.WithValue(ctx, specPathField, specDetails.ContentPath)
	return ctx, nil
}

func (j *pollProductsJob) buildServiceBody(ctx context.Context, product *models.ApiProduct) (*apic.ServiceBody, uint64, error) {
	logger := getLoggerFromContext(ctx)
	specPath := getStringFromContext(ctx, specPathField)

	var spec []byte
	var err error
	if strings.HasPrefix(specPath, specLocalTag) {
		logger = logger.WithField("specLocalDir", "true")
		fileName := strings.TrimPrefix(specPath, specLocalTag+"_")
		filePath := path.Join(j.client.GetConfig().Specs.LocalPath, fileName)
		spec, err = loadSpecFile(logger, filePath, nil)
	} else {
		logger = logger.WithField("specLocalDir", "false")
		// get the spec to build the service body
		spec, err = j.client.GetSpecFile(specPath)
	}

	if err != nil {
		logger.WithError(err).Error("could not download spec")
		return nil, 0, err
	}

	if len(spec) == 0 {
		return nil, 0, fmt.Errorf("spec had no content")
	}

	specHash, _ := coreutil.ComputeHash(spec)

	// create the agent details with the modification dates
	serviceDetails := map[string]interface{}{
		"productModDate":  time.UnixMilli(int64(product.LastModifiedAt)).Format(v1.APIServerTimeFormat),
		"specContentHash": specHash,
	}

	// create attributes to be added to service
	serviceAttributes := make(map[string]string)
	for _, att := range product.Attributes {
		name := strings.ToLower(att.Name)
		name = strings.ReplaceAll(name, " ", "_")
		serviceAttributes[name] = att.Value
	}

	logger.Debug("creating service body")
	sb, err := apic.NewServiceBodyBuilder().
		SetID(product.Name).
		SetAPIName(product.Name).
		SetDescription(product.Description).
		SetAPISpec(spec).
		SetTitle(product.DisplayName).
		SetServiceAttribute(serviceAttributes).
		SetServiceAgentDetails(serviceDetails).
		Build()
	return &sb, specHash, err
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
