package bestbuy

import (
	"fmt"
	"math/rand"
	"net/url"
	"strings"
	"time"

	"backend.juicedbot.io/juiced.client/client"
	"backend.juicedbot.io/juiced.infrastructure/common"
	"backend.juicedbot.io/juiced.infrastructure/common/entities"
	"backend.juicedbot.io/juiced.infrastructure/common/enums"
	"backend.juicedbot.io/juiced.infrastructure/common/events"
	"backend.juicedbot.io/juiced.sitescripts/base"
	"backend.juicedbot.io/juiced.sitescripts/util"
)

// CreateBestbuyMonitor takes a TaskGroup entity and turns it into a Bestbuy Monitor
func CreateBestbuyMonitor(taskGroup *entities.TaskGroup, proxies []entities.Proxy, eventBus *events.EventBus, singleMonitors []entities.BestbuySingleMonitorInfo) (Monitor, error) {
	storedBestbuyMonitors := make(map[string]entities.BestbuySingleMonitorInfo)
	bestbuyMonitor := Monitor{}
	skus := []string{}

	for _, monitor := range singleMonitors {
		storedBestbuyMonitors[monitor.SKU] = monitor
		skus = append(skus, monitor.SKU)
	}

	bestbuyMonitor = Monitor{
		Monitor: base.Monitor{
			TaskGroup: taskGroup,
			Proxies:   proxies,
			EventBus:  eventBus,
		},
		SKUs:        skus,
		SKUWithInfo: storedBestbuyMonitors,
	}

	return bestbuyMonitor, nil
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
		monitor.PublishEvent(enums.MonitorIdle, enums.MonitorComplete, nil)
	}()

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
	}

	monitor.RunSingleMonitor()
}

func (monitor *Monitor) RunSingleMonitor() {
	needToStop := monitor.CheckForStop()
	if needToStop {
		return
	}

	if len(monitor.Monitor.Proxies) > 0 {
		client.UpdateProxy(&monitor.Monitor.Client, common.ProxyCleaner(monitor.Monitor.Proxies[rand.Intn(len(monitor.Monitor.Proxies))]))
	}

	stockData := monitor.GetSKUStock()
	needToStop = monitor.CheckForStop()
	if needToStop {
		return
	}

	if stockData.SKU != "" {
		var inSlice bool
		for _, monitorStock := range monitor.InStock {
			inSlice = monitorStock.SKU == stockData.SKU
		}
		// If the sku isn't already in the array of in-stock skus then add it to the array
		if !inSlice {
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
			if monitorStock.SKU == stockData.SKU {
				monitor.InStock = append(monitor.InStock[:i], monitor.InStock[i+1:]...)
				break
			}
		}

		time.Sleep(time.Duration(monitor.Monitor.TaskGroup.MonitorDelay) * time.Millisecond)
		monitor.RunSingleMonitor()
	}
}

func (monitor *Monitor) GetSKUStock() BestbuyInStockData {
	stockData := BestbuyInStockData{}
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

	switch resp.StatusCode {
	case 200:
		for i := range monitorResponse {
			sku := monitorResponse[i].Sku.Skuid
			monitor.RunningMonitors = append(monitor.RunningMonitors, sku)

			if monitorResponse[i].Sku.Buttonstate.Buttonstate == "ADD_TO_CART" {
				price := int(monitorResponse[i].Sku.Price.Currentprice)
				if monitor.SKUWithInfo[sku].MaxPrice > price || monitor.SKUWithInfo[sku].MaxPrice == -1 {
					stockData.SKU = sku
					stockData.ProductName = monitorResponse[i].Sku.Names.Short
					stockData.ImageURL = fmt.Sprintf("https://pisces.bbystatic.com/image2/BestBuy_US/images/products/%v/%v_sd.jpg;canvasHeight=500;canvasWidth=500", sku[:4], sku)
					stockData.Price = int(monitorResponse[i].Sku.Price.Currentprice)
					if !common.InSlice(monitor.SKUsSentToTask, sku) {
						monitor.SKUsSentToTask = append(monitor.SKUsSentToTask, sku)
					}
				}
			} else {
				monitor.SKUsSentToTask = common.RemoveFromSlice(monitor.SKUsSentToTask, sku)
			}
		}
	}

	return stockData
}
