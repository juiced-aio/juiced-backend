package target

import (
	"strings"
	"time"

	"juiced/juiced.infrastructure/common/entities"
	"juiced/juiced.infrastructure/common/enums"
	"juiced/juiced.infrastructure/common/events"
	"juiced/juiced.sitescripts/base"
	"juiced/juiced.sitescripts/util"
)

// CreateTargetMonitor takes a TaskGroup entity and turns it into a Target Monitor
func CreateTargetMonitor(taskGroup *entities.TaskGroup, proxy entities.Proxy, eventBus *events.EventBus, monitorType enums.MonitorType, tcins []string, storeID string) (Monitor, error) {
	targetMonitor := Monitor{}
	client, err := util.CreateClient()
	if err != nil {
		return targetMonitor, err
	}
	targetMonitor = Monitor{
		Monitor: base.Monitor{
			TaskGroup: taskGroup,
			Proxy:     proxy,
			EventBus:  eventBus,
			Client:    client,
		},
		MonitorType: monitorType,
		TCINs:       tcins,
		StoreID:     storeID,
	}
	return targetMonitor, err
}

// PublishEvent wraps the EventBus's PublishMonitorEvent function
func (monitor *Monitor) PublishEvent(status enums.MonitorStatus, eventType enums.MonitorEventType) {
	monitor.Monitor.TaskGroup.SetMonitorStatus(status)
	monitor.Monitor.EventBus.PublishMonitorEvent(status, eventType, nil, monitor.Monitor.TaskGroup.GroupID)
}

// CheckForStop checks the stop flag and stops the monitor if it's true
func (monitor *Monitor) CheckForStop() bool {
	if monitor.Monitor.StopFlag {
		monitor.PublishEvent(enums.MonitorIdle, enums.MonitorStop)
		return true
	}
	return false
}

// Run is the script driver that calls all the individual requests
// Function order:
// 		1. Get___Stock (TCIN/URL/Keyword)
//		2. SendToTasks
func (monitor *Monitor) RunMonitor() {
	if monitor.Monitor.TaskGroup.MonitorStatus == enums.MonitorIdle {
		monitor.PublishEvent(enums.WaitingForProductData, enums.MonitorStart)
	}
	needToStop := monitor.CheckForStop()
	if needToStop {
		return
	}
	inStockForShip := make([]string, 0)
	outOfStockForShip := make([]string, 0)
	inStockForPickup := make([]string, 0)
	outOfStockForPickup := make([]string, 0)
	switch monitor.MonitorType {
	case enums.SKUMonitor:
		inStockForShip, outOfStockForShip, inStockForPickup, outOfStockForPickup = monitor.GetTCINStock()
	case enums.URLMonitor:
		// inStockForShip, inStockForPickup = monitor.GetURLStock()
	case enums.KeywordMonitor:
		// inStockForShip, inStockForPickup = monitor.GetKeywordStock()
	}

	somethingInStock := false
	if len(inStockForShip) > 0 {
		somethingInStock = true
	}
	if monitor.StoreID != "" && len(inStockForPickup) > 0 {
		somethingInStock = true
	}

	if somethingInStock {
		needToStop := monitor.CheckForStop()
		if needToStop {
			return
		}
		monitor.PublishEvent(enums.SendingProductInfoToTasks, enums.MonitorUpdate)
		monitor.SendToTasks(inStockForShip, inStockForPickup)
	} else {
		if len(outOfStockForShip) > 0 || len(outOfStockForPickup) > 0 {
			if monitor.Monitor.TaskGroup.MonitorStatus != enums.WaitingForInStock {
				monitor.PublishEvent(enums.WaitingForInStock, enums.MonitorUpdate)
			}
		}
		time.Sleep(time.Duration(monitor.Monitor.TaskGroup.MonitorDelay) * time.Millisecond)
		monitor.RunMonitor()
	}
}

// GetTCINStock returns a map of in stock TCINs given a list of TCINs joined by commas
func (monitor *Monitor) GetTCINStock() ([]string, []string, []string, []string) {
	inStockForShip := make([]string, 0)
	outOfStockForShip := make([]string, 0)
	inStockForPickup := make([]string, 0)
	outOfStockForPickup := make([]string, 0)

	getTCINStockRequest := GetTCINStockRequestToMap(GetTCINStockRequest{
		Key:                      "ff457966e64d5e877fdbad070f276d18ecec4a01",
		TCINs:                    strings.Join(monitor.TCINs, ","),
		StoreID:                  monitor.StoreID,
		ScheduledDeliveryStoreID: monitor.StoreID,
	})
	getTCINStockResponse := GetTCINStockResponse{}

	params := util.CreateParams(getTCINStockRequest)
	_, err := util.MakeRequest(&util.Request{
		Client:             monitor.Monitor.Client,
		Method:             "GET",
		URL:                GetTCINStockEndpoint + params,
		AddHeadersFunction: AddTargetHeaders,
		Referer:            GetTCINStockReferer,
		ResponseBodyStruct: &getTCINStockResponse,
	})
	if err != nil {
		return inStockForShip, outOfStockForShip, inStockForPickup, outOfStockForPickup
	}

	for _, product := range getTCINStockResponse.Data.ProductSummaries {
		if product.Fulfillment.ShippingOptions.AvailabilityStatus == "IN_STOCK" {
			inStockForShip = append(inStockForShip, product.TCIN)
		} else {
			outOfStockForShip = append(outOfStockForShip, product.TCIN)
		}
		for _, store := range product.Fulfillment.StoreOptions {
			if store.OrderPickup.AvailabilityStatus == "IN_STOCK" && store.LocationID == monitor.StoreID {
				inStockForPickup = append(inStockForPickup, product.TCIN)
			} else {
				outOfStockForPickup = append(outOfStockForPickup, product.TCIN)
			}
		}
	}

	return inStockForShip, outOfStockForShip, inStockForPickup, outOfStockForPickup
}

// SendToTasks sends the product info to tasks
func (monitor *Monitor) SendToTasks(inStockForShip []string, inStockForPickup []string) {
	data := events.TargetStockData{InStockForShip: inStockForShip, InStockForPickup: inStockForPickup}
	monitor.Monitor.EventBus.PublishProductEvent(enums.Target, data, monitor.Monitor.TaskGroup.GroupID)
}
