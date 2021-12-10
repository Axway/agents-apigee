package apigee

import (
	"encoding/json"
	"net/url"

	coreapi "github.com/Axway/agent-sdk/pkg/api"
	"github.com/Axway/agent-sdk/pkg/jobs"
	"github.com/Axway/agent-sdk/pkg/util/log"
)

const (
	apigeeAuthURL      = "https://login.apigee.com/oauth/token"
	apigeeAuthCheckURL = "https://login.apigee.com/login"
	apigeeAuthToken    = "ZWRnZWNsaTplZGdlY2xpc2VjcmV0" //hardcoded to edgecli:edgeclisecret
	grantTypeKey       = "grant_type"
	usernameKey        = "username"
	passwordKey        = "password"
	refreshTokenKey    = "refresh_token"
)

func newAuthJob(apiclient coreapi.Client, username, password string, tokenSetter func(string)) *authJob {
	return &authJob{
		apiClient:   apiclient,
		username:    username,
		password:    password,
		tokenSetter: tokenSetter,
	}
}

type authJob struct {
	jobs.Job
	apiClient    coreapi.Client
	refreshToken string
	username     string
	password     string
	tokenSetter  func(string)
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
	request := coreapi.Request{
		Method: coreapi.GET,
		URL:    apigeeAuthCheckURL,
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
	var err error
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
	request := coreapi.Request{
		Method: coreapi.POST,
		URL:    apigeeAuthURL,
		Headers: map[string]string{
			"Content-Type":  "application/x-www-form-urlencoded",
			"Authorization": "Basic " + apigeeAuthToken,
		},
		Body: []byte(authData.Encode()),
	}

	// Get the initial authentication token
	response, err := j.apiClient.Send(request)
	if err != nil {
		log.Errorf(err.Error())
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
