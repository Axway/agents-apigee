package beater

import (
	agenterrors "github.com/Axway/agent-sdk/pkg/util/errors"
	hc "github.com/Axway/agent-sdk/pkg/util/healthcheck"
	"github.com/Axway/agent-sdk/pkg/util/log"

	"github.com/elastic/beats/v7/libbeat/beat"
	"github.com/elastic/beats/v7/libbeat/common"

	"github.com/Axway/agents-apigee/traceability/pkg/apigee"
)

// customLogBeater configuration.
type customLogBeater struct {
	done   chan struct{}
	client beat.Client
}

var bt *customLogBeater

// New creates an instance of aws_apigw_traceability_agent.
func New(b *beat.Beat, cfg *common.Config) (beat.Beater, error) {
	bt := &customLogBeater{
		done: make(chan struct{}),
	}

	var err error

	if err != nil {
		return nil, err
	}

	// Validate that all necessary services are up and running. If not, return error
	if hc.RunChecks() != hc.OK {
		return nil, agenterrors.ErrInitServicesNotReady
	}

	return bt, nil
}

// Run starts ApigeeTraceabilityAgent.
func (bt *customLogBeater) Run(b *beat.Beat) error {
	log.Info("apigee_traceability_agent is running! Hit CTRL-C to stop it.")

	var err error
	bt.client, err = b.Publisher.Connect()
	if err != nil {
		return err
	}

	apigee.GetAgent().BeatsReady()
	for {
		select {
		case <-bt.done:
			return nil
		}
	}
}

// Stop stops customLogTraceabilityAgent.
func (bt *customLogBeater) Stop() {
	bt.client.Close()
	close(bt.done)
}
