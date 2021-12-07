package apigee

import (
	"time"

	coreapi "github.com/Axway/agent-sdk/pkg/api"
	"github.com/Axway/agent-sdk/pkg/jobs"

	"github.com/Axway/agents-apigee/discovery/pkg/config"
)

const (
	openapi     = "openapi"
	gatewayType = "APIGEE"
)

// GatewayClient - Represents the Gateway client
type GatewayClient struct {
	cfg          *config.ApigeeConfig
	apiClient    coreapi.Client
	accessToken  string
	pollInterval time.Duration
	envToURLs    map[string][]string
	stopChan     chan struct{}
}

// NewClient - Creates a new Gateway Client
func NewClient(apigeeCfg *config.ApigeeConfig) (*GatewayClient, error) {
	client := &GatewayClient{
		apiClient:    coreapi.NewClient(nil, ""),
		cfg:          apigeeCfg,
		pollInterval: apigeeCfg.GetPollInterval(),
		stopChan:     make(chan struct{}),
	}

	// Start the authentication
	err := client.registerJobs()
	if err != nil {
		return nil, err
	}

	return client, nil
}

// registerJobs - registers the agent jobs
func (a *GatewayClient) registerJobs() error {
	// create the auth job and register it
	authentication := newAuthJob(a.apiClient, a.cfg.Auth.GetUsername(), a.cfg.Auth.GetPassword(), a.setAccessToken)
	jobs.RegisterIntervalJobWithName(authentication, 10*time.Minute, "APIGEE Auth Token")

	// create the channel for portal poller to handler communication
	newPortalChan := make(chan string)
	removedPortalChan := make(chan string)

	// create the portals/portal poller job and register it
	portals := newPollPortalsJob(a, newPortalChan, removedPortalChan)
	jobs.RegisterIntervalJobWithName(portals, a.cfg.GetPollInterval(), "Poll Portals")

	// create the channel for the portal api jobs to handler communication
	apiChan := make(chan *apiDocData)

	// create the portal handler job and register it
	portalHandler := newPortalHandlerJob(a, newPortalChan, removedPortalChan, apiChan)
	jobs.RegisterChannelJobWithName(portalHandler, portalHandler.stopChan, "Portal Handler")

	// create the api handler job and register it
	apiHandler := newPortalAPIHandlerJob(a, apiChan)
	jobs.RegisterChannelJobWithName(apiHandler, apiHandler.stopChan, "New API Handler")

	return nil
}

func (a *GatewayClient) setAccessToken(token string) {
	a.accessToken = token
}

// AgentRunning - waits for a signal to stop the agent
func (a *GatewayClient) AgentRunning() {
	<-a.stopChan
}

// Stop - signals the agent to stop
func (a *GatewayClient) Stop() {
	a.stopChan <- struct{}{}
}

// // DiscoverAPIs - Process the API discovery
// func (a *GatewayClient) DiscoverAPIs() {
// 	// Get the env -> virtual hosts map in case we need to deploy the shared floe
// 	a.updateVirtualHosts()

// 	a.addSharedFlow()
// 	for {
// 		// Update the env -> virtual host mapping
// 		a.updateVirtualHosts()

// 		// Loop all the api proxies
// 		apiProxies := a.getAPIsWithData()
// 		// Get all deployments in all api proxies
// 		for _, proxy := range apiProxies {
// 			deployments := a.getDeployments(proxy.Name)

// 			for _, depEnv := range deployments.Environment {
// 				for _, revision := range depEnv.Revision {
// 					if revision.State == "deployed" {
// 						apigeeProxy := apigeeProxyDetails{
// 							Proxy:       proxy,
// 							Revision:    revision,
// 							Environment: depEnv.Name,
// 						}
// 						a.handleDeployedRevision(apigeeProxy)
// 					}
// 				}
// 			}
// 		}
// 		time.Sleep(a.pollInterval)
// 	}
// }

// // handleDeployedRevision - this is called with each deployed revision
// func (a *GatewayClient) handleDeployedRevision(apigeeProxy apigeeProxyDetails) {
// 	apigeeProxy.APIRevision = a.getRevisionsDetails(apigeeProxy.Proxy.Name, apigeeProxy.Revision.Name)

// 	cacheHash, _ := cache.GetCache().Get(apigeeProxy.GetCacheKey())
// 	if cacheHash != nil {
// 		a.handleExistingProxy(apigeeProxy)
// 	} else {
// 		a.handleNewProxy(apigeeProxy)
// 	}
// }

// func (a *GatewayClient) updateVirtualHosts() {
// 	envToURLs := map[string][]string{}
// 	// Get all virtual host details
// 	environments := a.getEnvironments()
// 	// Loop all environments
// 	for _, env := range environments {
// 		hosts := a.getVirtualHosts(env)
// 		envToURLs[env] = []string{}
// 		// loop all hosts in each environment
// 		for _, host := range hosts {
// 			vHost := a.getVirtualHost(env, host)
// 			for _, alias := range vHost.HostAliases {
// 				basePath := ""
// 				if len(vHost.BaseUrl) > 0 {
// 					basePath = vHost.BaseUrl
// 				}
// 				url := url.URL{
// 					Scheme: "http",
// 					Host:   fmt.Sprintf("%v:%v", alias, vHost.Port),
// 					Path:   basePath,
// 				}
// 				if vHost.SSLInfo.Enabled == "true" {
// 					url.Scheme = "https"
// 				}
// 				envToURLs[env] = append(envToURLs[env], url.String())
// 			}
// 		}
// 	}
// 	a.envToURLs = envToURLs
// }

// func (a *GatewayClient) serviceBodyBuilder(apigeeProxy apigeeProxyDetails) (apic.ServiceBody, error) {
// 	// Create the service body
// 	spec := a.retrieveOrBuildSpec(&apigeeProxy)

// 	// update spec
// 	spec = apigeeProxy.Bundle.UpdateSpec(spec)

// 	authPolicy := apic.Passthrough
// 	if apigeeProxy.Bundle.VerifyAPIKey.Enabled == "true" {
// 		authPolicy = apic.Apikey
// 	}

// 	return apic.NewServiceBodyBuilder().
// 		SetID(apigeeProxy.Proxy.Name).
// 		SetAPIName(apigeeProxy.Proxy.Name).
// 		SetDescription(apigeeProxy.APIRevision.Description).
// 		SetAPISpec(spec).
// 		SetStage(apigeeProxy.Environment).
// 		SetVersion(apigeeProxy.GetVersion()).
// 		SetAuthPolicy(authPolicy).
// 		SetTitle(apigeeProxy.APIRevision.DisplayName).
// 		Build()
// }

// // handleExistingProxy - the details on the proxy that has not yet been added to the cache
// func (a *GatewayClient) handleExistingProxy(data interface{}) {
// 	apigeeProxy := data.(apigeeProxyDetails)
// 	serviceBody, _ := a.serviceBodyBuilder(apigeeProxy)
// 	serviceBodyHash, _ := coreutil.ComputeHash(serviceBody)
// 	cacheHash, _ := cache.GetCache().Get(apigeeProxy.GetCacheKey())
// 	if serviceBodyHash != cacheHash.(uint64) {
// 		serviceBody.APIUpdateSeverity = "MINOR"
// 		agent.PublishAPI(serviceBody)
// 		cache.GetCache().Set(apigeeProxy.GetCacheKey(), serviceBodyHash)
// 	} else {
// 		log.Debug("Current API revision already exists")
// 	}
// }

// //handleNewProxy - the details on the proxy that has not yet been added to the cache
// func (a *GatewayClient) handleNewProxy(data interface{}) {
// 	apigeeProxy := data.(apigeeProxyDetails)
// 	serviceBody, _ := a.serviceBodyBuilder(apigeeProxy)
// 	serviceBodyHash, _ := coreutil.ComputeHash(serviceBody)

// 	if coreagent.IsAPIPublished(serviceBody.RestAPIID) {
// 		publishedMajorHash := util.ConvertStringToUint(agent.GetAttributeOnPublishedAPI(serviceBody.RestAPIID, apigeeProxy.Environment+"Hash"))
// 		if publishedMajorHash == serviceBodyHash {
// 			log.Debugf("No changes detected for API %s in environment %s", serviceBody.APIName, apigeeProxy.Environment)
// 			cache.GetCache().Set(apigeeProxy.GetCacheKey(), serviceBodyHash)
// 			return
// 		}
// 	} else {
// 		log.Infof("Create new API service in AMPLIFY Central for API %s in environment %s", serviceBody.APIName, apigeeProxy.Environment)
// 	}

// 	log.Infof("Published API %s in environment %s to AMPLIFY Central", serviceBody.APIName, apigeeProxy.Environment)
// 	serviceBody.ServiceAttributes[apigeeProxy.Environment+"Hash"] = util.ConvertUnitToString(serviceBodyHash)
// 	serviceBody.ServiceAttributes["GatewayType"] = gatewayType
// 	agent.PublishAPI(serviceBody)
// 	currentHash, _ := coreutil.ComputeHash(serviceBody)
// 	cache.GetCache().Set(apigeeProxy.GetCacheKey(), currentHash)
// }

// //retrieveOrBuildSpec - attempts to retrieve a spec or genrerates a spec if one is not found
// func (a *GatewayClient) retrieveOrBuildSpec(apigeeProxy *apigeeProxyDetails) []byte {
// 	zipBundle := a.getRevisionDefinitionBundle(apigeeProxy.Proxy.Name, apigeeProxy.Revision.Name)
// 	// generate apigeebundle from zip file
// 	apigeeProxy.Bundle = apigeebundle.NewAPIGEEBundle(zipBundle, apigeeProxy.Proxy.Name, a.envToURLs[apigeeProxy.Environment])

// 	// Check the revisionDetails for a value in spec
// 	specString := apigeeProxy.APIRevision.Spec.(string)
// 	if specString != "" {
// 		// The revision has a spec value
// 		if util.IsValidURL(specString) {
// 			// the spec value is a full url, lets attempt a request to get it
// 			response, _ := a.getRequest(specString)
// 			return response.Body
// 		}
// 		// the spec value is not a full url, must be a path in the spec store
// 		return a.getSwagger(specString)
// 	}

// 	// Check the resource files on the revision for a spec link
// 	resourceFiles := a.getResourceFiles(apigeeProxy.Proxy.Name, apigeeProxy.Revision.Name)

// 	// find spec path
// 	var path string
// 	for _, file := range resourceFiles.ResourceFile {
// 		if file.Type == openapi {
// 			path = file.Name
// 			break
// 		}
// 	}

// 	if path != "" {
// 		resourceFileData := a.getRevisionSpec(apigeeProxy.Proxy.Name, apigeeProxy.Revision.Name, path)
// 		// retrieve the spec
// 		var association specAssociationFile
// 		json.Unmarshal(resourceFileData, &association)
// 		return a.getSwagger(association.URL)

// 	}

// 	// Build the spec as a last resort
// 	return apigeeProxy.Bundle.Generate(apigeeProxy.APIRevision.DisplayName, apigeeProxy.APIRevision.Description, apigeeProxy.GetVersion())
// }
