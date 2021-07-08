package walmart

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"backend.juicedbot.io/juiced.infrastructure/common/entities"
	"backend.juicedbot.io/juiced.infrastructure/common/enums"
	"backend.juicedbot.io/juiced.infrastructure/common/events"
	"backend.juicedbot.io/juiced.sitescripts/base"
	"backend.juicedbot.io/juiced.sitescripts/util"
	"github.com/anaskhan96/soup"
)

// CreateWalmartMonitor takes a TaskGroup entity and turns it into a Walmart Monitor
func CreateWalmartMonitor(taskGroup *entities.TaskGroup, proxy entities.Proxy, eventBus *events.EventBus, monitorType enums.MonitorType, skus []string) (Monitor, error) {
	walmartMonitor := Monitor{}
	client, err := util.CreateClient(proxy)
	if err != nil {
		return walmartMonitor, err
	}
	walmartMonitor = Monitor{
		Monitor: base.Monitor{
			TaskGroup: taskGroup,
			Proxy:     proxy,
			EventBus:  eventBus,
			Client:    client,
		},
		MonitorType: monitorType,
		SKUs:        skus,
	}
	return walmartMonitor, err
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
		if recover() != nil {
			monitor.Monitor.StopFlag = true
			monitor.PublishEvent(enums.MonitorIdle, enums.MonitorFail)
		}
		monitor.PublishEvent(enums.MonitorIdle, enums.MonitorComplete)
	}()

	if monitor.Monitor.TaskGroup.MonitorStatus == enums.MonitorIdle {
		monitor.PublishEvent(enums.WaitingForProductData, enums.MonitorStart)
	}
	if monitor.PXValues.RefreshAt == 0 {
		go monitor.RefreshPX3()
		for monitor.PXValues.RefreshAt == 0 {
		}
	}
	needToStop := monitor.CheckForStop()
	if needToStop {
		return
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
	defer func() {
		recover()
		monitor.RefreshPX3()
	}()

	for {
		if monitor.PXValues.RefreshAt == 0 || time.Now().Unix() > monitor.PXValues.RefreshAt {
			pxValues, err := SetPXCookie(monitor.Monitor.Proxy, &monitor.Monitor.Client)

			if err != nil {
				return // TODO @silent
			}
			monitor.PXValues = pxValues
			monitor.PXValues.RefreshAt = time.Now().Unix() + 240
		}
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
		Task:    base.Task{},
		Monitor: monitor.Monitor,
	})
	if err != nil {
		fmt.Println(err.Error())
	}
	switch resp.StatusCode {
	case 200:
		if strings.Contains(resp.Request.URL.String(), "blocked") {
			err := SetPXCapCookie(strings.ReplaceAll(resp.Request.URL.String(), "affil.", ""), &monitor.PXValues, monitor.Monitor.Proxy, &monitor.Monitor.Client)
			if err != nil {
				fmt.Println(err.Error())
			}
			fmt.Println("Cookie updated.")
		} else {
			offerID := ""
			responseBody := soup.HTMLParse(string(body))
			sellers := responseBody.FindAll("div", "class", "product-seller-card")
			for _, seller := range sellers {
				sellerNameLink := seller.Find("a", "class", "seller-shipping-msg")
				if sellerNameLink.Error == nil {
					// Check if sold by Walmart first
					if sellerNameLink.Text() == "Walmart" {
						scriptText := ""
						script := responseBody.Find("script", "id", "items")
						if script.Error != nil {
							script = responseBody.Find("script", "id", "item")
						}
						if script.Error == nil {
							scriptText = script.Text()
						}
						if scriptText != "" {
							offerID, _ = util.FindInString(scriptText, `"carePlans":{"`, `"`)
							break
						}
					}
					// If not sold by Walmart, offerID will still be empty string, so continue
					// If sold by Walmart, but we couldn't find Walmart's offerID, continue
					if offerID == "" {
						// If MaxPrice is set, check price
						if monitor.Monitor.TaskGroup.WalmartMonitorInfo.MaxPrice > -1 {
							priceDiv := seller.Find("span", "class", "price")
							if priceDiv.Error == nil {
								priceSpan := priceDiv.Find("span", "class", "visuallyhidden")
								if priceSpan.Error == nil {
									priceStr := strings.ReplaceAll(priceSpan.Text(), "$", "")
									price, err := strconv.ParseFloat(priceStr, 64)
									if err == nil {
										// If price is good, set offerID
										if price < float64(monitor.Monitor.TaskGroup.WalmartMonitorInfo.MaxPrice) {
											href, ok := sellerNameLink.Attrs()["href"]
											if ok && strings.Contains(href, "offerId=") {
												offerID = strings.Split(href, "offerId=")[1]
												break
											}
										}
									}
								}
							}
							// Otherwise, just set offerID
						} else {
							href, ok := sellerNameLink.Attrs()["href"]
							if ok && strings.Contains(href, "offerId=") {
								offerID = strings.Split(href, "offerId=")[1]
								break
							}
						}
					}
				}
			}

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
