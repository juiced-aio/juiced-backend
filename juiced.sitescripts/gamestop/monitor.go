package gamestop

import (
	"fmt"
	"math/rand"
	"net/url"
	"strconv"
	"strings"
	"sync"
	"time"

	"backend.juicedbot.io/juiced.client/client"
	"backend.juicedbot.io/juiced.infrastructure/common"
	"backend.juicedbot.io/juiced.infrastructure/common/entities"
	"backend.juicedbot.io/juiced.infrastructure/common/enums"
	"backend.juicedbot.io/juiced.infrastructure/common/events"
	"backend.juicedbot.io/juiced.sitescripts/base"
	"backend.juicedbot.io/juiced.sitescripts/util"
)

// CreateGamestopMonitor takes a TaskGroup entity and turns it into a Gamestop Monitor
func CreateGamestopMonitor(taskGroup *entities.TaskGroup, proxies []entities.Proxy, eventBus *events.EventBus, singleMonitors []entities.GamestopSingleMonitorInfo) (Monitor, error) {
	storedGamestopMonitors := make(map[string]entities.GamestopSingleMonitorInfo)
	gamestopMonitor := Monitor{}
	pids := []string{}

	for _, monitor := range singleMonitors {
		storedGamestopMonitors[monitor.PID] = monitor
		pids = append(pids, monitor.PID)
	}

	gamestopMonitor = Monitor{
		Monitor: base.Monitor{
			TaskGroup: taskGroup,
			Proxies:   proxies,
			EventBus:  eventBus,
		},

		PIDs:        pids,
		PIDWithInfo: storedGamestopMonitors,
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

// So theres a few different ways we can make the monitoring groups for Amazon, for now I'm going to make it so it runs a goroutine for each ASIN
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

		if len(monitor.Monitor.Proxies) > 0 {
			client.UpdateProxy(&monitor.Monitor.Client, common.ProxyCleaner(monitor.Monitor.Proxies[rand.Intn(len(monitor.Monitor.Proxies))]))
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
	wg.Add(len(monitor.PIDs))
	for _, pid := range monitor.PIDs {
		go func(x string) {
			monitor.RunSingleMonitor(x)
			wg.Done()
		}(pid)
	}
	wg.Wait()

}

func (monitor *Monitor) RunSingleMonitor(pid string) {
	needToStop := monitor.CheckForStop()
	if needToStop {
		return
	}

	if !common.InSlice(monitor.RunningMonitors, pid) {
		defer func() {
			recover()
			// TODO @silent: Re-run this specific monitor
		}()

		if len(monitor.Monitor.Proxies) > 0 {
			client.UpdateProxy(&monitor.Monitor.Client, common.ProxyCleaner(monitor.Monitor.Proxies[rand.Intn(len(monitor.Monitor.Proxies))]))
		}

		stockData := monitor.GetPIDStock(pid)
		if stockData.PID != "" {
			needToStop := monitor.CheckForStop()
			if needToStop {
				return
			}

			var inSlice bool
			for _, monitorStock := range monitor.InStock {
				inSlice = monitorStock.PID == stockData.PID
			}
			if !inSlice {
				monitor.InStock = append(monitor.InStock, stockData)
				monitor.RunningMonitors = common.RemoveFromSlice(monitor.RunningMonitors, pid)
				monitor.PublishEvent(enums.SendingProductInfoToTasks, enums.MonitorUpdate, events.ProductInfo{
					Products: []events.Product{
						{ProductName: stockData.ItemName, ProductImageURL: stockData.ImageURL}},
				})
			}
		} else {
			if len(monitor.RunningMonitors) > 0 {
				if monitor.Monitor.TaskGroup.MonitorStatus != enums.WaitingForInStock {
					monitor.PublishEvent(enums.WaitingForInStock, enums.MonitorUpdate, nil)
				}
			}
			for i, monitorStock := range monitor.InStock {
				if monitorStock.PID == stockData.PID {
					monitor.InStock = append(monitor.InStock[:i], monitor.InStock[i+1:]...)
					break
				}
			}
			time.Sleep(time.Duration(monitor.Monitor.TaskGroup.MonitorDelay) * time.Millisecond)
			monitor.RunSingleMonitor(pid)
		}
	}
}

// Checks if the item is instock and fills the monitors EventInfo if so
func (monitor *Monitor) GetPIDStock(pid string) GamestopInStockData {
	params := url.Values{}
	if monitor.PIDWithInfo[pid].Color != "" {
		params.Add(fmt.Sprintf("dwvar_%v_color", pid), monitor.PIDWithInfo[pid].Color)
	}
	if monitor.PIDWithInfo[pid].Size != "" {
		params.Add(fmt.Sprintf("dwvar_%v_size", pid), monitor.PIDWithInfo[pid].Size)
	}
	if monitor.PIDWithInfo[pid].Condition != "" {
		params.Add(fmt.Sprintf("dwvar_%v_condition", pid), monitor.PIDWithInfo[pid].Condition)
	}
	currentEndpoint := fmt.Sprintf(MonitorEndpoint, pid)
	if params.Encode() != "" {
		currentEndpoint += "&" + params.Encode()
	}

	stockData := GamestopInStockData{}
	monitorResponse := MonitorResponse{}
	resp, _, err := util.MakeRequest(&util.Request{
		Client: monitor.Monitor.Client,
		Method: "GET",
		URL:    currentEndpoint,
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
		monitor.RunningMonitors = append(monitor.RunningMonitors, pid)
		if monitorResponse.Gtmdata.Productinfo.Availability == "Available" || monitorResponse.Product.Availability.ButtonText == "Pre-Order" {
			stockData.Price, _ = strconv.ParseFloat(monitorResponse.Gtmdata.Price.Sellingprice, 64)
			inBudget := monitor.PIDWithInfo[pid].MaxPrice >= int(stockData.Price) || monitor.PIDWithInfo[pid].MaxPrice == -1
			if inBudget {
				for _, event := range monitorResponse.Mccevents[0][1].([]interface{}) {
					stockData.ImageURL = fmt.Sprint(event.(map[string]interface{})["image_url"])
				}
				stockData.PID = pid
				stockData.ItemName = monitorResponse.Gtmdata.Productinfo.Name
				stockData.VID = monitorResponse.Gtmdata.Productinfo.SKU

				stockData.ProductURL = BaseEndpoint + strings.Split(monitorResponse.Product.Selectedproducturl, "?")[0]

				monitor.PIDsSentToTask = append(monitor.PIDsSentToTask, pid)

			}
			return stockData
		} else {
			monitor.PIDsSentToTask = common.RemoveFromSlice(monitor.PIDsSentToTask, pid)
			return stockData
		}
	default:
		return stockData
	}

}
