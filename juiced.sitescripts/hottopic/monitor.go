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
	"github.com/anaskhan96/soup"
	browser "github.com/eddycjy/fake-useragent"
)

// CreateHottopicMonitor takes a TaskGroup entity and turns it into a Hottopic Monitor
func CreateHottopicMonitor(taskGroup *entities.TaskGroup, proxy entities.Proxy, eventBus *events.EventBus, singleMonitors []entities.HottopicSingleMonitorInfo) (Monitor, error) {
	storedHottopicMonitors := make(map[string]entities.HottopicSingleMonitorInfo)
	hottopicMonitor := Monitor{}

	pids := []PidSingle{}
	for _, monitor := range singleMonitors {
		client, err := util.CreateClient(proxy)
		if err != nil {
			return hottopicMonitor, err
		}
		storedHottopicMonitors[monitor.Pid] = entities.HottopicSingleMonitorInfo{
			Pid:    monitor.Pid,
			Client: client,
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
	var client http.Client
	BuildEndpoint := MonitorEndpoint + pid.Pid

	if len(pid.size) > 0 {
		BuildEndpoint = BuildEndpoint + "&dwvar_" + pid.Pid + "_size=" + pid.size
	}
	if len(pid.color) > 0 {
		BuildEndpoint = BuildEndpoint + "&dwvar_" + pid.Pid + "_color=" + pid.color
	}
	client = monitor.PidWithInfo[pid.Pid].Client
	resp, err := util.MakeRequest(&util.Request{
		Client: client,
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
		stockInfo := monitor.StockInfo(resp, pid)
		return stockInfo
	case 404:
		monitor.PublishEvent(enums.UnableToFindProduct, enums.MonitorUpdate)
		return false
	default:
		fmt.Printf("Unkown Code:%v", resp.StatusCode)
		return false
	}
}

func (monitor *Monitor) StockInfo(resp *http.Response, pid PidSingle) bool {
	body := util.ReadBody(resp)
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
	monitor.EventInfo = events.HotTopicSingleStockData{
		PID:         pid.Pid,
		MonitorType: enums.SKUMonitor,
	}

	//EventInfo updated now we return true
	return true
}
func (monitor *Monitor) SendToTasks() {
	data := events.HottopicStockData{
		InStock: []events.HotTopicSingleStockData{monitor.EventInfo},
	}
	monitor.Monitor.EventBus.PublishProductEvent(enums.Hottopic, data, monitor.Monitor.TaskGroup.GroupID)
}
