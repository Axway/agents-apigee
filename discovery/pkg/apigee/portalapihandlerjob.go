package apigee

import (
	"fmt"
	"strconv"

	"github.com/Axway/agent-sdk/pkg/agent"
	"github.com/Axway/agent-sdk/pkg/apic"
	"github.com/Axway/agent-sdk/pkg/cache"
	"github.com/Axway/agent-sdk/pkg/jobs"
	coreutil "github.com/Axway/agent-sdk/pkg/util"
	"github.com/Axway/agent-sdk/pkg/util/log"
	"github.com/Axway/agents-apigee/client/pkg/apigee"
	"github.com/Axway/agents-apigee/discovery/pkg/util"
)

const (
	gatewayType  = "Apigee"
	catalogIDKey = "PortalCatalogID"
)

//newPortalAPIHandler - job that waits for
type newPortalAPIHandler struct {
	jobs.Job
	apigeeClient   *apigee.ApigeeClient
	processAPIChan chan *apigee.APIDocData
	removedAPIChan chan string
	stopChan       chan interface{}
	isRunning      bool
	runningChan    chan bool
	productChan    chan productRequest
	shouldPushAPI  func(map[string]string) bool
}

func newPortalAPIHandlerJob(apigeeClient *apigee.ApigeeClient, channels *agentChannels, shouldPushAPI func(map[string]string) bool) *newPortalAPIHandler {
	job := &newPortalAPIHandler{
		apigeeClient:   apigeeClient,
		stopChan:       make(chan interface{}),
		isRunning:      false,
		processAPIChan: channels.processAPIChan,
		removedAPIChan: channels.removedAPIChan,
		runningChan:    make(chan bool),
		productChan:    channels.productChan,
		shouldPushAPI:  shouldPushAPI,
	}
	go job.statusUpdate()
	return job
}

func (j *newPortalAPIHandler) Ready() bool {
	return j.apigeeClient.IsReady()
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
		case newAPI, ok := <-j.processAPIChan:
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

func (j *newPortalAPIHandler) getPortalData(portalID string) (*apigee.PortalData, error) {
	portalDataInterface, err := cache.GetCache().Get(portalMapCacheKey)
	portalDataMap := make(map[string]apigee.PortalData, 0)
	if err != nil {
		return nil, fmt.Errorf("error getting portal data from cache")
	}
	ok := false
	if portalDataMap, ok = portalDataInterface.(map[string]apigee.PortalData); !ok {
		return nil, fmt.Errorf("portal data in cache could not be read")
	}

	if portal, ok := portalDataMap[portalID]; ok {
		return &portal, nil
	}
	return nil, fmt.Errorf("portal with ID %s not found", portalID)
}

func (j *newPortalAPIHandler) buildServiceBody(newAPI *apigee.APIDocData, productAttributes map[string]string) (*apic.ServiceBody, error) {
	// get the spec to build the service body
	spec := j.getAPISpec(newAPI.SpecContent)

	portal, err := j.getPortalData(newAPI.PortalID)
	if err != nil {
		return nil, err
	}

	// get the image, if set
	image := ""
	imageContentType := ""
	if newAPI.ImageURL != nil {
		image, imageContentType = j.apigeeClient.GetImageWithURL(*newAPI.ImageURL, portal.CurrentURL)
	}

	// create the service body to use for update or create
	apiID := fmt.Sprint(newAPI.ID)

	// create attributes to be added to revision and instance
	serviceAttributes := make(map[string]string)
	revisionAttributes := make(map[string]string)
	for k, v := range productAttributes {
		serviceAttributes[k] = v
		revisionAttributes[k] = v
	}
	revisionAttributes[catalogIDKey] = apiID
	revisionAttributes["PortalID"] = newAPI.PortalID

	state := apic.UnpublishedState
	if newAPI.Visibility {
		state = apic.PublishedState
	}

	sb, err := apic.NewServiceBodyBuilder().
		SetID(newAPI.ProductName).
		SetAPIName(newAPI.ProductName).
		SetStage(newAPI.GetPortalTitle()).
		SetStageDescriptor("Portal").
		SetDescription(newAPI.Description).
		SetAPISpec(spec).
		SetImage(image).
		SetImageContentType(imageContentType).
		SetState(state).
		SetStatus(state).
		SetTitle(newAPI.Title).
		SetSubscriptionName(defaultSubscriptionSchema).
		SetServiceAttribute(serviceAttributes).
		SetRevisionAttribute(revisionAttributes).
		SetInstanceAttribute(revisionAttributes).
		Build()
	return &sb, err
}

func (j *newPortalAPIHandler) handleAPI(newAPI *apigee.APIDocData) {
	log.Tracef("handling api %v from portal %v", newAPI.Title, newAPI.PortalID)

	// check if the APIs product can be discovered
	discover, attributes := j.checkProduct(newAPI.ProductName)
	if !discover {
		log.Infof("Skipping API %s in Portal %s as the attached Product %s did not match the discovery filter", newAPI.Title, newAPI.GetPortalTitle(), newAPI.ProductName)
		return
	}

	serviceBody, err := j.buildServiceBody(newAPI, attributes)
	if err != nil {
		log.Error(err)
		return
	}
	serviceBodyHash, _ := coreutil.ComputeHash(*serviceBody)

	// add auth info to cached api for subscriptions
	newAPI.SetAPIKeyInfo(serviceBody.GetAPIKeyInfo())
	newAPI.SetSecurityPolicies(serviceBody.GetAuthPolicies())

	hashString := util.ConvertUnitToString(serviceBodyHash)

	// Check DiscoveryCache for API
	if !agent.IsAPIPublishedByID(newAPI.ProductName) {
		// call new API
		j.publishAPI(newAPI, *serviceBody, hashString)n
	} else if value := agent.GetAttributeOnPublishedAPIByID(newAPI.ProductName, fmt.Sprintf("%s-hash", newAPI.PortalID)); value != hashString {
		// handle update
		log.Tracef("%s has been updated, push new revision", newAPI.ProductName)
		serviceBody.APIUpdateSeverity = "Major"
		j.publishAPI(newAPI, *serviceBody, hashString)
	} else {
		// no changes made do not update the cache
		return
	}

	// update the cache
	cacheKey := fmt.Sprintf("%s-%s", newAPI.PortalID, newAPI.ProductName)
	cache.GetCache().SetWithSecondaryKey(cacheKey, fmt.Sprintf("%s-%s", newAPI.GetPortalTitle(), newAPI.ProductName), *newAPI)
	cache.GetCache().SetSecondaryKey(cacheKey, strconv.Itoa(newAPI.ID))
}

func (j *newPortalAPIHandler) getAPISpec(contentID string) []byte {
	specData := []byte{}
	if contentID != "" {
		// Get API Spec
		specData = j.apigeeClient.GetSpecContent(contentID)
	}
	// handle products without a spec
	return specData
}

func (j *newPortalAPIHandler) publishAPI(newAPI *apigee.APIDocData, serviceBody apic.ServiceBody, hashString string) {

	// Add a few more attributes to the service body
	serviceBody.ServiceAttributes[fmt.Sprintf("%s-hash", newAPI.PortalID)] = hashString
	serviceBody.ServiceAttributes["GatewayType"] = gatewayType

	err := agent.PublishAPI(serviceBody)
	if err == nil {
		log.Infof("Published API %s to AMPLIFY Central", serviceBody.NameToPush)
	}
}

func (j *newPortalAPIHandler) handleRemovedAPI(removedAPIID string) {
	log.Tracef("handling removed api %s", removedAPIID)

	// find api by id in cache
	_, err := cache.GetCache().GetBySecondaryKey(removedAPIID)
	if err != nil {
		log.Errorf("could not find the removed api, %s, in the cache", removedAPIID)
		return
	}

	// remove from the cache
	cache.GetCache().DeleteBySecondaryKey(removedAPIID)
}

func (j *newPortalAPIHandler) checkProduct(productName string) (bool, map[string]string) {
	// get the product attributes
	attributesChan := make(chan map[string]string)
	j.productChan <- productRequest{
		name:     productName,
		response: attributesChan,
	}
	attributes := <-attributesChan

	if !j.shouldPushAPI(attributes) {
		// product should not be discovered
		return false, map[string]string{}
	}

	return true, attributes
}
