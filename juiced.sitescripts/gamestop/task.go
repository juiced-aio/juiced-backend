package gamestop

import (
	"encoding/json"
	"fmt"
	"log"
	"net/url"
	"strings"
	"time"

	"backend.juicedbot.io/juiced.infrastructure/common/entities"
	"backend.juicedbot.io/juiced.infrastructure/common/enums"
	"backend.juicedbot.io/juiced.infrastructure/common/events"
	"backend.juicedbot.io/juiced.infrastructure/queries"
	sec "backend.juicedbot.io/juiced.security/auth/util"
	"backend.juicedbot.io/juiced.sitescripts/base"
	"backend.juicedbot.io/juiced.sitescripts/util"

	"github.com/anaskhan96/soup"
)

// CreateGamestopTask takes a Task entity and turns it into a Gamestop Task
func CreateGamestopTask(task *entities.Task, profile entities.Profile, proxy entities.Proxy, eventBus *events.EventBus, taskType enums.TaskType, email, password string) (Task, error) {
	gamestopTask := Task{}
	client, err := util.CreateClient(proxy)
	if err != nil {
		return gamestopTask, err
	}
	gamestopTask = Task{
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
		TaskType: taskType,
	}
	return gamestopTask, err
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
		recover()
		// TODO @silent: Let the UI know that a task failed
	}()

	task.PublishEvent(enums.LoggingIn, enums.TaskStart)
	// 1. Login / Become a guest
	sessionMade := false
	for !sessionMade {
		needToStop := task.CheckForStop()
		if needToStop {
			return
		}
		switch task.TaskType {
		case enums.TaskTypeAccount:
			sessionMade = task.Login()
		case enums.TaskTypeGuest:
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

	startTime := time.Now()

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
	// 7. PlaceOrder
	placedOrder := false
	for !placedOrder {
		needToStop := task.CheckForStop()
		if needToStop {
			return
		}
		placedOrder = task.PlaceOrder()
		if !placedOrder {
			time.Sleep(time.Duration(task.Task.Task.TaskDelay) * time.Millisecond)
		}
	}

	endTime := time.Now()

	log.Println("STARTED AT: " + startTime.String())
	log.Println("  ENDED AT: " + endTime.String())
	log.Println("TIME TO CHECK OUT: " + endTime.Sub(startTime).String())

	task.PublishEvent(enums.CheckedOut, enums.TaskComplete)
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
		if task.CheckoutInfo.SKUInStock != "" {
			return false
		}
	}
}

// AddToCart adds an item to the cart
func (task *Task) AddToCart() bool {
	addToCartResponse := AddToCartResponse{}
	form := url.Values{
		"pid":            {task.CheckoutInfo.PID},
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
		URL:    fmt.Sprintf(AddToCartEndpoint, task.CheckoutInfo.SKUInStock),
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
			{"referer", task.CheckoutInfo.ProductURL},
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

// Leaving this here for now incase NewAbck doesn't workout
func (task *Task) GetAbck() bool {
	var abckCookie string
	var akamaiResponse AkamaiResponse
	for _, cookie := range task.Task.Client.Jar.Cookies(ParsedBase) {
		if cookie.Name == "_abck" {
			abckCookie = cookie.Value
		}
	}
	form := url.Values{
		"authKey":        {"UFBfkndVOYYGZ58Uu8Mv49TrODCEiVE0gKpOAj679Z0dJoQThi9wGpoq6tpIRvrR"},
		"pageurl":        {"https://www.gamestop.com"},
		"skipkact":       {"true"},
		"skipmact":       {"true"},
		"onblur":         {"false"},
		"onfocus":        {"false"},
		"abck":           {abckCookie},
		"sensordatalink": {"https://www.gamestop.com/webcontent/de13b139ui242f1ac76f82d4de634d"},
		"ver":            {"1.7"},
		"firstpost":      {"true"},
		"pixelid":        {""},
		"pixelg":         {""},
		"json":           {"true"},
	}
	resp, _, err := util.MakeRequest(&util.Request{
		Client: task.Task.Client,
		Method: "POST",
		URL:    GenEndpoint,
		RawHeaders: [][2]string{
			{"Content-Type", "application/x-www-form-urlencoded"},
			{"User-Agent", "Juiced/1.0"},
			{"Accept", "*/*"},
			{"Accept-Encoding", "gzip, deflate, br"},
			{"Connection", "keep-alive"},
			{"Content-Length", fmt.Sprint(len(form.Encode()))},
		},
		Data:               []byte(form.Encode()),
		ResponseBodyStruct: &akamaiResponse,
	})
	if err != nil {
		fmt.Println(err.Error())
	}

	sensorRequest := SensorRequest{
		SensorData: akamaiResponse.Sensordata,
	}
	data, _ := json.Marshal(sensorRequest)
	resp, _, err = util.MakeRequest(&util.Request{
		Client: task.Task.Client,
		Method: "POST",
		URL:    AkamaiEndpoint,
		RawHeaders: [][2]string{
			{"content-length", fmt.Sprint(len(data))},
			{"sec-ch-ua", "\" Not A;Brand\";v=\"99\", \"Chromium\";v=\"90\", \"Google Chrome\";v=\"90\""},
			{"sec-ch-ua-mobile", "?0"},
			{"user-agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/90.0.4430.212 Safari/537.36"},
			{"content-type", "text/plain;charset=UTF-8"},
			{"accept", "*/*"},
			{"origin", "https://www.gamestop.com"},
			{"sec-fetch-site", "same-origin"},
			{"sec-fetch-mode", "cors"},
			{"sec-fetch-dest", "empty"},
			{"referer", CheckoutEndpoint + "/"},
			{"accept-encoding", "gzip, deflate, br"},
			{"accept-language", "en-US,en;q=0.9"},
		},
		Data:               data,
		ResponseBodyStruct: &akamaiResponse,
	})
	if err != nil {
		fmt.Println(err.Error())
	}

	switch resp.StatusCode {
	case 201:
		return true
	}
	return false
}

// Setting the payment for the order
func (task *Task) SetPaymentInfo() bool {
	// Not sure how this will work for gamestop compared to bestbuy, it will be a small change after testing if it's a problem
	util.NewAbck(&task.Task.Client, CheckoutEndpoint+"/", BaseEndpoint, AkamaiEndpoint)
	// Format is not 02 instead just 2
	if task.Task.Profile.CreditCard.ExpMonth[0:1] == "0" {
		task.Task.Profile.CreditCard.ExpMonth = task.Task.Profile.CreditCard.ExpMonth[1:]
	}
	form := url.Values{
		"dwfrm_giftCard_balance_accountNumber":               {""},
		"dwfrm_giftCard_balance_pinNumber":                   {""},
		"g-recaptcha-response":                               {""},
		"dwfrm_billing_paymentMethod":                        {"CREDIT_CARD"},
		"dwfrm_billing_creditCardFields_cardType":            {task.Task.Profile.CreditCard.CardType},
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
		return true
	default:
		return false
	}
}

// The final request to place the order
func (task *Task) PlaceOrder() bool {
	placeOrderResponse := PlaceOrderResponse{}
	form := url.Values{
		"klarnaOrderId": {""},
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
	if err != nil {
		fmt.Println(err.Error())
	}

	var status enums.OrderStatus
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

	_, user, err := queries.GetUserInfo()
	if err != nil {
		fmt.Println("Could not get user info")
		return false
	}
	sec.DiscordWebhook(success, "", task.CreateGamestopEmbed(status, task.CheckoutInfo.ImageURL), user)

	return success

}