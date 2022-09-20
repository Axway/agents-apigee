package apigee

import "github.com/Axway/agents-apigee/client/pkg/apigee/models"

const (
	openapi     = "openapi"
	association = "association.json"
)

type Association struct {
	URL string `json:"url"`
}

type JobFirstRunDone func() bool

type APIRevision struct {
	models.ApiProxyRevision
}
