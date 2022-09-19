package apigee

import (
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/Axway/agent-sdk/pkg/apic"
	"github.com/Axway/agent-sdk/pkg/jobs"
	"github.com/Axway/agent-sdk/pkg/util"
	"github.com/Axway/agent-sdk/pkg/util/log"

	"github.com/Axway/agents-apigee/client/pkg/apigee"
)

type specClient interface {
	GetSpecFile(specPath string) ([]byte, error)
	GetAllSpecs() ([]apigee.SpecDetails, error)
	IsReady() bool
}

type specCache interface {
	AddSpecToCache(id, path string, contentHash uint64, modDate time.Time, endpoints ...string)
}

// job that will poll for any new portals on APIGEE Edge
type pollSpecsJob struct {
	jobs.Job
	client   specClient
	cache    specCache
	firstRun bool
	logger   log.FieldLogger
}

func newPollSpecsJob(client specClient, cache specCache) *pollSpecsJob {
	job := &pollSpecsJob{
		client:   client,
		cache:    cache,
		firstRun: true,
		logger:   log.NewFieldLogger().WithComponent("pollSpecs").WithPackage("apigee"),
	}
	return job
}

func (j *pollSpecsJob) Ready() bool {
	j.logger.Trace("checking if the apigee client is ready for calls")
	return j.client.IsReady()
}

func (j *pollSpecsJob) Status() error {
	return nil
}

func (j *pollSpecsJob) Execute() error {
	j.logger.Trace("executing")
	allSpecs, err := j.client.GetAllSpecs()
	if err != nil {
		j.logger.WithError(err).Error("getting specs")
		return err
	}

	wg := sync.WaitGroup{}
	for _, spec := range allSpecs {
		wg.Add(1)
		go func(specDetails apigee.SpecDetails) {
			defer wg.Done()
			j.handleSpec(specDetails)
		}(spec)
	}

	wg.Wait()

	j.firstRun = false
	return nil
}

func (j *pollSpecsJob) FirstRunComplete() bool {
	return !j.firstRun
}

func (j *pollSpecsJob) handleSpec(spec apigee.SpecDetails) {
	logger := j.logger.WithField("specName", spec.Name)
	logger.Trace("handling spec")

	// get the spec content
	content, err := j.client.GetSpecFile(spec.SelfLink)
	if err != nil {
		j.logger.WithError(err).Error("getting spec content")
		return
	}

	// parse the spec
	parser := apic.NewSpecResourceParser(content, "")
	err = parser.Parse()
	if err != nil {
		j.logger.WithError(err).Error("could not parse spec")
		return
	}

	// gather spec info
	endpoints := []string{}
	endpointDefs, err := parser.GetSpecProcessor().GetEndpoints()
	if err != nil {
		j.logger.WithError(err).Error("could not get spec endpoints")
		return
	}
	for _, ep := range endpointDefs {
		endpoints = append(endpoints, endpointToString(ep))
	}

	// add spec details to cache
	hash, _ := util.ComputeHash(content)
	modDate, _ := time.Parse("2006-01-02T15:04:05.000Z", spec.Modified)
	j.cache.AddSpecToCache(spec.ID, spec.SelfLink, hash, modDate, endpoints...)
}

func endpointToString(endpoint apic.EndpointDefinition) string {
	if endpoint.Port > 0 &&
		((strings.ToLower(endpoint.Protocol) == "http" && endpoint.Port != 80) ||
			(strings.ToLower(endpoint.Protocol) == "https" && endpoint.Port != 443)) {
		return fmt.Sprintf("%v://%v:%v%v", endpoint.Protocol, endpoint.Host, endpoint.Port, endpoint.BasePath)
	}
	return fmt.Sprintf("%v://%v%v", endpoint.Protocol, endpoint.Host, endpoint.BasePath)
}
