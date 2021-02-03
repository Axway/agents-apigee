package apigee

// Headers - Type for request/response headers
type Headers map[string]string

// LogglyEvent - Event from loggly
type LogglyEvent struct {
	ID        string `json:"id"`
	Raw       string `json:"raw"`
	Timestamp int64  `json:"timestamp"`
	Logmsg    string `json:"logmsg"`
}

// LogglyEventsCollection - Represents the array of events the agent will receive
type LogglyEventsCollection struct {
	Events []LogglyEvent `json:"events"`
	Next   string        `json:"next"`
}

// LogEntry - Represents the structure of log entry the agent will receive
type LogEntry struct {
	Organization          string `json:"organization"`
	Environment           string `json:"environment"`
	APIName               string `json:"api"`
	APIRevision           string `json:"revision"`
	MessageID             string `json:"messageId"`
	ClientTLSVersion      string `json:"clientTlsVersion"`
	Verb                  string `json:"verb"`
	Path                  string `json:"path"`
	QueryString           int    `json:"queryString"`
	ClientIP              int    `json:"clientIP"`
	ClientHost            int    `json:"clientHost"`
	ClientStartTimeStamp  int    `json:"clientStartTimeStamp"`
	ClientEndTimeStamp    int    `json:"clientEndTimeStamp"`
	BytesReceived         int    `json:"bytesReceived"`
	BytesSent             int    `json:"bytesSent"`
	UserAgent             string `json:"UserAgent"`
	HTTPVersion           int    `json:"httpVersion"`
	ProxyURL              string `json:"proxyURL"`
	IsError               bool   `json:"isError"`
	StatusCode            int    `json:"statusCode"`
	ErrorStatusCode       int    `json:"errorStatusCode"`
	RequestHost           string `json:"requestHost"`
	RequestContentType    string `json:"requestContentType"`
	ResponseHost          string `json:"responseHost"`
	ResponseContentLength string `json:responseContentLength`
}
