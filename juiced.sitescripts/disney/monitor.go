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
		filteredSizes := []string{}
		for _, size := range sizes {
			if strings.Contains(strings.ToLower(sizesJoined), strings.ToLower(size)) {
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

	if len(sizes) > 0 || len(colors) > 0 {
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

func (monitor *SKUMonitor) GetSizeAndColor(pid string) ([]string, []string, []entities.ProductInfo, error) {
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
		return []string{}, []string{}, productInfos, err
	}

	switch resp.StatusCode {
	case 200:
		return monitor.GetVariationInfo(body, pid)
	case 404:
		return []string{}, []string{}, productInfos, errors.New("404 product not found")
	}

	return []string{}, []string{}, productInfos, errors.New("unknown error")
}

func (monitor *SKUMonitor) GetVariationInfo(body, pid string) ([]string, []string, []entities.ProductInfo, error) {
	var productInfos []entities.ProductInfo
	var sizes []string
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
	monitor.ItemName = productInfo.ProductName

	price, err := strconv.ParseFloat(productInfo.Price.Sales.Price, 64)
	if err != nil {
		price, err = strconv.ParseFloat(productInfo.Price.Min.Sales.Price, 64)
		if err != nil {
			price, err = strconv.ParseFloat(productInfo.Price.Max.Sales.Price, 64)
			if err != nil {
				return sizes, colors, productInfos, nil
			}
		}
	}
	monitor.Price = price

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
	monitor.ImageURL = imageURL
	// if imageURL == "" {
	// 	return sizes, colors, productInfos, nil
	// }

	if len(productInfo.Variants) == 0 {
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

	return sizes, colors, productInfos, nil
}

func (monitor *SKUMonitor) GetInStockVariations(pid string, sizes []string, colors []string) []entities.ProductInfo {
	var productInfos []entities.ProductInfo
	var combinations []Combination

	wg := sync.WaitGroup{}
	if len(sizes) > 0 && len(colors) == 0 {
		for _, size := range sizes {
			combinations = append(combinations, Combination{
				Size: size,
			})
		}
		wg.Add(len(sizes))
	}
	if len(sizes) == 0 && len(colors) > 0 {
		for _, color := range colors {
			combinations = append(combinations, Combination{
				Color: color,
			})
		}
		wg.Add(len(colors))
	}
	if len(sizes) > 0 && len(colors) > 0 {
		for _, size := range sizes {
			for _, color := range colors {
				combinations = append(combinations, Combination{
					Size:  size,
					Color: color,
				})
			}
		}
		wg.Add(len(sizes) * len(colors))
	}

	for _, combination := range combinations {
		go func(x string, y Combination) {
			productInfos = append(productInfos, monitor.GetProductInfosForCombination(x, y)...)
			wg.Done()
		}(pid, combination)
	}
	wg.Wait()
	return productInfos
}

func (monitor *SKUMonitor) GetProductInfosForCombination(pid string, combination Combination) []entities.ProductInfo {
	var productInfos []entities.ProductInfo

	endpoint := fmt.Sprintf(MonitorEndpoint2, pid, pid, combination.Color, pid, combination.Size)
	if combination.Color == "" {
		endpoint = fmt.Sprintf(MonitorEndpoint2NoColor, pid, pid, combination.Size)
	}
	if combination.Size == "" {
		endpoint = fmt.Sprintf(MonitorEndpoint2NoSize, pid, pid, combination.Color)
	}
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
			productInfos = append(productInfos, monitor.GetCombinationVariationInfo(body, pid, combination)...)
		}
	}

	return productInfos
}

func (monitor *SKUMonitor) GetCombinationVariationInfo(body, pid string, combination Combination) []entities.ProductInfo {
	var productInfos []entities.ProductInfo

	monitorResponse := MonitorResponse{}
	err := json.Unmarshal([]byte(body), &monitorResponse)
	if err != nil {
		return productInfos
	}

	productInfo := monitorResponse.Product

	price, err := strconv.ParseFloat(productInfo.Price.Sales.Price, 64)
	if err != nil {
		price, err = strconv.ParseFloat(productInfo.Price.Min.Sales.Price, 64)
		if err != nil {
			price, err = strconv.ParseFloat(productInfo.Price.Max.Sales.Price, 64)
			if err != nil {
				price = monitor.Price
			}
		}
	}

	if productInfo.Available || !monitorResponse.ATCState.IsDisabled || !monitorResponse.ATCState.IsSoldOut {
		productInfos = append(productInfos, entities.ProductInfo{
			InStock:      productInfo.Available || !monitorResponse.ATCState.IsDisabled || !monitorResponse.ATCState.IsSoldOut,
			InPriceRange: monitor.Input.MaxPrice == -1 || int(price) <= monitor.Input.MaxPrice,
			SKU:          pid,
			Price:        price,
			ItemName:     productInfo.ProductName,
			ItemURL:      BaseEndpoint + productInfo.ProductURL,
			ImageURL:     monitor.ImageURL,
			SiteSpecificInfo: structs.Map(DisneyProductInfo{
				VID:           productInfo.ID,
				Size:          combination.Size,
				Color:         combination.Color,
				QuantityLimit: productInfo.QuantityLimit,
				IsPreOrder:    productInfo.Availability.IsPreOrder,
				IsBackOrder:   productInfo.Availability.IsBackOrder,
			}),
		})
	}

	return productInfos
}
