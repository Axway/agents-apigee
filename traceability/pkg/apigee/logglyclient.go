package apigee

import (
	"encoding/json"
	"fmt"
	"time"

	coreapi "github.com/Axway/agent-sdk/pkg/api"
	corecfg "github.com/Axway/agent-sdk/pkg/config"
	"github.com/Axway/agent-sdk/pkg/util/log"
	"github.com/Axway/agents-apigee/discovery/pkg/config"
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
		startTime:    time.Now().Add(-1 * time.Minute).Unix(),
		stopChannel:  make(chan bool),
	}

	return gatewayClient, nil
}

// Start - starts the client processing
func (a *LogglyClient) Start() {
	go func() {
		for {
			nowTime := time.Now()
			rsID, err := a.createSearch()
			if err != nil {
				log.Error(err.Error())
			}
			a.readEvents(rsID)

			// Todo : persist the startTime to ./data/somefile and read it on agent start
			a.startTime = nowTime.Unix()

			fmt.Println("Sleeping for 30 seconds")
			time.Sleep(30 * time.Second)
		}
	}()
}

// Stop - Stop processing subscriptions
func (a *LogglyClient) Stop() {
	a.stopChannel <- true
}

func (a *LogglyClient) getAPIBaseURL() string {
	return "https://" + a.cfg.Organization + ".loggly.com/apiv2"
}

func (a *LogglyClient) getSearchURL() string {
	return a.getAPIBaseURL() + "/search"
}

func (a *LogglyClient) getEventURL() string {
	return a.getAPIBaseURL() + "/events"
}

func (a *LogglyClient) createSearch() (string, error) {
	startTimeUTC := time.Unix(a.startTime, 0)
	// yyyy-MM-ddTHH:mm:ss.SSSZ
	formattedStartTime := startTimeUTC.UTC().Format("2006-01-02T15:04:05.000Z")
	log.Debugf("Creating search for event after " + formattedStartTime)

	request := coreapi.Request{
		Method: coreapi.GET,
		URL:    a.getSearchURL(),
		Headers: map[string]string{
			"Authorization": "Bearer " + a.cfg.APIToken,
		},
		QueryParams: map[string]string{
			"q":     "tag:apic-logs",
			"from":  formattedStartTime,
			"until": "now",
			"size":  "1000",
			"order": "desc",
		},
	}

	response, err := a.apiClient.Send(request)
	if err != nil {
		fmt.Println("Error in creating search for events : " + err.Error())
		return "", err
	}

	var searchResponse LogglySearchResponse
	json.Unmarshal(response.Body, &searchResponse)

	return searchResponse.RSID.ID, nil
}

func (a *LogglyClient) readEvents(rsID string) error {
	request := coreapi.Request{
		Method: coreapi.GET,
		URL:    a.getEventURL(),
		Headers: map[string]string{
			"Authorization": "Bearer " + a.cfg.APIToken,
		},
		QueryParams: map[string]string{
			"rsid": rsID,
		},
	}
	response, err := a.apiClient.Send(request)
	if err != nil {
		fmt.Println("Error in getting events : " + err.Error())
		return err
	}

	var eventCollection LogglyEventsCollection
	json.Unmarshal(response.Body, &eventCollection)
	for _, event := range eventCollection.Events {
		logEntry := []byte(event.Logmsg)
		a.eventChannel <- logEntry
	}
	return nil
}
