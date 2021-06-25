package gamestop

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"backend.juicedbot.io/juiced.infrastructure/common"
	"backend.juicedbot.io/juiced.infrastructure/common/entities"
	"backend.juicedbot.io/juiced.infrastructure/common/enums"
	"backend.juicedbot.io/juiced.infrastructure/common/events"
	"backend.juicedbot.io/juiced.sitescripts/base"
	"backend.juicedbot.io/juiced.sitescripts/util"
)

// CreateGamestopMonitor takes a TaskGroup entity and turns it into a Gamestop Monitor
func CreateGamestopMonitor(taskGroup *entities.TaskGroup, proxy entities.Proxy, eventBus *events.EventBus, singleMonitors []entities.GamestopSingleMonitorInfo) (Monitor, error) {
	storedGamestopMonitors := make(map[string]entities.GamestopSingleMonitorInfo)
	gamestopMonitor := Monitor{}
	skus := []string{}

	for _, monitor := range singleMonitors {
		storedGamestopMonitors[monitor.SKU] = monitor
		skus = append(skus, monitor.SKU)
	}

	client, err := util.CreateClient(proxy)
	if err != nil {
		return gamestopMonitor, err
	}
	gamestopMonitor = Monitor{
		Monitor: base.Monitor{
			TaskGroup: taskGroup,
			Proxy:     proxy,
			EventBus:  eventBus,
			Client:    client,
		},

		SKUs:        skus,
		SKUWithInfo: storedGamestopMonitors,
	}

	becameGuest := false
	for !becameGuest {
		needToStop := gamestopMonitor.CheckForStop()
		if needToStop {
			return gamestopMonitor, nil
		}
		becameGuest = BecomeGuest(&gamestopMonitor.Monitor.Client)
		if !becameGuest {
			time.Sleep(1000 * time.Millisecond)
		}
	}

	return gamestopMonitor, nil
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

// So theres a few different ways we can make the monitoring groups for Amazon, for now I'm going to make it so it runs a goroutine for each ASIN
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
	for _, sku := range monitor.SKUs {
		if !common.InSlice(monitor.RunningMonitors, sku) {
			// TODO @Humphrey: THIS IS GOING TO CAUSE A MASSIVE MEMORY LEAK -- IF YOU HAVE 2 MONITORS, AND EACH ONE CALLS THE RUNMONITOR FUNCTION FROM WITHIN, YOU'LL START MULTIPLYING AND VERY QUICKLY YOU'LL HAVE THOUSANDS OF MONITORS
			// 		--> We should turn this into a RunSingleMonitor function, and have it call itself from within
			go func(t string) {
				// If the function panics due to a runtime error, recover from it
				defer func() {
					recover()
					// TODO @silent: Re-run this specific monitor
				}()

				stockData := monitor.GetSKUStock(t)
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
						monitor.RunningMonitors = common.RemoveFromSlice(monitor.RunningMonitors, t)
						monitor.PublishEvent(enums.SendingProductInfoToTasks, enums.MonitorUpdate)
					}
				} else {
					if len(monitor.RunningMonitors) > 0 {
						if monitor.Monitor.TaskGroup.MonitorStatus != enums.WaitingForInStock {
							monitor.PublishEvent(enums.WaitingForInStock, enums.MonitorUpdate)
						}
					}
					for i, monitorStock := range monitor.InStock {
						if monitorStock.SKU == stockData.SKU {
							monitor.InStock = append(monitor.InStock[:i], monitor.InStock[i+1:]...)
							break
						}
					}
					time.Sleep(time.Duration(monitor.Monitor.TaskGroup.MonitorDelay) * time.Millisecond)
					monitor.RunMonitor()
				}
			}(sku)
		}

	}

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
		fmt.Println(err.Error())
	}

	switch resp.StatusCode {
	case 200:
		monitor.RunningMonitors = append(monitor.RunningMonitors, sku)

		switch monitorResponse.Gtmdata.Productinfo.Availability {
		case "Available":
			stockData.Price, _ = strconv.Atoi(monitorResponse.Gtmdata.Price.Sellingprice)

			var inBudget bool
			inBudget = monitor.SKUWithInfo[sku].MaxPrice > stockData.Price || monitor.SKUWithInfo[sku].MaxPrice == -1
			if inBudget {
				for _, event := range monitorResponse.Mccevents[0][1].([]interface{}) {
					stockData.ImageURL = fmt.Sprint(event.(map[string]interface{})["image_url"])
				}
				stockData.SKU = sku
				stockData.ItemName = monitorResponse.Gtmdata.Productinfo.Name
				stockData.PID = monitorResponse.Gtmdata.Productinfo.SKU

				stockData.ProductURL = BaseEndpoint + strings.Split(monitorResponse.Product.Selectedproducturl, "?")[0]

				monitor.SKUsSentToTask = append(monitor.SKUsSentToTask, sku)

			}
			return stockData
		case "Not Available":
			monitor.SKUsSentToTask = common.RemoveFromSlice(monitor.SKUsSentToTask, sku)
			return stockData
		default:
			fmt.Println(monitorResponse.Gtmdata.Productinfo.Availability)
			return stockData
		}
	default:
		return stockData
	}

}
