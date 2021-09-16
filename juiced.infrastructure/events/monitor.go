package events

import (
	"backend.juicedbot.io/juiced.infrastructure/enums"
)

// MonitorEvent is fired whenever a Monitor updates
type MonitorEvent struct {
	Status      enums.MonitorStatus    `json:"status"`
	EventType   enums.MonitorEventType `json:"eventType"`
	Data        interface{}            `json:"data"`
	TaskGroupID string                 `json:"taskGroupID"`
	MonitorID   string                 `json:"monitorID"`
}
