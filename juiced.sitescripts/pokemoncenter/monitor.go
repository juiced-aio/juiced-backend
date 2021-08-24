package pokemoncenter

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"sync"
	"time"

	"backend.juicedbot.io/juiced.infrastructure/common"
	"backend.juicedbot.io/juiced.infrastructure/common/entities"
	"backend.juicedbot.io/juiced.infrastructure/common/enums"
	"backend.juicedbot.io/juiced.infrastructure/common/events"
	"backend.juicedbot.io/juiced.sitescripts/util"
	"github.com/anaskhan96/soup"
)

// CreatePokemonCenterMonitor takes a TaskGroup entity and turns it into a pokemoncenter Monitor
func CreatePokemonCenterMonitor(singleMonitors []entities.PokemonCenterSingleMonitorInfo) (Monitor, error) {
	storedPokemonCenterMonitors := make(map[string]entities.PokemonCenterSingleMonitorInfo)
	pokemonCenterMonitor := Monitor{}
	skus := []string{}

	for _, monitor := range singleMonitors {
		storedPokemonCenterMonitors[monitor.SKU] = monitor
		skus = append(skus, monitor.SKU)
	}

	pokemonCenterMonitor = Monitor{
		SKUs:        skus,
		SKUWithInfo: storedPokemonCenterMonitors,
	}

	return pokemonCenterMonitor, nil
}

// PublishEvent wraps the EventBus's PublishMonitorEvent function
func (monitor *Monitor) PublishEvent(status enums.MonitorStatus, eventType enums.MonitorEventType, data interface{}) {
	monitor.MonitorInfo.TaskGroup.SetMonitorStatus(status)
	monitor.MonitorInfo.EventBus.PublishMonitorEvent(status, eventType, data, monitor.MonitorInfo.TaskGroup.GroupID)
}

//This checks if we want to stop
func (monitor *Monitor) CheckForStop() bool {
	if monitor.MonitorInfo.StopFlag {
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
			monitor.MonitorInfo.StopFlag = true
			monitor.PublishEvent(enums.MonitorIdle, enums.MonitorFail, nil)
		}
	}()

	if monitor.MonitorInfo.TaskGroup.MonitorStatus == enums.MonitorIdle {
		monitor.PublishEvent(enums.WaitingForProductData, enums.MonitorStart, nil)
	}

	if monitor.MonitorInfo.Client.Transport == nil {
		monitorClient, err := util.CreateClient()
		if err != nil {
			return
		}
		monitor.MonitorInfo.Client = monitorClient

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
			recover()
			// TODO @silent: Re-run this specific monitor
		}()

		if monitor.MonitorInfo.ProxyGroup != nil {
			if len(monitor.MonitorInfo.ProxyGroup.Proxies) > 0 {
				proxy := util.RandomLeastUsedProxy(monitor.MonitorInfo.ProxyGroup.Proxies)
				monitor.MonitorInfo.UpdateProxy(proxy)
			}
		}

		stockData := monitor.GetSKUStock(sku)
		if stockData.StockInfo.SKU != "" {
			needToStop := monitor.CheckForStop()
			if needToStop {
				return
			}

			var inSlice bool
			for _, monitorStock := range monitor.InStock {
				inSlice = monitorStock.StockInfo.SKU == stockData.StockInfo.SKU
			}
			if !inSlice {
				monitor.InStock = append(monitor.InStock, stockData)
				monitor.RunningMonitors = common.RemoveFromSlice(monitor.RunningMonitors, sku)
				monitor.PublishEvent(enums.SendingProductInfoToTasks, enums.MonitorUpdate, events.ProductInfo{
					Products: []events.Product{
						{ProductName: stockData.StockInfo.ItemName, ProductImageURL: stockData.StockInfo.ImageURL}},
				})
			}
		} else {
			if len(monitor.RunningMonitors) > 0 {
				monitor.PublishEvent(enums.WaitingForInStock, enums.MonitorUpdate, events.ProductInfo{
					Products: []events.Product{
						{ProductName: stockData.StockInfo.ItemName, ProductImageURL: stockData.StockInfo.ImageURL}},
				})
			}
			for i, monitorStock := range monitor.InStock {
				if monitorStock.StockInfo.SKU == stockData.StockInfo.SKU {
					monitor.InStock = append(monitor.InStock[:i], monitor.InStock[i+1:]...)
					break
				}
			}
			time.Sleep(time.Duration(monitor.MonitorInfo.TaskGroup.MonitorDelay) * time.Millisecond)
			monitor.RunSingleMonitor(sku)
		}
	}
}

func (monitor *Monitor) GetSKUStock(sku string) PokemonCenterInStockData {
	stockData := PokemonCenterInStockData{}
	monitorResponse := MonitorResponse{}
	resp, _, err := util.MakeRequest(&util.Request{
		Client: monitor.MonitorInfo.Client,
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
	})
	if err != nil {
		fmt.Println(err.Error())
	}

	switch resp.StatusCode {
	case 200:
		monitor.RunningMonitors = append(monitor.RunningMonitors, sku)

		//Get response data as this is embedded into the HTML in a script.
		body, _ := ioutil.ReadAll(resp.Body)
		responseBody := soup.HTMLParse(string(body))
		nextData := responseBody.Find("script", "id", "__NEXT_DATA__")
		nextDataString := nextData.Pointer.FirstChild.Data
		json.Unmarshal([]byte(nextDataString), &monitorResponse)

		switch monitorResponse.Props.InitialState.Product.Availability {
		case "AVAILABLE":
			stockData.StockInfo.Price = monitorResponse.Props.InitialState.Product.ListPrice.Amount
			fmt.Println(monitorResponse.Props.InitialState.Product.ListPrice.Amount)
			var inBudget bool
			inBudget = monitor.SKUWithInfo[sku].MaxPrice > int(stockData.StockInfo.Price) || monitor.SKUWithInfo[sku].MaxPrice == -1
			if inBudget {
				stockData.StockInfo.ImageURL = monitorResponse.Props.InitialState.Product.Images.Original
				stockData.StockInfo.SKU = sku
				stockData.StockInfo.ItemName = monitorResponse.Props.InitialState.Product.Name
				stockData.AddToCartForm = monitorResponse.Props.InitialState.Product.AddToCartForm
				monitor.SKUsSentToTask = append(monitor.SKUsSentToTask, sku)
			}
			return stockData
		case "NOT_AVAILABLE":
			monitor.SKUsSentToTask = common.RemoveFromSlice(monitor.SKUsSentToTask, sku)
			return stockData
		default:
			fmt.Println(monitorResponse.Props.InitialState.Product.Availability)
			return stockData
		}
	default:
		return stockData
	}
}
