package boxlunch

import (
	"fmt"
	"math/rand"
	"strconv"
	"sync"
	"time"

	"backend.juicedbot.io/juiced.client/client"
	"backend.juicedbot.io/juiced.infrastructure/common"
	"backend.juicedbot.io/juiced.infrastructure/common/entities"
	"backend.juicedbot.io/juiced.infrastructure/common/enums"
	"backend.juicedbot.io/juiced.infrastructure/common/events"
	"backend.juicedbot.io/juiced.sitescripts/base"
	"backend.juicedbot.io/juiced.sitescripts/util"
	"github.com/anaskhan96/soup"
	browser "github.com/eddycjy/fake-useragent"
)

// CreateboxlunchMonitor takes a TaskGroup entity and turns it into a boxlunch Monitor
func CreateBoxlunchMonitor(taskGroup *entities.TaskGroup, proxies []entities.Proxy, eventBus *events.EventBus, singleMonitors []entities.BoxlunchSingleMonitorInfo) (Monitor, error) {
	storedBoxlunchMonitors := make(map[string]entities.BoxlunchSingleMonitorInfo)
	boxlunchMonitor := Monitor{}

	pids := []string{}
	for _, monitor := range singleMonitors {
		storedBoxlunchMonitors[monitor.Pid] = monitor
		pids = append(pids, monitor.Pid)
	}

	boxlunchMonitor = Monitor{
		Monitor: base.Monitor{
			TaskGroup: taskGroup,
			Proxies:   proxies,
			EventBus:  eventBus,
		},
		Pids:        pids,
		PidWithInfo: storedBoxlunchMonitors,
	}

	return boxlunchMonitor, nil
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

	stockData := BoxlunchInStockData{}

	if len(monitor.Monitor.Proxies) > 0 {
		client.UpdateProxy(&monitor.Monitor.Client, common.ProxyCleaner(monitor.Monitor.Proxies[rand.Intn(len(monitor.Monitor.Proxies))]))
	}

	switch monitor.PidWithInfo[pid].MonitorType {
	case enums.SKUMonitor:
		stockData = monitor.StockMonitor(pid)
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

func (monitor *Monitor) StockMonitor(pid string) BoxlunchInStockData {
	stockData := BoxlunchInStockData{}
	resp, body, err := util.MakeRequest(&util.Request{
		Client: monitor.Monitor.Client,
		Method: "GET",
		URL:    fmt.Sprintf(MonitorEndpoint, pid),
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
		monitor.RunningMonitors = append(monitor.RunningMonitors, pid)
		return monitor.StockInfo(body, pid)
	case 404:
		monitor.PublishEvent(enums.UnableToFindProduct, enums.MonitorUpdate, nil)
		return stockData
	default:
		fmt.Printf("Unkown Code:%v", resp.StatusCode)
		return stockData
	}
}

func (monitor *Monitor) StockInfo(body string, pid string) BoxlunchInStockData {
	stockData := BoxlunchInStockData{}
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
	InBudget := monitor.PidWithInfo[pid].MaxPrice >= Price || monitor.PidWithInfo[pid].MaxPrice == -1

	if !InBudget {
		//not in budget return false
		return stockData
	}

	ProductName := doc.Find("a", "class", "name-link").Text()
	if !InBudget {
		//not in budget return false
		return stockData
	}

	//EventInfo updated now we return true
	return BoxlunchInStockData{
		PID:         pid,
		Size:        monitor.PidWithInfo[pid].Size,
		Color:       monitor.PidWithInfo[pid].Color,
		ProductName: ProductName,
		// Boxlunch and HotTopic use the same image links
		ImageURL: "https://hottopic.scene7.com/is/image/HotTopic/" + pid + "_hi",
	}
}
