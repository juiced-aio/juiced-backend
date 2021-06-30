package walmart

import (
	"encoding/json"
	"fmt"
	"log"
	"strconv"
	"strings"
	"time"

	"backend.juicedbot.io/juiced.client/http"
	"backend.juicedbot.io/juiced.infrastructure/common/entities"
	"backend.juicedbot.io/juiced.infrastructure/common/enums"
	"backend.juicedbot.io/juiced.infrastructure/common/events"
	"backend.juicedbot.io/juiced.sitescripts/base"
	"backend.juicedbot.io/juiced.sitescripts/util"
)

// CreateWalmartTask takes a Task entity and turns it into a Walmart Task
func CreateWalmartTask(task *entities.Task, profile entities.Profile, proxy entities.Proxy, eventBus *events.EventBus) (Task, error) {
	walmartTask := Task{}
	client, err := util.CreateClient(proxy)
	if err != nil {
		return walmartTask, err
	}

	if task.TaskDelay == 0 {
		task.TaskDelay = 2000
	}

	walmartTask = Task{
		Task: base.Task{
			Task:     task,
			Profile:  profile,
			Proxy:    proxy,
			EventBus: eventBus,
			Client:   client,
		},
	}
	return walmartTask, err
}

// RefreshPX3 refreshes the px3 cookie every 4 minutes since it expires every 5 minutes
func (task *Task) RefreshPX3() {
	defer func() {
		recover()
		task.RefreshPX3()
	}()

	for {
		if task.PXValues.RefreshAt == 0 || time.Now().Unix() > task.PXValues.RefreshAt {
			pxValues, err := SetPXCookie(task.Task.Proxy, &task.Task.Client)

			if err != nil {
				log.Println("Error setting px cookie for task: " + err.Error())
				return // TODO @silent
			}
			task.PXValues = pxValues
			task.PXValues.RefreshAt = time.Now().Unix() + 240
		}
	}
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
// 		1. WaitForMonitor
//		2. AddToCart
//		3. GetCartInfo
// 		4. SetPCID
//		5. SetShippingInfo
// 		6. WaitForEncryptedPaymentInfo
//		7. SetPaymentInfo
//		8. PlaceOrder
func (task *Task) RunTask() {
	// If the function panics due to a runtime error, recover from it
	defer func() {
		recover()
		task.Task.StopFlag = true
		task.PublishEvent(enums.TaskIdle, enums.TaskFail)
	}()

	task.PublishEvent(enums.SettingUp, enums.TaskStart)
	go task.RefreshPX3()
	for task.PXValues.RefreshAt == 0 {
	}

	setup := false
	for !setup {
		needToStop := task.CheckForStop()
		if needToStop {
			return
		}
		setup = task.Setup()
		if !setup {
			time.Sleep(time.Duration(task.Task.Task.TaskDelay) * time.Millisecond)
		}
	}

	// 1. WaitForMonitor
	task.PublishEvent(enums.WaitingForMonitor, enums.TaskUpdate)
	needToStop := task.WaitForMonitor()
	if needToStop {
		return
	}

	startTime := time.Now()

	// @Tehnic: The endpoint that you are monitoring with automatically adds it to the cart so you should somehow pass the
	// cookies/client to here and then completely cut out the AddToCart request, otherwise using a faster endpoint to monitor would be better.
	// 2. AddToCart
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

	go func() {
		pieValues := PIEValues{}
		for pieValues.K == "" {
			pieValues = task.GetPIEValues()
			if pieValues.K == "" {
				time.Sleep(time.Duration(task.Task.Task.TaskDelay) * time.Millisecond)
			}
		}
		cardInfo := EncryptCardInfo{
			CardNumber: task.Task.Profile.CreditCard.CardNumber,
			CardCVV:    task.Task.Profile.CreditCard.CVV,
			PIEValues:  pieValues,
		}
		task.Task.EventBus.PublishTaskEvent(enums.EncryptingCardInfo, enums.TaskUpdate, cardInfo, task.Task.Task.ID)
	}()

	// 3. GetCartInfo
	task.PublishEvent(enums.GettingCartInfo, enums.TaskUpdate)
	gotCartInfo := false
	for !gotCartInfo {
		needToStop := task.CheckForStop()
		if needToStop {
			return
		}
		gotCartInfo = task.GetCartInfo()
		if !gotCartInfo {
			time.Sleep(time.Duration(task.Task.Task.TaskDelay) * time.Millisecond)
		}
	}

	// // 4. SetPCID
	// task.PublishEvent(enums.SettingCartInfo, enums.TaskUpdate)
	// setPCID := false
	// for !setPCID {
	// 	needToStop := task.CheckForStop()
	// 	if needToStop {
	// 		return
	// 	}
	// 	setPCID = task.SetPCID()
	// 	if !setPCID {
	// 		time.Sleep(time.Duration(task.Task.Task.TaskDelay) * time.Millisecond)
	// 	}
	// }

	// 5. SetShippingInfo
	task.PublishEvent(enums.SettingShippingInfo, enums.TaskUpdate)
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

	// 6. WaitForEncryptedPaymentInfo
	task.PublishEvent(enums.GettingBillingInfo, enums.TaskUpdate)
	needToStop = task.WaitForEncryptedPaymentInfo()
	if needToStop {
		return
	}

	// 7. SetPaymentInfo
	task.PublishEvent(enums.SettingBillingInfo, enums.TaskUpdate)
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

	// 8. PlaceOrder
	task.PublishEvent(enums.CheckingOut, enums.TaskUpdate)
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
	log.Println("TIME TO CHECK OUT: ", endTime.Sub(startTime).Milliseconds())

	task.PublishEvent(enums.CheckedOut, enums.TaskComplete)
}

// WaitForMonitor waits until the Monitor has sent the info to the task to continue
func (task *Task) WaitForMonitor() bool {
	for {
		needToStop := task.CheckForStop()
		if needToStop {
			return true
		}
		if task.OfferID != "" && task.Sku != "" {
			return false
		}
	}
}

func (task *Task) HandlePXCap(resp *http.Response, redirectURL string) bool {
	task.PublishEvent(enums.WaitingForCaptcha, enums.TaskUpdate)
	captchaURL := resp.Request.URL.String()
	if redirectURL != "" {
		captchaURL = BaseEndpoint + redirectURL
	}
	err := SetPXCapCookie(strings.ReplaceAll(captchaURL, "affil.", ""), &task.PXValues, task.Task.Proxy, &task.Task.Client)
	if err != nil {
		log.Println(err.Error())
		return false
	} else {
		log.Println("Cookie updated.")
		return true
	}
}

// Setup sends a GET request to the BaseEndpoint
func (task *Task) Setup() bool {
	_, _, err := util.MakeRequest(&util.Request{
		Client: task.Task.Client,
		Method: "GET",
		URL:    BaseEndpoint,
		RawHeaders: [][2]string{
			{"accept", "text/html,application/xhtml+xml,application/xml;q=0.9,image/avif,image/webp,image/apng,*/*;q=0.8,application/signed-exchange;v=b3;q=0.9"},
			{"accept-encoding", "gzip, deflate, br"},
			{"accept-language", "en-US,en;q=0.9"},
			{"sec-ch-ua", `" Not;A Brand";v="99", "Google Chrome";v="91", "Chromium";v="91"`},
			{"sec-ch-ua-mobile", "?0"},
			{"sec-fetch-dest", "document"},
			{"sec-fetch-mode", "navigate"},
			{"sec-fetch-site", "none"},
			{"sec-fetch-user", "?1"},
			{"upgrade-insecure-requests", "1"},
			{"user-agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/91.0.4472.77 Safari/537.36"},
		},
	})
	if err != nil {
		log.Println("Setup error: " + err.Error())
	}

	return err == nil
}

func (task *Task) GetPIEValues() PIEValues {
	pieValues := PIEValues{}
	resp, body, err := util.MakeRequest(&util.Request{
		Client: task.Task.Client,
		Method: "GET",
		URL:    PIEEndpoint + fmt.Sprint(time.Now().Unix()),
		RawHeaders: [][2]string{
			{"accept", "application/json"},
			{"accept-encoding", "gzip, deflate, br"},
			{"accept-language", "en-US,en;q=0.9"},
			{"content-type", "application/json"},
			{"sec-ch-ua", `" Not;A Brand";v="99", "Google Chrome";v="91", "Chromium";v="91"`},
			{"sec-ch-ua-mobile", "?0"},
			{"sec-fetch-dest", "document"},
			{"sec-fetch-mode", "navigate"},
			{"sec-fetch-site", "none"},
			{"sec-fetch-user", "?1"},
			{"upgrade-insecure-requests", "1"},
			{"user-agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/91.0.4472.77 Safari/537.36"},
			{"referer", PIEReferer},
		},
	})

	if err != nil || resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return pieValues
	}

	LStr, err := util.FindInString(body, "PIE.L = ", ";")
	if err != nil || LStr == "" {
		return pieValues
	}
	L, err := strconv.Atoi(LStr)
	if err != nil {
		return pieValues
	}
	EStr, err := util.FindInString(body, "PIE.E = ", ";")
	if err != nil || EStr == "" {
		return pieValues
	}
	E, err := strconv.Atoi(EStr)
	if err != nil {
		return pieValues
	}
	K, err := util.FindInString(body, `PIE.K = "`, `";`)
	if err != nil || K == "" {
		return pieValues
	}
	KeyID, err := util.FindInString(body, `PIE.key_id = "`, `";`)
	if err != nil || KeyID == "" {
		return pieValues
	}

	pieValues = PIEValues{
		L:     L,
		E:     E,
		K:     K,
		KeyID: KeyID,
		Phase: 0,
	}

	return pieValues
}

// AddToCart sends a POST request to the AddToCartEndpoint with an AddToCartRequest body
func (task *Task) AddToCart() bool {
	addToCartResponse := AddToCartResponse{}
	data := AddToCartRequest{
		OfferID:               task.OfferID,
		Quantity:              1,
		ShipMethodDefaultRule: "SHIP_RULE_1",
	}
	dataStr, err := json.Marshal(data)
	if err != nil {
		log.Println("ATC Request Error: " + err.Error())
		return false
	}
	resp, _, err := util.MakeRequest(&util.Request{
		Client: task.Task.Client,
		Method: "POST",
		URL:    AddToCartEndpoint,
		RawHeaders: [][2]string{
			{"accept", "application/json"},
			{"accept-encoding", "gzip, deflate, br"},
			{"accept-language", "en-US,en;q=0.9"},
			{"content-type", "application/json"},
			{"sec-ch-ua", `" Not;A Brand";v="99", "Google Chrome";v="91", "Chromium";v="91"`},
			{"sec-ch-ua-mobile", "?0"},
			{"sec-fetch-dest", "document"},
			{"sec-fetch-mode", "navigate"},
			{"sec-fetch-site", "none"},
			{"sec-fetch-user", "?1"},
			{"upgrade-insecure-requests", "1"},
			{"user-agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/91.0.4472.77 Safari/537.36"},
			{"content-length", fmt.Sprint(len(dataStr))},
			{"referer", AddToCartReferer + "ip/" + task.Sku + "/sellers"},
		},
		RequestBodyStruct:  data,
		ResponseBodyStruct: &addToCartResponse,
	})
	if strings.Contains(resp.Request.URL.String(), "blocked") || (addToCartResponse.RedirectURL != "" && strings.Contains(addToCartResponse.RedirectURL, "blocked")) {
		handled := task.HandlePXCap(resp, addToCartResponse.RedirectURL)
		if handled {
			task.PublishEvent(enums.AddingToCart, enums.TaskUpdate)
		}
		return false
	}
	if err != nil {
		log.Println("ATC Request Error: " + err.Error())
		return false
	}

	switch resp.StatusCode {
	case 404:
		log.Printf("Not Found: %v\n", resp.Status)
	default:
		log.Printf("Unknown Code: %v\n", resp.StatusCode)
	}
	if resp.StatusCode >= 200 && resp.StatusCode < 300 && addToCartResponse.Cart.ItemCount > 0 {
		return true
	}
	return false
}

// GetCartInfo is required for setting the PCID cookie
func (task *Task) GetCartInfo() bool {
	getCartInfoResponse := GetCartInfoResponse{}
	data := GetCartInfoRequest{
		StoreListIds:  []StoreList{},
		ZipCode:       task.Task.Profile.ShippingAddress.ZipCode,
		City:          task.Task.Profile.ShippingAddress.City,
		State:         task.Task.Profile.ShippingAddress.StateCode,
		IsZipLocated:  true, //setting true as we are populating with values
		Crt:           "",
		CustomerId:    "",
		CustomerType:  "",
		AffiliateInfo: "",
	}
	dataStr, err := json.Marshal(data)
	if err != nil {
		log.Println("GetCartInfo Request Error: " + err.Error())
		return false
	}
	resp, _, err := util.MakeRequest(&util.Request{
		Client: task.Task.Client,
		Method: "POST",
		URL:    GetCartInfoEndpoint,
		RawHeaders: [][2]string{
			{"accept", "application/json"},
			{"accept-encoding", "gzip, deflate, br"},
			{"accept-language", "en-US,en;q=0.9"},
			{"content-type", "application/json"},
			{"sec-ch-ua", `" Not;A Brand";v="99", "Google Chrome";v="91", "Chromium";v="91"`},
			{"sec-ch-ua-mobile", "?0"},
			{"sec-fetch-dest", "document"},
			{"sec-fetch-mode", "navigate"},
			{"sec-fetch-site", "none"},
			{"sec-fetch-user", "?1"},
			{"upgrade-insecure-requests", "1"},
			{"user-agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/91.0.4472.77 Safari/537.36"},
			{"content-length", fmt.Sprint(len(dataStr))},
			{"referer", GetCartInfoReferer},
		},
		RequestBodyStruct:  data,
		ResponseBodyStruct: &getCartInfoResponse,
	})
	if strings.Contains(resp.Request.URL.String(), "blocked") || (getCartInfoResponse.RedirectURL != "" && strings.Contains(getCartInfoResponse.RedirectURL, "blocked")) {
		handled := task.HandlePXCap(resp, getCartInfoResponse.RedirectURL)
		if handled {
			task.PublishEvent(enums.GettingCartInfo, enums.TaskUpdate)
		}
	}
	if err != nil {
		log.Println("GetCartInfo Request Error: " + err.Error())
		return false
	}

	switch resp.StatusCode {
	case 201:
	case 404:
		log.Printf("Not Found: %v\n", resp.Status)
	default:
		log.Printf("Unknown Code: %v\n", resp.StatusCode)
	}
	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		return true
	}
	return false
}

// SetPCID sets the PCID cookie
func (task *Task) SetPCID() bool {
	setPCIDResponse := SetPCIDResponse{}
	resp, _, err := util.MakeRequest(&util.Request{
		Client: task.Task.Client,
		Method: "POST",
		URL:    SetPcidEndpoint,
		RawHeaders: [][2]string{
			{"accept", "application/json"},
			{"accept-encoding", "gzip, deflate, br"},
			{"accept-language", "en-US,en;q=0.9"},
			{"content-type", "application/json"},
			{"sec-ch-ua", `" Not;A Brand";v="99", "Google Chrome";v="91", "Chromium";v="91"`},
			{"sec-ch-ua-mobile", "?0"},
			{"sec-fetch-dest", "document"},
			{"sec-fetch-mode", "navigate"},
			{"sec-fetch-site", "none"},
			{"sec-fetch-user", "?1"},
			{"upgrade-insecure-requests", "1"},
			{"user-agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/91.0.4472.77 Safari/537.36"},
			{"content-length", "0"},
			{"referer", SetPcidReferer},
		},
		ResponseBodyStruct: &setPCIDResponse,
	})
	if strings.Contains(resp.Request.URL.String(), "blocked") || (setPCIDResponse.RedirectURL != "" && strings.Contains(setPCIDResponse.RedirectURL, "blocked")) {
		handled := task.HandlePXCap(resp, setPCIDResponse.RedirectURL)
		if handled {
			task.PublishEvent(enums.SettingCartInfo, enums.TaskUpdate)
		}
	}
	if err != nil {
		log.Println("SetPCID Request Error: " + err.Error())
		return false
	}

	switch resp.StatusCode {
	case 404:
		log.Printf("Not Found: %v\n", resp.Status)
	default:
		log.Printf("Unknown Code: %v\n", resp.StatusCode)
	}
	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		return true
	}
	return false
}

// SetShippingInfo sets the shipping address
func (task *Task) SetShippingInfo() bool {
	setShippingInfoResponse := SetShippingInfoResponse{}
	data := SetShippingInfoRequest{
		AddressLineOne:     task.Task.Profile.ShippingAddress.Address1,
		City:               task.Task.Profile.ShippingAddress.City,
		FirstName:          task.Task.Profile.ShippingAddress.FirstName,
		LastName:           task.Task.Profile.ShippingAddress.LastName,
		Phone:              task.Task.Profile.PhoneNumber,
		Email:              task.Task.Profile.Email,
		MarketingEmailPref: false,
		PostalCode:         task.Task.Profile.ShippingAddress.ZipCode,
		State:              task.Task.Profile.ShippingAddress.StateCode,
		CountryCode:        task.Task.Profile.ShippingAddress.CountryCode,
		ChangedFields:      []string{},
		Storelist:          []Storelist{},
	}
	dataStr, err := json.Marshal(data)
	if err != nil {
		log.Println("SetShippingInfo Request Error: " + err.Error())
		return false
	}
	resp, _, err := util.MakeRequest(&util.Request{
		Client: task.Task.Client,
		Method: "POST",
		URL:    SetShippingInfoEndpoint,
		RawHeaders: [][2]string{
			{"accept", "application/json, text/javascript, */*; q=0.01"},
			{"accept-encoding", "gzip, deflate, br"},
			{"accept-language", "en-US,en;q=0.9"},
			{"content-type", "application/json"},
			{"origin", "https://www.walmart.com"},
			{"sec-ch-ua", `" Not;A Brand";v="99", "Google Chrome";v="91", "Chromium";v="91"`},
			{"sec-ch-ua-mobile", "?0"},
			{"sec-fetch-dest", "empty"},
			{"sec-fetch-mode", "cors"},
			{"sec-fetch-site", "same-origin"},
			{"upgrade-insecure-requests", "1"},
			{"user-agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/91.0.4472.77 Safari/537.36"},
			{"content-length", fmt.Sprint(len(dataStr))},
			{"referer", SetShippingInfoReferer},
			{"wm_cvv_in_session", "true"},
			{"wm_vertical_id", "0"},
			{"inkiru_precedence", "false"},
		},
		RequestBodyStruct:  data,
		ResponseBodyStruct: &setShippingInfoResponse,
	})
	if strings.Contains(resp.Request.URL.String(), "blocked") || (setShippingInfoResponse.RedirectURL != "" && strings.Contains(setShippingInfoResponse.RedirectURL, "blocked")) {
		handled := task.HandlePXCap(resp, setShippingInfoResponse.RedirectURL)
		if handled {
			task.PublishEvent(enums.SettingShippingInfo, enums.TaskUpdate)
		}
	}
	if err != nil {
		log.Println("SetShippingInfo Request Error: " + err.Error())
		return false
	}

	switch resp.StatusCode {
	case 200:
	case 404:
		log.Printf("Not Found: %v\n", resp.Status)
	default:
		log.Printf("Unknown Code: %v\n", resp.StatusCode)
	}
	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		return true
	}
	return false
}

// WaitForEncryptedPaymentInfo waits until the Monitor has sent the info to the task to continue
func (task *Task) WaitForEncryptedPaymentInfo() bool {
	for {
		needToStop := task.CheckForStop()
		if needToStop {
			return true
		}
		if task.CardInfo.EncryptedPan != "" {
			return false
		}
	}
}

// SetPaymentInfo sets the payment info to prepare for placing an order
func (task *Task) SetPaymentInfo() bool {
	setPaymentInfoResponse := SetPaymentInfoResponse{}
	data := PaymentsRequest{
		[]Payment{{
			PaymentType:    "CreditCard",
			CardType:       task.Task.Profile.CreditCard.CardType,
			FirstName:      task.Task.Profile.BillingAddress.FirstName,
			LastName:       task.Task.Profile.BillingAddress.LastName,
			AddressLineOne: task.Task.Profile.BillingAddress.Address1,
			AddressLineTwo: task.Task.Profile.BillingAddress.Address2,
			City:           task.Task.Profile.BillingAddress.City,
			State:          task.Task.Profile.BillingAddress.StateCode,
			PostalCode:     task.Task.Profile.BillingAddress.ZipCode,
			ExpiryMonth:    task.Task.Profile.CreditCard.ExpMonth,
			ExpiryYear:     task.Task.Profile.CreditCard.ExpYear,
			Email:          task.Task.Profile.Email,
			Phone:          task.Task.Profile.PhoneNumber,
			EncryptedPan:   task.CardInfo.EncryptedPan,
			EncryptedCvv:   task.CardInfo.EncryptedCvv,
			IntegrityCheck: task.CardInfo.IntegrityCheck,
			KeyId:          task.CardInfo.KeyId,
			Phase:          task.CardInfo.Phase,
			PiHash:         task.CardInfo.PiHash,
		}},
		true,
	}
	dataStr, err := json.Marshal(data)
	if err != nil {
		log.Println("SetPaymentInfo Request Error: " + err.Error())
		return false
	}
	resp, _, err := util.MakeRequest(&util.Request{
		Client: task.Task.Client,
		Method: "POST",
		URL:    SetPaymentInfoEndpoint,
		RawHeaders: [][2]string{
			{"accept", "application/json"},
			{"accept-encoding", "gzip, deflate, br"},
			{"accept-language", "en-US,en;q=0.9"},
			{"content-type", "application/json"},
			{"sec-ch-ua", `" Not;A Brand";v="99", "Google Chrome";v="91", "Chromium";v="91"`},
			{"sec-ch-ua-mobile", "?0"},
			{"sec-fetch-dest", "document"},
			{"sec-fetch-mode", "navigate"},
			{"sec-fetch-site", "none"},
			{"sec-fetch-user", "?1"},
			{"upgrade-insecure-requests", "1"},
			{"user-agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/91.0.4472.77 Safari/537.36"},
			{"content-length", fmt.Sprint(len(dataStr))},
			{"referer", SetPaymentInfoReferer},
		},
		RequestBodyStruct:  data,
		ResponseBodyStruct: &setPaymentInfoResponse,
	})
	if strings.Contains(resp.Request.URL.String(), "blocked") || (setPaymentInfoResponse.RedirectURL != "" && strings.Contains(setPaymentInfoResponse.RedirectURL, "blocked")) {
		handled := task.HandlePXCap(resp, setPaymentInfoResponse.RedirectURL)
		if handled {
			task.PublishEvent(enums.SettingBillingInfo, enums.TaskUpdate)
		}
	}
	if err != nil {
		log.Println("SetPaymentInfo Request Error: " + err.Error())
		return false
	}

	switch resp.StatusCode {
	case 404:
		log.Printf("Not Found: %v\n", resp.Status)
	default:
		log.Printf("Unknown Code: %v\n", resp.StatusCode)
	}
	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		return true
	}
	return false
}

// PlaceOrder completes the checkout process
func (task *Task) PlaceOrder() bool {
	placeOrderResponse := PlaceOrderResponse{}
	data := PlaceOrderRequest{
		CvvInSession: true,
		VoltagePayment: []VoltagePayment{{
			PaymentType:    "CREDITCARD",
			EncryptedCvv:   task.CardInfo.EncryptedCvv,
			EncryptedPan:   task.CardInfo.EncryptedPan,
			IntegrityCheck: task.CardInfo.IntegrityCheck,
			KeyId:          task.CardInfo.KeyId,
			Phase:          task.CardInfo.Phase,
		}},
	}
	dataStr, err := json.Marshal(data)
	if err != nil {
		log.Println("PlaceOrder Request Error: " + err.Error())
		return false
	}
	resp, _, err := util.MakeRequest(&util.Request{
		Client: task.Task.Client,
		Method: "POST",
		URL:    PlaceOrderEndpoint,
		RawHeaders: [][2]string{
			{"accept", "application/json"},
			{"accept-encoding", "gzip, deflate, br"},
			{"accept-language", "en-US,en;q=0.9"},
			{"content-type", "application/json"},
			{"sec-ch-ua", `" Not;A Brand";v="99", "Google Chrome";v="91", "Chromium";v="91"`},
			{"sec-ch-ua-mobile", "?0"},
			{"sec-fetch-dest", "document"},
			{"sec-fetch-mode", "navigate"},
			{"sec-fetch-site", "none"},
			{"sec-fetch-user", "?1"},
			{"upgrade-insecure-requests", "1"},
			{"user-agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/91.0.4472.77 Safari/537.36"},
			{"content-length", fmt.Sprint(len(dataStr))},
			{"referer", PlaceOrderReferer},
		},
		RequestBodyStruct:  data,
		ResponseBodyStruct: &placeOrderResponse,
	})
	if strings.Contains(resp.Request.URL.String(), "blocked") || (placeOrderResponse.RedirectURL != "" && strings.Contains(placeOrderResponse.RedirectURL, "blocked")) {
		handled := task.HandlePXCap(resp, placeOrderResponse.RedirectURL)
		if handled {
			task.PublishEvent(enums.CheckingOut, enums.TaskUpdate)
		}
	}
	if err != nil {
		log.Println("PlaceOrder Request Error: " + err.Error())
		return false
	}

	switch resp.StatusCode {
	case 404:
		log.Printf("Not Found: %v\n", resp.Status)
	default:
		log.Printf("Unknown Code: %v\n", resp.StatusCode)
	}
	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		return true
	}
	return false
}
