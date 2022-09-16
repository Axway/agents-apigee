package apigee

// AuthResponse - response struct from APIGEE auth call
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

type metricCache struct {
	ProxyName           string `json:"proxy"`
	Timestamp           int64  `json:"timestamp"`
	Total               int    `json:"total"`
	PolicyError         int    `json:"policyError"`
	ServerError         int    `json:"serverError"`
	Success             int    `json:"success"`
	ReportedPolicyError int    `json:"reportedPolicyError"`
	ReportedServerError int    `json:"reportedServerError"`
	ReportedSuccess     int    `json:"reportedSuccess"`
	ResponseTime        int64  `json:"responseTime"`
}
