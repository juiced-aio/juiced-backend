package disney

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"strconv"
	"strings"

	// "strings"
	"sync"
	"time"

	"backend.juicedbot.io/juiced.client/client"
	"backend.juicedbot.io/juiced.infrastructure/common"
	"backend.juicedbot.io/juiced.infrastructure/common/entities"
	"backend.juicedbot.io/juiced.infrastructure/common/enums"
	"backend.juicedbot.io/juiced.infrastructure/common/events"
	"backend.juicedbot.io/juiced.sitescripts/base"
	"backend.juicedbot.io/juiced.sitescripts/util"

	// "github.com/anaskhan96/soup"
	browser "github.com/eddycjy/fake-useragent"
)

// CreateDisneyMonitor takes a TaskGroup entity and turns it into a Disney Monitor
func CreateDisneyMonitor(taskGroup *entities.TaskGroup, proxies []entities.Proxy, eventBus *events.EventBus, singleMonitors []entities.DisneySingleMonitorInfo) (Monitor, error) {
	storedDisneyMonitors := make(map[string]entities.DisneySingleMonitorInfo)
	disneyMonitor := Monitor{}

	pids := []string{}
	for _, monitor := range singleMonitors {
		storedDisneyMonitors[monitor.PID] = monitor
		pids = append(pids, monitor.PID)
	}

	disneyMonitor = Monitor{
		Monitor: base.Monitor{
			TaskGroup: taskGroup,
			Proxies:   proxies,
			EventBus:  eventBus,
		},
		Pids:        pids,
		PidWithInfo: storedDisneyMonitors,
	}

	return disneyMonitor, nil
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
	var stockData DisneyInStockData
	var err error

	if len(monitor.Monitor.Proxies) > 0 {
		client.UpdateProxy(&monitor.Monitor.Client, common.ProxyCleaner(monitor.Monitor.Proxies[rand.Intn(len(monitor.Monitor.Proxies))]))
	}

	sizes, colors, stockData, err = monitor.GetSizeAndColor(pid)

	needToStop = monitor.CheckForStop()
	if needToStop {
		return
	}

	price := stockData.Price
	productName := stockData.ProductName
	itemURL := stockData.ItemURL
	imageURL := stockData.ImageURL
	quantityLimit := stockData.QuantityLimit
	isPreOrder := stockData.IsPreOrder
	isBackOrder := stockData.IsBackOrder
	outOfPriceRange := stockData.OutOfPriceRange

	if err == nil {
		// GetSizeAndColor will only return a populated stockData if the product has no size/color variations
		// 		stockData.ProductName and stockData.ImageURL will always be populated
		if stockData.PID == "" && !stockData.OutOfPriceRange {
			noColorsBeforeFilter := len(colors) == 0
			noSizesBeforeFilter := len(sizes) == 0

			// If stockData.PID == "", then the size and color lists should be populated
			// Filter the sizes we found with the ones the monitor has been provided
			sizesJoined := monitor.PidWithInfo[pid].Size
			if sizesJoined != "" {
				filteredSizes := []string{}
				for _, size := range sizes {
					if strings.Contains(strings.ToLower(sizesJoined), strings.ToLower(size)) {
						filteredSizes = append(filteredSizes, size)
					}
				}
				sizes = filteredSizes
			}
			// Filter the colors we found with the ones the monitor has been provided
			colorsJoined := monitor.PidWithInfo[pid].Color
			if colorsJoined != "" {
				filteredColors := []string{}
				for _, color := range colors {
					if strings.Contains(strings.ToLower(colorsJoined), strings.ToLower(color)) {
						filteredColors = append(filteredColors, color)
					}
				}
				colors = filteredColors
			}

			// If sizes and colors remain, continue
			if (len(sizes) > 0 && len(colors) > 0) ||
				(len(sizes) > 0 && noColorsBeforeFilter) || // (Or if there are sizes but no color variants)
				(len(colors) > 0 && noSizesBeforeFilter) { // (Or if there are colors but no size variants)
				needToStop = monitor.CheckForStop()
				if needToStop {
					return
				}

				// GetInStockVariations returns a list of DisneyStockData items for each size/color combination that's in stock
				stockDatas := monitor.GetInStockVariations(pid, sizes, colors)
				needToStop = monitor.CheckForStop()
				if needToStop {
					return
				}

				if len(stockDatas) > 0 {
					atLeastOneInPriceRange := false
					for _, stockData := range stockDatas { // Since we omitted these fields in the function below, add them back here
						stockData.Price = price
						stockData.ProductName = productName
						stockData.ItemURL = itemURL
						stockData.ImageURL = imageURL
						stockData.QuantityLimit = quantityLimit
						stockData.IsPreOrder = isPreOrder
						stockData.IsBackOrder = isBackOrder
						stockData.OutOfPriceRange = outOfPriceRange
						// Add each in stock combination to the monitor's InStock list, then update the status
						monitor.InStock = append(monitor.InStock, stockData)
						atLeastOneInPriceRange = true
					}
					if atLeastOneInPriceRange {
						// If at least one combination is in stock and in our price range, remove this monitor from the running monitors
						monitor.RunningMonitors = common.RemoveFromSlice(monitor.RunningMonitors, pid)
						monitor.PublishEvent(enums.SendingProductInfoToTasks, enums.MonitorUpdate, events.ProductInfo{
							Products: []events.Product{
								{ProductName: productName, ProductImageURL: imageURL}},
						})
					} else {
						// Otherwise, let the frontend know that none of the in stock combinations are in our price range
						monitor.PublishEvent(enums.OutOfPriceRange, enums.MonitorUpdate, events.ProductInfo{
							Products: []events.Product{
								{ProductName: productName, ProductImageURL: imageURL}},
						})
					}
				} else {
					// None of the filtered sizes/colors are in stock
					if stockData.ProductName != "" && stockData.ImageURL != "" {
						monitor.PublishEvent(enums.WaitingForInStock, enums.MonitorUpdate, events.ProductInfo{
							Products: []events.Product{
								{ProductName: productName, ProductImageURL: imageURL}},
						})
					}

					time.Sleep(time.Duration(monitor.Monitor.TaskGroup.MonitorDelay) * time.Millisecond)
					monitor.RunSingleMonitor(pid)
				}
			} else {
				// None of the available sizes/colors match the task's size/color filters
				if stockData.ProductName != "" && stockData.ImageURL != "" {
					monitor.PublishEvent(enums.WaitingForInStock, enums.MonitorUpdate, events.ProductInfo{
						Products: []events.Product{
							{ProductName: productName, ProductImageURL: imageURL}},
					})
				}

				time.Sleep(time.Duration(monitor.Monitor.TaskGroup.MonitorDelay) * time.Millisecond)
				monitor.RunSingleMonitor(pid)
			}
		} else {
			// This code is only run for items that have no size/color variations
			if stockData.PID != "" && !stockData.OutOfPriceRange {
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
						if !stockData.OutOfPriceRange {
							if stockData.ProductName != "" && stockData.ImageURL != "" {
								monitor.PublishEvent(enums.WaitingForInStock, enums.MonitorUpdate, events.ProductInfo{
									Products: []events.Product{
										{ProductName: stockData.ProductName, ProductImageURL: stockData.ImageURL}},
								})
							}
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
	} else {
		monitor.RunSingleMonitor(pid)
	}
}

func (monitor *Monitor) GetSizeAndColor(pid string) ([]string, []string, DisneyInStockData, error) {
	var sizes []string
	var colors []string
	var stockData DisneyInStockData
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
		return monitor.GetVariationInfo(body, pid)
	case 404:
		monitor.PublishEvent(enums.UnableToFindProduct, enums.MonitorUpdate, nil)
	default:
		fmt.Printf("Unknown Code:%v", resp.StatusCode)
	}

	return sizes, colors, stockData, nil
}

func (monitor *Monitor) GetVariationInfo(body, pid string) ([]string, []string, DisneyInStockData, error) {
	var sizes []string
	var colors []string
	var stockData DisneyInStockData

	stockResponse := DisneyStockResponse{}
	err := json.Unmarshal([]byte(body), &stockResponse)
	if err != nil {
		return sizes, colors, stockData, err
	}

	productInfo := stockResponse.Product
	if productInfo.ID == "" || productInfo.ProductName == "" || productInfo.ProductType == "" ||
		!productInfo.Available || stockResponse.ATCState.IsDisabled || stockResponse.ATCState.IsSoldOut {
		return sizes, colors, stockData, nil
	}

	price, err := strconv.ParseFloat(productInfo.Price.Sales.Price, 64)
	if err != nil {
		return sizes, colors, stockData, nil
	}

	imageURL := ""
	for _, image := range productInfo.Images.HighResImages {
		if image.ImageURL != "" {
			imageURL = image.ImageURL
			break
		}
	}
	if imageURL == "" {
		for _, image := range productInfo.Images.LargeImages {
			if image.ImageURL != "" {
				imageURL = image.ImageURL
				break
			}
		}
	}
	if imageURL == "" {
		for _, image := range productInfo.Images.SmallImages {
			if image.ImageURL != "" {
				imageURL = image.ImageURL
				break
			}
		}
	}
	if imageURL == "" {
		return sizes, colors, stockData, nil
	}

	if productInfo.ProductType == "standard" {
		stockData = DisneyInStockData{
			PID:             productInfo.ID,
			VID:             productInfo.ID,
			ProductName:     productInfo.ProductName,
			ItemURL:         BaseEndpoint + productInfo.ProductURL,
			ImageURL:        imageURL,
			Price:           int(price),
			OutOfPriceRange: monitor.PidWithInfo[pid].MaxPrice != -1 && monitor.PidWithInfo[pid].MaxPrice < int(price),
			QuantityLimit:   productInfo.QuantityLimit,
			IsPreOrder:      productInfo.Availability.IsPreOrder,
			IsBackOrder:     productInfo.Availability.IsBackOrder,
		}
		return sizes, colors, stockData, nil
	} else {
		for _, variant := range productInfo.Variants {
			if variant.Attribute == "size" {
				for _, size := range variant.Values {
					if size.Selectable {
						sizes = append(sizes, size.Value)
					}
				}
			}
			if variant.Attribute == "color" {
				for _, color := range variant.Values {
					if color.Selectable {
						colors = append(colors, color.Value)
					}
				}
			}
		}
	}

	stockData = DisneyInStockData{
		Price:         int(price),
		ProductName:   productInfo.ProductName,
		ItemURL:       BaseEndpoint + productInfo.ProductURL,
		ImageURL:      imageURL,
		QuantityLimit: productInfo.QuantityLimit,
		IsPreOrder:    productInfo.Availability.IsPreOrder,
		IsBackOrder:   productInfo.Availability.IsBackOrder,
	}
	return sizes, colors, stockData, nil
}

func (monitor *Monitor) GetInStockVariations(pid string, sizes, colors []string) []DisneyInStockData {
	if len(colors) == 0 {
		colors = append(colors, "")
	}
	if len(sizes) == 0 {
		sizes = append(sizes, "")
	}

	wg := sync.WaitGroup{}
	wg.Add(len(colors) * len(sizes))

	var stockDatas []DisneyInStockData
	for _, color := range colors {
		for _, size := range sizes {
			go func(x, y, z string) {
				stockData := monitor.GetInStockVariant(x, y, z)
				if stockData.PID != "" {
					stockDatas = append(stockDatas, stockData)
				}
				wg.Done()
			}(pid, size, color)
		}
	}
	wg.Wait()
	return stockDatas
}

func (monitor *Monitor) GetInStockVariant(pid string, size, color string) DisneyInStockData {
	var stockData DisneyInStockData

	endpoint := fmt.Sprintf(MonitorEndpoint2, pid, pid, size, pid, color)

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
		return stockData
	}

	switch resp.StatusCode {
	case 200:
		stockResponse := DisneyStockResponse{}
		err := json.Unmarshal([]byte(body), &stockResponse)
		if err != nil {
			return stockData
		}

		if stockResponse.Product.Available {
			stockData.PID = pid
			stockData.VID = stockResponse.Product.ID
			stockData.Size = size
			stockData.Color = color
		}
		return stockData
	case 404:
		monitor.PublishEvent(enums.UnableToFindProduct, enums.MonitorUpdate, nil)
	default:
		fmt.Printf("Unknown Code:%v", resp.StatusCode)
	}

	return stockData
}
