package apigee

import (
	"fmt"
	"strings"

	"github.com/Axway/agent-sdk/pkg/apic/definitions"
	prov "github.com/Axway/agent-sdk/pkg/apic/provisioning"
	"github.com/Axway/agent-sdk/pkg/util/log"
	"github.com/Axway/agents-apigee/client/pkg/apigee"
	"github.com/Axway/agents-apigee/client/pkg/apigee/models"
)

type provisioner struct {
	client client
}

type client interface {
	CreateDeveloperApp(newApp models.DeveloperApp) (*models.DeveloperApp, error)
	RemoveDeveloperApp(appName, developerID string) error
	GetDeveloperID() string
	GetDeveloperApp(name string) (*models.DeveloperApp, error)
	GetAppCredential(appName, devID, key string) (*models.DeveloperAppCredentials, error)
	AddProductCredential(appName, devID, key string, cpr apigee.CredentialProvisionRequest) (*models.DeveloperAppCredentials, error)
	RemoveProductCredential(appName, devID, key, productName string) error
}

// NewProvisioner creates a type to implement the SDK Provisioning methods for handling subscriptions
func NewProvisioner(client client) prov.Provisioning {
	return &provisioner{
		client: client,
	}
}

// AccessRequestDeprovision - removes an api from an application
func (p provisioner) AccessRequestDeprovision(req prov.AccessRequest) prov.RequestStatus {
	ps := prov.NewRequestStatusBuilder()
	devID := p.client.GetDeveloperID()

	appName := req.GetApplicationName()
	if appName == "" {
		return failed(ps, fmt.Errorf("application name not found"))
	}

	app, err := p.client.GetDeveloperApp(appName)
	if err != nil {
		if ok := strings.Contains(err.Error(), "404"); ok {
			return ps.Success()
		}

		return failed(ps, fmt.Errorf("failed to retrieve app: %s", err))
	}

	var cred models.DeveloperAppCredentials
	// find the credential that the api is linked to
	for _, c := range app.Credentials {
		for _, p := range c.ApiProducts {
			if p.Apiproduct == req.GetAPIID() {
				cred = c
			}
		}
	}

	apiID := req.GetAPIID()
	if apiID == "" {
		return failed(ps, fmt.Errorf("%s not found", definitions.AttrExternalAPIID))
	}

	err = p.client.RemoveProductCredential(appName, devID, cred.ConsumerKey, apiID)
	if err != nil {
		return failed(ps, fmt.Errorf("failed to remove api %s from app: %s", "api-product-name", err))
	}

	return ps.Success()
}

// AccessRequestProvision - adds an api to an application
func (p provisioner) AccessRequestProvision(req prov.AccessRequest) prov.RequestStatus {
	ps := prov.NewRequestStatusBuilder()
	devID := p.client.GetDeveloperID()

	apiID := req.GetAPIID()
	if apiID == "" {
		return failed(ps, fmt.Errorf("%s name not found", definitions.AttrExternalAPIID))
	}

	appName := req.GetApplicationName()
	if appName == "" {
		return failed(ps, fmt.Errorf("application name not found"))
	}

	app, err := p.client.GetDeveloperApp(appName)
	if err != nil {
		return failed(ps, fmt.Errorf("failed to retrieve app %s: %s", appName, err))
	}

	// check if the api is linked to a credential
	for _, cred := range app.Credentials {
		for _, p := range cred.ApiProducts {
			if p.Apiproduct == apiID {
				return failed(ps, fmt.Errorf("api %s already added to app %s", apiID, appName))
			}
		}
	}

	cred := app.Credentials[0]
	cpr := apigee.CredentialProvisionRequest{
		ApiProducts: []string{apiID},
	}

	_, err = p.client.AddProductCredential(appName, devID, cred.ConsumerKey, cpr)
	if err != nil {
		return failed(ps, fmt.Errorf("error: %s", err))
	}

	return ps.Success()
}

// ApplicationRequestDeprovision - removes an app from apigee
func (p provisioner) ApplicationRequestDeprovision(req prov.ApplicationRequest) prov.RequestStatus {
	ps := prov.NewRequestStatusBuilder()

	appName := req.GetManagedApplicationName()
	if appName == "" {
		return failed(ps, fmt.Errorf("managed application %s not found", appName))
	}

	devID := p.client.GetDeveloperID()
	if p.client.GetDeveloperID() == "" {
		return failed(ps, fmt.Errorf("developer id not found"))
	}

	err := p.client.RemoveDeveloperApp(appName, devID)
	if err != nil {
		return failed(ps, fmt.Errorf("failed to delete app: %s", err))
	}

	return ps.Success()
}

// ApplicationRequestProvision - creates an apigee app
func (p provisioner) ApplicationRequestProvision(req prov.ApplicationRequest) prov.RequestStatus {
	ps := prov.NewRequestStatusBuilder()
	app := models.DeveloperApp{
		Attributes: []models.Attribute{
			apigee.ApigeeAgentAttribute,
		},
		DeveloperId: p.client.GetDeveloperID(),
		Name:        req.GetManagedApplicationName(),
	}

	_, err := p.client.CreateDeveloperApp(app)
	if err != nil {
		return failed(ps, fmt.Errorf("failed to create app: %s", err))
	}

	return ps.Success()
}

// CredentialDeprovision - Return success because there are no credentials to remove until the app is deleted
func (p provisioner) CredentialDeprovision(_ prov.CredentialRequest) prov.RequestStatus {
	return prov.NewRequestStatusBuilder().
		SetMessage("credential still active until application is removed").
		Success()
}

// CredentialProvision - retrieves the app credentials for oauth or api key authentication
func (p provisioner) CredentialProvision(req prov.CredentialRequest) (prov.RequestStatus, prov.Credential) {
	ps := prov.NewRequestStatusBuilder()

	appName := req.GetApplicationName()
	if appName == "" {
		return failed(ps, fmt.Errorf("application name not found")), nil
	}

	app, err := p.client.GetDeveloperApp(appName)
	if err != nil {
		return failed(ps, fmt.Errorf("error retrieving app: %s", err)), nil
	}

	key := app.Credentials[0].ConsumerKey
	secret := app.Credentials[0].ConsumerSecret

	var cr prov.Credential

	t := req.GetCredentialType()
	if t == "oauth" {
		cr = prov.NewCredentialBuilder().SetOAuth(key, secret)
	} else {
		cr = prov.NewCredentialBuilder().SetAPIKey(key)
	}

	return ps.Success(), cr
}

func failed(ps prov.RequestStatusBuilder, err error) prov.RequestStatus {
	ps.SetMessage(err.Error())
	log.Error(fmt.Sprintf("subscription provisioning - %s", err))
	return ps.Failed()
}
