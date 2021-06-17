package boxlunch

import (
	"fmt"
	"strconv"

	"backend.juicedbot.io/juiced.infrastructure/common/entities"
	"backend.juicedbot.io/juiced.infrastructure/common/enums"
	"backend.juicedbot.io/juiced.infrastructure/common/events"
	"backend.juicedbot.io/juiced.sitescripts/base"
	"backend.juicedbot.io/juiced.sitescripts/util"
	"github.com/anaskhan96/soup"
	browser "github.com/eddycjy/fake-useragent"
)

// CreateboxlunchMonitor takes a TaskGroup entity and turns it into a boxlunch Monitor
func CreateBoxlunchMonitor(taskGroup *entities.TaskGroup, proxy entities.Proxy, eventBus *events.EventBus, singleMonitors []entities.BoxLunchSingleMonitorInfo) (Monitor, error) {
	storedBoxlunchMonitors := make(map[string]entities.BoxLunchSingleMonitorInfo)
	boxlunchMonitor := Monitor{}

	pids := []PidSingle{}
	for _, monitor := range singleMonitors {
		storedBoxlunchMonitors[monitor.Pid] = entities.BoxLunchSingleMonitorInfo{
			Pid: monitor.Pid,
		}
		pidV := PidSingle{
			Pid:   monitor.Pid,
			size:  monitor.Size,
			color: monitor.Color,
		}
		pids = append(pids, pidV)
		for created := false; !created; {
			boxlunchMonitor = Monitor{
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
	return boxlunchMonitor, nil
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
		switch monitor.PidWithInfo[pid.Pid].MonitorType {
		case enums.SKUMonitor:
			somethingInStock = monitor.StockMonitor(pid)
		}
		if somethingInStock {
			needToStop := monitor.CheckForStop()
			if needToStop {
				return
			}
			monitor.RunningMonitors = util.RemoveFromSlice(monitor.RunningMonitors, pid.Pid)
			monitor.PublishEvent(enums.SendingProductInfoToTasks, enums.MonitorUpdate)
			monitor.SendToTasks()
		}

	}
}

func (monitor *Monitor) StockMonitor(pid PidSingle) bool {
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
	})
	if err != nil {
		fmt.Println(err)
	}

	defer resp.Body.Close()

	switch resp.StatusCode {
	case 200:
		stockInfo := monitor.StockInfo(body, pid)
		return stockInfo
	case 404:
		monitor.PublishEvent(enums.UnableToFindProduct, enums.MonitorUpdate)
		return false
	default:
		fmt.Printf("Unkown Code:%v", resp.StatusCode)
		return false
	}
}

func (monitor *Monitor) StockInfo(body string, pid PidSingle) bool {
	doc := soup.HTMLParse(body)

	ShipTable := doc.Find("div", "class", "method-descr__label")
	InStock := ShipTable.Find("span", "class", "color-green").Text() == "In stock"
	if !InStock {
		InStock = ShipTable.Find("span", "class", "text-red").Text() == "Backorder"
	}
	if !InStock {
		//not instock or backorder return false
		return InStock
	}

	//We are in stock for this size/color, lets check price is in budget.
	PriceText := doc.Find("span", "class", "productdetail__info-pricing-original").Text()
	Price, _ := strconv.Atoi(PriceText)
	InBudget := monitor.PidWithInfo[pid.Pid].MaxPrice > Price

	if !InBudget {
		//not in budget return false
		return InBudget
	}

	//we are in stock and in budget
	monitor.EventInfo = events.BoxLunchSingleStockData{
		PID: pid.Pid,
	}

	//EventInfo updated now we return true
	return true
}
func (monitor *Monitor) SendToTasks() {
	data := events.BoxLunchStockData{
		InStock: []events.BoxLunchSingleStockData{monitor.EventInfo},
	}
	monitor.Monitor.EventBus.PublishProductEvent(enums.BoxLunch, data, monitor.Monitor.TaskGroup.GroupID)
}
