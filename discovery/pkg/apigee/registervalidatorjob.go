package apigee

import (
	"sync"

	"github.com/Axway/agent-sdk/pkg/jobs"
)

type registerAPIValidatorJob struct {
	jobs.Job
	waitGroup         sync.WaitGroup
	wgActionChan      chan interface{}
	topicID           string
	registerValidator func()
}

func newRegisterAPIValidatorJob(registerValidator func()) *registerAPIValidatorJob {
	wgActionChan, id, _ := subscribeToTopic(apiValidatorWait)
	job := &registerAPIValidatorJob{
		waitGroup:         sync.WaitGroup{},
		wgActionChan:      wgActionChan,
		topicID:           id,
		registerValidator: registerValidator,
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
			if action.(wgAction) == wgAdd {
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
	j.registerValidator()
	unsubscribeFromTopic(apiValidatorWait, j.topicID)
	return nil
}
