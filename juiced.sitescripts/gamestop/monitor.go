package gamestop

import (
	"errors"
	"fmt"
	"log"
	"strconv"
	"strings"

	"backend.juicedbot.io/juiced.antibot/akamai"
	"backend.juicedbot.io/juiced.infrastructure/entities"
	"backend.juicedbot.io/juiced.infrastructure/enums"
	"backend.juicedbot.io/juiced.sitescripts/util"
	"github.com/fatih/structs"
)

func CreateMonitor(input entities.MonitorInput, baseMonitor *entities.BaseMonitor) (entities.Monitor, error) {
	switch input.MonitorType {
	case enums.SKUMonitor:
		gameStopMonitorInput, err := ValidateMonitorInput(input.Input, input.MonitorType, input.SiteSpecificInfo)
		if err != nil {
			return nil, err
		}
		return &SKUMonitor{
			Input:         input,
			GameStopInput: gameStopMonitorInput,
			BaseMonitor:   baseMonitor,
		}, nil

	}
	return nil, &enums.UnsupportedMonitorTypeError{Retailer: enums.GameStop, MonitorType: input.MonitorType}
}

func (monitor *SKUMonitor) GetProductInfos() ([]entities.ProductInfo, error) {
	productInfos := []entities.ProductInfo{}

	monitorResponse := MonitorResponse{}
	resp, _, err := util.MakeRequest(&util.Request{
		Client: monitor.BaseMonitor.Client,
		Method: "GET",
		URL:    fmt.Sprintf(MonitorEndpoint, monitor.Input.Input),
		RawHeaders: [][2]string{
			{"sec-ch-ua", "\" Not A;Brand\";v=\"99\", \"Chromium\";v=\"90\", \"Google Chrome\";v=\"90\""},
			{"accept", "*/*"},
			{"x-requested-with", "XMLHttpRequest"},
			{"sec-ch-ua-mobile", "?0"},
			{"user-agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/90.0.4430.212 Safari/537.36"},
			{"sec-fetch-site", "same-origin"},
			{"sec-fetch-mode", "cors"},
			{"sec-fetch-dest", "empty"},
			{"referer", BaseEndpoint},
			{"accept-encoding", "gzip, deflate, br"},
			{"accept-language", "en-US,en;q=0.9"},
		},
		ResponseBodyStruct: &monitorResponse,
	})
	if err != nil {
		// return productInfos, err
		log.Println(err.Error())
	}

	switch resp.StatusCode {
	case 200:
		price, err := strconv.ParseFloat(monitorResponse.GtmData.Price.SellingPrice, 64)
		if err != nil {
			return productInfos, errors.New("bad price")
		}
		imageURL := ""
		if len(monitorResponse.MccEvents) > 0 && len(monitorResponse.MccEvents[0]) > 1 {
			for _, event := range monitorResponse.MccEvents[0][1].([]interface{}) {
				imageURL = fmt.Sprint(event.(map[string]interface{})["image_url"])
			}
		}
		itemName := monitorResponse.GtmData.ProductInfo.Name
		itemURL := ""
		if strings.Contains(monitorResponse.Product.SelectedProductURL, "?") {
			itemURL = BaseEndpoint + strings.Split(monitorResponse.Product.SelectedProductURL, "?")[0]
		}
		inStock := monitorResponse.GtmData.ProductInfo.Availability == "Available" || (monitorResponse.Product.Availability.ButtonText == "Pre-Order" && monitorResponse.Product.Available)
		inPriceRange := monitor.Input.MaxPrice == -1 || int(price) <= monitor.Input.MaxPrice
		sku := monitor.Input.Input
		vid := monitorResponse.GtmData.ProductInfo.SKU
		maxQuantity := monitorResponse.Product.MaxOrderQuantity

		productInfos = append(productInfos, entities.ProductInfo{
			InStock:      inStock,
			InPriceRange: inPriceRange,
			SKU:          sku,
			Price:        price,
			ItemName:     itemName,
			ItemURL:      itemURL,
			ImageURL:     imageURL,
			SiteSpecificInfo: structs.Map(GameStopProductInfo{
				VID:         vid,
				Size:        "",
				Color:       "",
				Condition:   "",
				MaxQuantity: maxQuantity,
			}),
		})
		return productInfos, nil
	case 403:
		err = akamai.HandleAkamaiMonitor(monitor.BaseMonitor, BaseEndpoint+"/", BaseEndpoint, AkamaiEndpoint, BaseURL)
		if err != nil {
			log.Println("AH1", err.Error())
			return productInfos, err
		}
	case 404:
		return productInfos, errors.New("404 product not found")
	}

	return productInfos, errors.New("unknown error")
}
