package apigee

// Headers - Type for request/response headers
type Headers map[string]string

// EventApigeeEntry - Type for event hub entry
type EventApigeeEntry struct {
	Records []EventEntry `json:"records"`
}

// EventEntry - Type for gateway transaction detail
type EventEntry struct {
	Level            int             `json:"Level"`
	IsRequestSuccess bool            `json:"isRequestSuccess"`
	Time             string          `json:"time"`
	OperationName    string          `json:"operationName"`
	Category         string          `json:"category"`
	DurationMs       int             `json:"durationMs"`
	CallerIPAddress  string          `json:"callerIpAddress"`
	CorrelationID    string          `json:"correlationId"`
	Location         string          `json:"location"`
	Properties       GatewayLogEntry `json:"properties"`
	ResourceID       string          `json:"resourceId"`
}

// GatewayLogEntry - Represents the structure of log entry the agent will receive
type GatewayLogEntry struct {
	APIID                  string  `json:"apiId"`
	OperationID            string  `json:"operationId"`
	APIMSubscriptionID     string  `json:"apimSubscriptionId"`
	APIRevision            string  `json:"apiRevision"`
	ClientProtocol         string  `json:"clientProtocol"`
	ClientTLSVersion       string  `json:"clientTlsVersion"`
	Method                 string  `json:"method"`
	URL                    string  `json:"url"`
	RequestSize            int     `json:"requestSize"`
	ResponseCode           int     `json:"responseCode"`
	ResponseSize           int     `json:"responseSize"`
	Cache                  string  `json:"cache"`
	BackendTime            int     `json:"backendTime"`
	BackendProtocol        string  `json:"backendProtocol"`
	BackendMethod          string  `json:"backendMethod"`
	BackendURL             string  `json:"backendUrl"`
	BackendResponseCode    int     `json:"backendResponseCode"`
	ResponseHeaders        Headers `json:"responseHeaders"`
	RequestHeaders         Headers `json:"requestHeaders"`
	BackendResponseHeaders Headers `json:"backendResponseHeaders"`
	BackendRequestHeaders  Headers `json:"backendRequestHeaders"`
}
