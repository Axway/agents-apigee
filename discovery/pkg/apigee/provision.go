package apigee

import (
	"fmt"

	prov "github.com/Axway/agent-sdk/pkg/apic/provisioning"
	"github.com/Axway/agent-sdk/pkg/util/log"
	"github.com/Axway/agents-apigee/client/pkg/apigee"
	"github.com/Axway/agents-apigee/client/pkg/apigee/models"
)

const (
	appName = "app-name"
	appID   = "app-id"
)

type provisioner struct {
	client client
}

type client interface {
	CreateDeveloperApp(newApp models.DeveloperApp) (*models.DeveloperApp, error)
	RemoveDeveloperApp(appName, developerID string) error
	GetDeveloperID() string
	GetDeveloperApp(name string) (*models.DeveloperApp, error)
	UpdateDeveloperApp(app models.DeveloperApp) (*models.DeveloperApp, error)
}

// NewProvisioner creates a type to implement the SDK Provisioning methods for handling subscriptions
func NewProvisioner(client client) prov.Provisioning {
	return &provisioner{
		client: client,
	}
}

// AccessRequestDeprovision -
func (p provisioner) AccessRequestDeprovision(req prov.AccessRequest) prov.RequestStatus {
	ps := prov.NewRequestStatusBuilder()

	app, err := p.client.GetDeveloperApp(req.GetApplicationName())
	if err != nil {
		return failed(ps, fmt.Errorf("error retrieving app: %s", err))
	}

	var apiProducts []string
	for _, api := range apiProducts {
		if api != "api-product-name" {
			apiProducts = append(apiProducts, api)
		}
	}

	_, err = p.client.UpdateDeveloperApp(*app)
	if err != nil {
		return failed(ps, fmt.Errorf("failed to remove api %s from app: %s", "api-product-name", err))
	}

	return ps.Success()
}

// AccessRequestProvision -
func (p provisioner) AccessRequestProvision(req prov.AccessRequest) prov.RequestStatus {
	ps := prov.NewRequestStatusBuilder()
	app, err := p.client.GetDeveloperApp(req.GetApplicationName())
	if err != nil {
		log.Error(err)
	}

	// TODO: should have a way to get the defs.AttrExternalAPIName
	app.ApiProducts = append(app.ApiProducts, "api-product-name")

	_, err = p.client.UpdateDeveloperApp(*app)
	if err != nil {
		return failed(ps, fmt.Errorf("failed to add api %s to app: %s", "api-product-name", err))
	}

	return ps.Success()
}

// ApplicationRequestDeprovision -
func (p provisioner) ApplicationRequestDeprovision(req prov.ApplicationRequest) prov.RequestStatus {
	ps := prov.NewRequestStatusBuilder()
	err := p.client.RemoveDeveloperApp(req.GetManagedApplicationName(), p.client.GetDeveloperID())
	if err != nil {
		return failed(ps, fmt.Errorf("failed to delete app: %s", err))
	}
	return ps.Success()
}

// ApplicationRequestProvision - creates an app
func (p provisioner) ApplicationRequestProvision(req prov.ApplicationRequest) prov.RequestStatus {
	ps := prov.NewRequestStatusBuilder()
	app := models.DeveloperApp{
		Attributes: []models.Attribute{
			apigee.ApigeeAgentAttribute,
		},
		DeveloperId: p.client.GetDeveloperID(),
		Name:        req.GetManagedApplicationName(),
	}

	res, err := p.client.CreateDeveloperApp(app)
	if err != nil {
		return failed(ps, fmt.Errorf("failed to create app: %s", err))
	}

	ps.AddProperty(appID, res.AppId)
	ps.AddProperty(appName, res.Name)

	return ps.Success()
}

// CredentialDeprovision -
func (p provisioner) CredentialDeprovision(req prov.CredentialRequest) prov.RequestStatus {
	ps := prov.NewRequestStatusBuilder()

	return ps.Success()
}

// CredentialProvision -
func (p provisioner) CredentialProvision(req prov.CredentialRequest) (prov.RequestStatus, prov.Credential) {
	ps := prov.NewRequestStatusBuilder()

	app, err := p.client.GetDeveloperApp(req.GetApplicationName())
	if err != nil {
		return failed(ps, fmt.Errorf("error retrieving app: %s", err)), nil
	}

	key := app.Credentials[0].ConsumerKey
	// secret := app.Credentials[0].ConsumerSecret

	cr := prov.NewCredentialBuilder().SetAPIKey(key)

	return ps.Success(), cr
}

func failed(ps prov.RequestStatusBuilder, err error) prov.RequestStatus {
	ps.SetMessage(err.Error())
	log.Error(fmt.Sprintf("subscription provisioning - %s", err))
	return ps.Failed()
}
