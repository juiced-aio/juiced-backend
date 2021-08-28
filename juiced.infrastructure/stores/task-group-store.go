package stores

import (
	"fmt"

	"backend.juicedbot.io/juiced.infrastructure/database"
	"backend.juicedbot.io/juiced.infrastructure/entities"
)

type TaskGroupStore struct {
	TaskGroups map[string]*entities.TaskGroup
}

var taskGroupStore TaskGroupStore

func (store *TaskGroupStore) Init() error {
	taskGroups, err := database.GetAllTaskGroups()
	if err != nil {
		return err
	}

	for _, taskGroup := range taskGroups {
		store.TaskGroups[taskGroup.GroupID] = &taskGroups
	}

	return nil
}

type TaskGroupNotFoundError struct {
	ID string
}

func (e *TaskGroupNotFoundError) Error() string {
	return fmt.Sprintf("TaskGroup with ID %s not found", e.ID)
}

func GetAllTaskGroups() []*entities.TaskGroup {
	taskGroups := []*entities.TaskGroup{}

	for _, taskGroup := range taskGroupStore.TaskGroups {
		taskGroups = append(taskGroups, taskGroup)
	}

	return taskGroups
}

func GetTaskGroup(groupID string) (*entities.TaskGroup, error) {
	taskGroup, ok := taskGroupStore.TaskGroups[groupID]
	if !ok {
		return nil, &TaskGroupNotFoundError{groupID}
	}

	return taskGroup, nil
}
