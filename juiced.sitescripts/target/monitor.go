package target

import (
	"strings"
	"time"

	"backend.juicedbot.io/juiced.infrastructure/common/entities"
	"backend.juicedbot.io/juiced.infrastructure/common/enums"
	"backend.juicedbot.io/juiced.infrastructure/common/events"
	"backend.juicedbot.io/juiced.sitescripts/base"
	"backend.juicedbot.io/juiced.sitescripts/util"
)

// CreateTargetMonitor takes a TaskGroup entity and turns it into a Target Monitor
func CreateTargetMonitor(taskGroup *entities.TaskGroup, proxy entities.Proxy, eventBus *events.EventBus, monitor entities.TargetMonitorInfo) (Monitor, error) {
	storedTargetMonitors := make(map[string]entities.TargetSingleMonitorInfo)
	targetMonitor := Monitor{}
	tcins := []string{}
	for _, monitor := range monitor.Monitors {
		storedTargetMonitors[monitor.TCIN] = monitor
		tcins = append(tcins, monitor.TCIN)
	}

	client, err := util.CreateClient(proxy)
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
		TCINs:         tcins,
		StoreID:       monitor.StoreID,
		TCINsWithInfo: storedTargetMonitors,
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
	// If the function panics due to a runtime error, recover from it
	defer func() {
		recover()
		// TODO @silent: Let the UI know that a monitor failed
	}()

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
	resp, _, err := util.MakeRequest(&util.Request{
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

	switch resp.StatusCode {
	case 200:
		for _, product := range getTCINStockResponse.Data.ProductSummaries {
			if product.Fulfillment.ShippingOptions.AvailabilityStatus == "IN_STOCK" && monitor.CheckPrice(product.TCIN) {
				TCINWithType := product.TCIN + "|" + monitor.TCINsWithInfo[product.TCIN].CheckoutType
				inStockForShip = append(inStockForShip, TCINWithType)
			} else {
				outOfStockForShip = append(outOfStockForShip, product.TCIN)
			}
			for _, store := range product.Fulfillment.StoreOptions {
				if store.OrderPickup.AvailabilityStatus == "IN_STOCK" && store.LocationID == monitor.StoreID && monitor.CheckPrice(product.TCIN) {
					TCINWithType := product.TCIN + "|" + monitor.TCINsWithInfo[product.TCIN].CheckoutType
					inStockForPickup = append(inStockForPickup, TCINWithType)
				} else {
					outOfStockForPickup = append(outOfStockForPickup, product.TCIN)
				}
			}
			switch monitor.TCINsWithInfo[product.TCIN].CheckoutType {
			case enums.CheckoutTypeSHIP:
				inStockForPickup = inStockForPickup[:0]
			case enums.CheckoutTypePICKUP:
				inStockForShip = inStockForShip[:0]
			}
		}
	}

	return inStockForShip, outOfStockForShip, inStockForPickup, outOfStockForPickup
}

func (monitor *Monitor) CheckPrice(sku string) bool {

	params := util.CreateParams(map[string]string{
		"key":                             "ff457966e64d5e877fdbad070f276d18ecec4a01",
		"tcin":                            sku,
		"store_id":                        monitor.StoreID,
		"has_store_id":                    "true",
		"pricing_store_id":                monitor.StoreID,
		"has_pricing_store_id":            "true",
		"scheduled_delivery_store_id":     monitor.StoreID,
		"has_scheduled_delivery_store_id": "true",
		"has_financing_options":           "true",
	})

	checkPriceResponse := CheckPriceResponse{}

	resp, _, err := util.MakeRequest(&util.Request{
		Client:             monitor.Monitor.Client,
		Method:             "GET",
		URL:                CheckPriceEndpoint + params,
		AddHeadersFunction: AddTargetHeaders,
		Referer:            CheckPriceReferer + sku,
		ResponseBodyStruct: &checkPriceResponse,
	})
	if err != nil || resp.StatusCode != 200 {
		return false
	}

	return monitor.TCINsWithInfo[sku].MaxPrice > int(checkPriceResponse.Data.Product.Price.CurrentRetail)
}

// SendToTasks sends the product info to tasks
func (monitor *Monitor) SendToTasks(inStockForShip []string, inStockForPickup []string) {
	data := events.TargetStockData{InStockForShip: inStockForShip, InStockForPickup: inStockForPickup}
	monitor.Monitor.EventBus.PublishProductEvent(enums.Target, data, monitor.Monitor.TaskGroup.GroupID)
}
