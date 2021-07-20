package queries

import (
	"errors"
	"strings"

	"backend.juicedbot.io/juiced.infrastructure/common"
	"backend.juicedbot.io/juiced.infrastructure/common/entities"
	"backend.juicedbot.io/juiced.infrastructure/common/enums"
)

func GetTaskInfos(task entities.Task) (entities.Task, error) {
	database := common.GetDatabase()
	if database == nil {
		return task, errors.New("database not initialized")
	}

	switch task.TaskRetailer {
	case enums.Target:
		statement, err := database.Preparex(`SELECT * FROM targetTaskInfos WHERE taskID = @p1`)
		if err != nil {
			return task, err
		}
		rows, err := statement.Queryx(task.ID)
		if err != nil {
			return task, err
		}

		defer rows.Close()
		for rows.Next() {
			err = rows.StructScan(&task.TargetTaskInfo)
			if err != nil {
				return task, err
			}
		}
	case enums.Walmart:
		statement, err := database.Preparex(`SELECT * FROM walmartTaskInfos WHERE taskID = @p1`)
		if err != nil {
			return task, err
		}
		rows, err := statement.Queryx(task.ID)
		if err != nil {
			return task, err
		}

		defer rows.Close()
		for rows.Next() {
			err = rows.StructScan(&task.WalmartTaskInfo)
			if err != nil {
				return task, err
			}
		}
	case enums.Amazon:
		statement, err := database.Preparex(`SELECT * FROM amazonTaskInfos WHERE taskID = @p1`)
		if err != nil {
			return task, err
		}
		rows, err := statement.Queryx(task.ID)
		if err != nil {
			return task, err
		}

		defer rows.Close()
		for rows.Next() {
			err = rows.StructScan(&task.AmazonTaskInfo)
			if err != nil {
				return task, err
			}
		}
	case enums.BestBuy:
		statement, err := database.Preparex(`SELECT * FROM bestbuyTaskInfos WHERE taskID = @p1`)
		if err != nil {
			return task, err
		}
		rows, err := statement.Queryx(task.ID)
		if err != nil {
			return task, err
		}

		defer rows.Close()
		for rows.Next() {
			err = rows.StructScan(&task.BestbuyTaskInfo)
			if err != nil {
				return task, err
			}
		}

	case enums.BoxLunch:
		statement, err := database.Preparex(`SELECT * FROM boxlunchTaskInfos WHERE taskID = @p1`)
		if err != nil {
			return task, err
		}
		rows, err := statement.Queryx(task.ID)
		if err != nil {
			return task, err
		}

		defer rows.Close()
		for rows.Next() {
			err = rows.StructScan(&task.BoxLunchTaskInfo)
			if err != nil {
				return task, err
			}
		}
		if task.BoxLunchTaskInfo.PidsJoined != "" {
			task.BoxLunchTaskInfo.Pids = strings.Split(task.BoxLunchTaskInfo.PidsJoined, ",")
		}

	case enums.GameStop:
		statement, err := database.Preparex(`SELECT * FROM gamestopTaskInfos WHERE taskID = @p1`)
		if err != nil {
			return task, err
		}
		rows, err := statement.Queryx(task.ID)
		if err != nil {
			return task, err
		}

		defer rows.Close()
		for rows.Next() {
			err = rows.StructScan(&task.GamestopTaskInfo)
			if err != nil {
				return task, err
			}
		}
	}
	return task, nil
}

func GetMonitorInfos(taskGroup entities.TaskGroup) (entities.TaskGroup, error) {
	database := common.GetDatabase()
	if database == nil {
		return taskGroup, errors.New("database not initialized")
	}
	switch taskGroup.MonitorRetailer {
	case enums.Target:
		statement, err := database.Preparex(`SELECT * FROM targetMonitorInfos WHERE taskGroupID = @p1`)
		if err != nil {
			return taskGroup, err
		}
		rows, err := statement.Queryx(taskGroup.GroupID)
		if err != nil {
			return taskGroup, err
		}

		defer rows.Close()
		for rows.Next() {
			err = rows.StructScan(&taskGroup.TargetMonitorInfo)
			if err != nil {
				return taskGroup, err
			}
		}

		statement, err = database.Preparex(`SELECT * FROM targetSingleMonitorInfos WHERE monitorID = @p1`)
		if err != nil {
			return taskGroup, err
		}

		rows, err = statement.Queryx(taskGroup.TargetMonitorInfo.ID)
		if err != nil {
			return taskGroup, err
		}

		defer rows.Close()
		for rows.Next() {
			tempSingleMonitor := entities.TargetSingleMonitorInfo{}
			err = rows.StructScan(&tempSingleMonitor)
			if err != nil {
				return taskGroup, err
			}
			taskGroup.TargetMonitorInfo.Monitors = append(taskGroup.TargetMonitorInfo.Monitors, tempSingleMonitor)
		}

	case enums.Walmart:
		statement, err := database.Preparex(`SELECT * FROM walmartMonitorInfos WHERE taskGroupID = @p1`)
		if err != nil {
			return taskGroup, err
		}

		rows, err := statement.Queryx(taskGroup.GroupID)
		if err != nil {
			return taskGroup, err
		}

		defer rows.Close()
		for rows.Next() {
			err = rows.StructScan(&taskGroup.WalmartMonitorInfo)
			if err != nil {
				return taskGroup, err
			}
		}

		if taskGroup.WalmartMonitorInfo.SKUsJoined != "" {
			taskGroup.WalmartMonitorInfo.SKUs = strings.Split(taskGroup.WalmartMonitorInfo.SKUsJoined, ",")
		}

	case enums.Amazon:
		statement, err := database.Preparex(`SELECT * FROM amazonMonitorInfos WHERE taskGroupID = @p1`)
		if err != nil {
			return taskGroup, err
		}
		rows, err := statement.Queryx(taskGroup.GroupID)
		if err != nil {
			return taskGroup, err
		}

		defer rows.Close()
		for rows.Next() {
			err = rows.StructScan(&taskGroup.AmazonMonitorInfo)
			if err != nil {
				return taskGroup, err
			}
		}

		statement, err = database.Preparex(`SELECT * FROM amazonSingleMonitorInfos WHERE monitorID = @p1`)
		if err != nil {
			return taskGroup, err
		}

		rows, err = statement.Queryx(taskGroup.AmazonMonitorInfo.ID)
		if err != nil {
			return taskGroup, err
		}
		defer rows.Close()

		for rows.Next() {
			tempSingleMonitor := entities.AmazonSingleMonitorInfo{}
			err = rows.StructScan(&tempSingleMonitor)
			if err != nil {
				return taskGroup, err
			}
			taskGroup.AmazonMonitorInfo.Monitors = append(taskGroup.AmazonMonitorInfo.Monitors, tempSingleMonitor)
		}

	case enums.BestBuy:
		statement, err := database.Preparex(`SELECT * FROM bestbuyMonitorInfos WHERE taskGroupID = @p1`)
		if err != nil {
			return taskGroup, err
		}

		rows, err := statement.Queryx(taskGroup.GroupID)
		if err != nil {
			return taskGroup, err
		}

		defer rows.Close()
		for rows.Next() {
			err = rows.StructScan(&taskGroup.BestbuyMonitorInfo)
			if err != nil {
				return taskGroup, err
			}
		}

		statement, err = database.Preparex(`SELECT * FROM bestbuySingleMonitorInfos WHERE monitorID = @p1`)
		if err != nil {
			return taskGroup, err
		}

		rows, err = statement.Queryx(taskGroup.BestbuyMonitorInfo.ID)
		if err != nil {
			return taskGroup, err
		}

		defer rows.Close()
		for rows.Next() {
			tempSingleMonitor := entities.BestbuySingleMonitorInfo{}
			err = rows.StructScan(&tempSingleMonitor)
			if err != nil {
				return taskGroup, err
			}
			taskGroup.BestbuyMonitorInfo.Monitors = append(taskGroup.BestbuyMonitorInfo.Monitors, tempSingleMonitor)
		}

	case enums.BoxLunch:
		statement, err := database.Preparex(`SELECT * FROM boxlunchMonitorInfos WHERE taskGroupID = @p1`)
		if err != nil {
			return taskGroup, err
		}

		rows, err := statement.Queryx(taskGroup.GroupID)
		if err != nil {
			return taskGroup, err
		}

		defer rows.Close()
		for rows.Next() {
			err = rows.StructScan(&taskGroup.BoxLunchMonitorInfo)
			if err != nil {
				return taskGroup, err
			}
		}

		statement, err = database.Preparex(`SELECT * FROM boxlunchSingleMonitorInfos WHERE monitorID = @p1`)
		if err != nil {
			return taskGroup, err
		}

		rows, err = statement.Queryx(taskGroup.BoxLunchMonitorInfo.ID)
		if err != nil {
			return taskGroup, err
		}

		defer rows.Close()
		for rows.Next() {
			tempSingleMonitor := entities.BoxLunchSingleMonitorInfo{}
			err = rows.StructScan(&tempSingleMonitor)
			if err != nil {
				return taskGroup, err
			}
			taskGroup.BoxLunchMonitorInfo.Monitors = append(taskGroup.BoxLunchMonitorInfo.Monitors, tempSingleMonitor)
		}

	case enums.GameStop:
		statement, err := database.Preparex(`SELECT * FROM gamestopMonitorInfos WHERE taskGroupID = @p1`)
		if err != nil {
			return taskGroup, err
		}

		rows, err := statement.Queryx(taskGroup.GroupID)
		if err != nil {
			return taskGroup, err
		}

		defer rows.Close()
		for rows.Next() {
			err = rows.StructScan(&taskGroup.GamestopMonitorInfo)
			if err != nil {
				return taskGroup, err
			}
		}
		statement, err = database.Preparex(`SELECT * FROM gamestopSingleMonitorInfos WHERE monitorID = @p1`)
		if err != nil {
			return taskGroup, err
		}

		rows, err = statement.Queryx(taskGroup.GamestopMonitorInfo.ID)
		if err != nil {
			return taskGroup, err
		}
		defer rows.Close()

		for rows.Next() {
			tempSingleMonitor := entities.GamestopSingleMonitorInfo{}
			err = rows.StructScan(&tempSingleMonitor)
			if err != nil {
				return taskGroup, err
			}
			taskGroup.GamestopMonitorInfo.Monitors = append(taskGroup.GamestopMonitorInfo.Monitors, tempSingleMonitor)
		}

	}

	return taskGroup, nil
}

func GetProxies(proxyGroup entities.ProxyGroup) (entities.ProxyGroup, error) {
	database := common.GetDatabase()
	if database == nil {
		return proxyGroup, errors.New("database not initialized")
	}

	statement, err := database.Preparex(`SELECT * FROM proxys WHERE proxyGroupID = @p1`)
	if err != nil {
		return proxyGroup, err
	}
	rows, err := statement.Queryx(proxyGroup.GroupID)
	if err != nil {
		return proxyGroup, err
	}

	defer rows.Close()
	for rows.Next() {
		tempProxy := entities.Proxy{}
		err = rows.StructScan(&tempProxy)
		if err != nil {
			return proxyGroup, err
		}
		proxyGroup.Proxies = append(proxyGroup.Proxies, tempProxy)
	}

	return proxyGroup, err
}

func GetShippingAddress(profile entities.Profile) (entities.Profile, error) {
	database := common.GetDatabase()
	if database == nil {
		return profile, errors.New("database not initialized")
	}

	statement, err := database.Preparex(`SELECT * FROM shippingAddresses WHERE profileID = @p1`)
	if err != nil {
		return profile, err
	}
	rows, err := statement.Queryx(profile.ID)
	if err != nil {
		return profile, err
	}

	defer rows.Close()
	for rows.Next() {
		err = rows.StructScan(&profile.ShippingAddress)
		if err != nil {
			return profile, err
		}

	}

	return profile, err
}

func GetBillingAddress(profile entities.Profile) (entities.Profile, error) {
	database := common.GetDatabase()
	if database == nil {
		return profile, errors.New("database not initialized")
	}

	statement, err := database.Preparex(`SELECT * FROM shippingAddresses WHERE profileID = @p1`)
	if err != nil {
		return profile, err
	}
	rows, err := statement.Queryx(profile.ID)
	if err != nil {
		return profile, err
	}

	defer rows.Close()
	for rows.Next() {
		err = rows.StructScan(&profile.BillingAddress)
		if err != nil {
			return profile, err
		}

	}

	return profile, err
}

func GetCard(profile entities.Profile) (entities.Profile, error) {
	database := common.GetDatabase()
	if database == nil {
		return profile, errors.New("database not initialized")
	}

	statement, err := database.Preparex(`SELECT * FROM cards WHERE profileID = @p1`)
	if err != nil {
		return profile, err
	}
	rows, err := statement.Queryx(profile.ID)
	if err != nil {
		return profile, err
	}

	defer rows.Close()
	for rows.Next() {
		err = rows.StructScan(&profile.CreditCard)
		if err != nil {
			return profile, err
		}

	}

	return profile, err
}

func GetProfileInfo(profile entities.Profile) (entities.Profile, error) {
	if profile.ProfileGroupIDsJoined != "" {
		profile.ProfileGroupIDs = strings.Split(profile.ProfileGroupIDsJoined, ",")
	}
	profile, err := GetShippingAddress(profile)
	if err != nil {
		return profile, err
	}
	profile, err = GetBillingAddress(profile)
	if err != nil {
		return profile, err
	}
	return GetCard(profile)
}
