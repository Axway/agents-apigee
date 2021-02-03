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
	// TODO
	ResponseHeaders Headers `json:"responseHeaders"`
	RequestHeaders  Headers `json:"requestHeaders"`
}

/*
Current raw message
{
	"organization":"jasonmscollins-eval",
	"environment": "prod",
	"api": "Petstore",
	"revision": "3",
	"messageId": "rrt-7724201184537129167-f-gce-10936-8390016-1",
	"verb": "GET",
	"path": "/petstore/store/inventory",
	"queryString": "",
	"clientIP": "184.101.205.182",
	"clientHost": "184.101.205.182",
	"clientStartTimeStamp": "1612378803270",
	"clientEndTimeStamp": "1612378803298",
	"bytesReceived": "",
	"bytesSent": "0",
	"userAgent": "PostmanRuntime/7.26.8",
	"httpVersion": "1.1",
	"proxyURL": "https://jasonmscollins-eval-prod.apigee.net/petstore/store/inventory?apikey=DTuyUFjHAgXmrPGjPJs6Auql43A5THVJ",
	"isError": "false",
	"statusCode": "200",
	"errorStatusCode": "",
	"requestHost":"jasonmscollins-eval-prod.apigee.net",
	"responseHost":"jasonmscollins-eval-prod.apigee.net",
	"responseContentLength":"0",
	"requestContentType":"application/json"
}
*/
