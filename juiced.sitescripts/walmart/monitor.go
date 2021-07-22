package walmart

import (
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
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
	"backend.juicedbot.io/juiced.sitescripts/util"
)

// CreateWalmartMonitor takes a TaskGroup entity and turns it into a Walmart Monitor
func CreateWalmartMonitor(taskGroup *entities.TaskGroup, proxies []entities.Proxy, eventBus *events.EventBus, singleMonitors []entities.WalmartSingleMonitorInfo) (Monitor, error) {
	storedWalmartMonitors := make(map[string]entities.WalmartSingleMonitorInfo)
	walmartMonitor := Monitor{}
	skus := []string{}

	for _, monitor := range singleMonitors {
		storedWalmartMonitors[monitor.SKU] = monitor
		skus = append(skus, monitor.SKU)
	}

	walmartMonitor = Monitor{
		Monitor: base.Monitor{
			TaskGroup: taskGroup,
			Proxies:   proxies,
			EventBus:  eventBus,
		},
		SKUs:        skus,
		SKUWithInfo: storedWalmartMonitors,
	}

	return walmartMonitor, nil
}

// PublishEvent wraps the EventBus's PublishMonitorEvent function
func (monitor *Monitor) PublishEvent(status enums.MonitorStatus, eventType enums.MonitorEventType, data interface{}) {
	monitor.Monitor.TaskGroup.SetMonitorStatus(status)
	monitor.Monitor.EventBus.PublishMonitorEvent(status, eventType, data, monitor.Monitor.TaskGroup.GroupID)
}

//This checks if we want to stop
func (monitor *Monitor) CheckForStop() bool {
	if monitor.Monitor.StopFlag {
		monitor.PublishEvent(enums.MonitorIdle, enums.MonitorStop, nil)
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
			monitor.PublishEvent(enums.MonitorIdle, enums.MonitorFail, nil)
		}
		monitor.PublishEvent(enums.MonitorIdle, enums.MonitorComplete, nil)
	}()

	if monitor.Monitor.Client.Transport == nil {
		monitorClient, err := util.CreateClient()
		if err != nil {
			return
		}
		monitor.Monitor.Client = monitorClient
	}

	if monitor.Monitor.TaskGroup.MonitorStatus == enums.MonitorIdle {
		monitor.PublishEvent(enums.SettingUpMonitor, enums.MonitorStart, nil)

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
		monitor.PublishEvent(enums.WaitingForProductData, enums.MonitorUpdate, nil)
	}

	if len(monitor.Monitor.Proxies) > 0 {
		client.UpdateProxy(&monitor.Monitor.Client, common.ProxyCleaner(monitor.Monitor.Proxies[rand.Intn(len(monitor.Monitor.Proxies))]))
	}

	needToStop := monitor.CheckForStop()
	if needToStop {
		return
	}

	wg := sync.WaitGroup{}
	wg.Add(len(monitor.SKUs))
	for _, sku := range monitor.SKUs {
		go func(x string) {
			monitor.RunSingleMonitor(x)
			wg.Done()
		}(sku)
	}
	wg.Wait()
}

func (monitor *Monitor) RunSingleMonitor(sku string) {
	needToStop := monitor.CheckForStop()
	if needToStop {
		return
	}

	if !common.InSlice(monitor.RunningMonitors, sku) {
		defer func() {
			recover()
			// TODO @silent: Re-run this specific monitor
		}()

		if len(monitor.Monitor.Proxies) > 0 {
			client.UpdateProxy(&monitor.Monitor.Client, common.ProxyCleaner(monitor.Monitor.Proxies[rand.Intn(len(monitor.Monitor.Proxies))]))
		}

		stockData := WalmartInStockData{}

		switch monitor.SKUWithInfo[sku].MonitorType {
		case enums.SKUMonitor:
			stockData = monitor.GetSkuStock(sku)
		}

		if stockData.OfferID != "" {
			needToStop := monitor.CheckForStop()
			if needToStop {
				return
			}

			var inSlice bool
			for _, monitorStock := range monitor.InStockForShip {
				inSlice = monitorStock.SKU == stockData.SKU
			}
			if !inSlice {
				monitor.InStockForShip = append(monitor.InStockForShip, stockData)
				monitor.RunningMonitors = common.RemoveFromSlice(monitor.RunningMonitors, sku)
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
			for i, monitorStock := range monitor.InStockForShip {
				if monitorStock.SKU == stockData.SKU {
					monitor.InStockForShip = append(monitor.InStockForShip[:i], monitor.InStockForShip[i+1:]...)
					break
				}
			}
			time.Sleep(time.Duration(monitor.Monitor.TaskGroup.MonitorDelay) * time.Millisecond)
			monitor.RunSingleMonitor(sku)
		}
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

	monitor.PublishEvent(enums.BypassingPXMonitor, enums.TaskUpdate, nil)
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

//This is for checking if a sku is instock. Here we also check if there is a maximum price.
func (monitor *Monitor) GetSkuStock(sku string) WalmartInStockData {
	stockData := WalmartInStockData{
		SKU: sku,
	}

	monitorResponse := MonitorResponse{}
	resp, _, err := util.MakeRequest(&util.Request{
		Client: monitor.Monitor.Client,
		Method: "GET",
		URL:    fmt.Sprintf(MonitorEndpoint, sku),
		RawHeaders: [][2]string{
			{"pragma", `no-cache`},
			{"cache-control", `no-cache`},
			{"sec-ch-ua", `" Not;A Brand";v="99", "Google Chrome";v="91", "Chromium";v="91"`},
			{"sec-ch-ua-mobile", `?0`},
			{"upgrade-insecure-requests", `1`},
			{"user-agent", `Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/91.0.4472.164 Safari/537.36`},
			{"accept", `text/html,application/xhtml+xml,application/xml;q=0.9,image/avif,image/webp,image/apng,*/*;q=0.8,application/signed-exchange;v=b3;q=0.9`},
			{"service-worker-navigation-preload", `true`},
			{"sec-fetch-site", `none`},
			{"sec-fetch-mode", `navigate`},
			{"sec-fetch-user", `?1`},
			{"sec-fetch-dest", `document`},
			{"accept-encoding", `gzip, deflate, br`},
			{"accept-language", `en-US,en;q=0.9`},
		},

		ResponseBodyStruct: &monitorResponse,
	})
	if err != nil {
		fmt.Println(err.Error())
		return stockData
	}
	switch resp.StatusCode {
	case 200:
		if strings.Contains(resp.Request.URL.String(), "blocked") {
			monitor.HandlePXCap(resp, resp.Request.URL.String())
		} else {
			var inStock bool

			// Item exists = Add to running monitors
			if len(monitorResponse.Payload.Offers.(map[string]interface{})) > 0 {
				monitor.RunningMonitors = append(monitor.RunningMonitors, sku)
			}
			for _, value := range monitorResponse.Payload.Offers.(map[string]interface{}) {
				offer := Offer{}
				tempJson, err := json.Marshal(value.(map[string]interface{}))
				if err != nil {
					return stockData
				}

				err = json.Unmarshal(tempJson, &offer)
				if err != nil {
					return stockData
				}

				if offer.Productavailability.Availabilitystatus == "IN_STOCK" {
					if monitor.SKUWithInfo[sku].MaxPrice > int(offer.Pricesinfo.Pricemap.Current.Price) {
						if monitor.SKUWithInfo[sku].SoldByWalmart {
							if offer.Sellerid != "F55CDC31AB754BB68FE0B39041159D63" {
								break
							}
							inStock = true
						} else {
							inStock = true
						}
					}
					if inStock {
						stockData.OfferID = offer.OfferInfo.OfferID
						stockData.MaxQty = offer.OfferInfo.QuantityOptions.OrderLimit
					}
				}
			}

			if inStock {
				product := Product{}
				tempJson, err := json.Marshal(monitorResponse.Payload.Products.(map[string]interface{})[monitorResponse.Payload.PrimaryProduct])
				if err != nil {
					return stockData
				}
				err = json.Unmarshal(tempJson, &product)
				if err != nil {
					return stockData
				}
				stockData.ProductName = product.ProductAttributes.ProductName

				for _, value := range monitorResponse.Payload.Images.(map[string]interface{}) {
					if stockData.ImageURL != "" {
						break
					}
					image := Image{}
					tempJson, err := json.Marshal(value.(map[string]interface{}))
					if err != nil {
						return stockData
					}

					err = json.Unmarshal(tempJson, &image)
					if err != nil {
						return stockData
					}

					stockData.ImageURL = image.AssetSizeUrls.Default
				}

			}

		}
	case 404:
		fmt.Printf("We have a bad response: %v\n", resp.Status)
	default:
		fmt.Printf("Unkown Code: %v\n", resp.StatusCode)
	}

	return stockData
}
