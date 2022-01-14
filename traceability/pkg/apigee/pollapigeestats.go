package apigee

import (
	"github.com/Axway/agent-sdk/pkg/jobs"
)

type pollApigeeStats struct {
	jobs.Job
}

func (j *pollApigeeStats) Ready() bool {
	return true
}

func (j *pollApigeeStats) Status() error {
	return nil
}

func (j *pollApigeeStats) Execute() error {
	return nil
}
