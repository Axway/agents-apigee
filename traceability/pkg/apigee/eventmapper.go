package apigee

import (
	"encoding/json"
	"time"

	"github.com/Axway/agent-sdk/pkg/transaction"
	"github.com/Axway/agent-sdk/pkg/util/log"
)

// EventMapper -
type EventMapper struct {
}

func (m *EventMapper) processMapping(apigeeLogEntry LogEntry) ([]transaction.LogEvent, error) {
	// centralCfg := agent.GetCentralConfig()
	// inboundReqHeaders := m.buildHeaders(apigeeLogEntry.Properties.RequestHeaders)
	// intboundResHeaders := m.buildHeaders(apigeeLogEntry.Properties.ResponseHeaders)
	// inboundHTTPProtocol, err := transaction.NewHTTPProtocolBuilder().
	// 	SetURI(apigeeLogEntry.Properties.URL).
	// 	SetMethod(apigeeLogEntry.Properties.Method).
	// 	SetStatus(apigeeLogEntry.Properties.ResponseCode, http.StatusText(apigeeLogEntry.Properties.ResponseCode)).
	// 	SetHost(apigeeLogEntry.Location).
	// 	SetHeaders(inboundReqHeaders, intboundResHeaders).
	// 	SetByteLength(apigeeLogEntry.Properties.RequestSize, apigeeLogEntry.Properties.ResponseSize).
	// 	SetRemoteAddress("", apigeeLogEntry.Location, 443).
	// 	SetLocalAddress("", 0).
	// 	Build()
	// if err != nil {
	// 	return nil, err
	// }
	// txEventStatus := transaction.TxEventStatusFail
	// if apigeeLogEntry.Properties.ResponseCode < 400 {
	// 	txEventStatus = transaction.TxEventStatusPass
	// }

	// transInboundLogEventLeg, err := transaction.NewTransactionEventBuilder().
	// 	SetTimestamp(makeTimestamp(apigeeLogEntry.Time)).
	// 	SetDuration(apigeeLogEntry.DurationMs).
	// 	SetTransactionID(apigeeLogEntry.CorrelationID).
	// 	SetID(apigeeLogEntry.Properties.OperationID + "-leg0").
	// 	SetSource("client").
	// 	SetDestination(util.GetURLHostName(apigeeLogEntry.Properties.URL)).
	// 	SetDirection("Inbound").
	// 	SetStatus(txEventStatus).
	// 	SetProtocolDetail(inboundHTTPProtocol).
	// 	Build()
	// if err != nil {
	// 	return nil, err
	// }

	// outboundReqHeaders := m.buildHeaders((apigeeLogEntry.Properties.BackendRequestHeaders))
	// outboundResHeaders := m.buildHeaders((apigeeLogEntry.Properties.BackendResponseHeaders))
	// outboundHTTPProtocol, err := transaction.NewHTTPProtocolBuilder().
	// 	SetURI(apigeeLogEntry.Properties.BackendURL).
	// 	SetMethod(apigeeLogEntry.Properties.BackendMethod).
	// 	SetStatus(apigeeLogEntry.Properties.BackendResponseCode, http.StatusText(apigeeLogEntry.Properties.BackendResponseCode)).
	// 	SetHost(apigeeLogEntry.Properties.BackendURL).
	// 	SetHeaders(outboundReqHeaders, outboundResHeaders).
	// 	SetByteLength(apigeeLogEntry.Properties.RequestSize, apigeeLogEntry.Properties.ResponseSize).
	// 	SetRemoteAddress("", apigeeLogEntry.Properties.BackendURL, 443).
	// 	SetLocalAddress("", 0). // Need to populate this
	// 	Build()
	// if err != nil {
	// 	return nil, err
	// }
	// txEventStatus = transaction.TxEventStatusFail
	// if apigeeLogEntry.Properties.BackendResponseCode < 400 {
	// 	txEventStatus = transaction.TxEventStatusPass
	// }

	// transOutboundLogEventLeg, err := transaction.NewTransactionEventBuilder().
	// 	SetTimestamp(makeTimestamp(apigeeLogEntry.Time)).
	// 	SetDuration(apigeeLogEntry.DurationMs).
	// 	SetTransactionID(apigeeLogEntry.CorrelationID).
	// 	SetID(apigeeLogEntry.Properties.OperationID + "-leg1").
	// 	SetParentID(apigeeLogEntry.Properties.OperationID + "-leg0").
	// 	SetSource(util.GetURLHostName(apigeeLogEntry.Properties.URL)).
	// 	SetDestination(util.GetURLHostName(apigeeLogEntry.Properties.BackendURL)).
	// 	SetDirection("Outbound").
	// 	SetStatus(txEventStatus).
	// 	SetProtocolDetail(outboundHTTPProtocol).
	// 	Build()
	// if err != nil {
	// 	return nil, err
	// }

	// transSummaryLogEvent, err := m.createSummaryEvent(apigeeLogEntry, centralCfg.GetTeamID())
	// if err != nil {
	// 	return nil, err
	// }

	// return []transaction.LogEvent{
	// 	*transSummaryLogEvent,
	// 	*transInboundLogEventLeg,
	// 	*transOutboundLogEventLeg,
	// }, nil
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
	// transSummaryStatus := transaction.TxSummaryStatusUnknown
	// statusCode := apigeeLogEntry.Properties.ResponseCode
	// if statusCode >= http.StatusOK && statusCode < http.StatusBadRequest {
	// 	transSummaryStatus = transaction.TxSummaryStatusSuccess
	// } else if statusCode >= http.StatusBadRequest && statusCode < http.StatusInternalServerError {
	// 	transSummaryStatus = transaction.TxSummaryStatusFailure
	// } else if statusCode >= http.StatusInternalServerError && statusCode < http.StatusNetworkAuthenticationRequired {
	// 	transSummaryStatus = transaction.TxSummaryStatusException
	// }

	// return transaction.NewTransactionSummaryBuilder().
	// 	SetTimestamp(makeTimestamp(apigeeLogEntry.Time)).
	// 	SetTransactionID(apigeeLogEntry.CorrelationID).
	// 	SetStatus(transSummaryStatus, strconv.Itoa(apigeeLogEntry.Properties.ResponseCode)).
	// 	SetDuration(apigeeLogEntry.DurationMs).
	// 	SetTeam(teamID).
	// 	SetEntryPoint("http", apigeeLogEntry.Properties.Method, apigeeLogEntry.Properties.URL, util.GetURLHostName(apigeeLogEntry.Properties.URL)).
	// 	// If the API is published to Central as unified catalog item/API service, see the Proxy details with the API definition
	// 	// The Proxy.Name represents the name of the API
	// 	// The Proxy.ID should be of format "remoteApiId_<ID Of the API on remote gateway>". Use transaction.FormatProxyID(<ID Of the API on remote gateway>) to get the formatted value.
	// 	SetProxy(transaction.FormatProxyID(apigeeLogEntry.Properties.APIID), apigeeLogEntry.Properties.APIID, 1).
	// 	Build()
	return nil, nil
}

func makeTimestamp(timeString string) int64 {
	t, err := time.Parse(time.RFC3339, timeString)
	if err != nil {
		t = time.Now()
	}
	return t.UnixNano() / int64(time.Millisecond)
}
