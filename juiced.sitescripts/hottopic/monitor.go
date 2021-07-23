package hottopic

import (
	"fmt"
	"math/rand"
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
	browser "github.com/eddycjy/fake-useragent"
)

// CreateHottopicMonitor takes a TaskGroup entity and turns it into a Hottopic Monitor
func CreateHottopicMonitor(taskGroup *entities.TaskGroup, proxies []entities.Proxy, eventBus *events.EventBus, singleMonitors []entities.HottopicSingleMonitorInfo) (Monitor, error) {
	storedHottopicMonitors := make(map[string]entities.HottopicSingleMonitorInfo)
	hottopicMonitor := Monitor{}

	pids := []string{}
	for _, monitor := range singleMonitors {
		storedHottopicMonitors[monitor.Pid] = monitor
		pids = append(pids, monitor.Pid)
	}

	hottopicMonitor = Monitor{
		Monitor: base.Monitor{
			TaskGroup: taskGroup,
			Proxies:   proxies,
			EventBus:  eventBus,
		},
		Pids:        pids,
		PidWithInfo: storedHottopicMonitors,
	}

	return hottopicMonitor, nil
}

// PublishEvent wraps the EventBus's PublishMonitorEvent function
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
	}

	wg := sync.WaitGroup{}
	wg.Add(len(monitor.Pids))
	for _, pid := range monitor.Pids {
		go func(x string) {
			monitor.RunSingleMonitor(x)
			wg.Done()
		}(pid)
	}
	wg.Wait()
}

func (monitor *Monitor) RunSingleMonitor(pid string) {
	needToStop := monitor.CheckForStop()
	if needToStop {
		return
	}

	var sizes []string
	var colors []string
	var stockData HotTopicInStockData
	var err error

	if len(monitor.Monitor.Proxies) > 0 {
		client.UpdateProxy(&monitor.Monitor.Client, common.ProxyCleaner(monitor.Monitor.Proxies[rand.Intn(len(monitor.Monitor.Proxies))]))
	}

	// Case 1: User provides Random size, Random color
	//		Case 1A: Product has size and color
	// 		Case 1B: Product has size, no colors
	// 		Case 1C: Product has color, no sizes
	// 		Case 1D: Product has neither color nor size
	// Case 2: User provides color, Random size
	// 		Case 2A: Product has size and color
	// 		Case 2B: Product has size, no colors
	// 		Case 2C: Product has color, no sizes
	// 		Case 2D: Product has neither color nor size
	// Case 3: User provides size, Random color
	// 		Case 3A: Product has size and color
	// 		Case 3B: Product has size, no colors
	// 		Case 3C: Product has color, no sizes
	// 		Case 3D: Product has neither color nor size
	// Case 4: User provides size & color
	// 		Case 4A: Product has size and color
	// 		Case 4B: Product has size, no colors
	// 		Case 4C: Product has color, no sizes
	// 		Case 4D: Product has neither color nor size

	// Size and color: https://www.hottopic.com/on/demandware.store/Sites-hottopic-Site/default/Product-Variation?pid=16304343&Quantity=1&format=ajax
	// Size, no colors: https://www.boxlunch.com/on/demandware.store/Sites-boxlunch-Site/default/Product-Variation?pid=13941979&Quantity=1&format=ajax
	// Color, no sizes:
	// Neither color nor size: https://www.boxlunch.com/on/demandware.store/Sites-boxlunch-Site/default/Product-Variation?pid=15647180&Quantity=1&format=ajax

	// 1. First, make a request to `endpoint` and check if In Stock
	// 		If In Stock, takes care of Case 1D, 2D, 3D, 4D --> Success
	// 2. If not In Stock, check if there is a list of color variations and a list of size variations
	//		If no list of color variations, input with ID `productColor` should have the default color selected
	//		If no list of size variations, the base PID is the correct PID
	//		Build a list of color variations and size variations (PIDs)
	//			Color variations are just the name of the color
	//			Size variations are incremented PIDs from the base PID, in order from smallest to largest
	// 3. Take care of remaining cases:
	//		Case 1A: Pick a random size and color from the lists
	//		Case 1B: Pick a random size from the size list and choose the default color
	//		Case 1C: Choose the default size and pick a random color from the color list
	//		Case 2A: Pick a random size from the size list and ensure the color matches the provided color
	//		Case 2B: Pick a random size from the size list and choose the default color (even if it doesn't match)
	//		Case 2C: Choose the default size and ensure the color matches the provided color
	//		Case 3A: Ensure the size matches the provided size and pick a random color from the color list
	//		Case 3B: Ensure the size matches the provided size and choose the default color
	//		Case 3C: Choose the default size (even if it doesn't match) and pick a random color from the color list
	//		Case 4A: Ensure the size matches the provided size and the color matches the provided color
	//		Case 4B: Ensure the size matches the provided size and choose the default color (even if it doesn't match)
	//		Case 4C: Choose the default size (even if it doesn't match) and ensure the color matches the provided color
	// 4. Use the chosen size and color to build an endpoint that includes these values and finds out if it's In Stock

	switch monitor.PidWithInfo[pid].MonitorType {
	case enums.SKUMonitor:
		sizes, colors, stockData, err = monitor.GetSizeAndColor(pid)
		if err == nil && stockData.PID == "" {
			if len(sizes) > 0 && len(colors) > 0 {

			}
		}
	}

	needToStop = monitor.CheckForStop()
	if needToStop {
		return
	}

	if stockData.PID != "" && stockData.InPriceRange {
		var inSlice bool
		for _, monitorStock := range monitor.InStock {
			inSlice = monitorStock.PID == stockData.PID
		}
		if !inSlice {
			monitor.RunningMonitors = common.RemoveFromSlice(monitor.RunningMonitors, pid)
			monitor.InStock = append(monitor.InStock, stockData)
			monitor.PublishEvent(enums.SendingProductInfoToTasks, enums.MonitorUpdate, events.ProductInfo{
				Products: []events.Product{
					{ProductName: stockData.ProductName, ProductImageURL: stockData.ImageURL}},
			})
		}
	} else {
		if len(monitor.RunningMonitors) > 0 {
			if monitor.Monitor.TaskGroup.MonitorStatus != enums.WaitingForInStock &&
				monitor.Monitor.TaskGroup.MonitorStatus != enums.UnableToFindProduct &&
				monitor.Monitor.TaskGroup.MonitorStatus != enums.OutOfPriceRange {
				if stockData.InPriceRange {
					monitor.PublishEvent(enums.WaitingForInStock, enums.MonitorUpdate, events.ProductInfo{
						Products: []events.Product{
							{ProductName: stockData.ProductName, ProductImageURL: stockData.ImageURL}},
					})
				} else {
					monitor.PublishEvent(enums.OutOfPriceRange, enums.MonitorUpdate, events.ProductInfo{
						Products: []events.Product{
							{ProductName: stockData.ProductName, ProductImageURL: stockData.ImageURL}},
					})
				}
			}
		}
		for i, monitorStock := range monitor.InStock {
			if monitorStock.PID == stockData.PID {
				monitor.InStock = append(monitor.InStock[:i], monitor.InStock[i+1:]...)
				break
			}
		}

		time.Sleep(time.Duration(monitor.Monitor.TaskGroup.MonitorDelay) * time.Millisecond)
		monitor.RunSingleMonitor(pid)
	}
}

func (monitor *Monitor) GetSizeAndColor(pid string) ([]string, []string, HotTopicInStockData, error) {
	var sizes []string
	var colors []string
	var stockData HotTopicInStockData
	endpoint := fmt.Sprintf(MonitorEndpoint, pid)

	resp, body, err := util.MakeRequest(&util.Request{
		Client: monitor.Monitor.Client,
		Method: "GET",
		URL:    endpoint,
		RawHeaders: [][2]string{
			{"sec-ch-ua", `" Not A;Brand";v="99", "Chromium";v="90", "Google Chrome";v="90"`},
			{"sec-ch-ua-mobile", "?0"},
			{"upgrade-insecure-requests", "1"},
			{"user-agent", browser.Chrome()},
			{"accept", "text/html,application/xhtml+xml,application/xml;q=0.9,image/avif,image/webp,image/apng,*/*;q=0.8,application/signed-exchange;v=b3;q=0.9"},
			{"sec-fetch-site", "none"},
			{"sec-fetch-mode", "navigate"},
			{"sec-fetch-user", "?1"},
			{"sec-fetch-dest", "document"},
			{"accept-encoding", "gzip, deflate, br"},
			{"accept-language", "en-US,en;q=0.9"},
		},
	})
	if err != nil {
		fmt.Println(err)
		return sizes, colors, stockData, err
	}

	switch resp.StatusCode {
	case 200:
		monitor.RunningMonitors = append(monitor.RunningMonitors, pid)
		return monitor.VariationInfo(body, pid)
	case 404:
		monitor.PublishEvent(enums.UnableToFindProduct, enums.MonitorUpdate, nil)
		return sizes, colors, stockData, nil
	default:
		fmt.Printf("Unknown Code:%v", resp.StatusCode)
		return sizes, colors, stockData, nil
	}
}

func (monitor *Monitor) VariationInfo(body, pid string) ([]string, []string, HotTopicInStockData, error) {
	var sizes []string
	var colors []string
	var stockData HotTopicInStockData

	doc := soup.HTMLParse(body)

	hasVariations := false
	inStock := false
	inStockText := doc.Find("span", "class", "color-green")
	if inStockText.Error == nil {
		if inStockText.Text() == "In Stock" {
			inStock = true
		}
	} else {
		backorderText := doc.Find("span", "class", "text-red")
		if backorderText.Error == nil {
			if backorderText.Text() == "Backorder" {
				inStock = true
			}
		} else {
			hasVariations = true
		}
	}

	if !hasVariations && inStock {
		priceText := doc.Find("span", "class", "productdetail__info-pricing-sale")
		if priceText.Error != nil {
			priceText = doc.Find("span", "class", "productdetail__info-pricing-original")
		}
		if priceText.Error != nil {
			return sizes, colors, stockData, priceText.Error
		}
		priceStr := strings.ReplaceAll(strings.ReplaceAll(priceText.Text(), " ", ""), "$", "")
		price, err := strconv.ParseFloat(priceStr, 64)
		if err != nil {
			return sizes, colors, stockData, err
		}
		defaultColorInput := doc.Find("input", "id", "productColor")
		if defaultColorInput.Error != nil {
			return sizes, colors, stockData, defaultColorInput.Error
		}
		defaultColor := defaultColorInput.Attrs()["value"]
		productNameHeader := doc.Find("h1", "class", "productdetail__info-title")
		if productNameHeader.Error != nil {
			return sizes, colors, stockData, productNameHeader.Error
		}
		productName := productNameHeader.Text()
		productImage := doc.Find("img", "class", "productdetail__image-active-each")
		if productImage.Error != nil {
			return sizes, colors, stockData, productImage.Error
		}
		imageURL := productImage.Attrs()["src"]
		stockData = HotTopicInStockData{
			PID:          pid,
			SizePID:      pid,
			Color:        defaultColor,
			ProductName:  productName,
			ImageURL:     imageURL,
			Price:        int(price),
			InPriceRange: monitor.PidWithInfo[pid].MaxPrice == -1 || monitor.PidWithInfo[pid].MaxPrice >= int(price),
		}
	} else {

	}
	return sizes, colors, stockData, nil
}

// func (monitor *Monitor) GetStock(pid string) HotTopicInStockData {
// 	stockData := HotTopicInStockData{}
// 	endpoint := fmt.Sprintf(MonitorEndpoint, pid)

// 	resp, body, err := util.MakeRequest(&util.Request{
// 		Client: monitor.Monitor.Client,
// 		Method: "GET",
// 		URL:    endpoint,
// 		RawHeaders: [][2]string{
// 			{"sec-ch-ua", `" Not A;Brand";v="99", "Chromium";v="90", "Google Chrome";v="90"`},
// 			{"sec-ch-ua-mobile", "?0"},
// 			{"upgrade-insecure-requests", "1"},
// 			{"user-agent", browser.Chrome()},
// 			{"accept", "text/html,application/xhtml+xml,application/xml;q=0.9,image/avif,image/webp,image/apng,*/*;q=0.8,application/signed-exchange;v=b3;q=0.9"},
// 			{"sec-fetch-site", "none"},
// 			{"sec-fetch-mode", "navigate"},
// 			{"sec-fetch-user", "?1"},
// 			{"sec-fetch-dest", "document"},
// 			{"accept-encoding", "gzip, deflate, br"},
// 			{"accept-language", "en-US,en;q=0.9"},
// 		},
// 	})
// 	if err != nil {
// 		fmt.Println(err)
// 		return stockData
// 	}

// 	switch resp.StatusCode {
// 	case 200:
// 		monitor.RunningMonitors = append(monitor.RunningMonitors, pid)
// 		return monitor.StockInfo(body, pid)
// 	case 404:
// 		monitor.PublishEvent(enums.UnableToFindProduct, enums.MonitorUpdate, nil)
// 		return stockData
// 	default:
// 		fmt.Printf("Unknown Code:%v", resp.StatusCode)
// 		return stockData
// 	}
// }

// func (monitor *Monitor) GetStockInfo(body string, pid string) HotTopicInStockData {
// 	stockData := HotTopicInStockData{}
// 	doc := soup.HTMLParse(body)

// 	ShipTable := doc.Find("div", "class", "method-descr__label")
// 	InStock := ShipTable.Find("span", "class", "color-green").Text() == "In stock"
// 	if !InStock {
// 		InStock = ShipTable.Find("span", "class", "text-red").Text() == "Backorder"
// 	}
// 	if !InStock {
// 		//not instock or backorder return false
// 		return stockData
// 	}

// 	//We are in stock for this size/color, lets check price is in budget.
// 	PriceText := doc.Find("span", "class", "productdetail__info-pricing-original").Text()
// 	Price, _ := strconv.Atoi(PriceText)
// 	InBudget := monitor.PidWithInfo[pid].MaxPrice >= Price || monitor.PidWithInfo[pid].MaxPrice == -1

// 	ProductName := doc.Find("a", "class", "name-link").Text()
// 	if !InBudget {
// 		//not in budget return false
// 		return stockData
// 	}

// 	//we are in stock and in budget
// 	return HotTopicInStockData{
// 		PID:         pid,
// 		Size:        monitor.PidWithInfo[pid].Size,
// 		Color:       monitor.PidWithInfo[pid].Color,
// 		ProductName: ProductName,
// 		ImageURL:    "https://hottopic.scene7.com/is/image/HotTopic/" + pid + "_hi",
// 	}
// }
