package gamestop

import (
	"fmt"
	"strconv"
	"strings"
	"sync"
	"time"

	"backend.juicedbot.io/juiced.infrastructure/common"
	"backend.juicedbot.io/juiced.infrastructure/common/entities"
	"backend.juicedbot.io/juiced.infrastructure/common/enums"
	"backend.juicedbot.io/juiced.infrastructure/common/events"
	"backend.juicedbot.io/juiced.sitescripts/base"
	"backend.juicedbot.io/juiced.sitescripts/util"
)

// CreateGamestopMonitor takes a TaskGroup entity and turns it into a Gamestop Monitor
func CreateGamestopMonitor(taskGroup *entities.TaskGroup, proxyGroup *entities.ProxyGroup, eventBus *events.EventBus, singleMonitors []entities.GamestopSingleMonitorInfo) (Monitor, error) {
	storedGamestopMonitors := make(map[string]entities.GamestopSingleMonitorInfo)
	gamestopMonitor := Monitor{}
	skus := []string{}

	for _, monitor := range singleMonitors {
		storedGamestopMonitors[monitor.SKU] = monitor
		skus = append(skus, monitor.SKU)
	}

	gamestopMonitor = Monitor{
		Monitor: base.Monitor{
			TaskGroup:  taskGroup,
			ProxyGroup: proxyGroup,
			EventBus:   eventBus,
		},

		SKUs:        skus,
		SKUWithInfo: storedGamestopMonitors,
	}

	return gamestopMonitor, nil
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

func (monitor *Monitor) RunMonitor() {
	// If the function panics due to a runtime error, recover from it
	defer func() {
		if recover() != nil {
			monitor.Monitor.StopFlag = true
			monitor.PublishEvent(enums.MonitorIdle, enums.MonitorFail, nil)
		}
	}()

	if monitor.Monitor.TaskGroup.MonitorStatus == enums.MonitorIdle {
		monitor.PublishEvent(enums.WaitingForProductData, enums.MonitorStart, nil)
	}

	if monitor.Monitor.Client.Transport == nil {
		err := monitor.Monitor.CreateClient()
		if err != nil {
			return
		}

		var proxy *entities.Proxy
		if monitor.Monitor.ProxyGroup != nil {
			if len(monitor.Monitor.ProxyGroup.Proxies) > 0 {
				proxy = util.RandomLeastUsedProxy(monitor.Monitor.ProxyGroup.Proxies)
				monitor.Monitor.UpdateProxy(proxy)
			}
		}

		becameGuest := false
		for !becameGuest {
			needToStop := monitor.CheckForStop()
			if needToStop {
				return
			}
			becameGuest = BecomeGuest(&monitor.Monitor.Client)
			if !becameGuest {
				time.Sleep(1000 * time.Millisecond)
			}
		}
	}

	needToStop := monitor.CheckForStop()
	if needToStop {
		return
	}

	wg := sync.WaitGroup{}
	wg.Add(len(monitor.SKUs))
	for _, sku := range monitor.SKUs {
		go func(x string) {
			monitor.RunSingleMonitor(x)
			wg.Done()
		}(sku)
	}
	wg.Wait()

}

func (monitor *Monitor) RunSingleMonitor(sku string) {
	needToStop := monitor.CheckForStop()
	if needToStop {
		return
	}

	defer func() {
		if recover() != nil {
			time.Sleep(time.Duration(monitor.Monitor.TaskGroup.MonitorDelay) * time.Millisecond)
			monitor.RunSingleMonitor(sku)
		}
	}()

	var proxy *entities.Proxy
	if monitor.Monitor.ProxyGroup != nil {
		if len(monitor.Monitor.ProxyGroup.Proxies) > 0 {
			proxy = util.RandomLeastUsedProxy(monitor.Monitor.ProxyGroup.Proxies)
			monitor.Monitor.UpdateProxy(proxy)
		}
	}

	stockData := monitor.GetSKUStock(sku)
	if stockData.SKU != "" {
		needToStop := monitor.CheckForStop()
		if needToStop {
			return
		}

		var inSlice bool
		for _, monitorStock := range monitor.InStock {
			inSlice = monitorStock.SKU == stockData.SKU
		}
		if !inSlice {
			monitor.InStock = append(monitor.InStock, stockData)
			monitor.PublishEvent(enums.SendingProductInfoToTasks, enums.MonitorUpdate, events.ProductInfo{
				Products: []events.Product{
					{ProductName: stockData.ItemName, ProductImageURL: stockData.ImageURL}},
			})
		}
	} else {
		if stockData.OutOfPriceRange {
			if monitor.Monitor.TaskGroup.MonitorStatus != enums.OutOfPriceRange {
				monitor.PublishEvent(enums.OutOfPriceRange, enums.MonitorUpdate, events.ProductInfo{
					Products: []events.Product{
						{ProductName: stockData.ItemName, ProductImageURL: stockData.ImageURL}},
				})
			}
		} else {
			if monitor.Monitor.TaskGroup.MonitorStatus != enums.WaitingForInStock {
				monitor.PublishEvent(enums.WaitingForInStock, enums.MonitorUpdate, events.ProductInfo{
					Products: []events.Product{
						{ProductName: stockData.ItemName, ProductImageURL: stockData.ImageURL}},
				})
			}
		}
		for i, monitorStock := range monitor.InStock {
			if monitorStock.SKU == stockData.SKU {
				monitor.InStock = append(monitor.InStock[:i], monitor.InStock[i+1:]...)
				break
			}
		}
	}

	time.Sleep(time.Duration(monitor.Monitor.TaskGroup.MonitorDelay) * time.Millisecond)
	monitor.RunSingleMonitor(sku)
}

// Checks if the item is instock and fills the monitors EventInfo if so
func (monitor *Monitor) GetSKUStock(sku string) GamestopInStockData {
	stockData := GamestopInStockData{}
	monitorResponse := MonitorResponse{}
	resp, _, err := util.MakeRequest(&util.Request{
		Client: monitor.Monitor.Client,
		Method: "GET",
		URL:    fmt.Sprintf(MonitorEndpoint, sku),
		RawHeaders: [][2]string{
			{"sec-ch-ua", "\" Not A;Brand\";v=\"99\", \"Chromium\";v=\"90\", \"Google Chrome\";v=\"90\""},
			{"accept", "*/*"},
			{"x-requested-with", "XMLHttpRequest"},
			{"sec-ch-ua-mobile", "?0"},
			{"user-agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/90.0.4430.212 Safari/537.36"},
			{"sec-fetch-site", "same-origin"},
			{"sec-fetch-mode", "cors"},
			{"sec-fetch-dest", "empty"},
			{"referer", BaseEndpoint},
			{"accept-encoding", "gzip, deflate, br"},
			{"accept-language", "en-US,en;q=0.9"},
		},
		ResponseBodyStruct: &monitorResponse,
	})
	if err != nil {
		return stockData
	}

	switch resp.StatusCode {
	case 200:
		stockData.Price, _ = strconv.ParseFloat(monitorResponse.Gtmdata.Price.Sellingprice, 64)
		for _, event := range monitorResponse.Mccevents[0][1].([]interface{}) {
			stockData.ImageURL = fmt.Sprint(event.(map[string]interface{})["image_url"])
		}
		stockData.ItemName = monitorResponse.Gtmdata.Productinfo.Name
		stockData.ProductURL = BaseEndpoint + strings.Split(monitorResponse.Product.Selectedproducturl, "?")[0]

		inBudget := monitor.SKUWithInfo[sku].MaxPrice == -1 || (stockData.Price != 0 && monitor.SKUWithInfo[sku].MaxPrice >= int(stockData.Price))
		if monitorResponse.Gtmdata.Productinfo.Availability == "Available" || (monitorResponse.Product.Availability.ButtonText == "Pre-Order" && monitorResponse.Product.Available) {
			if inBudget {
				monitor.SKUsSentToTask = append(monitor.SKUsSentToTask, sku)
				stockData.SKU = sku
				stockData.PID = monitorResponse.Gtmdata.Productinfo.SKU
				stockData.MaxQuantity = monitorResponse.Product.MaxOrderQuantity
			} else {
				monitor.SKUsSentToTask = common.RemoveFromSlice(monitor.SKUsSentToTask, sku)
				stockData.OutOfPriceRange = true
			}
			return stockData
		} else {
			monitor.SKUsSentToTask = common.RemoveFromSlice(monitor.SKUsSentToTask, sku)
			return stockData
		}
	default:
		return stockData
	}

}
