package walmart

import (
	"fmt"
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
func CreateWalmartMonitor(taskGroup *entities.TaskGroup, proxy entities.Proxy, eventBus *events.EventBus, monitorType enums.MonitorType, products []entities.WalmartProduct) (Monitor, error) {
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
		Products:    products,
	}
	return walmartMonitor, err
}

// PublishEvent wraps the EventBus's PublishMonitorEvent function
func (monitor *Monitor) PublishEvent(status enums.MonitorStatus, eventType enums.MonitorEventType) {
	monitor.Monitor.TaskGroup.SetMonitorStatus(status)
	monitor.Monitor.EventBus.PublishMonitorEvent(status, eventType, nil, monitor.Monitor.TaskGroup.GroupID)
}

func (monitor *Monitor) CheckForStop() bool {
	if monitor.Monitor.StopFlag {
		monitor.PublishEvent(enums.MonitorIdle, enums.MonitorStop)
		return true
	}
	return false
}

func (monitor *Monitor) RunMonitor() {
	// If the function panics due to a runtime error, recover from it
	defer func() {
		recover()
		// TODO @silent: Let the UI know that a monitor failed
	}()

	if monitor.Monitor.TaskGroup.MonitorStatus == enums.MonitorIdle {
		monitor.PublishEvent(enums.WaitingForProductData, enums.MonitorStart)
	}

	needToStop := monitor.CheckForStop()
	if needToStop {
		return
	}
	inStockForShip := []events.WalmartSingleStockData{}
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
		monitor.SendToTasks(inStockForShip)
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

func (monitor *Monitor) GetSkuStock() ([]events.WalmartSingleStockData, []string) {
	inStockForShip := make([]events.WalmartSingleStockData, 0)
	outOfStockForShip := make([]string, 0)

	var products []entities.WalmartProduct
	var skus []string
	for _, product := range monitor.Products {
		products = append(products, product)
		skus = append(skus, product.Sku)
	}

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
			fmt.Println("We are on the captcha page.")
			//captcha
		} else if strings.Contains(resp.Request.URL.String(), "cart") {
			fmt.Println("All requested items are in-stock.")
			inStockForShip = ConvertProductsToWalmartSingleStock(products)
		} else {
			responseBody := soup.HTMLParse(string(body))
			if !UrlExistsInResponse(responseBody) {
				fmt.Println("All requested items are out of stock.")
				outOfStockForShip = skus
			} else {
				foundItems := ParseInstockSku(responseBody)
				//remove from products where it wasnt found. We do this as we need to keep our maxPrice with it.
				for i, v := range products {
					if !findValueInList(v.Sku, foundItems) {
						//not in stock remove from our list
						products = append(products[:i], products[i+1:]...)
					} else {
						if v.MaxPrice < 99999 {
							price := GetPrice(v.Sku)
							if price > v.MaxPrice {
								//too expensive remove from our list
								products = append(products[:i], products[i+1:]...)
							}
						}
					}
				}
			}

			inStockForShip = ConvertProductsToWalmartSingleStock(products)
			fmt.Print(inStockForShip)
			fmt.Println(" items are in-stock")
		}
	case 404:
		fmt.Printf("We have a bad response:%v", resp.Status)
	default:
		fmt.Printf("Unkown Code:%v", resp.StatusCode)
	}

	return inStockForShip, outOfStockForShip
}

func GetPrice(Sku string) int {
	return 1
}

//takes a lists of needles and searches a list of haystacks. Filters out missing needles and returns list.
func ListUniqueFilter(needles []string, haystack []string) []string {
	for _, needle := range needles {
		if !findValueInList(needle, haystack) {
			for i, v := range needles {
				if v == needle {
					needles = append(needles[:i], needles[i+1:]...)
					break
				}
			}
		}
	}

	return needles
}

func findValueInList(needle string, haystack []string) bool {
	found := false
	for _, i := range haystack {
		if i == needle {
			found = true
			break
		}
	}
	return found
}
func (monitor *Monitor) SendToTasks(inStockForShip []events.WalmartSingleStockData) {
	data := events.WalmartStockData{InStockForShip: inStockForShip}
	monitor.Monitor.EventBus.PublishProductEvent(enums.Walmart, data, monitor.Monitor.TaskGroup.GroupID)
}
