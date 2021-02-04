package apigee

import (
	"encoding/json"
	"fmt"
	"time"

	coreapi "github.com/Axway/agent-sdk/pkg/api"
	corecfg "github.com/Axway/agent-sdk/pkg/config"
	"github.com/Axway/agent-sdk/pkg/util/log"
	"github.com/Axway/agents-apigee/traceability/pkg/config"
)

// LogglyClient - Represents the loggly client
type LogglyClient struct {
	cfg          *config.LogglyConfig
	apiClient    coreapi.Client
	accessToken  string
	startTime    int64
	eventChannel chan []byte
	stopChannel  chan bool
}

// NewLogglyClient - Creates a new loggly Client
func NewLogglyClient(logglyCfg *config.LogglyConfig, eventChannel chan []byte) (*LogglyClient, error) {
	gatewayClient := &LogglyClient{
		apiClient:    coreapi.NewClient(corecfg.NewTLSConfig(), ""),
		cfg:          logglyCfg,
		eventChannel: eventChannel,
		startTime:    time.Now().Add(-60 * time.Minute).Unix(),
		stopChannel:  make(chan bool),
	}

	return gatewayClient, nil
}

// Start - starts the client processing
func (a *LogglyClient) Start() {
	go func() {
		for {
			a.readEvents()
			time.Sleep(30 * time.Second)
		}
	}()
}

// Stop - Stop processing subscriptions
func (a *LogglyClient) Stop() {
	a.stopChannel <- true
}

func (a *LogglyClient) readEvents() {
	nowTime := time.Now()
	startTimeUTC := time.Unix(a.startTime, 0)

	// yyyy-MM-ddTHH:mm:ss.SSSZ
	formattedStartTime := startTimeUTC.UTC().Format("2006-01-02T15:04:05.000Z")
	formattedNowTime := nowTime.UTC().Format("2006-01-02T15:04:05.000Z")
	a.startTime = nowTime.Unix()

	log.Debugf("Getting event between " + formattedStartTime + " and " + formattedNowTime)
	url := "https://" + a.cfg.Organization + ".loggly.com/apiv2/events/iterate?q=tag:apic-logs&from=" + formattedStartTime + "&until=" + formattedNowTime
	a.readNextEvents(url)
	// Todo : persist the startTime to ./data/somefile and read it on agent start
}

// Call API to get events between startTime and now
func (a *LogglyClient) readNextEvents(url string) {
	request := coreapi.Request{
		Method: coreapi.GET,
		URL:    url,
		Headers: map[string]string{
			"Authorization": "Bearer " + a.cfg.APIToken,
		},
	}

	// return the api response
	response, err := a.apiClient.Send(request)
	if err != nil {
		fmt.Println("Error in getting events : " + err.Error())
	}

	var eventCollection LogglyEventsCollection
	json.Unmarshal(response.Body, &eventCollection)
	for _, event := range eventCollection.Events {
		logEntry := []byte(event.Raw)
		a.eventChannel <- logEntry
	}
	if eventCollection.Next != "" {
		a.readNextEvents(eventCollection.Next)
	}
}
