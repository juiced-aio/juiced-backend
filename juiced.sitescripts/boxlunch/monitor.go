package boxlunch

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
	"sync"

	"backend.juicedbot.io/juiced.infrastructure/entities"
	"backend.juicedbot.io/juiced.infrastructure/enums"
	"backend.juicedbot.io/juiced.sitescripts/util"

	"github.com/anaskhan96/soup"
	browser "github.com/eddycjy/fake-useragent"
	"github.com/fatih/structs"
)

func CreateMonitor(input entities.MonitorInput, baseMonitor *entities.BaseMonitor) (entities.Monitor, error) {
	switch input.MonitorType {
	case enums.SKUMonitor:
		boxLunchMonitorInput, err := ValidateMonitorInput(input.Input, input.MonitorType, input.SiteSpecificInfo)
		if err != nil {
			return nil, err
		}
		return &SKUMonitor{
			Input:         input,
			BoxLunchInput: boxLunchMonitorInput,
			BaseMonitor:   baseMonitor,
		}, nil

	}
	return nil, &enums.UnsupportedMonitorTypeError{Retailer: enums.BoxLunch, MonitorType: input.MonitorType}
}

func (monitor *SKUMonitor) GetProductInfos() ([]entities.ProductInfo, error) {
	monitor.ItemName = ""
	monitor.ImageURL = ""
	monitor.Price = 0
	sizes, colors, productInfos, err := monitor.GetSizeAndColor(monitor.Input.Input)
	if err != nil {
		return productInfos, err
	}

	if len(productInfos) > 0 {
		return productInfos, nil
	}

	// If stockData.PID == "", then the size and color lists should be populated
	// Filter the sizes we found with the ones the monitor has been provided
	sizesJoined := monitor.BoxLunchInput.Size
	if sizesJoined != "" {
		filteredSizes := []SizeInfo{}
		for _, size := range sizes {
			if strings.Contains(strings.ToLower(sizesJoined), strings.ToLower(size.Size)) {
				filteredSizes = append(filteredSizes, size)
			}
		}
		sizes = filteredSizes
	}
	// Filter the colors we found with the ones the monitor has been provided
	colorsJoined := monitor.BoxLunchInput.Color
	if colorsJoined != "" {
		filteredColors := []string{}
		for _, color := range colors {
			if strings.Contains(strings.ToLower(colorsJoined), strings.ToLower(color)) {
				filteredColors = append(filteredColors, color)
			}
		}
		colors = filteredColors
	}

	if len(sizes) > 0 && len(colors) > 0 {
		productInfos = monitor.GetInStockVariations(monitor.Input.Input, sizes, colors)
		if len(productInfos) > 0 {
			return productInfos, nil
		}
	}

	productInfos = append(productInfos, entities.ProductInfo{
		InStock:      false,
		InPriceRange: false,
		SKU:          monitor.Input.Input,
		Price:        0,
		ItemName:     monitor.ItemName,
		ItemURL:      fmt.Sprintf(ProductEndpoint, monitor.Input.Input),
		ImageURL:     monitor.ImageURL,
	})

	return productInfos, nil
}

func (monitor *SKUMonitor) GetSizeAndColor(pid string) ([]SizeInfo, []string, []entities.ProductInfo, error) {
	var productInfos []entities.ProductInfo
	endpoint := fmt.Sprintf(MonitorEndpoint, pid)

	resp, body, err := util.MakeRequest(&util.Request{
		Client: monitor.BaseMonitor.Client,
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
		return []SizeInfo{}, []string{}, productInfos, err
	}

	switch resp.StatusCode {
	case 200:
		return monitor.GetVariationInfo(body, pid)
	case 404:
		return []SizeInfo{}, []string{}, productInfos, errors.New("404 product not found")
	}

	return []SizeInfo{}, []string{}, productInfos, errors.New("unknown error")
}

func (monitor *SKUMonitor) GetVariationInfo(body, pid string) ([]SizeInfo, []string, []entities.ProductInfo, error) {
	var productInfos []entities.ProductInfo
	var sizes []SizeInfo
	var colors []string

	doc := soup.HTMLParse(body)

	// We need the product name and image URL no matter what
	productNameHeader := doc.Find("h1", "class", "productdetail__info-title")
	if productNameHeader.Error != nil {
		return sizes, colors, productInfos, productNameHeader.Error
	}
	productName := productNameHeader.Text()
	monitor.ItemName = productName

	productImage := doc.Find("img", "class", "productdetail__image-active-each")
	if productImage.Error != nil {
		return sizes, colors, productInfos, productImage.Error
	}
	imageURL := productImage.Attrs()["src"]
	monitor.ImageURL = imageURL

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
				productInfos = append(productInfos, entities.ProductInfo{
					InStock:      false,
					InPriceRange: false,
					SKU:          pid,
					Price:        0,
					ItemName:     productName,
					ItemURL:      fmt.Sprintf(ProductEndpoint, pid),
					ImageURL:     imageURL,
				})
				return sizes, colors, productInfos, nil
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
			return sizes, colors, productInfos, priceText.Error
		}
		priceStr := strings.ReplaceAll(strings.ReplaceAll(priceText.Text(), " ", ""), "$", "")
		price, err := strconv.ParseFloat(priceStr, 64)
		monitor.Price = price
		if err != nil {
			return sizes, colors, productInfos, err
		}
		defaultColorInput := doc.Find("input", "id", "productColor")
		if defaultColorInput.Error != nil {
			return sizes, colors, productInfos, defaultColorInput.Error
		}
		defaultColor := defaultColorInput.Attrs()["value"]
		productInfos = append(productInfos, entities.ProductInfo{
			InStock:      true,
			InPriceRange: monitor.Input.MaxPrice == -1 || int(price) <= monitor.Input.MaxPrice,
			SKU:          pid,
			Price:        price,
			ItemName:     productName,
			ItemURL:      fmt.Sprintf(ProductEndpoint, pid),
			ImageURL:     imageURL,
			SiteSpecificInfo: structs.Map(BoxLunchProductInfo{
				VID:   pid,
				Size:  "",
				Color: defaultColor,
			}),
		})
		return sizes, colors, productInfos, nil
	}

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
							sizes = append(sizes, SizeInfo{
								VID:  fmt.Sprint(intPid + index + 1),
								Size: size,
							})
						}
					}
				}
			}
		}
	} else {
		sizes = append(sizes, SizeInfo{
			VID: pid,
		})
	}

	if len(colors) == 1 {
		// If there is only one color, we can grab the rest of the info on this page
		priceText := doc.Find("span", "class", "productdetail__info-pricing-sale")
		if priceText.Error != nil {
			priceText = doc.Find("span", "class", "productdetail__info-pricing-original")
		}
		if priceText.Error != nil {
			return sizes, colors, productInfos, priceText.Error
		}
		priceStr := strings.ReplaceAll(strings.ReplaceAll(priceText.Text(), " ", ""), "$", "")
		price, err := strconv.ParseFloat(priceStr, 64)
		if err != nil {
			return sizes, colors, productInfos, err
		}

		for _, size := range sizes {
			productInfos = append(productInfos, entities.ProductInfo{
				InStock:      true,
				InPriceRange: monitor.Input.MaxPrice == -1 || int(price) <= monitor.Input.MaxPrice,
				SKU:          pid,
				Price:        price,
				ItemName:     productName,
				ItemURL:      fmt.Sprintf(ProductEndpoint, pid),
				ImageURL:     imageURL,
				SiteSpecificInfo: structs.Map(BoxLunchProductInfo{
					VID:   size.VID,
					Size:  size.Size,
					Color: colors[0],
				}),
			})
		}
	}
	return sizes, colors, productInfos, nil
}

func (monitor *SKUMonitor) GetInStockVariations(pid string, sizes []SizeInfo, colors []string) []entities.ProductInfo {
	var productInfos []entities.ProductInfo

	wg := sync.WaitGroup{}
	wg.Add(len(colors))

	for _, color := range colors {
		go func(x string, y []SizeInfo, z string) {
			productInfos = append(productInfos, monitor.GetInStockSizesForColor(x, y, z)...)
			wg.Done()
		}(pid, sizes, color)
	}
	wg.Wait()
	return productInfos
}

func (monitor *SKUMonitor) GetInStockSizesForColor(pid string, sizes []SizeInfo, color string) []entities.ProductInfo {
	var productInfos []entities.ProductInfo

	wg := sync.WaitGroup{}
	wg.Add(len(sizes))

	for _, size := range sizes {
		go func(s string) {
			defer wg.Done()

			endpoint := fmt.Sprintf(MonitorEndpoint2, pid, pid, color, pid, s)
			resp, body, err := util.MakeRequest(&util.Request{
				Client: monitor.BaseMonitor.Client,
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
			if err == nil {
				switch resp.StatusCode {
				case 200:
					productInfos = append(productInfos, monitor.GetColorVariationInfo(body, pid, color, sizes)...)
				}
			}
		}(size.Size)
	}

	wg.Wait()

	return productInfos
}

func (monitor *SKUMonitor) GetColorVariationInfo(body, pid, color string, sizes []SizeInfo) []entities.ProductInfo {
	var productInfos []entities.ProductInfo
	var vid string

	doc := soup.HTMLParse(body)

	pidItem := doc.Find("input", "id", "pid")
	if pidItem.Error == nil {
		vid = pidItem.Attrs()["value"]
	}

	priceText := doc.Find("span", "class", "productdetail__info-pricing-sale")
	if priceText.Error != nil {
		priceText = doc.Find("span", "class", "productdetail__info-pricing-original")
	}
	if priceText.Error != nil {
		return productInfos
	}
	priceStr := strings.ReplaceAll(strings.ReplaceAll(priceText.Text(), " ", ""), "$", "")
	if strings.Contains(priceStr, "-") {
		priceStr = strings.Split(priceStr, "-")[1]
	}
	price, err := strconv.ParseFloat(priceStr, 64)
	if err != nil {
		return productInfos
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
					productInfos = append(productInfos, entities.ProductInfo{
						InStock:      true,
						InPriceRange: monitor.Input.MaxPrice == -1 || int(price) <= monitor.Input.MaxPrice,
						SKU:          pid,
						Price:        price,
						ItemName:     monitor.ItemName,
						ItemURL:      fmt.Sprintf(ProductEndpoint, pid),
						ImageURL:     imageURL,
						SiteSpecificInfo: structs.Map(BoxLunchProductInfo{
							VID:   vid,
							Size:  size,
							Color: color,
						}),
					})
				}

			}
		}
	}

	return productInfos
}
