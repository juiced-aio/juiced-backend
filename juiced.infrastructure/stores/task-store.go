package stores

import (
	"fmt"

	"backend.juicedbot.io/juiced.infrastructure/database"
	"backend.juicedbot.io/juiced.infrastructure/entities"
)

type TaskStore struct {
	Tasks map[string]*entities.Task
}

var taskStore TaskStore

func (store *TaskStore) Init() error {
	tasks, err := database.GetAllTasks()
	if err != nil {
		return err
	}

	for _, task := range tasks {
		store.Tasks[task.ID] = &task
	}

	return nil
}

type TaskNotFoundError struct {
	ID string
}

func (e *TaskNotFoundError) Error() string {
	return fmt.Sprintf("Task with ID %s not found", e.ID)
}

func GetAllTasks() []*entities.Task {
	tasks := []*entities.Task{}
	for _, task := range taskStore.Tasks {
		tasks = append(tasks, task)
	}

	return tasks
}

func GetTasks(taskIDs []string) []*entities.Task {
	tasks := []*entities.Task{}
	for _, taskID := range taskIDs {
		if task, ok := taskStore.Tasks[taskID]; ok {
			tasks = append(tasks, task)
		}
	}

	return tasks
}

func GetTask(taskID string) (*entities.Task, error) {
	task, ok := taskStore.Tasks[taskID]
	if !ok {
		return nil, &TaskNotFoundError{taskID}
	}

	return task, nil
}
