package apigee

import (
	"github.com/Axway/agent-sdk/pkg/agent"
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
	stopChan        chan struct{}
	agentCache      *agentCache
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
		stopChan:        make(chan struct{}),
		agentCache:      newAgentCache(),
	}

	// newAgent.handleSubscriptions()
	provisioner := NewProvisioner(
		newAgent.apigeeClient,
		agentCfg.CentralCfg.GetCredentialConfig().GetExpirationDays(),
		agent.GetCacheManager(),
	)
	agent.RegisterProvisioner(provisioner)

	return newAgent, nil
}

func (a *Agent) Run() error {
	// Start the agent jobs
	err := a.registerJobs()
	if err != nil {
		return err
	}
	a.running()
	return nil
}

// registerJobs - registers the agent jobs
func (a *Agent) registerJobs() error {
	var err error

	specsJob := newPollSpecsJob(a.apigeeClient, a.agentCache, a.cfg.ApigeeCfg.GetWorkers().Spec)
	_, err = jobs.RegisterIntervalJobWithName(specsJob, a.apigeeClient.GetConfig().GetIntervals().Spec, "Poll Specs")
	if err != nil {
		return err
	}

	var validatorReady jobFirstRunDone

	if a.cfg.ApigeeCfg.IsProxyMode() {
		proxiesJob := newPollProxiesJob(a.apigeeClient, a.agentCache, specsJob.FirstRunComplete, a.cfg.ApigeeCfg.GetWorkers().Proxy)
		_, err = jobs.RegisterIntervalJobWithName(proxiesJob, a.apigeeClient.GetConfig().GetIntervals().Proxy, "Poll Proxies")
		if err != nil {
			return err
		}

		// register the api validator job
		validatorReady = proxiesJob.FirstRunComplete
	} else {
		productsJob := newPollProductsJob(a.apigeeClient, a.agentCache, specsJob.FirstRunComplete, a.cfg.ApigeeCfg.GetWorkers().Product, a.shouldPushAPI)
		_, err = jobs.RegisterIntervalJobWithName(productsJob, a.apigeeClient.GetConfig().GetIntervals().Product, "Poll Products")
		if err != nil {
			return err
		}

		// register the api validator job
		validatorReady = productsJob.FirstRunComplete
	}
	_, err = jobs.RegisterSingleRunJobWithName(newRegisterAPIValidatorJob(validatorReady, a.registerValidator), "Register API Validator")

	agent.NewAPIKeyCredentialRequestBuilder(agent.WithCRDIsSuspendable()).Register()
	agent.NewAPIKeyAccessRequestBuilder().Register()
	agent.NewOAuthCredentialRequestBuilder(agent.WithCRDOAuthSecret(), agent.WithCRDIsSuspendable()).Register()
	return err
}

// running - waits for a signal to stop the agent
func (a *Agent) running() {
	<-a.stopChan
}

// Stop - signals the agent to stop
func (a *Agent) Stop() {
	a.stopChan <- struct{}{}
}

// shouldPushAPI - callback used determine if the Product should be pushed to Central or not
func (a *Agent) shouldPushAPI(attributes map[string]string) bool {
	// Evaluate the filter condition
	return a.discoveryFilter.Evaluate(attributes)
}

// apiValidator - registers the agent jobs
func (a *Agent) apiValidator(proxyName, envName string) bool {
	// get the api with the product name and portal name
	return true
}

func (a *Agent) registerValidator() {
	agent.RegisterAPIValidator(a.apiValidator)
}
