package shopify

import (
	"errors"
	"fmt"
	"log"

	"backend.juicedbot.io/juiced.client/http"
	"backend.juicedbot.io/juiced.infrastructure/entities"
	"backend.juicedbot.io/juiced.infrastructure/enums"
	u "backend.juicedbot.io/juiced.infrastructure/util"
	"backend.juicedbot.io/juiced.sitescripts/util"
	"github.com/fatih/structs"
)

func CreateMonitor(input entities.MonitorInput, baseMonitor *entities.BaseMonitor) (entities.Monitor, error) {
	switch input.MonitorType {
	case enums.SKUMonitor:
		shopifyMonitorInput, err := ValidateMonitorInput(input.Input, input.MonitorType, input.SiteSpecificInfo)
		if err != nil {
			return nil, err
		}
		return &SKUMonitor{
			Input:        input,
			ShopifyInput: shopifyMonitorInput,
			BaseMonitor:  baseMonitor,
		}, nil

	case enums.URLMonitor:
		shopifyMonitorInput, err := ValidateMonitorInput(input.Input, input.MonitorType, input.SiteSpecificInfo)
		if err != nil {
			return nil, err
		}
		return &URLMonitor{
			Input:        input,
			ShopifyInput: shopifyMonitorInput,
			BaseMonitor:  baseMonitor,
		}, nil

	}
	return nil, &enums.UnsupportedMonitorTypeError{Retailer: enums.HotTopic, MonitorType: input.MonitorType}
}

func (monitor *SKUMonitor) GetProductInfos() ([]entities.ProductInfo, error) {
	return monitor.GetVIDStock(monitor.Input.Input)
}

func (monitor *SKUMonitor) GetVIDStock(vid string) ([]entities.ProductInfo, error) {
	var productInfos []entities.ProductInfo

	params := u.CreateParams(map[string]string{
		"form_type": "product",
		"utf8":      "✓",
		"id":        vid,
		"quantity":  "1",
	})
	monitorResponse := AddToCartResponse{}
	resp, body, err := util.MakeRequest(&util.Request{
		Client: monitor.BaseMonitor.Client,
		Method: "POST",
		URL:    monitor.ShopifyInput.SiteURL + AddToCartEndpoint,
		RawHeaders: http.RawHeader{
			{"content-length", fmt.Sprint(len(params))},
			{"sec-ch-ua", `" Not;A Brand";v="99", "Google Chrome";v="91", "Chromium";v="91"`},
			{"accept", "application/json, text/javascript, */*; q=0.01"},
			{"x-requested-with", "XMLHttpRequest"},
			{"sec-ch-ua-mobile", "?0"},
			{"user-agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/91.0.4472.106 Safari/537.36"},
			{"content-type", "application/x-www-form-urlencoded; charset=UTF-8"},
			{"origin", monitor.ShopifyInput.SiteURL},
			{"sec-fetch-site", "same-origin"},
			{"sec-fetch-site", "cors"},
			{"sec-fetch-dest", "empty"},
			{"referer", monitor.ShopifyInput.SiteURL + "/"},
			{"accept-encoding", "gzip, deflate"},
			{"accept-language", "en-US,en;q=0.9"},
		},
		Data:               []byte(params),
		ResponseBodyStruct: &monitorResponse,
	})
	if err != nil {
		log.Println(err.Error())
		log.Println(resp.Status)
		log.Println(body)
		return productInfos, err
	}

	switch resp.StatusCode {
	case 200:
		price := float64(monitorResponse.Price) / 100
		productInfos = append(productInfos, entities.ProductInfo{
			InStock:      true,
			InPriceRange: monitor.Input.MaxPrice == -1 || int(price) <= monitor.Input.MaxPrice,
			SKU:          monitorResponse.SKU,
			Price:        price,
			ItemName:     monitorResponse.Title,
			ItemURL:      monitorResponse.URL,
			ImageURL:     monitorResponse.Image,
			SiteSpecificInfo: structs.Map(ShopifyProductInfo{
				VID: fmt.Sprint(monitorResponse.VariantID),
			}),
		})
		return productInfos, nil
	case 404:
		return productInfos, errors.New("404 product not found")
	}

	log.Println(resp.Status)
	log.Println(body)

	return productInfos, errors.New("unknown error")
}

func (monitor *URLMonitor) GetProductInfos() ([]entities.ProductInfo, error) {
	return monitor.GetURLStock(monitor.Input.Input)
}

func (monitor *URLMonitor) GetURLStock(url string) ([]entities.ProductInfo, error) {
	var productInfos []entities.ProductInfo

	return productInfos, nil
}
