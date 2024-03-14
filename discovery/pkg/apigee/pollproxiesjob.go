package apigee

import (
	"context"
	"encoding/json"
	"fmt"
	"path"
	"sync"

	"github.com/Axway/agent-sdk/pkg/agent"
	"github.com/Axway/agent-sdk/pkg/apic"
	"github.com/Axway/agent-sdk/pkg/apic/provisioning"
	"github.com/Axway/agent-sdk/pkg/jobs"
	coreutil "github.com/Axway/agent-sdk/pkg/util"
	"github.com/Axway/agent-sdk/pkg/util/log"

	"github.com/Axway/agents-apigee/client/pkg/apigee"
	"github.com/Axway/agents-apigee/client/pkg/apigee/models"
	"github.com/Axway/agents-apigee/client/pkg/config"
	"github.com/Axway/agents-apigee/discovery/pkg/util"
)

const (
	gatewayType = "Apigee"

	proxyNameField       ctxKeys = "proxy"
	envNameField         ctxKeys = "environment"
	revNameField         ctxKeys = "revision"
	specPathField        ctxKeys = "specPath"
	hasQuotaPolicyField  ctxKeys = "hasQuota"
	hasAPIKeyPolicyField ctxKeys = "hasAPIKey"
	hasOAuthPolicyField  ctxKeys = "hasOauth"
	endpointsField       ctxKeys = "endpoints"
)

type proxyClient interface {
	GetConfig() *config.ApigeeConfig
	GetAllProxies() (apigee.Proxies, error)
	GetRevision(proxyName, revision string) (*models.ApiProxyRevision, error)
	GetRevisionResourceFile(proxyName, revision, resourceType, resourceName string) ([]byte, error)
	GetRevisionConnectionType(proxyName, revision string) (*apigee.HTTPProxyConnection, error)
	GetDeployments(apiName string) (*models.DeploymentDetails, error)
	GetVirtualHost(envName, virtualHostName string) (*models.VirtualHost, error)
	GetSpecFile(specPath string) ([]byte, error)
	GetSpecFromURL(url string, options ...apigee.RequestOption) ([]byte, error)
	GetRevisionPolicyByName(proxyName, revision, policyName string) (*apigee.PolicyDetail, error)
	IsReady() bool
}

type proxyCache interface {
	GetSpecWithPath(path string) (*specCacheItem, error)
	GetSpecWithName(name string) (*specCacheItem, error)
	GetSpecPathWithEndpoint(endpoint string) (string, error)
	AddPublishedServiceToCache(cacheKey string, serviceBody *apic.ServiceBody)
}

// job that will poll for any new portals on APIGEE Edge
type pollProxiesJob struct {
	jobs.Job
	client      proxyClient
	firstRun    bool
	cache       proxyCache
	logger      log.FieldLogger
	specsReady  jobFirstRunDone
	pubLock     sync.Mutex
	publishFunc agent.PublishAPIFunc
	workers     int
	running     bool
	matchOnURL  bool
	runningLock sync.Mutex
	lastTime    int
	runTime     int
}

func newPollProxiesJob() *pollProxiesJob {
	job := &pollProxiesJob{
		firstRun:    true,
		logger:      log.NewFieldLogger().WithComponent("pollProxies").WithPackage("apigee"),
		publishFunc: agent.PublishAPI,
		runningLock: sync.Mutex{},
	}
	return job
}

func (j *pollProxiesJob) SetSpecClient(client proxyClient) *pollProxiesJob {
	j.client = client
	return j
}

func (j *pollProxiesJob) SetSpecCache(cache proxyCache) *pollProxiesJob {
	j.cache = cache
	return j
}

func (j *pollProxiesJob) SetSpecsReady(specsReady jobFirstRunDone) *pollProxiesJob {
	j.specsReady = specsReady
	return j
}

func (j *pollProxiesJob) SetWorkers(workers int) *pollProxiesJob {
	j.workers = workers
	return j
}

func (j *pollProxiesJob) SetMatchOnURL(matchOnURL bool) *pollProxiesJob {
	j.matchOnURL = matchOnURL
	return j
}

func (j *pollProxiesJob) FirstRunComplete() bool {
	return !j.firstRun
}

func (j *pollProxiesJob) Ready() bool {
	j.logger.Trace("checking if the apigee client is ready for calls")
	if !j.client.IsReady() {
		return false
	}

	j.logger.Trace("checking if specs have been cached")
	return j.specsReady()
}

func (j *pollProxiesJob) Status() error {
	return nil
}

func (j *pollProxiesJob) updateRunning(running bool) {
	j.runningLock.Lock()
	defer j.runningLock.Unlock()
	j.running = running
}

func (j *pollProxiesJob) isRunning() bool {
	j.runningLock.Lock()
	defer j.runningLock.Unlock()
	return j.running
}

func (j *pollProxiesJob) Execute() error {
	j.logger.Trace("executing")

	if j.isRunning() {
		j.logger.Warn("previous proxies poll job run has not completed, will run again on next interval")
		return nil
	}
	j.updateRunning(true)
	defer j.updateRunning(false)

	allProxies, err := j.client.GetAllProxies()
	if err != nil {
		j.logger.WithError(err).Error("getting proxies")
		return err
	}

	limiter := make(chan string, j.workers)

	agent.PublishingLock()
	defer agent.PublishingUnlock()

	wg := sync.WaitGroup{}
	wg.Add(len(allProxies))
	j.runTime = j.lastTime
	for _, proxyName := range allProxies {
		go func() {
			defer wg.Done()
			name := <-limiter
			j.handleProxy(name)
		}()
		limiter <- proxyName
	}

	wg.Wait()
	close(limiter)

	j.firstRun = false
	return nil
}

func (j *pollProxiesJob) handleProxy(proxyName string) {
	logger := j.logger.WithField(proxyNameField.String(), proxyName)
	logger.Debug("handling proxy")

	ctx := addLoggerToContext(context.Background(), logger)
	ctx = context.WithValue(ctx, proxyNameField, proxyName)

	details, err := j.client.GetDeployments(proxyName)
	if err != nil {
		logger.WithError(err).Error("getting deployment")
		return // proxy may not have had any deployments
	}

	wg := sync.WaitGroup{}
	for _, env := range details.Environment {
		wg.Add(1)
		go func(environment models.DeploymentDetailsEnvironment) {
			defer wg.Done()
			j.handleEnvironment(ctx, environment)
		}(env)
	}

	wg.Wait()
}

func (j *pollProxiesJob) handleEnvironment(ctx context.Context, env models.DeploymentDetailsEnvironment) {
	logger := getLoggerFromContext(ctx).WithField(envNameField.String(), env.Name)
	addLoggerToContext(ctx, logger)
	logger.Debug("handling environment")

	ctx = context.WithValue(ctx, envNameField, env.Name)

	wg := sync.WaitGroup{}
	for _, rev := range env.Revision {
		wg.Add(1)
		go func(revName string) {
			defer wg.Done()
			j.handleRevision(ctx, revName)
		}(rev.Name)
	}

	wg.Wait()
}

func (j *pollProxiesJob) handleRevision(ctx context.Context, revName string) {
	logger := getLoggerFromContext(ctx).WithField(revNameField.String(), revName)
	addLoggerToContext(ctx, logger)
	logger.Debug("handling revision")
	pName := getStringFromContext(ctx, proxyNameField)

	_ = pName
	revision, err := j.client.GetRevision(getStringFromContext(ctx, proxyNameField), revName)
	if err != nil {
		logger.WithError(err).Error("getting revision")
		return
	}

	if revision.LastModifiedAt <= j.runTime {
		return
	}
	if j.lastTime < revision.LastModifiedAt {
		j.lastTime = revision.LastModifiedAt
	}

	ctx = context.WithValue(ctx, revNameField, revision)
	logger = logger.WithField(revNameField.String(), revision.Revision)
	addLoggerToContext(ctx, logger)

	ctx = j.checkPolicies(ctx)

	// get URLs
	ctx = j.getVirtualHostURLs(ctx)

	var specURL string
	if revision.Spec != nil && revision.Spec != "" {
		specURL = revision.Spec.(string)
		ctx = context.WithValue(ctx, specPathField, specURL)
	} else {
		specURL = j.specFromRevision(ctx)
		ctx = context.WithValue(ctx, specPathField, specURL)
	}

	if specURL != "" {
		logger = logger.WithField(specPathField.String(), specURL)
		addLoggerToContext(ctx, logger)
		logger.Debug("will download spec from URL in revision")
	}

	j.publish(ctx)
}

func (j *pollProxiesJob) checkPolicies(ctx context.Context) context.Context {
	logger := getLoggerFromContext(ctx)
	logger.Trace("checking revision policies for authentication")
	revision := ctx.Value(revNameField).(*models.ApiProxyRevision)

	for _, p := range revision.Policies {
		logger := logger.WithField("policyName", p)
		logger.Tracef("getting policy details")
		policyDetails, err := j.client.GetRevisionPolicyByName(getStringFromContext(ctx, proxyNameField), revision.Revision, p)
		if err != nil {
			logger.WithError(err).Debug("getting policy")
			continue
		}

		switch policyDetails.PolicyType {
		case quotaPolicy:
			ctx = context.WithValue(ctx, hasQuotaPolicyField, true)
		case apiKeyPolicy:
			ctx = context.WithValue(ctx, hasAPIKeyPolicyField, true)
		case oauthPolicy:
			ctx = context.WithValue(ctx, hasOAuthPolicyField, true)
		}
	}

	return ctx
}

func (j *pollProxiesJob) specFromRevision(ctx context.Context) string {
	logger := getLoggerFromContext(ctx)
	logger.Trace("checking revision resource files")

	// get the spec using the association.json file, if it exists
	revision := ctx.Value(revNameField).(*models.ApiProxyRevision)
	for _, resource := range revision.ResourceFiles.ResourceFile {
		if resource.Type != openapi || resource.Name != association {
			continue
		}
		if path := j.getSpecFromResourceFile(ctx, resource.Type, resource.Name); path != "" {
			return path
		}
	}

	// get a spec match based off the proxy name to the spec name
	specData, _ := j.cache.GetSpecWithName(revision.Name)
	if specData != nil {
		return specData.ContentPath
	}

	return j.getSpecFromVirtualHosts(ctx)
}

func (j *pollProxiesJob) getVirtualHostURLs(ctx context.Context) context.Context {
	logger := getLoggerFromContext(ctx)
	revision := ctx.Value(revNameField).(*models.ApiProxyRevision)
	envName := getStringFromContext(ctx, envNameField)
	proxyName := getStringFromContext(ctx, proxyNameField)
	allURLs := []string{}

	connection, err := j.client.GetRevisionConnectionType(proxyName, revision.Revision)
	if err != nil {
		logger.WithError(err).Error("could not get the revision connection type")
		return context.WithValue(ctx, endpointsField, allURLs)
	}

	virtualHostURLs := make(map[string]map[string][]string)

	if _, ok := virtualHostURLs[envName]; !ok {
		virtualHostURLs[envName] = make(map[string][]string)
	}

	if _, ok := virtualHostURLs[envName][connection.VirtualHost]; !ok {
		virtualHost, err := j.client.GetVirtualHost(envName, connection.VirtualHost)
		if err != nil {
			logger.WithError(err).Error("could not get the virtual host info")
			return context.WithValue(ctx, endpointsField, allURLs)
		}
		virtualHostURLs[envName][connection.VirtualHost] = urlsFromVirtualHost(virtualHost)
	}

	for _, url := range virtualHostURLs[envName][connection.VirtualHost] {
		allURLs = append(allURLs, fmt.Sprintf("%s%s", url, connection.BasePath))
	}

	return context.WithValue(ctx, endpointsField, allURLs)
}

func (j *pollProxiesJob) getSpecFromVirtualHosts(ctx context.Context) string {
	if !j.matchOnURL {
		return ""
	}

	logger := getLoggerFromContext(ctx)
	revision := ctx.Value(revNameField).(*models.ApiProxyRevision)

	urls := ctx.Value(endpointsField).([]string)

	// using the URLs find the first spec that has a match
	for _, url := range urls {
		logger.WithField("host", url)
		for _, path := range revision.Basepaths {
			logger.WithField("path", path)
			path, err := j.cache.GetSpecPathWithEndpoint(url)
			if err != nil {
				logger.WithError(err).Debug("could not get spec with endpoint")
				continue
			}
			logger.Debug("found spec with endpoint")
			return path
		}
	}
	return ""
}

func (j *pollProxiesJob) getSpecFromResourceFile(ctx context.Context, resourceType, resourceName string) string {
	logger := getLoggerFromContext(ctx)
	revision := ctx.Value(revNameField).(*models.ApiProxyRevision)
	logger.Debug("found openapi resource file on revision")

	// get the association.json file content
	resFileContent, err := j.client.GetRevisionResourceFile(getStringFromContext(ctx, proxyNameField), revision.Revision, resourceType, resourceName)
	if err != nil {
		logger.WithError(err).Debug("could not download resource file content")
	}
	associationFile := &Association{}
	err = json.Unmarshal(resFileContent, associationFile)
	if err != nil {
		logger.WithError(err).Debug("could not read resource file content")
	}

	// return the association.json file content
	return associationFile.URL
}

func (j *pollProxiesJob) publish(ctx context.Context) {
	logger := getLoggerFromContext(ctx)
	envName := getStringFromContext(ctx, envNameField)
	revision := ctx.Value(revNameField).(*models.ApiProxyRevision)
	//  start the service Body
	serviceBody, err := j.buildServiceBody(ctx)
	if err != nil {
		logger.WithError(err).Error("building service body")
		return
	}
	if serviceBody == nil {
		return
	}

	serviceBodyHash, _ := coreutil.ComputeHash(*serviceBody)
	hashString := util.ConvertUnitToString(serviceBodyHash)
	cacheKey := createProxyCacheKey(getStringFromContext(ctx, proxyNameField), envName)

	// Check DiscoveryCache for API
	j.pubLock.Lock() // only publish one at a time
	defer j.pubLock.Unlock()
	value := agent.GetAttributeOnPublishedAPIByID(revision.Name, fmt.Sprintf("%s-hash", envName))

	err = nil
	if !agent.IsAPIPublishedByID(revision.Name) {
		// call new API
		err = j.publishAPI(*serviceBody, envName, hashString, cacheKey)
	} else if value != hashString {
		// handle update
		log.Tracef("%s has been updated, push new revision", revision.Name)
		serviceBody.APIUpdateSeverity = "Major"
		serviceBody.SpecDefinition = []byte{}
		log.Tracef("%+v", serviceBody)
		err = j.publishAPI(*serviceBody, envName, hashString, cacheKey)
	}

	if err == nil {
		j.cache.AddPublishedServiceToCache(cacheKey, serviceBody)
	}
}

func (j *pollProxiesJob) buildServiceBody(ctx context.Context) (*apic.ServiceBody, error) {
	logger := getLoggerFromContext(ctx)
	revision := ctx.Value(revNameField).(*models.ApiProxyRevision)
	specPath := getStringFromContext(ctx, specPathField)

	spec, err := j.findSpecFile(specPath, revision)
	// if we should have a spec and can not get it then fall out
	if err != nil {
		logger.WithError(err).WithField("specInfo", specPath).Error("could not gather spec")
		return nil, err
	}

	if len(spec) == 0 && !j.client.GetConfig().Specs.Unstructured {
		log.Warn("skipping proxy creation without a spec")
		return nil, nil
	}
	logger.Debug("creating service body")

	specHash, _ := coreutil.ComputeHash(spec)
	specHashString := coreutil.ConvertUnitToString(specHash)

	// create the agent details with the modification dates
	serviceDetails := map[string]interface{}{
		"specContentHash": specHashString,
	}

	crds := []string{}
	if ctx.Value(hasAPIKeyPolicyField) != nil {
		crds = append(crds, provisioning.APIKeyCRD)
	}
	if ctx.Value(hasOAuthPolicyField) != nil {
		crds = append(crds, provisioning.OAuthSecretCRD)
	}

	urls := ctx.Value(endpointsField).([]string)
	endpoints := createEndpointsFromURLS(urls)

	sb, err := apic.NewServiceBodyBuilder().
		SetID(revision.Name).
		SetAPIName(revision.Name).
		SetStage(getStringFromContext(ctx, envNameField)).
		SetDescription(revision.Description).
		SetAPISpec(spec).
		SetTitle(revision.DisplayName).
		SetVersion(revision.Revision).
		SetAccessRequestDefinitionName(provisioning.APIKeyARD, false).
		SetCredentialRequestDefinitions(crds).
		SetServiceEndpoints(endpoints).
		SetServiceAgentDetails(serviceDetails).
		Build()
	return &sb, err
}

func (j *pollProxiesJob) findSpecFile(specPath string, revision *models.ApiProxyRevision) ([]byte, error) {
	// get the spec to build the service body
	if j.client.GetConfig().Specs.LocalPath != "" {
		specFilePath := path.Join(j.client.GetConfig().Specs.LocalPath, revision.Name)
		spec, err := findSpecFile(j.logger, specFilePath, j.client.GetConfig().Specs.Extensions)
		if len(spec) > 0 && err != nil {
			return spec, err
		}
	}

	if isFullURL(specPath) {
		return j.client.GetSpecFromURL(specPath)
	}

	if specPath != "" {
		// try to get the spec from the APIgee spec repo
		return j.client.GetSpecFile(specPath)
	}

	return nil, nil
}

func (j *pollProxiesJob) publishAPI(serviceBody apic.ServiceBody, envName, hashString, cacheKey string) error {
	// Add a few more attributes to the service body
	serviceBody.ServiceAttributes["GatewayType"] = gatewayType
	serviceBody.ServiceAgentDetails[fmt.Sprintf("%s-hash", envName)] = hashString
	serviceBody.InstanceAgentDetails[cacheKeyAttribute] = cacheKey

	err := j.publishFunc(serviceBody)
	if err == nil {
		log.Infof("Published API %s to AMPLIFY Central", serviceBody.NameToPush)
		return err
	}
	return nil
}
