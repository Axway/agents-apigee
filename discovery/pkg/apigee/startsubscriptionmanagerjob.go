package apigee

import (
	"github.com/Axway/agent-sdk/pkg/agent"
	"github.com/Axway/agent-sdk/pkg/jobs"
	"github.com/Axway/agents-apigee/client/pkg/apigee"
)

type startSubscriptionManagerJob struct {
	jobs.Job
	apigeeClient   *apigee.ApigeeClient
	getDeveloperID func() string
}

func newStartSubscriptionManager(apigeeClient *apigee.ApigeeClient, getDeveloperID func() string) *startSubscriptionManagerJob {
	return &startSubscriptionManagerJob{
		apigeeClient:   apigeeClient,
		getDeveloperID: getDeveloperID,
	}
}

func (j *startSubscriptionManagerJob) Ready() bool {
	return j.apigeeClient.IsReady() && j.getDeveloperID() != ""
}

func (j *startSubscriptionManagerJob) Status() error {
	return nil
}

func (j *startSubscriptionManagerJob) Execute() error {
	agent.GetCentralClient().GetSubscriptionManager().Start()
	return nil
}
