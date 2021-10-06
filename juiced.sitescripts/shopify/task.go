package shopify

import (
	"encoding/json"
	"fmt"
	"log"
	"net/url"
	"sort"
	"strconv"
	"strings"
	"time"

	"backend.juicedbot.io/juiced.client/http"
	"backend.juicedbot.io/juiced.infrastructure/entities"
	"backend.juicedbot.io/juiced.infrastructure/enums"
	u "backend.juicedbot.io/juiced.infrastructure/util"
	"backend.juicedbot.io/juiced.sitescripts/util"
)

const MAX_RETRIES = 5

func CreateTask(input entities.TaskInput, baseTask *entities.BaseTask) (entities.RetailerTask, error) {
	shopifyTaskInput, err := ValidateTaskInput(input.SiteSpecificInfo)
	if err != nil {
		return nil, err
	}
	return &Task{
		Input:        input,
		ShopifyInput: shopifyTaskInput,
		BaseTask:     baseTask,
	}, nil
}

func (task *Task) GetSetupFunctions() []entities.TaskFunction {
	setupTaskFunctions := []entities.TaskFunction{}

	if task.ShopifyInput.SitePassword != "" {
		setupTaskFunctions = append(setupTaskFunctions, []entities.TaskFunction{
			// 1. SubmitPassword
			{
				Function:         task.SubmitPassword,
				StatusBegin:      enums.SettingUp,
				StatusPercentage: 20,
				MaxRetries:       MAX_RETRIES,
			},
		}...)
	}

	if task.ShopifyInput.TaskType == enums.TaskTypeAccount {
		if task.ShopifyInput.ShopifyRetailer == enums.GenericShopify {
			setupTaskFunctions = append(setupTaskFunctions, []entities.TaskFunction{
				// 2. Login
				{
					Function:         task.Login,
					StatusBegin:      enums.LoggingIn,
					StatusPercentage: 30,
					MaxRetries:       MAX_RETRIES,
				},
				// TODO: 3. Refresh Login
				// {
				// 	Function:        task.Login,
				// 	StatusBegin:     enums.LoggingIn,
				// 	MaxRetries:      MAX_RETRIES,
				// 	RefreshFunction: true,
				// 	RefreshEvery:    1800,
				// },
			}...)
		} else {
			switch task.ShopifyInput.ShopifyRetailer {
			case enums.MattelCreations:
				setupTaskFunctions = append(setupTaskFunctions, []entities.TaskFunction{
					// 2. Login
					{
						Function:         task.MattelCreationsLogin,
						StatusBegin:      enums.LoggingIn,
						StatusPercentage: 30,
						MaxRetries:       MAX_RETRIES,
					},
					// TODO: 3. Refresh Login
					// {
					// 	Function:        task.Login,
					// 	StatusBegin:     enums.LoggingIn,
					// 	MaxRetries:      MAX_RETRIES,
					// 	RefreshFunction: true,
					// 	RefreshEvery:    1800,
					// },
				}...)
			default:
				setupTaskFunctions = append(setupTaskFunctions, []entities.TaskFunction{
					// 2. Login
					{
						Function:         task.Login,
						StatusBegin:      enums.LoggingIn,
						StatusPercentage: 30,
						MaxRetries:       MAX_RETRIES,
					},
					// TODO: 3. Refresh Login
					// {
					// 	Function:        task.Login,
					// 	StatusBegin:     enums.LoggingIn,
					// 	MaxRetries:      MAX_RETRIES,
					// 	RefreshFunction: true,
					// 	RefreshEvery:    1800,
					// },
				}...)
			}
		}
	} else {
		if task.ShopifyInput.SitePassword == "" {
			setupTaskFunctions = append(setupTaskFunctions, []entities.TaskFunction{
				// 2. LoginGuest
				{
					Function:         task.LoginGuest,
					StatusBegin:      enums.Preloading,
					StatusPercentage: 30,
					MaxRetries:       MAX_RETRIES,
				},
				// TODO: 3. Refresh LoginGuest
				// {
				// 	Function:        task.LoginGuest,
				// 	StatusBegin:     enums.SettingUp,
				// 	MaxRetries:      MAX_RETRIES,
				// 	RefreshFunction: true,
				// 	RefreshEvery:    1800,
				// },
			}...)
		}
	}
	setupTaskFunctions = append(setupTaskFunctions, []entities.TaskFunction{
		// 4. Preload
		{
			Function:         task.Preload,
			StatusBegin:      enums.SettingUp,
			StatusPercentage: 40,
			MaxRetries:       MAX_RETRIES,
		},
		// 5. ClearCart
		{
			Function:         task.ClearCart,
			StatusBegin:      enums.ClearingCart,
			StatusPercentage: 45,
			MaxRetries:       MAX_RETRIES,
		},
	}...)

	return setupTaskFunctions
}

func (task *Task) GetMainFunctions() []entities.TaskFunction {
	mainTaskFunctions := []entities.TaskFunction{
		// 5. AddToCart
		{
			Function:         task.AddToCart,
			StatusBegin:      enums.AddingToCart,
			StatusPercentage: 55,
			MsBetweenRetries: task.Input.DelayMS,
		},
		// 6. GoToCheckout
		{
			Function:         task.GoToCheckout,
			StatusBegin:      enums.GettingCartInfo,
			StatusPercentage: 60,
			MaxRetries:       MAX_RETRIES,
		},
		// 7. HandleCheckpoint
		{
			Function:         task.HandleCheckpoint,
			StatusBegin:      enums.HandlingCheckpoint,
			StatusPercentage: 65,
			MaxRetries:       MAX_RETRIES,
		},
		// 8. HandleQueue
		{
			Function:         task.HandleQueue,
			StatusBegin:      enums.HandlingQueue,
			StatusPercentage: 70,
			MaxRetries:       MAX_RETRIES,
		},
		// 9. SubmitShippingDetails
		{
			Function:         task.SubmitShippingDetails,
			StatusBegin:      enums.SettingShippingInfo,
			StatusPercentage: 75,
			MaxRetries:       MAX_RETRIES,
		},
		// 10. SubmitShippingRate
		{
			Function:         task.SubmitShippingRate,
			StatusBegin:      enums.SettingShippingInfo,
			StatusPercentage: 80,
			MaxRetries:       MAX_RETRIES,
		},
		// 11. GetCreditID
		{
			Function:         task.GetCreditID,
			StatusBegin:      enums.GettingBillingInfo,
			StatusPercentage: 85,
			MaxRetries:       MAX_RETRIES,
		},
		// 12. SubmitBillingDetails
		{
			Function:         task.SubmitBillingDetails,
			StatusBegin:      enums.SettingBillingInfo,
			StatusPercentage: 90,
			MaxRetries:       MAX_RETRIES,
		},
		// 12. SubmitOrder
		{
			Function:         task.SubmitOrder,
			StatusBegin:      enums.CheckingOut,
			StatusPercentage: 95,
			MaxRetries:       MAX_RETRIES,
			CheckoutFunction: true,
		},
	}
	return mainTaskFunctions
}

func (task *Task) Login() (bool, string) {
	// TODO
	return true, ""
}

func (task *Task) LoginGuest() (bool, string) {
	resp, _, err := util.MakeRequest(&util.Request{
		Client: task.BaseTask.Client,
		Method: "GET",
		URL:    task.ShopifyInput.SiteURL,
		RawHeaders: http.RawHeader{
			{"sec-ch-ua", `" Not;A Brand";v="99", "Google Chrome";v="91", "Chromium";v="91"`},
			{"accept", "application/json, text/javascript, */*; q=0.01"},
			{"x-requested-with", "XMLHttpRequest"},
			{"sec-ch-ua-mobile", "?0"},
			{"user-agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/91.0.4472.106 Safari/537.36"},
			{"content-type", "application/x-www-form-urlencoded; charset=UTF-8"},
			{"sec-fetch-site", "same-origin"},
			{"sec-fetch-site", "cors"},
			{"sec-fetch-dest", "empty"},
			{"accept-encoding", "gzip, deflate"},
			{"accept-language", "en-US,en;q=0.9"},
		},
	})
	if err != nil {
		log.Println(err.Error())
		// return false, fmt.Sprintf(enums.SettingUpFailure, err.Error())
	}

	switch resp.StatusCode {
	case 200:
		return true, enums.SettingUpSuccess
	}

	return false, fmt.Sprintf(enums.SettingUpFailure, fmt.Sprintf(enums.UnknownError, resp.StatusCode))
}

func (task *Task) SubmitPassword() (bool, string) {
	params := u.CreateParams(map[string]string{
		"form_type": "storefront_password",
		"utf8":      "✓",
		"password":  task.ShopifyInput.SitePassword,
	})
	resp, _, err := util.MakeRequest(&util.Request{
		Client: task.BaseTask.Client,
		Method: "POST",
		URL:    task.ShopifyInput.SiteURL,
		RawHeaders: http.RawHeader{
			{"content-length", fmt.Sprint(len(params))},
			{"sec-ch-ua", `" Not;A Brand";v="99", "Google Chrome";v="91", "Chromium";v="91"`},
			{"accept", "application/json, text/javascript, */*; q=0.01"},
			{"x-requested-with", "XMLHttpRequest"},
			{"sec-ch-ua-mobile", "?0"},
			{"user-agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/91.0.4472.106 Safari/537.36"},
			{"content-type", "application/x-www-form-urlencoded; charset=UTF-8"},
			{"sec-fetch-site", "same-origin"},
			{"sec-fetch-site", "cors"},
			{"sec-fetch-dest", "empty"},
			{"accept-encoding", "gzip, deflate"},
			{"accept-language", "en-US,en;q=0.9"},
		},
		Data: []byte(params),
	})
	if err != nil {
		log.Println(err.Error())
		// return false, fmt.Sprintf(enums.SettingUpFailure, err.Error())
	}

	switch resp.StatusCode {
	case 200:
		return true, enums.SettingUpSuccess
	}

	return false, fmt.Sprintf(enums.SettingUpFailure, fmt.Sprintf(enums.UnknownError, resp.StatusCode))
}

func (task *Task) Preload() (bool, string) {
	productsResponse := ProductsResponse{}
	resp, _, err := util.MakeRequest(&util.Request{
		Client: task.BaseTask.Client,
		Method: "GET",
		URL:    task.ShopifyInput.SiteURL + ProductsEndpoint,
		RawHeaders: http.RawHeader{
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
		ResponseBodyStruct: &productsResponse,
	})
	if err != nil {
		log.Println(err.Error())
		// return false, fmt.Sprintf(enums.PreloadingFailure, err.Error())
	}
	switch resp.StatusCode {
	case 200:
	default:
		// Just skip preloading if it doesn't work
		return true, enums.PreloadingSuccess
	}

	var added bool
	for _, product := range productsResponse.Products {
		if added {
			break
		}
		for _, variant := range product.Variants {
			_, resp, _, err := task.AddToCartRequest(fmt.Sprint(variant.ID))
			if err != nil {
				log.Println(err.Error())
			}
			if resp.StatusCode == 200 {
				added = true
				break
			}
		}
	}

	if !added {
		// Just skip preloading if it doesn't work
		return true, enums.PreloadingSuccess
	}

	resp, body, err := task.GoToCheckoutRequest()
	if err != nil {
		log.Println(err.Error())
		// return false, fmt.Sprintf(enums.PreloadingFailure, err.Error())
	}

	switch resp.StatusCode {
	case 200:
		authToken, err := u.FindInString(body, `"authenticity_token" value="`, `"`)
		if err != nil {
			// Couldn't find authToken
			// Just skip preloading if it doesn't work
			return true, enums.PreloadingSuccess
		}
		task.AuthToken = authToken

		checkoutURL := resp.Request.URL.String()
		if strings.Contains(checkoutURL, "throttle") || strings.Contains(checkoutURL, "checkpoint") {
			// If preloading runs into queue/checkpoint, it's not worth handling now — go through ATC process first and handle after
			return true, enums.PreloadingSuccess
		}
		task.CheckoutURL = checkoutURL

		task.Preloaded = true
		return true, enums.PreloadingSuccess
	}

	// Just skip preloading if it doesn't work
	return true, enums.PreloadingSuccess
}

func (task *Task) ClearCart() (bool, string) {
	resp, _, err := util.MakeRequest(&util.Request{
		Client: task.BaseTask.Client,
		Method: "GET",
		URL:    task.ShopifyInput.SiteURL + ClearCartEndpoint,
		RawHeaders: http.RawHeader{
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
	})
	if err != nil {
		log.Println(err.Error())
		// return false, fmt.Sprintf(enums.ClearingCartFailure, err.Error())
	}

	switch resp.StatusCode {
	case 200:
		return true, enums.ClearingCartSuccess
	}

	return false, fmt.Sprintf(enums.ClearingCartFailure, fmt.Sprintf(enums.UnknownError, resp.StatusCode))
}

func (task *Task) AddToCart() (bool, string) {
	return true, ""
}

func (task *Task) GoToCheckout() (bool, string) {
	if task.Preloaded {
		return true, enums.GettingCartInfoSuccess
	}

	resp, body, err := task.GoToCheckoutRequest()
	if err != nil {
		log.Println(err.Error())
		// return false, fmt.Sprintf(enums.GettingCartInfoFailure, err.Error())
	}

	switch resp.StatusCode {
	case 200:
		authToken, err := u.FindInString(body, `"authenticity_token" value="`, `"`)
		if err != nil {
			return false, fmt.Sprintf(enums.GettingCartInfoFailure, err.Error())
		}
		task.AuthToken = authToken

		checkoutURL := resp.Request.URL.String()
		if strings.Contains(checkoutURL, "throttle") || strings.Contains(checkoutURL, "checkpoint") {
			if strings.Contains(checkoutURL, "throttle") {
				task.HasQueue = true
			}
			if strings.Contains(checkoutURL, "checkpoint") {
				task.HasCheckpoint = true
			}
			return true, enums.GettingCartInfoSuccess
		}
		task.CheckoutURL = checkoutURL
		return true, enums.GettingCartInfoSuccess
	}

	return false, fmt.Sprintf(enums.GettingCartInfoFailure, fmt.Sprintf(enums.UnknownError, resp.StatusCode))
}

func (task *Task) HandleCheckpoint() (bool, string) {
	if !task.HasCheckpoint {
		return true, enums.HandlingCheckpointSuccess
	}
	return true, enums.HandlingCheckpointSuccess
	// TODO
}

func (task *Task) HandleQueue() (bool, string) {
	if !task.HasQueue {
		return true, enums.HandlingQueueSuccess
	}
	data := []byte(u.CreateParams(map[string]string{
		"authenticity_token":   task.AuthToken,
		"g-recaptcha-response": "", // TODO
		"data_via":             "cookie",
		"commit":               "",
	}))
	resp, _, err := util.MakeRequest(&util.Request{
		Client: task.BaseTask.Client,
		Method: "POST",
		URL:    task.ShopifyInput.SiteURL + "/throttle/queue",
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
			{"referer", task.ShopifyInput.SiteURL + "/"},
			{"accept-encoding", "gzip, deflate"},
			{"accept-language", "en-US,en;q=0.9"},
		},
		Data: data,
	})
	if err != nil {
		log.Println(err.Error())
		// return false, fmt.Sprintf(enums.HandlingQueueFailure, err.Error())
	}

	switch resp.StatusCode {
	case 200:
		baseURL, err := url.Parse(task.ShopifyInput.SiteURL)
		if err != nil {
			return false, fmt.Sprintf(enums.HandlingQueueFailure, err.Error())
		}
		var currentToken string
		for _, cookie := range task.BaseTask.Client.Jar.Cookies(baseURL) {
			if cookie.Name == "_checkout_queue_token" {
				currentToken = cookie.Value
			}
		}
		if currentToken == "" {
			return false, fmt.Sprintf(enums.HandlingQueueFailure, InvalidQueueTokenError)
		}
		inQueue := true
		for inQueue {
			if task.BaseTask.StopFlag {
				return true, ""
			}
			pollRequest := PollRequest{
				Query: "\n      {\n        poll(token: $token) {\n          token\n          pollAfter\n          queueEtaSeconds\n          productVariantAvailability\n        }\n      }\n    ",
				Variables: Variables{
					Token: currentToken,
				},
			}
			data, err := json.Marshal(pollRequest)
			if err != nil {
				return false, fmt.Sprintf(enums.HandlingQueueFailure, err.Error())
			}
			pollResponse := PollResponse{}
			_, _, err = util.MakeRequest(&util.Request{
				Client: task.BaseTask.Client,
				Method: "POST",
				URL:    task.ShopifyInput.SiteURL + "/queue/poll",
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
					{"referer", task.ShopifyInput.SiteURL + "/"},
					{"accept-encoding", "gzip, deflate"},
					{"accept-language", "en-US,en;q=0.9"},
				},
				Data:               data,
				ResponseBodyStruct: &pollResponse,
			})
			if err != nil {
				log.Println(err.Error())
				// return false, fmt.Sprintf(enums.HandlingQueueFailure, err.Error())
			}

			if len(pollResponse.Data.Poll.ProductVariantAvailability) == 0 || !pollResponse.Data.Poll.ProductVariantAvailability[0].Available {
				return false, fmt.Sprintf(enums.HandlingQueueFailure, enums.ProductOutOfStockError)
			}
			currentToken = pollResponse.Data.Poll.Token
			if pollResponse.Data.Poll.Typename == "PollComplete" {
				inQueue = false
			} else {
				time.Sleep(time.Until(pollResponse.Data.Poll.PollAfter))
			}
			log.Println("Seconds left: ", pollResponse.Data.Poll.QueueEtaSeconds)
		}
	}

	return false, fmt.Sprintf(enums.HandlingQueueFailure, fmt.Sprintf(enums.UnknownError, resp.StatusCode))
}

func (task *Task) SubmitShippingDetails() (bool, string) {
	data := []byte(u.CreateParams(map[string]string{
		"_method":                                      "patch",
		"authenticity_token":                           task.AuthToken,
		"previous_step":                                "contact_information",
		"step":                                         "shipping_method",
		"checkout[email]":                              task.BaseTask.Profile.Email,
		"checkout[buyer_accepts_marketing]":            "0",
		"checkout[shipping_address][first_name]":       task.BaseTask.Profile.ShippingAddress.FirstName,
		"checkout[shipping_address][last_name]":        task.BaseTask.Profile.ShippingAddress.LastName,
		"checkout[shipping_address][address1]":         task.BaseTask.Profile.ShippingAddress.Address1,
		"checkout[shipping_address][address2]":         task.BaseTask.Profile.ShippingAddress.Address2,
		"checkout[shipping_address][city]":             task.BaseTask.Profile.ShippingAddress.City,
		"checkout[shipping_address][country]":          task.BaseTask.Profile.ShippingAddress.CountryCode,
		"checkout[shipping_address][province]":         task.BaseTask.Profile.ShippingAddress.StateCode,
		"checkout[shipping_address][zip]":              task.BaseTask.Profile.ShippingAddress.ZipCode,
		"checkout[client_details][browser_width]":      "1920",
		"checkout[client_details][browser_height]":     "1080",
		"checkout[client_details][javascript_enabled]": "1",
		"checkout[client_details][color_depth]":        "30",
		"checkout[client_details][java_enabled]":       "false",
		"checkout[client_details][browser_tz]":         "420",
	}))
	resp, _, err := util.MakeRequest(&util.Request{
		Client: task.BaseTask.Client,
		Method: "POST",
		URL:    task.CheckoutURL,
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
			{"referer", task.ShopifyInput.SiteURL + "/"},
			{"accept-encoding", "gzip, deflate"},
			{"accept-language", "en-US,en;q=0.9"},
		},
		Data: data,
	})
	if err != nil {
		log.Println(err.Error())
		// return false, fmt.Sprintf(enums.SettingShippingInfoFailure, err.Error())
	}

	switch resp.StatusCode {
	case 200:
		return true, enums.SettingShippingInfoSuccess
	}

	return false, fmt.Sprintf(enums.SettingShippingInfoFailure, fmt.Sprintf(enums.UnknownError, resp.StatusCode))
}

func (task *Task) SubmitShippingRate() (bool, string) {
	getShippingRatesResponse := GetShippingRatesResponse{}
	resp, _, err := util.MakeRequest(&util.Request{
		Client: task.BaseTask.Client,
		Method: "GET",
		URL:    task.ShopifyInput.SiteURL + fmt.Sprintf(ShippingRatesEndpoint, task.BaseTask.Profile.ShippingAddress.ZipCode, task.BaseTask.Profile.ShippingAddress.CountryCode, task.BaseTask.Profile.ShippingAddress.StateCode),
		RawHeaders: http.RawHeader{
			{"sec-ch-ua", `" Not;A Brand";v="99", "Google Chrome";v="91", "Chromium";v="91"`},
			{"accept", "application/json, text/javascript, */*; q=0.01"},
			{"x-requested-with", "XMLHttpRequest"},
			{"sec-ch-ua-mobile", "?0"},
			{"user-agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/91.0.4472.106 Safari/537.36"},
			{"origin", task.ShopifyInput.SiteURL},
			{"sec-fetch-site", "same-origin"},
			{"sec-fetch-mode", "cors"},
			{"sec-fetch-dest", "empty"},
			{"referer", task.ShopifyInput.SiteURL + "/"},
			{"accept-encoding", "gzip, deflate"},
			{"accept-language", "en-US,en;q=0.9"},
		},
		ResponseBodyStruct: &getShippingRatesResponse,
	})
	if err != nil {
		log.Println(err.Error())
		// return false, fmt.Sprintf(enums.SettingShippingInfoFailure, err.Error())
	}
	switch resp.StatusCode {
	case 200:
	default:
		return false, fmt.Sprintf(enums.SettingShippingInfoFailure, fmt.Sprintf(enums.UnknownError, resp.StatusCode))
	}

	var prices []float64
	priceToShippingRateMap := make(map[float64]ShippingRates)
	for _, shippingRate := range getShippingRatesResponse.ShippingRates {
		priceFloat, err := strconv.ParseFloat(shippingRate.Price, 64)
		if err == nil {
			priceToShippingRateMap[priceFloat] = shippingRate
			prices = append(prices, priceFloat)
		}
	}
	if len(prices) == 0 {
		return false, fmt.Sprintf(enums.SettingShippingInfoFailure, NoShippingRatesAvailableError)
	}
	prices = sort.Float64Slice(prices)

	shippingRate := priceToShippingRateMap[prices[0]].Source + "-" + priceToShippingRateMap[prices[0]].Code + "-" + priceToShippingRateMap[prices[0]].Price
	data := []byte(u.CreateParams(map[string]string{
		"_method":                     "patch",
		"authenticity_token":          task.AuthToken,
		"previous_step":               "shipping_method",
		"step":                        "payment_method",
		"checkout[shipping_rate][id]": shippingRate,
		"checkout[client_details][browser_width]":      "1920",
		"checkout[client_details][browser_height]":     "1080",
		"checkout[client_details][javascript_enabled]": "1",
		"checkout[client_details][color_depth]":        "30",
		"checkout[client_details][java_enabled]":       "false",
		"checkout[client_details][browser_tz]":         "420",
	}))
	resp, body, err := util.MakeRequest(&util.Request{
		Client: task.BaseTask.Client,
		Method: "POST",
		URL:    task.CheckoutURL,
		RawHeaders: http.RawHeader{
			//{"content-length", fmt.Sprint(len(data))},
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
			{"referer", task.ShopifyInput.SiteURL + "/"},
			{"accept-encoding", "gzip, deflate"},
			{"accept-language", "en-US,en;q=0.9"},
		},
		Data: data,
	})
	if err != nil {
		log.Println(err.Error())
		// return false, fmt.Sprintf(enums.SettingShippingInfoFailure, err.Error())
	}

	switch resp.StatusCode {
	case 200:
		paymentGateway, err := u.FindInString(body, `data-select-gateway="`, `"`)
		if err != nil {
			return false, fmt.Sprintf(enums.SettingShippingInfoFailure, err.Error())
		}
		orderTotal, err := u.FindInString(body, `order_total" : "`, `"`)
		if err != nil {
			return false, fmt.Sprintf(enums.SettingShippingInfoFailure, err.Error())
		}
		task.PaymentGateway = paymentGateway
		task.OrderTotal = orderTotal
		return true, enums.SettingShippingInfoSuccess
	}

	return false, fmt.Sprintf(enums.SettingShippingInfoFailure, fmt.Sprintf(enums.UnknownError, resp.StatusCode))
}

func (task *Task) GetCreditID() (bool, string) {
	cardSplit := SplitCard(task.BaseTask.Profile.CreditCard.CardNumber, task.BaseTask.Profile.CreditCard.CardType)
	var expMonthString string
	if string(task.BaseTask.Profile.CreditCard.ExpMonth[0]) == "0" {
		expMonthString = string(task.BaseTask.Profile.CreditCard.ExpMonth[1])
	} else {
		expMonthString = task.BaseTask.Profile.CreditCard.ExpMonth
	}
	expMonthInt, err := strconv.Atoi(expMonthString)
	if err != nil {
		return false, fmt.Sprintf(enums.GettingBillingInfoFailure, err.Error())
	}
	expYearInt, err := strconv.Atoi(task.BaseTask.Profile.CreditCard.ExpYear)
	if err != nil {
		return false, fmt.Sprintf(enums.GettingBillingInfoFailure, err.Error())
	}

	creditIDRequest := CreditIDRequest{
		CreditCard: CreditCard{
			Number:            cardSplit,
			Name:              task.BaseTask.Profile.BillingAddress.FirstName + " " + task.BaseTask.Profile.BillingAddress.LastName,
			Month:             expMonthInt,
			Year:              expYearInt,
			VerificationValue: task.BaseTask.Profile.CreditCard.CVV,
		},
		PaymentSessionScope: task.ShopifyInput.SiteURL[7:],
	}

	getCreditIDResponse := GetCreditIDResponse{}
	data, err := json.Marshal(creditIDRequest)
	if err != nil {
		return false, fmt.Sprintf(enums.GettingBillingInfoFailure, err.Error())
	}
	resp, _, err := util.MakeRequest(&util.Request{
		Client: task.BaseTask.Client,
		Method: "POST",
		URL:    CreditIDEndpoint,
		Headers: http.Header{
			"sec-ch-ua-mobile": {"?0"},
			"User-Agent":       {"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/91.0.4472.114 Safari/537.36"},
			"Content-Type":     {"application/json"},
			"Origin":           {CreditIDReferer},
			"Sec-Fetch-Site":   {"same-site"},
			"Sec-Fetch-Mode":   {"cors"},
			"Sec-Fetch-Dest":   {"empty"},
			"Referer":          {CreditIDReferer + "/"},
			"Accept-Encoding":  {"deflate, br"},
			"Accept-Language":  {"en-US,en;q=0.9"},
			"Header-Order:":    {"sec-ch-ua-mobile", "User-Agent", "Content-Type", "Origin", "Sec-Fetch-Site", "Sec-Fetch-Mode", "Sec-Fetch-Dest", "Referer", "Accept-Encoding", "Accept-Language"},
		},
		Data:               data,
		ResponseBodyStruct: &getCreditIDResponse,
	})
	if err != nil {
		log.Println(err.Error())
		// return false, fmt.Sprintf(enums.GettingBillingInfoFailure, err.Error())
	}

	switch resp.StatusCode {
	case 200:
		task.CreditID = getCreditIDResponse.ID
		return true, enums.GettingBillingInfoSuccess
	}

	return false, fmt.Sprintf(enums.GettingBillingInfoFailure, fmt.Sprintf(enums.UnknownError, resp.StatusCode))
}

func (task *Task) SubmitBillingDetails() (bool, string) {
	totalFloat, _ := strconv.ParseFloat(task.OrderTotal, 64)
	data := []byte(u.CreateParams(map[string]string{
		"_method":                             "patch",
		"authenticity_token":                  task.AuthToken,
		`checkout[reduction_code]`:            task.ShopifyInput.CouponCode,
		"previous_step":                       "payment_method",
		"step":                                "",
		"s":                                   task.CreditID,
		"checkout[payment_gateway]":           task.PaymentGateway,
		"checkout[credit_card][vault]":        "false",
		"checkout[different_billing_address]": "false",
		"checkout[total_price]":               fmt.Sprint(int(totalFloat * 100)),
		"complete":                            "1",
		"checkout[client_details][browser_width]":      "1920",
		"checkout[client_details][browser_height]":     "1080",
		"checkout[client_details][javascript_enabled]": "1",
		"checkout[client_details][color_depth]":        "30",
		"checkout[client_details][java_enabled]":       "false",
		"checkout[client_details][browser_tz]":         "420",
	}))
	resp, _, err := util.MakeRequest(&util.Request{
		Client: task.BaseTask.Client,
		Method: "POST",
		URL:    task.CheckoutURL,
		RawHeaders: http.RawHeader{
			{"sec-ch-ua", `" Not;A Brand";v="99", "Google Chrome";v="91", "Chromium";v="91"`},
			{"sec-ch-ua-mobile", "?0"},
			{"upgrade-insecure-requests", "1"},
			{"origin", task.ShopifyInput.SiteURL},
			{"content-type", "application/x-www-form-urlencoded"},
			{"user-agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/91.0.4472.106 Safari/537.36"},
			{"accept", "text/html,application/xhtml+xml,application/xml;q=0.9,image/avif,image/webp,image/apng,*/*;q=0.8,application/signed-exchange;v=b3;q=0.9"},
			{"sec-fetch-site", "same-origin"},
			{"sec-fetch-mode", "cors"},
			{"sec-fetch-dest", "empty"},
			{"referer", task.ShopifyInput.SiteURL + "/"},
			{"accept-encoding", "gzip, deflate"},
			{"accept-language", "en-US,en;q=0.9"},
		},
		Data: data,
	})
	if err != nil {
		log.Println(err.Error())
		// return false, fmt.Sprintf(enums.SettingBillingInfoFailure, err.Error())
	}

	switch resp.StatusCode {
	case 200:
		return true, enums.SettingBillingInfoSuccess
	}

	return false, fmt.Sprintf(enums.SettingBillingInfoFailure, fmt.Sprintf(enums.UnknownError, resp.StatusCode))
}

func (task *Task) SubmitOrder() (bool, string) {
	time.Sleep(3 * time.Second)
	resp, body, err := util.MakeRequest(&util.Request{
		Client: task.BaseTask.Client,
		Method: "GET",
		URL:    task.CheckoutURL + "/processing?from_processing_page=1",
		RawHeaders: http.RawHeader{
			{"sec-ch-ua", `" Not;A Brand";v="99", "Google Chrome";v="91", "Chromium";v="91"`},
			{"sec-ch-ua-mobile", "?0"},
			{"upgrade-insecure-requests", "1"},
			{"user-agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/91.0.4472.106 Safari/537.36"},
			{"accept", "text/html,application/xhtml+xml,application/xml;q=0.9,image/avif,image/webp,image/apng,*/*;q=0.8,application/signed-exchange;v=b3;q=0.9"},
			{"sec-fetch-site", "same-origin"},
			{"sec-fetch-mode", "navigate"},
			{"sec-fetch-dest", "document"},
			{"referer", task.ShopifyInput.SiteURL + "/"},
			{"accept-encoding", "gzip, deflate"},
			{"accept-language", "en-US,en;q=0.9"},
		},
	})

	if err != nil {
		log.Println(err.Error())
		// return false, fmt.Sprintf(enums.CheckingOutFailure, err.Error())
	}

	switch resp.StatusCode {
	case 200:
		if strings.Contains(body, "declined") {
			return false, enums.CardDeclined
		} else {
			return true, enums.CheckingOutSuccess
		}
	}

	return false, fmt.Sprintf(enums.CheckingOutFailure, fmt.Sprintf(enums.UnknownError, resp.StatusCode))
}
