package events

type EventType = string

const (
	ConnectEventType EventType = "CONNECT_EVENT"
	AuthEventType    EventType = "AUTH_EVENT"
	CloseEventType   EventType = "CLOSE_EVENT"
	TaskEventType    EventType = "TASK_EVENT"
	MonitorEventType EventType = "MONITOR_EVENT"
	AccountEventType EventType = "ACCOUNT_EVENT"
)

// Event is any event that needs to be broadcasted
type Event struct {
	EventType    EventType    `json:"eventType"`
	ConnectEvent ConnectEvent `json:"connectEvent"`
	CloseEvent   CloseEvent   `json:"closeEvent"`
	AuthEvent    AuthEvent    `json:"authEvent"`
	TaskEvent    TaskEvent    `json:"taskEvent"`
	MonitorEvent MonitorEvent `json:"monitorEvent"`
	AccountEvent AccountEvent `json:"accountEvent"`
}

// EventChannel is a channel that can accept an Event
type EventChannel chan Event
