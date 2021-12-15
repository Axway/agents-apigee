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
	a.registerSubscriptionSchema()

	agent.GetCentralClient().GetSubscriptionManager()

	agent.GetCentralClient().GetSubscriptionManager().RegisterProcessor(apic.SubscriptionApproved, a.processSubscribe)
	agent.GetCentralClient().GetSubscriptionManager().RegisterProcessor(apic.SubscriptionUnsubscribeInitiated, a.processUnsubscribe)
	// agent.GetCentralClient().GetSubscriptionManager().RegisterValidator(a.validateSubscription)
}

func (a *Agent) sendSubscriptionNotification(subscription apic.Subscription, key string, newState apic.SubscriptionState, message string) {
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
	subNotif.SetAPIKeyInfo(key, "apikey")

	// force to apikey, since that is the only auth type available
	subNotif.SetAuthorizationTemplate(notify.Apikeys)

	err = subNotif.NotifySubscriber(recipient)
	if err != nil {
		return
	}
}

func (a *Agent) processSubscribe(sub apic.Subscription) {
	log.Tracef("received subscribe event for subscription id %s", sub.GetID())
	apiAttributes := sub.GetRemoteAPIAttributes()

	// check for the catalog id on the subscription event
	var catalogID string
	var ok bool
	if catalogID, ok = apiAttributes[catalogIDKey]; !ok {
		log.Errorf("subscrition did not have a catalog ID key, %s, as expected", catalogIDKey)
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
	apiKey := createdApp.Credentials[0].ConsumerKey

	// success
	// sub.UpdateState(apic.SubscriptionActive, "Successfully processed the subscription")

	// send notification
	a.sendSubscriptionNotification(sub, apiKey, apic.SubscriptionActive, "")
	return
}

func (a *Agent) processUnsubscribe(sub apic.Subscription) {
	log.Tracef("received unsubscribe event for subscription id %s", sub.GetID())
	apiAttributes := sub.GetRemoteAPIAttributes()
	c := cache.GetCache()
	api, err := cache.GetCache().Get(apiAttributes[catalogIDKey])
	_ = c
	_ = api
	_ = err

	// get product by name

	// get app by name
	return
}
