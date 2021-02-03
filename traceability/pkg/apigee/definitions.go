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

//AuthResponse - response struct from APIGEE auth call
type AuthResponse struct {
	AccessToken  string `json:"access_token"`
	TokenType    string `json:"token_type"`
	RefreshToken string `json:"refresh_token"`
	ExpiresIn    int    `json:"expires_in"`
	Scope        string `json:"scope"`
	JTI          string `json:"jti"`
}

// grantType values
type grantType int

const (
	password grantType = iota
	refresh
)

func (g grantType) String() string {
	return [...]string{"password", "refresh_token"}[g]
}

//Environments
type environments []string

// apiProxies
type apiProxies []string

// apigeeLogs - apigee logs
type apigeeLogs struct {
	Log []apigeeLog `json:"logs"`
}

type apigeeLog struct {
	FaultCode        string  `json:"fault_code"`
	FaultFlow        string  `json:"fault_flow"`
	FaultPolicy      string  `json:"fault_policy"`
	FaultProxy       string  `json:"fault_proxy"`
	FaultSource      string  `json:"fault_source"`
	Request          string  `json:"request"`
	RequestLength    int64   `json:"request_length"`
	RequestMessageID string  `json:"request_message_id"`
	ResponseSize     int64   `json:"response_size"`
	ResponseStatus   string  `json:"response_status"`
	ResponseTime     float64 `json:"response_time"`
	Timestamp        string  `json:"timestamp"`
	VirtualHost      string  `json:"virtual_host"`
}

// events
type apigeeEvents struct {
	Event []apigeeEvent `json:"events"`
}

type apigeeEvent struct {
	ID                   string `json:"id"`
	SharedID             string `json:"shared_id"`
	EntityKey            string `json:"entity_key"`
	EntityValue          string `json:"entity_value"`
	DependentEntityValue string `json:"dependent_entity_string"`
	Component            string `json:"component"`
	Pod                  string `json:"pod"`
	Region               string `json:"region"`
	Organization         string `json:organization"`
	Environment          string `json:environment"`
	Name                 string `json:"name"`
	Type                 string `json:"type"`
	Source               string `json:"source"`
	RawPayload           string `json:"raw_payload"` //TODO shane
	Time                 string `json:"time"`
}
