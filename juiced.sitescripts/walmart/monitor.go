package walmart

import (
	"fmt"
	"log"
	"math/rand"
	"strings"
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

// CreateWalmartMonitor takes a TaskGroup entity and turns it into a Walmart Monitor
func CreateWalmartMonitor(taskGroup *entities.TaskGroup, proxies []entities.Proxy, eventBus *events.EventBus, monitorType enums.MonitorType, skus []string) (Monitor, error) {
	walmartMonitor := Monitor{}

	walmartMonitor = Monitor{
		Monitor: base.Monitor{
			TaskGroup: taskGroup,
			Proxies:   proxies,
			EventBus:  eventBus,
		},
		MonitorType: monitorType,
		SKUs:        skus,
	}
	return walmartMonitor, nil
}

// PublishEvent wraps the EventBus's PublishMonitorEvent function
func (monitor *Monitor) PublishEvent(status enums.MonitorStatus, eventType enums.MonitorEventType) {
	monitor.Monitor.TaskGroup.SetMonitorStatus(status)
	monitor.Monitor.EventBus.PublishMonitorEvent(status, eventType, nil, monitor.Monitor.TaskGroup.GroupID)
}

//This checks if we want to stop
func (monitor *Monitor) CheckForStop() bool {
	if monitor.Monitor.StopFlag {
		monitor.PublishEvent(enums.MonitorIdle, enums.MonitorStop)
		return true
	}
	return false
}

//This is responsible for starting the Walmart Product monitor
func (monitor *Monitor) RunMonitor() {
	// If the function panics due to a runtime error, recover from it
	defer func() {
		if r := recover(); r != nil {
			log.Println(r)
			monitor.Monitor.StopFlag = true
			monitor.PublishEvent(enums.MonitorIdle, enums.MonitorFail)
		}
		monitor.PublishEvent(enums.MonitorIdle, enums.MonitorComplete)
	}()

	if monitor.Monitor.Client.Transport == nil {
		monitorClient, err := util.CreateClient()
		if err != nil {
			return
		}
		monitor.Monitor.Client = monitorClient
	}

	if monitor.Monitor.TaskGroup.MonitorStatus == enums.MonitorIdle {
		monitor.PublishEvent(enums.SettingUpMonitor, enums.MonitorStart)

		if monitor.PXValues.RefreshAt == 0 {
			go monitor.RefreshPX3()
			for monitor.PXValues.RefreshAt == 0 {
				needToStop := monitor.CheckForStop()
				if needToStop {
					return
				}
				time.Sleep(25 * time.Millisecond)
			}
		}
	}

	if monitor.Monitor.TaskGroup.MonitorStatus == enums.SettingUpMonitor {
		monitor.PublishEvent(enums.WaitingForProductData, enums.MonitorUpdate)
	}

	if len(monitor.Monitor.Proxies) > 0 {
		client.UpdateProxy(&monitor.Monitor.Client, common.ProxyCleaner(monitor.Monitor.Proxies[rand.Intn(len(monitor.Monitor.Proxies))]))
	}

	stockData := WalmartInStockData{}
	outOfStockForShip := make([]string, 0)

	switch monitor.MonitorType {
	case enums.SKUMonitor:
		stockData, outOfStockForShip = monitor.GetSkuStock(monitor.SKUs[0]) // TODO @silent
	}

	somethingInStock := false
	if stockData.OfferID != "" {
		somethingInStock = true
	}

	if somethingInStock {
		needToStop := monitor.CheckForStop()
		if needToStop {
			return
		}
		monitor.PublishEvent(enums.SendingProductInfoToTasks, enums.MonitorUpdate)
		monitor.InStockForShip = []WalmartInStockData{stockData}
	} else {
		if len(outOfStockForShip) > 0 {
			if monitor.Monitor.TaskGroup.MonitorStatus != enums.WaitingForInStock {
				monitor.PublishEvent(enums.WaitingForInStock, enums.MonitorUpdate)
			}
		}
		time.Sleep(time.Duration(monitor.Monitor.TaskGroup.MonitorDelay) * time.Millisecond)
		monitor.RunMonitor()
	}
}

// RefreshPX3 refreshes the px3 cookie every 4 minutes since it expires every 5 minutes
func (monitor *Monitor) RefreshPX3() {
	quit := make(chan bool)
	defer func() {
		quit <- true
		if r := recover(); r != nil {
			monitor.RefreshPX3()
		}
	}()

	cancellationToken := util.CancellationToken{Cancel: false}
	go func() {
		for {
			select {
			case <-quit:
				return
			default:
				needToStop := monitor.CheckForStop()
				if needToStop {
					cancellationToken.Cancel = true
					return
				}
			}
			time.Sleep(25 * time.Millisecond)
		}
	}()

	for {
		if monitor.PXValues.RefreshAt == 0 || time.Now().Unix() > monitor.PXValues.RefreshAt {
			pxValues, cancelled, err := SetPXCookie(monitor.Monitor.Proxy, &monitor.Monitor.Client, &cancellationToken)
			if cancelled {
				return
			}

			if err != nil {
				log.Println("Error setting px cookie for monitor: " + err.Error())
				panic(err)
			}
			monitor.PXValues = pxValues
			monitor.PXValues.RefreshAt = time.Now().Unix() + 240
		}
	}
}

func (monitor Monitor) HandlePXCap(resp *http.Response, redirectURL string) bool {
	quit := make(chan bool)
	defer func() {
		quit <- true
		if r := recover(); r != nil {
			monitor.HandlePXCap(resp, redirectURL)
		}
	}()

	cancellationToken := util.CancellationToken{Cancel: false}
	go func() {
		for {
			select {
			case <-quit:
				return
			default:
				needToStop := monitor.CheckForStop()
				if needToStop {
					cancellationToken.Cancel = true
					return
				}
			}
			time.Sleep(25 * time.Millisecond)
		}
	}()

	monitor.PublishEvent(enums.BypassingPXMonitor, enums.TaskUpdate)
	captchaURL := resp.Request.URL.String()
	if redirectURL != "" {
		captchaURL = BaseEndpoint + redirectURL[1:]
	}
	err := SetPXCapCookie(strings.ReplaceAll(captchaURL, "affil.", ""), &monitor.PXValues, monitor.Monitor.Proxy, &monitor.Monitor.Client, &cancellationToken)
	if err != nil {
		log.Println(err.Error())
		return false
	} else {
		log.Println("Cookie updated.")
		return true
	}
}

//This is for checking if a list of Skus are instock. Here we also check if there is a maximum price.
func (monitor *Monitor) GetSkuStock(sku string) (WalmartInStockData, []string) {
	stockData := WalmartInStockData{
		Sku: sku,
	}
	outOfStockForShip := make([]string, 0)

	resp, body, err := util.MakeRequest(&util.Request{
		Client: monitor.Monitor.Client,
		Method: "GET",
		URL:    fmt.Sprintf(MonitorEndpoint, sku),
		RawHeaders: [][2]string{
			{"accept", "text/html,application/xhtml+xml,application/xml;q=0.9,image/avif,image/webp,image/apng,*/*;q=0.8,application/signed-exchange;v=b3;q=0.9"},
			{"accept-encoding", "gzip, deflate, br"},
			{"accept-language", "en-US,en;q=0.9"},
			{"sec-ch-ua", `" Not;A Brand";v="99", "Google Chrome";v="91", "Chromium";v="91"`},
			{"sec-ch-ua-mobile", "?0"},
			{"sec-fetch-dest", "document"},
			{"sec-fetch-mode", "navigate"},
			{"sec-fetch-site", "none"},
			{"sec-fetch-user", "?1"},
			{"upgrade-insecure-requests", "1"},
			{"user-agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/91.0.4472.77 Safari/537.36"},
		},
	})
	if err != nil {
		fmt.Println(err.Error())
	}
	switch resp.StatusCode {
	case 200:
		if strings.Contains(resp.Request.URL.String(), "blocked") {
			monitor.HandlePXCap(resp, resp.Request.URL.String())
		} else {
			offerID := ""
			log.Println(body) // You can remove this, I'm just putting it here so line 220 doesn't error out since body is unused otherwise
			// TODO @Humphrey

			if offerID != "" {
				fmt.Printf("%s is in-stock.\n", sku)
			} else {
				if monitor.Monitor.TaskGroup.WalmartMonitorInfo.MaxPrice > -1 {
					fmt.Printf("%s is in-stock, but could not find an offerID below the monitor's MaxPrice.\n", sku)
				} else {
					fmt.Printf("%s is in-stock, but could not find an offerID.\n", sku)
				}
			}

			inSlice := false
			for _, monitorStock := range monitor.InStockForShip {
				if monitorStock.Sku == stockData.Sku {
					inSlice = true
				}
			}
			if !inSlice {
				stockData.OfferID = offerID
			}
		}
	case 404:
		fmt.Printf("We have a bad response: %v\n", resp.Status)
	default:
		fmt.Printf("Unkown Code: %v\n", resp.StatusCode)
	}

	return stockData, outOfStockForShip
}
