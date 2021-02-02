package apigee

import (
	"archive/zip"
	"bytes"
	"encoding/json"
	"fmt"
	"net/url"
	"strings"
	"time"

	"github.com/getkin/kin-openapi/openapi3"

	"github.com/Axway/agent-sdk/pkg/agent"
	coreagent "github.com/Axway/agent-sdk/pkg/agent"
	coreapi "github.com/Axway/agent-sdk/pkg/api"
	"github.com/Axway/agent-sdk/pkg/apic"
	"github.com/Axway/agent-sdk/pkg/cache"
	coreutil "github.com/Axway/agent-sdk/pkg/util"
	"github.com/Axway/agent-sdk/pkg/util/log"

	"github.com/Axway/agents-apigee/discovery/pkg/apigee/generatespec"
	"github.com/Axway/agents-apigee/discovery/pkg/apigee/models"
	"github.com/Axway/agents-apigee/discovery/pkg/config"
	"github.com/Axway/agents-apigee/discovery/pkg/util"
)

const (
	apigeeAuthURL      = "https://login.apigee.com/oauth/token"
	apigeeAuthToken    = "ZWRnZWNsaTplZGdlY2xpc2VjcmV0" //hardcoded to edgecli:edgeclisecret
	orgURL             = "https://api.enterprise.apigee.com/v1/organizations/%s/"
	openapi            = "openapi"
	gatewayType        = "APIGEE"
	newProxyTopic      = "newProxy"
	existingProxyTopic = "existingProxy"
	topicErr           = "Error creating topic %v: %v"
)

// GatewayClient - Represents the Gateway client
type GatewayClient struct {
	cfg               *config.ApigeeConfig
	apiClient         coreapi.Client
	accessToken       string
	pollInterval      time.Duration
	virtualHostsToEnv map[string][]models.VirtualHost
}

// NewClient - Creates a new Gateway Client
func NewClient(apigeeCfg *config.ApigeeConfig) (*GatewayClient, error) {
	client := &GatewayClient{
		apiClient:    coreapi.NewClient(nil, ""),
		cfg:          apigeeCfg,
		pollInterval: apigeeCfg.GetPollInterval(),
	}

	// Start the authentication
	client.Authenticate()

	return client, nil
}

// Authenticate - handles the initial authentication then starts a go routine to refresh the token
func (a *GatewayClient) Authenticate() error {
	authData := url.Values{}
	authData.Set("grant_type", password.String())
	authData.Set("username", a.cfg.GetAuth().GetUsername())
	authData.Set("password", a.cfg.GetAuth().GetPassword())

	authResponse := a.postAuth(authData)

	log.Debugf("APIGEE auth token: %s", authResponse.AccessToken)

	// Continually refresh the token
	go func() {
		for {
			// Refresh the token 5 minutes before expiration
			time.Sleep(time.Duration(authResponse.ExpiresIn-300) * time.Second)

			log.Debug("Refreshing auth token")
			authData := url.Values{}
			authData.Set("grant_type", refresh.String())
			authData.Set("refresh_token", authResponse.RefreshToken)

			authResponse = a.postAuth(authData)
			log.Debugf("APIGEE auth token: %s", authResponse.AccessToken)
		}
	}()

	return nil
}

// DiscoverAPIs - Process the API discovery
func (a *GatewayClient) DiscoverAPIs() {
	for {
		// Update the virtual host to environment mapping
		a.updateVirtualHosts()

		// Loop all the api proxies
		apiProxies := a.getAPIsWithData()
		// Get all deployments in all api proxies
		for _, proxy := range apiProxies {
			deployments := a.getDeployments(proxy.Name)

			for _, depEnv := range deployments.Environment {
				for _, revision := range depEnv.Revision {
					if revision.State == "deployed" {
						apigeeProxy := apigeeProxyDetails{
							Proxy:       proxy,
							Revision:    revision,
							Environment: depEnv.Name,
						}
						go a.handleDeployedRevision(apigeeProxy)
					}
				}
			}
		}
		time.Sleep(a.pollInterval)
	}
}

// handleDeployedRevision - this is called with each deployed revision
func (a *GatewayClient) handleDeployedRevision(apigeeProxy apigeeProxyDetails) {
	apigeeProxy.APIRevision = a.getRevisionsDetails(apigeeProxy.Proxy.Name, apigeeProxy.Revision.Name)
	a.retrieveOrBuildSpec(&apigeeProxy)

	cacheHash, _ := cache.GetCache().Get(apigeeProxy.GetCacheKey())
	if cacheHash != nil {
		go a.handleExistingProxy(apigeeProxy)
	} else {
		go a.handleNewProxy(apigeeProxy)
	}
}

func (a *GatewayClient) updateVirtualHosts() {
	virtualHostsToEnv := map[string][]models.VirtualHost{}
	// Get all virtual host details
	environments := a.getEnvironments()
	// Loop all environments
	for _, env := range environments {
		hosts := a.getVirtualHosts(env)
		virtualHostsToEnv[env] = []models.VirtualHost{}
		// loop all hosts in each environment
		for _, host := range hosts {
			virtualHostsToEnv[env] = append(virtualHostsToEnv[env], a.getVirtualHost(env, host))
		}
	}
	a.virtualHostsToEnv = virtualHostsToEnv
}

func (a *GatewayClient) serviceBodyBuilder(apigeeProxy apigeeProxyDetails) (apic.ServiceBody, error) {
	// Create the service body
	return apic.NewServiceBodyBuilder().
		SetID(apigeeProxy.Proxy.Name).
		SetAPIName(apigeeProxy.Proxy.Name).
		SetDescription(apigeeProxy.APIRevision.Description).
		SetAPISpec(apigeeProxy.Spec).
		SetStage(apigeeProxy.Environment).
		SetVersion(apigeeProxy.GetVersion()).
		SetAuthPolicy(apic.Passthrough).
		SetTitle(apigeeProxy.APIRevision.DisplayName).
		Build()
}

// handleExistingProxy - the details on the proxy that has not yet been added to the cache
func (a *GatewayClient) handleExistingProxy(data interface{}) {
	apigeeProxy := data.(apigeeProxyDetails)
	serviceBody, _ := a.serviceBodyBuilder(apigeeProxy)
	serviceBodyHash, _ := coreutil.ComputeHash(serviceBody)
	cacheHash, _ := cache.GetCache().Get(apigeeProxy.GetCacheKey())
	if serviceBodyHash != cacheHash.(uint64) {
		serviceBody.APIUpdateSeverity = "MINOR"
		agent.PublishAPI(serviceBody)
		cache.GetCache().Set(apigeeProxy.GetCacheKey(), serviceBodyHash)
	} else {
		log.Debug("Current API revision already exists")
	}
}

//handleNewProxy - the details on the proxy that has not yet been added to the cache
func (a *GatewayClient) handleNewProxy(data interface{}) {
	apigeeProxy := data.(apigeeProxyDetails)
	serviceBody, _ := a.serviceBodyBuilder(apigeeProxy)
	serviceBodyHash, _ := coreutil.ComputeHash(serviceBody)

	if coreagent.IsAPIPublished(serviceBody.RestAPIID) {
		publishedMajorHash := util.ConvertStringToUint(agent.GetAttributeOnPublishedAPI(serviceBody.RestAPIID, apigeeProxy.Environment+"Hash"))
		if publishedMajorHash == serviceBodyHash {
			log.Debugf("No changes detected for API %s in environment %s", serviceBody.APIName, apigeeProxy.Environment)
			cache.GetCache().Set(apigeeProxy.GetCacheKey(), serviceBodyHash)
			return
		}
	} else {
		log.Infof("Create new API service in AMPLIFY Central for API %s in environment %s", serviceBody.APIName, apigeeProxy.Environment)
	}

	log.Infof("Published API %s in environment %s to AMPLIFY Central", serviceBody.APIName, apigeeProxy.Environment)
	serviceBody.ServiceAttributes[apigeeProxy.Environment+"Hash"] = util.ConvertUnitToString(serviceBodyHash)
	serviceBody.ServiceAttributes["GatewayType"] = gatewayType
	agent.PublishAPI(serviceBody)
	currentHash, _ := coreutil.ComputeHash(serviceBody)
	cache.GetCache().Set(apigeeProxy.GetCacheKey(), currentHash)
}

//retrieveOrBuildSpec - attempts to retrieve a spec or genrerates a spec if one is not found
func (a *GatewayClient) retrieveOrBuildSpec(apigeeProxy *apigeeProxyDetails) {
	// Check the revisionDetails for a value in spec
	specString := apigeeProxy.APIRevision.Spec.(string)
	if specString != "" {
		// The revision has a spec value
		if util.IsValidURL(specString) {
			// the spec value is a full url, lets attempt a request to get it
			response, _ := a.getRequest(specString)
			apigeeProxy.Spec = response.Body
			return
		}
		// the spec value is not a full url, must be a path in the spec store
		apigeeProxy.Spec = a.getSwagger(specString)
		return
	}

	// Check the resource files on the revision for a spec link
	resourceFiles := a.getResourceFiles(apigeeProxy.Proxy.Name, apigeeProxy.Revision.Name)

	// find spec path
	var path string
	for _, file := range resourceFiles.ResourceFile {
		if file.Type == openapi {
			path = file.Name
			break
		}
	}

	if path != "" {
		resourceFileData := a.getRevisionSpec(apigeeProxy.Proxy.Name, apigeeProxy.Revision.Name, path)
		// retrieve the spec
		var association specAssociationFile
		json.Unmarshal(resourceFileData, &association)
		apigeeProxy.Spec = a.getSwagger(association.URL)
		return
	}

	// Build the spec as a last resort
	apigeeProxy.Spec = a.generateSpecFile(a.getRevisionDefinitionBundle(apigeeProxy.Proxy.Name, apigeeProxy.Revision.Name), apigeeProxy.APIRevision)
}

func (a *GatewayClient) generateSpecFile(data []byte, revisionDetails models.ApiProxyRevision) []byte {
	// data is the byte array of the zip archive
	spec := openapi3.Swagger{
		OpenAPI: "3.0.1",
		Info: &openapi3.Info{
			Title:       revisionDetails.DisplayName,
			Description: revisionDetails.Description,
			Version:     fmt.Sprintf("%v.%v", revisionDetails.ConfigurationVersion.MajorVersion, revisionDetails.ConfigurationVersion.MinorVersion),
		},
		Paths: openapi3.Paths{},
	}

	zipReader, err := zip.NewReader(bytes.NewReader(data), int64(len(data)))
	if err != nil {
		log.Error(err)
	}

	// Read all the files from zip archive
	for _, zipFile := range zipReader.File {
		// we only care about the files in proxies
		if strings.HasPrefix(zipFile.Name, "apiproxy/proxies/") && strings.HasSuffix(zipFile.Name, ".xml") {
			fileBytes, err := util.ReadZipFile(zipFile)
			if err != nil {
				log.Error(err)
				continue
			}
			generatespec.GenerateEndpoints(&spec, fileBytes)
		}
	}

	specBytes, _ := json.Marshal(spec)
	return specBytes
}
