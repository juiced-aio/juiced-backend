package events

import (
	"backend.juicedbot.io/juiced.infrastructure/common/enums"
)

// MonitorEvent is fired whenever a Monitor updates
type MonitorEvent struct {
	Status    enums.MonitorStatus    `json:"status"`
	EventType enums.MonitorEventType `json:"eventType"`
	Data      interface{}            `json:"data"`
	MonitorID string                 `json:"monitorID"`
}
