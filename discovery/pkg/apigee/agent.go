package apigee

import (
	"time"

	"github.com/Axway/agent-sdk/pkg/agent"
	"github.com/Axway/agent-sdk/pkg/apic"
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
}

// NewAgent - Creates a new Agent
func NewAgent(apigeeCfg *config.ApigeeConfig) (*Agent, error) {
	apigeeClient, err := apigee.NewClient(apigeeCfg)
	if err != nil {
		return nil, err
	}

	agent := &Agent{
		apigeeClient: apigeeClient,
		cfg:          apigeeCfg,
		pollInterval: apigeeCfg.GetPollInterval(),
		stopChan:     make(chan struct{}),
	}

	// Start the agent jobs
	err = agent.registerJobs()
	if err != nil {
		return nil, err
	}

	agent.handleSubscriptions()

	return agent, nil
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

	// create job that gets the developers

	// create job that gets the apps

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

// registerSubscriptionSchema - create a subscription schema on Central
func (a *Agent) registerSubscriptionSchema() {
	apic.NewSubscriptionSchemaBuilder(agent.GetCentralClient()).
		SetName(defaultSubscriptionSchema).
		AddProperty(apic.NewSubscriptionSchemaPropertyBuilder().
			SetName(appNameKey).
			SetRequired().
			IsString()).
		Register()
}

// handleSubscriptions - setup all things necessary to handle subscriptions from Central
func (a *Agent) handleSubscriptions() {
	a.registerSubscriptionSchema()

	agent.GetCentralClient().GetSubscriptionManager()

	agent.GetCentralClient().GetSubscriptionManager().RegisterProcessor(apic.SubscriptionApproved, a.processSubscribe)
	// agent.GetCentralClient().GetSubscriptionManager().RegisterProcessor(apic.SubscriptionUnsubscribeInitiated, a.processUnsubscribe)
	// agent.GetCentralClient().GetSubscriptionManager().RegisterValidator(a.validateSubscription)
	agent.GetCentralClient().GetSubscriptionManager().Start()
}

func (a *Agent) processSubscribe(sub apic.Subscription) {
	apiAttributes := sub.GetRemoteAPIAttributes()
	c := cache.GetCache()
	api, err := cache.GetCache().Get(apiAttributes[catalogIDKey])
	_ = c
	_ = api
	_ = err

	// get product by name

	// get app by name
	return
}
