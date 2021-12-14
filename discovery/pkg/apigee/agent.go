package apigee

import (
	"fmt"
	"time"

	"github.com/Axway/agent-sdk/pkg/agent"
	"github.com/Axway/agent-sdk/pkg/cache"
	corecfg "github.com/Axway/agent-sdk/pkg/config"
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
	cfg          *config.ApigeeConfig
	apigeeClient *apigee.ApigeeClient
	pollInterval time.Duration
	stopChan     chan struct{}
	devCreated   bool
}

// NewAgent - Creates a new Agent
func NewAgent(apigeeCfg *config.ApigeeConfig) (*Agent, error) {
	apigeeClient, err := apigee.NewClient(apigeeCfg)
	if err != nil {
		return nil, err
	}

	newAgent := &Agent{
		apigeeClient: apigeeClient,
		cfg:          apigeeCfg,
		pollInterval: apigeeCfg.GetPollInterval(),
		stopChan:     make(chan struct{}),
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
	// create the channel for portal poller to handler communication
	newPortalChan := make(chan string)
	removedPortalChan := make(chan string)

	// create the portals/portal poller job and register it
	portals := newPollPortalsJob(a.apigeeClient, newPortalChan, removedPortalChan)
	jobs.RegisterIntervalJobWithName(portals, a.cfg.GetPollInterval(), "Poll Portals")

	// create the channel for the portal api jobs to handler communication
	processAPIChan := make(chan *apigee.APIDocData)
	removedAPIChan := make(chan string)

	// create the portal handler job and register it
	portalHandler := newPortalHandlerJob(a.apigeeClient, newPortalChan, removedPortalChan, removedAPIChan, processAPIChan)
	jobs.RegisterChannelJobWithName(portalHandler, portalHandler.stopChan, "Portal Handler")

	// create the api handler job and register it
	apiHandler := newPortalAPIHandlerJob(a.apigeeClient, processAPIChan, removedAPIChan)
	jobs.RegisterChannelJobWithName(apiHandler, apiHandler.stopChan, "New API Handler")

	// create job that creates the developer profile used by the agent
	jobs.RegisterSingleRunJobWithName(newCreateDeveloperJob(a.apigeeClient, a.apigeeClient.SetDeveloperID), "Create Developer")

	// create job that start the subscription manager
	jobs.RegisterSingleRunJobWithName(newStartSubscriptionManager(a.apigeeClient, a.apigeeClient.GetDeveloperID), "Start Subscription Manager")

	return nil
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
