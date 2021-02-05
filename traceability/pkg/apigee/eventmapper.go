package apigee

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/Axway/agent-sdk/pkg/agent"
	"github.com/Axway/agent-sdk/pkg/transaction"
	"github.com/Axway/agent-sdk/pkg/util"
	"github.com/Axway/agent-sdk/pkg/util/log"
)

// EventMapper -
type EventMapper struct {
}

func (m *EventMapper) processMapping(apigeeLogEntry LogEntry) ([]transaction.LogEvent, error) {
	centralCfg := agent.GetCentralConfig()
	inboundHTTPProtocol, err := transaction.NewHTTPProtocolBuilder().
		SetURI(buildURI(apigeeLogEntry)).
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
		SetDuration(stringToInt(apigeeLogEntry.ClientStartTimeStamp) - stringToInt(apigeeLogEntry.ClientEndTimeStamp)).
		SetTransactionID(apigeeLogEntry.MessageID).
		SetID(apigeeLogEntry.MessageID + "-leg0").
		// TODO :
		// TransactionID and ID : SetID(apigeeLogEntry.Properties.OperationID + "-leg0").
		SetSource("client").
		SetDestination(util.GetURLHostName(apigeeLogEntry.RequestHost)).
		SetDirection("Inbound").
		SetStatus(txEventStatus).
		SetProtocolDetail(inboundHTTPProtocol).
		Build()
	if err != nil {
		return nil, err
	}

	//TODO - all outbound leg is same as inbound leg
	outboundHTTPProtocol, err := transaction.NewHTTPProtocolBuilder().
		SetURI(buildURI(apigeeLogEntry)).
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
		SetDuration(stringToInt(apigeeLogEntry.ClientStartTimeStamp) - stringToInt(apigeeLogEntry.ClientEndTimeStamp)).
		SetTransactionID(apigeeLogEntry.MessageID).
		SetID(apigeeLogEntry.MessageID + "-leg1"). //TODO diff between transactionID and ID
		SetParentID(apigeeLogEntry.MessageID + "-leg0").
		SetSource(util.GetURLHostName(apigeeLogEntry.RequestHost)).
		SetDestination(util.GetURLHostName(apigeeLogEntry.RequestHost)).
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
		SetDuration(stringToInt(apigeeLogEntry.ClientStartTimeStamp)-stringToInt(apigeeLogEntry.ClientEndTimeStamp)).
		SetTeam(teamID).
		SetEntryPoint("http", apigeeLogEntry.Verb, apigeeLogEntry.RequestHost, util.GetURLHostName(apigeeLogEntry.RequestHost)).
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
	uri := apigeeLogEntry.RequestHost + apigeeLogEntry.Path
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
