package database

import (
	"encoding/json"
	"sort"

	"backend.juicedbot.io/juiced.infrastructure/entities"
)

func GetAllTasks() ([]entities.Task, error) {
	tasks := []entities.Task{}
	if database == nil {
		return tasks, &DatabaseNotInitializedError{}
	}

	rows, err := database.Queryx("SELECT * FROM tasks")
	if err != nil {
		return tasks, err
	}

	defer rows.Close()
	for rows.Next() {
		tempTask := entities.Task{}
		err = rows.StructScan(&tempTask)
		if err != nil {
			return tasks, err
		}

		baseTask := entities.BaseTask{}
		if tempTask.TaskInfo != "" {
			err = json.Unmarshal([]byte(tempTask.TaskInfo), &baseTask)
			if err != nil {
				tempTask.Task = &baseTask
			}
		}

		tasks = append(tasks, tempTask)
	}

	sort.SliceStable(tasks, func(i, j int) bool {
		return tasks[i].CreationDate < tasks[j].CreationDate
	})

	return tasks, err
}

func GetTask(ID string) (entities.Task, error) {
	task := entities.Task{}
	if database == nil {
		return task, &DatabaseNotInitializedError{}
	}

	statement, err := database.Preparex("SELECT * FROM tasks WHERE ID = @p1")
	if err != nil {
		return task, err
	}

	rows, err := statement.Queryx(ID)
	if err != nil {
		return task, err
	}

	defer rows.Close()
	for rows.Next() {
		err = rows.StructScan(&task)
		if err != nil {
			return task, err
		}
	}

	baseTask := entities.BaseTask{}
	if task.TaskInfo != "" {
		err = json.Unmarshal([]byte(task.TaskInfo), &task)
		if err != nil {
			task.Task = &baseTask
		}
	}

	return task, nil
}

func CreateTask(task entities.Task) error {
	if database == nil {
		return &DatabaseNotInitializedError{}
	}

	statement, err := database.Preparex(`INSERT INTO tasks (ID, taskGroupID, profileID, proxyGroupID, taskInfo, creationDate) VALUES (?, ?, ?, ?, ?, ?)`)
	if err != nil {
		return err
	}

	taskInfo, err := json.Marshal(task.Task)
	if err != nil {
		return err
	}

	_, err = statement.Exec(task.ID, task.TaskGroupID, task.Profile, task.ProxyGroup, taskInfo, task.CreationDate)
	return err
}

func RemoveTask(ID string) error {
	if database == nil {
		return &DatabaseNotInitializedError{}
	}

	statement, err := database.Preparex(`DELETE FROM tasks WHERE ID = @p1`)
	if err != nil {
		return err
	}
	_, err = statement.Exec(ID)
	return err
}

func UpdateTask(ID string, newTask entities.Task) error {
	err := RemoveTask(ID)
	if err != nil {
		return err
	}
	return CreateTask(newTask)
}
