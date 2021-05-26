package events

import (
	"juiced/juiced.infrastructure/common/enums"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// TaskEvent is fired whenever a Task updates
type TaskEvent struct {
	Status    enums.TaskStatus    `json:"status"`
	EventType enums.TaskEventType `json:"eventType"`
	Data      interface{}         `json:"data"`
	TaskID    primitive.ObjectID  `json:"taskID"`
}
