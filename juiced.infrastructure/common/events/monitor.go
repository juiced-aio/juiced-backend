package events

import (
	"backend.juicedbot.io/m/v2/juiced.infrastructure/common/enums"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// MonitorEvent is fired whenever a Monitor updates
type MonitorEvent struct {
	Status    enums.MonitorStatus    `json:"status"`
	EventType enums.MonitorEventType `json:"eventType"`
	Data      interface{}            `json:"data"`
	MonitorID primitive.ObjectID     `json:"monitorID"`
}
