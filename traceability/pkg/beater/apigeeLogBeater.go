package beater

import (
	"fmt"

	agenterrors "github.com/Axway/agent-sdk/pkg/util/errors"
	hc "github.com/Axway/agent-sdk/pkg/util/healthcheck"

	"github.com/elastic/beats/v7/libbeat/beat"
	"github.com/elastic/beats/v7/libbeat/common"
	"github.com/elastic/beats/v7/libbeat/logp"

	"github.com/Axway/agents-apigee/discovery/pkg/config"
	"github.com/Axway/agents-apigee/traceability/pkg/apigee"
)

// customLogBeater configuration.
type customLogBeater struct {
	done           chan struct{}
	logglyClient   *apigee.LogglyClient
	eventProcessor *apigee.EventProcessor
	client         beat.Client
	eventChannel   chan []byte
}

var bt *customLogBeater
var logglyConfig *config.LogglyConfig

// New creates an instance of aws_apigw_traceability_agent.
func New(b *beat.Beat, cfg *common.Config) (beat.Beater, error) {
	bt := &customLogBeater{
		done:         make(chan struct{}),
		eventChannel: make(chan []byte),
	}

	var err error

	bt.logglyClient, err = apigee.NewLogglyClient(logglyConfig, bt.eventChannel)
	bt.eventProcessor = apigee.NewEventProcessor(logglyConfig)
	if err != nil {
		return nil, err
	}

	// Validate that all necessary services are up and running. If not, return error
	if hc.RunChecks() != hc.OK {
		return nil, agenterrors.ErrInitServicesNotReady
	}

	return bt, nil
}

// SetLogglyConfig - set parsed gateway config
func SetLogglyConfig(logglyCfg *config.LogglyConfig) {
	logglyConfig = logglyCfg
}

// Run starts ApigeeTraceabilityAgent.
func (bt *customLogBeater) Run(b *beat.Beat) error {
	logp.Info("apigee_traceability_agent is running! Hit CTRL-C to stop it.")

	var err error
	bt.client, err = b.Publisher.Connect()
	if err != nil {
		return err
	}

	bt.logglyClient.Start()

	for {
		select {
		case <-bt.done:
			return nil
		case eventData := <-bt.eventChannel:
			fmt.Println("EVENT TO PROCESS : " + string(eventData))

			// Todo : Uncomment
			eventsToPublish := bt.eventProcessor.ProcessRaw(eventData)
			if eventsToPublish != nil {
				bt.client.PublishAll(eventsToPublish)
			}
		}
	}
}

// Stop stops customLogTraceabilityAgent.
func (bt *customLogBeater) Stop() {
	bt.client.Close()
	close(bt.done)
}
