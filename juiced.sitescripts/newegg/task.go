package newegg

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"log"
	"net/url"
	"time"

	"backend.juicedbot.io/juiced.client/http"
	"backend.juicedbot.io/juiced.infrastructure/common"
	"backend.juicedbot.io/juiced.infrastructure/common/entities"
	"backend.juicedbot.io/juiced.infrastructure/common/enums"
	"backend.juicedbot.io/juiced.infrastructure/common/events"
	"backend.juicedbot.io/juiced.sitescripts/base"
	"backend.juicedbot.io/juiced.sitescripts/util"
)

// CreateNeweggTask takes a Task entity and turns it into a Newegg Task
func CreateNeweggTask(task *entities.Task, profile entities.Profile, proxyGroup *entities.ProxyGroup, eventBus *events.EventBus) (Task, error) {
	neweggTask := Task{
		Task: base.Task{
			Task:       task,
			Profile:    profile,
			ProxyGroup: proxyGroup,
			EventBus:   eventBus,
		},
	}
	return neweggTask, nil
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

// RunTask is the script driver that calls all the individual requests
// Function order:
//		1. BecomeGuest
//		2. WaitForMonitor
//		3. AddTocart
//		4. ProceedToCheckout
//		5. Checkout
//		6. SubmitShippingInfo
//		7. GetPaymentToken
//		8. SubmitPaymentInfo
//		9. InitOrder
//		10. PlaceOrder
//		11. Verify
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
	if task.Task.Task.TaskQty == 0 {
		task.Task.Task.TaskQty = 1
	}

	err := task.Task.CreateClient(task.Task.Proxy)
	if err != nil {
		return
	}

	task.PublishEvent(enums.SettingUp, enums.TaskStart)
	// 1. BecomeGuest
	becameGuest := false
	for !becameGuest {
		needToStop := task.CheckForStop()
		if needToStop {
			return
		}
		becameGuest = BecomeGuest(task.Task.Client)
		if !becameGuest {
			time.Sleep(time.Duration(task.Task.Task.TaskDelay) * time.Millisecond)
		}
	}

	// 2. WaitForMonitor
	task.PublishEvent(enums.WaitingForMonitor, enums.TaskUpdate)
	needToStop := task.WaitForMonitor()
	if needToStop {
		return
	}

	// 3. AddTocart
	task.PublishEvent(enums.AddingToCart, enums.TaskUpdate)
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

	startTime := time.Now()

	// 4. ProceedToCheckout
	task.PublishEvent(enums.SettingCartInfo, enums.TaskUpdate)
	preparedCheckout := false
	for !preparedCheckout {
		needToStop := task.CheckForStop()
		if needToStop {
			return
		}
		preparedCheckout = task.PrepareCheckout()
		if !preparedCheckout {
			time.Sleep(time.Duration(task.Task.Task.TaskDelay) * time.Millisecond)
		}
	}

	// 5. Checkout
	task.PublishEvent(enums.GettingCartInfo, enums.TaskUpdate)
	gotCheckout := false
	for !gotCheckout {
		needToStop := task.CheckForStop()
		if needToStop {
			return
		}
		gotCheckout = task.Checkout()
		if !gotCheckout {
			time.Sleep(time.Duration(task.Task.Task.TaskDelay) * time.Millisecond)
		}
	}

	// 6. SubmitShippingInfo
	task.PublishEvent(enums.SettingShippingInfo, enums.TaskUpdate)
	submittedShipping := false
	for !submittedShipping {
		needToStop := task.CheckForStop()
		if needToStop {
			return
		}
		submittedShipping = task.SubmitShippingInfo()
		if !submittedShipping {
			time.Sleep(time.Duration(task.Task.Task.TaskDelay) * time.Millisecond)
		}
	}

	// 7. GetPaymentToken
	task.PublishEvent(enums.SettingBillingInfo, enums.TaskUpdate)
	gotPaymentToken := false
	for !gotPaymentToken {
		needToStop := task.CheckForStop()
		if needToStop {
			return
		}
		gotPaymentToken = task.GetPaymentToken()
		if !gotPaymentToken {
			time.Sleep(time.Duration(task.Task.Task.TaskDelay) * time.Millisecond)
		}
	}

	// 8. SubmitPaymentInfo
	submittedPayment := false
	for !submittedPayment {
		needToStop := task.CheckForStop()
		if needToStop {
			return
		}
		submittedPayment = task.SubmitPaymentInfo()
		if !submittedPayment {
			time.Sleep(time.Duration(task.Task.Task.TaskDelay) * time.Millisecond)
		}
	}

	// 9. InitOrder
	task.PublishEvent(enums.CheckingOut, enums.TaskUpdate)
	initiatedOrder := false
	for !initiatedOrder {
		needToStop := task.CheckForStop()
		if needToStop {
			return
		}

		initiatedOrder = task.InitOrder()
		if !initiatedOrder {
			time.Sleep(time.Duration(task.Task.Task.TaskDelay) * time.Millisecond)
		}
	}

	// 10. PlaceOrder
	task.PublishEvent(enums.CheckingOut, enums.TaskUpdate)
	submittedOrder := false
	status := enums.OrderStatusFailed
	for !submittedOrder {
		needToStop := task.CheckForStop()
		if needToStop {
			return
		}
		if status == enums.OrderStatusDeclined {
			break
		}
		submittedOrder, status = task.PlaceOrder(startTime)
		if !submittedOrder {
			time.Sleep(time.Duration(task.Task.Task.TaskDelay) * time.Millisecond)
		}
	}

	// 11. Verify
	go task.Verify()

	endTime := time.Now()

	log.Println("STARTED AT: " + startTime.String())
	log.Println("  ENDED AT: " + endTime.String())
	log.Println("TIME TO CHECK OUT: ", endTime.Sub(startTime).Milliseconds())

	if status == enums.OrderStatusSuccess {
		task.PublishEvent(enums.CheckedOut, enums.TaskComplete)
	} else {
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
		if task.StockData.ItemNumber != "" {
			return false
		}
		time.Sleep(common.MS_TO_WAIT)
	}
}

func (task *Task) AddToCart() bool {
	addToCartRequest := AddToCartRequest{
		Itemlist: []Itemlist{
			{
				Itemgroup:     "Single",
				Itemnumber:    task.StockData.ItemNumber,
				Quantity:      task.Task.Task.TaskQty,
				Optionalinfos: nil,
				Saletype:      "Sales",
			},
		},
		Customernumber: 0,
	}

	task.Task.Client.Jar.SetCookies(ParsedBase, []*http.Cookie{{
		Name:   "NV%5FW57",
		Value:  "USA",
		Domain: ".newegg.com",
		Path:   "/",
	}})
	respMap := make(map[string]interface{})
	resp, _, err := util.MakeRequest(&util.Request{
		Client: task.Task.Client,
		Method: "POST",
		URL:    AddToCartEndpoint,
		RawHeaders: http.RawHeader{
			{"pragma", `no-cache`},
			{"cache-control", `no-cache`},
			{"sec-ch-ua", `"Chromium";v="92", " Not A;Brand";v="99", "Google Chrome";v="92"`},
			{"accept", `application/json, text/plain, */*`},
			{"sec-ch-ua-mobile", `?0`},
			{"user-agent", `Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/92.0.4515.131 Safari/537.36`},
			{"content-type", `application/json`},
			{"origin", BaseEndpoint},
			{"sec-fetch-site", `same-origin`},
			{"sec-fetch-mode", `cors`},
			{"sec-fetch-dest", `empty`},
			{"referer", task.StockData.ItemURL},
			{"accept-encoding", `gzip, deflate, br`},
			{"accept-language", `en-US,en;q=0.9`},
		},

		RequestBodyStruct:  addToCartRequest,
		ResponseBodyStruct: &respMap,
	})
	if err != nil || resp.StatusCode != 201 {
		fmt.Println(err)
		return false
	}

	return true
}

func (task *Task) PrepareCheckout() bool {
	prepareCheckoutRequest := PrepareCheckoutRequest{
		Itemlist: []Itemlist{
			{
				Itemnumber: task.StockData.ItemNumber,
				ItemKey:    base64.StdEncoding.EncodeToString([]byte(fmt.Sprintf(`{"SaleType":1,"ItemGroup":1,"ItemNumber":"%v","OptionalInfos":[]}`, task.StockData.ItemNumber))),
				Quantity:   task.Task.Task.TaskQty,
				Itemgroup:  "Single",
			},
		},
		Actions: []interface{}{},
	}

	data, _ := json.Marshal(prepareCheckoutRequest)

	params, newSign := CreateExtras()
	if params == "" || newSign == "" {
		return false
	}

	parsedURL, _ := url.Parse(SecureBaseEndpoint)
	task.Task.Client.Jar.SetCookies(parsedURL, []*http.Cookie{{
		Name:   "NV%5FW57",
		Value:  "USA",
		Domain: ".newegg.com",
		Path:   "/",
	}})

	respMap := make(map[string]interface{})
	resp, _, err := util.MakeRequest(&util.Request{
		Client: task.Task.Client,
		Method: "POST",
		URL:    PrepareCheckoutEndpoint + "?" + params,
		RawHeaders: http.RawHeader{
			{"sec-ch-ua", `"Chromium";v="92", " Not A;Brand";v="99", "Google Chrome";v="92"`},
			{"sec-ch-ua-mobile", `?0`},
			{"user-agent", `Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/92.0.4515.131 Safari/537.36`},
			{"content-type", `application/json`},
			{"accept", `application/json, text/plain, */*`},
			{"x-ne-sign", newSign},
			{"x-requested-with", `XMLHttpRequest`},
			{"x-ne-sign-type", `simple`},
			{"origin", SecureBaseEndpoint},
			{"sec-fetch-site", `same-origin`},
			{"sec-fetch-mode", `cors`},
			{"sec-fetch-dest", `empty`},
			{"referer", CartEndpoint},
			{"accept-encoding", `gzip, deflate, br`},
			{"accept-language", `en-US,en;q=0.9`},
		},
		Data:               data,
		ResponseBodyStruct: &respMap,
	})
	if err != nil || resp.StatusCode != 200 {
		fmt.Println(err)
		return false
	}

	var sessionID string
	var ok bool
	if sessionID, ok = respMap["SessionID"].(string); !ok {
		return false
	}

	task.TaskInfo.SessionID = sessionID

	// Auth

	authCheckoutRequest := AuthCheckoutRequest{
		Nextpage:      fmt.Sprintf(GuestCheckoutEndpoint, task.TaskInfo.SessionID),
		Sitetheme:     "Normal",
		Sessionid:     task.TaskInfo.SessionID,
		Lastloginname: "",
		Callbackpage:  "https://secure.newegg.com/login/landingpage",
	}
	data, _ = json.Marshal(authCheckoutRequest)

	params, newSign = CreateExtras()
	if params == "" || newSign == "" {
		return false
	}

	resp, _, err = util.MakeRequest(&util.Request{
		Client: task.Task.Client,
		Method: "POST",
		URL:    AuthCheckoutEndpoint + "?" + params,
		RawHeaders: http.RawHeader{
			{"sec-ch-ua", `"Chromium";v="92", " Not A;Brand";v="99", "Google Chrome";v="92"`},
			{"sec-ch-ua-mobile", `?0`},
			{"user-agent", `Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/92.0.4515.131 Safari/537.36`},
			{"content-type", `application/json`},
			{"accept", `application/json, text/plain, */*`},
			{"x-ne-sign", newSign},
			{"x-requested-with", `XMLHttpRequest`},
			{"x-ne-sign-type", `simple`},
			{"origin", SecureBaseEndpoint},
			{"sec-fetch-site", `same-origin`},
			{"sec-fetch-mode", `cors`},
			{"sec-fetch-dest", `empty`},
			{"referer", CartEndpoint},
			{"accept-encoding", `gzip, deflate, br`},
			{"accept-language", `en-US,en;q=0.9`},
		},
		Data: data,
	})
	if err != nil || resp.StatusCode != 200 {
		fmt.Println(err)
		return false
	}

	return true

}

func (task *Task) Checkout() bool {
	resp, _, err := util.MakeRequest(&util.Request{
		Client: task.Task.Client,
		Method: "GET",
		URL:    fmt.Sprintf(GuestCheckoutEndpoint, task.TaskInfo.SessionID),
		RawHeaders: http.RawHeader{
			{"sec-ch-ua", `"Chromium";v="92", " Not A;Brand";v="99", "Google Chrome";v="92"`},
			{"sec-ch-ua-mobile", `?0`},
			{"upgrade-insecure-requests", `1`},
			{"user-agent", `Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/92.0.4515.131 Safari/537.36`},
			{"accept", `text/html,application/xhtml+xml,application/xml;q=0.9,image/avif,image/webp,image/apng,*/*;q=0.8,application/signed-exchange;v=b3;q=0.9`},
			{"sec-fetch-site", `none`},
			{"sec-fetch-mode", `navigate`},
			{"sec-fetch-user", `?1`},
			{"sec-fetch-dest", `document`},
			{"accept-encoding", `gzip, deflate, br`},
			{"accept-language", `en-US,en;q=0.9`},
		},
	})
	if resp.StatusCode != 200 || err != nil {
		return false
	}

	return true
}

func (task *Task) SubmitShippingInfo() bool {
	submitShippingInfoRequest := SubmitShippingInfoRequest{
		Detailinfo: Detailinfo{
			Contactwith:       task.Task.Profile.ShippingAddress.FirstName + " " + task.Task.Profile.ShippingAddress.LastName,
			Phone:             task.Task.Profile.PhoneNumber,
			Country:           task.Task.Profile.ShippingAddress.CountryCode,
			State:             task.Task.Profile.ShippingAddress.StateCode,
			City:              task.Task.Profile.ShippingAddress.City,
			Address1:          task.Task.Profile.ShippingAddress.Address1,
			Address2:          task.Task.Profile.ShippingAddress.Address2,
			Zipcode:           task.Task.Profile.ShippingAddress.ZipCode,
			Displaylines:      []interface{}{},
			Email:             task.Task.Profile.Email,
			Firstname:         task.Task.Profile.ShippingAddress.FirstName,
			Lastname:          task.Task.Profile.ShippingAddress.LastName,
			Addressverifymark: "None",
		},
		Sessionid: task.TaskInfo.SessionID,
		Isguest:   true,
	}

	params, newSign := CreateExtras()
	if params == "" || newSign == "" {
		return false
	}

	respMap := make(map[string]interface{})
	resp, _, err := util.MakeRequest(&util.Request{
		Client: task.Task.Client,
		Method: "POST",
		URL:    SubmitShippingInfoEndpoint + "?" + params,
		RawHeaders: http.RawHeader{
			{"sec-ch-ua", `"Chromium";v="92", " Not A;Brand";v="99", "Google Chrome";v="92"`},
			{"x-sessionid", task.TaskInfo.SessionID},
			{"sec-ch-ua-mobile", `?0`},
			{"user-agent", `Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/92.0.4515.131 Safari/537.36`},
			{"content-type", `application/json`},
			{"accept", `application/json, text/plain, */*`},
			{"x-ne-sign", newSign},
			{"x-requested-with", `XMLHttpRequest`},
			{"x-ne-sign-type", `simple`},
			{"origin", SecureBaseEndpoint},
			{"sec-fetch-site", `same-origin`},
			{"sec-fetch-mode", `cors`},
			{"sec-fetch-dest", `empty`},
			{"referer", fmt.Sprintf(BaseCheckoutEndpoint, task.TaskInfo.SessionID)},
			{"accept-encoding", `gzip, deflate, br`},
			{"accept-language", `en-US,en;q=0.9`},
		},

		RequestBodyStruct:  submitShippingInfoRequest,
		ResponseBodyStruct: &respMap,
	})
	if err != nil || resp.StatusCode != 200 {
		fmt.Println(err)
		return false
	}
	var result string
	var ok bool
	if result, ok = respMap["Result"].(string); !ok {
		return false
	}

	return result == "Success"
}

func (task *Task) GetPaymentToken() bool {
	jsonMap := map[string]interface{}{
		"SessionID":         task.TaskInfo.SessionID,
		"TransactionNumber": 0,
	}
	data, _ := json.Marshal(jsonMap)

	params, newSign := CreateExtras()
	if params == "" || newSign == "" {
		return false
	}

	respMap := make(map[string]interface{})
	resp, _, err := util.MakeRequest(&util.Request{
		Client: task.Task.Client,
		Method: "POST",
		URL:    GetPaymentTokenEndpoint + "?" + params,
		RawHeaders: http.RawHeader{
			{"sec-ch-ua", `"Chromium";v="92", " Not A;Brand";v="99", "Google Chrome";v="92"`},
			{"x-sessionid", task.TaskInfo.SessionID},
			{"sec-ch-ua-mobile", `?0`},
			{"user-agent", `Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/92.0.4515.131 Safari/537.36`},
			{"content-type", `application/json`},
			{"accept", `application/json, text/plain, */*`},
			{"x-ne-sign", newSign},
			{"x-requested-with", `XMLHttpRequest`},
			{"x-ne-sign-type", `simple`},
			{"origin", SecureBaseEndpoint},
			{"sec-fetch-site", `same-origin`},
			{"sec-fetch-mode", `cors`},
			{"sec-fetch-dest", `empty`},
			{"referer", fmt.Sprintf(BaseCheckoutEndpoint, task.TaskInfo.SessionID)},
			{"accept-encoding", `gzip, deflate, br`},
			{"accept-language", `en-US,en;q=0.9`},
		},

		Data:               data,
		ResponseBodyStruct: &respMap,
	})
	if err != nil || resp.StatusCode != 200 {
		fmt.Println(err)
		return false
	}
	var result string
	var ok bool
	if result, ok = respMap["Result"].(string); !ok {
		return false
	}

	var token string
	if token, ok = respMap["Token"].(string); !ok {
		return false
	}

	task.TaskInfo.PaymentToken = token

	return result == "Success"
}

func (task *Task) SubmitPaymentInfo() bool {
	submitPaymentInfoRequest := SubmitPaymentInfoRequest{
		Billing: Billing{
			Address1:    task.Task.Profile.BillingAddress.Address1,
			Address2:    task.Task.Profile.BillingAddress.Address2,
			City:        task.Task.Profile.BillingAddress.City,
			State:       task.Task.Profile.BillingAddress.StateCode,
			Zipcode:     task.Task.Profile.BillingAddress.ZipCode,
			Contactwith: task.Task.Profile.BillingAddress.FirstName + " " + task.Task.Profile.BillingAddress.LastName,
			Phone:       task.Task.Profile.PhoneNumber,
		},
		Creditcard: Creditcard{
			Holdername:       task.Task.Profile.CreditCard.CardholderName,
			Creditcardnumber: task.Task.Profile.CreditCard.CardNumber,
			Expirationdate:   task.Task.Profile.CreditCard.ExpMonth + "/" + task.Task.Profile.CreditCard.ExpYear,
			Paytermscode:     task.Task.Profile.CreditCard.CardType,
		},
		Token: task.TaskInfo.PaymentToken,
	}

	params, newSign := CreateExtras()
	if params == "" || newSign == "" {
		return false
	}

	respMap := make(map[string]interface{})
	resp, _, err := util.MakeRequest(&util.Request{
		Client: task.Task.Client,
		Method: "POST",
		URL:    SubmitPaymentInfoEndpoint + "?" + params,
		RawHeaders: http.RawHeader{
			{"sec-ch-ua", `"Chromium";v="92", " Not A;Brand";v="99", "Google Chrome";v="92"`},
			{"x-sessionid", task.TaskInfo.SessionID},
			{"sec-ch-ua-mobile", `?0`},
			{"user-agent", `Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/92.0.4515.131 Safari/537.36`},
			{"content-type", `application/json`},
			{"accept", `application/json, text/plain, */*`},
			{"x-ne-sign", newSign},
			{"x-requested-with", `XMLHttpRequest`},
			{"x-ne-sign-type", `simple`},
			{"origin", SecureBaseEndpoint},
			{"sec-fetch-site", `same-origin`},
			{"sec-fetch-mode", `cors`},
			{"sec-fetch-dest", `empty`},
			{"referer", fmt.Sprintf(BaseCheckoutEndpoint, task.TaskInfo.SessionID)},
			{"accept-encoding", `gzip, deflate, br`},
			{"accept-language", `en-US,en;q=0.9`},
		},

		RequestBodyStruct:  submitPaymentInfoRequest,
		ResponseBodyStruct: &respMap,
	})
	if err != nil || resp.StatusCode != 200 {
		fmt.Println(err)
		return false
	}
	var result string
	var ok bool
	if result, ok = respMap["Result"].(string); !ok {
		return false
	}

	return result == "Success"
}

func (task *Task) InitOrder() bool {
	initOrderRequest := InitOrderRequest{
		Sessionid: task.TaskInfo.SessionID,
		Actions: []Actions{
			{
				Actiontype:  "ConfirmPayment",
				Jsoncontent: `{"ActionType":"ConfirmPayment","Cvv2":"260","TransactionNumber":0,"PaytermsCode":"Visa"}`,
			},
		},
		Enableasynctoken: true,
	}

	params, newSign := CreateExtras()
	if params == "" || newSign == "" {
		return false
	}

	respMap := make(map[string]interface{})
	resp, _, err := util.MakeRequest(&util.Request{
		Client: task.Task.Client,
		Method: "POST",
		URL:    InitOrderEndpoint + "?" + params,
		RawHeaders: http.RawHeader{
			{"sec-ch-ua", `"Chromium";v="92", " Not A;Brand";v="99", "Google Chrome";v="92"`},
			{"x-sessionid", task.TaskInfo.SessionID},
			{"sec-ch-ua-mobile", `?0`},
			{"user-agent", `Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/92.0.4515.131 Safari/537.36`},
			{"content-type", `application/json`},
			{"accept", `application/json, text/plain, */*`},
			{"x-ne-sign", newSign},
			{"x-requested-with", `XMLHttpRequest`},
			{"x-ne-sign-type", `simple`},
			{"origin", SecureBaseEndpoint},
			{"sec-fetch-site", `same-origin`},
			{"sec-fetch-mode", `cors`},
			{"sec-fetch-dest", `empty`},
			{"referer", fmt.Sprintf(BaseCheckoutEndpoint, task.TaskInfo.SessionID)},
			{"accept-encoding", `gzip, deflate, br`},
			{"accept-language", `en-US,en;q=0.9`},
		},

		RequestBodyStruct:  initOrderRequest,
		ResponseBodyStruct: &respMap,
	})
	if err != nil || resp.StatusCode != 200 {
		fmt.Println(err)
		return false
	}
	var result string
	var ok bool
	if result, ok = respMap["Result"].(string); !ok {
		return false
	}

	return result == "Success"
}

func (task *Task) PlaceOrder(startTime time.Time) (bool, enums.OrderStatus) {
	status := enums.OrderStatusFailed
	placeOrderRequest := PlaceOrderRequest{
		Sessionid:               task.TaskInfo.SessionID,
		Isacceptnsccauth:        false,
		Subscribenewsletterflag: false,
		Creactaccount:           false,
		Password:                "",
		Mobilesubscribe:         Mobilesubscribe{},
		Languagecode:            "en-US",
		Cvv2:                    task.Task.Profile.CreditCard.CVV,
	}

	params, newSign := CreateExtras()
	if params == "" || newSign == "" {
		return false, status
	}

	var placeOrderResponse PlaceOrderResponse
	resp, _, err := util.MakeRequest(&util.Request{
		Client: task.Task.Client,
		Method: "POST",
		URL:    PlaceOrderEndpoint + "?" + params,
		RawHeaders: http.RawHeader{
			{"sec-ch-ua", `"Chromium";v="92", " Not A;Brand";v="99", "Google Chrome";v="92"`},
			{"x-sessionid", task.TaskInfo.SessionID},
			{"sec-ch-ua-mobile", `?0`},
			{"user-agent", `Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/92.0.4515.131 Safari/537.36`},
			{"content-type", `application/json`},
			{"accept", `application/json, text/plain, */*`},
			{"x-ne-sign", newSign},
			{"x-requested-with", `XMLHttpRequest`},
			{"x-ne-sign-type", `simple`},
			{"origin", SecureBaseEndpoint},
			{"sec-fetch-site", `same-origin`},
			{"sec-fetch-mode", `cors`},
			{"sec-fetch-dest", `empty`},
			{"referer", fmt.Sprintf(BaseCheckoutEndpoint, task.TaskInfo.SessionID)},
			{"accept-encoding", `gzip, deflate, br`},
			{"accept-language", `en-US,en;q=0.9`},
		},

		RequestBodyStruct:  placeOrderRequest,
		ResponseBodyStruct: &placeOrderResponse,
	})
	if err != nil || resp.StatusCode != 200 {
		fmt.Println(err)
		return false, status
	}

	if placeOrderResponse.Result != "Success" {
		return false, status
	}
	status = enums.OrderStatusSuccess
	success := true

	go util.ProcessCheckout(&util.ProcessCheckoutInfo{
		BaseTask:     task.Task,
		Success:      success,
		Status:       status,
		Content:      "",
		Embeds:       task.CreateNeweggEmbed(status, task.StockData.ImageURL),
		ItemName:     task.StockData.ProductName,
		ImageURL:     task.StockData.ImageURL,
		Sku:          task.StockData.SKU,
		Retailer:     enums.Newegg,
		Price:        float64(task.StockData.Price),
		Quantity:     task.Task.Task.TaskQty,
		MsToCheckout: time.Since(startTime).Milliseconds(),
	})

	task.TaskInfo.VBVToken = placeOrderResponse.Vbvdata.Jwttoken
	task.TaskInfo.CardBin = placeOrderResponse.Vbvdata.Cardbin

	return true, status
}

// Pretty sure these requests just speed up the payment, order goes through no matter what then you get a decline email a minute later if you send these requests.
// If you don't send these requests it takes 10-15 min to get the decline email.
func (task *Task) Verify() {
	if task.VerifyPayment() {
		task.VerifyOrder()
	}
}

func (task *Task) VerifyPayment() bool {
	verifyPaymentRequest := VerifyPaymentRequest{
		Browserpayload: Browserpayload{
			Order: Order{
				Consumer: Consumer{
					Account: Account{
						Accountnumber: task.TaskInfo.CardBin,
					},
				},
				Cart: []interface{}{},
			},
			Supportsalternativepayments: Supportsalternativepayments{
				Cca: true,
			},
		},
		Client: Client{
			Agent:   "SongbirdJS",
			Version: "1.34.0",
		},
		Consumersessionid: nil,
		Serverjwt:         task.TaskInfo.VBVToken,
	}

	resp, _, err := util.MakeRequest(&util.Request{
		Client: task.Task.Client,
		Method: "POST",
		URL:    VerifyPaymentEndpoint,
		RawHeaders: http.RawHeader{
			{"sec-ch-ua", `"Chromium";v="92", " Not A;Brand";v="99", "Google Chrome";v="92"`},
			//{"x-cardinal-tid", `Tid-024bd543-1de0-487b-98f1-5de71b4c2f39`},
			{"sec-ch-ua-mobile", `?0`},
			{"user-agent", `Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/92.0.4515.131 Safari/537.36`},
			{"content-type", `application/json;charset=UTF-8`},
			{"accept", `*/*`},
			{"origin", SecureBaseEndpoint},
			{"sec-fetch-site", `cross-site`},
			{"sec-fetch-mode", `cors`},
			{"sec-fetch-dest", `empty`},
			{"accept-encoding", `gzip, deflate, br`},
			{"accept-language", `en-US,en;q=0.9`},
		},

		RequestBodyStruct: verifyPaymentRequest,
	})
	if err != nil || resp.StatusCode != 200 {
		fmt.Println(err)
		return false
	}

	return true
}

func (task *Task) VerifyOrder() bool {
	verifyOrderRequest := VerifyOrderRequest{
		Useragent:     "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/92.0.4515.131 Safari/537.36",
		Browserheader: "text/html,application/xhtml+xml,application/xml;q=0.9,image/avif,image/webp,image/apng,*/*;q=0.8,application/signed-exchange;v=b3;q=0.9",
		Sessionid:     task.TaskInfo.SessionID,
	}

	params, newSign := CreateExtras()
	if params == "" || newSign == "" {
		return false
	}

	respMap := make(map[string]interface{})
	resp, _, err := util.MakeRequest(&util.Request{
		Client: task.Task.Client,
		Method: "POST",
		URL:    VerifyOrderEndpoint + "?" + params,
		RawHeaders: http.RawHeader{
			{"sec-ch-ua", `"Chromium";v="92", " Not A;Brand";v="99", "Google Chrome";v="92"`},
			{"x-sessionid", task.TaskInfo.SessionID},
			{"sec-ch-ua-mobile", `?0`},
			{"user-agent", `Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/92.0.4515.131 Safari/537.36`},
			{"content-type", `application/json`},
			{"accept", `application/json, text/plain, */*`},
			{"x-ne-sign", newSign},
			{"x-requested-with", `XMLHttpRequest`},
			{"x-ne-sign-type", `simple`},
			{"origin", SecureBaseEndpoint},
			{"sec-fetch-site", `same-origin`},
			{"sec-fetch-mode", `cors`},
			{"sec-fetch-dest", `empty`},
			{"referer", fmt.Sprintf(BaseCheckoutEndpoint, task.TaskInfo.SessionID)},
			{"accept-encoding", `gzip, deflate, br`},
			{"accept-language", `en-US,en;q=0.9`},
		},

		RequestBodyStruct:  verifyOrderRequest,
		ResponseBodyStruct: &respMap,
	})
	if err != nil || resp.StatusCode != 200 {
		fmt.Println(err)
		return false
	}
	var result string
	var ok bool
	if result, ok = respMap["Result"].(string); !ok {
		return false
	}

	return result == "Success"

}
