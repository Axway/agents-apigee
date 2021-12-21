package apigee

import (
	"fmt"
	"time"

	"github.com/Axway/agent-sdk/pkg/agent"
	"github.com/Axway/agent-sdk/pkg/cache"
	corecfg "github.com/Axway/agent-sdk/pkg/config"
	"github.com/Axway/agent-sdk/pkg/filter"
	"github.com/Axway/agent-sdk/pkg/jobs"

	"github.com/Axway/agents-apigee/client/pkg/apigee"
	"github.com/Axway/agents-apigee/client/pkg/config"
)

// AgentConfig - represents the config for agent
type AgentConfig struct {
	CentralCfg corecfg.CentralConfig `config:"central"`
	ApigeeCfg  *config.ApigeeConfig  `config:"apigee"`
}

// Agent - Represents the Gateway client
type Agent struct {
	cfg             *AgentConfig
	apigeeClient    *apigee.ApigeeClient
	discoveryFilter filter.Filter
	pollInterval    time.Duration
	stopChan        chan struct{}
	devCreated      bool
}

// NewAgent - Creates a new Agent
func NewAgent(agentCfg *AgentConfig) (*Agent, error) {
	apigeeClient, err := apigee.NewClient(agentCfg.ApigeeCfg)
	if err != nil {
		return nil, err
	}

	discoveryFilter, err := filter.NewFilter(agentCfg.ApigeeCfg.Filter)
	if err != nil {
		return nil, err
	}

	newAgent := &Agent{
		apigeeClient:    apigeeClient,
		cfg:             agentCfg,
		discoveryFilter: discoveryFilter,
		pollInterval:    agentCfg.ApigeeCfg.GetPollInterval(),
		stopChan:        make(chan struct{}),
	}

	// Start the agent jobs
	err = newAgent.registerJobs()
	if err != nil {
		return nil, err
	}

	newAgent.handleSubscriptions()

	// delay the start of the API validator
	go func() {
		// allow 2 poll intervals before starting validator
		time.Sleep(newAgent.pollInterval * 2)
		agent.RegisterAPIValidator(newAgent.apiValidator)
	}()

	return newAgent, nil
}

// registerJobs - registers the agent jobs
func (a *Agent) registerJobs() error {
	var err error
	// create the communication channels
	channels := &agentChannels{
		productChan:       make(chan productRequest),
		newPortalChan:     make(chan string),
		removedPortalChan: make(chan string),
		processAPIChan:    make(chan *apigee.APIDocData),
		removedAPIChan:    make(chan string),
	}

	// create the product handler job and register it
	productHandler := newProductHandlerJob(a.apigeeClient, channels)
	_, err = jobs.RegisterChannelJobWithName(productHandler, productHandler.stopChan, "Product Handler")
	if err != nil {
		return err
	}

	// create the portals/portal poller job and register it
	_, err = jobs.RegisterIntervalJobWithName(newPollPortalsJob(a.apigeeClient, channels), a.pollInterval, "Poll Portals")
	if err != nil {
		return err
	}

	// create the portal handler job and register it
	portalHandler := newPortalHandlerJob(a.apigeeClient, channels)
	_, err = jobs.RegisterChannelJobWithName(portalHandler, portalHandler.stopChan, "Portal Handler")
	if err != nil {
		return err
	}

	// create the api handler job and register it
	apiHandler := newPortalAPIHandlerJob(a.apigeeClient, channels, a.shouldPushAPI)
	_, err = jobs.RegisterChannelJobWithName(apiHandler, apiHandler.stopChan, "New API Handler")
	if err != nil {
		return err
	}

	// create job that creates the developer profile used by the agent
	_, err = jobs.RegisterSingleRunJobWithName(newCreateDeveloperJob(a.apigeeClient, a.apigeeClient.SetDeveloperID), "Create Developer")
	if err != nil {
		return err
	}

	// create job that starts the subscription manager
	_, err = jobs.RegisterSingleRunJobWithName(newStartSubscriptionManager(a.apigeeClient, a.apigeeClient.GetDeveloperID), "Start Subscription Manager")

	return err
}

// AgentRunning - waits for a signal to stop the agent
func (a *Agent) AgentRunning() {
	<-a.stopChan
}

// Stop - signals the agent to stop
func (a *Agent) Stop() {
	a.stopChan <- struct{}{}
}

// apiValidator - registers the agent jobs
func (a *Agent) apiValidator(productName, portalName string) bool {
	// get the api with the product name and portal name
	cacheKey := fmt.Sprintf("%s-%s", portalName, productName)

	tmp := cache.GetCache()
	_, err := cache.GetCache().GetBySecondaryKey(cacheKey)
	if err != nil {
		return false // api has been removed
	}
	_ = tmp

	return true
}

// setDeveloperCreated - once the developer creation job runs and sets that it is created
func (a *Agent) setDeveloperCreated() {
	a.devCreated = true
}

func (a *Agent) isDevCreated() bool {
	return a.devCreated
}

// shouldPushAPI - callback used determine if the Product should be pushed to Central or not
func (a *Agent) shouldPushAPI(attributes map[string]string) bool {
	// Evaluate the filter condition
	return a.discoveryFilter.Evaluate(attributes)
}
