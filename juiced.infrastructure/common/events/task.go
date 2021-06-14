package events

import (
	"backend.juicedbot.io/juiced.infrastructure/common/enums"
)

// TaskEvent is fired whenever a Task updates
type TaskEvent struct {
	Status    enums.TaskStatus    `json:"status"`
	EventType enums.TaskEventType `json:"eventType"`
	Data      interface{}         `json:"data"`
	TaskID    string              `json:"taskID"`
}
