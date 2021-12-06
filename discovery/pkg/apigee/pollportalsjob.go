package apigee

import (
	"github.com/Axway/agent-sdk/pkg/jobs"
	"github.com/Axway/agent-sdk/pkg/util/log"
)

// job that will poll for any new portals on APIGEE Edge
type pollPortalsJob struct {
	jobs.Job
	apigeeClient *GatewayClient
	portalsMap   map[string]string
	portalChan   chan string
}

func newPollPortalsJob(apigeeClient *GatewayClient, portalChan chan string) *pollPortalsJob {
	return &pollPortalsJob{
		apigeeClient: apigeeClient,
		portalsMap:   make(map[string]string),
		portalChan:   portalChan,
	}
}

func (j *pollPortalsJob) Ready() bool {
	if j.apigeeClient.accessToken == "" {
		return false
	}
	return true
}

func (j *pollPortalsJob) Status() error {
	return nil
}

func (j *pollPortalsJob) Execute() error {
	allPortals := j.apigeeClient.getPortals()
	for _, portal := range allPortals {
		if _, ok := j.portalsMap[portal.ID]; !ok {
			log.Debugf("Found new portal %s", portal.Name)
			// send to portal handler
			j.portalChan <- portal.ID
			j.portalsMap[portal.ID] = portal.Name
		}
	}
	return nil
}
