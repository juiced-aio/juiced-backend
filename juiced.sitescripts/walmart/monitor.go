package walmart

import (
	"fmt"
	"log"
	"regexp"
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
	inStockForShip := []WalmartInStockData{}
	outOfStockForShip := make([]string, 0)

	switch monitor.MonitorType {
	case enums.SKUMonitor:
		inStockForShip, outOfStockForShip = monitor.GetSkuStock()
	}

	somethingInStock := false
	if len(inStockForShip) > 0 {
		somethingInStock = true
	}

	if somethingInStock {
		needToStop := monitor.CheckForStop()
		if needToStop {
			return
		}
		monitor.PublishEvent(enums.SendingProductInfoToTasks, enums.MonitorUpdate)
		monitor.InStockForShip = inStockForShip
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
		}
	}
}

//This is for checking if a list of Skus are instock. Here we also check if there is a maximum price.
func (monitor *Monitor) GetSkuStock() ([]WalmartInStockData, []string) {
	inStockForShip := make([]WalmartInStockData, 0)
	outOfStockForShip := make([]string, 0)
	var skus []string

	resp, body, err := util.MakeRequest(&util.Request{
		Client: monitor.Monitor.Client,
		Method: "GET",
		URL:    MonitorEndpoint + strings.Join(skus, ","),
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
			err := SetPXCapCookie(resp.Request.URL.String(), &monitor.PXValues, monitor.Monitor.Proxy, &monitor.Monitor.Client)
			if err != nil {
				fmt.Println(err.Error())
			}
			fmt.Println("Cookie updated.")
		} else if strings.Contains(resp.Request.URL.String(), "cart") {
			fmt.Println("All requested items are in-stock.")
			for _, thisStock := range ConvertSkuListToWalmartSingleStock(skus) {
				var inSlice bool
				for _, monitorStock := range monitor.InStockForShip {
					if monitorStock.Sku == thisStock.Sku {
						inSlice = true
					}
				}
				if !inSlice {
					inStockForShip = append(inStockForShip, thisStock)
				}
			}
		} else {
			responseBody := soup.HTMLParse(string(body))
			if !UrlExistsInResponse(responseBody) {
				for _, sku := range skus {
					for i, monitorStock := range monitor.InStockForShip {
						if monitorStock.Sku == sku {
							monitor.InStockForShip = append(monitor.InStockForShip[:i], monitor.InStockForShip[i+1:]...)
							break
						}
					}
				}
				outOfStockForShip = skus
			} else {
				foundItems := ParseInstockSku(responseBody)
				if monitor.Monitor.TaskGroup.WalmartMonitorInfo.MaxPrice > -1 {
					for i, sku := range foundItems {
						price := monitor.GetPrice(sku)
						if price > monitor.Monitor.TaskGroup.WalmartMonitorInfo.MaxPrice || monitor.Monitor.TaskGroup.WalmartMonitorInfo.MaxPrice == -1 {
							foundItems = append(foundItems[:i], foundItems[i+1:]...)
						}
					}
				}
				for _, thisStock := range ConvertSkuListToWalmartSingleStock(foundItems) {
					var inSlice bool
					for _, monitorStock := range monitor.InStockForShip {
						inSlice = monitorStock.Sku == thisStock.Sku
					}
					if !inSlice {
						inStockForShip = append(inStockForShip, thisStock)
					}
				}
				fmt.Print(inStockForShip)
				fmt.Println(" items are in-stock")
			}
		}
	case 404:
		fmt.Printf("We have a bad response:%v", resp.Status)
	default:
		fmt.Printf("Unkown Code:%v", resp.StatusCode)
	}

	return inStockForShip, outOfStockForShip
}

//This is for checking the maximum price. It is called from GetSkuStock() if maximum price is over -1
func (monitor *Monitor) GetPrice(Sku string) int {
	var price = 0

	resp, body, err := util.MakeRequest(&util.Request{
		Client: monitor.Monitor.Client,
		Method: "GET",
		URL:    PriceMonitorEndpoint + Sku,
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
			err := SetPXCapCookie(resp.Request.URL.String(), &monitor.PXValues, monitor.Monitor.Proxy, &monitor.Monitor.Client)
			if err != nil {
				fmt.Println(err.Error())
			}
			fmt.Println("Cookie updated.")
		} else if strings.Contains(resp.Request.URL.String(), "walmart.com/ip/seort") {
			fmt.Println("Invalid Sku")
		} else {
			reg, _ := regexp.Compile("[^0-9]+")
			responseBody := soup.HTMLParse(string(body))
			priceBlock := responseBody.Find("span", "class", "display-inline-block")
			priceText := priceBlock.Find("span", "class", "visuallyhidden").Text()
			price, err = strconv.Atoi(reg.ReplaceAllString(priceText, ""))
			if err != nil {
				log.Fatal(err)
			}
		}
	case 404:
		fmt.Printf("We have a bad response:%v", resp.Status)
	default:
		fmt.Printf("Unkown Code:%v", resp.StatusCode)
	}
	return price
}
