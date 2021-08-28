package database

import (
	"encoding/json"
	"sort"
	"strings"

	"backend.juicedbot.io/juiced.infrastructure/entities"
)

func GetAllTaskGroups() ([]entities.TaskGroup, error) {
	taskGroups := []entities.TaskGroup{}
	if database == nil {
		return taskGroups, &DatabaseNotInitializedError{}
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

		monitor := entities.BaseMonitor{}
		if tempTaskGroup.MonitorInfo != "" {
			err = json.Unmarshal([]byte(tempTaskGroup.MonitorInfo), &monitor)
			if err != nil {
				tempTaskGroup.Monitor = &monitor
			}
		}

		taskGroups = append(taskGroups, tempTaskGroup)
	}

	sort.SliceStable(taskGroups, func(i, j int) bool {
		return taskGroups[i].CreationDate < taskGroups[j].CreationDate
	})

	return taskGroups, err
}

func GetTaskGroup(groupID string) (entities.TaskGroup, error) {
	taskGroup := entities.TaskGroup{}
	if database == nil {
		return taskGroup, &DatabaseNotInitializedError{}
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

	monitor := entities.BaseMonitor{}
	if taskGroup.MonitorInfo != "" {
		err = json.Unmarshal([]byte(taskGroup.MonitorInfo), &monitor)
		if err != nil {
			taskGroup.Monitor = &monitor
		}
	}
	return taskGroup, nil
}
