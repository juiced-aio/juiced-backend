package target

import (
	"math/rand"

	"strings"
	"time"

	"backend.juicedbot.io/juiced.client/client"
	"backend.juicedbot.io/juiced.infrastructure/common"
	"backend.juicedbot.io/juiced.infrastructure/common/entities"
	"backend.juicedbot.io/juiced.infrastructure/common/enums"
	"backend.juicedbot.io/juiced.infrastructure/common/events"
	"backend.juicedbot.io/juiced.sitescripts/base"
	"backend.juicedbot.io/juiced.sitescripts/util"
	cmap "github.com/orcaman/concurrent-map"
)

// CreateTargetMonitor takes a TaskGroup entity and turns it into a Target Monitor
func CreateTargetMonitor(taskGroup *entities.TaskGroup, proxies []entities.Proxy, eventBus *events.EventBus, monitor *entities.TargetMonitorInfo) (Monitor, error) {
	storedTargetMonitors := make(map[string]entities.TargetSingleMonitorInfo)
	targetMonitor := Monitor{}
	tcins := []string{}
	for _, monitor := range monitor.Monitors {
		storedTargetMonitors[monitor.TCIN] = monitor
		tcins = append(tcins, monitor.TCIN)
	}

	targetMonitor = Monitor{
		Monitor: base.Monitor{
			TaskGroup: taskGroup,
			Proxies:   proxies,
			EventBus:  eventBus,
		},
		TCINs:            tcins,
		StoreID:          monitor.StoreID,
		TCINsWithInfo:    storedTargetMonitors,
		InStockForShip:   cmap.New(),
		InStockForPickup: cmap.New(),
		MonitorType:      monitor.MonitorType,
	}
	return targetMonitor, nil
}

// PublishEvent wraps the EventBus's PublishMonitorEvent function
func (monitor *Monitor) PublishEvent(status enums.MonitorStatus, eventType enums.MonitorEventType, data interface{}) {
	monitor.Monitor.TaskGroup.SetMonitorStatus(status)
	monitor.Monitor.EventBus.PublishMonitorEvent(status, eventType, data, monitor.Monitor.TaskGroup.GroupID)
}

// CheckForStop checks the stop flag and stops the monitor if it's true
func (monitor *Monitor) CheckForStop() bool {
	if monitor.Monitor.StopFlag {
		monitor.PublishEvent(enums.MonitorIdle, enums.MonitorStop, nil)
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
		if recover() != nil {
			monitor.Monitor.StopFlag = true
			monitor.PublishEvent(enums.MonitorIdle, enums.MonitorFail, nil)
		}
		monitor.PublishEvent(enums.MonitorIdle, enums.MonitorComplete, nil)
	}()

	if monitor.Monitor.TaskGroup.MonitorStatus == enums.MonitorIdle {
		monitor.PublishEvent(enums.WaitingForProductData, enums.MonitorStart, nil)
	}
	needToStop := monitor.CheckForStop()
	if needToStop {
		return
	}

	if monitor.Monitor.Client.Transport == nil {
		monitorClient, err := util.CreateClient()
		if err != nil {
			return
		}
		monitor.Monitor.Client = monitorClient

	}
	if len(monitor.Monitor.Proxies) > 0 {
		client.UpdateProxy(&monitor.Monitor.Client, common.ProxyCleaner(monitor.Monitor.Proxies[rand.Intn(len(monitor.Monitor.Proxies))]))
	}

	stockData := TargetStockData{}
	switch monitor.MonitorType {
	case enums.SKUMonitor:
		stockData = monitor.GetTCINStock()
	case enums.URLMonitor:
		// inStockForShip, inStockForPickup = monitor.GetURLStock()
	case enums.KeywordMonitor:
		// inStockForShip, inStockForPickup = monitor.GetKeywordStock()
	default:
		stockData = monitor.GetTCINStock()
	}

	somethingInStock := false
	if len(stockData.InStockForShip) > 0 {
		somethingInStock = true
	}
	if monitor.StoreID != "" && len(stockData.InStockForPickup) > 0 {
		somethingInStock = true
	}

	if somethingInStock {
		needToStop := monitor.CheckForStop()
		if needToStop {
			return
		}

		var productInfo events.ProductInfo

		for _, singleStockData := range stockData.InStockForShip {
			monitor.InStockForShip.Set(singleStockData.TCIN, singleStockData)
			productInfo.Products = append(productInfo.Products, events.Product{
				ProductName:     singleStockData.ProductName,
				ProductImageURL: singleStockData.ProductImageURL,
			})
		}

		for _, singleStockData := range stockData.InStockForPickup {
			monitor.InStockForPickup.Set(singleStockData.TCIN, singleStockData)
			productInfo.Products = append(productInfo.Products, events.Product{
				ProductName:     singleStockData.ProductName,
				ProductImageURL: singleStockData.ProductImageURL,
			})
		}

		monitor.PublishEvent(enums.SendingProductInfoToTasks, enums.MonitorUpdate, productInfo)

	} else {
		if len(stockData.OutOfStockForShip) > 0 || len(stockData.OutOfStockForPickup) > 0 {
			if monitor.Monitor.TaskGroup.MonitorStatus != enums.WaitingForInStock {
				monitor.PublishEvent(enums.WaitingForInStock, enums.MonitorUpdate, nil)
			}
		}

		time.Sleep(time.Duration(monitor.Monitor.TaskGroup.MonitorDelay) * time.Millisecond)
		monitor.RunMonitor()
	}
}

// GetTCINStock returns a map of in stock TCINs given a list of TCINs joined by commas
func (monitor *Monitor) GetTCINStock() TargetStockData {
	targetStockData := TargetStockData{}

	getTCINStockRequest := map[string]string{}
	if monitor.StoreID != "" {
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
	})
	if err != nil {
		return targetStockData
	}

	switch resp.StatusCode {
	case 200:

		// For Ship
		for _, product := range getTCINStockResponse.Data.ProductSummaries {
			if product.Fulfillment.ShippingOptions.AvailabilityStatus == "IN_STOCK" || product.Fulfillment.ShippingOptions.AvailabilityStatus == "LIMITED_STOCK" || product.Fulfillment.ShippingOptions.AvailabilityStatus == "PRE_ORDER_SELLABLE" {
				productName, productImageURL, inBudget := monitor.GetTCINInfo(product.TCIN)
				if inBudget {
					if ok := monitor.InStockForShip.Has(product.TCIN); !ok {
						targetStockData.InStockForShip = append(targetStockData.InStockForShip, SingleStockData{
							TCIN:            product.TCIN,
							TCINType:        monitor.TCINsWithInfo[product.TCIN].CheckoutType,
							ProductName:     productName,
							ProductImageURL: productImageURL,
						})
					}
				} else {
					monitor.InStockForShip.Remove(product.TCIN)
					targetStockData.OutOfStockForShip = append(targetStockData.OutOfStockForShip, SingleStockData{TCIN: product.TCIN})
				}
			} else {
				monitor.InStockForShip.Remove(product.TCIN)
				targetStockData.OutOfStockForShip = append(targetStockData.OutOfStockForShip, SingleStockData{TCIN: product.TCIN})
			}

			// For Pickup
			for _, store := range product.Fulfillment.StoreOptions {
				if store.OrderPickup.AvailabilityStatus == "IN_STOCK" || store.OrderPickup.AvailabilityStatus == "LIMITED_STOCK" || store.OrderPickup.AvailabilityStatus == "PRE_ORDER_SELLABLE" && store.LocationID == monitor.StoreID {
					productName, productImageURL, inBudget := monitor.GetTCINInfo(product.TCIN)
					if inBudget {
						if ok := monitor.InStockForPickup.Has(product.TCIN); !ok {
							targetStockData.InStockForPickup = append(targetStockData.InStockForPickup, SingleStockData{
								TCIN:            product.TCIN,
								TCINType:        monitor.TCINsWithInfo[product.TCIN].CheckoutType,
								ProductName:     productName,
								ProductImageURL: productImageURL,
							})
						}
					} else {
						monitor.InStockForPickup.Remove(product.TCIN)
						targetStockData.OutOfStockForPickup = append(targetStockData.OutOfStockForPickup, SingleStockData{TCIN: product.TCIN})
					}
				} else {
					monitor.InStockForPickup.Remove(product.TCIN)
					targetStockData.OutOfStockForPickup = append(targetStockData.OutOfStockForPickup, SingleStockData{TCIN: product.TCIN})
				}
			}

			switch monitor.TCINsWithInfo[product.TCIN].CheckoutType {
			case enums.CheckoutTypeSHIP:
				targetStockData.InStockForPickup = targetStockData.InStockForPickup[:0]
			case enums.CheckoutTypePICKUP:
				targetStockData.InStockForShip = targetStockData.InStockForShip[:0]
			}
		}
	}

	return targetStockData
}

func (monitor *Monitor) GetTCINInfo(sku string) (string, string, bool) {
	var storeID string
	storeID = monitor.StoreID
	if monitor.StoreID == "" {
		storeID = "199"
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

	getTCINInfoResponse := GetTCINInfoResponse{}

	resp, _, err := util.MakeRequest(&util.Request{
		Client:             monitor.Monitor.Client,
		Method:             "GET",
		URL:                TCINInfoEndpoint + params,
		AddHeadersFunction: AddTargetHeaders,
		Referer:            TCINInfoReferer + sku,
		ResponseBodyStruct: &getTCINInfoResponse,
	})
	if err != nil || resp.StatusCode != 200 {
		return "", "", false
	}

	return getTCINInfoResponse.Data.Product.Item.ProductDescription.Title, getTCINInfoResponse.Data.Product.Item.Enrichment.Images.PrimaryImageURL, monitor.TCINsWithInfo[sku].MaxPrice >= int(getTCINInfoResponse.Data.Product.Price.CurrentRetail) || monitor.TCINsWithInfo[sku].MaxPrice == -1
}
