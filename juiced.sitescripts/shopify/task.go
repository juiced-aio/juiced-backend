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
	"backend.juicedbot.io/juiced.infrastructure/common"
	"backend.juicedbot.io/juiced.infrastructure/common/entities"
	"backend.juicedbot.io/juiced.infrastructure/common/enums"
	"backend.juicedbot.io/juiced.infrastructure/common/events"
	"backend.juicedbot.io/juiced.infrastructure/queries"
	"backend.juicedbot.io/juiced.sitescripts/base"
	"backend.juicedbot.io/juiced.sitescripts/util"
)

func CreateShopifyTask(task *entities.Task, profile entities.Profile, proxy entities.Proxy, eventBus *events.EventBus, siteURL string, email string, password string) (Task, error) {
	shopifyTask := Task{}
	client, err := util.CreateClient(proxy)
	if err != nil {
		return shopifyTask, err
	}
	shopifyTask = Task{
		Task: base.Task{
			Task:     task,
			Profile:  profile,
			Proxy:    proxy,
			EventBus: eventBus,
			Client:   client,
		},
		AccountInfo: AccountInfo{
			Email:    email,
			Password: password,
		},
		SiteURL: siteURL,
	}

	return shopifyTask, err
}

// PublishEvent wraps the EventBus's PublishTaskEvent function
func (task *Task) PublishEvent(status enums.TaskStatus, eventType enums.TaskEventType) {
	task.Task.Task.SetTaskStatus(status)
	task.Task.EventBus.PublishTaskEvent(status, eventType, nil, task.Task.Task.ID)
}

// CheckForStop checks the stop flag and stops the monitor if it's true
func (task *Task) CheckForStop() bool {
	if task.Task.StopFlag {
		task.PublishEvent(enums.TaskIdle, enums.TaskStop)
		return true
	}
	return false
}

func (task *Task) CheckForAdditionalSteps() {
	switch task.ShopifyRetailer {
	case enums.HotWheels:
		if task.Step == SettingUp {
			task.HotWheelsLogin()
		}
		// Future Shopify site-specific functions will be added here
	}
}

func (task *Task) RunTask() {
	// If the function panics due to a runtime error, recover from it
	defer func() {
		if recover() != nil {
			task.Task.StopFlag = true
			task.PublishEvent(enums.TaskIdle, enums.TaskFail)
		}
		task.PublishEvent(enums.TaskIdle, enums.TaskComplete)
	}()

	task.Step = SettingUp

	task.CheckForAdditionalSteps()

	task.Step = WaitingForMonitor

	// 1. WaitForMonitor
	task.PublishEvent(enums.WaitingForMonitor, enums.TaskUpdate)
	needToStop := task.WaitForMonitor()
	if needToStop {
		return
	}

	task.Step = AddingToCart

	startTime := time.Now()

	task.CheckForAdditionalSteps()

	task.PublishEvent(enums.AddingToCart, enums.TaskUpdate)
	// 2. AddtoCart
	addedToCart := false
	for !addedToCart {
		needToStop := task.CheckForStop()
		if needToStop {
			return
		}
		addedToCart = task.AddToCart()
		if !addedToCart {
			time.Sleep(time.Duration(task.Task.Task.TaskDelay) * time.Millisecond)
		}
	}

	task.PublishEvent(enums.CheckingOut, enums.TaskUpdate)
	// 3. Checkout
	checkout := false
	for !addedToCart {
		needToStop := task.CheckForStop()
		if needToStop {
			return
		}
		checkout = task.Checkout()
		if !checkout {
			time.Sleep(time.Duration(task.Task.Task.TaskDelay) * time.Millisecond)
		}
	}

	task.PublishEvent(enums.SettingShippingInfo, enums.TaskUpdate)
	// 4. SetShipping
	setShippingInfo := false
	for !setShippingInfo {
		needToStop := task.CheckForStop()
		if needToStop {
			return
		}
		setShippingInfo = task.SetShippingInfo()
		if !setShippingInfo {
			time.Sleep(time.Duration(task.Task.Task.TaskDelay) * time.Millisecond)
		}
	}

	setShippingRate := false
	for !setShippingInfo {
		needToStop := task.CheckForStop()
		if needToStop {
			return
		}
		setShippingRate = task.SetShippingRate()
		if !setShippingRate {
			time.Sleep(time.Duration(task.Task.Task.TaskDelay) * time.Millisecond)
		}
	}

	task.PublishEvent(enums.SettingBillingInfo, enums.TaskUpdate)
	// 5. SetPayment
	getCreditID := false
	for !getCreditID {
		needToStop := task.CheckForStop()
		if needToStop {
			return
		}
		getCreditID = task.GetCreditID()
		if !getCreditID {
			time.Sleep(time.Duration(task.Task.Task.TaskDelay) * time.Millisecond)
		}
	}

	setPaymentInfo := false
	for !setPaymentInfo {
		needToStop := task.CheckForStop()
		if needToStop {
			return
		}
		setPaymentInfo = task.SetPaymentInfo()
		if !setPaymentInfo {
			time.Sleep(time.Duration(task.Task.Task.TaskDelay) * time.Millisecond)
		}
	}

	task.PublishEvent(enums.CheckingOut, enums.TaskUpdate)
	// 6. PlaceOrder
	processOrder := false
	var status enums.OrderStatus
	for !processOrder {
		needToStop := task.CheckForStop()
		if needToStop {
			return
		}
		if status == enums.OrderStatusDeclined {
			break
		}
		processOrder, status = task.ProcessOrder(startTime)
		if !processOrder {
			time.Sleep(time.Duration(task.Task.Task.TaskDelay) * time.Millisecond)
		}
	}

	endTime := time.Now()

	log.Println("STARTED AT: " + startTime.String())
	log.Println("  ENDED AT: " + endTime.String())
	log.Println("TIME TO CHECK OUT: ", endTime.Sub(startTime).Milliseconds())
	switch status {
	case enums.OrderStatusSuccess:
		task.PublishEvent(enums.CheckedOut, enums.TaskComplete)
	case enums.OrderStatusDeclined:
		task.PublishEvent(enums.CheckoutFailed, enums.TaskComplete)
	case enums.OrderStatusFailed:
		task.PublishEvent(enums.CheckoutFailed, enums.TaskComplete)
	}

}

// WaitForMonitor waits until the Monitor has sent the info to the task to continue
func (task *Task) WaitForMonitor() bool {
	for {
		needToStop := task.CheckForStop()
		if needToStop {
			return true
		}
		if task.InStockData.VariantID != "" {
			time.Sleep(1 * time.Millisecond)
			return false
		}
		task.VariantID = task.InStockData.VariantID
	}
}

func (task *Task) AddToCart() bool {
	paramsString := util.CreateParams(map[string]string{
		"form_type": "product",
		"utf8":      "✓",
		"id":        task.VariantID,
		"quantity":  "1",
	})

	addToCartResponse := AddToCartResponse{}
	resp, _, err := util.MakeRequest(&util.Request{
		Client: task.Client,
		Method: "POST",
		URL:    AddToCartEndpoint,
		RawHeaders: http.RawHeader{
			{"content-length", fmt.Sprint(len(paramsString))},
			{"sec-ch-ua", `" Not;A Brand";v="99", "Google Chrome";v="91", "Chromium";v="91"`},
			{"accept", "application/json, text/javascript, */*; q=0.01"},
			{"x-requested-with", "XMLHttpRequest"},
			{"sec-ch-ua-mobile", "?0"},
			{"user-agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/91.0.4472.106 Safari/537.36"},
			{"content-type", "application/x-www-form-urlencoded; charset=UTF-8"},
			{"origin", task.SiteURL},
			{"sec-fetch-site", "same-origin"},
			{"sec-fetch-site", "cors"},
			{"sec-fetch-dest", "empty"},
			{"referer", task.SiteURL + "/"},
			{"accept-encoding", "gzip, deflate"},
			{"accept-language", "en-US,en;q=0.9"},
		},
		Data:               []byte(paramsString),
		ResponseBodyStruct: &addToCartResponse,
	})
	if err != nil {
		fmt.Println("err")
	}

	switch resp.StatusCode {
	case 200:
		task.TaskInfo.Image = addToCartResponse.Image
		task.TaskInfo.Name = addToCartResponse.Title
		task.TaskInfo.ItemURL = addToCartResponse.URL
		task.TaskInfo.Price = addToCartResponse.Price
		return true
	case 422:
		//Out Of Stock
		return false
	case 404:
		//Item does not exist
		return false
	default:
		return false
	}

}

func (task *Task) Checkout() bool {
	data := []byte("checkout=")
	resp, body, err := util.MakeRequest(&util.Request{
		Client: task.Client,
		Method: "POST",
		URL:    CartEndpoint,
		RawHeaders: http.RawHeader{
			{"sec-ch-ua", `" Not;A Brand";v="99", "Google Chrome";v="91", "Chromium";v="91"`},
			{"accept", "text/html,application/xhtml+xml,application/xml;q=0.9,image/avif,image/webp,image/apng,*/*;q=0.8,application/signed-exchange;v=b3;q=0.9"},
			{"x-requested-with", "XMLHttpRequest"},
			{"sec-ch-ua-mobile", "?0"},
			{"origin", task.SiteURL},
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
	if err != nil {
		fmt.Println(err)
	}

	switch resp.StatusCode {
	case 200:
		AuthToken, err := common.FindInString(body, `"authenticity_token" value="`, `"`)
		if err != nil {
			//Couldn't find auth
			return false
		}
		task.TaskInfo.AuthToken = AuthToken
		checkoutURL := resp.Request.URL.String()
		if strings.Contains(checkoutURL, "throttle") {
			return task.HandleQueue()
		}

		if strings.Contains(checkoutURL, "checkpoint") {
			return task.HandleQueue()
		}

		task.TaskInfo.CheckoutURL = checkoutURL
		return true
	default:
		return false
	}
}

func (task *Task) HandleQueue() bool {
	data := []byte(util.CreateParams(map[string]string{
		"authenticity_token": task.TaskInfo.AuthToken,
		// Empty for now
		"g-recaptcha-response": "",
		"data_via":             "cookie",
		"commit":               "",
	}))
	resp, _, err := util.MakeRequest(&util.Request{
		Client: task.Client,
		Method: "POST",
		URL:    task.SiteURL + "/throttle/queue",
		RawHeaders: http.RawHeader{
			{"sec-ch-ua", `" Not;A Brand";v="99", "Google Chrome";v="91", "Chromium";v="91"`},
			{"accept", "text/html,application/xhtml+xml,application/xml;q=0.9,image/avif,image/webp,image/apng,*/*;q=0.8,application/signed-exchange;v=b3;q=0.9"},
			{"x-requested-with", "XMLHttpRequest"},
			{"sec-ch-ua-mobile", "?0"},
			{"origin", task.SiteURL},
			{"content-type", "application/x-www-form-urlencoded"},
			{"user-agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/91.0.4472.106 Safari/537.36"},
			{"sec-fetch-site", "same-origin"},
			{"sec-fetch-mode", "cors"},
			{"sec-fetch-dest", "empty"},
			{"referer", task.SiteURL + "/"},
			{"accept-encoding", "gzip, deflate"},
			{"accept-language", "en-US,en;q=0.9"},
		},
		Data: data,
	})
	if err != nil {
		fmt.Println(err)
	}

	if resp.StatusCode != 200 {
		return false
	}

	parsedBase, _ := url.Parse(task.SiteURL)
	var currentToken string
	for _, cookie := range task.Client.Jar.Cookies(parsedBase) {
		if cookie.Name == "_checkout_queue_token" {
			currentToken = cookie.Value
		}
	}

	inQueue := true
	for inQueue {
		pollRequest := PollRequest{
			Query: "\n      {\n        poll(token: $token) {\n          token\n          pollAfter\n          queueEtaSeconds\n          productVariantAvailability\n        }\n      }\n    ",
			Variables: Variables{
				Token: currentToken,
			},
		}
		data, _ := json.Marshal(pollRequest)
		pollResponse := PollResponse{}
		_, _, err := util.MakeRequest(&util.Request{
			Client: task.Client,
			Method: "POST",
			URL:    task.SiteURL + "/queue/poll",
			RawHeaders: http.RawHeader{
				{"sec-ch-ua", `" Not;A Brand";v="99", "Google Chrome";v="91", "Chromium";v="91"`},
				{"accept", "text/html,application/xhtml+xml,application/xml;q=0.9,image/avif,image/webp,image/apng,*/*;q=0.8,application/signed-exchange;v=b3;q=0.9"},
				{"x-requested-with", "XMLHttpRequest"},
				{"sec-ch-ua-mobile", "?0"},
				{"origin", task.SiteURL},
				{"content-type", "application/x-www-form-urlencoded"},
				{"user-agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/91.0.4472.106 Safari/537.36"},
				{"sec-fetch-site", "same-origin"},
				{"sec-fetch-mode", "cors"},
				{"sec-fetch-dest", "empty"},
				{"referer", task.SiteURL + "/"},
				{"accept-encoding", "gzip, deflate"},
				{"accept-language", "en-US,en;q=0.9"},
			},
			Data:               data,
			ResponseBodyStruct: &pollResponse,
		})
		if err != nil {
			fmt.Println(err)
		}

		if !pollResponse.Data.Poll.Productvariantavailability[0].Available {
			return false
		}

		currentToken = pollResponse.Data.Poll.Token

		if pollResponse.Data.Poll.Typename == "PollComplete" {
			inQueue = false
		} else {
			time.Sleep(1 * time.Second)
		}

	}

	return task.Checkout()
}

func (task *Task) SetShippingInfo() bool {
	data := []byte(util.CreateParams(map[string]string{
		"_method":                                      "patch",
		"authenticity_token":                           task.TaskInfo.AuthToken,
		"previous_step":                                "contact_information",
		"step":                                         "shipping_method",
		"checkout[email]":                              task.Task.Profile.Email,
		"checkout[buyer_accepts_marketing]":            "0",
		"checkout[shipping_address][first_name]":       task.Task.Profile.ShippingAddress.FirstName,
		"checkout[shipping_address][last_name]":        task.Task.Profile.ShippingAddress.LastName,
		"checkout[shipping_address][address1]":         task.Task.Profile.ShippingAddress.Address1,
		"checkout[shipping_address][address2]":         task.Task.Profile.ShippingAddress.Address2,
		"checkout[shipping_address][city]":             task.Task.Profile.ShippingAddress.City,
		"checkout[shipping_address][country]":          task.Task.Profile.ShippingAddress.CountryCode,
		"checkout[shipping_address][province]":         task.Task.Profile.ShippingAddress.StateCode,
		"checkout[shipping_address][zip]":              task.Task.Profile.ShippingAddress.ZipCode,
		"checkout[client_details][browser_width]":      "1920",
		"checkout[client_details][browser_height]":     "1080",
		"checkout[client_details][javascript_enabled]": "1",
		"checkout[client_details][color_depth]":        "30",
		"checkout[client_details][java_enabled]":       "false",
		"checkout[client_details][browser_tz]":         "420",
	}))
	resp, _, err := util.MakeRequest(&util.Request{
		Client: task.Client,
		Method: "POST",
		URL:    task.TaskInfo.CheckoutURL,
		RawHeaders: http.RawHeader{
			{"sec-ch-ua", `" Not;A Brand";v="99", "Google Chrome";v="91", "Chromium";v="91"`},
			{"accept", "text/html,application/xhtml+xml,application/xml;q=0.9,image/avif,image/webp,image/apng,*/*;q=0.8,application/signed-exchange;v=b3;q=0.9"},
			{"x-requested-with", "XMLHttpRequest"},
			{"sec-ch-ua-mobile", "?0"},
			{"origin", task.SiteURL},
			{"content-type", "application/x-www-form-urlencoded"},
			{"user-agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/91.0.4472.106 Safari/537.36"},
			{"sec-fetch-site", "same-origin"},
			{"sec-fetch-mode", "cors"},
			{"sec-fetch-dest", "empty"},
			{"referer", task.SiteURL + "/"},
			{"accept-encoding", "gzip, deflate"},
			{"accept-language", "en-US,en;q=0.9"},
		},
		Data: data,
	})
	if err != nil {
		fmt.Println(err)
	}

	return resp.StatusCode == 200
}

func (task *Task) SetShippingRate() bool {
	shippingRatesResponse := ShippingRatesResponse{}
	resp, _, err := util.MakeRequest(&util.Request{
		Client: task.Client,
		Method: "GET",
		URL:    fmt.Sprintf(ShippingRatesEndpoint, task.Task.Profile.ShippingAddress.ZipCode, task.Task.Profile.ShippingAddress.CountryCode, task.Task.Profile.ShippingAddress.StateCode),
		RawHeaders: http.RawHeader{
			{"sec-ch-ua", `" Not;A Brand";v="99", "Google Chrome";v="91", "Chromium";v="91"`},
			{"accept", "application/json, text/javascript, */*; q=0.01"},
			{"x-requested-with", "XMLHttpRequest"},
			{"sec-ch-ua-mobile", "?0"},
			{"user-agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/91.0.4472.106 Safari/537.36"},
			{"origin", task.SiteURL},
			{"sec-fetch-site", "same-origin"},
			{"sec-fetch-mode", "cors"},
			{"sec-fetch-dest", "empty"},
			{"referer", task.SiteURL + "/"},
			{"accept-encoding", "gzip, deflate"},
			{"accept-language", "en-US,en;q=0.9"},
		},
		ResponseBodyStruct: &shippingRatesResponse,
	})
	if err != nil {
		fmt.Println(err)
	}
	if resp.StatusCode != 200 {
		return false
	}
	var prices []float64

	shippingRateUnderPrice := make(map[float64]ShippingRates)
	for _, shippingRate := range shippingRatesResponse.ShippingRates {
		priceFloat, _ := strconv.ParseFloat(shippingRate.Price, 64)
		shippingRateUnderPrice[priceFloat] = shippingRate
		prices = append(prices, priceFloat)
	}
	prices = sort.Float64Slice(prices)

	task.TaskInfo.ShippingRate = shippingRateUnderPrice[prices[0]].Source + "-" + shippingRateUnderPrice[prices[0]].Code + "-" + shippingRateUnderPrice[prices[0]].Price

	data := []byte(util.CreateParams(map[string]string{
		"_method":                     "patch",
		"authenticity_token":          task.TaskInfo.AuthToken,
		"previous_step":               "shipping_method",
		"step":                        "payment_method",
		"checkout[shipping_rate][id]": task.TaskInfo.ShippingRate,
		"checkout[client_details][browser_width]":      "1920",
		"checkout[client_details][browser_height]":     "1080",
		"checkout[client_details][javascript_enabled]": "1",
		"checkout[client_details][color_depth]":        "30",
		"checkout[client_details][java_enabled]":       "false",
		"checkout[client_details][browser_tz]":         "420",
	}))
	resp, body, err := util.MakeRequest(&util.Request{
		Client: task.Client,
		Method: "POST",
		URL:    task.TaskInfo.CheckoutURL,
		RawHeaders: http.RawHeader{
			//{"content-length", fmt.Sprint(len(data))},
			{"sec-ch-ua", `" Not;A Brand";v="99", "Google Chrome";v="91", "Chromium";v="91"`},
			{"accept", "text/html,application/xhtml+xml,application/xml;q=0.9,image/avif,image/webp,image/apng,*/*;q=0.8,application/signed-exchange;v=b3;q=0.9"},
			{"x-requested-with", "XMLHttpRequest"},
			{"sec-ch-ua-mobile", "?0"},
			{"origin", task.SiteURL},
			{"content-type", "application/x-www-form-urlencoded"},
			{"user-agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/91.0.4472.106 Safari/537.36"},
			{"sec-fetch-site", "same-origin"},
			{"sec-fetch-mode", "cors"},
			{"sec-fetch-dest", "empty"},
			{"referer", task.SiteURL + "/"},
			{"accept-encoding", "gzip, deflate"},
			{"accept-language", "en-US,en;q=0.9"},
		},
		Data: data,
	})
	if err != nil {
		fmt.Println(err)
	}

	switch resp.StatusCode {
	case 200:
		paymentGateway, err := common.FindInString(body, `data-select-gateway="`, `"`)
		if err != nil {
			//Could not find Payment Gateway
			return false
		}
		orderTotal, err := common.FindInString(body, `order_total" : "`, `"`)
		if err != nil {
			//Could not find Order Total
			return false
		}
		task.TaskInfo.PaymentGateway = paymentGateway
		task.TaskInfo.OrderTotal = orderTotal
		return true
	default:
		return false
	}
}

func (task *Task) GetCreditID() bool {
	cardSplit := SplitCard(task.Task.Profile.CreditCard.CardNumber)
	var expMonthString string
	if string(task.Task.Profile.CreditCard.ExpMonth[0]) == "0" {
		expMonthString = string(task.Task.Profile.CreditCard.ExpMonth[1])
	} else {
		expMonthString = task.Task.Profile.CreditCard.ExpMonth
	}
	expMonthInt, _ := strconv.Atoi(expMonthString)
	expYearInt, _ := strconv.Atoi(task.Task.Profile.CreditCard.ExpYear)

	creditIDRequest := CreditIDRequest{
		CreditCard: CreditCard{
			Number:            cardSplit[0] + " " + cardSplit[1] + " " + cardSplit[2] + " " + cardSplit[3],
			Name:              task.Task.Profile.BillingAddress.FirstName + " " + task.Task.Profile.BillingAddress.LastName,
			Month:             expMonthInt,
			Year:              expYearInt,
			VerificationValue: task.Task.Profile.CreditCard.CVV,
		},
		PaymentSessionScope: task.SiteURL[7:],
	}

	creditIDResponse := CreditIDResponse{}
	data, _ := json.Marshal(creditIDRequest)
	resp, _, err := util.MakeRequest(&util.Request{
		Client: task.Client,
		Method: "POST",
		URL:    CreditIDEndpoint,
		Headers: http.Header{
			"sec-ch-ua-mobile": {"?0"},
			"User-Agent":       {"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/91.0.4472.114 Safari/537.36"},
			"Content-Type":     {"application/json"},
			"Origin":           {"https://checkout.shopifycs.com/"},
			"Sec-Fetch-Site":   {"same-site"},
			"Sec-Fetch-Mode":   {"cors"},
			"Sec-Fetch-Dest":   {"empty"},
			"Referer":          {"https://checkout.shopifycs.com/"},
			"Accept-Encoding":  {"deflate, br"},
			"Accept-Language":  {"en-US,en;q=0.9"},
			"Header-Order:":    {"sec-ch-ua-mobile", "User-Agent", "Content-Type", "Origin", "Sec-Fetch-Site", "Sec-Fetch-Mode", "Sec-Fetch-Dest", "Referer", "Accept-Encoding", "Accept-Language"},
		},
		Data:               data,
		ResponseBodyStruct: &creditIDResponse,
	})
	if err != nil {
		fmt.Println(err)
	}
	task.TaskInfo.CreditID = creditIDResponse.ID

	return resp.StatusCode == 200
}

func (task *Task) SetPaymentInfo() bool {
	totalFloat, _ := strconv.ParseFloat(task.TaskInfo.OrderTotal, 64)
	data := []byte(util.CreateParams(map[string]string{
		"_method":                             "patch",
		"authenticity_token":                  task.TaskInfo.AuthToken,
		"previous_step":                       "payment_method",
		"step":                                "",
		"s":                                   task.TaskInfo.CreditID,
		"checkout[payment_gateway]":           task.TaskInfo.PaymentGateway,
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
		Client: task.Client,
		Method: "POST",
		URL:    task.TaskInfo.CheckoutURL,
		RawHeaders: http.RawHeader{
			{"sec-ch-ua", `" Not;A Brand";v="99", "Google Chrome";v="91", "Chromium";v="91"`},
			{"sec-ch-ua-mobile", "?0"},
			{"upgrade-insecure-requests", "1"},
			{"origin", task.SiteURL},
			{"content-type", "application/x-www-form-urlencoded"},
			{"user-agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/91.0.4472.106 Safari/537.36"},
			{"accept", "text/html,application/xhtml+xml,application/xml;q=0.9,image/avif,image/webp,image/apng,*/*;q=0.8,application/signed-exchange;v=b3;q=0.9"},
			{"sec-fetch-site", "same-origin"},
			{"sec-fetch-mode", "cors"},
			{"sec-fetch-dest", "empty"},
			{"referer", task.SiteURL + "/"},
			{"accept-encoding", "gzip, deflate"},
			{"accept-language", "en-US,en;q=0.9"},
		},
		Data: data,
	})
	if err != nil {
		fmt.Println(err)
	}

	return resp.StatusCode == 200
}

func (task *Task) ProcessOrder(startTime time.Time) (bool, enums.OrderStatus) {
	var status enums.OrderStatus

	time.Sleep(3 * time.Second)
	resp, body, err := util.MakeRequest(&util.Request{
		Client: task.Client,
		Method: "GET",
		URL:    task.TaskInfo.CheckoutURL + "/processing?from_processing_page=1",
		RawHeaders: http.RawHeader{
			{"sec-ch-ua", `" Not;A Brand";v="99", "Google Chrome";v="91", "Chromium";v="91"`},
			{"sec-ch-ua-mobile", "?0"},
			{"upgrade-insecure-requests", "1"},
			{"user-agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/91.0.4472.106 Safari/537.36"},
			{"accept", "text/html,application/xhtml+xml,application/xml;q=0.9,image/avif,image/webp,image/apng,*/*;q=0.8,application/signed-exchange;v=b3;q=0.9"},
			{"sec-fetch-site", "same-origin"},
			{"sec-fetch-mode", "navigate"},
			{"sec-fetch-dest", "document"},
			{"referer", task.SiteURL + "/"},
			{"accept-encoding", "gzip, deflate"},
			{"accept-language", "en-US,en;q=0.9"},
		},
	})
	if err != nil {
		fmt.Println(err)
	}

	var success bool
	// I have absolutely no idea what happens when the order actually goes through, I'll use my card and hopefully will be able to cancel
	if resp.StatusCode == 200 {
		if strings.Contains(body, "declined") {
			success = false
			status = enums.OrderStatusDeclined
		} else {
			success = true
			status = enums.OrderStatusSuccess
		}
	}

	_, user, err := queries.GetUserInfo()
	if err != nil {
		fmt.Println("Could not get user info")
		return success, status
	}

	util.ProcessCheckout(util.ProcessCheckoutInfo{
		BaseTask:     task.Task,
		Success:      success,
		Content:      "",
		Embeds:       task.CreateShopifyEmbed(status, task.TaskInfo.Image),
		UserInfo:     user,
		ItemName:     task.TaskInfo.Name,
		Sku:          task.VariantID,
		Retailer:     enums.Shopify,
		Price:        float64(task.TaskInfo.Price),
		Quantity:     1,
		MsToCheckout: time.Since(startTime).Milliseconds(),
	})

	return success, status
}
