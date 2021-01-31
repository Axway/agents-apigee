package apigee

import (
	"encoding/json"
	"net/url"
	"time"

	coreapi "github.com/Axway/agent-sdk/pkg/api"
	"github.com/Axway/agent-sdk/pkg/util/log"
	"github.com/Axway/agents-apigee/discovery/pkg/apigee/models"
	"github.com/Axway/agents-apigee/discovery/pkg/config"
)

const (
	apigeeAuthURL   = "https://login.apigee.com/oauth/token"
	apigeeAuthToken = "ZWRnZWNsaTplZGdlY2xpc2VjcmV0" //hardcoded to edgecli:edgeclisecret
	orgURL          = "https://api.enterprise.apigee.com/v1/organizations/%s/"
)

// GatewayClient - Represents the Gateway client
type GatewayClient struct {
	cfg         *config.ApigeeConfig
	apiClient   coreapi.Client
	accessToken string
}

// NewClient - Creates a new Gateway Client
func NewClient(apigeeCfg *config.ApigeeConfig) (*GatewayClient, error) {
	gatewayClient := &GatewayClient{
		apiClient: coreapi.NewClient(nil, ""),
		cfg:       apigeeCfg,
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

// ExternalAPI - Sample struct representing the API definition in API gateway
type ExternalAPI struct {
	swaggerSpec   []byte
	id            string
	name          string
	description   string
	version       string
	url           string
	documentation []byte
}

// DiscoverAPIs - Process the API discovery
func (a *GatewayClient) DiscoverAPIs() error {
	// Gateway specific implementation to get the details for discovered API goes here
	// Set the service definition
	// As sample the implementation reads the swagger for musical-instrument from local directory

	hostsPerEnv := map[string][]models.VirtualHost{}
	// Get all virtual host details
	environments := a.getEnvironments()
	// Loop all enviornments
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
				resourceFiles := a.getResourceFiles(proxy, revision.Name)
				// revisionDetails := a.getRevisionsDetails(proxy, revision.Name)

				// find spec path
				var path string
				for _, file := range resourceFiles.ResourceFile {
					if file.Type == "openapi" {
						path = file.Name
						break
					}
				}

				if path != "" {
					resourceFileData := a.getRevisionSpec(proxy, revision.Name, path)
					// retrieve the spec
					var association specAssociationFile
					json.Unmarshal(resourceFileData, &association)
					spec := a.getSwagger(association.URL)

					log.Debugf(spec)
				}
			}
		}

		log.Debug("end")
	}

	return nil
	// swaggerSpec, err := a.getSpec()
	// if err != nil {
	// 	log.Infof("Failed to load sample API specification from %s: %s ", a.cfg.SpecPath, err.Error())
	// }

	// externalAPI := ExternalAPI{
	// 	id:            "65c79285-f550-4617-bf6e-003e617841f2",
	// 	name:          "Musical-Instrument-Sample",
	// 	description:   "Sample for API discovery agent",
	// 	version:       "1.0.0",
	// 	url:           "",
	// 	documentation: []byte("\"Sample documentation for API discovery agent\""),
	// 	swaggerSpec:   swaggerSpec,
	// }

	// serviceBody, err := a.buildServiceBody(externalAPI)
	// if err != nil {
	// 	return err
	// }
	// err = agent.PublishAPI(serviceBody)
	// if err != nil {
	// 	return err
	// }
	// log.Info("Published API " + serviceBody.APIName + "to AMPLIFY Central")
	// return err
}

/*
// buildServiceBody - creates the service definition
func (a *GatewayClient) buildServiceBody(externalAPI ExternalAPI) (apic.ServiceBody, error) {
	return apic.NewServiceBodyBuilder().
		SetID(externalAPI.id).
		SetTitle(externalAPI.name).
		SetURL(externalAPI.url).
		SetDescription(externalAPI.description).
		SetAPISpec(externalAPI.swaggerSpec).
		SetVersion(externalAPI.version).
		SetAuthPolicy(apic.Passthrough).
		SetDocumentation(externalAPI.documentation).
		SetResourceType(apic.Oas2).
		Build()
}

func (a *GatewayClient) getSpec() ([]byte, error) {
	bytes, err := ioutil.ReadFile(a.cfg.SpecPath)
	if err != nil {
		return nil, err
	}
	return bytes, nil
}
*/
