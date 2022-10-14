package apigee

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"

	coreapi "github.com/Axway/agent-sdk/pkg/api"
	"github.com/Axway/agent-sdk/pkg/jobs"
	"github.com/Axway/agent-sdk/pkg/util/log"
)

const (
	apigeeAuthPath      = "/oauth/token"
	apigeeAuthCheckPath = "/login"
	grantTypeKey        = "grant_type"
	usernameKey         = "username"
	passwordKey         = "password"
	refreshTokenKey     = "refresh_token"
)

type authJobOpt func(*authJob)

func newAuthJob(opts ...authJobOpt) *authJob {
	a := &authJob{}
	for _, o := range opts {
		o(a)
	}
	return a
}

func withAPIClient(apiClient coreapi.Client) authJobOpt {
	return func(a *authJob) {
		a.apiClient = apiClient
	}
}

func withUsername(username string) authJobOpt {
	return func(a *authJob) {
		a.username = username
	}
}

func withPassword(password string) authJobOpt {
	return func(a *authJob) {
		a.password = password
	}
}

func withTokenSetter(tokenSetter func(string)) authJobOpt {
	return func(a *authJob) {
		a.tokenSetter = tokenSetter
	}
}

func withURL(url string) authJobOpt {
	return func(a *authJob) {
		a.url = url
	}
}

func withAuthServerUsername(username string) authJobOpt {
	return func(a *authJob) {
		a.serverUsername = username
	}
}

func withAuthServerPassword(password string) authJobOpt {
	return func(a *authJob) {
		a.serverPassword = password
	}
}

type authJob struct {
	jobs.Job
	apiClient      coreapi.Client
	refreshToken   string
	username       string
	password       string
	url            string
	serverUsername string
	serverPassword string
	tokenSetter    func(string)
}

func (j *authJob) Ready() bool {
	err := j.passwordAuth()
	if err != nil {
		log.Error(err)
		return false
	}
	return true
}

func (j *authJob) Status() error {
	return nil
}

func (j *authJob) checkConnection() error {
	request := coreapi.Request{
		Method: coreapi.GET,
		URL:    fmt.Sprintf("%s%s", j.url, apigeeAuthCheckPath),
	}

	// Validate we can reach the apigee auth server
	_, err := j.apiClient.Send(request)
	if err != nil {
		log.Errorf(err.Error())
		return err
	}

	return nil
}

func (j *authJob) Execute() error {
	err := j.checkConnection()
	if err != nil {
		return err
	}

	if j.refreshToken != "" {
		err = j.refreshAuth()
	}
	if err != nil {
		err = j.passwordAuth()
	}
	return err
}

func (j *authJob) passwordAuth() error {
	log.Tracef("Getting new auth token")
	authData := url.Values{}
	authData.Set(grantTypeKey, password.String())
	authData.Set(usernameKey, j.username)
	authData.Set(passwordKey, j.password)

	err := j.postAuth(authData)
	if err != nil {
		// clear out the refreshToken attribute
		j.refreshToken = ""
	}
	return err
}

func (j *authJob) refreshAuth() error {
	log.Tracef("Refreshing auth token")
	authData := url.Values{}
	authData.Set(grantTypeKey, refresh.String())
	authData.Set(refreshTokenKey, j.refreshToken)

	return j.postAuth(authData)
}

func (j *authJob) postAuth(authData url.Values) error {
	basicAuth := base64.StdEncoding.EncodeToString([]byte(fmt.Sprintf("%s:%s", j.serverUsername, j.serverPassword)))
	request := coreapi.Request{
		Method: coreapi.POST,
		URL:    fmt.Sprintf("%s%s", j.url, apigeeAuthPath),
		Headers: map[string]string{
			"Content-Type":  "application/x-www-form-urlencoded",
			"Authorization": "Basic " + basicAuth,
		},
		Body: []byte(authData.Encode()),
	}

	// Get the initial authentication token
	response, err := j.apiClient.Send(request)
	if err != nil {
		log.Errorf(err.Error())
		return err
	}

	// if the response code is not ok log and return an err
	if response.Code != http.StatusOK {
		err := fmt.Errorf("unexpected response code %d from authentication call: %s", response.Code, response.Body)
		log.Error(err)
		return err
	}

	// save this refreshToken and send the token to the client
	authResponse := AuthResponse{}
	json.Unmarshal(response.Body, &authResponse)
	log.Trace(authResponse.AccessToken)
	j.refreshToken = authResponse.RefreshToken
	j.tokenSetter(authResponse.AccessToken)
	return nil
}
