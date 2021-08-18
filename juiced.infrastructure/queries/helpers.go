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
	var encryptedEmail string
	var encryptedPassword string
	var currentTaskInfo string
	switch task.TaskRetailer {
	case enums.Amazon:
		currentTaskInfo = "amazonTaskInfos"
		statement, err := database.Preparex(`SELECT * FROM ` + currentTaskInfo + ` WHERE taskID = @p1`)
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

			decryptedEmail, err := common.Aes256Decrypt(tempTaskInfo.Email, enums.UserKey)
			if err == nil {
				tempTaskInfo.Email = decryptedEmail
			} else {
				encryptedEmail, err = common.Aes256Encrypt(tempTaskInfo.Email, enums.UserKey)
				if err != nil {
					return task, err
				}

			}

			decryptedPassword, err := common.Aes256Decrypt(tempTaskInfo.Password, enums.UserKey)
			if err == nil {
				tempTaskInfo.Password = decryptedPassword
			} else {
				encryptedPassword, err = common.Aes256Encrypt(tempTaskInfo.Password, enums.UserKey)
				if err != nil {
					return task, err
				}

			}

			task.AmazonTaskInfo = &tempTaskInfo
		}

	case enums.BestBuy:
		currentTaskInfo = "bestbuyTaskInfos"
		statement, err := database.Preparex(`SELECT * FROM ` + currentTaskInfo + ` WHERE taskID = @p1`)
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

			decryptedEmail, err := common.Aes256Decrypt(tempTaskInfo.Email, enums.UserKey)
			if err == nil {
				tempTaskInfo.Email = decryptedEmail
			} else {
				encryptedEmail, err = common.Aes256Encrypt(tempTaskInfo.Email, enums.UserKey)
				if err != nil {
					return task, err
				}

			}

			decryptedPassword, err := common.Aes256Decrypt(tempTaskInfo.Password, enums.UserKey)
			if err == nil {
				tempTaskInfo.Password = decryptedPassword
			} else {
				encryptedPassword, err = common.Aes256Encrypt(tempTaskInfo.Password, enums.UserKey)
				if err != nil {
					return task, err
				}

			}

			task.BestbuyTaskInfo = &tempTaskInfo
		}

	case enums.BoxLunch:
		currentTaskInfo = "boxlunchTaskInfos"
		statement, err := database.Preparex(`SELECT * FROM ` + currentTaskInfo + ` WHERE taskID = @p1`)
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
		currentTaskInfo = "disneyTaskInfos"
		statement, err := database.Preparex(`SELECT * FROM ` + currentTaskInfo + ` WHERE taskID = @p1`)
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

			decryptedEmail, err := common.Aes256Decrypt(tempTaskInfo.Email, enums.UserKey)
			if err == nil {
				tempTaskInfo.Email = decryptedEmail
			} else {
				encryptedEmail, err = common.Aes256Encrypt(tempTaskInfo.Email, enums.UserKey)
				if err != nil {
					return task, err
				}

			}

			decryptedPassword, err := common.Aes256Decrypt(tempTaskInfo.Password, enums.UserKey)
			if err == nil {
				tempTaskInfo.Password = decryptedPassword
			} else {
				encryptedPassword, err = common.Aes256Encrypt(tempTaskInfo.Password, enums.UserKey)
				if err != nil {
					return task, err
				}

			}

			task.DisneyTaskInfo = &tempTaskInfo
		}

	case enums.GameStop:
		currentTaskInfo = "gamestopTaskInfos"
		statement, err := database.Preparex(`SELECT * FROM ` + currentTaskInfo + ` WHERE taskID = @p1`)
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

			decryptedEmail, err := common.Aes256Decrypt(tempTaskInfo.Email, enums.UserKey)
			if err == nil {
				tempTaskInfo.Email = decryptedEmail
			} else {
				encryptedEmail, err = common.Aes256Encrypt(tempTaskInfo.Email, enums.UserKey)
				if err != nil {
					return task, err
				}

			}

			decryptedPassword, err := common.Aes256Decrypt(tempTaskInfo.Password, enums.UserKey)
			if err == nil {
				tempTaskInfo.Password = decryptedPassword
			} else {
				encryptedPassword, err = common.Aes256Encrypt(tempTaskInfo.Password, enums.UserKey)
				if err != nil {
					return task, err
				}

			}

			task.GamestopTaskInfo = &tempTaskInfo
		}

	case enums.HotTopic:
		currentTaskInfo = "hottopicTaskInfos"
		statement, err := database.Preparex(`SELECT * FROM ` + currentTaskInfo + ` WHERE taskID = @p1`)
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

	case enums.Newegg:
		statement, err := database.Preparex(`SELECT * FROM neweggTaskInfos WHERE taskID = @p1`)
		if err != nil {
			return task, err
		}
		rows, err := statement.Queryx(task.ID)
		if err != nil {
			return task, err
		}

		defer rows.Close()
		for rows.Next() {
			tempTaskInfo := entities.NeweggTaskInfo{}
			err = rows.StructScan(&tempTaskInfo)
			if err != nil {
				return task, err
			}
			task.NeweggTaskInfo = &tempTaskInfo
		}

	case enums.Shopify:
		currentTaskInfo = "shopifyTaskInfos"
		statement, err := database.Preparex(`SELECT * FROM ` + currentTaskInfo + ` WHERE taskID = @p1`)
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

		switch task.ShopifyTaskInfo.ShopifyRetailer {
		case enums.HotWheels:
			currentTaskInfo = "hotwheelsTaskInfos"
			statement, err := database.Preparex(`SELECT * FROM ` + currentTaskInfo + ` WHERE taskID = @p1`)
			if err != nil {
				return task, err
			}
			rows, err := statement.Queryx(task.ID)
			if err != nil {
				return task, err
			}

			defer rows.Close()
			for rows.Next() {
				tempTaskInfo := entities.HotWheelsTaskInfo{}
				err = rows.StructScan(&tempTaskInfo)
				if err != nil {
					return task, err
				}

				decryptedEmail, err := common.Aes256Decrypt(tempTaskInfo.Email, enums.UserKey)
				if err == nil {
					tempTaskInfo.Email = decryptedEmail
				} else {
					encryptedEmail, err = common.Aes256Encrypt(tempTaskInfo.Email, enums.UserKey)
					if err != nil {
						return task, err
					}

				}

				decryptedPassword, err := common.Aes256Decrypt(tempTaskInfo.Password, enums.UserKey)
				if err == nil {
					tempTaskInfo.Password = decryptedPassword
				} else {
					encryptedPassword, err = common.Aes256Encrypt(tempTaskInfo.Password, enums.UserKey)
					if err != nil {
						return task, err
					}

				}

				task.ShopifyTaskInfo.HotWheelsTaskInfo = &tempTaskInfo
			}
		}

	case enums.Target:
		currentTaskInfo = "targetTaskInfos"
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

			decryptedEmail, err := common.Aes256Decrypt(tempTaskInfo.Email, enums.UserKey)
			if err == nil {
				tempTaskInfo.Email = decryptedEmail
			} else {
				encryptedEmail, err = common.Aes256Encrypt(tempTaskInfo.Email, enums.UserKey)
				if err != nil {
					return task, err
				}

			}

			decryptedPassword, err := common.Aes256Decrypt(tempTaskInfo.Password, enums.UserKey)
			if err == nil {
				tempTaskInfo.Password = decryptedPassword
			} else {
				encryptedPassword, err = common.Aes256Encrypt(tempTaskInfo.Password, enums.UserKey)
				if err != nil {
					return task, err
				}

			}

			task.TargetTaskInfo = &tempTaskInfo
		}

	case enums.Topps:
		statement, err := database.Preparex(`SELECT * FROM toppsTaskInfos WHERE taskID = @p1`)
		if err != nil {
			return task, err
		}
		rows, err := statement.Queryx(task.ID)
		if err != nil {
			return task, err
		}

		defer rows.Close()
		for rows.Next() {
			tempTaskInfo := entities.ToppsTaskInfo{}
			err = rows.StructScan(&tempTaskInfo)
			if err != nil {
				return task, err
			}
			task.ToppsTaskInfo = &tempTaskInfo
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

	if encryptedEmail != "" {
		go func() {
			for {
				_, err := database.Exec(fmt.Sprintf(`UPDATE %v SET email = "%v" WHERE ID = "%v"`, currentTaskInfo, encryptedEmail, task.ID))
				if err != nil {
					time.Sleep(1 * time.Second)
					continue
				} else {
					break
				}
			}
		}()
	}

	if encryptedPassword != "" {
		go func() {
			for {
				_, err := database.Exec(fmt.Sprintf(`UPDATE %v SET password = "%v" WHERE ID = "%v"`, currentTaskInfo, encryptedPassword, task.ID))
				if err != nil {
					time.Sleep(1 * time.Second)
					continue
				} else {
					break
				}
			}
		}()
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

	case enums.Newegg:
		statement, err := database.Preparex(`SELECT * FROM neweggMonitorInfos WHERE taskGroupID = @p1`)
		if err != nil {
			return taskGroup, err
		}

		rows, err := statement.Queryx(taskGroup.GroupID)
		if err != nil {
			return taskGroup, err
		}

		defer rows.Close()
		for rows.Next() {
			tempMonitorInfo := entities.NeweggMonitorInfo{}
			err = rows.StructScan(&tempMonitorInfo)
			if err != nil {
				return taskGroup, err
			}
			taskGroup.NeweggMonitorInfo = &tempMonitorInfo
		}
		statement, err = database.Preparex(`SELECT * FROM neweggSingleMonitorInfos WHERE monitorID = @p1`)
		if err != nil {
			return taskGroup, err
		}

		rows, err = statement.Queryx(taskGroup.NeweggMonitorInfo.ID)
		if err != nil {
			return taskGroup, err
		}
		defer rows.Close()

		for rows.Next() {
			tempSingleMonitor := entities.NeweggSingleMonitorInfo{}
			err = rows.StructScan(&tempSingleMonitor)
			if err != nil {
				return taskGroup, err
			}
			taskGroup.NeweggMonitorInfo.Monitors = append(taskGroup.NeweggMonitorInfo.Monitors, tempSingleMonitor)
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

	case enums.Topps:
		statement, err := database.Preparex(`SELECT * FROM toppsMonitorInfos WHERE taskGroupID = @p1`)
		if err != nil {
			return taskGroup, err
		}

		rows, err := statement.Queryx(taskGroup.GroupID)
		if err != nil {
			return taskGroup, err
		}

		defer rows.Close()
		for rows.Next() {
			tempMonitorInfo := entities.ToppsMonitorInfo{}
			err = rows.StructScan(&tempMonitorInfo)
			if err != nil {
				return taskGroup, err
			}
			taskGroup.ToppsMonitorInfo = &tempMonitorInfo
		}
		statement, err = database.Preparex(`SELECT * FROM toppsSingleMonitorInfos WHERE monitorID = @p1`)
		if err != nil {
			return taskGroup, err
		}

		rows, err = statement.Queryx(taskGroup.ToppsMonitorInfo.ID)
		if err != nil {
			return taskGroup, err
		}
		defer rows.Close()

		for rows.Next() {
			tempSingleMonitor := entities.ToppsSingleMonitorInfo{}
			err = rows.StructScan(&tempSingleMonitor)
			if err != nil {
				return taskGroup, err
			}
			taskGroup.ToppsMonitorInfo.Monitors = append(taskGroup.ToppsMonitorInfo.Monitors, tempSingleMonitor)
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
		proxyGroup.Proxies = append(proxyGroup.Proxies, &tempProxy)
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
		var encryptedFirstName string
		var encryptedLastName string
		var encryptedAddress1 string
		var encryptedAddress2 string
		var encryptedCity string
		var encryptedZipCode string
		var encryptedStateCode string
		var encryptedCountryCode string

		err = rows.StructScan(&profile.ShippingAddress)
		if err != nil {
			return profile, err
		}

		decryptedFirstName, err := common.Aes256Decrypt(profile.ShippingAddress.FirstName, enums.UserKey)
		if err == nil {
			profile.ShippingAddress.FirstName = decryptedFirstName
		} else {
			encryptedFirstName, err = common.Aes256Encrypt(profile.ShippingAddress.FirstName, enums.UserKey)
			if err != nil {
				return profile, err
			}
		}
		decryptedLastName, err := common.Aes256Decrypt(profile.ShippingAddress.LastName, enums.UserKey)
		if err == nil {
			profile.ShippingAddress.LastName = decryptedLastName
		} else {
			encryptedLastName, err = common.Aes256Encrypt(profile.ShippingAddress.LastName, enums.UserKey)
			if err != nil {
				return profile, err
			}
		}
		decryptedAddress1, err := common.Aes256Decrypt(profile.ShippingAddress.Address1, enums.UserKey)
		if err == nil {
			profile.ShippingAddress.Address1 = decryptedAddress1
		} else {
			encryptedAddress1, err = common.Aes256Encrypt(profile.ShippingAddress.Address1, enums.UserKey)
			if err != nil {
				return profile, err
			}
		}
		decryptedAddress2, err := common.Aes256Decrypt(profile.ShippingAddress.Address2, enums.UserKey)
		if err == nil {
			profile.ShippingAddress.Address2 = decryptedAddress2
		} else {
			encryptedAddress2, err = common.Aes256Encrypt(profile.ShippingAddress.Address2, enums.UserKey)
			if err != nil {
				return profile, err
			}
		}
		decryptedCity, err := common.Aes256Decrypt(profile.ShippingAddress.City, enums.UserKey)
		if err == nil {
			profile.ShippingAddress.City = decryptedCity
		} else {
			encryptedCity, err = common.Aes256Encrypt(profile.ShippingAddress.City, enums.UserKey)
			if err != nil {
				return profile, err
			}
		}
		decryptedZipCode, err := common.Aes256Decrypt(profile.ShippingAddress.ZipCode, enums.UserKey)
		if err == nil {
			profile.ShippingAddress.ZipCode = decryptedZipCode
		} else {
			encryptedZipCode, err = common.Aes256Encrypt(profile.ShippingAddress.ZipCode, enums.UserKey)
			if err != nil {
				return profile, err
			}
		}
		decryptedStateCode, err := common.Aes256Decrypt(profile.ShippingAddress.StateCode, enums.UserKey)
		if err == nil {
			profile.ShippingAddress.StateCode = decryptedStateCode
		} else {
			encryptedStateCode, err = common.Aes256Encrypt(profile.ShippingAddress.StateCode, enums.UserKey)
			if err != nil {
				return profile, err
			}
		}
		decryptedCountryCode, err := common.Aes256Decrypt(profile.ShippingAddress.CountryCode, enums.UserKey)
		if err == nil {
			profile.ShippingAddress.CountryCode = decryptedCountryCode
		} else {
			encryptedCountryCode, err = common.Aes256Encrypt(profile.ShippingAddress.CountryCode, enums.UserKey)
			if err != nil {
				return profile, err
			}
		}

		if encryptedFirstName != "" {
			go func() {
				for {
					_, err = database.Exec(fmt.Sprintf(`UPDATE shippingAddresses SET firstName = "%v" WHERE ID = "%v"`, encryptedFirstName, profile.ShippingAddress.ID))
					if err != nil {
						time.Sleep(1 * time.Second)
						continue
					} else {
						break
					}
				}
			}()
		}

		if encryptedLastName != "" {
			go func() {
				for {
					_, err = database.Exec(fmt.Sprintf(`UPDATE shippingAddresses SET lastName = "%v" WHERE ID = "%v"`, encryptedLastName, profile.ShippingAddress.ID))
					if err != nil {
						time.Sleep(1 * time.Second)
						continue
					} else {
						break
					}
				}
			}()
		}

		if encryptedAddress1 != "" {
			go func() {
				for {
					_, err = database.Exec(fmt.Sprintf(`UPDATE shippingAddresses SET address1 = "%v" WHERE ID = "%v"`, encryptedAddress1, profile.ShippingAddress.ID))
					if err != nil {
						time.Sleep(1 * time.Second)
						continue
					} else {
						break
					}
				}
			}()
		}

		if encryptedAddress2 != "" {
			go func() {
				for {
					_, err = database.Exec(fmt.Sprintf(`UPDATE shippingAddresses SET address2 = "%v" WHERE ID = "%v"`, encryptedAddress2, profile.ShippingAddress.ID))
					if err != nil {
						time.Sleep(1 * time.Second)
						continue
					} else {
						break
					}
				}
			}()
		}

		if encryptedCity != "" {
			go func() {
				for {
					_, err = database.Exec(fmt.Sprintf(`UPDATE shippingAddresses SET city = "%v" WHERE ID = "%v"`, encryptedCity, profile.ShippingAddress.ID))
					if err != nil {
						time.Sleep(1 * time.Second)
						continue
					} else {
						break
					}
				}

			}()
		}

		if encryptedZipCode != "" {
			go func() {
				for {
					_, err = database.Exec(fmt.Sprintf(`UPDATE shippingAddresses SET zipCode = "%v" WHERE ID = "%v"`, encryptedZipCode, profile.ShippingAddress.ID))
					if err != nil {
						time.Sleep(1 * time.Second)
						continue
					} else {
						break
					}
				}
			}()
		}

		if encryptedStateCode != "" {
			go func() {
				for {
					_, err = database.Exec(fmt.Sprintf(`UPDATE shippingAddresses SET stateCode = "%v" WHERE ID = "%v"`, encryptedStateCode, profile.ShippingAddress.ID))
					if err != nil {
						time.Sleep(1 * time.Second)
						continue
					} else {
						break
					}
				}
			}()
		}

		if encryptedCountryCode != "" {
			go func() {
				for {
					_, err = database.Exec(fmt.Sprintf(`UPDATE shippingAddresses SET countryCode = "%v" WHERE ID = "%v"`, encryptedCountryCode, profile.ShippingAddress.ID))
					if err != nil {
						time.Sleep(1 * time.Second)
						continue
					} else {
						break
					}
				}
			}()
		}

	}

	return profile, err
}

func GetBillingAddress(profile entities.Profile) (entities.Profile, error) {
	database := common.GetDatabase()
	if database == nil {
		return profile, errors.New("database not initialized")
	}

	statement, err := database.Preparex(`SELECT * FROM billingAddresses WHERE profileID = @p1`)
	if err != nil {
		return profile, err
	}
	rows, err := statement.Queryx(profile.ID)
	if err != nil {
		return profile, err
	}

	defer rows.Close()
	for rows.Next() {
		var encryptedFirstName string
		var encryptedLastName string
		var encryptedAddress1 string
		var encryptedAddress2 string
		var encryptedCity string
		var encryptedZipCode string
		var encryptedStateCode string
		var encryptedCountryCode string

		err = rows.StructScan(&profile.BillingAddress)
		if err != nil {
			return profile, err
		}

		decryptedFirstName, err := common.Aes256Decrypt(profile.BillingAddress.FirstName, enums.UserKey)
		if err == nil {
			profile.BillingAddress.FirstName = decryptedFirstName
		} else {
			encryptedFirstName, err = common.Aes256Encrypt(profile.BillingAddress.FirstName, enums.UserKey)
			if err != nil {
				return profile, err
			}
		}
		decryptedLastName, err := common.Aes256Decrypt(profile.BillingAddress.LastName, enums.UserKey)
		if err == nil {
			profile.BillingAddress.LastName = decryptedLastName
		} else {
			encryptedLastName, err = common.Aes256Encrypt(profile.BillingAddress.LastName, enums.UserKey)
			if err != nil {
				return profile, err
			}
		}
		decryptedAddress1, err := common.Aes256Decrypt(profile.BillingAddress.Address1, enums.UserKey)
		if err == nil {
			profile.BillingAddress.Address1 = decryptedAddress1
		} else {
			encryptedAddress1, err = common.Aes256Encrypt(profile.BillingAddress.Address1, enums.UserKey)
			if err != nil {
				return profile, err
			}
		}
		decryptedAddress2, err := common.Aes256Decrypt(profile.BillingAddress.Address2, enums.UserKey)
		if err == nil {
			profile.BillingAddress.Address2 = decryptedAddress2
		} else {
			encryptedAddress2, err = common.Aes256Encrypt(profile.BillingAddress.Address2, enums.UserKey)
			if err != nil {
				return profile, err
			}
		}
		decryptedCity, err := common.Aes256Decrypt(profile.BillingAddress.City, enums.UserKey)
		if err == nil {
			profile.BillingAddress.City = decryptedCity
		} else {
			encryptedCity, err = common.Aes256Encrypt(profile.BillingAddress.City, enums.UserKey)
			if err != nil {
				return profile, err
			}
		}
		decryptedZipCode, err := common.Aes256Decrypt(profile.BillingAddress.ZipCode, enums.UserKey)
		if err == nil {
			profile.BillingAddress.ZipCode = decryptedZipCode
		} else {
			encryptedZipCode, err = common.Aes256Encrypt(profile.BillingAddress.ZipCode, enums.UserKey)
			if err != nil {
				return profile, err
			}
		}
		decryptedStateCode, err := common.Aes256Decrypt(profile.BillingAddress.StateCode, enums.UserKey)
		if err == nil {
			profile.BillingAddress.StateCode = decryptedStateCode
		} else {
			encryptedStateCode, err = common.Aes256Encrypt(profile.BillingAddress.StateCode, enums.UserKey)
			if err != nil {
				return profile, err
			}
		}
		decryptedCountryCode, err := common.Aes256Decrypt(profile.BillingAddress.CountryCode, enums.UserKey)
		if err == nil {
			profile.BillingAddress.CountryCode = decryptedCountryCode
		} else {
			encryptedCountryCode, err = common.Aes256Encrypt(profile.BillingAddress.CountryCode, enums.UserKey)
			if err != nil {
				return profile, err
			}
		}

		if encryptedFirstName != "" {
			go func() {
				for {
					_, err = database.Exec(fmt.Sprintf(`UPDATE billingAddresses SET firstName = "%v" WHERE ID = "%v"`, encryptedFirstName, profile.BillingAddress.ID))
					if err != nil {
						time.Sleep(1 * time.Second)
						continue
					} else {
						break
					}
				}
			}()
		}

		if encryptedLastName != "" {
			go func() {
				for {
					_, err = database.Exec(fmt.Sprintf(`UPDATE billingAddresses SET lastName = "%v" WHERE ID = "%v"`, encryptedLastName, profile.BillingAddress.ID))
					if err != nil {
						time.Sleep(1 * time.Second)
						continue
					} else {
						break
					}
				}
			}()
		}

		if encryptedAddress1 != "" {
			go func() {
				for {
					_, err = database.Exec(fmt.Sprintf(`UPDATE billingAddresses SET address1 = "%v" WHERE ID = "%v"`, encryptedAddress1, profile.BillingAddress.ID))
					if err != nil {
						time.Sleep(1 * time.Second)
						continue
					} else {
						break
					}
				}
			}()
		}

		if encryptedAddress2 != "" {
			go func() {
				for {
					_, err = database.Exec(fmt.Sprintf(`UPDATE billingAddresses SET address2 = "%v" WHERE ID = "%v"`, encryptedAddress2, profile.BillingAddress.ID))
					if err != nil {
						time.Sleep(1 * time.Second)
						continue
					} else {
						break
					}
				}
			}()
		}

		if encryptedCity != "" {
			go func() {
				for {
					_, err = database.Exec(fmt.Sprintf(`UPDATE billingAddresses SET city = "%v" WHERE ID = "%v"`, encryptedCity, profile.BillingAddress.ID))
					if err != nil {
						time.Sleep(1 * time.Second)
						continue
					} else {
						break
					}
				}

			}()
		}

		if encryptedZipCode != "" {
			go func() {
				for {
					_, err = database.Exec(fmt.Sprintf(`UPDATE billingAddresses SET zipCode = "%v" WHERE ID = "%v"`, encryptedZipCode, profile.BillingAddress.ID))
					if err != nil {
						time.Sleep(1 * time.Second)
						continue
					} else {
						break
					}
				}
			}()
		}

		if encryptedStateCode != "" {
			go func() {
				for {
					_, err = database.Exec(fmt.Sprintf(`UPDATE billingAddresses SET stateCode = "%v" WHERE ID = "%v"`, encryptedStateCode, profile.BillingAddress.ID))
					if err != nil {
						time.Sleep(1 * time.Second)
						continue
					} else {
						break
					}
				}
			}()
		}

		if encryptedCountryCode != "" {
			go func() {
				for {
					_, err = database.Exec(fmt.Sprintf(`UPDATE billingAddresses SET countryCode = "%v" WHERE ID = "%v"`, encryptedCountryCode, profile.BillingAddress.ID))
					if err != nil {
						time.Sleep(1 * time.Second)
						continue
					} else {
						break
					}
				}
			}()
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
		var encryptedCardholderName string
		var encryptedCardNumber string
		var encryptedExpMonth string
		var encryptedExpYear string
		var encryptedCVV string
		var encryptedCardType string

		err = rows.StructScan(&profile.CreditCard)
		if err != nil {
			return profile, err
		}

		decryptedCardholderName, err := common.Aes256Decrypt(profile.CreditCard.CardholderName, enums.UserKey)
		if err == nil {
			profile.CreditCard.CardholderName = decryptedCardholderName
		} else {
			encryptedCardholderName, err = common.Aes256Encrypt(profile.CreditCard.CardholderName, enums.UserKey)
			if err != nil {
				return profile, err
			}
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
		decryptedExpMonth, err := common.Aes256Decrypt(profile.CreditCard.ExpMonth, enums.UserKey)
		if err == nil {
			profile.CreditCard.ExpMonth = decryptedExpMonth
		} else {
			encryptedExpMonth, err = common.Aes256Encrypt(profile.CreditCard.ExpMonth, enums.UserKey)
			if err != nil {
				return profile, err
			}
		}
		decryptedExpYear, err := common.Aes256Decrypt(profile.CreditCard.ExpYear, enums.UserKey)
		if err == nil {
			profile.CreditCard.ExpYear = decryptedExpYear
		} else {
			encryptedExpYear, err = common.Aes256Encrypt(profile.CreditCard.ExpYear, enums.UserKey)
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
		decryptedCardType, err := common.Aes256Decrypt(profile.CreditCard.CardType, enums.UserKey)
		if err == nil {
			profile.CreditCard.CardType = decryptedCardType
		} else {
			encryptedCardType, err = common.Aes256Encrypt(profile.CreditCard.CardType, enums.UserKey)
			if err != nil {
				return profile, err
			}
		}

		if encryptedCardholderName != "" {
			go func() {
				for {
					_, err = database.Exec(fmt.Sprintf(`UPDATE cards SET cardHolderName = "%v" WHERE ID = "%v"`, encryptedCardholderName, profile.CreditCard.ID))
					if err != nil {
						time.Sleep(1 * time.Second)
						continue
					} else {
						break
					}
				}
			}()
		}

		if encryptedCardNumber != "" {
			go func() {
				for {
					_, err = database.Exec(fmt.Sprintf(`UPDATE cards SET cardNumber = "%v" WHERE ID = "%v"`, encryptedCardNumber, profile.CreditCard.ID))
					if err != nil {
						time.Sleep(1 * time.Second)
						continue
					} else {
						break
					}
				}
			}()
		}

		if encryptedExpMonth != "" {
			go func() {
				for {
					_, err = database.Exec(fmt.Sprintf(`UPDATE cards SET expMonth = "%v" WHERE ID = "%v"`, encryptedExpMonth, profile.CreditCard.ID))
					if err != nil {
						time.Sleep(1 * time.Second)
						continue
					} else {
						break
					}
				}
			}()
		}

		if encryptedExpYear != "" {
			go func() {
				for {
					_, err = database.Exec(fmt.Sprintf(`UPDATE cards SET expYear = "%v" WHERE ID = "%v"`, encryptedExpYear, profile.CreditCard.ID))
					if err != nil {
						time.Sleep(1 * time.Second)
						continue
					} else {
						break
					}
				}
			}()
		}

		if encryptedCVV != "" {
			go func() {
				for {
					_, err = database.Exec(fmt.Sprintf(`UPDATE cards SET cvv = "%v" WHERE ID = "%v"`, encryptedCVV, profile.CreditCard.ID))
					if err != nil {
						time.Sleep(1 * time.Second)
						continue
					} else {
						break
					}
				}

			}()
		}

		if encryptedCardType != "" {
			go func() {
				for {
					_, err = database.Exec(fmt.Sprintf(`UPDATE cards SET cardType = "%v" WHERE ID = "%v"`, encryptedCardType, profile.CreditCard.ID))
					if err != nil {
						time.Sleep(1 * time.Second)
						continue
					} else {
						break
					}
				}
			}()
		}
	}

	return profile, err
}

func GetProfileInfo(profile entities.Profile) (entities.Profile, error) {
	database := common.GetDatabase()
	if database == nil {
		return profile, errors.New("database not initialized")
	}

	if profile.ProfileGroupIDsJoined != "" {
		profile.ProfileGroupIDs = strings.Split(profile.ProfileGroupIDsJoined, ",")
	}

	var (
		encryptedEmail       string
		encryptedPhoneNumber string
	)

	decryptedEmail, err := common.Aes256Decrypt(profile.Email, enums.UserKey)
	if err == nil {
		profile.Email = decryptedEmail
	} else {
		encryptedEmail, err = common.Aes256Encrypt(profile.Email, enums.UserKey)
		if err != nil {
			return profile, err
		}
	}

	decryptedPhoneNumber, err := common.Aes256Decrypt(profile.PhoneNumber, enums.UserKey)
	if err == nil {
		profile.PhoneNumber = decryptedPhoneNumber
	} else {
		encryptedPhoneNumber, err = common.Aes256Encrypt(profile.PhoneNumber, enums.UserKey)
		if err != nil {
			return profile, err
		}
	}

	if encryptedEmail != "" {
		go func() {
			for {
				_, err = database.Exec(fmt.Sprintf(`UPDATE profiles SET email = "%v" WHERE ID = "%v"`, encryptedEmail, profile.ID))
				if err != nil {
					time.Sleep(1 * time.Second)
					continue
				} else {
					break
				}
			}
		}()
	}
	if encryptedPhoneNumber != "" {
		go func() {
			for {
				_, err = database.Exec(fmt.Sprintf(`UPDATE profiles SET phoneNumber = "%v" WHERE ID = "%v"`, encryptedPhoneNumber, profile.ID))
				if err != nil {
					time.Sleep(1 * time.Second)
					continue
				} else {
					break
				}
			}
		}()
	}

	profile, err = GetShippingAddress(profile)
	if err != nil {
		return profile, err
	}
	profile, err = GetBillingAddress(profile)
	if err != nil {
		return profile, err
	}

	return GetCard(profile)
}
