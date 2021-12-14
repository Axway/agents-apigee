package apigee

import (
	"github.com/Axway/agent-sdk/pkg/jobs"
	"github.com/Axway/agents-apigee/client/pkg/apigee"
	"github.com/Axway/agents-apigee/client/pkg/apigee/models"
)

type createDeveloperJob struct {
	jobs.Job
	apigeeClient   *apigee.ApigeeClient
	setDeveloperID func(string)
}

func newCreateDeveloperJob(apigeeClient *apigee.ApigeeClient, setDeveloperID func(string)) *createDeveloperJob {
	return &createDeveloperJob{
		apigeeClient:   apigeeClient,
		setDeveloperID: setDeveloperID,
	}
}

func (j *createDeveloperJob) Ready() bool {
	return j.apigeeClient.IsReady()
}

func (j *createDeveloperJob) Status() error {
	return nil
}

func (j *createDeveloperJob) Execute() error {
	// Creates the developer that will be associated with all agent created apps

	// apigee agent developer TODO - make configurable
	agentDev := models.Developer{
		Email:     "apigee-agent@axway.com",
		FirstName: "Apigee",
		LastName:  "Agent",
		UserName:  "apigee-agent",
		Attributes: []models.Attribute{
			apigee.ApigeeAgentAttribute,
		},
	}

	// get the developers first
	devs := j.apigeeClient.GetDevelopers()
	for _, dev := range devs {
		if dev == agentDev.Email {
			existingDev, err := j.apigeeClient.GetDeveloper(agentDev.Email)
			if err != nil {
				return err
			}
			j.setDeveloperID(existingDev.DeveloperId)
			return nil // found the apigee agent developer
		}
	}

	// create the apigee agent developer
	newDev, err := j.apigeeClient.CreateDeveloper(agentDev)
	if err == nil {
		j.setDeveloperID(newDev.DeveloperId)
	}
	return err
}
