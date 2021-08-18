package gamestop

import (
	"fmt"
	"log"
	"net/url"
	"strings"
	"time"

	"backend.juicedbot.io/juiced.infrastructure/common"
	"backend.juicedbot.io/juiced.infrastructure/common/entities"
	"backend.juicedbot.io/juiced.infrastructure/common/enums"
	"backend.juicedbot.io/juiced.infrastructure/common/events"
	"backend.juicedbot.io/juiced.sitescripts/base"
	"backend.juicedbot.io/juiced.sitescripts/util"

	"github.com/anaskhan96/soup"
)

// CreateGamestopTask takes a Task entity and turns it into a Gamestop Task
func CreateGamestopTask(task *entities.Task, profile entities.Profile, proxyGroup *entities.ProxyGroup, eventBus *events.EventBus, taskType enums.TaskType, email, password string) (Task, error) {
	gamestopTask := Task{}

	gamestopTask = Task{
		Task: base.Task{
			Task:       task,
			Profile:    profile,
			ProxyGroup: proxyGroup,
			EventBus:   eventBus,
		},
		AccountInfo: AccountInfo{
			Email:    email,
			Password: password,
		},
		TaskType: taskType,
	}
	if proxyGroup != nil {
		gamestopTask.Task.Proxy = util.RandomLeastUsedProxy(proxyGroup.Proxies)
	}
	return gamestopTask, nil
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
// 		1. Login / Become a guest
// 		2. WaitForMonitor
// 		3. AddToCart
// 		4. Checkout
//		5. SetShippingInfo
// 		6. SetPaymentInfo
// 		7. PlaceOrder
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

	// 1. Login / Become a guest
	sessionMade := false
	for !sessionMade {
		needToStop := task.CheckForStop()
		if needToStop {
			return
		}
		switch task.TaskType {
		case enums.TaskTypeAccount:
			task.PublishEvent(enums.LoggingIn, enums.TaskStart)
			sessionMade = task.Login()
		case enums.TaskTypeGuest:
			task.PublishEvent(enums.SettingUp, enums.TaskStart)
			sessionMade = BecomeGuest(&task.Task.Client)
		}

		if !sessionMade {
			time.Sleep(time.Duration(task.Task.Task.TaskDelay) * time.Millisecond)
		}
	}

	task.PublishEvent(enums.WaitingForMonitor, enums.TaskUpdate)
	// 2. WaitForMonitor
	needToStop := task.WaitForMonitor()
	if needToStop {
		return
	}

	task.PublishEvent(enums.AddingToCart, enums.TaskUpdate)
	// 3. AddToCart
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

	task.PublishEvent(enums.GettingCartInfo, enums.TaskUpdate)
	startTime := time.Now()
	// 4. Checkout
	gotCartInfo := false
	for !gotCartInfo {
		needToStop := task.CheckForStop()
		if needToStop {
			return
		}
		gotCartInfo = task.Checkout()
		if !gotCartInfo {
			time.Sleep(time.Duration(task.Task.Task.TaskDelay) * time.Millisecond)
		}
	}

	task.PublishEvent(enums.SettingShippingInfo, enums.TaskUpdate)
	// 5. SetShippingInfo

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

	task.PublishEvent(enums.SettingBillingInfo, enums.TaskUpdate)
	// 6. SetPaymentInfo
	setPaymentInfo := false
	doNotRetry := false
	for !setPaymentInfo {
		needToStop := task.CheckForStop()
		if needToStop || doNotRetry {
			return
		}
		setPaymentInfo, doNotRetry = task.SetPaymentInfo()
		if !setPaymentInfo {
			time.Sleep(time.Duration(task.Task.Task.TaskDelay) * time.Millisecond)
		}
	}

	task.PublishEvent(enums.CheckingOut, enums.TaskUpdate)
	// 7. PlaceOrder
	placedOrder := false
	status := enums.OrderStatusFailed
	for !placedOrder {
		needToStop := task.CheckForStop()
		if needToStop {
			return
		}
		if status == enums.OrderStatusDeclined {
			break
		}

		placedOrder, status = task.PlaceOrder(startTime)
		if !placedOrder {
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
		task.PublishEvent(enums.CardDeclined, enums.TaskComplete)
	case enums.OrderStatusFailed:
		task.PublishEvent(enums.CheckoutFailed, enums.TaskComplete)
	}

}

// Logs the main client in
func (task *Task) Login() bool {
	_, _, err := util.MakeRequest(&util.Request{
		Client:     task.Task.Client,
		Method:     "GET",
		URL:        BaseEndpoint,
		RawHeaders: DefaultRawHeaders,
	})
	if err != nil {
		fmt.Println(err.Error())
	}

	err = util.NewAbck(&task.Task.Client, BaseEndpoint+"/", BaseEndpoint, AkamaiEndpoint)
	if err != nil {
		return false
	}

	loginResponse := LoginResponse{}
	form := url.Values{
		"loginEmail":         {task.AccountInfo.Email},
		"loginPassword":      {task.AccountInfo.Password},
		"loginRememberMe":    {"true"},
		"userTimezoneOffset": {"420"},
	}
	_, _, err = util.MakeRequest(&util.Request{
		Client: task.Task.Client,
		Method: "POST",
		URL:    LoginEndpoint,
		RawHeaders: [][2]string{
			{"content-length", fmt.Sprint(len(form.Encode()))},
			{"pragma", "no-cache"},
			{"cache-control", "no-cache"},
			{"sec-ch-ua", "\" Not A;Brand\";v=\"99\", \"Chromium\";v=\"90\", \"Google Chrome\";v=\"90\""},
			{"accept", "application/json, text/javascript, */*; q=0.01"},
			{"x-requested-with", "XMLHttpRequest"},
			{"sec-ch-ua-mobile", "?0"},
			{"user-agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/90.0.4430.212 Safari/537.36"},
			{"content-type", "application/x-www-form-urlencoded; charset=UTF-8"},
			{"origin", BaseEndpoint},
			{"sec-fetch-site", "same-origin"},
			{"sec-fetch-mode", "cors"},
			{"sec-fetch-dest", "empty"},
			{"referer", BaseEndpoint + "/"},
			{"accept-encoding", "gzip, deflate, br"},
			{"accept-language", "en-US,en;q=0.9"},
		},
		Data:               []byte(form.Encode()),
		ResponseBodyStruct: &loginResponse,
	})
	if err != nil {
		fmt.Println(err.Error())
	}

	if !loginResponse.Loginstatus.Success {
		return false
	}
	_, _, err = util.MakeRequest(&util.Request{
		Client: task.Task.Client,
		Method: "GET",
		URL:    AccountEndpoint + "/",
		RawHeaders: [][2]string{
			{"pragma", "no-cache"},
			{"cache-control", "no-cache"},
			{"sec-ch-ua", "\" Not A;Brand\";v=\"99\", \"Chromium\";v=\"90\", \"Google Chrome\";v=\"90\""},
			{"sec-ch-ua-mobile", "?0"},
			{"upgrade-insecure-requests", "1"},
			{"user-agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/90.0.4430.212 Safari/537.36"},
			{"accept", "text/html,application/xhtml+xml,application/xml;q=0.9,image/avif,image/webp,image/apng,*/*;q=0.8,application/signed-exchange;v=b3;q=0.9"},
			{"sec-fetch-site", "same-origin"},
			{"sec-fetch-mode", "navigate"},
			{"sec-fetch-user", "?1"},
			{"sec-fetch-dest", "document"},
			{"referer", BaseEndpoint + "/"},
			{"accept-encoding", "gzip, deflate, br"},
			{"accept-language", "en-US,en;q=0.9"},
		},
	})
	if err != nil {
		fmt.Println(err.Error())
	}

	return loginResponse.Loginstatus.Success
}

// WaitForMonitor waits until the Monitor has sent the info to the task to continue
func (task *Task) WaitForMonitor() bool {
	for {
		needToStop := task.CheckForStop()
		if needToStop {
			return true
		}
		if task.StockData.SKU != "" {
			return false
		}
		time.Sleep(common.MS_TO_WAIT)
	}
}

// AddToCart adds an item to the cart
func (task *Task) AddToCart() bool {
	quantity := task.Task.Task.TaskQty
	if quantity > task.StockData.MaxQuantity {
		quantity = task.StockData.MaxQuantity
	}
	addToCartResponse := AddToCartResponse{}
	form := url.Values{
		"pid":            {task.StockData.PID},
		"quantity":       {fmt.Sprint(quantity)},
		"upsellID":       {""},
		"purPROID":       {""},
		"options":        {"[]"},
		"isSDD":          {"false"},
		"deliveryOption": {"home"},
		"pageSpecified":  {"PDP"},
		"recommTitle":    {""},
	}
	resp, _, err := util.MakeRequest(&util.Request{
		Client: task.Task.Client,
		Method: "POST",
		URL:    fmt.Sprintf(AddToCartEndpoint, task.StockData.SKU),
		RawHeaders: [][2]string{
			{"content-length", fmt.Sprint(len(form.Encode()))},
			{"sec-ch-ua", "\" Not A;Brand\";v=\"99\", \"Chromium\";v=\"90\", \"Google Chrome\";v=\"90\""},
			{"accept", "*/*"},
			{"x-requested-with", "XMLHttpRequest"},
			{"sec-ch-ua-mobile", "?0"},
			{"user-agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/90.0.4430.212 Safari/537.36"},
			{"content-type", "application/x-www-form-urlencoded; charset=UTF-8"},
			{"origin", BaseEndpoint},
			{"sec-fetch-site", "same-origin"},
			{"sec-fetch-mode", "cors"},
			{"sec-fetch-dest", "empty"},
			{"referer", task.StockData.ProductURL},
			{"accept-encoding", "gzip, deflate, br"},
			{"accept-language", "en-US,en;q=0.9"},
		},
		Data:               []byte(form.Encode()),
		ResponseBodyStruct: &addToCartResponse,
	})
	if err != nil {
		fmt.Println(err.Error())
	}

	switch resp.StatusCode {
	case 200:
		switch addToCartResponse.Message {
		case "Added to cart":
			return true
		default:
			return false
		}
	case 500:
		return false
	default:
		return false
	}
}

// This is the longest request but a very important one because it gets the cross site scripting (csrf) token which is embedded in the html of the page
func (task *Task) Checkout() bool {
	resp, body, err := util.MakeRequest(&util.Request{
		Client: task.Task.Client,
		Method: "GET",
		URL:    CheckoutEndpoint + "/",
		RawHeaders: [][2]string{
			{"sec-ch-ua", "\" Not A;Brand\";v=\"99\", \"Chromium\";v=\"90\", \"Google Chrome\";v=\"90\""},
			{"sec-ch-ua-mobile", "?0"},
			{"upgrade-insecure-requests", "1"},
			{"user-agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/90.0.4430.212 Safari/537.36"},
			{"accept", "text/html,application/xhtml+xml,application/xml;q=0.9,image/avif,image/webp,image/apng,*/*;q=0.8,application/signed-exchange;v=b3;q=0.9"},
			{"sec-fetch-site", "same-origin"},
			{"sec-fetch-mode", "navigate"},
			{"sec-fetch-user", "?1"},
			{"sec-fetch-dest", "document"},
			{"referer", CheckoutLoginEndpoint + "/"},
			{"accept-encoding", "gzip, deflate, br"},
			{"accept-language", "en-US,en;q=0.9"},
		},
	})
	if err != nil {
		fmt.Println(err.Error())
	}

	doc := soup.HTMLParse(string(body))
	switch resp.StatusCode {
	case 200:
		task.CheckoutInfo.ShipmentUUID = doc.Find("input", "name", "shipmentUUID").Attrs()["value"]
		task.CheckoutInfo.OriginalShipmentUUID = doc.Find("input", "name", "originalShipmentUUID").Attrs()["value"]
		task.CheckoutInfo.CSRF = doc.Find("input", "name", "csrf_token").Attrs()["value"]
		if task.CheckoutInfo.ShipmentUUID != "" && task.CheckoutInfo.OriginalShipmentUUID != "" && task.CheckoutInfo.CSRF != "" {
			return true
		}
		return false

	default:
		return false
	}

}

// Setting the shipping for the order
func (task *Task) SetShippingInfo() bool {
	form := url.Values{
		"originalShipmentUUID":             {task.CheckoutInfo.OriginalShipmentUUID},
		"shipmentUUID":                     {task.CheckoutInfo.ShipmentUUID},
		"shippingAddressUpdateLinkClicked": {"false"},
		"dwfrm_shipping_shippingAddress_addressFields_firstName":        {task.Task.Profile.ShippingAddress.FirstName},
		"dwfrm_shipping_shippingAddress_addressFields_lastName":         {task.Task.Profile.ShippingAddress.LastName},
		"dwfrm_shipping_shippingAddress_addressFields_address1":         {task.Task.Profile.ShippingAddress.Address1},
		"dwfrm_shipping_shippingAddress_addressFields_address2":         {task.Task.Profile.ShippingAddress.Address2},
		"dwfrm_shipping_shippingAddress_addressFields_country":          {task.Task.Profile.ShippingAddress.CountryCode},
		"dwfrm_shipping_shippingAddress_addressFields_postalCode":       {task.Task.Profile.ShippingAddress.ZipCode},
		"dwfrm_shipping_shippingAddress_addressFields_city":             {task.Task.Profile.ShippingAddress.City},
		"dwfrm_shipping_shippingAddress_addressFields_states_stateCode": {task.Task.Profile.ShippingAddress.StateCode},
		"dwfrm_billing_email_emailAddress":                              {task.Task.Profile.Email},
		"dwfrm_shipping_shippingAddress_addressFields_phone":            {task.Task.Profile.PhoneNumber},
		"dwfrm_shipping_shippingAddress_shippingMethodID":               {"16"},
		"csrf_token": {task.CheckoutInfo.CSRF},
	}
	resp, _, err := util.MakeRequest(&util.Request{
		Client: task.Task.Client,
		Method: "POST",
		URL:    ShippingEndpoint,
		RawHeaders: [][2]string{
			{"content-length", fmt.Sprint(len(form.Encode()))},
			{"sec-ch-ua", "\" Not A;Brand\";v=\"99\", \"Chromium\";v=\"90\", \"Google Chrome\";v=\"90\""},
			{"accept", "application/json, text/javascript, */*; q=0.01"},
			{"x-requested-with", "XMLHttpRequest"},
			{"sec-ch-ua-mobile", "?0"},
			{"user-agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/90.0.4430.212 Safari/537.36"},
			{"content-type", "application/x-www-form-urlencoded; charset=UTF-8"},
			{"origin", BaseEndpoint},
			{"sec-fetch-site", "same-origin"},
			{"sec-fetch-mode", "cors"},
			{"sec-fetch-dest", "empty"},
			{"referer", CheckoutEndpoint + "/"},
			{"accept-encoding", "gzip, deflate, br"},
			{"accept-language", "en-US,en;q=0.9"},
		},
		Data: []byte(form.Encode()),
	})
	if err != nil {
		fmt.Println(err.Error())
	}

	switch resp.StatusCode {
	case 200:
		return true
	default:
		return false
	}

}

// Setting the payment for the order
func (task *Task) SetPaymentInfo() (bool, bool) {
	if !common.ValidCardType([]byte(task.Task.Profile.CreditCard.CardNumber), task.Task.Task.TaskRetailer) {
		return false, true
	}

	cardType := util.GetCardType([]byte(task.Task.Profile.CreditCard.CardNumber), task.Task.Task.TaskRetailer)

	// Not sure how this will work for gamestop compared to bestbuy, it will be a small change after testing if it's a problem
	util.NewAbck(&task.Task.Client, PaymentEndpoint+"/", BaseEndpoint, AkamaiEndpoint)

	// Format is not 02 instead just 2
	if task.Task.Profile.CreditCard.ExpMonth[0:1] == "0" {
		task.Task.Profile.CreditCard.ExpMonth = task.Task.Profile.CreditCard.ExpMonth[1:]
	}
	form := url.Values{
		"dwfrm_giftCard_balance_accountNumber":               {""},
		"dwfrm_giftCard_balance_pinNumber":                   {""},
		"g-recaptcha-response":                               {""},
		"dwfrm_billing_paymentMethod":                        {"CREDIT_CARD"},
		"dwfrm_billing_creditCardFields_cardType":            {cardType},
		"dwfrm_billing_creditCardFields_cardNumber":          {task.Task.Profile.CreditCard.CardNumber},
		"dwfrm_billing_creditCardFields_expirationMonth":     {task.Task.Profile.CreditCard.ExpMonth},
		"dwfrm_billing_creditCardFields_expirationYear":      {task.Task.Profile.CreditCard.ExpYear},
		"dwfrm_billing_creditCardFields_securityCode":        {task.Task.Profile.CreditCard.CVV},
		"addressSelector":                                    {"new"},
		"dwfrm_billing_addressFields_firstName":              {task.Task.Profile.BillingAddress.FirstName},
		"dwfrm_billing_addressFields_lastName":               {task.Task.Profile.BillingAddress.LastName},
		"dwfrm_billing_addressFields_address1":               {strings.ToUpper(task.Task.Profile.BillingAddress.Address1)},
		"dwfrm_billing_addressFields_address2":               {strings.ToUpper(task.Task.Profile.BillingAddress.Address2)},
		"dwfrm_billing_addressFields_country":                {task.Task.Profile.BillingAddress.CountryCode},
		"dwfrm_billing_addressFields_postalCode":             {task.Task.Profile.BillingAddress.ZipCode},
		"dwfrm_billing_addressFields_city":                   {strings.ToUpper(task.Task.Profile.BillingAddress.City)},
		"dwfrm_billing_addressFields_states_stateCode":       {task.Task.Profile.BillingAddress.StateCode},
		"dwfrm_billing_email_emailAddress":                   {task.Task.Profile.Email},
		"dwfrm_billing_addressFields_phone":                  {task.Task.Profile.PhoneNumber},
		"dwfrm_billing_purCreditCardFields_purAccountNumber": {""},
		"flexPay":              {"nonFlexPayment"},
		"flexPayImgUrls":       {""},
		"flexPayEnable":        {"true"},
		"flexPayEnableQuadPay": {""},
		"flexPayEnableKlarna":  {""},
		"flexPayEnableSezzle":  {""},
		"csrf_token":           {task.CheckoutInfo.CSRF},
		"flexpay":              {"nonFlexPayment"},
		"dwfrm_billing_accertify_accertifyDeviceFingerprint": {""},
	}

	resp, _, err := util.MakeRequest(&util.Request{
		Client: task.Task.Client,
		Method: "POST",
		URL:    PaymentEndpoint,
		RawHeaders: [][2]string{
			{"content-length", fmt.Sprint(len(form.Encode()))},
			{"sec-ch-ua", "\" Not A;Brand\";v=\"99\", \"Chromium\";v=\"90\", \"Google Chrome\";v=\"90\""},
			{"accept", "*/*"},
			{"x-requested-with", "XMLHttpRequest"},
			{"sec-ch-ua-mobile", "?0"},
			{"user-agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/90.0.4430.212 Safari/537.36"},
			{"content-type", "application/x-www-form-urlencoded; charset=UTF-8"},
			{"origin", BaseEndpoint},
			{"sec-fetch-site", "same-origin"},
			{"sec-fetch-mode", "cors"},
			{"sec-fetch-dest", "empty"},
			{"referer", CheckoutEndpoint + "/"},
			{"accept-encoding", "gzip, deflate, br"},
			{"accept-language", "en-US,en;q=0.9"},
		},
		Data: []byte(form.Encode()),
	})
	if err != nil {
		fmt.Println(err.Error())
	}

	switch resp.StatusCode {
	case 200:
		return true, false
	default:
		return false, false
	}
}

// The final request to place the order
func (task *Task) PlaceOrder(startTime time.Time) (bool, enums.OrderStatus) {
	status := enums.OrderStatusFailed
	placeOrderResponse := PlaceOrderResponse{}
	form := url.Values{
		"klarnaOrderId":              {""},
		"accertifyDeviceFingerprint": {""},
	}
	resp, _, err := util.MakeRequest(&util.Request{
		Client: task.Task.Client,
		Method: "POST",
		URL:    PlaceOrderEndpoint,
		RawHeaders: [][2]string{
			{"content-length", fmt.Sprint(len(form.Encode()))},
			{"sec-ch-ua", "\" Not A;Brand\";v=\"99\", \"Chromium\";v=\"90\", \"Google Chrome\";v=\"90\""},
			{"accept", "*/*"},
			{"x-requested-with", "XMLHttpRequest"},
			{"sec-ch-ua-mobile", "?0"},
			{"user-agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/90.0.4430.212 Safari/537.36"},
			{"content-type", "application/x-www-form-urlencoded; charset=UTF-8"},
			{"origin", BaseEndpoint},
			{"sec-fetch-site", "same-origin"},
			{"sec-fetch-mode", "cors"},
			{"sec-fetch-dest", "empty"},
			{"referer", CheckoutEndpoint + "/"},
			{"accept-encoding", "gzip, deflate, br"},
			{"accept-language", "en-US,en;q=0.9"},
		},
		Data:               []byte(form.Encode()),
		ResponseBodyStruct: &placeOrderResponse,
	})
	ok := util.HandleErrors(err, util.RequestDoError)
	if !ok {
		return false, status
	}

	var success bool
	switch resp.StatusCode {
	case 200:
		if placeOrderResponse.Error {
			status = enums.OrderStatusFailed
			if strings.Contains(placeOrderResponse.Errormessage, "another form of payment") {
				fmt.Println("Card Declined")
				status = enums.OrderStatusDeclined
				success = false
			}
		} else {
			status = enums.OrderStatusSuccess
			success = true
		}
	default:
		status = enums.OrderStatusFailed
		success = false
	}

	quantity := task.Task.Task.TaskQty
	if quantity > task.StockData.MaxQuantity {
		quantity = task.StockData.MaxQuantity
	}
	go util.ProcessCheckout(util.ProcessCheckoutInfo{
		BaseTask:     task.Task,
		Success:      success,
		Status:       status,
		Content:      "",
		Embeds:       task.CreateGamestopEmbed(status, task.StockData.ImageURL),
		ItemName:     task.StockData.ItemName,
		ImageURL:     task.StockData.ImageURL,
		Sku:          task.StockData.SKU,
		Retailer:     enums.GameStop,
		Price:        task.StockData.Price,
		Quantity:     quantity,
		MsToCheckout: time.Since(startTime).Milliseconds(),
	})

	return success, status

}
