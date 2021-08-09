package bigcartel

import (
	"encoding/json"
	"fmt"
	"log"
	"net/url"
	"time"

	"backend.juicedbot.io/juiced.client/http"
	"backend.juicedbot.io/juiced.infrastructure/common/entities"
	"backend.juicedbot.io/juiced.infrastructure/common/enums"
	"backend.juicedbot.io/juiced.infrastructure/common/events"
	"backend.juicedbot.io/juiced.sitescripts/base"
	"backend.juicedbot.io/juiced.sitescripts/util"
)

func CreateBigCartelTask(task *entities.Task, profile entities.Profile, proxy entities.Proxy, eventBus *events.EventBus, siteUrl string) (Task, error) {
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
		SiteURL: siteUrl,
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
	// Future BigCartel site-specific functions will be added here
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

func (task *Task) RunTask() {
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

	task.Step = SettingUp

	task.MaxRetry = 5
	task.CheckForAdditionalSteps()

	task.Step = WaitingForMonitor

	// WaitForMonitor
	needToStop := task.WaitForMonitor()
	if needToStop {
		return
	}

	task.Step = AddingToCart

	startTime := time.Now()

	task.PublishEvent(enums.SettingEmailAddress, enums.TaskUpdate)
	for isSuccess, needtostop := task.RunUntilSuccessful(task.NameAndEmail()); !isSuccess || needtostop; {
		if needtostop {
			return
		}
	}

	task.PublishEvent(enums.SettingShippingInfo, enums.TaskUpdate)
	for isSuccess, needtostop := task.RunUntilSuccessful(task.Address()); !isSuccess || needtostop; {
		if needtostop {
			return
		}
	}

	task.PublishEvent(enums.GettingBillingInfo, enums.TaskUpdate)
	for isSuccess, needtostop := task.RunUntilSuccessful(task.PaymentMethod()); !isSuccess || needtostop; {
		if needtostop {
			return
		}
	}

	task.PublishEvent(enums.SettingBillingInfo, enums.TaskUpdate)
	for isSuccess, needtostop := task.RunUntilSuccessful(task.PaymentInfo()); !isSuccess || needtostop; {
		if needtostop {
			return
		}
	}

	task.PublishEvent(enums.CheckingOut, enums.TaskUpdate)
	for isSuccess, needtostop := task.RunUntilSuccessful(task.Checkout()); !isSuccess || needtostop; {
		if needtostop {
			return
		}
	}

	endTime := time.Now()

	log.Println("STARTED AT: " + startTime.String())
	log.Println("  ENDED AT: " + endTime.String())
	log.Println("TIME TO CHECK OUT: ", endTime.Sub(startTime).Milliseconds())
}

func (task *Task) NameAndEmail() (bool, string) {
	payloadBytes, _ := json.Marshal(BigCartelRequestSubmitNameAndEmail{
		Buyer_email:                 task.Task.Profile.Email,
		Buyer_first_name:            task.Task.Profile.ShippingAddress.FirstName,
		Buyer_last_name:             task.Task.Profile.ShippingAddress.LastName,
		Buyer_opted_in_to_marketing: false,
		Buyer_phone_number:          task.Task.Profile.PhoneNumber, //Number must be correct format example: +1 (231) 231-2312
	})

	resp, _, err := util.MakeRequest(&util.Request{
		Client: task.Task.Client,
		Method: "POST",
		URL:    task.SiteURL + AddressEmailEndpoint,
		RawHeaders: http.RawHeader{
			{"content-length", fmt.Sprint(len(payloadBytes))},
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
		RequestBodyStruct: &payloadBytes,
	})
	if err != nil {
		return false, enums.SettingEmailAddressFailure
	}

	switch resp.StatusCode {
	case 200:
		return true, enums.SettingEmailAddressSuccess
	case 422:
		return false, enums.SettingEmailAddressFailure
	case 404:
		return false, enums.SettingEmailAddressFailure
	default:
		return false, enums.SettingEmailAddressFailure
	}
}

func (task *Task) Address() (bool, string) {
	payloadBytes, _ := json.Marshal(BigCartelRequestSubmitAddress{
		Shipping_address_1:             task.Task.Profile.ShippingAddress.Address1,
		Shipping_address_2:             task.Task.Profile.ShippingAddress.Address2,
		Shipping_city:                  task.Task.Profile.ShippingAddress.City,
		Shipping_country_autofill_name: "",
		Shipping_country_id:            "43", //43 = USA i assume we arnt supporting anywhere else anyway? If so we'll need to set them up specificaly.
		Shipping_state:                 task.Task.Profile.ShippingAddress.StateCode,
		Shipping_zip:                   task.Task.Profile.ShippingAddress.ZipCode,
	})

	resp, _, err := util.MakeRequest(&util.Request{
		Client: task.Task.Client,
		Method: "POST",
		URL:    task.SiteURL + AddressEmailEndpoint,
		RawHeaders: http.RawHeader{
			{"content-length", fmt.Sprint(len(payloadBytes))},
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
		RequestBodyStruct: &payloadBytes,
	})
	if err != nil {
		return false, ""
	}

	switch resp.StatusCode {
	case 200:

		return true, ""
	case 422:
		return false, ""
	case 404:
		return false, ""
	default:
		return false, ""
	}
}

func (task *Task) PaymentMethod() (bool, string) {
	payload := "{\"cart_token\":\"" + task.InStockData.CartToken + "\",\"stripe_payment_method_id\":\"" + task.InStockData.StoreId + "\",\"stripe_payment_intent_id\":null}"

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
			{"sec-fetch-site", "same-origin"},
			{"sec-fetch-site", "cors"},
			{"sec-fetch-dest", "empty"},
			{"accept-encoding", "gzip, deflate"},
			{"accept-language", "en-US,en;q=0.9"},
		},
		Data: []byte(payload),
	})
	if err != nil {
		return false, ""
	}

	switch resp.StatusCode {
	case 200:
		return true, ""
	case 422:
		return false, ""
	case 404:
		return false, ""
	default:
		return false, ""
	}
}

func (task *Task) PaymentInfo() (bool, string) {
	payload := url.Values{
		"type":                                  {"card"},
		"billing_details[name]":                 {task.Task.Profile.BillingAddress.FirstName + task.Task.Profile.BillingAddress.LastName},
		"billing_details[address][line1]":       {task.Task.Profile.BillingAddress.Address1},
		"billing_details[address][line2]":       {task.Task.Profile.BillingAddress.Address2},
		"billing_details[address][city]":        {task.Task.Profile.BillingAddress.City},
		"billing_details[address][state]":       {task.Task.Profile.BillingAddress.StateCode},
		"billing_details[address][postal_code]": {task.Task.Profile.BillingAddress.ZipCode},
		"billing_details[address][country]":     {"US"},
		"card[number]":                          {task.Task.Profile.CreditCard.CardNumber},
		"card[cvc]":                             {task.Task.Profile.CreditCard.CVV},
		"card[exp_month]":                       {task.Task.Profile.CreditCard.ExpMonth},
		"card[exp_year]":                        {task.Task.Profile.CreditCard.ExpYear},
		"pasted_fields":                         {"number"},
		"time_on_page":                          {"13709"}, ///this time seems fine? Maybe we could randomise it slightly? doesnt seem important.
		"referrer":                              {"https://checkout.bigcartel.com/"},
		"key":                                   {"pk_live_HAopYDMYyyhaXP505VRbXQtT"}, //Must get from the checkout page (currently done on monitor)
	}
	resp, _, err := util.MakeRequest(&util.Request{
		Client: task.Task.Client,
		Method: "POST",
		URL:    task.SiteURL + PaymentInfoEndpoint,
		RawHeaders: http.RawHeader{
			{"content-length", fmt.Sprint(len(payload.Encode()))},
			{"sec-ch-ua", `" Not;A Brand";v="99", "Google Chrome";v="91", "Chromium";v="91"`},
			{"accept", "application/json, text/javascript, */*; q=0.01"},
			{"x-requested-with", "XMLHttpRequest"},
			{"sec-ch-ua-mobile", "?0"},
			{"user-agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/91.0.4472.106 Safari/537.36"},
			{"content-type", "application/x-www-form-urlencoded; charset=UTF-8"},
			{"origin", "https://js.stripe.com"},
			{"sec-fetch-site", "same-origin"},
			{"sec-fetch-site", "cors"},
			{"sec-fetch-dest", "empty"},
			{"referer", "https://js.stripe.com"},
			{"accept-encoding", "gzip, deflate"},
			{"accept-language", "en-US,en;q=0.9"},
		},
		Data: []byte(payload.Encode()),
	})
	if err != nil {
		return false, ""
	}

	switch resp.StatusCode {
	case 200:
		return true, ""
	case 422:
		return false, ""
	case 404:
		return false, ""
	default:
		return false, ""
	}
}

func (task *Task) Checkout() (bool, string) {
	payloadBytes, _ := json.Marshal(Payment{
		Stripe_payment_method_id: "",
	})

	resp, _, err := util.MakeRequest(&util.Request{
		Client: task.Task.Client,
		Method: "POST",
		URL:    task.SiteURL + PaymentInfoEndpoint,
		RawHeaders: http.RawHeader{
			{"content-length", fmt.Sprint(len(payloadBytes))},
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
		RequestBodyStruct: payloadBytes,
	})
	if err != nil {
		return false, ""
	}

	switch resp.StatusCode {
	case 200:

		return true, ""
	case 422:
		return false, ""
	case 404:
		return false, ""
	default:
		return false, ""
	}
}
