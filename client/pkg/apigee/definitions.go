package apigee

const (
	defaultSubscriptionSchema = "apigee-subscription-schema"
	appNameKey                = "appName"
)

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

//Products
type Products []string

// PortalResponse
type PortalResponse struct {
	Status    string     `json:"status"`
	Message   string     `json:"message"`
	Code      string     `json:"code"`
	ErrorCode string     `json:"error_code"`
	RequestID string     `json:"request_id"`
	Data      PortalData `json:"data"`
}

// PortalsResponse
type PortalsResponse struct {
	Status    string       `json:"status"`
	Message   string       `json:"message"`
	Code      string       `json:"code"`
	ErrorCode string       `json:"error_code"`
	RequestID string       `json:"request_id"`
	Data      []PortalData `json:"data"`
}

type PortalData struct {
	ID                   string `json:"id"`
	Name                 string `json:"name"`
	Description          string `json:"description"`
	CustomDomain         string `json:"customDomain"`
	OrgName              string `json:"orgName"`
	Status               string `json:"status"`
	VisibleToCustomers   bool   `json:"visibleToCustomers"`
	HTTPS                bool   `json:"https"`
	DefaultDomain        string `json:"defaultDomain"`
	CustomeDomainEnabled bool   `json:"customDomainEnabled"`
	DefaultURL           string `json:"defaultURL"`
	CurrentURL           string `json:"currentURL"`
	CurrentDomain        string `json:"currentDomain"`
}

// APIDocDataResponse
type APIDocDataResponse struct {
	Status    string        `json:"status"`
	Message   string        `json:"message"`
	Code      string        `json:"code"`
	ErrorCode string        `json:"error_code"`
	RequestID string        `json:"request_id"`
	Data      []*APIDocData `json:"data"`
}

type APIDocData struct {
	ID               int     `json:"id"`
	PortalID         string  `json:"siteId"`
	Title            string  `json:"title"`
	Description      string  `json:"description"`
	APIID            string  `json:"apiId"`
	ProductName      string  `json:"edgeAPIProductName"`
	SpecContent      string  `json:"specContent"`
	SpecTitle        string  `json:"specTitle"`
	SpecID           string  `json:"specId"`
	ProductExists    bool    `json:"productExists"`
	Modified         int     `json:"modified"`
	SnapshotModified int     `json:"snapshotModified"`
	ImageURL         *string `json:"imageUrl"`
	CategoryIds      []int   `json:"categoryIds"`
	PortalTitle      string
}

func (a *APIDocData) SetPortalTitle(title string) {
	a.PortalTitle = title
}
