package commands

import (
	"errors"
	"strings"

	"backend.juicedbot.io/juiced.infrastructure/common"
	"backend.juicedbot.io/juiced.infrastructure/common/entities"
	"backend.juicedbot.io/juiced.infrastructure/queries"
	_ "github.com/jmoiron/sqlx"
	_ "github.com/mattn/go-sqlite3"
)

// CreateTaskGroup adds the TaskGroup object to the database
func CreateTaskGroup(taskGroup entities.TaskGroup) error {
	database := common.GetDatabase()
	if database == nil {
		return errors.New("database not initialized")
	}

	statement, err := database.Preparex(`INSERT INTO taskGroups (groupID, name, proxyGroupID, retailer, input, delay, status, taskIDsJoined, creationDate) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`)
	if err != nil {
		return err
	}
	taskIDsJoined := strings.Join(taskGroup.TaskIDs, ",")

	_, err = statement.Exec(taskGroup.GroupID, taskGroup.Name, taskGroup.MonitorProxyGroupID, taskGroup.MonitorRetailer, taskGroup.MonitorInput, taskGroup.MonitorDelay, taskGroup.MonitorStatus, taskIDsJoined, taskGroup.CreationDate)
	if err != nil {
		return err
	}

	return CreateMonitorInfos(taskGroup)
}

// RemoveTaskGroup removes the TaskGroup from the database with the given groupID and returns it (if it exists)
func RemoveTaskGroup(groupID string, deleteTasks bool) (entities.TaskGroup, error) {
	taskGroup := entities.TaskGroup{}
	database := common.GetDatabase()
	if database == nil {
		return taskGroup, errors.New("database not initialized")
	}

	taskGroup, err := queries.GetTaskGroup(groupID)
	if err != nil {
		return taskGroup, err
	}

	if deleteTasks {
		for _, taskID := range taskGroup.TaskIDs {
			_, err := RemoveTask(taskID)
			if err != nil {
				return taskGroup, err
			}
		}
	}

	statement, err := database.Preparex(`DELETE FROM taskGroups WHERE groupID = @p1`)
	if err != nil {
		return taskGroup, err
	}
	_, err = statement.Exec(groupID)
	if err != nil {
		return taskGroup, err
	}

	err = DeleteMonitorInfos(groupID, taskGroup.MonitorRetailer)
	return taskGroup, err

}

// UpdateTaskGroup updates the TaskGroup from the database with the given groupID and returns it (if it exists)
func UpdateTaskGroup(groupID string, newTaskGroup entities.TaskGroup) (entities.TaskGroup, error) {
	taskGroup, err := RemoveTaskGroup(groupID, false)
	if err != nil {
		return taskGroup, err
	}

	err = CreateTaskGroup(newTaskGroup)
	if err != nil {
		return taskGroup, err
	}

	return queries.GetTaskGroup(groupID)
}

// CreateTask adds the Task object to the database
func CreateTask(task entities.Task) error {
	database := common.GetDatabase()
	if database == nil {
		return errors.New("database not initialized")
	}

	statement, err := database.Preparex(`INSERT INTO tasks (ID, taskGroupID, profileID, proxyGroupID, retailer, sizeJoined, qty, status, taskDelay, creationDate) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`)
	if err != nil {
		return err
	}

	sizeJoined := strings.Join(task.TaskSize, ",")
	_, err = statement.Exec(task.ID, task.TaskGroupID, task.TaskProfileID, task.TaskProxyGroupID, task.TaskRetailer, sizeJoined, task.TaskQty, task.TaskStatus, task.TaskDelay, task.CreationDate)
	if err != nil {
		return err
	}

	return CreateTaskInfos(task)

}

// RemoveTask removes the Task from the database with the given ID and returns it (if it exists)
func RemoveTask(ID string) (entities.Task, error) {
	task := entities.Task{}
	database := common.GetDatabase()
	if database == nil {
		return task, errors.New("database not initialized")
	}
	task, err := queries.GetTask(ID)
	if err != nil {
		return task, err
	}

	statement, err := database.Preparex(`DELETE FROM tasks WHERE ID = @p1`)
	if err != nil {
		return task, err
	}
	_, err = statement.Exec(ID)
	if err != nil {
		return task, err
	}

	if task.TaskSizeJoined != "" {
		task.TaskSize = strings.Split(task.TaskSizeJoined, ",")
	}

	return task, DeleteTaskInfos(task.ID, task.TaskRetailer)
}

// UpdateTask updates the Task from the database with the given ID and returns it (if it exists)
func UpdateTask(ID string, newTask entities.Task) (entities.Task, error) {
	// @silent: Good or bad practice?
	task := entities.Task{}
	_, err := RemoveTask(ID)
	if err != nil {
		return task, err
	}

	err = CreateTask(newTask)
	if err != nil {
		return task, err
	}

	return task, err
}

// RemoveTasksWithProfileID removes any Tasks with the given profileID and returns any errors
func RemoveTasksByProfileID(profileID string) error {
	database := common.GetDatabase()
	if database == nil {
		return errors.New("database not initialized")
	}

	tasks, err := queries.GetTasksByProfileID(profileID)
	if err != nil {
		return err
	}

	for _, task := range tasks {
		_, err = RemoveTask(task.ID)
		if err != nil {
			return err
		}
	}

	return err
}
