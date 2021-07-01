package target

import (
	"fmt"
	"strings"
	"time"

	"backend.juicedbot.io/juiced.infrastructure/common"
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
		MonitorType:   monitor.MonitorType,
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
		if recover() != "" {
			monitor.Monitor.StopFlag = true
			monitor.PublishEvent(enums.MonitorIdle, enums.MonitorFail)
		}
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
	default:
		inStockForShip, outOfStockForShip, inStockForPickup, outOfStockForPickup = monitor.GetTCINStock()
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

		monitor.InStockForShip = inStockForShip
		monitor.InStockForPickup = inStockForPickup
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
	getTCINStockRequest := map[string]string{}
	if monitor.StoreID != "" {
		fmt.Println(monitor.StoreID)
		getTCINStockRequest = GetTCINStockRequestToMap(GetTCINStockRequest{
			Key:                      "ff457966e64d5e877fdbad070f276d18ecec4a01",
			TCINs:                    strings.Join(monitor.TCINs, ","),
			StoreID:                  monitor.StoreID,
			ScheduledDeliveryStoreID: monitor.StoreID,
		})
	} else {
		getTCINStockRequest["key"] = "ff457966e64d5e877fdbad070f276d18ecec4a01"
		getTCINStockRequest["tcins"] = strings.Join(monitor.TCINs, ",")
	}

	getTCINStockResponse := GetTCINStockResponse{}

	params := util.CreateParams(getTCINStockRequest)
	resp, _, err := util.MakeRequest(&util.Request{
		Client:             monitor.Monitor.Client,
		Method:             "GET",
		URL:                GetTCINStockEndpoint + params,
		AddHeadersFunction: AddTargetHeaders,
		Referer:            GetTCINStockReferer,
		ResponseBodyStruct: &getTCINStockResponse,
		Task:               base.Task{},
		Monitor:            monitor.Monitor,
	})
	if err != nil {
		return inStockForShip, outOfStockForShip, inStockForPickup, outOfStockForPickup
	}

	switch resp.StatusCode {
	case 200:
		for _, product := range getTCINStockResponse.Data.ProductSummaries {
			TCINWithType := product.TCIN + "|" + monitor.TCINsWithInfo[product.TCIN].CheckoutType
			if product.Fulfillment.ShippingOptions.AvailabilityStatus == "IN_STOCK" && monitor.CheckPrice(product.TCIN) {
				if !common.InSlice(monitor.InStockForShip, TCINWithType) {
					inStockForShip = append(inStockForShip, TCINWithType)
				}
			} else {
				outOfStockForShip = append(outOfStockForShip, product.TCIN)
				monitor.InStockForShip = common.RemoveFromSlice(monitor.InStockForShip, TCINWithType)
			}
			for _, store := range product.Fulfillment.StoreOptions {
				if store.OrderPickup.AvailabilityStatus == "IN_STOCK" && store.LocationID == monitor.StoreID && monitor.CheckPrice(product.TCIN) {
					TCINWithType := product.TCIN + "|" + monitor.TCINsWithInfo[product.TCIN].CheckoutType
					inStockForPickup = append(inStockForPickup, TCINWithType)
				} else {
					outOfStockForPickup = append(outOfStockForPickup, product.TCIN)
					monitor.InStockForPickup = common.RemoveFromSlice(monitor.InStockForPickup, TCINWithType)
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
	var storeID string
	storeID = monitor.StoreID
	if monitor.StoreID == "" {
		storeID = "1"
	}
	params := util.CreateParams(map[string]string{
		"key":                             "ff457966e64d5e877fdbad070f276d18ecec4a01",
		"tcin":                            sku,
		"store_id":                        storeID,
		"has_store_id":                    "true",
		"pricing_store_id":                storeID,
		"has_pricing_store_id":            "true",
		"scheduled_delivery_store_id":     storeID,
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
		Task:               base.Task{},
		Monitor:            monitor.Monitor,
	})
	if err != nil || resp.StatusCode != 200 {
		return false
	}

	return monitor.TCINsWithInfo[sku].MaxPrice > int(checkPriceResponse.Data.Product.Price.CurrentRetail) || monitor.TCINsWithInfo[sku].MaxPrice == -1
}
