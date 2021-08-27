package responses

import (
	"backend.juicedbot.io/juiced.infrastructure/entities"
)

// TaskGroupResponse is the response that any /api/task/group request receives
type TaskGroupResponse struct {
	Success  bool                 `json:"success"`
	Data     []entities.TaskGroup `json:"data"`
	Errors   []string             `json:"errors"`
	Warnings []string             `json:"warnings"`
}

// TaskResponse is the response that any /api/task request receives
type TaskResponse struct {
	Success bool            `json:"success"`
	Data    []entities.Task `json:"data"`
	Errors  []string        `json:"errors"`
}
