package topps

import (
	"fmt"
	"math/rand"
	"strconv"
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
	"backend.juicedbot.io/juiced.sitescripts/hawk-go"
	"backend.juicedbot.io/juiced.sitescripts/util"
	"github.com/anaskhan96/soup"
)

// CreateToppsMonitor takes a TaskGroup entity and turns it into a Topps Monitor
func CreateToppsMonitor(taskGroup *entities.TaskGroup, proxies []entities.Proxy, eventBus *events.EventBus, singleMonitors []entities.ToppsSingleMonitorInfo) (Monitor, error) {
	storedToppsMonitors := make(map[string]entities.ToppsSingleMonitorInfo)

	items := []string{}

	for _, monitor := range singleMonitors {
		storedToppsMonitors[monitor.Item] = monitor
		items = append(items, monitor.Item)
	}

	toppsMonitor := Monitor{
		Monitor: base.Monitor{
			TaskGroup: taskGroup,
			Proxies:   proxies,
			EventBus:  eventBus,
		},
		Items:        items,
		ItemWithInfo: storedToppsMonitors,
	}

	return toppsMonitor, nil
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
		monitor.Monitor.Scraper = hawk.Init(monitorClient, common.HAWK_KEY, false)

		if len(monitor.Monitor.Proxies) > 0 {
			client.UpdateProxy(&monitor.Monitor.Scraper.Client, common.ProxyCleaner(monitor.Monitor.Proxies[rand.Intn(len(monitor.Monitor.Proxies))]))
		}

		becameGuest := false
		for !becameGuest {
			needToStop := monitor.CheckForStop()
			if needToStop {
				return
			}
			becameGuest = BecomeGuest(monitor.Monitor.Scraper)
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
	wg.Add(len(monitor.Items))
	for _, item := range monitor.Items {
		go func(x string) {
			monitor.RunSingleMonitor(x)
			wg.Done()
		}(item)
	}
	wg.Wait()

}

func (monitor *Monitor) RunSingleMonitor(item string) {
	needToStop := monitor.CheckForStop()
	if needToStop {
		return
	}

	if !common.InSlice(monitor.RunningMonitors, item) {
		defer func() {
			recover()
			// TODO @silent: Re-run this specific monitor
		}()

		if len(monitor.Monitor.Proxies) > 0 {
			client.UpdateProxy(&monitor.Monitor.Scraper.Client, common.ProxyCleaner(monitor.Monitor.Proxies[rand.Intn(len(monitor.Monitor.Proxies))]))
		}

		stockData := monitor.GetItemStock(item)
		if stockData.SKU != "" && stockData.AddURL != "" && stockData.FormKey != "" {
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
				monitor.RunningMonitors = common.RemoveFromSlice(monitor.RunningMonitors, item)
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
				if monitorStock.SKU == stockData.SKU {
					monitor.InStock = append(monitor.InStock[:i], monitor.InStock[i+1:]...)
					break
				}
			}
			time.Sleep(time.Duration(monitor.Monitor.TaskGroup.MonitorDelay) * time.Millisecond)
			monitor.RunSingleMonitor(item)
		}
	}
}

func (monitor *Monitor) GetItemStock(item string) ToppsInStockData {
	stockData := ToppsInStockData{}
	resp, body, err := util.MakeRequest(&util.Request{
		Scraper: monitor.Monitor.Scraper,
		Method:  "GET",
		URL:     fmt.Sprintf(MonitorEndpoint, item),
		RawHeaders: http.RawHeader{
			{"cache-control", "no-store"},
			{"pragma", "no-cache"},
			{"sec-ch-ua", `"Chromium";v="92", " Not A;Brand";v="99", "Google Chrome";v="92"`},
			{"sec-ch-ua-mobile", `?0`},
			{"upgrade-insecure-requests", `1`},
			{"user-agent", `Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/92.0.4515.131 Safari/537.36`},
			{"accept", `text/html,application/xhtml+xml,application/xml;q=0.9,image/avif,image/webp,image/apng,*/*;q=0.8,application/signed-exchange;v=b3;q=0.9`},
			{"sec-fetch-site", `same-origin`},
			{"sec-fetch-mode", `navigate`},
			{"sec-fetch-dest", `document`},
			// Bypassing cache by having a random origin
			{"origin", RandomString("qwertyuiopasdfghjklzxcvbnm1234567890", 10)},
			{"referer", BaseEndpoint + "/"},
			{"accept-encoding", `gzip, deflate`},
			{"accept-language", `en-US,en;q=0.9`},
		},
	})
	if err != nil || resp.StatusCode != 200 {
		fmt.Println(err)
		return stockData
	}

	return monitor.ParseInfos(item, body)
}

func (monitor *Monitor) ParseInfos(item, body string) ToppsInStockData {
	stockData := ToppsInStockData{Item: item}

	doc := soup.HTMLParse(body)

	elem := doc.Find("button", "id", "product-addtocart-button")
	if elem.Error != nil {
		return stockData
	}
	if elem.Find("span").Text() != "Add to Cart" {
		return stockData
	}

	elem = doc.Find("span", "class", "price")
	if elem.Error != nil {
		return stockData
	}

	price, err := strconv.ParseFloat(strings.ReplaceAll(elem.Text(), "$", ""), 64)
	if err != nil {
		fmt.Println(err)
		return stockData
	}
	stockData.Price = price

	elem = doc.Find("img", "class", "gallery-placeholder__image")
	if elem.Error != nil {
		return stockData
	}
	stockData.ImageURL = elem.Attrs()["src"]

	elem = doc.Find("title")
	if elem.Error != nil {
		return stockData
	}
	stockData.ProductName = elem.Text()

	elem = doc.Find("div", "class", "product-add-form")
	if elem.Error != nil {
		return stockData
	}

	elems := elem.FindAll("input")
	for i := range elems {
		if _, ok := elems[i].Attrs()["name"]; ok {
			switch elems[i].Attrs()["name"] {
			case "product":
				stockData.SKU = elems[i].Attrs()["value"]
			case "form_key":
				stockData.FormKey = elems[i].Attrs()["value"]
			}
		}
	}
	if stockData.SKU == "" || stockData.FormKey == "" {
		return stockData
	}

	elem = elem.Find("form")
	if elem.Error != nil {
		return stockData
	}

	stockData.AddURL = elem.Attrs()["action"]
	elems = doc.FindAll("select")
	fmt.Println(elems[0].Children())

	var options []Option
	elems = doc.FindAll("select")
	for i := range elems {
		if strings.Contains(elems[i].Attrs()["id"], "select") {
			stockData.OptionName = elems[i].Attrs()["name"]
			for _, child := range elems[i].Children() {
				fmt.Println(child.Text())
				tempPrice, _ := strconv.ParseFloat(child.Attrs()["price"], 64)
				fmt.Println(tempPrice)
				if float64(monitor.ItemWithInfo[item].MaxPrice) > tempPrice || monitor.ItemWithInfo[item].MaxPrice == -1 {
					options = append(options, Option{child.Attrs()["value"], tempPrice})
				}
			}
		}
	}

	if len(options) != 0 {
		stockData.Price += options[len(options)-1].Price
		stockData.OptionID += options[len(options)-1].ID
	} else {
		return ToppsInStockData{}
	}

	stockData.ItemURL = BaseEndpoint + "/" + item

	return stockData
}
