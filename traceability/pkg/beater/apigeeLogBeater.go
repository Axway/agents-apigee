package beater

import (
	"github.com/Axway/agent-sdk/pkg/traceability"
	agenterrors "github.com/Axway/agent-sdk/pkg/util/errors"
	hc "github.com/Axway/agent-sdk/pkg/util/healthcheck"

	"github.com/elastic/beats/v7/libbeat/beat"
	"github.com/elastic/beats/v7/libbeat/common"
	"github.com/elastic/beats/v7/libbeat/logp"

	"github.com/Axway/agents-apigee/traceability/pkg/config"
	"github.com/Axway/agents-apigee/traceability/pkg/gateway"
)

// customLogBeater configuration.
type customLogBeater struct {
	done           chan struct{}
	logReader      *gateway.LogReader
	eventProcessor *gateway.EventProcessor
	client         beat.Client
	eventChannel   chan string
}

var bt *customLogBeater
var gatewayConfig *config.GatewayConfig

// New creates an instance of aws_apigw_traceability_agent.
func New(b *beat.Beat, cfg *common.Config) (beat.Beater, error) {
	bt := &customLogBeater{
		done:         make(chan struct{}),
		eventChannel: make(chan string),
	}

	var err error
	bt.logReader, err = gateway.NewLogReader(gatewayConfig, bt.eventChannel)
	bt.eventProcessor = gateway.NewEventProcessor(gatewayConfig)
	if err != nil {
		return nil, err
	}

	if !gatewayConfig.ProcessOnInput {
		traceability.SetOutputEventProcessor(bt.eventProcessor)
	}

	// Validate that all necessary services are up and running. If not, return error
	if hc.RunChecks() != hc.OK {
		return nil, agenterrors.ErrInitServicesNotReady
	}

	return bt, nil
}

// SetGatewayConfig - set parsed gateway config
func SetGatewayConfig(gatewayCfg *config.GatewayConfig) {
	gatewayConfig = gatewayCfg
}

// Run starts awsApigwTraceabilityAgent.
func (bt *customLogBeater) Run(b *beat.Beat) error {
	logp.Info("apic_traceability_agent is running! Hit CTRL-C to stop it.")

	var err error
	bt.client, err = b.Publisher.Connect()
	if err != nil {
		return err
	}

	bt.logReader.Start()

	for {
		select {
		case <-bt.done:
			return nil
		case eventData := <-bt.eventChannel:
			if gatewayConfig.ProcessOnInput {
				eventsToPublish := bt.eventProcessor.ProcessRaw([]byte(eventData))
				if eventsToPublish != nil {
					bt.client.PublishAll(eventsToPublish)
				}
			} else {
				eventToPublish := beat.Event{
					Fields: common.MapStr{
						"message": eventData,
					},
				}
				bt.client.Publish(eventToPublish)
			}
		}
	}
}

// Stop stops customLogTraceabilityAgent.
func (bt *customLogBeater) Stop() {
	bt.client.Close()
	close(bt.done)
}
