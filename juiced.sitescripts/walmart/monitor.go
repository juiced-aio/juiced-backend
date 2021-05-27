package walmart

import (
	"fmt"
	"io/ioutil"
	"strings"
	"time"

	"backend.juicedbot.io/m/v2/juiced.infrastructure/common/entities"
	"backend.juicedbot.io/m/v2/juiced.infrastructure/common/enums"
	"backend.juicedbot.io/m/v2/juiced.infrastructure/common/events"
	"backend.juicedbot.io/m/v2/juiced.sitescripts/base"
	"backend.juicedbot.io/m/v2/juiced.sitescripts/util"

	"github.com/anaskhan96/soup"
)

// CreateWalmartMonitor takes a TaskGroup entity and turns it into a Walmart Monitor
func CreateWalmartMonitor(taskGroup *entities.TaskGroup, proxy entities.Proxy, eventBus *events.EventBus, monitorType enums.MonitorType, skus []string) (Monitor, error) {
	walmartMonitor := Monitor{}
	client, err := util.CreateClient()
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

func (monitor *Monitor) CheckForStop() bool {
	if monitor.Monitor.StopFlag {
		monitor.PublishEvent(enums.MonitorIdle, enums.MonitorStop)
		return true
	}
	return false
}

func (monitor *Monitor) RunMonitor() {
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

	resp, err := util.MakeRequest(&util.Request{
		Client: monitor.Monitor.Client,
		Method: "GET",
		URL:    MonitorEndpoint + strings.Join(monitor.SKUs, ","),
	})
	if err != nil {
		fmt.Println(err)
	}

	switch resp.StatusCode {
	case 200:
		if strings.Contains(resp.Request.URL.String(), "blocked") {
			fmt.Println("We are on the captcha page.")
			//captcha
		} else if strings.Contains(resp.Request.URL.String(), "cart") {
			fmt.Println("All requested items are in-stock.")
			inStockForShip = ConvertSkuListToWalmartSingleStock(monitor.SKUs)
		} else {
			body, _ := ioutil.ReadAll(resp.Body)
			responseBody := soup.HTMLParse(string(body))

			if !UrlExistsInResponse(responseBody) {
				fmt.Println("All requested items are out of stock.")
				outOfStockForShip = monitor.SKUs
			} else {
				inStockForShip = ConvertSkuListToWalmartSingleStock(ParseInstockSku(responseBody))
				fmt.Print(inStockForShip)
				fmt.Println(" items are in-stock")
			}
			return inStockForShip, outOfStockForShip
		}
	case 404:
		fmt.Printf("We have a bad response:%v", resp.Status)
	default:
		fmt.Printf("Unkown Code:%v", resp.StatusCode)
	}

	return inStockForShip, outOfStockForShip
}

func (monitor *Monitor) SendToTasks(inStockForShip []events.WalmartSingleStockData) {
	data := events.WalmartStockData{InStockForShip: inStockForShip}
	monitor.Monitor.EventBus.PublishProductEvent(enums.Walmart, data, monitor.Monitor.TaskGroup.GroupID)
}
