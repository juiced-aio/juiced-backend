package queries

import (
	"errors"
	"fmt"
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
	// Let's see if we can abstract this
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

	case enums.PokemonCenter:
		statement, err := database.Preparex(`SELECT * FROM pokemoncenterTaskInfos WHERE taskID = @p1`)
		if err != nil {
			return task, err
		}
		rows, err := statement.Queryx(task.ID)
		if err != nil {
			return task, err
		}

		defer rows.Close()
		for rows.Next() {
			tempTaskInfo := entities.PokemonCenterTaskInfo{}
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

			task.PokemonCenterTaskInfo = &tempTaskInfo
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
	// Let's see if we can abstract this
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

	case enums.PokemonCenter:
		statement, err := database.Preparex(`SELECT * FROM pokemoncenterMonitorInfos WHERE taskGroupID = @p1`)
		if err != nil {
			return taskGroup, err
		}

		rows, err := statement.Queryx(taskGroup.GroupID)
		if err != nil {
			return taskGroup, err
		}

		defer rows.Close()
		for rows.Next() {
			tempSingleMonitor := entities.PokemonCenterSingleMonitorInfo{}
			err = rows.StructScan(&tempSingleMonitor)
			if err != nil {
				return taskGroup, err
			}
			taskGroup.PokemonCenterMonitorInfo.Monitors = append(taskGroup.PokemonCenterMonitorInfo.Monitors, tempSingleMonitor)
		}
		statement, err = database.Preparex(`SELECT * FROM pokemoncenterSingleMonitorInfos WHERE monitorID = @p1`)
		if err != nil {
			return taskGroup, err
		}

		rows, err = statement.Queryx(taskGroup.PokemonCenterMonitorInfo.ID)
		if err != nil {
			return taskGroup, err
		}
		defer rows.Close()

		for rows.Next() {
			tempSingleMonitor := entities.PokemonCenterSingleMonitorInfo{}
			err = rows.StructScan(&tempSingleMonitor)
			if err != nil {
				return taskGroup, err
			}
			taskGroup.PokemonCenterMonitorInfo.Monitors = append(taskGroup.PokemonCenterMonitorInfo.Monitors, tempSingleMonitor)
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
