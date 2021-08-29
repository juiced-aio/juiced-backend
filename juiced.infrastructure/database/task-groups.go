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
		if tempTaskGroup.TaskIDsSerialized != "" {
			tempTaskGroup.TaskIDs = strings.Split(tempTaskGroup.TaskIDsSerialized, ",")
		}

		monitors := []*entities.BaseMonitor{}
		if tempTaskGroup.MonitorsSerialized != "" {
			err = json.Unmarshal([]byte(tempTaskGroup.MonitorsSerialized), &monitors)
			if err != nil {
				tempTaskGroup.Monitors = monitors
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
	if taskGroup.TaskIDsSerialized != "" {
		taskGroup.TaskIDs = strings.Split(taskGroup.TaskIDsSerialized, ",")
	}

	monitors := []*entities.BaseMonitor{}
	if taskGroup.MonitorsSerialized != "" {
		err = json.Unmarshal([]byte(taskGroup.MonitorsSerialized), &monitors)
		if err != nil {
			taskGroup.Monitors = monitors
		}
	}
	return taskGroup, nil
}

func CreateTaskGroup(taskGroup entities.TaskGroup) error {
	if database == nil {
		return &DatabaseNotInitializedError{}
	}

	statement, err := database.Preparex(`INSERT INTO taskGroups (groupID, name, retailer, taskIDsSerialized, monitorsSerialized, creationDate) VALUES (?, ?, ?, ?, ?, ?)`)
	if err != nil {
		return err
	}
	taskIDsSerialized := strings.Join(taskGroup.TaskIDs, ",")

	monitorsSerialized, err := json.Marshal(taskGroup.Monitors)
	if err != nil {
		return err
	}

	_, err = statement.Exec(taskGroup.GroupID, taskGroup.Name, taskGroup.Retailer, taskIDsSerialized, monitorsSerialized, taskGroup.CreationDate)
	return err
}

func RemoveTaskGroup(groupID string, deleteTasks bool) error {
	taskGroup := entities.TaskGroup{}
	if database == nil {
		return &DatabaseNotInitializedError{}
	}

	if deleteTasks {
		for _, taskID := range taskGroup.TaskIDs {
			err := RemoveTask(taskID)
			if err != nil {
				return err
			}
		}
	}

	statement, err := database.Preparex(`DELETE FROM taskGroups WHERE groupID = @p1`)
	if err != nil {
		return err
	}
	_, err = statement.Exec(groupID)
	if err != nil {
		return err
	}

	return err

}

func UpdateTaskGroup(groupID string, newTaskGroup entities.TaskGroup) error {
	err := RemoveTaskGroup(groupID, false)
	if err != nil {
		return err
	}
	return CreateTaskGroup(newTaskGroup)
}
