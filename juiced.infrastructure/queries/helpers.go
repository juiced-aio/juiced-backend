package queries

import (
	"errors"
	"fmt"
	"strings"
	"time"

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
			tempTaskInfo := entities.AmazonTaskInfo{}
			err = rows.StructScan(&tempTaskInfo)
			if err != nil {
				return task, err
			}
			task.AmazonTaskInfo = &tempTaskInfo
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
			tempTaskInfo := entities.BestbuyTaskInfo{}
			err = rows.StructScan(&tempTaskInfo)
			if err != nil {
				return task, err
			}
			task.BestbuyTaskInfo = &tempTaskInfo
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
			tempTaskInfo := entities.BoxlunchTaskInfo{}
			err = rows.StructScan(&tempTaskInfo)
			if err != nil {
				return task, err
			}
			task.BoxlunchTaskInfo = &tempTaskInfo
		}

	case enums.Disney:
		statement, err := database.Preparex(`SELECT * FROM disneyTaskInfos WHERE taskID = @p1`)
		if err != nil {
			return task, err
		}
		rows, err := statement.Queryx(task.ID)
		if err != nil {
			return task, err
		}

		defer rows.Close()
		for rows.Next() {
			tempTaskInfo := entities.DisneyTaskInfo{}
			err = rows.StructScan(&tempTaskInfo)
			if err != nil {
				return task, err
			}
			task.DisneyTaskInfo = &tempTaskInfo
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
			tempTaskInfo := entities.GamestopTaskInfo{}
			err = rows.StructScan(&tempTaskInfo)
			if err != nil {
				return task, err
			}
			task.GamestopTaskInfo = &tempTaskInfo
		}

	case enums.HotTopic:
		statement, err := database.Preparex(`SELECT * FROM hottopicTaskInfos WHERE taskID = @p1`)
		if err != nil {
			return task, err
		}
		rows, err := statement.Queryx(task.ID)
		if err != nil {
			return task, err
		}

		defer rows.Close()
		for rows.Next() {
			tempTaskInfo := entities.HottopicTaskInfo{}
			err = rows.StructScan(&tempTaskInfo)
			if err != nil {
				return task, err
			}
			task.HottopicTaskInfo = &tempTaskInfo
		}

	case enums.Shopify:
		statement, err := database.Preparex(`SELECT * FROM shopifyTaskInfos WHERE taskID = @p1`)
		if err != nil {
			return task, err
		}
		rows, err := statement.Queryx(task.ID)
		if err != nil {
			return task, err
		}

		defer rows.Close()
		for rows.Next() {
			tempTaskInfo := entities.ShopifyTaskInfo{}
			err = rows.StructScan(&tempTaskInfo)
			if err != nil {
				return task, err
			}
			task.ShopifyTaskInfo = &tempTaskInfo
		}

	case enums.BigCartel:
		statement, err := database.Preparex(`SELECT * FROM bigcartelTaskInfos WHERE taskID = @p1`)
		if err != nil {
			return task, err
		}
		rows, err := statement.Queryx(task.ID)
		if err != nil {
			return task, err
		}

		defer rows.Close()
		for rows.Next() {
			tempTaskInfo := entities.BigCartelTaskInfo{}
			err = rows.StructScan(&tempTaskInfo)
			if err != nil {
				return task, err
			}
			task.BigCartelTaskInfo = &tempTaskInfo
		}

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
			tempTaskInfo := entities.TargetTaskInfo{}
			err = rows.StructScan(&tempTaskInfo)
			if err != nil {
				return task, err
			}
			task.TargetTaskInfo = &tempTaskInfo
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
			tempTaskInfo := entities.WalmartTaskInfo{}
			err = rows.StructScan(&tempTaskInfo)
			if err != nil {
				return task, err
			}
			task.WalmartTaskInfo = &tempTaskInfo
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
			tempMonitorInfo := entities.AmazonMonitorInfo{}
			err = rows.StructScan(&tempMonitorInfo)
			if err != nil {
				return taskGroup, err
			}
			taskGroup.AmazonMonitorInfo = &tempMonitorInfo
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
			tempMonitorInfo := entities.BestbuyMonitorInfo{}
			err = rows.StructScan(&tempMonitorInfo)
			if err != nil {
				return taskGroup, err
			}
			taskGroup.BestbuyMonitorInfo = &tempMonitorInfo
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
			tempMonitorInfo := entities.BoxlunchMonitorInfo{}
			err = rows.StructScan(&tempMonitorInfo)
			if err != nil {
				return taskGroup, err
			}
			taskGroup.BoxlunchMonitorInfo = &tempMonitorInfo
		}

		statement, err = database.Preparex(`SELECT * FROM boxlunchSingleMonitorInfos WHERE monitorID = @p1`)
		if err != nil {
			return taskGroup, err
		}

		rows, err = statement.Queryx(taskGroup.BoxlunchMonitorInfo.ID)
		if err != nil {
			return taskGroup, err
		}

		defer rows.Close()
		for rows.Next() {
			tempSingleMonitor := entities.BoxlunchSingleMonitorInfo{}
			err = rows.StructScan(&tempSingleMonitor)
			if err != nil {
				return taskGroup, err
			}
			taskGroup.BoxlunchMonitorInfo.Monitors = append(taskGroup.BoxlunchMonitorInfo.Monitors, tempSingleMonitor)
		}

	case enums.Disney:
		statement, err := database.Preparex(`SELECT * FROM disneyMonitorInfos WHERE taskGroupID = @p1`)
		if err != nil {
			return taskGroup, err
		}

		rows, err := statement.Queryx(taskGroup.GroupID)
		if err != nil {
			return taskGroup, err
		}

		defer rows.Close()
		for rows.Next() {
			tempMonitorInfo := entities.DisneyMonitorInfo{}
			err = rows.StructScan(&tempMonitorInfo)
			if err != nil {
				return taskGroup, err
			}
			taskGroup.DisneyMonitorInfo = &tempMonitorInfo
		}

		statement, err = database.Preparex(`SELECT * FROM disneySingleMonitorInfos WHERE monitorID = @p1`)
		if err != nil {
			return taskGroup, err
		}

		rows, err = statement.Queryx(taskGroup.DisneyMonitorInfo.ID)
		if err != nil {
			return taskGroup, err
		}

		defer rows.Close()
		for rows.Next() {
			tempSingleMonitor := entities.DisneySingleMonitorInfo{}
			err = rows.StructScan(&tempSingleMonitor)
			if err != nil {
				return taskGroup, err
			}
			taskGroup.DisneyMonitorInfo.Monitors = append(taskGroup.DisneyMonitorInfo.Monitors, tempSingleMonitor)
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
			tempMonitorInfo := entities.GamestopMonitorInfo{}
			err = rows.StructScan(&tempMonitorInfo)
			if err != nil {
				return taskGroup, err
			}
			taskGroup.GamestopMonitorInfo = &tempMonitorInfo
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

	case enums.HotTopic:
		statement, err := database.Preparex(`SELECT * FROM hottopicMonitorInfos WHERE taskGroupID = @p1`)
		if err != nil {
			return taskGroup, err
		}

		rows, err := statement.Queryx(taskGroup.GroupID)
		if err != nil {
			return taskGroup, err
		}
		defer rows.Close()

		for rows.Next() {
			tempMonitorInfo := entities.HottopicMonitorInfo{}
			err = rows.StructScan(&tempMonitorInfo)
			if err != nil {
				return taskGroup, err
			}
			taskGroup.HottopicMonitorInfo = &tempMonitorInfo
		}
		statement, err = database.Preparex(`SELECT * FROM hottopicSingleMonitorInfos WHERE monitorID = @p1`)
		if err != nil {
			return taskGroup, err
		}

		rows, err = statement.Queryx(taskGroup.HottopicMonitorInfo.ID)
		if err != nil {
			return taskGroup, err
		}
		defer rows.Close()

		for rows.Next() {
			tempSingleMonitor := entities.HottopicSingleMonitorInfo{}
			err = rows.StructScan(&tempSingleMonitor)
			if err != nil {
				return taskGroup, err
			}
			taskGroup.HottopicMonitorInfo.Monitors = append(taskGroup.HottopicMonitorInfo.Monitors, tempSingleMonitor)
		}

	case enums.Shopify:
		statement, err := database.Preparex(`SELECT * FROM shopifyMonitorInfos WHERE taskGroupID = @p1`)
		if err != nil {
			return taskGroup, err
		}
		rows, err := statement.Queryx(taskGroup.GroupID)
		if err != nil {
			return taskGroup, err
		}

		defer rows.Close()
		for rows.Next() {
			tempMonitorInfo := entities.ShopifyMonitorInfo{}
			err = rows.StructScan(&tempMonitorInfo)
			if err != nil {
				return taskGroup, err
			}
			taskGroup.ShopifyMonitorInfo = &tempMonitorInfo
		}

		statement, err = database.Preparex(`SELECT * FROM shopifySingleMonitorInfos WHERE monitorID = @p1`)
		if err != nil {
			return taskGroup, err
		}

		rows, err = statement.Queryx(taskGroup.ShopifyMonitorInfo.ID)
		if err != nil {
			return taskGroup, err
		}

		defer rows.Close()
		for rows.Next() {
			tempSingleMonitor := entities.ShopifySingleMonitorInfo{}
			err = rows.StructScan(&tempSingleMonitor)
			if err != nil {
				return taskGroup, err
			}
			taskGroup.ShopifyMonitorInfo.Monitors = append(taskGroup.ShopifyMonitorInfo.Monitors, tempSingleMonitor)
		}

	case enums.BigCartel:
		statement, err := database.Preparex(`SELECT * FROM bigcartelMonitorInfos WHERE taskGroupID = @p1`)
		if err != nil {
			return taskGroup, err
		}
		rows, err := statement.Queryx(taskGroup.GroupID)
		if err != nil {
			return taskGroup, err
		}

		defer rows.Close()
		for rows.Next() {
			tempMonitorInfo := entities.BigCartelMonitorInfo{}
			err = rows.StructScan(&tempMonitorInfo)
			if err != nil {
				return taskGroup, err
			}
			taskGroup.BigCartelMonitorInfo = &tempMonitorInfo
		}

		statement, err = database.Preparex(`SELECT * FROM bigcartelSingleMonitorInfos WHERE monitorID = @p1`)
		if err != nil {
			return taskGroup, err
		}

		rows, err = statement.Queryx(taskGroup.BigCartelMonitorInfo.ID)
		if err != nil {
			return taskGroup, err
		}

		defer rows.Close()
		for rows.Next() {
			tempSingleMonitor := entities.BigCartelSingleMonitorInfo{}
			err = rows.StructScan(&tempSingleMonitor)
			if err != nil {
				return taskGroup, err
			}
			taskGroup.BigCartelMonitorInfo.Monitors = append(taskGroup.BigCartelMonitorInfo.Monitors, tempSingleMonitor)
		}

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
			tempMonitorInfo := entities.TargetMonitorInfo{}
			err = rows.StructScan(&tempMonitorInfo)
			if err != nil {
				return taskGroup, err
			}
			taskGroup.TargetMonitorInfo = &tempMonitorInfo
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
			tempMonitorInfo := entities.WalmartMonitorInfo{}
			err = rows.StructScan(&tempMonitorInfo)
			if err != nil {
				return taskGroup, err
			}
			taskGroup.WalmartMonitorInfo = &tempMonitorInfo
		}

		statement, err = database.Preparex(`SELECT * FROM walmartSingleMonitorInfos WHERE monitorID = @p1`)
		if err != nil {
			return taskGroup, err
		}

		rows, err = statement.Queryx(taskGroup.WalmartMonitorInfo.ID)
		if err != nil {
			return taskGroup, err
		}
		defer rows.Close()

		for rows.Next() {
			tempSingleMonitor := entities.WalmartSingleMonitorInfo{}
			err = rows.StructScan(&tempSingleMonitor)
			if err != nil {
				return taskGroup, err
			}
			taskGroup.WalmartMonitorInfo.Monitors = append(taskGroup.WalmartMonitorInfo.Monitors, tempSingleMonitor)
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

	var encryptedCardNumber string
	var encryptedCVV string
	defer rows.Close()
	for rows.Next() {
		err = rows.StructScan(&profile.CreditCard)
		if err != nil {
			return profile, err
		}
		decryptedCardNumber, err := common.Aes256Decrypt(profile.CreditCard.CardNumber, enums.UserKey)
		if err == nil {
			profile.CreditCard.CardNumber = decryptedCardNumber
		} else {
			encryptedCardNumber, err = common.Aes256Encrypt(profile.CreditCard.CardNumber, enums.UserKey)
			if err != nil {
				return profile, err
			}
		}
		decryptedCVV, err := common.Aes256Decrypt(profile.CreditCard.CVV, enums.UserKey)
		if err == nil {
			profile.CreditCard.CVV = decryptedCVV
		} else {
			encryptedCVV, err = common.Aes256Encrypt(profile.CreditCard.CVV, enums.UserKey)
			if err != nil {
				return profile, err
			}
		}
	}
	if encryptedCardNumber != "" {
		go func() {
			for {
				_, err = database.Exec(fmt.Sprintf(`UPDATE cards SET cardNumber = "%v" WHERE ID = "%v"`, encryptedCardNumber, profile.CreditCard.ID))
				if err != nil {
					continue
				} else {
					time.Sleep(1 * time.Second)
				}
			}

		}()
	}

	if encryptedCVV != "" {
		go func() {
			for {
				_, err = database.Exec(fmt.Sprintf(`UPDATE cards SET cvv = "%v" WHERE ID = "%v"`, encryptedCVV, profile.CreditCard.ID))
				if err != nil {
					continue
				} else {
					time.Sleep(1 * time.Second)
				}
			}

		}()
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
