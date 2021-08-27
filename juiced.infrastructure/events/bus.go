package events

import (
	"backend.juicedbot.io/juiced.infrastructure/enums"

	"sync"
)

// EventBus stores the information about subscribers
type EventBus struct {
	Subscribers []EventChannel
	RM          sync.RWMutex
}

// Unsubscribe removes the EventChannel from the EventBus's subscribers
func (eb *EventBus) Unsubscribe(ch EventChannel) {
	eb.RM.Lock()
	newSubscribers := []EventChannel{}
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

// PublishConnectEvent publishes a ConnectEvent
func (eb *EventBus) PublishConnectEvent() {
	eb.RM.RLock()
	// Will panic if any channel is closed
	go func(event Event, channels []EventChannel) {
		defer func() {
			recover()
		}()
		for _, ch := range channels {
			ch <- event
		}
	}(Event{
		EventType:    ConnectEventType,
		ConnectEvent: ConnectEvent{},
	}, eb.Subscribers)
	eb.RM.RUnlock()
}

// PublishAuthEvent publishes an AuthEvent
func (eb *EventBus) PublishAuthEvent() {
	eb.RM.RLock()
	// Will panic if any channel is closed
	go func(event Event, channels []EventChannel) {
		defer func() {
			recover()
		}()
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
	// Will panic if any channel is closed
	go func(event Event, channels []EventChannel) {
		defer func() {
			recover()
		}()
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
func (eb *EventBus) PublishMonitorEvent(monitorStatus enums.MonitorStatus, eventType enums.MonitorEventType, data interface{}, monitorID string) {
	eb.RM.RLock()
	// Will panic if any channel is closed
	go func(event Event, channels []EventChannel) {
		defer func() {
			recover()
		}()
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
func (eb *EventBus) PublishTaskEvent(taskStatus enums.TaskStatus, eventType enums.TaskEventType, data interface{}, taskID string) {
	eb.RM.RLock()
	// Will panic if any channel is closed
	go func(event Event, channels []EventChannel) {
		defer func() {
			recover()
		}()
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

var eventBus *EventBus

// InitEventBus initializes the singleton instance of the EventBus
func InitEventBus() {
	eventBus = &EventBus{Subscribers: []EventChannel{}}
}

// GetEventBus returns the singleton instance of the EventBus
func GetEventBus() *EventBus {
	return eventBus
}
