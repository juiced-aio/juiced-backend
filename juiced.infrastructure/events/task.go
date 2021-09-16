package events

import (
	"backend.juicedbot.io/juiced.infrastructure/enums"
)

// TaskEvent is fired whenever a Task updates
type TaskEvent struct {
	Status           enums.TaskStatus    `json:"status"`
	StatusPercentage int                 `json:"statusPercentage"`
	EventType        enums.TaskEventType `json:"eventType"`
	Data             interface{}         `json:"data"`
	TaskGroupID      string              `json:"taskGroupID"`
	TaskID           string              `json:"taskID"`
}
