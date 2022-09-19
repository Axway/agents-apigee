package apigee

import (
	coreapi "github.com/Axway/agent-sdk/pkg/api"
)

type RequestOption func(*apigeeRequest)

type apigeeRequest struct {
	method      string
	url         string
	token       string
	headers     map[string]string
	queryParams map[string]string
	body        []byte
	client      coreapi.Client
}

func (r *apigeeRequest) Execute() (*coreapi.Response, error) {
	// return the api response
	request := coreapi.Request{
		Method:      r.method,
		URL:         r.url,
		Headers:     r.headers,
		QueryParams: r.queryParams,
		Body:        r.body,
	}
	return r.client.Send(request)
}

func (a *ApigeeClient) newRequest(method, url string, options ...RequestOption) *apigeeRequest {
	req := &apigeeRequest{method: method, url: url, client: a.apiClient, token: a.accessToken}
	for _, o := range options {
		o(req)
	}
	return req
}

// WithDefaultHeaders - add the default headers needed for apigee
func WithDefaultHeaders() RequestOption {
	return func(r *apigeeRequest) {
		if r.headers == nil {
			r.headers = make(map[string]string)
		}
		r.headers["Accept"] = "application/json"
		r.headers["Authorization"] = "Bearer " + r.token
	}
}

// WithHeaders - add additional headers to the request
func WithHeaders(headers map[string]string) RequestOption {
	return func(r *apigeeRequest) {
		if r.headers == nil {
			r.headers = make(map[string]string)
		}
		for key, val := range headers {
			r.headers[key] = val
		}
	}
}

// WithHeader - add an additional header to the request
func WithHeader(name, value string) RequestOption {
	return func(r *apigeeRequest) {
		if r.headers == nil {
			r.headers = make(map[string]string)
		}
		r.headers[name] = value
	}
}

// WithQueryParams - add query parameters to the request
func WithQueryParams(queryParams map[string]string) RequestOption {
	return func(r *apigeeRequest) {
		if r.queryParams == nil {
			r.queryParams = make(map[string]string)
		}
		for key, val := range queryParams {
			r.queryParams[key] = val
		}
	}
}

// WithQueryParam - add a query parameter to the request
func WithQueryParam(name, value string) RequestOption {
	return func(r *apigeeRequest) {
		if r.queryParams == nil {
			r.queryParams = make(map[string]string)
		}
		r.queryParams[name] = value
	}
}

// WithBody - add a JSON body to the request
func WithBody(body []byte) RequestOption {
	return func(r *apigeeRequest) {
		r.body = body
		if r.headers == nil {
			r.headers = make(map[string]string)
		}
		r.headers["Content-Type"] = "application/json"
	}
}

// WithStringBody - add a JSON body, from a string, to the request
func WithStringBody(body string) RequestOption {
	return func(r *apigeeRequest) {
		r.body = []byte(body)
		if r.headers == nil {
			r.headers = make(map[string]string)
		}
		r.headers["Content-Type"] = "application/json"
	}
}
