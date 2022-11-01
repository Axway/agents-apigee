package apigee

import (
	"github.com/Axway/agent-sdk/pkg/jobs"
)

type registerAPIValidatorJob struct {
	jobs.Job
	validatorReady    jobFirstRunDone
	registerValidator func()
}

func newRegisterAPIValidatorJob(proxiesReady jobFirstRunDone, registerValidator func()) *registerAPIValidatorJob {
	job := &registerAPIValidatorJob{
		validatorReady:    proxiesReady,
		registerValidator: registerValidator,
	}
	return job
}

func (j *registerAPIValidatorJob) Ready() bool {
	return j.validatorReady()
}

func (j *registerAPIValidatorJob) Status() error {
	return nil
}

func (j *registerAPIValidatorJob) Execute() error {
	j.registerValidator()
	return nil
}
