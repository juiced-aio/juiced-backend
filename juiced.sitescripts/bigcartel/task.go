package bigcartel

import (
	"encoding/json"
	"fmt"
	"net/url"
	"time"

	"backend.juicedbot.io/juiced.client/http"
	"backend.juicedbot.io/juiced.infrastructure/common/entities"
	"backend.juicedbot.io/juiced.infrastructure/common/enums"
	"backend.juicedbot.io/juiced.infrastructure/common/events"
	"backend.juicedbot.io/juiced.sitescripts/base"
	"backend.juicedbot.io/juiced.sitescripts/util"
)

func CreateBigCartelTask(task *entities.Task, profile entities.Profile, proxy entities.Proxy, eventBus *events.EventBus, couponCode, siteURL, siteInfo SiteInfo) (Task, error) {
	bigcartelTask := Task{}
	client, err := util.CreateClient(proxy)
	if err != nil {
		return bigcartelTask, err
	}
	bigcartelTask = Task{
		Task: base.Task{
			Task:     task,
			Profile:  profile,
			Proxy:    proxy,
			EventBus: eventBus,
			Client:   client,
		},
		SiteInfo: siteInfo,
	}

	return bigcartelTask, err
}

func (task *Task) PublishEvent(status enums.TaskStatus, eventType enums.TaskEventType) {
	task.Task.Task.SetTaskStatus(status)
	task.Task.EventBus.PublishTaskEvent(status, eventType, nil, task.Task.Task.ID)
}

func (task *Task) CheckForStop() bool {
	if task.Task.StopFlag {
		task.PublishEvent(enums.TaskIdle, enums.TaskStop)
		return true
	}
	return false
}

func (task *Task) CheckForAdditionalSteps() {
	switch task.BigCartelRetailer {
	// Future Shopify site-specific functions will be added here
	}
}

// WaitForMonitor waits until the Monitor has sent the info to the task to continue
func (task *Task) WaitForMonitor() bool {
	for {
		needToStop := task.CheckForStop()
		if needToStop {
			return true
		}
		if task.InStockData.Sku != "" && task.InStockData.CartToken != "" && task.InStockData.StoreId != "" {
			return false
		}
		time.Sleep(25 * time.Millisecond)
	}
}

func (task *Task) loop() {
	check := false
	for !check {
		needToStop := task.CheckForStop()
		if needToStop {
			return
		}
		check = task.NameAndEmail()
		if !becameGuest {
			time.Sleep(time.Duration(task.Task.Task.TaskDelay) * time.Millisecond)
		}
	}

	task.Step = Preloading
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

	if task.Task.Task.TaskDelay == 0 {
		task.Task.Task.TaskDelay = 2000
	}

	task.Step = SettingUp //not srue what this is for, just adding anyway for now.

	task.CheckForAdditionalSteps()

	//First check if we need to login

	//Wait for monitor

	//Add to cart

	//Submit address and email

	//Submit payment details

	//Checkout
}

func (task *Task) AddToCart(vid string) bool {
	payload := url.Values{
		"cart[add][id]": {"268318059"},
		"submit":        {""},
	}

	addToCartResponse := AddToCartResponse{}
	resp, _, err := util.MakeRequest(&util.Request{
		Client: task.Task.Client,
		Method: "POST",
		URL:    task.SiteInfo.BaseUrl + AddToCartEndpoint,
		RawHeaders: http.RawHeader{
			{"content-length", fmt.Sprint(len(payload.Encode()))},
			{"sec-ch-ua", `" Not;A Brand";v="99", "Google Chrome";v="91", "Chromium";v="91"`},
			{"accept", "application/json, text/javascript, */*; q=0.01"},
			{"x-requested-with", "XMLHttpRequest"},
			{"sec-ch-ua-mobile", "?0"},
			{"user-agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/91.0.4472.106 Safari/537.36"},
			{"content-type", "application/x-www-form-urlencoded; charset=UTF-8"},
			{"origin", task.SiteInfo.BaseUrl},
			{"sec-fetch-site", "same-origin"},
			{"sec-fetch-site", "cors"},
			{"sec-fetch-dest", "empty"},
			{"referer", task.SiteInfo.BaseUrl + "/"},
			{"accept-encoding", "gzip, deflate"},
			{"accept-language", "en-US,en;q=0.9"},
		},
		Data:               []byte(payload.Encode()),
		ResponseBodyStruct: &addToCartResponse,
	})
	if err != nil {
		return false
	}

	switch resp.StatusCode {
	case 200:

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

func (task *Task) NameAndEmail() bool {
	payloadBytes, _ := json.Marshal(BigCartelRequestSubmitNameAndEmail{
		Buyer_email:                 "@gmail.com",
		Buyer_first_name:            "Anthony",
		Buyer_last_name:             "Reeder",
		Buyer_opted_in_to_marketing: false,
		Buyer_phone_number:          "+1 (231) 231-2312", //Number must be correct format example: +1 (231) 231-2312
	})

	addToCartResponse := AddToCartResponse{}
	resp, _, err := util.MakeRequest(&util.Request{
		Client: task.Task.Client,
		Method: "POST",
		URL:    task.SiteInfo.BaseUrl + AddressEmailEndpoint,
		RawHeaders: http.RawHeader{
			{"content-length", fmt.Sprint(len(payloadBytes))},
			{"sec-ch-ua", `" Not;A Brand";v="99", "Google Chrome";v="91", "Chromium";v="91"`},
			{"accept", "application/json, text/javascript, */*; q=0.01"},
			{"x-requested-with", "XMLHttpRequest"},
			{"sec-ch-ua-mobile", "?0"},
			{"user-agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/91.0.4472.106 Safari/537.36"},
			{"content-type", "application/x-www-form-urlencoded; charset=UTF-8"},
			{"origin", task.SiteInfo.BaseUrl},
			{"sec-fetch-site", "same-origin"},
			{"sec-fetch-site", "cors"},
			{"sec-fetch-dest", "empty"},
			{"referer", task.SiteInfo.BaseUrl + "/"},
			{"accept-encoding", "gzip, deflate"},
			{"accept-language", "en-US,en;q=0.9"},
		},
		RequestBodyStruct:  &payloadBytes,
		ResponseBodyStruct: &addToCartResponse,
	})
	if err != nil {
		return false
	}

	switch resp.StatusCode {
	case 200:

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

func (task *Task) Address(vid string) bool {
	payloadBytes, _ := json.Marshal(BigCartelRequestSubmitAddress{
		Shipping_address_1:             "49 Thackeray Close",
		Shipping_address_2:             "",
		Shipping_city:                  "Royston",
		Shipping_country_autofill_name: "",
		Shipping_country_id:            "43",
		Shipping_state:                 "Hawaii",
		Shipping_zip:                   "4353453453",
	})

	resp, _, err := util.MakeRequest(&util.Request{
		Client: task.Task.Client,
		Method: "POST",
		URL:    task.SiteInfo.BaseUrl + AddressEmailEndpoint,
		RawHeaders: http.RawHeader{
			{"content-length", fmt.Sprint(len(payloadBytes))},
			{"sec-ch-ua", `" Not;A Brand";v="99", "Google Chrome";v="91", "Chromium";v="91"`},
			{"accept", "application/json, text/javascript, */*; q=0.01"},
			{"x-requested-with", "XMLHttpRequest"},
			{"sec-ch-ua-mobile", "?0"},
			{"user-agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/91.0.4472.106 Safari/537.36"},
			{"content-type", "application/x-www-form-urlencoded; charset=UTF-8"},
			{"origin", task.SiteInfo.BaseUrl},
			{"sec-fetch-site", "same-origin"},
			{"sec-fetch-site", "cors"},
			{"sec-fetch-dest", "empty"},
			{"referer", task.SiteInfo.BaseUrl + "/"},
			{"accept-encoding", "gzip, deflate"},
			{"accept-language", "en-US,en;q=0.9"},
		},
		RequestBodyStruct: &payloadBytes,
	})
	if err != nil {
		return false
	}

	switch resp.StatusCode {
	case 200:

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

func (task *Task) PaymentMethod(cartToken, paymentId string) bool {
	payload := "{\"cart_token\":\"" + cartToken + "\",\"stripe_payment_method_id\":\"" + paymentId + "\",\"stripe_payment_intent_id\":null}"

	resp, _, err := util.MakeRequest(&util.Request{
		Client: task.Task.Client,
		Method: "POST",
		URL:    PaymentMethodEndpoint,
		RawHeaders: http.RawHeader{
			{"content-length", fmt.Sprint(len(payload))},
			{"sec-ch-ua", `" Not;A Brand";v="99", "Google Chrome";v="91", "Chromium";v="91"`},
			{"accept", "application/json, text/javascript, */*; q=0.01"},
			{"x-requested-with", "XMLHttpRequest"},
			{"sec-ch-ua-mobile", "?0"},
			{"user-agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/91.0.4472.106 Safari/537.36"},
			{"content-type", "application/x-www-form-urlencoded; charset=UTF-8"},
			{"origin", task.SiteInfo.BaseUrl},
			{"sec-fetch-site", "same-origin"},
			{"sec-fetch-site", "cors"},
			{"sec-fetch-dest", "empty"},
			{"referer", task.SiteInfo.BaseUrl + "/"},
			{"accept-encoding", "gzip, deflate"},
			{"accept-language", "en-US,en;q=0.9"},
		},
		Data: []byte(payload),
	})
	if err != nil {
		return false
	}

	switch resp.StatusCode {
	case 200:

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

func (task *Task) PaymentInfo(cartToken, paymentId string) bool {
	payload := url.Values{
		"type":                                  {"card"},
		"billing_details[name]":                 {"Anthony Reeder"},
		"billing_details[address][line1]":       {"49 Thackeray Close"},
		"billing_details[address][line2]":       {""},
		"billing_details[address][city]":        {"Royston"},
		"billing_details[address][state]":       {"Hawaii"},
		"billing_details[address][postal_code]": {"4353453453"},
		"billing_details[address][country]":     {"US"},
		"card[number]":                          {"4767718212263745"},
		"card[cvc]":                             {"260"},
		"card[exp_month]":                       {"02"},
		"card[exp_year]":                        {"26"},
		"guid":                                  {"13bedf9f-0e16-4243-9024-2eecbc91113ee4bb60"},
		"muid":                                  {"f5023d3c-907a-42db-8dc4-d4ef7f2413436b0869"},
		"sid":                                   {"39c026e7-5d19-4c2f-93f5-058749b4de5fa54e58"},
		"pasted_fields":                         {"number"},
		"payment_user_agent":                    {"stripe.js/76aee18e6; stripe-js-v3/76aee18e6"},
		"time_on_page":                          {"13709"},
		"referrer":                              {"https://checkout.bigcartel.com/"},
		"key":                                   {"pk_live_HAopYDMYyyhaXP505VRbXQtT"}, //i think this is speicfic to each site.
	}
	resp, _, err := util.MakeRequest(&util.Request{
		Client: task.Task.Client,
		Method: "POST",
		URL:    PaymentInfoEndpoint,
		RawHeaders: http.RawHeader{
			{"content-length", fmt.Sprint(len(payload.Encode()))},
			{"sec-ch-ua", `" Not;A Brand";v="99", "Google Chrome";v="91", "Chromium";v="91"`},
			{"accept", "application/json, text/javascript, */*; q=0.01"},
			{"x-requested-with", "XMLHttpRequest"},
			{"sec-ch-ua-mobile", "?0"},
			{"user-agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/91.0.4472.106 Safari/537.36"},
			{"content-type", "application/x-www-form-urlencoded; charset=UTF-8"},
			{"origin", task.SiteInfo.BaseUrl},
			{"sec-fetch-site", "same-origin"},
			{"sec-fetch-site", "cors"},
			{"sec-fetch-dest", "empty"},
			{"referer", task.SiteInfo.BaseUrl + "/"},
			{"accept-encoding", "gzip, deflate"},
			{"accept-language", "en-US,en;q=0.9"},
		},
		Data: []byte(payload.Encode()),
	})
	if err != nil {
		return false
	}

	switch resp.StatusCode {
	case 200:

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

func (task *Task) Checkout(vid string) bool {
	payloadBytes, _ := json.Marshal(Payment{
		Stripe_payment_method_id: "",
	})

	resp, _, err := util.MakeRequest(&util.Request{
		Client: task.Task.Client,
		Method: "POST",
		URL:    PaymentInfoEndpoint,
		RawHeaders: http.RawHeader{
			{"content-length", fmt.Sprint(len(payloadBytes))},
			{"sec-ch-ua", `" Not;A Brand";v="99", "Google Chrome";v="91", "Chromium";v="91"`},
			{"accept", "application/json, text/javascript, */*; q=0.01"},
			{"x-requested-with", "XMLHttpRequest"},
			{"sec-ch-ua-mobile", "?0"},
			{"user-agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/91.0.4472.106 Safari/537.36"},
			{"content-type", "application/x-www-form-urlencoded; charset=UTF-8"},
			{"origin", task.SiteInfo.BaseUrl},
			{"sec-fetch-site", "same-origin"},
			{"sec-fetch-site", "cors"},
			{"sec-fetch-dest", "empty"},
			{"referer", task.SiteInfo.BaseUrl + "/"},
			{"accept-encoding", "gzip, deflate"},
			{"accept-language", "en-US,en;q=0.9"},
		},
		RequestBodyStruct: payloadBytes,
	})
	if err != nil {
		return false
	}

	switch resp.StatusCode {
	case 200:

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
