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
	"github.com/tidwall/gjson"
)

const (
	gatewayType  = "APIGEE"
	catalogIDKey = "PortalCatalogID"
)

//newPortalAPIHandler - job that waits for
type newPortalAPIHandler struct {
	jobs.Job
	apigeeClient   *GatewayClient
	processAPIChan chan *apiDocData
	removedAPIChan chan string
	stopChan       chan interface{}
	isRunning      bool
	runningChan    chan bool
}

func newPortalAPIHandlerJob(apigeeClient *GatewayClient, processAPIChan chan *apiDocData, removedAPIChan chan string) *newPortalAPIHandler {
	job := &newPortalAPIHandler{
		apigeeClient:   apigeeClient,
		stopChan:       make(chan interface{}),
		isRunning:      false,
		processAPIChan: processAPIChan,
		removedAPIChan: removedAPIChan,
		runningChan:    make(chan bool),
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

func (j *newPortalAPIHandler) getPortalData(portalID string) (*portalData, error) {
	portalDataInterface, err := cache.GetCache().Get(portalMapCacheKey)
	portalDataMap := make(map[string]portalData, 0)
	if err != nil {
		return nil, fmt.Errorf("error getting portal data from cache")
	}
	ok := false
	if portalDataMap, ok = portalDataInterface.(map[string]portalData); !ok {
		return nil, fmt.Errorf("portal data in cache could not be read")
	}

	if portal, ok := portalDataMap[portalID]; ok {
		return &portal, nil
	}
	return nil, fmt.Errorf("portal with ID %s not found", portalID)
}

func (j *newPortalAPIHandler) buildServiceBody(newAPI *apiDocData) (*apic.ServiceBody, error) {
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
		image, imageContentType = j.apigeeClient.getImageWithURL(*newAPI.ImageURL, portal.CurrentURL)
	}

	// create the service body to use for update or create
	apiID := fmt.Sprint(newAPI.ID)

	// create attributes to be added to revision and instance
	attributes := make(map[string]string)
	attributes[catalogIDKey] = apiID
	attributes["PortalID"] = newAPI.PortalID

	sb, err := apic.NewServiceBodyBuilder().
		SetID(newAPI.ProductName).
		SetAPIName(newAPI.ProductName).
		SetStage(newAPI.PortalTitle).
		SetStageDescriptor("Portal").
		SetDescription(newAPI.Description).
		SetAPISpec(spec).
		SetImage(image).
		SetImageContentType(imageContentType).
		SetAuthPolicy(j.determineAuthPolicyFromSpec(spec)).
		SetTitle(newAPI.Title).
		SetSubscriptionName(defaultSubscriptionSchema).
		SetRevisionAttribute(attributes).
		SetInstanceAttribute(attributes).
		Build()
	return &sb, err
}

func (j *newPortalAPIHandler) handleAPI(newAPI *apiDocData) {
	log.Tracef("handling api %v from portal %v", newAPI.Title, newAPI.PortalID)

	serviceBody, err := j.buildServiceBody(newAPI)
	if err != nil {
		log.Error(err)
		return
	}

	serviceBodyHash, _ := coreutil.ComputeHash(*serviceBody)

	// Check DiscoveryCache for API
	if !agent.IsAPIPublishedByID(newAPI.ProductName) {
		// call new API
		j.publishAPI(newAPI, serviceBody, serviceBodyHash)
		return
	}

	// Check to see if the API has changed
	if value := agent.GetAttributeOnPublishedAPIByID(newAPI.ProductName, fmt.Sprintf("%s-hash", newAPI.PortalID)); value != fmt.Sprint(serviceBodyHash) {
		// handle update
		log.Tracef("%s has been updated, push new revision", newAPI.ProductName)
		serviceBody.APIUpdateSeverity = "Major"
		j.publishAPI(newAPI, serviceBody, serviceBodyHash)
	}
	// update the cache
	cache.GetCache().Set(fmt.Sprintf("%s-%s", newAPI.PortalID, newAPI.ProductName), *newAPI)
}

func (j *newPortalAPIHandler) getAPISpec(contentID string) []byte {
	specData := []byte{}
	if contentID != "" {
		// Get API Spec
		specData = j.apigeeClient.getSpecContent(contentID)
	}
	// handle products without a spec
	return specData
}

func (j *newPortalAPIHandler) publishAPI(newAPI *apiDocData, serviceBody *apic.ServiceBody, serviceBodyHash uint64) {

	// Add a few more attributes to the service body
	serviceBody.ServiceAttributes[fmt.Sprintf("%s-hash", newAPI.PortalID)] = util.ConvertUnitToString(serviceBodyHash)
	serviceBody.ServiceAttributes["GatewayType"] = gatewayType

	err := agent.PublishAPI(*serviceBody)
	if err == nil {
		log.Infof("Published API %s to AMPLIFY Central", serviceBody.NameToPush)
	}
}

func (j *newPortalAPIHandler) handleRemovedAPI(removedAPIID string) {
	log.Tracef("Handling removed api %s", removedAPIID)

	// TODO - handle removed API
}

func (j *newPortalAPIHandler) determineAuthPolicyFromSpec(swagger []byte) string {
	// Check for a security definition in the PAS spec
	var authPolicy = apic.Passthrough
	const (
		apiKey = "apiKey"
		oauth  = "oauth2"
	)

	// OAS2
	securityDefs := gjson.GetBytes(swagger, "securityDefinitions.*.type")
	for _, def := range securityDefs.Array() {
		if def.String() == apiKey {
			authPolicy = apic.Apikey
			return authPolicy
		}
		if def.String() == oauth {
			authPolicy = apic.Oauth
		}
	}

	// OAS3
	securityDefs = gjson.GetBytes(swagger, "components.securitySchemes.*.type")
	for _, def := range securityDefs.Array() {
		if def.String() == apiKey {
			authPolicy = apic.Apikey
			return authPolicy
		}
		if def.String() == oauth {
			authPolicy = apic.Oauth
		}
	}

	return authPolicy
}
