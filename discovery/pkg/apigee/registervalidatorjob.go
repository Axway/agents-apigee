package apigee

import (
	"sync"

	"github.com/Axway/agent-sdk/pkg/agent"
	"github.com/Axway/agent-sdk/pkg/jobs"
)

type registerAPIValidatorJob struct {
	jobs.Job
	waitGroup    sync.WaitGroup
	wgActionChan chan wgAction
	apiValidator agent.APIValidator
}

func newRegisterAPIValidatorJob(wgActionChan chan wgAction, apiValidator agent.APIValidator) *registerAPIValidatorJob {
	job := &registerAPIValidatorJob{
		waitGroup:    sync.WaitGroup{},
		wgActionChan: wgActionChan,
		apiValidator: apiValidator,
	}
	go job.acceptActions()
	return job
}

func (j *registerAPIValidatorJob) acceptActions() {
	for {
		select {
		case action, ok := <-j.wgActionChan:
			if !ok {
				return
			}
			if action == wgAdd {
				j.waitGroup.Add(1)
			} else {
				j.waitGroup.Done()
			}
		}
	}
}

func (j *registerAPIValidatorJob) Ready() bool {
	return true
}

func (j *registerAPIValidatorJob) Status() error {
	return nil
}

func (j *registerAPIValidatorJob) Execute() error {
	j.waitGroup.Wait()
	agent.RegisterAPIValidator(j.apiValidator)
	close(j.wgActionChan)
	return nil
}
