package disney

import (
	"math/rand"
	"net/url"
	"strings"
	"sync"
	"time"

	"backend.juicedbot.io/juiced.client/client"
	"backend.juicedbot.io/juiced.client/http"
	"backend.juicedbot.io/juiced.infrastructure/common"
	"backend.juicedbot.io/juiced.infrastructure/common/entities"
	"backend.juicedbot.io/juiced.infrastructure/common/enums"
	"backend.juicedbot.io/juiced.infrastructure/common/events"
	"backend.juicedbot.io/juiced.sitescripts/base"
	"backend.juicedbot.io/juiced.sitescripts/util"
)

// CreateDisneyMonitor takes a TaskGroup entity and turns it into a Disney Monitor
func CreateDisneyMonitor(taskGroup *entities.TaskGroup, proxies []entities.Proxy, eventBus *events.EventBus, singleMonitors []entities.DisneySingleMonitorInfo) (Monitor, error) {
	storedDisneyMonitors := make(map[string]entities.DisneySingleMonitorInfo)
	disneyMonitor := Monitor{}

	pids := []string{}
	for _, monitor := range singleMonitors {
		storedDisneyMonitors[monitor.PID] = monitor
		pids = append(pids, monitor.PID)
	}

	disneyMonitor = Monitor{
		Monitor: base.Monitor{
			TaskGroup: taskGroup,
			Proxies:   proxies,
			EventBus:  eventBus,
		},
		Pids:        pids,
		PidWithInfo: storedDisneyMonitors,
	}

	return disneyMonitor, nil
}

// PublishEvent wraps the EventBus's PublishMonitorEvent function
func (monitor *Monitor) PublishEvent(status enums.MonitorStatus, eventType enums.MonitorEventType, data interface{}) {
	monitor.Monitor.TaskGroup.SetMonitorStatus(status)
	monitor.Monitor.EventBus.PublishMonitorEvent(status, eventType, data, monitor.Monitor.TaskGroup.GroupID)
}

func (monitor *Monitor) CheckForStop() bool {
	if monitor.Monitor.StopFlag {
		monitor.PublishEvent(enums.MonitorIdle, enums.MonitorStop, nil)
		return true
	}
	return false
}

func (monitor *Monitor) RunMonitor() {
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

		if len(monitor.Monitor.Proxies) > 0 {
			client.UpdateProxy(&monitor.Monitor.Client, common.ProxyCleaner(monitor.Monitor.Proxies[rand.Intn(len(monitor.Monitor.Proxies))]))
		}
	}

	wg := sync.WaitGroup{}
	wg.Add(len(monitor.Pids))
	for _, pid := range monitor.Pids {
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

	stockData := DisneyInStockData{}

	if len(monitor.Monitor.Proxies) > 0 {
		client.UpdateProxy(&monitor.Monitor.Client, common.ProxyCleaner(monitor.Monitor.Proxies[rand.Intn(len(monitor.Monitor.Proxies))]))
	}

	stockData = monitor.StockMonitor(pid)

	needToStop = monitor.CheckForStop()
	if needToStop {
		return
	}

	if stockData.PID != "" {
		var inSlice bool
		for _, monitorStock := range monitor.InStock {
			inSlice = monitorStock.PID == stockData.PID
		}
		if !inSlice {
			monitor.RunningMonitors = common.RemoveFromSlice(monitor.RunningMonitors, pid)
			monitor.InStock = append(monitor.InStock, stockData)
			monitor.PublishEvent(enums.SendingProductInfoToTasks, enums.MonitorUpdate, events.ProductInfo{
				Products: []events.Product{
					{ProductName: stockData.ProductName, ProductImageURL: stockData.ImageURL}},
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

func (monitor *Monitor) StockMonitor(pid string) DisneyInStockData {
	stockData := DisneyInStockData{}

	params := url.Values{}
	params.Add("pid", pid)

	if monitor.PidWithInfo[pid].Size != "" {
		params.Add("dwvar_"+pid+"_size", monitor.PidWithInfo[pid].Size)
	}
	if monitor.PidWithInfo[pid].Color != "" {
		params.Add("dwvar_"+pid+"_color", monitor.PidWithInfo[pid].Color)
	}

	monitorResponse := MonitorResponse{}

	resp, _, err := util.MakeRequest(&util.Request{
		Client: monitor.Monitor.Client,
		Method: "GET",
		URL:    MonitorEndpoint + "?" + params.Encode(),
		RawHeaders: http.RawHeader{
			{"sec-ch-ua", `" Not;A Brand";v="99", "Google Chrome";v="91", "Chromium";v="91"`},
			{"accept", "application/json, text/javascript, */*; q=0.01"},
			{"x-requested-with", "XMLHttpRequest"},
			{"sec-ch-ua-mobile", "?0"},
			{"user-agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/91.0.4472.106 Safari/537.36"},
			{"content-type", "application/json; charset=UTF-8"},
			{"sec-fetch-site", "same-origin"},
			{"sec-fetch-site", "cors"},
			{"sec-fetch-dest", "empty"},
			{"accept-encoding", "gzip, deflate"},
			{"accept-language", "en-US,en;q=0.9"},
		},
		ResponseBodyStruct: &monitorResponse,
	})
	if err != nil || resp.StatusCode != 200 || !monitorResponse.Product.Available || (monitor.PidWithInfo[pid].MaxPrice != -1 && monitor.PidWithInfo[pid].MaxPrice < int(monitorResponse.Product.Price.Sales.Value)) {
		return stockData
	}

	stockData.ProductName = monitorResponse.Product.ProductName
	stockData.VID = monitorResponse.Product.ID
	stockData.ItemURL = BaseEndpoint + "/" + strings.Split(monitorResponse.Product.SelectedProductUrl, "?")[0] + "?isProductSearch=0&plpPosition=1"
	if len(monitorResponse.Product.Images.Small) > 0 {
		stockData.ImageURL = monitorResponse.Product.Images.Small[0].URL
	} else {
		stockData.ImageURL = monitorResponse.Product.Images.HighRes[0].URL
	}

	return stockData
}
