package shopify

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	cmap "github.com/orcaman/concurrent-map"

	"backend.juicedbot.io/juiced.client/http"
	"backend.juicedbot.io/juiced.infrastructure/common/entities"
	"backend.juicedbot.io/juiced.infrastructure/common/enums"
	"backend.juicedbot.io/juiced.infrastructure/common/events"
	"backend.juicedbot.io/juiced.sitescripts/base"
	"backend.juicedbot.io/juiced.sitescripts/util"
)

// CreateShopifyMonitor takes a TaskGroup entity and turns it into a Shopify Monitor
func CreateShopifyMonitor(taskGroup *entities.TaskGroup, proxyGroup *entities.ProxyGroup, eventBus *events.EventBus, siteURL, sitePassword string, singleMonitors []entities.ShopifySingleMonitorInfo) (Monitor, error) {
	storedShopifyMonitors := make(map[string]entities.ShopifySingleMonitorInfo)
	shopifyMonitor := Monitor{}
	keywords := []string{}

	for _, monitor := range singleMonitors {
		storedShopifyMonitors[monitor.Keyword] = monitor
		keywords = append(keywords, monitor.Keyword)
	}

	shopifyMonitor = Monitor{
		Monitor: base.Monitor{
			TaskGroup:  taskGroup,
			ProxyGroup: proxyGroup,
			EventBus:   eventBus,
		},
		SiteURL:         siteURL,
		SitePassword:    sitePassword,
		Keywords:        keywords,
		KeywordWithInfo: storedShopifyMonitors,
		InStock:         cmap.New(),
	}

	return shopifyMonitor, nil
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
		err := monitor.Monitor.CreateClient()
		if err != nil {
			return
		}

		var proxy *entities.Proxy
		if monitor.Monitor.ProxyGroup != nil {
			if len(monitor.Monitor.ProxyGroup.Proxies) > 0 {
				proxy = util.RandomLeastUsedProxy(monitor.Monitor.ProxyGroup.Proxies)
				monitor.Monitor.UpdateProxy(proxy)
			}
		}

		becameGuest := false
		for !becameGuest {
			needToStop := monitor.CheckForStop()
			if needToStop {
				return
			}
			becameGuest = BecomeGuest(monitor.Monitor.Client, monitor.SiteURL, monitor.SitePassword)
			if !becameGuest {
				time.Sleep(1000 * time.Millisecond)
			}
		}

	}

	needToStop := monitor.CheckForStop()
	if needToStop {
		return
	}

	monitor.RunSingleMonitor()

}

func (monitor *Monitor) RunSingleMonitor() {
	defer func() {
		recover()
		// TODO @silent: Re-run this specific monitor
	}()

	needToStop := monitor.CheckForStop()
	if needToStop {
		return
	}

	var proxy *entities.Proxy
	if monitor.Monitor.ProxyGroup != nil {
		if len(monitor.Monitor.ProxyGroup.Proxies) > 0 {
			proxy = util.RandomLeastUsedProxy(monitor.Monitor.ProxyGroup.Proxies)
			monitor.Monitor.UpdateProxy(proxy)
		}
	}

	stockData := monitor.GetSiteStock()
	if len(stockData.InStock) > 0 {
		needToStop := monitor.CheckForStop()
		if needToStop {
			return
		}

		var productInfo events.ProductInfo

		for _, singleStockData := range stockData.InStock {
			monitor.InStock.Set(singleStockData.VariantID, singleStockData)
			productInfo.Products = append(productInfo.Products, events.Product{
				ProductName:     singleStockData.ItemName,
				ProductImageURL: singleStockData.ImageURL,
			})
		}

		monitor.PublishEvent(enums.SendingProductInfoToTasks, enums.MonitorUpdate, productInfo)

	} else {
		if len(stockData.OutOfStock) > 0 {
			if monitor.Monitor.TaskGroup.MonitorStatus != enums.WaitingForInStock {
				monitor.PublishEvent(enums.WaitingForInStock, enums.MonitorUpdate, nil)
			}
		}

		time.Sleep(time.Duration(monitor.Monitor.TaskGroup.MonitorDelay) * time.Millisecond)
		monitor.RunSingleMonitor()
	}

}

// Getting stock by adding to cart
func (monitor *Monitor) GetSiteStock() ShopifyStockData {
	stockData := ShopifyStockData{}

	page := 1
	products := []Products{}
	morePages := true
	for morePages {
		monitorResponse := MonitorResponse{}
		resp, _, err := util.MakeRequest(&util.Request{
			Client: monitor.Monitor.Client,
			Method: "GET",
			URL:    monitor.SiteURL + MonitorEndpoint + "?limit=250&page=" + fmt.Sprint(page),
			RawHeaders: http.RawHeader{
				{"sec-ch-ua", `" Not;A Brand";v="99", "Google Chrome";v="91", "Chromium";v="91"`},
				{"accept", "application/json, text/javascript, */*; q=0.01"},
				{"x-requested-with", "XMLHttpRequest"},
				{"sec-ch-ua-mobile", "?0"},
				{"user-agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/91.0.4472.106 Safari/537.36"},
				{"origin", monitor.SiteURL},
				{"sec-fetch-site", "same-origin"},
				{"sec-fetch-site", "cors"},
				{"sec-fetch-dest", "empty"},
				{"referer", monitor.SiteURL + "/"},
				{"accept-encoding", "gzip, deflate"},
				{"accept-language", "en-US,en;q=0.9"},
			},
			ResponseBodyStruct: &monitorResponse,
		})
		if err != nil {
			return stockData
		}

		switch resp.StatusCode {
		case 200:
			if len(monitorResponse.Products) == 0 {
				morePages = true
				break
			}
			products = append(products, monitorResponse.Products...)
			page++
			time.Sleep(250 * time.Millisecond)
		default:
			time.Sleep(time.Duration(monitor.Monitor.TaskGroup.MonitorDelay))
		}

	}
	for _, product := range products {
		var wanted bool
		var currentKeyword string
		for _, keyword := range monitor.Keywords {
			wanted = strings.Contains(strings.ToLower(product.Title), keyword)
			if wanted {
				currentKeyword = keyword
				break
			}
		}
		if wanted {
			for _, variant := range product.Variants {
				switch variant.Available {
				case true:
					if ok := monitor.InStock.Has(fmt.Sprint(variant.ID)); !ok {
						price, err := strconv.ParseFloat(variant.Price, 64)
						if err != nil {
							return stockData
						}
						if float64(monitor.KeywordWithInfo[currentKeyword].MaxPrice) > price {
							title := variant.Title
							if title == "Default Title" {
								title = product.Title
							}
							var imageURL string
							for _, image := range product.Images {
								imageURL = image.Src
							}
							stockData.InStock = append(stockData.InStock, SingleStockData{
								VariantID: fmt.Sprint(variant.ID),
								Price:     price,
								ItemName:  title,
								ImageURL:  imageURL,
							})
						}
					}
				case false:
					monitor.InStock.Remove(fmt.Sprint(variant.ID))
					stockData.OutOfStock = append(stockData.OutOfStock, SingleStockData{VariantID: fmt.Sprint(variant.ID)})
				}
			}
		}

	}

	return stockData
}
