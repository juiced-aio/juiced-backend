package events

import (
	"backend.juicedbot.io/juiced.infrastructure/common/enums"
)

// TaskEvent is fired whenever a Task updates
type TaskEvent struct {
	Status           enums.TaskStatus    `json:"status"`
	StatusPercentage int                 `json:"statusPercentage"`
	EventType        enums.TaskEventType `json:"eventType"`
	Data             interface{}         `json:"data"`
	TaskID           string              `json:"taskID"`
}

// ProductInfo is sent when publishing the SendingProductInfoToTasks event
type ProductInfo struct {
	Products []Product `json:"products"`
}

type Product struct {
	ProductName     string `json:"productName"`
	ProductImageURL string `json:"productImageURL"`
}
