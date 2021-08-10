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
		if tempTaskGroup.TaskIDsJoined != "" {
			tempTaskGroup.TaskIDs = strings.Split(tempTaskGroup.TaskIDsJoined, ",")
		}
		tempTaskGroup, err = GetMonitorInfos(tempTaskGroup)
		if err != nil {
			return taskGroups, err
		}
		taskGroups = append(taskGroups, tempTaskGroup)
	}

	sort.SliceStable(taskGroups, func(i, j int) bool {
		return taskGroups[i].CreationDate < taskGroups[j].CreationDate
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
	if taskGroup.TaskIDsJoined != "" {
		taskGroup.TaskIDs = strings.Split(taskGroup.TaskIDsJoined, ",")
	}

	return GetMonitorInfos(taskGroup)
}

// GetTaskGroupByProxyGroupID returns the TaskGroup object from the database with the given proxyGroupID (if it exists)
func GetTaskGroupsByProxyGroupID(proxyGroupID string) ([]entities.TaskGroup, error) {
	taskGroups := make([]entities.TaskGroup, 0)
	database := common.GetDatabase()
	if database == nil {
		return taskGroups, errors.New("database not initialized")
	}

	statement, err := database.Preparex("SELECT * FROM taskGroups WHERE proxyGroupID = @p1")
	if err != nil {
		return taskGroups, err
	}

	rows, err := statement.Queryx(proxyGroupID)
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
		taskGroup, err := GetTaskGroup(tempTaskGroup.GroupID)
		if err != nil {
			return taskGroups, err
		}
		taskGroups = append(taskGroups, taskGroup)
	}

	return taskGroups, err
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
		if tempTask.TaskSizeJoined != "" {
			tempTask.TaskSize = strings.Split(tempTask.TaskSizeJoined, ",")
		}
		tempTask, err = GetTaskInfos(tempTask)
		if err != nil {
			return tasks, err
		}
		tasks = append(tasks, tempTask)
	}

	sort.SliceStable(tasks, func(i, j int) bool {
		return tasks[i].CreationDate < tasks[j].CreationDate
	})

	return tasks, err
}

// @silent: Do you want to change these two GetTasksBy functions into something like this?
//	tasks := []entities.Task{}
//	allTasks, err := GetAllTasks()
//	if err != nil {
//		return tasks, err
//	}
//	for _, task := range allTasks {
//		if task.TaskProfileID == profileID {
//			tasks = append(tasks, task)
//		}
//	}

// GetTasksByProfileID returns all tasks with the given profileID
func GetTasksByProfileID(profileID string) ([]entities.Task, error) {
	tasks := []entities.Task{}
	database := common.GetDatabase()
	if database == nil {
		return tasks, errors.New("database not initialized")
	}

	statement, err := database.Preparex("SELECT * FROM tasks WHERE profileID = @p1")
	if err != nil {
		return tasks, err
	}

	rows, err := statement.Queryx(profileID)
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
		if tempTask.TaskSizeJoined != "" {
			tempTask.TaskSize = strings.Split(tempTask.TaskSizeJoined, ",")
		}
		tempTask, err = GetTaskInfos(tempTask)
		if err != nil {
			return tasks, err
		}
		tasks = append(tasks, tempTask)
	}

	sort.SliceStable(tasks, func(i, j int) bool {
		return tasks[i].CreationDate < tasks[j].CreationDate
	})

	return tasks, err
}

// GetTasksByProxyGroupID returns all tasks with the given proxyGroupID
func GetTasksByProxyGroupID(proxyGroupID string) ([]entities.Task, error) {
	tasks := []entities.Task{}
	database := common.GetDatabase()
	if database == nil {
		return tasks, errors.New("database not initialized")
	}

	statement, err := database.Preparex("SELECT * FROM tasks WHERE proxyGroupID = @p1")
	if err != nil {
		return tasks, err
	}

	rows, err := statement.Queryx(proxyGroupID)
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
		if tempTask.TaskSizeJoined != "" {
			tempTask.TaskSize = strings.Split(tempTask.TaskSizeJoined, ",")
		}
		tempTask, err = GetTaskInfos(tempTask)
		if err != nil {
			return tasks, err
		}
		tasks = append(tasks, tempTask)
	}

	sort.SliceStable(tasks, func(i, j int) bool {
		return tasks[i].CreationDate < tasks[j].CreationDate
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
	if task.TaskSizeJoined != "" {
		task.TaskSize = strings.Split(task.TaskSizeJoined, ",")
	}

	return GetTaskInfos(task)
}

// ConvertTaskIDsToTasks returns a TaskGroupWithTasks object from a TaskGroup object
func ConvertTaskIDsToTasks(taskGroup *entities.TaskGroup) (entities.TaskGroupWithTasks, error) {
	taskGroupWithTasks := entities.TaskGroupWithTasks{
		GroupID: taskGroup.GroupID, Name: taskGroup.Name,
		MonitorProxyGroupID:  taskGroup.MonitorProxyGroupID,
		MonitorRetailer:      taskGroup.MonitorRetailer,
		MonitorDelay:         taskGroup.MonitorDelay,
		MonitorStatus:        taskGroup.MonitorStatus,
		AmazonMonitorInfo:    taskGroup.AmazonMonitorInfo,
		BestbuyMonitorInfo:   taskGroup.BestbuyMonitorInfo,
		BoxlunchMonitorInfo:  taskGroup.BoxlunchMonitorInfo,
		DisneyMonitorInfo:    taskGroup.DisneyMonitorInfo,
		GamestopMonitorInfo:  taskGroup.GamestopMonitorInfo,
		HottopicMonitorInfo:  taskGroup.HottopicMonitorInfo,
		ShopifyMonitorInfo:   taskGroup.ShopifyMonitorInfo,
		BigCartelMonitorInfo: taskGroup.BigCartelMonitorInfo,
		TargetMonitorInfo:    taskGroup.TargetMonitorInfo,
		WalmartMonitorInfo:   taskGroup.WalmartMonitorInfo,

		Tasks: []entities.Task{},
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
