package hottopic

import (
	"fmt"
	"strconv"

	"backend.juicedbot.io/juiced.infrastructure/common"
	"backend.juicedbot.io/juiced.infrastructure/common/entities"
	"backend.juicedbot.io/juiced.infrastructure/common/enums"
	"backend.juicedbot.io/juiced.infrastructure/common/events"
	"backend.juicedbot.io/juiced.sitescripts/base"
	"backend.juicedbot.io/juiced.sitescripts/util"
	"github.com/anaskhan96/soup"
	browser "github.com/eddycjy/fake-useragent"
)

// CreateHottopicMonitor takes a TaskGroup entity and turns it into a Hottopic Monitor
func CreateHottopicMonitor(taskGroup *entities.TaskGroup, proxy entities.Proxy, eventBus *events.EventBus, singleMonitors []entities.HottopicSingleMonitorInfo) (Monitor, error) {
	storedHottopicMonitors := make(map[string]entities.HottopicSingleMonitorInfo)
	hottopicMonitor := Monitor{}

	pids := []PidSingle{}
	for _, monitor := range singleMonitors {
		storedHottopicMonitors[monitor.Pid] = entities.HottopicSingleMonitorInfo{
			Pid: monitor.Pid,
		}
		pidV := PidSingle{
			Pid:   monitor.Pid,
			size:  monitor.Size,
			color: monitor.Color,
		}
		pids = append(pids, pidV)
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
		stockData := HotTopicInStockData{}
		switch monitor.PidWithInfo[pid.Pid].MonitorType {
		case enums.SKUMonitor:
			stockData = monitor.StockMonitor(pid)
		}
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
				monitor.RunningMonitors = common.RemoveFromSlice(monitor.RunningMonitors, pid.Pid)
				monitor.PublishEvent(enums.SendingProductInfoToTasks, enums.MonitorUpdate)
			}
		}

	}
}

func (monitor *Monitor) StockMonitor(pid PidSingle) HotTopicInStockData {
	stockData := HotTopicInStockData{}
	BuildEndpoint := MonitorEndpoint + pid.Pid

	//Values have to be exact and case sensistive
	//XXS
	//SM
	//MD
	//LG
	//XL
	//2X
	//3X
	if len(pid.size) > 0 {
		BuildEndpoint = BuildEndpoint + "&dwvar_" + pid.Pid + "_size=" + pid.size
	}

	//Default seems to be BLACK case sensitive
	if len(pid.color) > 0 {
		BuildEndpoint = BuildEndpoint + "&dwvar_" + pid.Pid + "_color=" + pid.color
	}

	resp, body, err := util.MakeRequest(&util.Request{
		Client: monitor.Monitor.Client,
		Method: "GET",
		URL:    BuildEndpoint,
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
		Task:    base.Task{},
		Monitor: monitor.Monitor,
	})
	if err != nil {
		fmt.Println(err)
		return stockData
	}

	switch resp.StatusCode {
	case 200:
		return monitor.StockInfo(body, pid)
	case 404:
		monitor.PublishEvent(enums.UnableToFindProduct, enums.MonitorUpdate)
		return stockData
	default:
		fmt.Printf("Unkown Code:%v", resp.StatusCode)
		return stockData
	}
}

func (monitor *Monitor) StockInfo(body string, pid PidSingle) HotTopicInStockData {
	stockData := HotTopicInStockData{}
	doc := soup.HTMLParse(body)

	ShipTable := doc.Find("div", "class", "method-descr__label")
	InStock := ShipTable.Find("span", "class", "color-green").Text() == "In stock"
	if !InStock {
		InStock = ShipTable.Find("span", "class", "text-red").Text() == "Backorder"
	}
	if !InStock {
		//not instock or backorder return false
		return stockData
	}

	//We are in stock for this size/color, lets check price is in budget.
	PriceText := doc.Find("span", "class", "productdetail__info-pricing-original").Text()
	Price, _ := strconv.Atoi(PriceText)
	InBudget := monitor.PidWithInfo[pid.Pid].MaxPrice > Price

	if !InBudget {
		//not in budget return false
		return stockData
	}

	//we are in stock and in budget
	return HotTopicInStockData{
		PID: pid.Pid,
	}
}
