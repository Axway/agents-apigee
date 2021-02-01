package apigee

import (
	"archive/zip"
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/url"
	"time"

	"github.com/Axway/agent-sdk/pkg/agent"
	coreapi "github.com/Axway/agent-sdk/pkg/api"
	"github.com/Axway/agent-sdk/pkg/apic"
	"github.com/Axway/agent-sdk/pkg/util/log"
	"github.com/getkin/kin-openapi/openapi3"

	"github.com/Axway/agents-apigee/discovery/pkg/apigee/generatespec"
	"github.com/Axway/agents-apigee/discovery/pkg/apigee/models"
	"github.com/Axway/agents-apigee/discovery/pkg/config"
	"github.com/Axway/agents-apigee/discovery/pkg/util"
)

const (
	apigeeAuthURL   = "https://login.apigee.com/oauth/token"
	apigeeAuthToken = "ZWRnZWNsaTplZGdlY2xpc2VjcmV0" //hardcoded to edgecli:edgeclisecret
	orgURL          = "https://api.enterprise.apigee.com/v1/organizations/%s/"
	openapi         = "openapi"
)

// GatewayClient - Represents the Gateway client
type GatewayClient struct {
	cfg          *config.ApigeeConfig
	apiClient    coreapi.Client
	accessToken  string
	pollInterval time.Duration
}

// NewClient - Creates a new Gateway Client
func NewClient(apigeeCfg *config.ApigeeConfig) (*GatewayClient, error) {
	gatewayClient := &GatewayClient{
		apiClient:    coreapi.NewClient(nil, ""),
		cfg:          apigeeCfg,
		pollInterval: apigeeCfg.GetPollInterval(),
	}

	// Start the authentication
	gatewayClient.Authenticate()

	return gatewayClient, nil
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
		hostsPerEnv := map[string][]models.VirtualHost{}
		// Get all virtual host details
		environments := a.getEnvironments()
		// Loop all environments
		for _, env := range environments {
			hosts := a.getVirtualHosts(env)
			hostsPerEnv[env] = []models.VirtualHost{}
			// loop all hosts in each environment
			for _, host := range hosts {
				hostsPerEnv[env] = append(hostsPerEnv[env], a.getVirtualHost(env, host))
			}
		}

		apiProxies := a.getAPIs()
		// Get all deployments in all api proxies
		for _, proxy := range apiProxies {
			proxyDetails := a.getAPI(proxy)
			deployments := a.getDeployments(proxy)

			// Get an array of deployed revisions only
			deployedRevisions := []environmentRevision{}
			for _, depEnv := range deployments.Environment {
				envRev := environmentRevision{
					Name:      depEnv.Name,
					Revisions: []models.DeploymentDetailsRevision{},
				}
				for _, revision := range depEnv.Revision {
					if revision.State == "deployed" {
						envRev.Revisions = append(envRev.Revisions, revision)
					}
				}
				deployedRevisions = append(deployedRevisions, envRev)
			}

			for _, depRev := range deployedRevisions {
				for _, revision := range depRev.Revisions {
					revisionDetails := a.getRevisionsDetails(proxy, revision.Name)
					spec := a.retrieveOrBuildSpec(proxyDetails, revision.Name, revisionDetails)

					serviceBody, _ := apic.NewServiceBodyBuilder().
						SetID(proxyDetails.Name).
						SetAPIName(proxyDetails.Name).
						SetDescription(revisionDetails.Description).
						SetAPISpec(spec).
						SetStage(depRev.Name).
						SetVersion(fmt.Sprintf("%d.%d", revisionDetails.ConfigurationVersion.MajorVersion, revisionDetails.ConfigurationVersion.MinorVersion)).
						SetAuthPolicy(apic.Passthrough).
						SetTitle(revisionDetails.DisplayName).
						Build()

					agent.PublishAPI(serviceBody)
					log.Info("Published API " + serviceBody.APIName + " to AMPLIFY Central")
				}
			}
		}
		time.Sleep(a.pollInterval)
		return
	}
}

//retrieveOrBuildSpec - attempts to retrieve a spec or genrerates a spec if one is not found
func (a *GatewayClient) retrieveOrBuildSpec(proxy models.ApiProxy, revisionName string, revisionDetails models.ApiProxyRevision) []byte {
	// Check the revisionDetails for a value in spec
	specString := revisionDetails.Spec.(string)
	if revisionDetails.Spec != "" {
		// The revision has a spec value
		if util.IsValidURL(specString) {
			// the spec value is a full url, lets attempt a request to get it
			response, _ := a.getRequest(specString)
			return response.Body
		}
		// the spec value is not a full url, must be a path in the spec store
		return a.getSwagger(specString)
	}

	// Check the resource files on the revision for a spec link
	resourceFiles := a.getResourceFiles(proxy.Name, revisionName)

	// find spec path
	var path string
	for _, file := range resourceFiles.ResourceFile {
		if file.Type == openapi {
			path = file.Name
			break
		}
	}

	if path != "" {
		resourceFileData := a.getRevisionSpec(proxy.Name, revisionName, path)
		// retrieve the spec
		var association specAssociationFile
		json.Unmarshal(resourceFileData, &association)
		return a.getSwagger(association.URL)
	}

	// Build the spec as a last resort
	return a.generateSpecFile(a.getRevisionDefinitionBundle(proxy.Name, revisionName), revisionDetails)
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
		for _, proxyFile := range revisionDetails.Proxies {
			if zipFile.Name == fmt.Sprintf("apiproxy/proxies/%s.xml", proxyFile) {
				fileBytes, err := readZipFile(zipFile)
				if err != nil {
					log.Error(err)
					break
				}
				generatespec.GenerateEndpoints(&spec, fileBytes)
				break
			}
		}
	}

	specBytes, _ := json.Marshal(spec)
	return specBytes
}

func readZipFile(zf *zip.File) ([]byte, error) {
	f, err := zf.Open()
	if err != nil {
		return nil, err
	}
	defer f.Close()
	return ioutil.ReadAll(f)
}
