package newegg

import (
	"fmt"
	"sync"
	"time"

	"backend.juicedbot.io/juiced.client/http"
	"backend.juicedbot.io/juiced.infrastructure/common"
	"backend.juicedbot.io/juiced.infrastructure/common/entities"
	"backend.juicedbot.io/juiced.infrastructure/common/enums"
	"backend.juicedbot.io/juiced.infrastructure/common/events"
	"backend.juicedbot.io/juiced.sitescripts/base"
	"backend.juicedbot.io/juiced.sitescripts/util"
)

// CreateNeweggMonitor takes a TaskGroup entity and turns it into a Newegg Monitor
func CreateNeweggMonitor(taskGroup *entities.TaskGroup, proxyGroup *entities.ProxyGroup, eventBus *events.EventBus, singleMonitors []entities.NeweggSingleMonitorInfo) (Monitor, error) {
	storedNeweggMonitors := make(map[string]entities.NeweggSingleMonitorInfo)
	neweggMonitor := Monitor{}
	skus := []string{}

	for _, monitor := range singleMonitors {
		storedNeweggMonitors[monitor.SKU] = monitor
		skus = append(skus, monitor.SKU)
	}

	neweggMonitor = Monitor{
		Monitor: base.Monitor{
			TaskGroup:  taskGroup,
			ProxyGroup: proxyGroup,
			EventBus:   eventBus,
		},

		SKUs:        skus,
		SKUWithInfo: storedNeweggMonitors,
	}

	return neweggMonitor, nil
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

// So theres a few different ways we can make the monitoring groups for Amazon, for now I'm going to make it so it runs a goroutine for each ASIN
func (monitor *Monitor) RunMonitor() {
	// If the function panics due to a runtime error, recover from it
	defer func() {
		if r := recover(); r != nil {
			monitor.PublishEvent(fmt.Sprintf(enums.MonitorFailed, r), enums.MonitorFail, 0)
			monitor.Monitor.StopFlag = true
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
			becameGuest = BecomeGuest(monitor.Monitor.Client)
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

	if !common.InSlice(monitor.RunningMonitors, sku) {
		defer func() {
			if r := recover(); r != nil {
				monitor.PublishEvent(fmt.Sprintf(enums.MonitorFailed, r), enums.MonitorFail, 0)
				monitor.Monitor.StopFlag = true
			}
		}()

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
				monitor.RunningMonitors = common.RemoveFromSlice(monitor.RunningMonitors, sku)
				monitor.PublishEvent(enums.SendingProductInfoToTasks, enums.MonitorUpdate, events.ProductInfo{
					Products: []events.Product{
						{ProductName: stockData.ProductName, ProductImageURL: stockData.ImageURL}},
				})
			}
		} else {
			if len(monitor.RunningMonitors) > 0 {
				if stockData.OutOfPriceRange {
					if monitor.Monitor.TaskGroup.MonitorStatus != enums.OutOfPriceRange {
						monitor.PublishEvent(enums.OutOfPriceRange, enums.MonitorUpdate, events.ProductInfo{
							Products: []events.Product{
								{ProductName: stockData.ProductName, ProductImageURL: stockData.ImageURL}},
						})
					}
				} else {
					if monitor.Monitor.TaskGroup.MonitorStatus != enums.WaitingForInStock {
						monitor.PublishEvent(enums.WaitingForInStock, enums.MonitorUpdate, events.ProductInfo{
							Products: []events.Product{
								{ProductName: stockData.ProductName, ProductImageURL: stockData.ImageURL}},
						})
					}
				}
			}
			for i, monitorStock := range monitor.InStock {
				if monitorStock.SKU == stockData.SKU {
					monitor.InStock = append(monitor.InStock[:i], monitor.InStock[i+1:]...)
					break
				}
			}
			time.Sleep(time.Duration(monitor.Monitor.TaskGroup.MonitorDelay) * time.Millisecond)
			monitor.RunSingleMonitor(sku)
		}
	}
}

func (monitor *Monitor) GetSKUStock(sku string) NeweggInStockData {
	stockData := NeweggInStockData{}
	var monitorResponse MonitorResponse
	resp, _, err := util.MakeRequest(&util.Request{
		Client: monitor.Monitor.Client,
		Method: "GET",
		URL:    fmt.Sprintf(MonitorEndpoint, sku),
		RawHeaders: http.RawHeader{
			{"pragma", `no-cache`},
			{"cache-control", `no-cache`},
			{"sec-ch-ua", `"Chromium";v="92", " Not A;Brand";v="99", "Google Chrome";v="92"`},
			{"sec-ch-ua-mobile", `?0`},
			{"upgrade-insecure-requests", `1`},
			{"user-agent", `Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/92.0.4515.131 Safari/537.36`},
			{"accept", `text/html,application/xhtml+xml,application/xml;q=0.9,image/avif,image/webp,image/apng,*/*;q=0.8,application/signed-exchange;v=b3;q=0.9`},
			{"sec-fetch-site", `none`},
			{"sec-fetch-mode", `navigate`},
			{"sec-fetch-user", `?1`},
			{"sec-fetch-dest", `document`},
			{"accept-encoding", `gzip, deflate, br`},
			{"accept-language", `en-US,en;q=0.9`},
		},

		ResponseBodyStruct: &monitorResponse,
	})
	if err != nil || resp.StatusCode != 200 {
		fmt.Println(err)
		return stockData
	}

	monitor.RunningMonitors = append(monitor.RunningMonitors, sku)
	price := monitorResponse.MainItem.FinalPrice
	inBudget := monitor.SKUWithInfo[sku].MaxPrice == -1 || (price != 0 && float64(monitor.SKUWithInfo[sku].MaxPrice) > price)
	stockData.ProductName = monitorResponse.MainItem.Description.ProductName
	stockData.ItemURL = "https://www.newegg.com/p/" + sku
	stockData.ImageURL = "https://c1.neweggimages.com/NeweggImage/ProductImage/" + monitorResponse.MainItem.Image.Normal.ImageName
	stockData.Price = price
	if monitorResponse.MainItem.InStock {
		if inBudget {
			stockData.SKU = sku
			stockData.ItemNumber = monitorResponse.MainItem.Item
			monitor.SKUsSentToTask = append(monitor.SKUsSentToTask, sku)
		} else {
			stockData.OutOfPriceRange = true
			monitor.SKUsSentToTask = common.RemoveFromSlice(monitor.SKUsSentToTask, sku)
		}
	} else {
		monitor.SKUsSentToTask = common.RemoveFromSlice(monitor.SKUsSentToTask, sku)
	}

	return stockData
}
