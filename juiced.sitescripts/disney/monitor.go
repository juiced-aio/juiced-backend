package disney

import (
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"sync"

	"backend.juicedbot.io/juiced.infrastructure/entities"
	"backend.juicedbot.io/juiced.infrastructure/enums"
	"backend.juicedbot.io/juiced.sitescripts/util"

	browser "github.com/eddycjy/fake-useragent"
	"github.com/fatih/structs"
)

// Test
//	1. Standard
// 	2. Size variants
// 	3. Color variants
//	4. Size and color variants
// 	5. Non-standard but still no variants

func CreateMonitor(input entities.MonitorInput, baseMonitor *entities.BaseMonitor) (entities.Monitor, error) {
	switch input.MonitorType {
	case enums.SKUMonitor:
		disneyMonitorInput, err := ValidateMonitorInput(input.Input, input.MonitorType, input.SiteSpecificInfo)
		if err != nil {
			return nil, err
		}
		return &SKUMonitor{
			Input:       input,
			DisneyInput: disneyMonitorInput,
			BaseMonitor: baseMonitor,
		}, nil

	}
	return nil, &enums.UnsupportedMonitorTypeError{Retailer: enums.Disney, MonitorType: input.MonitorType}
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
	sizesJoined := monitor.DisneyInput.Size
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
	colorsJoined := monitor.DisneyInput.Color
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

	monitorResponse := MonitorResponse{}
	err := json.Unmarshal([]byte(body), &monitorResponse)
	if err != nil {
		return sizes, colors, productInfos, err
	}

	productInfo := monitorResponse.Product
	if productInfo.ID == "" || productInfo.ProductName == "" || productInfo.ProductType == "" {
		return sizes, colors, productInfos, nil
	}

	price, err := strconv.ParseFloat(productInfo.Price.Sales.Price, 64)
	if err != nil {
		return sizes, colors, productInfos, nil
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
	// if imageURL == "" {
	// 	return sizes, colors, productInfos, nil
	// }

	if productInfo.ProductType == "standard" {
		productInfos = append(productInfos, entities.ProductInfo{
			InStock:      productInfo.Available || !monitorResponse.ATCState.IsDisabled || !monitorResponse.ATCState.IsSoldOut,
			InPriceRange: monitor.Input.MaxPrice == -1 || int(price) <= monitor.Input.MaxPrice,
			SKU:          productInfo.ID,
			Price:        price,
			ItemName:     productInfo.ProductName,
			ItemURL:      BaseEndpoint + productInfo.ProductURL,
			ImageURL:     imageURL,
			SiteSpecificInfo: structs.Map(DisneyProductInfo{
				VID:           productInfo.ID,
				Size:          "",
				Color:         "",
				QuantityLimit: productInfo.QuantityLimit,
				IsPreOrder:    productInfo.Availability.IsPreOrder,
				IsBackOrder:   productInfo.Availability.IsBackOrder,
			}),
		})
		return sizes, colors, productInfos, nil
	} else {
		for _, variant := range productInfo.Variants {
			if variant.Attribute == "size" {
				for _, size := range variant.Values {
					if size.Selectable {
						sizes = append(sizes, SizeInfo{
							VID:  size.Value,
							Size: size.Value,
						})
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

	// productInfos = append(productInfos, entities.ProductInfo{
	// 	InStock:      productInfo.Available || !monitorResponse.ATCState.IsDisabled || !monitorResponse.ATCState.IsSoldOut,
	// 	InPriceRange: monitor.Input.MaxPrice == -1 || int(price) <= monitor.Input.MaxPrice,
	// 	SKU:          productInfo.ID,
	// 	Price:        price,
	// 	ItemName:     productInfo.ProductName,
	// 	ItemURL:      BaseEndpoint + productInfo.ProductURL,
	// 	ImageURL:     imageURL,
	// 	SiteSpecificInfo: structs.Map(DisneyProductInfo{
	// 		VID:           "",
	// 		Size:          "",
	// 		Color:         "",
	// 		QuantityLimit: productInfo.QuantityLimit,
	// 		IsPreOrder:    productInfo.Availability.IsPreOrder,
	// 		IsBackOrder:   productInfo.Availability.IsBackOrder,
	// 	}),
	// })

	// // Filling VID just incase the ProductType is not standard but still has no variants like this product for some reason https://www.shopdisney.com/elsa-costume-wig-for-kids-frozen-2-428423206036.html?isProductSearch=0&plpPosition=9&guestFacing=Halloween%2520Shop-Costume%2520Accessories
	// stockData = DisneyInStockData{
	// 	VID:           productInfo.ID,
	// 	Price:         int(price),
	// 	ProductName:   productInfo.ProductName,
	// 	ItemURL:       BaseEndpoint + productInfo.ProductURL,
	// 	ImageURL:      imageURL,
	// 	QuantityLimit: productInfo.QuantityLimit,
	// 	IsPreOrder:    productInfo.Availability.IsPreOrder,
	// 	IsBackOrder:   productInfo.Availability.IsBackOrder,
	// }

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

	monitorResponse := MonitorResponse{}
	err := json.Unmarshal([]byte(body), &monitorResponse)
	if err != nil {
		return productInfos
	}

	price := float64(0)
	imageURL := ""

	productInfo := monitorResponse.Product
	if productInfo.Available || !monitorResponse.ATCState.IsDisabled || !monitorResponse.ATCState.IsSoldOut {
		productInfos = append(productInfos, entities.ProductInfo{
			InStock:      productInfo.Available || !monitorResponse.ATCState.IsDisabled || !monitorResponse.ATCState.IsSoldOut,
			InPriceRange: monitor.Input.MaxPrice == -1 || int(price) <= monitor.Input.MaxPrice,
			SKU:          productInfo.ID,
			Price:        price,
			ItemName:     productInfo.ProductName,
			ItemURL:      BaseEndpoint + productInfo.ProductURL,
			ImageURL:     imageURL,
			SiteSpecificInfo: structs.Map(DisneyProductInfo{
				VID:           productInfo.ID,
				Size:          "",
				Color:         "",
				QuantityLimit: productInfo.QuantityLimit,
				IsPreOrder:    productInfo.Availability.IsPreOrder,
				IsBackOrder:   productInfo.Availability.IsBackOrder,
			}),
		})
	}

	return productInfos
}
