package apigee

import (
	"encoding/json"
	"fmt"
	"net/url"
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
	gatewayType    = "Apigee"
	proxyNameField = "proxy"
	envNameField   = "environment"
	revNameField   = "revision"
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
	GetSpecWithPath(path string) (uint64, error)
	GetSpecPathWithEndpoint(endpoint string) (string, error)
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
	logger := j.logger.WithField(proxyNameField, proxyName)
	logger.Debug("handling proxy")

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
			j.handleEnvironment(proxyName, environment)
		}(env)
	}

	wg.Wait()
}

func (j *pollProxiesJob) handleEnvironment(proxyName string, env models.DeploymentDetailsEnvironment) {
	logger := j.logger.WithField(proxyNameField, proxyName).WithField(envNameField, env.Name)
	logger.Debug("handling environment")

	wg := sync.WaitGroup{}
	for _, rev := range env.Revision {
		wg.Add(1)
		go func(revision models.DeploymentDetailsRevision) {
			defer wg.Done()
			j.handleRevision(proxyName, env.Name, revision)
		}(rev)
	}

	wg.Wait()
}

func (j *pollProxiesJob) handleRevision(proxyName, envName string, rev models.DeploymentDetailsRevision) {
	logger := j.logger.WithField(proxyNameField, proxyName).WithField(envNameField, envName).WithField(revNameField, rev.Name)
	logger.Debug("handling revision")

	revision, err := j.client.GetRevision(proxyName, rev.Name)

	if err != nil {
		logger.WithError(err).Error("getting revision")
		return
	}

	var specURL string
	if revision.Spec != nil && revision.Spec != "" {
		logger.WithField("specURL", revision.Spec).Info("will download spec from URL in revision")
		specURL = revision.Spec.(string)
	} else {
		specURL = j.specFromRevision(proxyName, envName, revision)
	}

	j.publish(proxyName, envName, revision, specURL)
}

func (j *pollProxiesJob) specFromRevision(proxyName, envName string, revision *models.ApiProxyRevision) string {
	logger := j.logger.WithField(proxyNameField, proxyName).WithField(envNameField, envName).WithField(revNameField, revision.Revision)
	logger.Trace("checking revision resource files")
	for _, resource := range revision.ResourceFiles.ResourceFile {
		if resource.Type == openapi && resource.Name == association {
			return j.getSpecFromResourceFile(proxyName, envName, revision, resource.Type, resource.Name)
		}
	}

	return j.getSpecFromVirtualHosts(proxyName, envName, revision)
}

func (j *pollProxiesJob) getSpecFromVirtualHosts(proxyName, envName string, revision *models.ApiProxyRevision) string {
	logger := j.logger.WithField(proxyNameField, proxyName).WithField(envNameField, envName).WithField(revNameField, revision.Revision)
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

func (j *pollProxiesJob) getSpecFromResourceFile(proxyName, envName string, revision *models.ApiProxyRevision, resourceName, resourceType string) string {
	logger := j.logger.WithField(proxyNameField, proxyName).WithField(envNameField, envName).WithField(revNameField, revision.Revision).WithField("filename", resourceName)
	logger.Info("found openapi resource file on revision")

	// get the association.json file content
	resFileContent, err := j.client.GetRevisionResourceFile(proxyName, revision.Revision, resourceType, resourceName)
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
	}
	return associationFile.URL
}

func urlsFromVirtualHost(virtualHost *models.VirtualHost) []string {
	urls := []string{}

	scheme := "http"
	port := virtualHost.Port
	if virtualHost.SSLInfo != nil {
		scheme = "https"
		if port == "443" {
			port = ""
		}
	}
	if scheme == "http" && port == "80" {
		port = ""
	}

	for _, host := range virtualHost.HostAliases {
		thisURL := fmt.Sprintf("%s://%s:%s", scheme, host, port)
		if port == "" {
			thisURL = fmt.Sprintf("%s://%s", scheme, host)
		}
		urls = append(urls, thisURL)
	}

	return urls
}

func (j *pollProxiesJob) publish(proxyName, envName string, revision *models.ApiProxyRevision, specPath string) {
	logger := j.logger.WithField(proxyNameField, proxyName).WithField(envNameField, envName)

	//  start the service Body
	serviceBody, err := j.buildServiceBody(proxyName, envName, revision, specPath)
	if err != nil {
		logger.WithError(err).Error("building service body")
		return
	}
	serviceBodyHash, _ := coreutil.ComputeHash(*serviceBody)

	hashString := util.ConvertUnitToString(serviceBodyHash)

	// Check DiscoveryCache for API
	j.pubLock.Lock() // only publish one at a time
	defer j.pubLock.Unlock()
	if !agent.IsAPIPublishedByID(revision.Name) {
		// call new API
		j.publishAPI(*serviceBody, envName, hashString)
	} else if value := agent.GetAttributeOnPublishedAPIByID(revision.Name, fmt.Sprintf("%s-hash", envName)); value != hashString {
		// handle update
		log.Tracef("%s has been updated, push new revision", revision.Name)
		serviceBody.APIUpdateSeverity = "Major"
		serviceBody.SpecDefinition = []byte{}
		log.Tracef("%+v", serviceBody)
		j.publishAPI(*serviceBody, envName, hashString)
	}
}

func (j *pollProxiesJob) buildServiceBody(proxyName, envName string, revision *models.ApiProxyRevision, specPath string) (*apic.ServiceBody, error) {
	// get the spec to build the service body
	logger := j.logger.WithField(proxyNameField, proxyName).WithField(envNameField, envName)
	spec := []byte{}
	if specPath != "" && isFullURL(specPath) {
		spec, _ = j.client.GetSpecFromURL(specPath)
	} else if specPath != "" {
		spec, _ = j.client.GetSpecFile(specPath)
	}

	if len(spec) == 0 {
		logger.Info("creating without a spec")
	} else {
		logger.WithField("spec", specPath)
	}
	logger.Info("creating service body")

	sb, err := apic.NewServiceBodyBuilder().
		SetID(revision.Name).
		SetAPIName(revision.Name).
		SetStage(envName).
		SetDescription(revision.Description).
		SetAPISpec(spec).
		SetTitle(revision.DisplayName).
		Build()
	return &sb, err
}

func (j *pollProxiesJob) publishAPI(serviceBody apic.ServiceBody, envName, hashString string) {
	// Add a few more attributes to the service body
	serviceBody.ServiceAttributes["GatewayType"] = gatewayType
	serviceBody.ServiceAttributes[fmt.Sprintf("%s-hash", envName)] = hashString

	err := agent.PublishAPI(serviceBody)
	if err == nil {
		log.Infof("Published API %s to AMPLIFY Central", serviceBody.NameToPush)
	}
}

// isFullURL - returns true if the url arg is a fully qualified URL
func isFullURL(urlString string) bool {
	if _, err := url.ParseRequestURI(urlString); err != nil {
		return true
	}
	return false
}
