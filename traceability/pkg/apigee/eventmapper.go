package apigee

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/Axway/agent-sdk/pkg/agent"
	"github.com/Axway/agent-sdk/pkg/transaction"
	"github.com/Axway/agent-sdk/pkg/util/log"
)

// EventMapper -
type EventMapper struct {
}

func (m *EventMapper) processMapping(apigeeLogEntry LogEntry) ([]transaction.LogEvent, error) {
	centralCfg := agent.GetCentralConfig()
	inboundHTTPProtocol, err := transaction.NewHTTPProtocolBuilder().
		SetURI(apigeeLogEntry.Path).
		SetMethod(apigeeLogEntry.Verb).
		SetStatus(stringToInt(apigeeLogEntry.StatusCode), http.StatusText(stringToInt(apigeeLogEntry.StatusCode))).
		SetHost(apigeeLogEntry.RequestHost).
		SetHeaders(apigeeLogEntry.RequestHeaders, apigeeLogEntry.ResponseHeaders).
		SetByteLength(stringToInt(apigeeLogEntry.BytesSent), stringToInt(apigeeLogEntry.BytesReceived)).
		SetRemoteAddress("", apigeeLogEntry.RequestHost, 443).
		SetLocalAddress("", 0).
		Build()
	if err != nil {
		return nil, err
	}
	txEventStatus := transaction.TxEventStatusFail
	if (stringToInt(apigeeLogEntry.StatusCode)) < 400 {
		txEventStatus = transaction.TxEventStatusPass
	}

	transInboundLogEventLeg, err := transaction.NewTransactionEventBuilder().
		SetTimestamp(stringToInt64(apigeeLogEntry.ClientStartTimeStamp)).
		SetDuration(getDuration(apigeeLogEntry)).
		SetTransactionID(apigeeLogEntry.MessageID).
		SetID(apigeeLogEntry.MessageID + "-leg0").
		SetSource(apigeeLogEntry.ClientHost).
		SetDestination(buildURILeg(apigeeLogEntry)).
		SetDirection("Inbound").
		SetStatus(txEventStatus).
		SetProtocolDetail(inboundHTTPProtocol).
		Build()
	if err != nil {
		return nil, err
	}

	outboundHTTPProtocol, err := transaction.NewHTTPProtocolBuilder().
		SetURI(apigeeLogEntry.Path).
		SetMethod(apigeeLogEntry.Verb).
		SetStatus(stringToInt(apigeeLogEntry.StatusCode), http.StatusText(stringToInt(apigeeLogEntry.StatusCode))).
		SetHost(apigeeLogEntry.RequestHost).
		SetHeaders(apigeeLogEntry.RequestHeaders, apigeeLogEntry.ResponseHeaders).
		SetByteLength(stringToInt(apigeeLogEntry.BytesSent), stringToInt(apigeeLogEntry.BytesReceived)).
		SetRemoteAddress("", apigeeLogEntry.RequestHost, 443).
		SetLocalAddress("", 0).
		Build()
	if err != nil {
		return nil, err
	}
	txEventStatus = transaction.TxEventStatusFail
	if (stringToInt(apigeeLogEntry.StatusCode)) < 400 {
		txEventStatus = transaction.TxEventStatusPass
	}

	transOutboundLogEventLeg, err := transaction.NewTransactionEventBuilder().
		SetTimestamp(stringToInt64(apigeeLogEntry.ClientStartTimeStamp)).
		SetDuration(getDuration(apigeeLogEntry)).
		SetTransactionID(apigeeLogEntry.MessageID).
		SetID(apigeeLogEntry.MessageID + "-leg1"). //TODO diff between transactionID and ID
		SetParentID(apigeeLogEntry.MessageID + "-leg0").
		SetSource(buildURILeg(apigeeLogEntry)).
		SetDestination(apigeeLogEntry.RequestHost).
		SetDirection("Outbound").
		SetStatus(txEventStatus).
		SetProtocolDetail(outboundHTTPProtocol).
		Build()

	if err != nil {
		return nil, err
	}

	transSummaryLogEvent, err := m.createSummaryEvent(apigeeLogEntry, centralCfg.GetTeamID())
	if err != nil {
		return nil, err
	}

	return []transaction.LogEvent{
		*transSummaryLogEvent,
		*transInboundLogEventLeg,
		*transOutboundLogEventLeg,
	}, nil
	return nil, nil
}

func (m *EventMapper) getTransactionStatus(code int) string {
	if code >= 400 {
		return "FAIL"
	}
	return "PASS"
}

func (m *EventMapper) buildHeaders(headers map[string]string) string {
	jsonHeader, err := json.Marshal(headers)
	if err != nil {
		log.Error(err.Error())
	}
	return string(jsonHeader)
}

func (m *EventMapper) createSummaryEvent(apigeeLogEntry LogEntry, teamID string) (*transaction.LogEvent, error) {
	transSummaryStatus := transaction.TxSummaryStatusUnknown
	statusCode := stringToInt(apigeeLogEntry.StatusCode)
	if statusCode >= http.StatusOK && statusCode < http.StatusBadRequest {
		transSummaryStatus = transaction.TxSummaryStatusSuccess
	} else if statusCode >= http.StatusBadRequest && statusCode < http.StatusInternalServerError {
		transSummaryStatus = transaction.TxSummaryStatusFailure
	} else if statusCode >= http.StatusInternalServerError && statusCode < http.StatusNetworkAuthenticationRequired {
		transSummaryStatus = transaction.TxSummaryStatusException
	}

	return transaction.NewTransactionSummaryBuilder().
		SetTimestamp(stringToInt64(apigeeLogEntry.ClientStartTimeStamp)).
		SetTransactionID(apigeeLogEntry.MessageID).
		SetStatus(transSummaryStatus, apigeeLogEntry.StatusCode).
		SetDuration(getDuration(apigeeLogEntry)).
		SetTeam(teamID).
		SetEntryPoint("http", apigeeLogEntry.Verb, buildURI(apigeeLogEntry), buildURI(apigeeLogEntry)).
		// If the API is published to Central as unified catalog item/API service, see the Proxy details with the API definition
		// The Proxy.Name represents the name of the API
		// The Proxy.ID should be of format "remoteApiId_<ID Of the API on remote gateway>". Use transaction.FormatProxyID(<ID Of the API on remote gateway>) to get the formatted value.
		SetProxy(transaction.FormatProxyID(apigeeLogEntry.APIName), apigeeLogEntry.APIName, 1).
		Build()
	return nil, nil
}

func makeTimestamp(timeString string) int64 {
	t, err := time.Parse(time.RFC3339, timeString)
	if err != nil {
		t = time.Now()
	}
	return t.UnixNano() / int64(time.Millisecond)
}

func buildURI(apigeeLogEntry LogEntry) string {
	uri := fmt.Sprintf("%s-%s.apigee.net/%s%s", apigeeLogEntry.Organization, apigeeLogEntry.Environment, apigeeLogEntry.APIName, apigeeLogEntry.Path)
	return strings.ToLower(uri)
}

func buildURILeg(apigeeLogEntry LogEntry) string {
	uri := apigeeLogEntry.Organization + "-" + apigeeLogEntry.Environment + ".apigee.net"
	return uri
}

func stringToInt(s string) int {
	newString, _ := strconv.Atoi(s)
	return newString
}

func stringToInt64(s string) int64 {
	newString, _ := strconv.ParseInt(s, 10, 64)
	return newString
}

func getDuration(apigeeLogEntry LogEntry) int {
	duration := stringToInt(apigeeLogEntry.ClientEndTimeStamp) - stringToInt(apigeeLogEntry.ClientStartTimeStamp)
	return duration
}
