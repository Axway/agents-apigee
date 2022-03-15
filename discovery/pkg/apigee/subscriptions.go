package apigee

import (
	"fmt"

	"github.com/Axway/agent-sdk/pkg/agent"
	"github.com/Axway/agent-sdk/pkg/apic"
	"github.com/Axway/agent-sdk/pkg/cache"
	"github.com/Axway/agent-sdk/pkg/notify"
	"github.com/Axway/agent-sdk/pkg/util/log"
	"github.com/Axway/agents-apigee/client/pkg/apigee"
	"github.com/Axway/agents-apigee/client/pkg/apigee/models"
)

// registerSubscriptionSchema - create a subscription schema on Central
func (a *Agent) registerSubscriptionSchema() {
	apic.NewSubscriptionSchemaBuilder(agent.GetCentralClient()).
		SetName(defaultSubscriptionSchema).
		AddProperty(apic.NewSubscriptionSchemaPropertyBuilder().
			SetName(appDisplayNameKey).
			SetRequired().
			SetDescription("The application display name set in Apigee").
			IsString()).
		Register()
}

// handleSubscriptions - setup all things necessary to handle subscriptions from Central, but don't start the manager
func (a *Agent) handleSubscriptions() {
	err := agent.GetCentralClient().RegisterSubscriptionWebhook()
	if err != nil {
		log.Errorf("Unable to register subscription webhook: %s", err.Error())
		return
	}

	a.registerSubscriptionSchema()

	agent.GetCentralClient().GetSubscriptionManager()

	agent.GetCentralClient().GetSubscriptionManager().RegisterProcessor(apic.SubscriptionApproved, a.processSubscribe)
	agent.GetCentralClient().GetSubscriptionManager().RegisterProcessor(apic.SubscriptionUnsubscribeInitiated, a.processUnsubscribe)
	// agent.GetCentralClient().GetSubscriptionManager().RegisterValidator(a.validateSubscription)
}

func (a *Agent) sendSubscriptionNotification(subscription apic.Subscription, api apigee.APIDocData, key, secret string, newState apic.SubscriptionState, message string) {
	// Verify that at least 1 notification type was set.  If none was set, then do not attempt to gather user info or send notification
	if len(a.cfg.CentralCfg.GetSubscriptionConfig().GetNotificationTypes()) == 0 {
		log.Debug("No subscription notifications are configured.")
		return
	}

	catalogItemName, _ := agent.GetCentralClient().GetCatalogItemName(subscription.GetCatalogItemID())

	createdUserID := subscription.GetCreatedUserID()

	recipient, err := agent.GetCentralClient().GetUserEmailAddress(createdUserID)
	if err != nil {
		log.Error(err)
		return
	}
	catalogItemURL := fmt.Sprintf(a.cfg.CentralCfg.GetURL()+"/catalog/explore/%s", subscription.GetCatalogItemID())
	subNotif := notify.NewSubscriptionNotification(recipient, message, newState)
	subNotif.SetCatalogItemInfo(subscription.GetCatalogItemID(), catalogItemName, catalogItemURL)
	if api.HasAPIKey() {
		subNotif.SetAuthorizationTemplate(notify.Apikeys)
		subNotif.SetAPIKeyInfoAndLocation(key, api.GetAPIKeyInfo()[0].Name, api.GetAPIKeyInfo()[0].Location)
	}
	if api.HasOauth() {
		subNotif.SetOauthInfo(key, secret)
	}

	err = subNotif.NotifySubscriber(recipient)
	if err != nil {
		return
	}
}

func (a *Agent) processSubscribe(sub apic.Subscription) {
	// TODO: add rollback handling

	log.Tracef("received subscribe event for subscription id %s", sub.GetID())
	apiAttributes := sub.GetRemoteAPIAttributes()

	// check for the catalog id on the subscription event
	var catalogID string
	var ok bool
	if catalogID, ok = apiAttributes[catalogIDKey]; !ok {
		log.Errorf("subscription did not have a catalog ID key, %s, as expected", catalogIDKey)
		return
	}

	// get the api from the cache
	apiInterface, err := cache.GetCache().GetBySecondaryKey(catalogID)
	if err != nil {
		return // API not found
	}
	api, ok := apiInterface.(apigee.APIDocData)
	if !ok {
		log.Errorf("found item in cache with secondary key %s but it was not of the expected type", catalogID)
		return // Found cache item but not of APIDocData type
	}

	attributes := []models.Attribute{
		apigee.ApigeeAgentAttribute,
	}
	displayName := sub.GetPropertyValue(appDisplayNameKey)
	if displayName != "" {
		attributes = append(attributes, models.Attribute{
			Name:  "DisplayName",
			Value: displayName,
		})
	}

	// create app by name
	newApp := models.DeveloperApp{
		ApiProducts: []string{api.ProductName},
		Attributes:  attributes,
		DeveloperId: a.apigeeClient.GetDeveloperID(),
		Name:        sub.GetID(),
	}

	createdApp, err := a.apigeeClient.CreateDeveloperApp(newApp)
	if err != nil {
		log.Errorf("error attempting to create an app %s (%s): %s", displayName, sub.GetID(), err.Error())
		sub.UpdateState(apic.SubscriptionFailedToSubscribe, "Could not process the subscription, contact the administrator")
		return
	}
	if len(createdApp.Credentials) < 1 {
		log.Errorf("error getting credentials from new app %s (%s)", displayName, sub.GetID())
		sub.UpdateState(apic.SubscriptionFailedToSubscribe, "Could not process the subscription, contact the administrator")
		return
	}
	key := createdApp.Credentials[0].ConsumerKey
	secret := createdApp.Credentials[0].ConsumerSecret

	// success
	sub.UpdateState(apic.SubscriptionActive, "Successfully subscribed")

	// send notification
	a.sendSubscriptionNotification(sub, api, key, secret, apic.SubscriptionActive, "")
}

func (a *Agent) processUnsubscribe(sub apic.Subscription) {
	log.Tracef("received unsubscribe event for subscription id %s", sub.GetID())

	// delete the application using the subscription id
	a.apigeeClient.RemoveDeveloperApp(sub.GetID(), a.apigeeClient.GetDeveloperID())

	sub.UpdateState(apic.SubscriptionUnsubscribed, "Successfully unsubscribed")
}
