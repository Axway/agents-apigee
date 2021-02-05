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
