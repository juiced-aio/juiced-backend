package hottopic

import (
	"fmt"
	"strconv"
	"strings"
	"sync"
	"time"

	"backend.juicedbot.io/juiced.infrastructure/common/entities"
	"backend.juicedbot.io/juiced.infrastructure/common/enums"
	"backend.juicedbot.io/juiced.infrastructure/common/events"
	"backend.juicedbot.io/juiced.sitescripts/base"
	"backend.juicedbot.io/juiced.sitescripts/util"
	"github.com/anaskhan96/soup"
	browser "github.com/eddycjy/fake-useragent"
)

// CreateHottopicMonitor takes a TaskGroup entity and turns it into a Hottopic Monitor
func CreateHottopicMonitor(taskGroup *entities.TaskGroup, proxyGroup *entities.ProxyGroup, eventBus *events.EventBus, singleMonitors []entities.HottopicSingleMonitorInfo) (Monitor, error) {
	storedHottopicMonitors := make(map[string]entities.HottopicSingleMonitorInfo)
	hottopicMonitor := Monitor{}

	pids := []string{}
	for _, monitor := range singleMonitors {
		storedHottopicMonitors[monitor.Pid] = monitor
		pids = append(pids, monitor.Pid)
	}

	hottopicMonitor = Monitor{
		Monitor: base.Monitor{
			TaskGroup:  taskGroup,
			ProxyGroup: proxyGroup,
			EventBus:   eventBus,
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
		err := monitor.Monitor.CreateClient()
		if err != nil {
			return
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

	defer func() {
		if recover() != nil {
			time.Sleep(time.Duration(monitor.Monitor.TaskGroup.MonitorDelay) * time.Millisecond)
			monitor.RunSingleMonitor(pid)
		}
	}()

	var sizes []HottopicSizeInfo
	var colors []string
	var stockData HottopicInStockData
	var err error

	var proxy *entities.Proxy
	if monitor.Monitor.ProxyGroup != nil {
		if len(monitor.Monitor.ProxyGroup.Proxies) > 0 {
			proxy = util.RandomLeastUsedProxy(monitor.Monitor.ProxyGroup.Proxies)
			monitor.Monitor.UpdateProxy(proxy)
		}
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
	}

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
				filteredSizes := []HottopicSizeInfo{}
				for _, size := range sizes {
					if strings.Contains(strings.ToLower(sizesJoined), strings.ToLower(size.Size)) {
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
			if len(sizes) > 0 && len(colors) > 0 {
				needToStop = monitor.CheckForStop()
				if needToStop {
					return
				}

				var stockDatas []HottopicInStockData
				if oneColorBeforeFilter {
					for _, size := range sizes {
						stockData := HottopicInStockData{
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
							if stockData.ImageURL == "" {
								stockData.ImageURL = imageURL
							}
							// Add each in stock combination to the monitor's InStock list, then update the status
							monitor.InStock = append(monitor.InStock, stockData)
							atLeastOneInPriceRange = true
						}
					}
					if atLeastOneInPriceRange {
						// If at least one combination is in stock and in our price range, remove this monitor from the running monitors
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
				}
			} else {
				// None of the available sizes/colors match the task's size/color filters
				if stockData.ProductName != "" && stockData.ImageURL != "" {
					monitor.PublishEvent(enums.WaitingForInStock, enums.MonitorUpdate, events.ProductInfo{
						Products: []events.Product{
							{ProductName: productName, ProductImageURL: imageURL}},
					})
				}
			}
		} else {
			// This code is only run for items that have no size/color variations
			if stockData.PID != "" && !stockData.OutOfPriceRange {
				var inSlice bool
				for _, monitorStock := range monitor.InStock {
					inSlice = monitorStock.PID == stockData.PID
				}
				if !inSlice {
					monitor.InStock = append(monitor.InStock, stockData)
					monitor.PublishEvent(enums.SendingProductInfoToTasks, enums.MonitorUpdate, events.ProductInfo{
						Products: []events.Product{
							{ProductName: stockData.ProductName, ProductImageURL: stockData.ImageURL}},
					})
				}
			} else {
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
				for i, monitorStock := range monitor.InStock {
					if monitorStock.PID == stockData.PID {
						monitor.InStock = append(monitor.InStock[:i], monitor.InStock[i+1:]...)
						break
					}
				}
			}
		}
	}

	time.Sleep(time.Duration(monitor.Monitor.TaskGroup.MonitorDelay) * time.Millisecond)
	monitor.RunSingleMonitor(pid)
}

func (monitor *Monitor) GetSizeAndColor(pid string) ([]HottopicSizeInfo, []string, HottopicInStockData, error) {
	var sizes []HottopicSizeInfo
	var colors []string
	var stockData HottopicInStockData
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
		return monitor.GetVariationInfo(body, pid)
	case 404:
		monitor.PublishEvent(enums.UnableToFindProduct, enums.MonitorUpdate, nil)
	default:
		fmt.Printf("Unknown Code:%v", resp.StatusCode)
	}

	return sizes, colors, stockData, nil
}

func (monitor *Monitor) GetVariationInfo(body, pid string) ([]HottopicSizeInfo, []string, HottopicInStockData, error) {
	var sizes []HottopicSizeInfo
	var colors []string
	var stockData HottopicInStockData

	doc := soup.HTMLParse(body)

	// We need the product name and image URL no matter what
	productNameHeader := doc.Find("h1", "class", "productdetail__info-title")
	if productNameHeader.Error != nil {
		return sizes, colors, stockData, productNameHeader.Error
	}
	productName := productNameHeader.Text()
	stockData.ProductName = productName

	productImage := doc.Find("img", "class", "productdetail__image-active-each")
	if productImage.Error != nil {
		return sizes, colors, stockData, productImage.Error
	}
	imageURL := productImage.Attrs()["src"]
	stockData.ImageURL = imageURL

	hasVariations := false
	// This element will only exist if the product has no size/color variations and the item is in stock
	inStockText := doc.Find("p", "class", "in-stock-msg")
	if inStockText.Error != nil || !strings.Contains(strings.ToLower(inStockText.Text()), "in stock") {
		// This element will only exist if the product has no size/color variations and the item is on backorder
		backorderText := doc.Find("span", "class", "text-red")
		if backorderText.Error != nil || !strings.Contains(strings.ToLower(backorderText.Text()), "backorder") {
			// This element will only exist if the product has no size/color variations and the item is out of stock
			outOfStockText := doc.Find("p", "class", "not-available-msg")
			if outOfStockText.Error == nil && strings.Contains(strings.ToLower(outOfStockText.Text()), "out of stock") {
				return sizes, colors, stockData, nil
			} else {
				// If none of these exist, then the product has variations
				hasVariations = true
			}
		}
	}

	if !hasVariations {
		// If the product is in stock and has no size/color variations, we have all we need to know on this page
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
		stockData = HottopicInStockData{
			PID:             pid,
			SizePID:         pid,
			Color:           defaultColor,
			ProductName:     productName,
			ImageURL:        imageURL,
			Price:           int(price),
			OutOfPriceRange: monitor.PidWithInfo[pid].MaxPrice != -1 && monitor.PidWithInfo[pid].MaxPrice < int(price),
		}
		return sizes, colors, stockData, nil
	} else {
		// Otherwise, we need to build a list of available colors and sizes and get more info
		colorList := doc.Find("div", "class", "productdetail__info-form-color-swatch-container")
		if colorList.Error == nil {
			colorListItems := colorList.FindAll("a", "class", "productdetail__info-form-color-swatch-link")
			for _, colorListItem := range colorListItems {
				if colorListItem.Error == nil {
					color := colorListItem.Attrs()["title"]
					colors = append(colors, color)
				}
			}
		} else {
			defaultColorInput := doc.Find("input", "id", "productColor")
			defaultColor := ""
			if defaultColorInput.Error == nil {
				defaultColor = defaultColorInput.Attrs()["value"]
			}
			colors = append(colors, defaultColor)
		}

		sizeList := doc.Find("div", "class", "productdetail__info-form-size-swatch-container")
		if sizeList.Error == nil {
			sizeListItems := sizeList.FindAll("li", "class", "productdetail__info-form-size-swatch-link-container")
			for index, sizeListItem := range sizeListItems {
				if sizeListItem.Error == nil {
					intPid, err := strconv.Atoi(pid)
					if err == nil {
						sizeListLink := sizeListItem.Find("a")
						if sizeListLink.Error == nil {
							size := sizeListLink.Attrs()["title"]
							if len(colors) > 1 || !strings.Contains(sizeListItem.Attrs()["class"], "unselectable") {
								sizes = append(sizes, HottopicSizeInfo{
									SizePID: fmt.Sprint(intPid + index + 1),
									Size:    size,
								})
							}
						}
					}
				}
			}
		} else {
			sizes = append(sizes, HottopicSizeInfo{
				SizePID: pid,
			})
		}
	}
	stockData = HottopicInStockData{
		ProductName: productName,
		ImageURL:    imageURL,
	}
	if len(colors) == 1 {
		// If there is only one color, we can grab the rest of the info on this page
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
		stockData.Color = colors[0]
		stockData.Price = int(price)
		stockData.OutOfPriceRange = monitor.PidWithInfo[pid].MaxPrice != -1 && monitor.PidWithInfo[pid].MaxPrice < int(price)
	}
	return sizes, colors, stockData, nil
}

func (monitor *Monitor) GetInStockVariations(pid string, sizes []HottopicSizeInfo, colors []string) []HottopicInStockData {
	// Each color page shows us whether the individual sizes are in stock or not
	wg := sync.WaitGroup{}
	wg.Add(len(colors))

	var stockDatas []HottopicInStockData
	for _, color := range colors {
		go func(x string, y []HottopicSizeInfo, z string) {
			stockDatas = append(stockDatas, monitor.GetInStockSizesForColor(x, y, z)...)
			wg.Done()
		}(pid, sizes, color)
	}
	wg.Wait()
	return stockDatas
}

func (monitor *Monitor) GetInStockSizesForColor(pid string, sizes []HottopicSizeInfo, color string) []HottopicInStockData {
	var stockDatas []HottopicInStockData

	wg := sync.WaitGroup{}
	wg.Add(len(sizes))

	for _, size := range sizes {
		go func(s string) {
			defer wg.Done()

			endpoint := fmt.Sprintf(MonitorEndpoint2, pid, pid, color, pid, s)
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
			}

			switch resp.StatusCode {
			case 200:
				stockDatas = append(stockDatas, monitor.GetColorVariationInfo(body, pid, color, sizes)...)
			case 404:
				monitor.PublishEvent(enums.UnableToFindProduct, enums.MonitorUpdate, nil)
			default:
				fmt.Printf("Unknown Code:%v", resp.StatusCode)
			}
		}(size.Size)
	}

	wg.Wait()

	return stockDatas
}

func (monitor *Monitor) GetColorVariationInfo(body, pid, color string, sizes []HottopicSizeInfo) []HottopicInStockData {
	var stockDatas []HottopicInStockData
	var sizePID string

	doc := soup.HTMLParse(body)

	pidItem := doc.Find("input", "id", "pid")
	if pidItem.Error == nil {
		sizePID = pidItem.Attrs()["value"]
	}

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

	var imageURL string
	imageItem := doc.Find("img", "class", "productdetail__image-active-each")
	if imageItem.Error == nil {
		imageURL = imageItem.Attrs()["src"]
	}

	sizeList := doc.Find("ul", "class", "productdetail__info-form-size-swatch")
	if sizeList.Error == nil {
		sizeListLinks := sizeList.FindAll("a", "class", "productdetail__info-form-size-swatch-link")
		for _, sizeListLink := range sizeListLinks {
			if sizeListLink.Error == nil {
				size := sizeListLink.Attrs()["title"]
				matchedSize := false
				for _, s := range sizes {
					if s.Size == size {
						matchedSize = true
						break
					}
				}
				if matchedSize {
					stockDatas = append(stockDatas, HottopicInStockData{
						PID:             pid,
						SizePID:         sizePID,
						Size:            size,
						Color:           color,
						Price:           int(price),
						ImageURL:        imageURL,
						OutOfPriceRange: monitor.PidWithInfo[pid].MaxPrice != -1 && monitor.PidWithInfo[pid].MaxPrice < int(price),
					})
				}

			}
		}
	}

	return stockDatas
}
