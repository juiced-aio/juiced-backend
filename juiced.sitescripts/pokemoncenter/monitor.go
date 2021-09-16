package pokemoncenter

import (
	"encoding/json"
	"fmt"

	"backend.juicedbot.io/juiced.infrastructure/entities"
	"backend.juicedbot.io/juiced.infrastructure/enums"
	"backend.juicedbot.io/juiced.sitescripts/util"
	"github.com/anaskhan96/soup"
)

func CreateMonitor(input entities.MonitorInput, baseMonitor *entities.BaseMonitor) (entities.Monitor, error) {
	switch input.MonitorType {
	case enums.SKUMonitor:
		if err := ValidateMonitorInput(input.Input, input.MonitorType, input.SiteSpecificInfo); err != nil {
			return nil, err
		}
		return &SKUMonitor{
			Input:       input,
			BaseMonitor: baseMonitor,
		}, nil

	}
	return nil, &enums.UnsupportedMonitorTypeError{Retailer: enums.PokemonCenter, MonitorType: input.MonitorType}
}

func (monitor *SKUMonitor) GetProductInfos() ([]entities.ProductInfo, error) {
	productInfos := []entities.ProductInfo{}
	return productInfos, nil

	resp, body, err := util.MakeRequest(&util.Request{
		Client: monitor.BaseMonitor.Client,
		Method: "GET",
		URL:    fmt.Sprintf(MonitorEndpoint, monitor.Input.Input),
		RawHeaders: [][2]string{
			{"sec-ch-ua", "\" Not A;Brand\";v=\"99\", \"Chromium\";v=\"90\", \"Google Chrome\";v=\"90\""},
			{"accept", "*/*"},
			{"x-requested-with", "XMLHttpRequest"},
			{"sec-ch-ua-mobile", "?0"},
			{"user-agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/15.0 Safari/605.1.15"},
			{"sec-fetch-site", "same-origin"},
			{"sec-fetch-mode", "cors"},
			{"sec-fetch-dest", "empty"},
			{"referer", BaseEndpoint},
			{"accept-encoding", "gzip, deflate, br"},
			{"accept-language", "en-US,en;q=0.9"},
		},
	})
	if err != nil {
		return productInfos, err
	}

	switch resp.StatusCode {
	case 403:
		err = HandleDatadomeMonitor(monitor.BaseMonitor, body)
		if err != nil {
			return productInfos, err
		}
	case 200:
		monitorResponse := MonitorResponse{}
		responseBody := soup.HTMLParse(string(body))
		nextData := responseBody.Find("script", "id", "__NEXT_DATA__")
		if nextData.Error != nil {
			return productInfos, nextData.Error
		}
		nextDataString := nextData.Pointer.FirstChild.Data
		err = json.Unmarshal([]byte(nextDataString), &monitorResponse)
		if err != nil {
			return productInfos, err
		}

		productInfo := entities.ProductInfo{}
		productInfo.SKU = monitor.Input.Input
		productInfo.Price = monitorResponse.Props.InitialState.Product.ListPrice.Amount
		productInfo.ItemName = monitorResponse.Props.InitialState.Product.Name
		productInfo.ItemURL = fmt.Sprintf(MonitorEndpoint, monitor.Input.Input)
		if len(monitorResponse.Props.InitialState.Product.Images) > 0 {
			productInfo.ImageURL = monitorResponse.Props.InitialState.Product.Images[0].Original
		}

		if monitorResponse.Props.InitialState.Product.Availability == "AVAILABLE" {
			productInfo.InStock = true
		}
		if monitor.Input.MaxPrice == -1 || int(productInfo.Price) <= monitor.Input.MaxPrice {
			productInfo.InPriceRange = true
		}

		productInfos = append(productInfos, productInfo)
	}

	return productInfos, nil
}
