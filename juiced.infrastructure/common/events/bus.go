package events

import (
	"backend.juicedbot.io/juiced.infrastructure/common/enums"
	"backend.juicedbot.io/juiced.infrastructure/queries"

	"sync"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// EventBus stores the information about subscribers
type EventBus struct {
	Subscribers EventChannelSlice
	RM          sync.RWMutex
}

// Unsubscribe removes the EventChannel from the EventBus's subscribers
func (eb *EventBus) Unsubscribe(ch EventChannel) {
	eb.RM.Lock()
	newSubscribers := EventChannelSlice{}
	for _, subscriber := range eb.Subscribers {
		if ch != subscriber {
			newSubscribers = append(newSubscribers, subscriber)
		}
	}
	eb.Subscribers = newSubscribers
	close(ch)
	eb.RM.Unlock()
}

// Subscribe adds the EventChannel to the EventBus's subscribers
func (eb *EventBus) Subscribe(ch EventChannel) {
	eb.RM.Lock()
	eb.Subscribers = append(eb.Subscribers, ch)
	eb.RM.Unlock()
}

// PublishAuthEvent publishes an AuthEvent
func (eb *EventBus) PublishAuthEvent() {
	eb.RM.RLock()
	defer func() {
		recover()
	}()
	// Will panic if any channel is closed
	go func(event Event, channels EventChannelSlice) {
		for _, ch := range channels {
			ch <- event
		}
	}(Event{
		EventType: AuthEventType,
		AuthEvent: AuthEvent{},
	}, eb.Subscribers)
	eb.RM.RUnlock()
}

// PublishCloseEvent publishes a CloseEvent
func (eb *EventBus) PublishCloseEvent() {
	eb.RM.RLock()
	defer func() {
		recover()
	}()
	// Will panic if any channel is closed
	go func(event Event, channels EventChannelSlice) {
		for _, ch := range channels {
			ch <- event
		}
	}(Event{
		EventType:  CloseEventType,
		CloseEvent: CloseEvent{},
	}, eb.Subscribers)
	eb.RM.RUnlock()
}

// PublishMonitorEvent publishes a MonitorEvent
func (eb *EventBus) PublishMonitorEvent(monitorStatus enums.MonitorStatus, eventType enums.MonitorEventType, data interface{}, monitorID primitive.ObjectID) {
	eb.RM.RLock()
	defer func() {
		recover()
	}()
	// Will panic if any channel is closed
	go func(event Event, channels EventChannelSlice) {
		for _, ch := range channels {
			ch <- event
		}
	}(Event{
		EventType: MonitorEventType,
		MonitorEvent: MonitorEvent{
			Status:    monitorStatus,
			EventType: eventType,
			Data:      data,
			MonitorID: monitorID,
		},
	}, eb.Subscribers)
	eb.RM.RUnlock()
}

// PublishTaskEvent publishes a TaskEvent
func (eb *EventBus) PublishTaskEvent(taskStatus enums.TaskStatus, eventType enums.TaskEventType, data interface{}, taskID primitive.ObjectID) {
	eb.RM.RLock()
	defer func() {
		recover()
	}()
	// Will panic if any channel is closed
	go func(event Event, channels EventChannelSlice) {
		for _, ch := range channels {
			ch <- event
		}
	}(Event{
		EventType: TaskEventType,
		TaskEvent: TaskEvent{
			Status:    taskStatus,
			EventType: eventType,
			Data:      data,
			TaskID:    taskID,
		},
	}, eb.Subscribers)
	eb.RM.RUnlock()
}

// PublishProductEvent publishes a ProductEvent for any site
func (eb *EventBus) PublishProductEvent(retailer enums.Retailer, data interface{}, monitorID primitive.ObjectID) {
	eb.RM.RLock()
	// By retrieving this right after picking up the product data, we ensure we have the latest webhook info
	// without having to retrieve it once per task.
	settings, err := queries.GetSettings()
	discordWebhook := ""
	if err != nil {
		discordWebhook = settings.DiscordWebhook
	}
	event := Event{
		EventType: ProductEventType,
		ProductEvent: ProductEvent{
			DiscordWebhook: discordWebhook,
			Retailer:       retailer,
			MonitorID:      monitorID,
		},
	}
	switch retailer {
	case enums.Target:
		targetData, ok := data.(TargetStockData)
		if !ok {
			return
		}
		event.ProductEvent.TargetData = targetData
	case enums.Walmart:
		walmartData, ok := data.(WalmartStockData)
		if !ok {
			return
		}
		event.ProductEvent.WalmartData = walmartData
	}
	defer func() {
		recover()
	}()
	// Will panic if any channel is closed
	go func(event Event, channels EventChannelSlice) {
		for _, ch := range channels {
			ch <- event
		}
	}(event, eb.Subscribers)
	eb.RM.RUnlock()
}

var eventBus *EventBus

// InitEventBus initializes the singleton instance of the EventBus
func InitEventBus() {
	eventBus = &EventBus{Subscribers: EventChannelSlice{}}
}

// GetEventBus returns the singleton instance of the EventBus
func GetEventBus() *EventBus {
	return eventBus
}
