package gateway

// CHANGE_HERE - Change the structures below to represent the log entry the agent is going to receive

// Headers - Type for request/response headers
type Headers map[string]string

// ApigeeLogEntry - apigee transaction detail
type ApigeeLogEntry struct {
	ID              string  `json:"id"`
	SourceHost      string  `json:"srcHost"`
	SourcePort      int     `json:"srcPort"`
	DesHost         string  `json:"destHost"`
	DestPort        int     `json:"destPort"`
	URI             string  `json:"uri"`
	Method          string  `json:"method"`
	StatusCode      int     `json:"statusCode"`
	RequestHeaders  Headers `json:"requestHeaders"`
	ResponseHeaders Headers `json:"responseHeaders"`
	RequestBytes    int     `json:"requestByte"`
	ResponseBytes   int     `json:"responseByte"`
}

// LogEntry - Represents the structure of log entry the agent will receive
type LogEntry struct {
	TraceID             string        `json:"traceId"`
	APIName             string        `json:"apiName"`
	InboundTransaction  GwTransaction `json:"inbound"`
	OutboundTransaction GwTransaction `json:"outbound"`
}
