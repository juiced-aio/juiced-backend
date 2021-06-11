package bestbuy

import (
	"fmt"
	"net/url"
	"strings"
	"time"

	"backend.juicedbot.io/juiced.infrastructure/common/entities"
	"backend.juicedbot.io/juiced.infrastructure/common/enums"
	"backend.juicedbot.io/juiced.infrastructure/common/events"
	"backend.juicedbot.io/juiced.sitescripts/base"
	"backend.juicedbot.io/juiced.sitescripts/util"
)

// CreateBestbuyMonitor takes a TaskGroup entity and turns it into a Bestbuy Monitor
func CreateBestbuyMonitor(taskGroup *entities.TaskGroup, proxy entities.Proxy, eventBus *events.EventBus, singleMonitors []entities.BestbuySingleMonitorInfo) (Monitor, error) {
	storedBestbuyMonitors := make(map[string]entities.BestbuySingleMonitorInfo)
	bestbuyMonitor := Monitor{}
	skus := []string{}

	for _, monitor := range singleMonitors {
		storedBestbuyMonitors[monitor.SKU] = monitor
		skus = append(skus, monitor.SKU)
	}

	client, err := util.CreateClient(proxy)
	if err != nil {
		return bestbuyMonitor, err
	}
	bestbuyMonitor = Monitor{
		Monitor: base.Monitor{
			TaskGroup: taskGroup,
			Proxy:     proxy,
			EventBus:  eventBus,
			Client:    client,
		},
		SKUs:        skus,
		SKUWithInfo: storedBestbuyMonitors,
	}

	return bestbuyMonitor, nil
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

func (monitor *Monitor) RunMonitor() {
	// If the function panics due to a runtime error, recover from it
	defer func() {
		recover()
		// TODO @silent: Let the UI know that a monitor failed
	}()

	if monitor.Monitor.TaskGroup.MonitorStatus == enums.MonitorIdle {
		monitor.PublishEvent(enums.WaitingForProductData, enums.MonitorStart)
	}
	needToStop := monitor.CheckForStop()
	if needToStop {
		return
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

	somethingInStock := monitor.GetSKUStock()

	if somethingInStock {
		needToStop := monitor.CheckForStop()
		if needToStop {
			return
		}
		monitor.PublishEvent(enums.SendingProductInfoToTasks, enums.MonitorUpdate)
		monitor.SendToTasks()
	} else {
		if len(monitor.RunningMonitors) > 0 {
			if monitor.Monitor.TaskGroup.MonitorStatus != enums.WaitingForInStock {
				monitor.PublishEvent(enums.WaitingForInStock, enums.MonitorUpdate)
			}
		}
		time.Sleep(time.Duration(monitor.Monitor.TaskGroup.MonitorDelay) * time.Millisecond)
		monitor.RunMonitor()
	}
}

func (monitor *Monitor) GetSKUStock() bool {
	skus := url.PathEscape(strings.Join(monitor.SKUs, ","))
	monitorResponse := MonitorResponse{}
	resp, _, err := util.MakeRequest(&util.Request{
		Client: monitor.Monitor.Client,
		Method: "GET",
		URL:    fmt.Sprintf(MonitorEndpoint, skus),
		RawHeaders: [][2]string{
			{"pragma", "no-cache"},
			{"cache-control", "no-cache"},
			{"sec-ch-ua", "\" Not A;Brand\";v=\"99\", \"Chromium\";v=\"90\", \"Google Chrome\";v=\"90\""},
			{"accept", "application/json"},
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

	stockData := events.BestbuySingleStockData{}
	switch resp.StatusCode {
	case 200:
		for i := range monitorResponse {
			sku := monitorResponse[i].Sku.Skuid
			monitor.RunningMonitors = append(monitor.RunningMonitors, sku)

			if monitorResponse[i].Sku.Buttonstate.Buttonstate == "ADD_TO_CART" {
				price := int(monitorResponse[i].Sku.Price.Currentprice)
				if (monitor.SKUWithInfo[sku].MaxPrice > price || monitor.SKUWithInfo[sku].MaxPrice == -1) && !util.InSlice(monitor.SKUsSentToTask, sku) {
					stockData.SKU = sku
					stockData.Price = int(monitorResponse[i].Sku.Price.Currentprice)
					monitor.SKUsSentToTask = append(monitor.SKUsSentToTask, sku)
				}
			} else {
				monitor.SKUsSentToTask = util.RemoveFromSlice(monitor.SKUsSentToTask, sku)
			}
		}
	}

	monitor.EventInfo = stockData

	return stockData.SKU != ""
}

// SendToTasks sends the product info to tasks
func (monitor *Monitor) SendToTasks() {
	data := events.BestbuyStockData{
		InStock: []events.BestbuySingleStockData{monitor.EventInfo},
	}
	monitor.Monitor.EventBus.PublishProductEvent(enums.BestBuy, data, monitor.Monitor.TaskGroup.GroupID)
}
