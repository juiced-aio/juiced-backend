package pokemoncenter

import (
	"encoding/json"
	"fmt"
	"log"
	"math"
	"strconv"
	"sync"
	"time"

	"backend.juicedbot.io/juiced.infrastructure/common"
	"backend.juicedbot.io/juiced.infrastructure/common/entities"
	"backend.juicedbot.io/juiced.infrastructure/common/enums"
	"backend.juicedbot.io/juiced.infrastructure/common/events"
	"backend.juicedbot.io/juiced.sitescripts/base"
	"backend.juicedbot.io/juiced.sitescripts/util"
)

// CreatePokemonCenterMonitor takes a TaskGroup entity and turns it into a pokemoncenter Monitor
func CreatePokemonCenterMonitor(taskGroup *entities.TaskGroup, proxyGroup *entities.ProxyGroup, eventBus *events.EventBus, singleMonitors []entities.PokemonCenterSingleMonitorInfo) (Monitor, error) {
	storedPokemonCenterMonitors := make(map[string]entities.PokemonCenterSingleMonitorInfo)
	skus := []string{}

	for _, monitor := range singleMonitors {
		storedPokemonCenterMonitors[monitor.SKU] = monitor
		skus = append(skus, monitor.SKU)
	}

	return Monitor{
		Monitor: base.Monitor{
			TaskGroup:  taskGroup,
			ProxyGroup: proxyGroup,
			EventBus:   eventBus,
		},

		SKUs:        skus,
		SKUWithInfo: storedPokemonCenterMonitors,
	}, nil
}

// PublishEvent wraps the EventBus's PublishMonitorEvent function
func (monitor *Monitor) PublishEvent(status enums.MonitorStatus, eventType enums.MonitorEventType, data interface{}) {
	monitor.Monitor.TaskGroup.SetMonitorStatus(status)
	monitor.Monitor.EventBus.PublishMonitorEvent(status, eventType, data, monitor.Monitor.TaskGroup.GroupID)
}

//This checks if we want to stop
func (monitor *Monitor) CheckForStop() bool {
	if monitor.Monitor.StopFlag {
		monitor.PublishEvent(enums.MonitorIdle, enums.MonitorStop, nil)
		return true
	}
	return false
}

//This is responsible for starting the pokemoncenter Product monitor
func (monitor *Monitor) RunMonitor() {
	// If the function panics due to a runtime error, recover from it
	defer func() {
		if recover() != nil {
			monitor.Monitor.StopFlag = true
			monitor.PublishEvent(enums.MonitorIdle, enums.MonitorFail, nil)
		}
	}()

	log.Println("Monitor starting")

	if monitor.Monitor.TaskGroup.MonitorStatus == enums.MonitorIdle {
		monitor.PublishEvent(enums.WaitingForProductData, enums.MonitorStart, nil)
	}

	if monitor.Monitor.Client.Transport == nil {
		monitorClient, err := util.CreateClient()
		if err != nil {
			return
		}
		monitor.Monitor.Client = monitorClient

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

	if monitor.Monitor.ProxyGroup != nil {
		if len(monitor.Monitor.ProxyGroup.Proxies) > 0 {
			proxy := util.RandomLeastUsedProxy(monitor.Monitor.ProxyGroup.Proxies)
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
			if monitor.Monitor.TaskGroup.MonitorStatus != enums.WaitingForInStock && stockData.ItemName != "" {
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

func (monitor *Monitor) GetSKUStock(sku string) PokemonCenterInStockData {
	stockData := PokemonCenterInStockData{}
	monitorResponse := MonitorResponse{}
	resp, body, err := util.MakeRequest(&util.Request{
		Client: monitor.Monitor.Client,
		Method: "GET",
		// Random string here is to bypass caching
		URL: fmt.Sprintf(MonitorEndpoint, sku) + common.RandString(5),
		RawHeaders: [][2]string{
			{"pragma", `no-transform`},
			{"cache-control", `no-cache=#headers`},
			{"sec-ch-ua", `"Chromium";v="92", " Not A;Brand";v="99", "Google Chrome";v="92"`},
			{"content-type", `application/json`},
			{"sec-ch-ua-mobile", `?0`},
			{"user-agent", `Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/15.0 Safari/605.1.15`},
			{"x-store-scope", `pokemon`},
			{"accept", `*/*`},
			{"sec-fetch-site", `same-origin`},
			{"sec-fetch-mode", `cors`},
			{"sec-fetch-dest", `empty`},
			{"referer", BaseEndpoint + "/"},
			{"accept-language", `en-SP`},
			{"origin", BaseEndpoint},
			{"Accept-Encoding", `identity`},
		},
	})
	if err != nil {
		fmt.Println(err.Error())
		return stockData
	}

	switch resp.StatusCode {
	case 403:
		monitor.HandleDatadome(body)
		return stockData
	case 200:
		err = json.Unmarshal([]byte(body), &monitorResponse)
		if err != nil {
			fmt.Println(err.Error())
			return stockData
		}

		if len(monitorResponse.Items) == 0 || len(monitorResponse.Definition) == 0 || len(monitorResponse.Images) == 0 {
			return stockData
		}

		//Get response data as this is embedded into the HTML in a script.
		for _, detail := range monitorResponse.Definition[0].Details {
			switch detail.DisplayName {
			case "Item Name (Web)":
				stockData.ItemName = detail.Value.(string)
			case "Item Name":
				stockData.ItemName = detail.Value.(string)
			case "Purchase Quantity Limit":
				stockData.MaxQuantity, err = strconv.Atoi(detail.DisplayValue)
				if err != nil {
					fmt.Println(err.Error())
					return stockData
				}
			}
		}
		stockData.ImageURL = monitorResponse.Images[0].Original

		for _, element := range monitorResponse.Items[0].Element {
			stockData.AddToCartForm = element.Addtocartform[0].Self.URI
			stockData.Price = math.Floor(element.Price[0].PurchasePrice[0].Amount*100) / 100
			inBudget := monitor.SKUWithInfo[sku].MaxPrice > int(stockData.Price) || monitor.SKUWithInfo[sku].MaxPrice == -1
			if !inBudget {
				stockData.OutOfPriceRange = true
			} else {
				if element.Availability[0].State == "AVAILABLE" {
					stockData.SKU = sku
				}
			}
		}
	}

	return stockData

}
