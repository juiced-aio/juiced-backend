package shopify

import (
	"fmt"
	"strings"

	"backend.juicedbot.io/juiced.client/http"
	u "backend.juicedbot.io/juiced.infrastructure/util"
	"backend.juicedbot.io/juiced.sitescripts/util"
)

func (task *Task) AddToCartRequest(vid string) (AddToCartResponse, *http.Response, string, error) {
	paramsString := u.CreateParams(map[string]string{
		"form_type": "product",
		"utf8":      "✓",
		"id":        vid,
		"quantity":  fmt.Sprint(task.Input.Quantity),
	})

	addToCartResponse := AddToCartResponse{}
	resp, body, err := util.MakeRequest(&util.Request{
		Client: task.BaseTask.Client,
		Method: "POST",
		URL:    task.ShopifyInput.SiteURL + AddToCartEndpoint,
		RawHeaders: http.RawHeader{
			{"content-length", fmt.Sprint(len(paramsString))},
			{"sec-ch-ua", `" Not;A Brand";v="99", "Google Chrome";v="91", "Chromium";v="91"`},
			{"accept", "application/json, text/javascript, */*; q=0.01"},
			{"x-requested-with", "XMLHttpRequest"},
			{"sec-ch-ua-mobile", "?0"},
			{"user-agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/91.0.4472.106 Safari/537.36"},
			{"content-type", "application/x-www-form-urlencoded; charset=UTF-8"},
			{"origin", task.ShopifyInput.SiteURL},
			{"sec-fetch-site", "same-origin"},
			{"sec-fetch-site", "cors"},
			{"sec-fetch-dest", "empty"},
			{"referer", task.ShopifyInput.SiteURL + "/"},
			{"accept-encoding", "gzip, deflate"},
			{"accept-language", "en-US,en;q=0.9"},
		},
		Data:               []byte(paramsString),
		ResponseBodyStruct: &addToCartResponse,
	})
	return addToCartResponse, resp, body, err
}

func (task *Task) GoToCheckoutRequest() (*http.Response, string, error) {
	data := []byte("checkout=")
	return util.MakeRequest(&util.Request{
		Client: task.BaseTask.Client,
		Method: "POST",
		URL:    task.ShopifyInput.SiteURL + CartEndpoint,
		RawHeaders: http.RawHeader{
			{"sec-ch-ua", `" Not;A Brand";v="99", "Google Chrome";v="91", "Chromium";v="91"`},
			{"accept", "text/html,application/xhtml+xml,application/xml;q=0.9,image/avif,image/webp,image/apng,*/*;q=0.8,application/signed-exchange;v=b3;q=0.9"},
			{"x-requested-with", "XMLHttpRequest"},
			{"sec-ch-ua-mobile", "?0"},
			{"origin", task.ShopifyInput.SiteURL},
			{"content-type", "application/x-www-form-urlencoded"},
			{"user-agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/91.0.4472.106 Safari/537.36"},
			{"sec-fetch-site", "same-origin"},
			{"sec-fetch-mode", "cors"},
			{"sec-fetch-dest", "empty"},
			{"referer", CartEndpoint},
			{"accept-encoding", "gzip, deflate"},
			{"accept-language", "en-US,en;q=0.9"},
		},
		Data: data,
	})
}

func SplitCard(card string, cardType string) string {
	var cardSplit string
	switch cardType {
	case "AMEX":
		cardSplit = strings.Join([]string{card[:4], card[4:10], card[10:15]}, " ")
	case "Diners":
		cardSplit = strings.Join([]string{card[:4], card[4:10], card[10:14]}, " ")
	default:
		cardSplit = strings.Join([]string{card[:4], card[4:8], card[8:12], card[12:16]}, " ")
	}
	return cardSplit
}
