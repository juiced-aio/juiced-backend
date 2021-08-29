package pokemoncenter

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"strings"
	"sync"
	"time"

	"backend.juicedbot.io/juiced.infrastructure/common"
	"backend.juicedbot.io/juiced.infrastructure/common/entities"
	"backend.juicedbot.io/juiced.infrastructure/common/enums"
	"backend.juicedbot.io/juiced.infrastructure/common/events"
	"backend.juicedbot.io/juiced.sitescripts/base"
	"backend.juicedbot.io/juiced.sitescripts/util"
	"github.com/anaskhan96/soup"
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
		monitor.PublishEvent(enums.WaitingForInStock, enums.MonitorUpdate, events.ProductInfo{
			Products: []events.Product{
				{ProductName: stockData.ItemName, ProductImageURL: stockData.ImageURL}},
		})
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
		URL:    fmt.Sprintf(MonitorEndpoint, sku),
		RawHeaders: [][2]string{
			{"sec-ch-ua", "\" Not A;Brand\";v=\"99\", \"Chromium\";v=\"90\", \"Google Chrome\";v=\"90\""},
			{"accept", "*/*"},
			{"x-requested-with", "XMLHttpRequest"},
			{"sec-ch-ua-mobile", "?0"},
			{"user-agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/15.0 Safari/605.1.15"},
			{"sec-fetch-site", "same-origin"},
			{"sec-fetch-mode", "cors"},
			{"sec-fetch-dest", "empty"},
			{"referer", BaseEndpoint},
			{"accept-encoding", "gzip, deflate, br"},
			{"accept-language", "en-US,en;q=0.9"},
			{"cache-control", "max-age=0"},
		},
	})
	if err != nil {
		fmt.Println(err.Error())
	}

	switch resp.StatusCode {
	case 403:
		monitor.HandleDatadome(body)
		return stockData
	case 200:
		//Get response data as this is embedded into the HTML in a script.
		body, _ := ioutil.ReadAll(resp.Body)
		responseBody := soup.HTMLParse(string(body))
		nextData := responseBody.Find("script", "id", "__NEXT_DATA__")
		nextDataString := nextData.Pointer.FirstChild.Data
		json.Unmarshal([]byte(nextDataString), &monitorResponse)

		switch monitorResponse.Props.InitialState.Product.Availability {
		case "AVAILABLE":
			stockData.Price = monitorResponse.Props.InitialState.Product.ListPrice.Amount
			fmt.Println(monitorResponse.Props.InitialState.Product.ListPrice.Amount)
			var inBudget bool
			inBudget = monitor.SKUWithInfo[sku].MaxPrice > int(stockData.Price) || monitor.SKUWithInfo[sku].MaxPrice == -1
			if inBudget {
				stockData.ImageURL = monitorResponse.Props.InitialState.Product.Images.Original
				stockData.SKU = sku
				stockData.ItemName = monitorResponse.Props.InitialState.Product.Name
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

func (monitor Monitor) HandleDatadome(body string) {
	status := monitor.Monitor.TaskGroup.MonitorStatus
	monitor.PublishEvent(enums.WaitingForCaptchaMonitor, enums.MonitorUpdate, nil)
	quit := make(chan bool)
	defer func() {
		quit <- true
	}()

	cancellationToken := util.CancellationToken{Cancel: false}
	go func() {
		for {
			select {
			case <-quit:
				return
			default:
				needToStop := monitor.CheckForStop()
				if needToStop {
					cancellationToken.Cancel = true
					return
				}
			}
			time.Sleep(common.MS_TO_WAIT)
		}
	}()

	datadomeStr, err := util.FindInString(body, "<script>var dd=", "}")
	if err != nil {
		return
	}
	datadomeStr += "}"
	datadomeStr = strings.ReplaceAll(datadomeStr, "'", "\"")

	datadomeInfo := DatadomeInfo{}
	err = json.Unmarshal([]byte(datadomeStr), &datadomeInfo)
	if err != nil {
		log.Println(err.Error())
		return
	}

	cookies := monitor.Monitor.Client.Jar.Cookies(BaseURL)
	for _, cookie := range cookies {
		if cookie.Name == "datadome" {
			datadomeInfo.CID = cookie.Value
		}
	}

	if datadomeInfo.CID == "" {
		return
	}

	err = SetDatadomeCookie(datadomeInfo, monitor.Monitor.Proxy, &monitor.Monitor.Client, &cancellationToken)
	if err != nil {
		log.Println(err.Error())
		return
	} else {
		monitor.PublishEvent(status, enums.MonitorUpdate, nil)
		log.Println("Cookie updated.")
		return
	}
}
