package apigee

import (
	"context"
	"fmt"
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
	"github.com/Axway/agents-apigee/discovery/pkg/util"
)

const (
	productNameField        ctxKeys = "product"
	productDisplayNameField ctxKeys = "productDisplay"
	productDetailsField     ctxKeys = "productDetails"
	productModDateField     ctxKeys = "productModDate"
	specDetailsField        ctxKeys = "specDetails"
	specModDateField        ctxKeys = "specModDate"
)

type productClient interface {
	GetProducts() (apigee.Products, error)
	GetProduct(productName string) (*models.ApiProduct, error)
	GetSpecFile(specPath string) ([]byte, error)
	IsReady() bool
}

type productCacheItem struct {
	Name        string
	ModDate     time.Time
	SpecModDate time.Time
}

type productCache interface {
	GetSpecWithName(name string) (*specCacheItem, error)
	AddProductToCache(name string, modDate time.Time, specModDate time.Time)
	HasProductChanged(name string, modDate time.Time, specModDate time.Time) bool
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
	logger := j.logger.WithField(productNameField.String(), productName)
	logger.Trace("handling product")

	// get product full details
	ctx := addLoggerToContext(context.Background(), logger)
	ctx = context.WithValue(ctx, productNameField, productName)

	// get the full product details
	productDetails, err := j.client.GetProduct(productName)
	if err != nil {
		logger.WithError(err).Trace("could not retrieve product details")
		return
	}
	ctx = context.WithValue(ctx, productDetailsField, productDetails)
	ctx = context.WithValue(ctx, productDisplayNameField, productDetails.DisplayName)
	logger = logger.WithField(productDisplayNameField.String(), productDetails.DisplayName)

	if !j.shouldPublishProduct(ctx) {
		logger.Trace("product has been filtered out")
		return
	}

	// try to get spec by using the name of the product
	specDetails, err := j.getSpecDetails(ctx)
	ctx = context.WithValue(ctx, specDetailsField, specDetails)
	if err != nil {
		logger.Trace("could not find spec for product by name")
		return
	}
	ctx = context.WithValue(ctx, specPathField, specDetails.ContentPath)

	ctx, changed := j.hasProductChanged(ctx)
	if !changed {
		logger.Trace("no changes detected for product")
		return
	}

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
		j.cache.AddProductToCache(productName, ctx.Value(productModDateField).(time.Time), specDetails.ModDate)
	}
}

func (j *pollProductsJob) hasProductChanged(ctx context.Context) (context.Context, bool) {
	logger := getLoggerFromContext(ctx)
	logger.Trace("checking for product changes")

	productDetails := ctx.Value(productDetailsField).(*models.ApiProduct)
	specDetails := ctx.Value(specDetailsField).(*specCacheItem)
	productName := getStringFromContext(ctx, productNameField)
	productModDate := time.UnixMilli(int64(productDetails.LastModifiedAt)).UTC()
	ctx = context.WithValue(ctx, productModDateField, productModDate)
	ctx = context.WithValue(ctx, specModDateField, specDetails.ModDate)

	if cachedProduct, _ := j.cache.GetProductWithName(productName); cachedProduct != nil {
		// check if any changes have been made to the product or spec
		if j.cache.HasProductChanged(productName, productModDate, specDetails.ModDate) {
			logger.Trace("no changes detected for product based on cache")
			return ctx, true
		}

		return ctx, false
	}

	// check if the product exists
	isPublished := j.isPublishedFunc(productName)
	if isPublished {
		logger.Debug("checking published products for changes to the modification date")
		// check if the mod dates have changed prior to continuing
		curProdModStr := j.getAttributeFunc(productName, productModDateField.String())
		curSpecModStr := j.getAttributeFunc(productName, specModDateField.String())

		// did not retrieve attributes, mark as changed
		if curProdModStr == "" && curSpecModStr == "" {
			return ctx, true
		}

		curProdMod, _ := time.Parse(v1.APIServerTimeFormat, curProdModStr)
		curSpecMod, _ := time.Parse(v1.APIServerTimeFormat, curSpecModStr)
		logger := logger.WithField("currentProductModDate", curProdMod).
			WithField("productModDate", productModDate).
			WithField("currentSpecModDate", curSpecMod).
			WithField("specModDate", specDetails.ModDate)

		if productModDate.After(curProdMod) || specDetails.ModDate.After(curSpecMod) {
			logger.Trace("published product change detected")
			return ctx, true
		}

		logger.Trace("published product no change detected")
		return ctx, false
	}

	logger.Trace("no published product with name")
	return ctx, true
}

func (j *pollProductsJob) shouldPublishProduct(ctx context.Context) bool {
	product := ctx.Value(productDetailsField).(*models.ApiProduct)
	// get the product attributes in a map
	attributes := make(map[string]string)
	for _, att := range product.Attributes {
		// ignore access attribute
		if strings.ToLower(att.Name) == "access" {
			continue
		}
		attributes[att.Name] = att.Value
	}
	logger := j.logger.WithField("attributes", attributes)

	if val, ok := attributes[agentProductTagName]; ok && val == agentProductTagValue {
		logger.Trace("product was created by agent, skipping")
		return false
	}

	logger.WithField("attributes", attributes).Trace("checking against discovery filter")
	return j.shouldPushAPI(attributes)
}

func (j *pollProductsJob) getSpecDetails(ctx context.Context) (*specCacheItem, error) {
	productName := getStringFromContext(ctx, productNameField)
	displayName := getStringFromContext(ctx, productDisplayNameField)

	specDetails, err := j.cache.GetSpecWithName(productName)
	if err != nil {
		// try to find the spec details with the display name before giving up
		specDetails, err = j.cache.GetSpecWithName(displayName)
	}
	return specDetails, err
}

func (j *pollProductsJob) buildServiceBody(ctx context.Context) (*apic.ServiceBody, error) {
	logger := getLoggerFromContext(ctx)
	product := ctx.Value(productDetailsField).(*models.ApiProduct)
	specPath := getStringFromContext(ctx, specPathField)
	productModDate := ctx.Value(productModDateField).(time.Time)
	specModDate := ctx.Value(specModDateField).(time.Time)

	// get the spec to build the service body
	spec, err := j.client.GetSpecFile(specPath)
	if err != nil {
		logger.WithError(err).Error("could not download spec")
		return nil, err
	}

	if len(spec) == 0 {
		return nil, fmt.Errorf("spec had no content")
	}

	// create the agent details with the modification dates
	serviceDetails := map[string]interface{}{
		productModDateField.String(): productModDate.Format(v1.APIServerTimeFormat),
		specModDateField.String():    specModDate.Format(v1.APIServerTimeFormat),
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
