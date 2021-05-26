package events

type EventType string

const (
	AuthEventType    EventType = "AUTH_EVENT"
	CloseEventType   EventType = "CLOSE_EVENT"
	TaskEventType    EventType = "TASK_EVENT"
	MonitorEventType EventType = "MONITOR_EVENT"
	ProductEventType EventType = "PRODUCT_EVENT"
)

// Event is any event that needs to be broadcasted
type Event struct {
	EventType    EventType    `json:"eventType"`
	CloseEvent   CloseEvent   `json:"closeEvent"`
	AuthEvent    AuthEvent    `json:"authEvent"`
	TaskEvent    TaskEvent    `json:"taskEvent"`
	MonitorEvent MonitorEvent `json:"monitorEvent"`
	ProductEvent ProductEvent `json:"productEvent"`
}

// EventChannel is a channel that can accept an Event
type EventChannel chan Event

// EventChannelSlice is a slice of TaskChannels
type EventChannelSlice []EventChannel
