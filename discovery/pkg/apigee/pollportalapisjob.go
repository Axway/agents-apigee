package apigee

import (
	"fmt"
	"strconv"

	"github.com/Axway/agent-sdk/pkg/cache"
	"github.com/Axway/agent-sdk/pkg/jobs"
	"github.com/Axway/agent-sdk/pkg/util/log"
	"github.com/Axway/agents-apigee/client/pkg/apigee"
)

// job that will poll for any new apis in the portal on APIGEE Edge
type pollPortalAPIsJob struct {
	jobs.Job
	apigeeClient      *apigee.ApigeeClient
	portalID          string
	portalName        string
	portalAPIsMap     map[string]string
	jobID             string
	consecutiveErrors int // job only fails after 3 consecutive execution errors
	firstRun          bool
}

func newPollPortalAPIsJob(apigeeClient *apigee.ApigeeClient, portalID, portalName string) *pollPortalAPIsJob {
	job := &pollPortalAPIsJob{
		apigeeClient:  apigeeClient,
		portalID:      portalID,
		portalName:    portalName,
		portalAPIsMap: make(map[string]string),
		firstRun:      true,
	}
	publishToTopic(apiValidatorWait, wgAdd)
	return job
}

func (j *pollPortalAPIsJob) Register() error {
	jobID, err := jobs.RegisterIntervalJobWithName(j, j.apigeeClient.GetConfig().GetIntervals().API, fmt.Sprintf("%s Portal Poller", j.portalName))
	if err != nil {
		return err
	}
	j.jobID = jobID
	return nil
}

func (j *pollPortalAPIsJob) Ready() bool {
	return j.apigeeClient.IsReady()
}

func (j *pollPortalAPIsJob) Status() error {
	if j.consecutiveErrors >= 3 {
		return fmt.Errorf("job failed to execute 3 or more consecutive times")
	}
	return nil
}

func (j *pollPortalAPIsJob) Execute() error {
	if j.firstRun {
		defer func() {
			publishToTopic(apiValidatorWait, wgDone)
			j.firstRun = false
		}()
	}
	log.Tracef("Executing %s Portal poller", j.portalName)
	allPortalAPIs, err := j.apigeeClient.GetPortalAPIs(j.portalID)
	if err != nil {
		log.Errorf("error getting APIs for portal %s: %s", j.portalID, err)
		j.consecutiveErrors++
		return nil
	}
	j.consecutiveErrors = 0

	log.Tracef("%s Portal APIs: %+v", j.portalName, allPortalAPIs)
	apisFound := make(map[string]string)
	for _, api := range allPortalAPIs {
		id := strconv.Itoa(api.ID)
		apisFound[id] = api.ProductName
		if _, ok := j.portalAPIsMap[id]; !ok {
			log.Debugf("Found new api product %s", api.ProductName)
			j.portalAPIsMap[id] = api.ProductName
		}
		changed, err := cache.GetCache().HasItemChanged(id, *api)
		if err != nil || changed {
			api.SetPortalTitle(j.portalName)
			// send to new api handler
			publishToTopic(processAPI, api)
		}
	}

	// check if any api has been removed
	for id := range j.portalAPIsMap {
		if _, ok := apisFound[id]; !ok {
			publishToTopic(removedAPI, id)
			defer func(id string) {
				delete(j.portalAPIsMap, id)
			}(id) // remove apis from the map
		}
	}
	log.Tracef("Finished %s Portal poller", j.portalName)
	return nil
}

func (j *pollPortalAPIsJob) PortalRemoved() {
	// Loop all apis in this portal and remove them
	for _, id := range j.portalAPIsMap {
		publishToTopic(removedAPI, id)
	}

	// Unregister this portals job
	jobs.UnregisterJob(j.jobID)
}
