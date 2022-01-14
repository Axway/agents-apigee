package apigee

// grantType values
type grantType int

const (
	password grantType = iota
	refresh
)

func (g grantType) String() string {
	return [...]string{"password", "refresh_token"}[g]
}

//AuthResponse - response struct from APIGEE auth call
type AuthResponse struct {
	AccessToken  string `json:"access_token"`
	TokenType    string `json:"token_type"`
	RefreshToken string `json:"refresh_token"`
	ExpiresIn    int    `json:"expires_in"`
	Scope        string `json:"scope"`
	JTI          string `json:"jti"`
}

// Headers - Type for request/response headers
type Headers map[string]string

// LogglyEvent - Event from loggly
type LogglyEvent struct {
	ID        string `json:"id"`
	Raw       string `json:"raw"`
	Timestamp int64  `json:"timestamp"`
	Logmsg    string `json:"logmsg"`
}

// LogglyRSID -
type LogglyRSID struct {
	ID     string `json:"id"`
	Status string `json:"status"`
}

// LogglySearchResponse - Represents the array of events the agent will receive
type LogglySearchResponse struct {
	RSID LogglyRSID `json:"rsid"`
}

// LogglyEventsCollection - Represents the array of events the agent will receive
type LogglyEventsCollection struct {
	Events      []LogglyEvent `json:"events"`
	Page        int           `json:"page"`
	TotalEvents int           `json:"total_events"`
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
	QueryString           string `json:"queryString"`
	ClientIP              string `json:"clientIP"`
	ClientHost            string `json:"clientHost"`
	ClientStartTimeStamp  string `json:"clientStartTimeStamp"`
	ClientEndTimeStamp    string `json:"clientEndTimeStamp"`
	BytesReceived         string `json:"bytesReceived"`
	BytesSent             string `json:"bytesSent"`
	UserAgent             string `json:"UserAgent"`
	HTTPVersion           string `json:"httpVersion"`
	ProxyURL              string `json:"proxyURL"`
	IsError               string `json:"isError"`
	StatusCode            string `json:"statusCode"`
	ErrorStatusCode       string `json:"errorStatusCode"`
	RequestHost           string `json:"requestHost"`
	RequestContentType    string `json:"requestContentType"`
	ResponseHost          string `json:"responseHost"`
	ResponseContentLength string `json:"responseContentLength"`
	ResponseHeaders       string `json:"responseHeaders"`
	RequestHeaders        string `json:"requestHeaders"`
}
