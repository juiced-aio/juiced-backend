package disney

import (
	"encoding/json"
	"fmt"
	"log"
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

	var sizes []DisneySizeInfo
	var colors []string
	var stockData DisneyInStockData
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
	// Size, no colors: https://www.Disney.com/on/demandware.store/Sites-Disney-Site/default/Product-Variation?pid=13941979&Quantity=1&format=ajax
	// Color, no sizes:
	// Neither color nor size: https://www.Disney.com/on/demandware.store/Sites-Disney-Site/default/Product-Variation?pid=15647180&Quantity=1&format=ajax

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

	sizes, colors, stockData, err = monitor.GetSizeAndColor(pid)

	log.Println(sizes)
	log.Println(colors)
	log.Println(stockData)

	needToStop = monitor.CheckForStop()
	if needToStop {
		return
	}

	productName := stockData.ProductName
	imageURL := stockData.ImageURL

	if err == nil {
		// GetSizeAndColor will only return a populated stockData if the product has no size/color variations
		// 		stockData.ProductName and stockData.ImageURL will always be populated
		if stockData.PID == "" {
			// If there's only one color, we already know that the size list only has in-stock sizes
			oneColorBeforeFilter := len(colors) == 1

			// If stockData.PID == "", then the size and color lists should be populated
			// Filter the sizes we found with the ones the monitor has been provided
			sizesJoined := monitor.PidWithInfo[pid].Size
			if sizesJoined != "" {
				filteredSizes := []DisneySizeInfo{}
				for _, size := range sizes {
					if strings.Contains(sizesJoined, size.Size) {
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
					if strings.Contains(colorsJoined, color) {
						filteredColors = append(filteredColors, color)
					}
				}
				colors = filteredColors
			}

			// If sizes and colors remain, continue
			if len(sizes) > 0 && len(colors) > 0 {
				needToStop = monitor.CheckForStop()
				if needToStop {
					return
				}

				var stockDatas []DisneyInStockData
				if oneColorBeforeFilter {
					for _, size := range sizes {
						stockData := DisneyInStockData{
							PID:             pid,
							Price:           stockData.Price,
							SizePID:         size.SizePID,
							Size:            size.Size,
							Color:           colors[0],
							ProductName:     stockData.ProductName,
							ImageURL:        stockData.ImageURL,
							OutOfPriceRange: stockData.OutOfPriceRange,
						}
						stockDatas = append(stockDatas, stockData)
					}
				} else {
					// GetInStockVariations returns a list of HottopicStockData items for each size/color combination that's in stock
					stockDatas = monitor.GetInStockVariations(pid, sizes, colors)
					needToStop = monitor.CheckForStop()
					if needToStop {
						return
					}
				}

				if len(stockDatas) > 0 {
					atLeastOneInPriceRange := false
					for _, stockData := range stockDatas {
						if !stockData.OutOfPriceRange {
							// Since we omitted these fields in the function below, add them back here
							stockData.ProductName = productName
							stockData.ImageURL = imageURL
							// Add each in stock combination to the monitor's InStock list, then update the status
							monitor.InStock = append(monitor.InStock, stockData)
							atLeastOneInPriceRange = true
						}
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

func (monitor *Monitor) GetSizeAndColor(pid string) ([]DisneySizeInfo, []string, DisneyInStockData, error) {
	var sizes []DisneySizeInfo
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

func (monitor *Monitor) GetVariationInfo(body, pid string) ([]DisneySizeInfo, []string, DisneyInStockData, error) {
	var sizes []DisneySizeInfo
	var colors []string
	var stockData DisneyInStockData

	stockResponse := DisneyStockResponse{}
	err := json.Unmarshal([]byte(body), &stockResponse)
	if err != nil {
		log.Println(1)
		log.Println(err.Error())
		return sizes, colors, stockData, err
	}

	productInfo := stockResponse.Product
	if productInfo.ID == "" || productInfo.ProductName == "" || productInfo.ProductType == "" ||
		!productInfo.Available || stockResponse.ATCState.IsDisabled || stockResponse.ATCState.IsSoldOut {
		log.Println(2)
		log.Println(productInfo.ID)
		log.Println(productInfo.ProductName)
		log.Println(productInfo.ProductType)
		log.Println(productInfo.Available)
		log.Println(stockResponse.ATCState.IsDisabled)
		log.Println(stockResponse.ATCState.IsSoldOut)
		return sizes, colors, stockData, nil
	}

	price, err := strconv.ParseFloat(productInfo.Price.Sales.Price, 64)
	if err != nil {
		log.Println(3)
		log.Println(productInfo.Price.Sales.Price)
		log.Println(err.Error())
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
		log.Println(4)
		log.Println(productInfo.Images)
		return sizes, colors, stockData, nil
	}

	if productInfo.ProductType == "standard" {
		stockData = DisneyInStockData{
			PID:             productInfo.ID,
			SizePID:         productInfo.ID,
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
	} else if productInfo.ProductType == "master" {
		// Multiple sizes, one color
	} else {
		// Multiple sizes and colors
	}

	stockData = DisneyInStockData{
		ProductName: productInfo.ProductName,
		ImageURL:    imageURL,
	}
	if len(colors) == 1 {

	}
	return sizes, colors, stockData, nil
}

func (monitor *Monitor) GetInStockVariations(pid string, sizes []DisneySizeInfo, colors []string) []DisneyInStockData {
	// Each color page shows us whether the individual sizes are in stock or not
	var stockDatas []DisneyInStockData
	for _, color := range colors {
		stockDatas = append(stockDatas, monitor.GetInStockSizesForColor(pid, sizes, color)...)
		needToStop := monitor.CheckForStop()
		if needToStop {
			return stockDatas
		}
		time.Sleep(100 * time.Millisecond)
	}
	return stockDatas
}

func (monitor *Monitor) GetInStockSizesForColor(pid string, sizes []DisneySizeInfo, color string) []DisneyInStockData {
	var stockDatas []DisneyInStockData

	endpoint := fmt.Sprintf(MonitorEndpoint2, pid, pid) + color

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
		return stockDatas
	}

	switch resp.StatusCode {
	case 200:
		return monitor.GetColorVariationInfo(body, pid, color, sizes)
	case 404:
		monitor.PublishEvent(enums.UnableToFindProduct, enums.MonitorUpdate, nil)
	default:
		fmt.Printf("Unknown Code:%v", resp.StatusCode)
	}

	return stockDatas
}

func (monitor *Monitor) GetColorVariationInfo(body, pid, color string, sizes []DisneySizeInfo) []DisneyInStockData {
	var stockDatas []DisneyInStockData

	doc := soup.HTMLParse(body)

	priceText := doc.Find("span", "class", "productdetail__info-pricing-sale")
	if priceText.Error != nil {
		priceText = doc.Find("span", "class", "productdetail__info-pricing-original")
	}
	if priceText.Error != nil {
		return stockDatas
	}
	priceStr := strings.ReplaceAll(strings.ReplaceAll(priceText.Text(), " ", ""), "$", "")
	if strings.Contains(priceStr, "-") {
		priceStr = strings.Split(priceStr, "-")[1]
	}
	price, err := strconv.ParseFloat(priceStr, 64)
	if err != nil {
		return stockDatas
	}

	sizeList := doc.Find("ul", "class", "productdetail__info-form-size-swatch")
	if sizeList.Error == nil {
		sizeListLinks := sizeList.FindAll("a", "class", "productdetail__info-form-size-swatch-link")
		for index, sizeListLink := range sizeListLinks {
			if sizeListLink.Error == nil {
				intPid, err := strconv.Atoi(pid)
				if err == nil {
					size := sizeListLink.Attrs()["title"]
					matchedSize := false
					for _, s := range sizes {
						if s.Size == size {
							matchedSize = true
							break
						}
					}
					if matchedSize {
						stockDatas = append(stockDatas, DisneyInStockData{
							PID:             pid,
							SizePID:         fmt.Sprint(intPid + index + 1),
							Size:            size,
							Color:           color,
							Price:           int(price),
							OutOfPriceRange: monitor.PidWithInfo[pid].MaxPrice != -1 && monitor.PidWithInfo[pid].MaxPrice < int(price),
						})
					}
				}
			}
		}
	}

	return stockDatas
}
