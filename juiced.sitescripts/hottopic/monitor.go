package hottopic

import (
	"fmt"
	"net/http"
	"strconv"

	"backend.juicedbot.io/m/v2/juiced.infrastructure/common/entities"
	"backend.juicedbot.io/m/v2/juiced.infrastructure/common/enums"
	"backend.juicedbot.io/m/v2/juiced.infrastructure/common/events"
	"backend.juicedbot.io/m/v2/juiced.sitescripts/base"
	"backend.juicedbot.io/m/v2/juiced.sitescripts/util"
	browser "github.com/eddycjy/fake-useragent"
)

// CreateHottopicMonitor takes a TaskGroup entity and turns it into a Hottopic Monitor
func CreateHottopicMonitor(taskGroup *entities.TaskGroup, proxy entities.Proxy, eventBus *events.EventBus, singleMonitors []entities.HottopicSingleMonitorInfo) (Monitor, error) {
	storedHottopicMonitors := make(map[string]entities.HottopicSingleMonitorInfo)
	hottopicMonitor := Monitor{}

	pids := []string{}
	for _, monitor := range singleMonitors {
		client, err := util.CreateClient(proxy)
		if err != nil {
			return hottopicMonitor, err
		}
		storedHottopicMonitors[monitor.Pid] = entities.HottopicSingleMonitorInfo{
			Pid:    monitor.Pid,
			Client: client,
		}
		pids = append(pids, monitor.Pid)
		for created := false; !created; {
			hottopicMonitor = Monitor{
				Monitor: base.Monitor{
					TaskGroup: taskGroup,
					Proxy:     proxy,
					EventBus:  eventBus,
				},
				Pids: pids,
			}
			created = true
		}
	}
	return hottopicMonitor, nil
}

// PublishEvent wraps the EventBus's PublishMonitorEvent function
func (monitor *Monitor) PublishEvent(status enums.MonitorStatus, eventType enums.MonitorEventType) {
	monitor.Monitor.TaskGroup.SetMonitorStatus(status)
	monitor.Monitor.EventBus.PublishMonitorEvent(status, eventType, nil, monitor.Monitor.TaskGroup.GroupID)
}

func (monitor *Monitor) CheckForStop() bool {
	if monitor.Monitor.StopFlag {
		monitor.PublishEvent(enums.MonitorIdle, enums.MonitorStop)
		return true
	}
	return false
}

func (monitor *Monitor) RunMonitor() {
	if monitor.Monitor.TaskGroup.MonitorStatus == enums.MonitorIdle {
		monitor.PublishEvent(enums.WaitingForProductData, enums.MonitorStart)
	}
	needToStop := monitor.CheckForStop()
	if needToStop {
		return
	}
	for _, pid := range monitor.Pids {
		somethingInStock := false
		switch monitor.PidWithInfo[pid].MonitorType {
		case enums.SKUMonitor:
			somethingInStock = monitor.StockMonitor(pid)
		}
		if somethingInStock {
			needToStop := monitor.CheckForStop()
			if needToStop {
				return
			}
			monitor.RunningMonitors = util.RemoveFromSlice(monitor.RunningMonitors, pid)
			monitor.PublishEvent(enums.SendingProductInfoToTasks, enums.MonitorUpdate)
			monitor.SendToTasks()
		}

	}
}

func (monitor *Monitor) StockMonitor(pid string) bool {
	var client http.Client
	client = monitor.PidWithInfo[pid].Client
	resp, err := util.MakeRequest(&util.Request{
		Client: client,
		Method: "GET",
		URL:    MonitorEndpoint + pid,
		RawHeaders: [][2]string{
			{"sec-ch-ua", `" Not A;Brand";v="99", "Chromium";v="90", "Google Chrome";v="90"`},
			{"sec-ch-ua-mobile", "?0"},
			{"upgrade-insecure-requests", "1"},
			{"user-agent", browser.Chrome()},
			{"accept", "text/html,application/xhtml+xml,application/xml;q=0.9,image/avif,image/webp,image/apng,*/*;q=0.8,application/signed-exchange;v=b3;q=0.9"},
			{"sec-fetch-site", "none"},
			{"sec-fetch-mode", "navigate"},
			{"sec-fetch-user", "?1"},
			{"sec-fetch-dest", "document"},
			{"accept-encoding", "gzip, deflate, br"},
			{"accept-language", "en-US,en;q=0.9"},
		},
	})
	if err != nil {
		fmt.Println(err)
	}

	defer resp.Body.Close()

	switch resp.StatusCode {
	case 200:
		test := monitor.StockInfo(resp, pid)
		fmt.Println(test)
		return test
	case 404:
		monitor.PublishEvent(enums.UnableToFindProduct, enums.MonitorUpdate)
		return false
	default:
		fmt.Printf("Unkown Code:%v", resp.StatusCode)
		return false
	}
}

func (monitor *Monitor) StockInfo(resp *http.Response, pid string) bool {
	//use soup to parse the HTML and get values we want
	//
	var priceStr string
	price, _ := strconv.Atoi(priceStr)
	inBudget := monitor.PidWithInfo[pid].MaxPrice > price
	if inBudget {
		monitor.EventInfo = events.HotTopicSingleStockData{
			PID:         pid,
			MonitorType: enums.SKUMonitor,
			//anymore data you want to scrape put here such as item name, size, color etc.
		}
	}

	return inBudget
}
func (monitor *Monitor) SendToTasks() {
	data := events.HottopicStockData{
		InStock: []events.HotTopicSingleStockData{monitor.EventInfo},
	}
	monitor.Monitor.EventBus.PublishProductEvent(enums.Hottopic, data, monitor.Monitor.TaskGroup.GroupID)
}
