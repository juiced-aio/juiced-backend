package commands

import (
	"errors"

	"backend.juicedbot.io/juiced.infrastructure/common"
	"backend.juicedbot.io/juiced.infrastructure/common/entities"
	"backend.juicedbot.io/juiced.infrastructure/common/enums"
	"github.com/google/uuid"
	_ "github.com/jmoiron/sqlx"
	_ "github.com/mattn/go-sqlite3"
)

func CreateMonitorInfos(taskGroup entities.TaskGroup) error {
	database := common.GetDatabase()
	if database == nil {
		return errors.New("database not initialized")
	}
	monitorID := uuid.New().String()
	switch taskGroup.MonitorRetailer {
	case enums.Amazon:
		statement, err := database.Preparex(`INSERT INTO amazonMonitorInfos (ID, taskGroupID) VALUES (?, ?)`)
		if err != nil {
			return err
		}
		taskGroup.AmazonMonitorInfo.ID = monitorID
		taskGroup.AmazonMonitorInfo.TaskGroupID = taskGroup.GroupID
		_, err = statement.Exec(taskGroup.AmazonMonitorInfo.ID, taskGroup.AmazonMonitorInfo.TaskGroupID)
		if err != nil {
			return err
		}
		for _, monitor := range taskGroup.AmazonMonitorInfo.Monitors {
			statement, err := database.Preparex(`INSERT INTO amazonSingleMonitorInfos (monitorID, taskGroupID, monitorType, asin, ofid, maxPrice) VALUES (?, ?, ?, ?, ?, ?)`)
			if err != nil {
				return err
			}
			monitor.MonitorID = monitorID
			monitor.TaskGroupID = taskGroup.GroupID
			_, err = statement.Exec(monitor.MonitorID, monitor.TaskGroupID, monitor.MonitorType, monitor.ASIN, monitor.OFID, monitor.MaxPrice)
			if err != nil {
				return err
			}
		}

	case enums.BestBuy:
		statement, err := database.Preparex(`INSERT INTO bestbuyMonitorInfos (ID, taskGroupID) VALUES (?, ?)`)
		if err != nil {
			return err
		}
		taskGroup.BestbuyMonitorInfo.ID = monitorID
		taskGroup.BestbuyMonitorInfo.TaskGroupID = taskGroup.GroupID
		_, err = statement.Exec(taskGroup.BestbuyMonitorInfo.ID, taskGroup.BestbuyMonitorInfo.TaskGroupID)
		if err != nil {
			return err
		}
		for _, monitor := range taskGroup.BestbuyMonitorInfo.Monitors {
			statement, err := database.Preparex(`INSERT INTO bestbuySingleMonitorInfos (monitorID, taskGroupID, sku, maxPrice) VALUES (?, ?, ?, ?)`)
			if err != nil {
				return err
			}
			monitor.MonitorID = monitorID
			monitor.TaskGroupID = taskGroup.GroupID
			_, err = statement.Exec(monitor.MonitorID, monitor.TaskGroupID, monitor.SKU, monitor.MaxPrice)
			if err != nil {
				return err
			}
		}

	case enums.BoxLunch:
		statement, err := database.Preparex(`INSERT INTO boxlunchMonitorInfos (ID, taskGroupID) VALUES (?, ?)`)
		if err != nil {
			return err
		}
		taskGroup.BoxlunchMonitorInfo.ID = monitorID
		taskGroup.BoxlunchMonitorInfo.TaskGroupID = taskGroup.GroupID
		_, err = statement.Exec(taskGroup.BoxlunchMonitorInfo.ID, taskGroup.BoxlunchMonitorInfo.TaskGroupID)
		if err != nil {
			return err
		}
		for _, monitor := range taskGroup.BoxlunchMonitorInfo.Monitors {
			statement, err := database.Preparex(`INSERT INTO boxlunchSingleMonitorInfos (monitorID, taskGroupID, pid, size, color, maxPrice, monitorType) VALUES (?, ?, ?, ?, ?, ?, ?)`)
			if err != nil {
				return err
			}
			monitor.MonitorID = monitorID
			monitor.TaskGroupID = taskGroup.GroupID
			_, err = statement.Exec(monitor.MonitorID, monitor.TaskGroupID, monitor.Pid, monitor.Size, monitor.Color, monitor.MaxPrice, monitor.MonitorType)
			if err != nil {
				return err
			}
		}

	case enums.Disney:
		statement, err := database.Preparex(`INSERT INTO disneyMonitorInfos (ID, taskGroupID) VALUES (?, ?)`)
		if err != nil {
			return err
		}
		taskGroup.DisneyMonitorInfo.ID = monitorID
		taskGroup.DisneyMonitorInfo.TaskGroupID = taskGroup.GroupID
		_, err = statement.Exec(taskGroup.DisneyMonitorInfo.ID, taskGroup.DisneyMonitorInfo.TaskGroupID)
		if err != nil {
			return err
		}
		for _, monitor := range taskGroup.DisneyMonitorInfo.Monitors {
			statement, err := database.Preparex(`INSERT INTO disneySingleMonitorInfos (monitorID, taskGroupID, pid, size, color, maxPrice) VALUES (?, ?, ?, ?, ?, ?)`)
			if err != nil {
				return err
			}
			monitor.MonitorID = monitorID
			monitor.TaskGroupID = taskGroup.GroupID
			_, err = statement.Exec(monitor.MonitorID, monitor.TaskGroupID, monitor.PID, monitor.Size, monitor.Color, monitor.MaxPrice)
			if err != nil {
				return err
			}
		}

	case enums.GameStop:
		statement, err := database.Preparex(`INSERT INTO gamestopMonitorInfos (ID, taskGroupID) VALUES (?, ?)`)
		if err != nil {
			return err
		}
		taskGroup.GamestopMonitorInfo.ID = monitorID
		taskGroup.GamestopMonitorInfo.TaskGroupID = taskGroup.GroupID
		_, err = statement.Exec(taskGroup.GamestopMonitorInfo.ID, taskGroup.GamestopMonitorInfo.TaskGroupID)
		if err != nil {
			return err
		}
		for _, monitor := range taskGroup.GamestopMonitorInfo.Monitors {
			statement, err := database.Preparex(`INSERT INTO gamestopSingleMonitorInfos (monitorID, taskGroupID, sku, maxPrice) VALUES (?, ?, ?, ?)`)
			if err != nil {
				return err
			}
			monitor.MonitorID = monitorID
			monitor.TaskGroupID = taskGroup.GroupID
			_, err = statement.Exec(monitor.MonitorID, monitor.TaskGroupID, monitor.SKU, monitor.MaxPrice)
			if err != nil {
				return err
			}
		}

	case enums.HotTopic:
		statement, err := database.Preparex(`INSERT INTO hottopicMonitorInfos (ID, taskGroupID) VALUES (?, ?)`)
		if err != nil {
			return err
		}
		taskGroup.HottopicMonitorInfo.ID = monitorID
		taskGroup.HottopicMonitorInfo.TaskGroupID = taskGroup.GroupID
		_, err = statement.Exec(taskGroup.HottopicMonitorInfo.ID, taskGroup.HottopicMonitorInfo.TaskGroupID)
		if err != nil {
			return err
		}
		for _, monitor := range taskGroup.HottopicMonitorInfo.Monitors {
			statement, err := database.Preparex(`INSERT INTO hottopicSingleMonitorInfos (monitorID, taskGroupID, pid, size, color, maxPrice, monitorType) VALUES (?, ?, ?, ?, ?, ?, ?)`)
			if err != nil {
				return err
			}
			monitor.MonitorID = monitorID
			monitor.TaskGroupID = taskGroup.GroupID
			_, err = statement.Exec(monitor.MonitorID, monitor.TaskGroupID, monitor.Pid, monitor.Size, monitor.Color, monitor.MaxPrice, monitor.MonitorType)
			if err != nil {
				return err
			}
		}

	case enums.Newegg:
		statement, err := database.Preparex(`INSERT INTO neweggMonitorInfos (ID, taskGroupID) VALUES (?, ?)`)
		if err != nil {
			return err
		}
		taskGroup.NeweggMonitorInfo.ID = monitorID
		taskGroup.NeweggMonitorInfo.TaskGroupID = taskGroup.GroupID
		_, err = statement.Exec(taskGroup.NeweggMonitorInfo.ID, taskGroup.NeweggMonitorInfo.TaskGroupID)
		if err != nil {
			return err
		}
		for _, monitor := range taskGroup.NeweggMonitorInfo.Monitors {
			statement, err := database.Preparex(`INSERT INTO neweggSingleMonitorInfos (monitorID, taskGroupID, sku, maxPrice) VALUES (?, ?, ?, ?)`)
			if err != nil {
				return err
			}
			monitor.MonitorID = monitorID
			monitor.TaskGroupID = taskGroup.GroupID
			_, err = statement.Exec(monitor.MonitorID, monitor.TaskGroupID, monitor.SKU, monitor.MaxPrice)
			if err != nil {
				return err
			}
		}

	case enums.PokemonCenter:
		statement, err := database.Preparex(`INSERT INTO pokemoncenterMonitorInfos (ID, taskGroupID) VALUES (?, ?)`)
		if err != nil {
			return err
		}
		taskGroup.PokemonCenterMonitorInfo.ID = monitorID
		taskGroup.PokemonCenterMonitorInfo.TaskGroupID = taskGroup.GroupID
		_, err = statement.Exec(taskGroup.PokemonCenterMonitorInfo.ID, taskGroup.PokemonCenterMonitorInfo.TaskGroupID)
		if err != nil {
			return err
		}
		for _, monitor := range taskGroup.PokemonCenterMonitorInfo.Monitors {
			statement, err := database.Preparex(`INSERT INTO pokemoncenterSingleMonitorInfos (monitorID, taskGroupID, sku, maxPrice) VALUES (?, ?, ?, ?)`)
			if err != nil {
				return err
			}
			monitor.MonitorID = monitorID
			monitor.TaskGroupID = taskGroup.GroupID
			_, err = statement.Exec(monitor.MonitorID, monitor.TaskGroupID, monitor.SKU, monitor.MaxPrice)
			if err != nil {
				return err
			}
		}

	case enums.Shopify:
		statement, err := database.Preparex(`INSERT INTO shopifyMonitorInfos (ID, taskGroupID, siteURL, sitePassword) VALUES (?, ?, ?, ?)`)
		if err != nil {
			return err
		}
		taskGroup.ShopifyMonitorInfo.ID = monitorID
		taskGroup.ShopifyMonitorInfo.TaskGroupID = taskGroup.GroupID
		_, err = statement.Exec(taskGroup.ShopifyMonitorInfo.ID, taskGroup.ShopifyMonitorInfo.TaskGroupID, taskGroup.ShopifyMonitorInfo.SiteURL, taskGroup.ShopifyMonitorInfo.SitePassword)
		if err != nil {
			return err
		}
		for _, monitor := range taskGroup.ShopifyMonitorInfo.Monitors {
			statement, err := database.Preparex(`INSERT INTO shopifySingleMonitorInfos (monitorID, taskGroupID, variantID, maxPrice) VALUES (?, ?, ?, ?)`)
			if err != nil {
				return err
			}
			monitor.MonitorID = monitorID
			monitor.TaskGroupID = taskGroup.GroupID
			_, err = statement.Exec(monitor.MonitorID, monitor.TaskGroupID, monitor.VariantID, monitor.MaxPrice)
			if err != nil {
				return err
			}
		}

	case enums.Target:
		statement, err := database.Preparex(`INSERT INTO targetMonitorInfos (ID, taskGroupID, storeID, monitorType) VALUES (?, ?, ?, ?)`)
		if err != nil {
			return err
		}
		taskGroup.TargetMonitorInfo.ID = monitorID
		taskGroup.TargetMonitorInfo.TaskGroupID = taskGroup.GroupID
		_, err = statement.Exec(taskGroup.TargetMonitorInfo.ID, taskGroup.TargetMonitorInfo.TaskGroupID, taskGroup.TargetMonitorInfo.StoreID, taskGroup.TargetMonitorInfo.MonitorType)
		if err != nil {
			return err
		}
		for _, monitor := range taskGroup.TargetMonitorInfo.Monitors {
			statement, err := database.Preparex(`INSERT INTO targetSingleMonitorInfos (monitorID, taskGroupID, tcin, maxPrice, checkoutType) VALUES (?, ?, ?, ?, ?)`)
			if err != nil {
				return err
			}
			monitor.MonitorID = monitorID
			monitor.TaskGroupID = taskGroup.GroupID
			_, err = statement.Exec(monitor.MonitorID, monitor.TaskGroupID, monitor.TCIN, monitor.MaxPrice, monitor.CheckoutType)
			if err != nil {
				return err
			}
		}

	case enums.Topps:
		statement, err := database.Preparex(`INSERT INTO toppsMonitorInfos (ID, taskGroupID) VALUES (?, ?)`)
		if err != nil {
			return err
		}
		taskGroup.ToppsMonitorInfo.ID = monitorID
		taskGroup.ToppsMonitorInfo.TaskGroupID = taskGroup.GroupID
		_, err = statement.Exec(taskGroup.ToppsMonitorInfo.ID, taskGroup.ToppsMonitorInfo.TaskGroupID)
		if err != nil {
			return err
		}
		for _, monitor := range taskGroup.ToppsMonitorInfo.Monitors {
			statement, err := database.Preparex(`INSERT INTO toppsSingleMonitorInfos (monitorID, taskGroupID, item, maxPrice) VALUES (?, ?, ?, ?)`)
			if err != nil {
				return err
			}
			monitor.MonitorID = monitorID
			monitor.TaskGroupID = taskGroup.GroupID
			_, err = statement.Exec(monitor.MonitorID, monitor.TaskGroupID, monitor.Item, monitor.MaxPrice)
			if err != nil {
				return err
			}
		}

	case enums.Walmart:
		statement, err := database.Preparex(`INSERT INTO walmartMonitorInfos (ID, taskGroupID) VALUES (?, ?)`)
		if err != nil {
			return err
		}
		taskGroup.WalmartMonitorInfo.ID = monitorID
		taskGroup.WalmartMonitorInfo.TaskGroupID = taskGroup.GroupID
		_, err = statement.Exec(taskGroup.WalmartMonitorInfo.ID, taskGroup.WalmartMonitorInfo.TaskGroupID)
		if err != nil {
			return err
		}
		for _, monitor := range taskGroup.WalmartMonitorInfo.Monitors {
			statement, err := database.Preparex(`INSERT INTO walmartSingleMonitorInfos (monitorID, taskGroupID, id, maxPrice, soldByWalmart, monitorType) VALUES (?, ?, ?, ?, ?, ?)`)
			if err != nil {
				return err
			}
			monitor.MonitorID = monitorID
			monitor.TaskGroupID = taskGroup.GroupID
			_, err = statement.Exec(monitor.MonitorID, monitor.TaskGroupID, monitor.ID, monitor.MaxPrice, monitor.SoldByWalmart, monitor.MonitorType)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func DeleteMonitorInfos(groupID string, retailer enums.Retailer) error {
	var monitorInfoSchema string
	var singleMonitorInfoSchema string
	switch retailer {
	case enums.Amazon:
		monitorInfoSchema = "amazonMonitorInfos"
		singleMonitorInfoSchema = "amazonSingleMonitorInfos"
	case enums.BestBuy:
		monitorInfoSchema = "bestbuyMonitorInfos"
		singleMonitorInfoSchema = "bestbuySingleMonitorInfos"
	case enums.BoxLunch:
		monitorInfoSchema = "boxlunchMonitorInfos"
		singleMonitorInfoSchema = "boxlunchSingleMonitorInfos"
	case enums.Disney:
		monitorInfoSchema = "disneyMonitorInfos"
		singleMonitorInfoSchema = "disneySingleMonitorInfos"
	case enums.GameStop:
		monitorInfoSchema = "gamestopMonitorInfos"
		singleMonitorInfoSchema = "gamestopSingleMonitorInfos"
	case enums.HotTopic:
		monitorInfoSchema = "hottopicMonitorInfos"
		singleMonitorInfoSchema = "hottopicSingleMonitorInfos"
	case enums.Newegg:
		monitorInfoSchema = "neweggMonitorInfos"
		singleMonitorInfoSchema = "neweggSingleMonitorInfos"
	case enums.PokemonCenter:
		monitorInfoSchema = "pokemoncenterMonitorInfos"
		singleMonitorInfoSchema = "pokemoncenterSingleMonitorInfos"
	case enums.Shopify:
		monitorInfoSchema = "shopifyMonitorInfos"
		singleMonitorInfoSchema = "shopifySingleMonitorInfos"
	case enums.Target:
		monitorInfoSchema = "targetMonitorInfos"
		singleMonitorInfoSchema = "targetSingleMonitorInfos"
	case enums.Topps:
		monitorInfoSchema = "toppsMonitorInfos"
		singleMonitorInfoSchema = "toppsSingleMonitorInfos"
	case enums.Walmart:
		monitorInfoSchema = "walmartMonitorInfos"

	}

	database := common.GetDatabase()
	if database == nil {
		return errors.New("database not initialized")
	}

	statement, err := database.Preparex(`DELETE FROM ` + monitorInfoSchema + ` WHERE taskGroupID = @p1`)
	if err != nil {
		return err
	}

	_, err = statement.Exec(groupID)
	if err != nil {
		return err
	}

	if monitorInfoSchema == "walmartMonitorInfos" {
		return nil
	}

	statement, err = database.Preparex(`DELETE FROM ` + singleMonitorInfoSchema + ` WHERE taskGroupID = @p1`)
	if err != nil {
		return err
	}
	_, err = statement.Exec(groupID)
	if err != nil {
		return err
	}

	return nil
}

func CreateTaskInfos(task entities.Task) error {
	database := common.GetDatabase()
	if database == nil {
		return errors.New("database not initialized")
	}

	switch task.TaskRetailer {
	case enums.Amazon:
		statement, err := database.Preparex(`INSERT INTO amazonTaskInfos (taskID, taskGroupID, email, password, loginType) VALUES (?, ?, ?, ?, ?)`)
		if err != nil {
			return err
		}

		encryptedEmail, err := common.Aes256Encrypt(task.AmazonTaskInfo.Email, enums.UserKey)
		if err != nil {
			return err
		}

		encryptedPassword, err := common.Aes256Encrypt(task.AmazonTaskInfo.Password, enums.UserKey)
		if err != nil {
			return err
		}

		_, err = statement.Exec(task.ID, task.TaskGroupID, encryptedEmail, encryptedPassword, task.AmazonTaskInfo.LoginType)
		if err != nil {
			return err
		}

	case enums.BestBuy:
		statement, err := database.Preparex(`INSERT INTO bestbuyTaskInfos (taskID, taskGroupID, email, password, locationID, taskType) VALUES (?, ?, ?, ?, ?, ?)`)
		if err != nil {
			return err
		}

		encryptedEmail, err := common.Aes256Encrypt(task.BestbuyTaskInfo.Email, enums.UserKey)
		if err != nil {
			return err
		}

		encryptedPassword, err := common.Aes256Encrypt(task.BestbuyTaskInfo.Password, enums.UserKey)
		if err != nil {
			return err
		}

		_, err = statement.Exec(task.ID, task.TaskGroupID, encryptedEmail, encryptedPassword, task.BestbuyTaskInfo.LocationID, task.BestbuyTaskInfo.TaskType)
		if err != nil {
			return err
		}

	case enums.BoxLunch:
		statement, err := database.Preparex(`INSERT INTO boxlunchTaskInfos (taskID, taskGroupID) VALUES (?, ?)`)
		if err != nil {
			return err
		}
		_, err = statement.Exec(task.ID, task.TaskGroupID)
		if err != nil {
			return err
		}

	case enums.Disney:
		statement, err := database.Preparex(`INSERT INTO disneyTaskInfos (taskID, taskGroupID, email, password, taskType) VALUES (?, ?, ?, ?, ?)`)
		if err != nil {
			return err
		}

		encryptedEmail, err := common.Aes256Encrypt(task.DisneyTaskInfo.Email, enums.UserKey)
		if err != nil {
			return err
		}

		encryptedPassword, err := common.Aes256Encrypt(task.DisneyTaskInfo.Password, enums.UserKey)
		if err != nil {
			return err
		}

		_, err = statement.Exec(task.ID, task.TaskGroupID, encryptedEmail, encryptedPassword, task.DisneyTaskInfo.TaskType)
		if err != nil {
			return err
		}

	case enums.GameStop:
		statement, err := database.Preparex(`INSERT INTO gamestopTaskInfos (taskID, taskGroupID, email, password, taskType) VALUES (?, ?, ?, ?, ?)`)
		if err != nil {
			return err
		}

		encryptedEmail, err := common.Aes256Encrypt(task.GamestopTaskInfo.Email, enums.UserKey)
		if err != nil {
			return err
		}

		encryptedPassword, err := common.Aes256Encrypt(task.GamestopTaskInfo.Password, enums.UserKey)
		if err != nil {
			return err
		}

		_, err = statement.Exec(task.ID, task.TaskGroupID, encryptedEmail, encryptedPassword, task.GamestopTaskInfo.TaskType)
		if err != nil {
			return err
		}

	case enums.HotTopic:
		statement, err := database.Preparex(`INSERT INTO hottopicTaskInfos (taskID, taskGroupID) VALUES (?, ?)`)
		if err != nil {
			return err
		}
		_, err = statement.Exec(task.ID, task.TaskGroupID)
		if err != nil {
			return err
		}

	case enums.Newegg:
		statement, err := database.Preparex(`INSERT INTO neweggTaskInfos (taskID, taskGroupID) VALUES (?, ?)`)
		if err != nil {
			return err
		}
		_, err = statement.Exec(task.ID, task.TaskGroupID)
		if err != nil {
			return err
		}

	case enums.PokemonCenter:
		statement, err := database.Preparex(`INSERT INTO pokemoncenterTaskInfos (taskID, taskGroupID, email, password, taskType, addressType) VALUES (?, ?, ?, ?, ?, ?)`)
		if err != nil {
			return err
		}

		encryptedEmail, err := common.Aes256Encrypt(task.PokemonCenterTaskInfo.Email, enums.UserKey)
		if err != nil {
			return err
		}

		encryptedPassword, err := common.Aes256Encrypt(task.PokemonCenterTaskInfo.Password, enums.UserKey)
		if err != nil {
			return err
		}

		_, err = statement.Exec(task.ID, task.TaskGroupID, encryptedEmail, encryptedPassword, task.PokemonCenterTaskInfo.TaskType, task.PokemonCenterTaskInfo.AddressType)
		if err != nil {
			return err
		}

	case enums.Shopify:
		statement, err := database.Preparex(`INSERT INTO shopifyTaskInfos (taskID, taskGroupID, couponCode, siteURL, sitePassword, shopifyRetailer) VALUES (?, ?, ?, ?, ?, ?)`)
		if err != nil {
			return err
		}
		_, err = statement.Exec(task.ID, task.TaskGroupID, task.ShopifyTaskInfo.CouponCode, task.ShopifyTaskInfo.SiteURL, task.ShopifyTaskInfo.SitePassword, task.ShopifyTaskInfo.ShopifyRetailer)
		if err != nil {
			return err
		}
		switch task.ShopifyTaskInfo.ShopifyRetailer {
		case enums.HotWheels:
			statement, err := database.Preparex(`INSERT INTO hotwheelsTaskInfos (taskID, taskGroupID, email, password) VALUES (?, ?, ?, ?)`)
			if err != nil {
				return err
			}

			encryptedEmail, err := common.Aes256Encrypt(task.ShopifyTaskInfo.HotWheelsTaskInfo.Email, enums.UserKey)
			if err != nil {
				return err
			}

			encryptedPassword, err := common.Aes256Encrypt(task.ShopifyTaskInfo.HotWheelsTaskInfo.Password, enums.UserKey)
			if err != nil {
				return err
			}

			_, err = statement.Exec(task.ID, task.TaskGroupID, encryptedEmail, encryptedPassword)
			if err != nil {
				return err
			}
		}

	case enums.Target:
		statement, err := database.Preparex(`INSERT INTO targetTaskInfos (taskID, taskGroupID, checkoutType, email, password, paymentType) VALUES (?, ?, ?, ?, ?, ?)`)
		if err != nil {
			return err
		}

		encryptedEmail, err := common.Aes256Encrypt(task.TargetTaskInfo.Email, enums.UserKey)
		if err != nil {
			return err
		}

		encryptedPassword, err := common.Aes256Encrypt(task.TargetTaskInfo.Password, enums.UserKey)
		if err != nil {
			return err
		}

		_, err = statement.Exec(task.ID, task.TaskGroupID, task.TargetTaskInfo.CheckoutType, encryptedEmail, encryptedPassword, task.TargetTaskInfo.PaymentType)
		if err != nil {
			return err
		}

	case enums.Topps:
		statement, err := database.Preparex(`INSERT INTO toppsTaskInfos (taskID, taskGroupID, email, password, taskType) VALUES (?, ?, ?, ?, ?)`)
		if err != nil {
			return err
		}
		_, err = statement.Exec(task.ID, task.TaskGroupID, task.ToppsTaskInfo.Email, task.ToppsTaskInfo.Password, task.ToppsTaskInfo.TaskType)
		if err != nil {
			return err
		}

	case enums.Walmart:
		statement, err := database.Preparex(`INSERT INTO walmartTaskInfos (taskID, taskGroupID) VALUES (?, ?)`)
		if err != nil {
			return err
		}
		_, err = statement.Exec(task.ID, task.TaskGroupID)
		if err != nil {
			return err
		}

	}
	return nil
}
func DeleteTaskInfos(taskID string, retailer enums.Retailer) error {
	var taskInfoSchema string

	switch retailer {
	case enums.Amazon:
		taskInfoSchema = "amazonTaskInfos"
	case enums.BestBuy:
		taskInfoSchema = "bestbuyTaskInfos"
	case enums.BoxLunch:
		taskInfoSchema = "boxlunchTaskInfos"
	case enums.Disney:
		taskInfoSchema = "disneyTaskInfos"
	case enums.GameStop:
		taskInfoSchema = "gamestopTaskInfos"
	case enums.HotTopic:
		taskInfoSchema = "hottopicTaskInfos"
	case enums.Newegg:
		taskInfoSchema = "neweggTaskInfos"
	case enums.PokemonCenter:
		taskInfoSchema = "pokemoncenterTaskInfos"
	case enums.Shopify:
		taskInfoSchema = "shopifyTaskInfos"
	case enums.Target:
		taskInfoSchema = "targetTaskInfos"
	case enums.Topps:
		taskInfoSchema = "toppsTaskInfos"
	case enums.Walmart:
		taskInfoSchema = "walmartTaskInfos"

	}
	if taskInfoSchema == "" {
		return nil
	}
	database := common.GetDatabase()
	if database == nil {
		return errors.New("database not initialized")
	}

	statement, err := database.Preparex(`DELETE FROM ` + taskInfoSchema + ` WHERE taskID = @p1`)
	if err != nil {
		return err
	}

	_, err = statement.Exec(taskID)
	if err != nil {
		return err
	}

	if retailer == enums.Shopify {
		statement, err := database.Preparex(`DELETE FROM hotwheelsTaskInfos WHERE taskID = @p1`)
		if err != nil {
			return err
		}

		_, err = statement.Exec(taskID)
		if err != nil {
			return err
		}
	}

	return err
}

func CreateShippingAddresses(profile entities.Profile) error {
	database := common.GetDatabase()
	if database == nil {
		return errors.New("database not initialized")
	}

	encryptedValues, err := common.EncryptValues(enums.UserKey, profile.ShippingAddress.FirstName, profile.ShippingAddress.LastName, profile.ShippingAddress.Address1, profile.ShippingAddress.Address2, profile.ShippingAddress.City, profile.ShippingAddress.ZipCode, profile.ShippingAddress.StateCode, profile.ShippingAddress.CountryCode)
	if err != nil {
		return err
	}

	statement, err := database.Preparex(`INSERT INTO shippingAddresses (ID, profileID, firstName, lastName, address1, address2, city, zipCode, stateCode, countryCode) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`)
	if err != nil {
		return err
	}

	_, err = statement.Exec(profile.ShippingAddress.ID, profile.ID, encryptedValues[0], encryptedValues[1], encryptedValues[2], encryptedValues[3], encryptedValues[4], encryptedValues[5], encryptedValues[6], encryptedValues[7])
	if err != nil {
		return err
	}

	return err
}
func CreateBillingAddresses(profile entities.Profile) error {
	database := common.GetDatabase()
	if database == nil {
		return errors.New("database not initialized")
	}

	encryptedValues, err := common.EncryptValues(enums.UserKey, profile.BillingAddress.FirstName, profile.BillingAddress.LastName, profile.BillingAddress.Address1, profile.BillingAddress.Address2, profile.BillingAddress.City, profile.BillingAddress.ZipCode, profile.BillingAddress.StateCode, profile.BillingAddress.CountryCode)
	if err != nil {
		return err
	}

	statement, err := database.Preparex(`INSERT INTO billingAddresses (ID, profileID, firstName, lastName, address1, address2, city, zipCode, stateCode, countryCode) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`)
	if err != nil {
		return err
	}

	_, err = statement.Exec(profile.BillingAddress.ID, profile.ID, encryptedValues[0], encryptedValues[1], encryptedValues[2], encryptedValues[3], encryptedValues[4], encryptedValues[5], encryptedValues[6], encryptedValues[7])
	if err != nil {
		return err
	}

	return err
}
func CreateCards(profile entities.Profile) error {
	database := common.GetDatabase()
	if database == nil {
		return errors.New("database not initialized")
	}

	encryptedValues, err := common.EncryptValues(enums.UserKey, profile.CreditCard.CardholderName, profile.CreditCard.CardNumber, profile.CreditCard.ExpMonth, profile.CreditCard.ExpYear, profile.CreditCard.CVV, profile.CreditCard.CardType)
	if err != nil {
		return err
	}

	statement, err := database.Preparex(`INSERT INTO cards (ID, profileID, cardHolderName, cardNumber, expMonth, expYear, cvv, cardType) VALUES (?, ?, ?, ?, ?, ?, ?, ?)`)
	if err != nil {
		return err
	}

	_, err = statement.Exec(profile.CreditCard.ID, profile.ID, encryptedValues[0], encryptedValues[1], encryptedValues[2], encryptedValues[3], encryptedValues[4], encryptedValues[5])
	if err != nil {
		return err
	}

	return err
}

func CreateProfileInfos(profile entities.Profile) error {
	err := CreateShippingAddresses(profile)
	if err != nil {
		return err
	}
	err = CreateBillingAddresses(profile)
	if err != nil {
		return err
	}
	return CreateCards(profile)
}

func DeleteShippingAddresses(ID string) error {
	database := common.GetDatabase()
	if database == nil {
		return errors.New("database not initialized")
	}

	statement, err := database.Preparex(`DELETE FROM shippingAddresses WHERE profileID = @p1`)
	if err != nil {
		return err
	}
	_, err = statement.Exec(ID)

	return err
}

func DeleteBillingAddresses(ID string) error {
	database := common.GetDatabase()
	if database == nil {
		return errors.New("database not initialized")
	}

	statement, err := database.Preparex(`DELETE FROM billingAddresses WHERE profileID = @p1`)
	if err != nil {
		return err
	}
	_, err = statement.Exec(ID)

	return err
}

func DeleteCards(ID string) error {
	database := common.GetDatabase()
	if database == nil {
		return errors.New("database not initialized")
	}

	statement, err := database.Preparex(`DELETE FROM cards WHERE profileID = @p1`)
	if err != nil {
		return err
	}
	_, err = statement.Exec(ID)

	return err
}

func DeleteProfileInfos(ID string) error {
	err := DeleteShippingAddresses(ID)
	if err != nil {
		return err
	}
	err = DeleteBillingAddresses(ID)
	if err != nil {
		return err
	}
	return DeleteCards(ID)
}
