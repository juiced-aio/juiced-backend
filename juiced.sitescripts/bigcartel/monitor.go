package bigcartel

import (
	"math/rand"
	"net/url"
	"strconv"
	"strings"
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
)

func CreateBigCartelMonitor(taskGroup *entities.TaskGroup, proxies []entities.Proxy, eventBus *events.EventBus, siteURL string, singleMonitors []entities.BigCartelSingleMonitorInfo) (Monitor, error) {
	storedBigCartelMonitors := make(map[string]entities.BigCartelSingleMonitorInfo) //List of storedsingle monitors
	bigcartelMonitor := Monitor{}
	_skus := []string{}

	for _, monitor := range singleMonitors {
		storedBigCartelMonitors[monitor.Sku] = monitor //list of monitors
		_skus = append(_skus, monitor.Sku)             //gets list of skus from the passed in single monitors.
	}

	bigcartelMonitor = Monitor{
		Monitor: base.Monitor{
			TaskGroup: taskGroup,
			Proxies:   proxies,
			EventBus:  eventBus,
		},
		Skus:        _skus, //set the list of skus in monitor struct
		SKUWithInfo: storedBigCartelMonitors,
		SiteURL:     siteURL,
	}

	return bigcartelMonitor, nil
}

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
		monitor.Monitor.Client = monitorClient
	}

	needToStop := monitor.CheckForStop()
	if needToStop {
		return
	}

	wg := sync.WaitGroup{}
	wg.Add(len(monitor.Skus))
	for _, sku := range monitor.Skus { //iterate over the skus frrom monitor, run a single monitor for each
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

		stockData := monitor.GetStockWithSku(sku) //return instockData struct
		if stockData.Sku != "" {
			needToStop := monitor.CheckForStop()
			if needToStop {
				return
			}

			var inSlice bool
			for _, monitorStock := range monitor.InStock {
				if inSlice = monitorStock.Sku == stockData.Sku; inSlice {
					break
				}
			}
			if !inSlice {
				monitor.InStock = append(monitor.InStock, stockData)
				monitor.RunningMonitors = common.RemoveFromSlice(monitor.RunningMonitors, sku)
				monitor.PublishEvent(enums.SendingProductInfoToTasks, enums.MonitorUpdate, events.ProductInfo{ //Were not  getting this product info yet as its easier to grab it from a resposne later when setting up the payment info.
					Products: []events.Product{
						{ProductName: stockData.ItemName, ProductImageURL: stockData.ImageURL}},
				})
			}
		} else {
			if len(monitor.RunningMonitors) > 0 {
				if monitor.Monitor.TaskGroup.MonitorStatus != enums.WaitingForInStock {
					monitor.PublishEvent(enums.WaitingForInStock, enums.MonitorUpdate, nil)
				}
			}
			for i, monitorStock := range monitor.InStock {
				if monitorStock.Sku == stockData.Sku {
					monitor.InStock = append(monitor.InStock[:i], monitor.InStock[i+1:]...)
					break
				}
			}
			time.Sleep(time.Duration(monitor.Monitor.TaskGroup.MonitorDelay) * time.Millisecond)
			monitor.RunSingleMonitor(sku)
		}
	}
}

func (monitor *Monitor) GetStockWithSku(sku string) BigCartelInStockData {
	bigCartelInStockData := BigCartelInStockData{}

	payload := url.Values{
		"cart[add][id]": {sku},
		"submit":        {""},
	}

	addToCartResponse := AddToCartResponse{}
	resp, body, err := util.MakeRequest(&util.Request{
		Client:             monitor.Monitor.Client,
		Method:             "POST",
		URL:                monitor.SiteURL + "/cart",
		Data:               []byte(payload.Encode()),
		ResponseBodyStruct: &addToCartResponse,
	})
	if err != nil {
		//theres some error so just return this empty
		return bigCartelInStockData
	}

	switch resp.StatusCode {
	case 200:

		responseBody := soup.HTMLParse(string(body))
		price := ""
		skuData := responseBody.Find("div", "class", "remove")
		price = responseBody.Find("div", "class", "price").Text()
		if skuData.Pointer != nil && price != "" {
			link := skuData.Find("a")
			bigCartelInStockData.Sku = link.Pointer.Attr[1].Val
		}

		if bigCartelInStockData.Sku == sku {
			//product and price found
			responsePrice, err := strconv.Atoi(price)
			if err != nil {
				//Error converting price string -> int
				//unable to confirm price
				bigCartelInStockData = BigCartelInStockData{}
			} else {
				if responsePrice > monitor.SKUWithInfo[sku].MaxPrice {
					//too expensive
					bigCartelInStockData = BigCartelInStockData{}
				} else {
					//We have found the product and we have added to cart and its in budget
					//Now we do a GET request on the checkout to get the cart ID for passover
					storeId, cartToken, key := monitor.StoreAndCartid()
					if storeId != "" && cartToken != "" {
						bigCartelInStockData.StoreId = storeId
						bigCartelInStockData.CartToken = cartToken
						bigCartelInStockData.Key = key
					} else {
						//Unable to locate storeId and cartToken this is required
						bigCartelInStockData = BigCartelInStockData{}
					}
				}
			}

		} else {
			//out of stock/not in cart
			bigCartelInStockData = BigCartelInStockData{}
		}

	case 422:
		//Out Of Stock
		bigCartelInStockData = BigCartelInStockData{}
	case 404:
		//Item does not exist
		bigCartelInStockData = BigCartelInStockData{}
	default:
	}

	//we are always going to return the stockdata
	return bigCartelInStockData
}

func (monitor *Monitor) StoreAndCartid() (string, string, string) {
	resp, str, err := util.MakeRequest(&util.Request{
		Client: monitor.Monitor.Client,
		Method: "GET",
		URL:    GetStockEndpoint,
	})
	if err != nil {
		//error getting page
		return "", "", ""
	} else {
		responseBody := soup.HTMLParse(str)
		nextData := responseBody.Find("script", "type", "text/javascript").Text()
		out, _ := common.FindInString(nextData, "stripePublishableKey': \"", "\",")

		s := []string{}
		if strings.Contains(resp.Request.URL.String(), "checkout.bigcartel.com/") {
			s = strings.Split(string(resp.Request.URL.String()), "/")
			return s[3], s[4], out
		} else {
			//unable to get values from URL
			return "", "", ""
		}
	}

}
