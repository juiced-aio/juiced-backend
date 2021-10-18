package requests

import "backend.juicedbot.io/juiced.infrastructure/entities"

type TaskGroupsRequest struct {
	TaskGroupIDs []string `json:"taskGroupIDs"`
}

type UpdateTaskGroupsRequest struct {
	TaskGroupIDs   []string                `json:"taskGroupIDs"`
	Name           string                  `json:"name"`
	UpdateName     bool                    `json:"updateName"`
	Monitors       []*entities.BaseMonitor `json:"monitors"`
	UpdateMonitors bool                    `json:"updateMonitors"`
}
