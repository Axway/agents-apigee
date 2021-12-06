package apigee

import (
	"github.com/Axway/agent-sdk/pkg/jobs"
	"github.com/Axway/agent-sdk/pkg/util/log"
)

// job that will poll for any new apis in the portal on APIGEE Edge
type pollPortalAPIsJob struct {
	jobs.Job
	apigeeClient  *GatewayClient
	portalName    string
	portalAPIsMap map[string]string
	apiChan       chan *apiDocData
}

func newPollPortalAPIsJob(apigeeClient *GatewayClient, portalName string, apiChan chan *apiDocData) *pollPortalAPIsJob {
	return &pollPortalAPIsJob{
		apigeeClient:  apigeeClient,
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
	allPortalAPIs := j.apigeeClient.getPortalAPIs(j.portalName)
	log.Tracef("%s Portal APIs: %+v", j.portalName, allPortalAPIs)
	// TODO cache portal apis list
	for _, api := range allPortalAPIs {
		if _, ok := j.portalAPIsMap[api.ID]; !ok {
			log.Debugf("Found new api product %s", api.ProductName)
			// send to new api handler
			j.apiChan <- &api
			j.portalAPIsMap[api.ID] = api.ProductName
		}
	}
	log.Tracef("Finished %s Portal poller", j.portalName)
	return nil
}
