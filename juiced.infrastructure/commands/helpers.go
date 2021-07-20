package commands

import (
	"errors"
	"strings"

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
	case enums.Walmart:
		statement, err := database.Preparex(`INSERT INTO walmartMonitorInfos (ID, taskGroupID, monitorType, skusJoined, maxPrice) VALUES (?, ?, ?, ?, ?)`)
		if err != nil {
			return err
		}
		taskGroup.WalmartMonitorInfo.ID = monitorID
		taskGroup.WalmartMonitorInfo.TaskGroupID = taskGroup.GroupID
		taskGroup.WalmartMonitorInfo.SKUsJoined = strings.Join(taskGroup.WalmartMonitorInfo.SKUs, ",")
		_, err = statement.Exec(taskGroup.WalmartMonitorInfo.ID, taskGroup.WalmartMonitorInfo.TaskGroupID, taskGroup.WalmartMonitorInfo.MonitorType, taskGroup.WalmartMonitorInfo.SKUsJoined, taskGroup.WalmartMonitorInfo.MaxPrice)
		if err != nil {
			return err
		}
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
		taskGroup.BoxLunchMonitorInfo.ID = monitorID
		taskGroup.BoxLunchMonitorInfo.TaskGroupID = taskGroup.GroupID
		_, err = statement.Exec(taskGroup.BoxLunchMonitorInfo.ID, taskGroup.BoxLunchMonitorInfo.TaskGroupID)
		if err != nil {
			return err
		}
		for _, monitor := range taskGroup.BoxLunchMonitorInfo.Monitors {
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

	}

	return nil
}

func DeleteMonitorInfos(groupID string, retailer enums.Retailer) error {
	var monitorInfoSchema string
	var singleMonitorInfoSchema string
	switch retailer {
	case enums.Target:
		monitorInfoSchema = "targetMonitorInfos"
		singleMonitorInfoSchema = "targetSingleMonitorInfos"
	case enums.Walmart:
		monitorInfoSchema = "walmartMonitorInfos"
	case enums.Amazon:
		monitorInfoSchema = "amazonMonitorInfos"
		singleMonitorInfoSchema = "amazonSingleMonitorInfos"
	case enums.BestBuy:
		monitorInfoSchema = "bestbuyMonitorInfos"
		singleMonitorInfoSchema = "bestbuySingleMonitorInfos"
	case enums.BoxLunch:
		monitorInfoSchema = "boxlunchMonitorInfos"
		singleMonitorInfoSchema = "boxlunchSingleMonitorInfos"
	case enums.GameStop:
		monitorInfoSchema = "gamestopMonitorInfos"
		singleMonitorInfoSchema = "gamestopSingleMonitorInfos"
	case enums.HotTopic:
		monitorInfoSchema = "hottopicMonitorInfos"
		singleMonitorInfoSchema = "hottopicSingleMonitorInfos"
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
	case enums.Target:
		statement, err := database.Preparex(`INSERT INTO targetTaskInfos (taskID, taskGroupID, checkoutType, email, password, paymentType) VALUES (?, ?, ?, ?, ?, ?)`)
		if err != nil {
			return err
		}
		_, err = statement.Exec(task.ID, task.TaskGroupID, task.TargetTaskInfo.CheckoutType, task.TargetTaskInfo.Email, task.TargetTaskInfo.Password, task.TargetTaskInfo.PaymentType)
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
	case enums.Amazon:
		statement, err := database.Preparex(`INSERT INTO amazonTaskInfos (taskID, taskGroupID, email, password, loginType) VALUES (?, ?, ?, ?, ?)`)
		if err != nil {
			return err
		}
		_, err = statement.Exec(task.ID, task.TaskGroupID, task.AmazonTaskInfo.Email, task.AmazonTaskInfo.Password, task.AmazonTaskInfo.LoginType)
		if err != nil {
			return err
		}
	case enums.BestBuy:
		statement, err := database.Preparex(`INSERT INTO bestbuyTaskInfos (taskID, taskGroupID, email, password, taskType) VALUES (?, ?, ?, ?, ?)`)
		if err != nil {
			return err
		}
		_, err = statement.Exec(task.ID, task.TaskGroupID, task.BestbuyTaskInfo.Email, task.BestbuyTaskInfo.Password, task.BestbuyTaskInfo.TaskType)
		if err != nil {
			return err
		}

	case enums.BoxLunch:
		statement, err := database.Preparex(`INSERT INTO boxlunchTaskInfos (taskID, taskGroupID, pidsJoined) VALUES (?, ?, ?)`)
		if err != nil {
			return err
		}

		task.BoxLunchTaskInfo.PidsJoined = strings.Join(task.BoxLunchTaskInfo.Pids, ",")

		_, err = statement.Exec(task.ID, task.TaskGroupID, task.BoxLunchTaskInfo.PidsJoined)
		if err != nil {
			return err
		}

	case enums.GameStop:
		statement, err := database.Preparex(`INSERT INTO gamestopTaskInfos (taskID, taskGroupID, email, password, taskType) VALUES (?, ?, ?, ?, ?)`)
		if err != nil {
			return err
		}
		_, err = statement.Exec(task.ID, task.TaskGroupID, task.GamestopTaskInfo.Email, task.GamestopTaskInfo.Password, task.GamestopTaskInfo.TaskType)
		if err != nil {
			return err
		}
	case enums.HotTopic:
		task.HottopicTaskInfo.PidsJoined = strings.Join(task.HottopicTaskInfo.Pids, ",")
		statement, err := database.Preparex(`INSERT INTO hottopicTaskInfos (taskID, taskGroupID, pidsJoined) VALUES (?, ?, ?)`)
		if err != nil {
			return err
		}
		_, err = statement.Exec(task.ID, task.TaskGroupID, task.HottopicTaskInfo.PidsJoined)
		if err != nil {
			return err
		}
	}
	return nil
}
func DeleteTaskInfos(taskID string, retailer enums.Retailer) error {
	var taskInfoSchema string

	switch retailer {
	case enums.Target:
		taskInfoSchema = "targetTaskInfos"
	case enums.Walmart:
		taskInfoSchema = "walmartTaskInfos"
	case enums.Amazon:
		taskInfoSchema = "amazonTaskInfos"
	case enums.BestBuy:
		taskInfoSchema = "bestbuyTaskInfos"
	case enums.BoxLunch:
		taskInfoSchema = "boxlunchTaskInfos"
	case enums.GameStop:
		taskInfoSchema = "gamestopTaskInfos"
	case enums.HotTopic:
		taskInfoSchema = "hottopicTaskInfos"
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

	return err
}

func CreateShippingAddresses(profile entities.Profile) error {
	database := common.GetDatabase()
	if database == nil {
		return errors.New("database not initialized")
	}

	statement, err := database.Preparex(`INSERT INTO shippingAddresses (ID, profileID, firstName, lastName, address1, address2, city, zipCode, stateCode, countryCode) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`)
	if err != nil {
		return err
	}

	_, err = statement.Exec(profile.ShippingAddress.ID, profile.ID, profile.ShippingAddress.FirstName, profile.ShippingAddress.LastName, profile.ShippingAddress.Address1, profile.ShippingAddress.Address2, profile.ShippingAddress.City, profile.ShippingAddress.ZipCode, profile.ShippingAddress.StateCode, profile.ShippingAddress.CountryCode)
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

	statement, err := database.Preparex(`INSERT INTO billingAddresses (ID, profileID, firstName, lastName, address1, address2, city, zipCode, stateCode, countryCode) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`)
	if err != nil {
		return err
	}

	_, err = statement.Exec(profile.BillingAddress.ID, profile.ID, profile.BillingAddress.FirstName, profile.BillingAddress.LastName, profile.BillingAddress.Address1, profile.BillingAddress.Address2, profile.BillingAddress.City, profile.BillingAddress.ZipCode, profile.BillingAddress.StateCode, profile.BillingAddress.CountryCode)
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

	statement, err := database.Preparex(`INSERT INTO cards (ID, profileID, cardHolderName, cardNumber, expMonth, expYear, cvv, cardType) VALUES (?, ?, ?, ?, ?, ?, ?, ?)`)
	if err != nil {
		return err
	}

	_, err = statement.Exec(profile.CreditCard.ID, profile.ID, profile.CreditCard.CardholderName, profile.CreditCard.CardNumber, profile.CreditCard.ExpMonth, profile.CreditCard.ExpYear, profile.CreditCard.CVV, profile.CreditCard.CardType)
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
