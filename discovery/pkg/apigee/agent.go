package apigee

import (
	"time"

	"github.com/Axway/agent-sdk/pkg/agent"
	"github.com/Axway/agent-sdk/pkg/apic"
	"github.com/Axway/agent-sdk/pkg/jobs"

	"github.com/Axway/agents-apigee/discovery/pkg/config"
)

// Agent - Represents the Gateway client
type Agent struct {
	cfg          *config.ApigeeConfig
	apigeeClient *GatewayClient
	pollInterval time.Duration
	stopChan     chan struct{}
}

// NewAgent - Creates a new Agent
func NewAgent(apigeeCfg *config.ApigeeConfig) (*Agent, error) {
	apigeeClient, err := NewClient(apigeeCfg)
	if err != nil {
		return nil, err
	}

	agent := &Agent{
		apigeeClient: apigeeClient,
		cfg:          apigeeCfg,
		pollInterval: apigeeCfg.GetPollInterval(),
		stopChan:     make(chan struct{}),
	}

	agent.handleSubscriptions()

	// Start the agent jobs
	err = agent.registerJobs()
	if err != nil {
		return nil, err
	}

	return agent, nil
}

// registerJobs - registers the agent jobs
func (a *Agent) registerJobs() error {
	// create the auth job and register it
	authentication := newAuthJob(a.apigeeClient.apiClient, a.cfg.Auth.GetUsername(), a.cfg.Auth.GetPassword(), a.setAccessToken)
	jobs.RegisterIntervalJobWithName(authentication, 10*time.Minute, "APIGEE Auth Token")

	// create the channel for portal poller to handler communication
	newPortalChan := make(chan string)
	removedPortalChan := make(chan string)

	// create the portals/portal poller job and register it
	portals := newPollPortalsJob(a.apigeeClient, newPortalChan, removedPortalChan)
	jobs.RegisterIntervalJobWithName(portals, a.cfg.GetPollInterval(), "Poll Portals")

	// create the channel for the portal api jobs to handler communication
	newAPIChan := make(chan *apiDocData)
	removedAPIChan := make(chan string)

	// create the portal handler job and register it
	portalHandler := newPortalHandlerJob(a.apigeeClient, newPortalChan, removedPortalChan, removedAPIChan, newAPIChan)
	jobs.RegisterChannelJobWithName(portalHandler, portalHandler.stopChan, "Portal Handler")

	// create the api handler job and register it
	apiHandler := newPortalAPIHandlerJob(a.apigeeClient, newAPIChan, removedAPIChan)
	jobs.RegisterChannelJobWithName(apiHandler, apiHandler.stopChan, "New API Handler")

	return nil
}

func (a *Agent) setAccessToken(token string) {
	a.apigeeClient.setAccessToken(token)
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
	return
}
