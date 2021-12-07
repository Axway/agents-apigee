package apigee

import (
	"strconv"

	"github.com/Axway/agent-sdk/pkg/jobs"
	"github.com/Axway/agent-sdk/pkg/util/log"
)

// job that will poll for any new apis in the portal on APIGEE Edge
type pollPortalAPIsJob struct {
	jobs.Job
	apigeeClient  *GatewayClient
	portalID      string
	portalName    string
	portalAPIsMap map[string]string
	apiChan       chan *apiDocData
}

func newPollPortalAPIsJob(apigeeClient *GatewayClient, portalID, portalName string, apiChan chan *apiDocData) *pollPortalAPIsJob {
	return &pollPortalAPIsJob{
		apigeeClient:  apigeeClient,
		portalID:      portalID,
		portalName:    portalName,
		portalAPIsMap: make(map[string]string),
		apiChan:       apiChan,
	}
}

func (j *pollPortalAPIsJob) Ready() bool {
	if j.apigeeClient.accessToken == "" {
		return false
	}
	return true
}

func (j *pollPortalAPIsJob) Status() error {
	return nil
}

func (j *pollPortalAPIsJob) Execute() error {
	log.Tracef("Executing %s Portal poller", j.portalName)
	allPortalAPIs := j.apigeeClient.getPortalAPIs(j.portalID)
	log.Tracef("%s Portal APIs: %+v", j.portalName, allPortalAPIs)
	// TODO cache portal apis list
	for _, api := range allPortalAPIs {
		id := strconv.Itoa(api.ID)
		if _, ok := j.portalAPIsMap[id]; !ok {
			log.Debugf("Found new api product %s", api.ProductName)
			// send to new api handler
			api.SetPortalTitle(j.portalName)
			j.apiChan <- api
			j.portalAPIsMap[id] = api.ProductName
		}
	}
	log.Tracef("Finished %s Portal poller", j.portalName)
	return nil
}
