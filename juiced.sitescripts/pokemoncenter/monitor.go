package pokemoncenter

import (
	"encoding/json"
	"fmt"
	"io/ioutil"

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

func (monitor *SKUMonitor) GetProductInfo() (entities.ProductInfo, error) {
	productInfo := entities.ProductInfo{}

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
		return productInfo, err
	}

	switch resp.StatusCode {
	case 403:
		err = HandleDatadome(monitor.BaseMonitor, body)
		if err != nil {
			return productInfo, err
		}
	case 200:
		monitorResponse := MonitorResponse{}
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return productInfo, err
		}

		responseBody := soup.HTMLParse(string(body))
		nextData := responseBody.Find("script", "id", "__NEXT_DATA__")
		nextDataString := nextData.Pointer.FirstChild.Data
		err = json.Unmarshal([]byte(nextDataString), &monitorResponse)
		if err != nil {
			return productInfo, err
		}

		productInfo.SKU = monitor.Input.Input
		productInfo.Price = monitorResponse.Props.InitialState.Product.ListPrice.Amount
		productInfo.ItemName = monitorResponse.Props.InitialState.Product.Name
		productInfo.ItemURL = fmt.Sprintf(MonitorEndpoint, monitor.Input.Input)
		productInfo.ImageURL = monitorResponse.Props.InitialState.Product.Images.Original

		if monitorResponse.Props.InitialState.Product.Availability == "AVAILABLE" {
			productInfo.InStock = true
		}
		if int(productInfo.Price) <= monitor.Input.MaxPrice {
			productInfo.InPriceRange = true
		}
	}

	return productInfo, nil
}
