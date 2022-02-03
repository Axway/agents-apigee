package apigee

import (
	"fmt"

	"github.com/Axway/agent-sdk/pkg/notification"
	"github.com/Axway/agent-sdk/pkg/util/log"
)

const (
	newProduct       = "New Product"
	newPortal        = "New Portal"
	removedPortal    = "Removed Portal"
	processAPI       = "API Process"
	removedAPI       = "Removed API"
	apiValidatorWait = "API Validator"
)

type topic struct {
	inputChannel chan interface{}
	notifier     notification.Notifier
}

var topics = map[string]*topic{
	newProduct:       nil,
	newPortal:        nil,
	removedPortal:    nil,
	processAPI:       nil,
	removedAPI:       nil,
	apiValidatorWait: nil,
}

func createTopics() {
	// create all of the notification pubsubs
	for topicName := range topics {
		newTopic := &topic{
			inputChannel: make(chan interface{}),
		}
		var err error
		newTopic.notifier, err = notification.RegisterNotifier(topicName, newTopic.inputChannel)
		newTopic.notifier.Start()
		if err != nil {
			log.Errorf("could not create the necessary notifier: %s", err)
			return
		}
		topics[topicName] = newTopic
	}
}

func subscribeToTopic(topicName string) (chan interface{}, string, error) {
	if _, found := topics[topicName]; !found {
		return nil, "", fmt.Errorf("cannot subscribe to unknown topic %s", topicName)
	}
	subChan := make(chan interface{})
	sub, err := notification.Subscribe(topicName, subChan)
	if err != nil {
		return nil, "", err
	}
	return subChan, sub.GetID(), nil
}

func publishToTopic(topicName string, data interface{}) error {
	if _, found := topics[topicName]; !found {
		return fmt.Errorf("cannot publish to unknown topic %s", topicName)
	}
	topics[topicName].inputChannel <- data
	return nil
}

func unsubscribeFromTopic(topicName, id string) error {
	if _, found := topics[topicName]; !found {
		return fmt.Errorf("cannot unsubscribe to unknown topic %s", topicName)
	}
	// create all of the notification pubsubs
	return notification.Unsubscribe(topicName, id)
}
