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
	done           chan struct{}
	logglyClient   *apigee.LogglyClient
	eventProcessor *apigee.EventProcessor
	client         beat.Client
	eventChannel   chan []byte
	statChannel    chan interface{}
}

var bt *customLogBeater

// var logglyConfig *config.LogglyConfig

// New creates an instance of aws_apigw_traceability_agent.
func New(b *beat.Beat, cfg *common.Config) (beat.Beater, error) {
	bt := &customLogBeater{
		done:         make(chan struct{}),
		eventChannel: make(chan []byte),
	}

	var err error

	// bt.logglyClient, err = apigee.NewLogglyClient(logglyConfig, bt.eventChannel)
	// bt.eventProcessor = apigee.NewEventProcessor(logglyConfig)
	bt.eventProcessor = apigee.NewEventProcessor()
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
// func SetLogglyConfig(logglyCfg *config.LogglyConfig) {
// 	logglyConfig = logglyCfg
// }

// SetStat - set parsed gateway config
// func SetLogglyConfig(logglyCfg *config.LogglyConfig) {
// 	logglyConfig = logglyCfg
// }

// Run starts ApigeeTraceabilityAgent.
func (bt *customLogBeater) Run(b *beat.Beat) error {
	log.Info("apigee_traceability_agent is running! Hit CTRL-C to stop it.")

	var err error
	bt.client, err = b.Publisher.Connect()
	if err != nil {
		return err
	}

	// bt.logglyClient.Start()

	for {
		select {
		case <-bt.done:
			return nil
		case statData := <-bt.statChannel:
			log.Debugf("STAT TO PROCESS: %+v", statData)
		case eventData := <-bt.eventChannel:
			log.Debug("EVENT TO PROCESS : " + string(eventData))
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
