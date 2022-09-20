package apigee

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"

	"github.com/Axway/agent-sdk/pkg/agent"
	"github.com/Axway/agent-sdk/pkg/apic"
	"github.com/Axway/agent-sdk/pkg/jobs"
	coreutil "github.com/Axway/agent-sdk/pkg/util"
	"github.com/Axway/agent-sdk/pkg/util/log"

	"github.com/Axway/agents-apigee/client/pkg/apigee"
	"github.com/Axway/agents-apigee/client/pkg/apigee/models"
	"github.com/Axway/agents-apigee/discovery/pkg/util"
)

const (
	gatewayType = "Apigee"

	proxyNameField ctxKeys = "proxy"
	envNameField   ctxKeys = "environment"
	revNameField   ctxKeys = "revision"
	specPathField  ctxKeys = "specPath"
)

type proxyClient interface {
	GetAllProxies() (apigee.Proxies, error)
	GetRevision(proxyName, revision string) (*models.ApiProxyRevision, error)
	GetRevisionResourceFile(proxyName, revision, resourceType, resourceName string) ([]byte, error)
	GetDeployments(apiname string) (*models.DeploymentDetails, error)
	GetVirtualHost(envName, virtualHostName string) (*models.VirtualHost, error)
	GetSpecFile(specPath string) ([]byte, error)
	GetSpecFromURL(url string, options ...apigee.RequestOption) ([]byte, error)
	IsReady() bool
}

type proxyCache interface {
	GetSpecWithPath(path string) (string, error)
	GetSpecPathWithEndpoint(endpoint string) (string, error)
	AddPublishedProxyToCache(cacheKey string, serviceBody *apic.ServiceBody)
}

// job that will poll for any new portals on APIGEE Edge
type pollProxiesJob struct {
	jobs.Job
	client     proxyClient
	cache      proxyCache
	firstRun   bool
	logger     log.FieldLogger
	specsReady JobFirstRunDone
	pubLock    sync.Mutex
}

func newPollProxiesJob(client proxyClient, cache proxyCache, specsReady JobFirstRunDone) *pollProxiesJob {
	job := &pollProxiesJob{
		client:     client,
		cache:      cache,
		firstRun:   true,
		specsReady: specsReady,
		logger:     log.NewFieldLogger().WithComponent("pollProxies").WithPackage("apigee"),
	}
	return job
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

func (j *pollProxiesJob) Execute() error {
	j.logger.Trace("executing")
	allProxies, err := j.client.GetAllProxies()
	if err != nil {
		j.logger.WithError(err).Error("getting proxies")
		return err
	}

	wg := sync.WaitGroup{}
	for _, proxyName := range allProxies {
		if proxyName != "Petstore-X" {
			continue
		}
		wg.Add(1)
		go func(name string) {
			defer wg.Done()
			j.handleProxy(name)
		}(proxyName)
	}

	wg.Wait()

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

	ctx = context.WithValue(ctx, revNameField, revName)

	revision, err := j.client.GetRevision(getStringFromContext(ctx, proxyNameField), revName)
	if err != nil {
		logger.WithError(err).Error("getting revision")
		return
	}
	ctx = context.WithValue(ctx, revNameField, revision)
	logger = logger.WithField(revNameField.String(), revision.Revision)
	addLoggerToContext(ctx, logger)

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
		logger.Info("will download spec from URL in revision")
	}

	j.publish(ctx)
}

func (j *pollProxiesJob) specFromRevision(ctx context.Context) string {
	logger := getLoggerFromContext(ctx)
	logger.Trace("checking revision resource files")

	revision := ctx.Value(revNameField).(*models.ApiProxyRevision)
	for _, resource := range revision.ResourceFiles.ResourceFile {
		if resource.Type == openapi && resource.Name == association {
			return j.getSpecFromResourceFile(ctx, resource.Type, resource.Name)
		}
	}

	return j.getSpecFromVirtualHosts(ctx)
}

func (j *pollProxiesJob) getSpecFromVirtualHosts(ctx context.Context) string {
	logger := getLoggerFromContext(ctx)
	revision := ctx.Value(revNameField).(*models.ApiProxyRevision)
	envName := getStringFromContext(ctx, envNameField)

	// attempt to get the spec from the endpoints the revision is hosted on
	for _, virtualHostName := range revision.Proxies {
		logger := logger.WithField("virtualHostName", virtualHostName)
		virtualHost, err := j.client.GetVirtualHost(envName, virtualHostName)
		if err != nil {
			logger.WithError(err).Debug("could not get virtual host details")
			continue
		}
		urls := urlsFromVirtualHost(virtualHost)

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
	}
	return ""
}

func (j *pollProxiesJob) getSpecFromResourceFile(ctx context.Context, resourceType, resourceName string) string {
	logger := getLoggerFromContext(ctx)
	revision := ctx.Value(revNameField).(*models.ApiProxyRevision)
	logger.Info("found openapi resource file on revision")

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

	// get the association.json file content
	_, err = j.cache.GetSpecWithPath(associationFile.URL)
	if err != nil {
		logger.WithError(err).Error("spec path not found in cache")
		return ""
	}
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
	serviceBodyHash, _ := coreutil.ComputeHash(*serviceBody)
	hashString := util.ConvertUnitToString(serviceBodyHash)
	cacheKey := createProxyCacheKey(getStringFromContext(ctx, proxyNameField), envName)

	// Check DiscoveryCache for API
	j.pubLock.Lock() // only publish one at a time
	defer j.pubLock.Unlock()
	value := agent.GetAttributeOnPublishedAPIByID(revision.Name, fmt.Sprintf("%s-hash", envName))
	update := false
	if !agent.IsAPIPublishedByID(revision.Name) {
		// call new API
		update = true
		j.publishAPI(*serviceBody, envName, hashString, cacheKey)
	} else if value != hashString {
		// handle update
		log.Tracef("%s has been updated, push new revision", revision.Name)
		serviceBody.APIUpdateSeverity = "Major"
		serviceBody.SpecDefinition = []byte{}
		log.Tracef("%+v", serviceBody)
		update = true
		j.publishAPI(*serviceBody, envName, hashString, cacheKey)
	}

	if update {
		j.cache.AddPublishedProxyToCache(cacheKey, serviceBody)
	}
}

func (j *pollProxiesJob) buildServiceBody(ctx context.Context) (*apic.ServiceBody, error) {
	logger := getLoggerFromContext(ctx)
	revision := ctx.Value(revNameField).(*models.ApiProxyRevision)
	specPath := getStringFromContext(ctx, specPathField)
	// get the spec to build the service body
	spec := []byte{}
	if isFullURL(specPath) {
		spec, _ = j.client.GetSpecFromURL(specPath)
	} else if specPath != "" {
		spec, _ = j.client.GetSpecFile(specPath)
	}

	if len(spec) == 0 {
		logger.Info("creating without a spec")
	}

	logger.Info("creating service body")
	sb, err := apic.NewServiceBodyBuilder().
		SetID(revision.Name).
		SetAPIName(revision.Name).
		SetStage(getStringFromContext(ctx, envNameField)).
		SetDescription(revision.Description).
		SetAPISpec(spec).
		SetTitle(revision.DisplayName).
		SetVersion(revision.Revision).
		Build()
	return &sb, err
}

func (j *pollProxiesJob) publishAPI(serviceBody apic.ServiceBody, envName, hashString, cacheKey string) {
	// Add a few more attributes to the service body
	serviceBody.ServiceAttributes["GatewayType"] = gatewayType
	serviceBody.ServiceAgentDetails[fmt.Sprintf("%s-hash", envName)] = hashString
	serviceBody.InstanceAgentDetails["cacheKey"] = cacheKey

	err := agent.PublishAPI(serviceBody)
	if err == nil {
		log.Infof("Published API %s to AMPLIFY Central", serviceBody.NameToPush)
	}
}