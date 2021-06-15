package queries

import (
	"errors"
	"sort"
	"strings"

	"backend.juicedbot.io/juiced.infrastructure/common"
	"backend.juicedbot.io/juiced.infrastructure/common/entities"
)

// GetAllTaskGroups returns all TaskGroup objects from the database
func GetAllTaskGroups() ([]entities.TaskGroup, error) {
	taskGroups := []entities.TaskGroup{}
	database := common.GetDatabase()
	if database == nil {
		return taskGroups, errors.New("database not initialized")
	}

	rows, err := database.Queryx("SELECT * FROM taskGroups")
	if err != nil {
		return taskGroups, err
	}

	defer rows.Close()
	for rows.Next() {
		tempTaskGroup := entities.TaskGroup{}
		err = rows.StructScan(&tempTaskGroup)
		if err != nil {
			return taskGroups, err
		}
		tempTaskGroup.TaskIDs = strings.Split(tempTaskGroup.TaskIDsJoined, ",")
		tempTaskGroup, err = GetMonitorInfos(tempTaskGroup)
		if err != nil {
			return taskGroups, err
		}
		taskGroups = append(taskGroups, tempTaskGroup)
	}

	sort.SliceStable(taskGroups, func(i, j int) bool {
		return taskGroups[i].CreationDate > taskGroups[j].CreationDate
	})

	return taskGroups, err
}

// GetTaskGroup returns the TaskGroup object from the database with the given groupID (if it exists)
func GetTaskGroup(groupID string) (entities.TaskGroup, error) {
	taskGroup := entities.TaskGroup{}
	database := common.GetDatabase()
	if database == nil {
		return taskGroup, errors.New("database not initialized")
	}

	statement, err := database.Preparex("SELECT * FROM taskGroups WHERE groupID = @p1")
	if err != nil {
		return taskGroup, err
	}

	rows, err := statement.Queryx(groupID)
	if err != nil {
		return taskGroup, err
	}

	defer rows.Close()
	for rows.Next() {
		err = rows.StructScan(&taskGroup)
		if err != nil {
			return taskGroup, err
		}
	}

	taskGroup.TaskIDs = strings.Split(taskGroup.TaskIDsJoined, ",")

	return GetMonitorInfos(taskGroup)
}

// GetAllTasks returns all Task objects from the database
func GetAllTasks() ([]entities.Task, error) {
	tasks := []entities.Task{}
	database := common.GetDatabase()
	if database == nil {
		return tasks, errors.New("database not initialized")
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
		tempTask.TaskSize = strings.Split(tempTask.TaskSizeJoined, ",")
		tempTask, err = GetTaskInfos(tempTask)
		if err != nil {
			return tasks, err
		}
		tasks = append(tasks, tempTask)
	}

	sort.SliceStable(tasks, func(i, j int) bool {
		return tasks[i].CreationDate > tasks[j].CreationDate
	})

	return tasks, err
}

// GetTask returns the Task object from the database with the given ID (if it exists)
func GetTask(ID string) (entities.Task, error) {
	task := entities.Task{}
	database := common.GetDatabase()
	if database == nil {
		return task, errors.New("database not initialized")
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

	task.TaskSize = strings.Split(task.TaskSizeJoined, ",")

	return GetTaskInfos(task)
}

// ConvertTaskIDsToTasks returns a TaskGroupWithTasks object from a TaskGroup object
func ConvertTaskIDsToTasks(taskGroup *entities.TaskGroup) (entities.TaskGroupWithTasks, error) {
	taskGroupWithTasks := entities.TaskGroupWithTasks{
		GroupID: taskGroup.GroupID, Name: taskGroup.Name,
		MonitorProxyGroupID: taskGroup.MonitorProxyGroupID,
		MonitorRetailer:     taskGroup.MonitorRetailer,
		MonitorDelay:        taskGroup.MonitorDelay,
		MonitorStatus:       taskGroup.MonitorStatus,
		TargetMonitorInfo:   taskGroup.TargetMonitorInfo,
		WalmartMonitorInfo:  taskGroup.WalmartMonitorInfo,
		AmazonMonitorInfo:   taskGroup.AmazonMonitorInfo,
		BestbuyMonitorInfo:  taskGroup.BestbuyMonitorInfo,
		HottopicMonitorInfo: taskGroup.HottopicMonitorInfo,
		GamestopMonitorInfo: taskGroup.GamestopMonitorInfo,
		Tasks:               []entities.Task{},
	}
	tasks := []entities.Task{}
	for i := 0; i < len(taskGroup.TaskIDs); i++ {
		task, err := GetTask(taskGroup.TaskIDs[i])
		if err != nil {
			return taskGroupWithTasks, err
		}
		tasks = append(tasks, task)
	}
	taskGroupWithTasks.SetTasks(tasks)
	return taskGroupWithTasks, nil
}
