package shopify

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

// CreateShopifyMonitor takes a TaskGroup entity and turns it into a Shopify Monitor
func CreateShopifyMonitor(taskGroup *entities.TaskGroup, proxyGroup *entities.ProxyGroup, eventBus *events.EventBus, siteURL, sitePassword string, singleMonitors []entities.ShopifySingleMonitorInfo) (Monitor, error) {
	storedShopifyMonitors := make(map[string]entities.ShopifySingleMonitorInfo)
	shopifyMonitor := Monitor{}
	vIDs := []string{}

	for _, monitor := range singleMonitors {
		storedShopifyMonitors[monitor.VariantID] = monitor
		vIDs = append(vIDs, monitor.VariantID)
	}

	shopifyMonitor = Monitor{
		Monitor: base.Monitor{
			TaskGroup:  taskGroup,
			ProxyGroup: proxyGroup,
			EventBus:   eventBus,
		},
		SiteURL:      siteURL,
		SitePassword: sitePassword,
		VIDs:         vIDs,
		SKUWithInfo:  storedShopifyMonitors,
	}

	return shopifyMonitor, nil
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
			becameGuest = BecomeGuest(monitor.Monitor.Client, monitor.SiteURL, monitor.SitePassword)
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
	wg.Add(len(monitor.VIDs))
	for _, vid := range monitor.VIDs {
		go func(x string) {
			monitor.RunSingleMonitor(x)
			wg.Done()
		}(vid)
	}
	wg.Wait()

}

func (monitor *Monitor) RunSingleMonitor(vid string) {
	needToStop := monitor.CheckForStop()
	if needToStop {
		return
	}

	defer func() {
		if recover() != nil {
			time.Sleep(time.Duration(monitor.Monitor.TaskGroup.MonitorDelay) * time.Millisecond)
			monitor.RunSingleMonitor(vid)
		}
	}()

	var proxy *entities.Proxy
	if monitor.Monitor.ProxyGroup != nil {
		if len(monitor.Monitor.ProxyGroup.Proxies) > 0 {
			proxy = util.RandomLeastUsedProxy(monitor.Monitor.ProxyGroup.Proxies)
			monitor.Monitor.UpdateProxy(proxy)
		}
	}

	stockData := monitor.GetVIDstock(vid)
	if stockData.VariantID != "" {
		needToStop := monitor.CheckForStop()
		if needToStop {
			return
		}

		var inSlice bool
		for _, monitorStock := range monitor.InStock {
			inSlice = monitorStock.VariantID == stockData.VariantID
		}
		if !inSlice {
			monitor.InStock = append(monitor.InStock, stockData)
			monitor.PublishEvent(enums.SendingProductInfoToTasks, enums.MonitorUpdate, events.ProductInfo{
				Products: []events.Product{
					{ProductName: stockData.ItemName, ProductImageURL: stockData.ImageURL}},
			})
		}
	} else {
		if monitor.Monitor.TaskGroup.MonitorStatus != enums.WaitingForInStock {
			monitor.PublishEvent(enums.WaitingForInStock, enums.MonitorUpdate, nil)
		}
		for i, monitorStock := range monitor.InStock {
			if monitorStock.VariantID == stockData.VariantID {
				monitor.InStock = append(monitor.InStock[:i], monitor.InStock[i+1:]...)
				break
			}
		}
	}

	time.Sleep(time.Duration(monitor.Monitor.TaskGroup.MonitorDelay) * time.Millisecond)
	monitor.RunSingleMonitor(vid)
}

// Getting stock by adding to cart
func (monitor *Monitor) GetVIDstock(vid string) ShopifyInStockData {
	stockData := ShopifyInStockData{}
	paramsString := common.CreateParams(map[string]string{
		"form_type": "product",
		"utf8":      "✓",
		"id":        vid,
		"quantity":  "1",
	})

	monitorResponse := AddToCartResponse{}
	resp, _, err := util.MakeRequest(&util.Request{
		Client: monitor.Monitor.Client,
		Method: "POST",
		URL:    monitor.SiteURL + AddToCartEndpoint,
		RawHeaders: http.RawHeader{
			{"content-length", fmt.Sprint(len(paramsString))},
			{"sec-ch-ua", `" Not;A Brand";v="99", "Google Chrome";v="91", "Chromium";v="91"`},
			{"accept", "application/json, text/javascript, */*; q=0.01"},
			{"x-requested-with", "XMLHttpRequest"},
			{"sec-ch-ua-mobile", "?0"},
			{"user-agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/91.0.4472.106 Safari/537.36"},
			{"content-type", "application/x-www-form-urlencoded; charset=UTF-8"},
			{"origin", monitor.SiteURL},
			{"sec-fetch-site", "same-origin"},
			{"sec-fetch-site", "cors"},
			{"sec-fetch-dest", "empty"},
			{"referer", monitor.SiteURL + "/"},
			{"accept-encoding", "gzip, deflate"},
			{"accept-language", "en-US,en;q=0.9"},
		},
		Data:               []byte(paramsString),
		ResponseBodyStruct: &monitorResponse,
	})
	if err != nil {
		return stockData
	}

	switch resp.StatusCode {
	case 200:
		if monitor.SKUWithInfo[vid].MaxPrice >= monitorResponse.Price/100 || monitor.SKUWithInfo[vid].MaxPrice == -1 {
			stockData.ItemName = monitorResponse.Title
			stockData.ItemName = monitorResponse.Title
			stockData.Price = float64(monitorResponse.Price / 100)
			stockData.VariantID = vid
		}
	case 422:
		fmt.Println("Out Of Stock")
	case 404:
		fmt.Println("Item does not exist")
	default:
		fmt.Println("Unknown StatusCode:", resp.StatusCode)
	}
	return stockData
}
