package apigee

// import (
// 	"github.com/Axway/agent-sdk/pkg/cache"
// 	"github.com/Axway/agent-sdk/pkg/jobs"
// 	"github.com/Axway/agent-sdk/pkg/util/log"

// 	"github.com/Axway/agents-apigee/client/pkg/apigee"
// )

// const (
// 	portalMapCacheKey = "PortalMap"
// )

// // job that will poll for any new portals on APIGEE Edge
// type pollPortalsJob struct {
// 	jobs.Job
// 	apigeeClient *apigee.ApigeeClient
// 	portalsMap   map[string]apigee.PortalData
// 	firstRun     bool
// }

// func newPollPortalsJob(apigeeClient *apigee.ApigeeClient) *pollPortalsJob {
// 	job := &pollPortalsJob{
// 		apigeeClient: apigeeClient,
// 		portalsMap:   make(map[string]apigee.PortalData),
// 		firstRun:     true,
// 	}
// 	log.Tracef("+++++++++Increment+++++++++")
// 	publishToTopic(apiValidatorWait, wgAdd)
// 	return job
// }

// func (j *pollPortalsJob) Ready() bool {
// 	return j.apigeeClient.IsReady()
// }

// func (j *pollPortalsJob) Status() error {
// 	return nil
// }

// func (j *pollPortalsJob) Execute() error {
// 	if j.firstRun {
// 		defer func() {
// 			publishToTopic(apiValidatorWait, wgDone)
// 			j.firstRun = false
// 		}()
// 	}
// 	allPortals := j.apigeeClient.GetPortals()
// 	portalsFound := make(map[string]string)
// 	for _, portal := range allPortals {
// 		portalsFound[portal.ID] = portal.Name
// 		if _, ok := j.portalsMap[portal.ID]; !ok {
// 			log.Debugf("Found new portal %s", portal.Name)
// 			j.portalsMap[portal.ID] = portal
// 			cache.GetCache().Set(portalMapCacheKey, j.portalsMap)
// 			// send to portal handler
// 			publishToTopic(newPortal, portal.ID)
// 		}
// 	}

// 	// check if any portal has been removed
// 	for id := range j.portalsMap {
// 		if _, ok := portalsFound[id]; !ok {
// 			publishToTopic(removedPortal, id)
// 			defer func(id string) {
// 				delete(j.portalsMap, id)
// 				cache.GetCache().Set(portalMapCacheKey, j.portalsMap)
// 			}(id) // remove apis from the map
// 		}
// 	}
// 	return nil
// }
