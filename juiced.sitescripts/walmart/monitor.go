package walmart

import (
	"encoding/json"
	"fmt"
	"log"
	"strings"
	"sync"
	"time"

	"backend.juicedbot.io/juiced.client/http"
	"backend.juicedbot.io/juiced.infrastructure/common"
	"backend.juicedbot.io/juiced.infrastructure/common/entities"
	"backend.juicedbot.io/juiced.infrastructure/common/enums"
	"backend.juicedbot.io/juiced.infrastructure/common/events"
	"backend.juicedbot.io/juiced.sitescripts/base"
	"backend.juicedbot.io/juiced.sitescripts/util"
)

// CreateWalmartMonitor takes a TaskGroup entity and turns it into a Walmart Monitor
func CreateWalmartMonitor(taskGroup *entities.TaskGroup, proxyGroup *entities.ProxyGroup, eventBus *events.EventBus, singleMonitors []entities.WalmartSingleMonitorInfo) (Monitor, error) {
	storedWalmartMonitors := make(map[string]entities.WalmartSingleMonitorInfo)
	walmartMonitor := Monitor{}
	ids := []string{}

	for _, monitor := range singleMonitors {
		storedWalmartMonitors[monitor.ID] = monitor
		ids = append(ids, monitor.ID)

	}

	walmartMonitor = Monitor{
		Monitor: base.Monitor{
			TaskGroup:  taskGroup,
			ProxyGroup: proxyGroup,
			EventBus:   eventBus,
		},
		IDs:        ids,
		IDWithInfo: storedWalmartMonitors,
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
		err := monitor.Monitor.CreateClient()
		if err != nil {
			return
		}

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
				time.Sleep(common.MS_TO_WAIT)
			}
		}
	}

	if monitor.Monitor.TaskGroup.MonitorStatus == enums.SettingUpMonitor {
		monitor.PublishEvent(enums.WaitingForProductData, enums.MonitorUpdate, nil)
	}

	needToStop := monitor.CheckForStop()
	if needToStop {
		return
	}

	wg := sync.WaitGroup{}
	wg.Add(len(monitor.IDs))
	for _, id := range monitor.IDs {
		go func(x string) {
			monitor.RunSingleMonitor(x)
			wg.Done()
		}(id)
	}
	wg.Wait()
}

func (monitor *Monitor) RunSingleMonitor(id string) {
	defer func() {
		if r := recover(); r != nil {
			log.Println(r)
		}
		// TODO @silent: Re-run this specific monitor
	}()

	needToStop := monitor.CheckForStop()
	if needToStop {
		return
	}

	var proxy *entities.Proxy
	if !common.InSlice(monitor.RunningMonitors, id) {
		if monitor.Monitor.ProxyGroup != nil {
			if len(monitor.Monitor.ProxyGroup.Proxies) > 0 {
				proxy = util.RandomLeastUsedProxy(monitor.Monitor.ProxyGroup.Proxies)
				monitor.Monitor.UpdateProxy(proxy)
			}
		}

		stockData := WalmartInStockData{}

		switch monitor.IDWithInfo[id].MonitorType {
		case enums.SKUMonitor:
			stockData = monitor.GetSkuStock(id)
		case enums.FastSKUMonitor:
			stockData = monitor.GetOfferIDStock(id)
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
				monitor.RunningMonitors = common.RemoveFromSlice(monitor.RunningMonitors, id)
				monitor.PublishEvent(enums.SendingProductInfoToTasks, enums.MonitorUpdate, events.ProductInfo{
					Products: []events.Product{
						{ProductName: stockData.ProductName, ProductImageURL: stockData.ImageURL}},
				})
			}
		} else {
			if len(monitor.RunningMonitors) > 0 {
				if monitor.Monitor.TaskGroup.MonitorStatus != enums.WaitingForInStock {
					monitor.PublishEvent(enums.WaitingForInStock, enums.MonitorUpdate, events.ProductInfo{
						Products: []events.Product{
							{ProductName: stockData.ProductName, ProductImageURL: stockData.ImageURL}},
					})
				}
			}
			for i, monitorStock := range monitor.InStockForShip {
				if monitorStock.SKU == stockData.SKU {
					monitor.InStockForShip = append(monitor.InStockForShip[:i], monitor.InStockForShip[i+1:]...)
					break
				}
			}

			time.Sleep(time.Duration(monitor.Monitor.TaskGroup.MonitorDelay) * time.Millisecond)
			monitor.RunningMonitors = common.RemoveFromSlice(monitor.RunningMonitors, id)
			monitor.RunSingleMonitor(id)
		}
	}
}

// RefreshPX3 refreshes the px3 cookie every 4 minutes since it expires every 5 minutes
func (monitor *Monitor) RefreshPX3() {
	quit := make(chan bool)
	defer func() {
		quit <- true
	}()

	cancellationToken := &util.CancellationToken{Cancel: false}
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
			time.Sleep(common.MS_TO_WAIT)
		}
	}()

	retry := true
	for retry {
		retry = monitor.RefreshPX3Helper(cancellationToken)
		time.Sleep(common.MS_TO_WAIT)
	}
}

func (monitor *Monitor) RefreshPX3Helper(cancellationToken *util.CancellationToken) bool {
	for {
		if cancellationToken.Cancel {
			return false
		}
		if monitor.PXValues.RefreshAt == 0 || time.Now().Unix() > monitor.PXValues.RefreshAt {
			pxValues, cancelled, err := SetPXCookie(monitor.Monitor.Proxy, &monitor.Monitor.Client, cancellationToken)
			if cancelled {
				return false
			}

			if err != nil {
				log.Println("Error setting px cookie for monitor: " + err.Error())
				return true
			}
			monitor.PXValues = pxValues
			monitor.PXValues.RefreshAt = time.Now().Unix() + 240
		}
		time.Sleep(common.MS_TO_WAIT)
	}
}

func (monitor Monitor) HandlePXCap(resp *http.Response, redirectURL string) bool {
	quit := make(chan bool)
	defer func() {
		quit <- true
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
			time.Sleep(common.MS_TO_WAIT)
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

func (monitor *Monitor) GetOfferIDStock(offerID string) WalmartInStockData {
	stockData := WalmartInStockData{}

	data := AddToCartRequest{
		OfferID:               offerID,
		Quantity:              1,
		ShipMethodDefaultRule: "SHIP_RULE_1",
	}
	dataStr, err := json.Marshal(data)
	if err != nil {
		log.Println("ATC Request Error: " + err.Error())
		return stockData
	}

	addToCartResponse := AddToCartResponse{}
	resp, _, err := util.MakeRequest(&util.Request{
		Client: monitor.Monitor.Client,
		Method: "POST",
		URL:    AddToCartEndpoint,
		RawHeaders: [][2]string{
			{"accept", "application/json"},
			{"accept-encoding", "gzip, deflate, br"},
			{"accept-language", "en-US,en;q=0.9"},
			{"content-type", "application/json"},
			{"sec-ch-ua", `" Not;A Brand";v="99", "Google Chrome";v="91", "Chromium";v="91"`},
			{"sec-ch-ua-mobile", "?0"},
			{"sec-fetch-dest", "document"},
			{"sec-fetch-mode", "navigate"},
			{"sec-fetch-site", "none"},
			{"sec-fetch-user", "?1"},
			{"upgrade-insecure-requests", "1"},
			{"user-agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/91.0.4472.77 Safari/537.36"},
			{"content-length", fmt.Sprint(len(dataStr))},
		},
		RequestBodyStruct:  data,
		ResponseBodyStruct: &addToCartResponse,
	})
	if err != nil || resp.StatusCode != 200 {
		fmt.Println(err.Error())
		return stockData
	}

	if strings.Contains(resp.Request.URL.String(), "blocked") {
		monitor.HandlePXCap(resp, resp.Request.URL.String())
		return stockData
	}

	var inStock bool
	for _, item := range addToCartResponse.Items {
		if item.OfferID == offerID {
			if item.AvailableQuantity > 0 {
				if monitor.IDWithInfo[offerID].MaxPrice >= int(item.Price) || monitor.IDWithInfo[offerID].MaxPrice == -1 {
					if monitor.IDWithInfo[offerID].SoldByWalmart {
						if item.Seller.ID == "F55CDC31AB754BB68FE0B39041159D63" {
							inStock = true
						}
					} else {
						inStock = true
					}

					if inStock {
						stockData.SKU = item.USItemID
						stockData.MaxQty = int(item.MaxItemCountPerOrder)
						stockData.OfferID = item.OfferID
						stockData.ProductName = item.Name
						stockData.Price = item.Price
						stockData.ImageURL = strings.ReplaceAll(item.Assets.Primary[0].Num100, "FFFFFF", "FFFFF")
					}
				}
			}

		}
	}

	return stockData
}

//This is for checking if a sku is instock. Here we also check if there is a maximum price.
func (monitor *Monitor) GetSkuStock(sku string) WalmartInStockData {
	stockData := WalmartInStockData{
		SKU: sku,
	}

	resp, body, err := util.MakeRequest(&util.Request{
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
	})
	if err != nil {
		fmt.Println(err.Error())
		return stockData
	}

	monitorResponse := MonitorResponse{}
	switch resp.StatusCode {
	case 200:
		if strings.Contains(resp.Request.URL.String(), "blocked") {
			monitor.HandlePXCap(resp, resp.Request.URL.String())
		} else {
			var inStock bool

			err = json.Unmarshal([]byte(body), &monitorResponse)
			if err != nil {
				return stockData
			}

			// Item exists = Add to running monitors
			if len(monitorResponse.Payload.Offers.(map[string]interface{})) > 0 {
				monitor.RunningMonitors = append(monitor.RunningMonitors, sku)
			}
			lowestPrice := -1
			for _, value := range monitorResponse.Payload.Offers.(map[string]interface{}) {
				offer := Offer{}
				tempJson, err := json.Marshal(value.(map[string]interface{}))
				if err == nil {
					err = json.Unmarshal(tempJson, &offer)
					if err == nil {
						if offer.Productavailability.Availabilitystatus == "IN_STOCK" {
							if monitor.IDWithInfo[sku].MaxPrice >= int(offer.Pricesinfo.Pricemap.Current.Price) || monitor.IDWithInfo[sku].MaxPrice == -1 {
								if monitor.IDWithInfo[sku].SoldByWalmart {
									if offer.Sellerid == "F55CDC31AB754BB68FE0B39041159D63" {
										inStock = true
									}
								} else {
									inStock = true
								}
							}
							if inStock && (lowestPrice == -1 || int(offer.Pricesinfo.Pricemap.Current.Price) <= lowestPrice) {
								stockData.OfferID = offer.OfferInfo.OfferID
								stockData.MaxQty = offer.OfferInfo.QuantityOptions.OrderLimit
								lowestPrice = int(offer.Pricesinfo.Pricemap.Current.Price)
							}
						}
					}
				}
			}

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

			secondaryImageUrl := ""
			secondaryImageRank := -1
			for _, value := range monitorResponse.Payload.Images.(map[string]interface{}) {
				if stockData.ImageURL != "" {
					break
				}
				image := Image{}
				tempJson, err := json.Marshal(value.(map[string]interface{}))
				if err == nil {
					err = json.Unmarshal(tempJson, &image)
					if err == nil {
						if image.Type == "PRIMARY" {
							stockData.ImageURL = image.AssetSizeUrls.Default
						} else if secondaryImageRank == -1 || image.Rank < secondaryImageRank {
							secondaryImageUrl = image.AssetSizeUrls.Default
							secondaryImageRank = image.Rank
						}
					}
				}
			}

			if stockData.ImageURL == "" && secondaryImageUrl != "" {
				stockData.ImageURL = secondaryImageUrl
			}
		}
	case 404:
		fmt.Printf("We have a bad response: %v\n", resp.Status)
	default:
		fmt.Printf("Unkown Code: %v\n", resp.StatusCode)
	}

	return stockData
}
