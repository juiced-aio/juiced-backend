package walmart

import (
	"encoding/json"
	"fmt"
	"log"
	"net/url"
	"strconv"
	"strings"
	"time"

	"backend.juicedbot.io/juiced.client/http"
	"backend.juicedbot.io/juiced.infrastructure/common"
	"backend.juicedbot.io/juiced.infrastructure/common/entities"
	"backend.juicedbot.io/juiced.infrastructure/common/enums"
	"backend.juicedbot.io/juiced.infrastructure/common/events"
	"backend.juicedbot.io/juiced.sitescripts/base"
	"backend.juicedbot.io/juiced.sitescripts/util"
)

// CreateWalmartTask takes a Task entity and turns it into a Walmart Task
func CreateWalmartTask(task *entities.Task, profile entities.Profile, proxyGroup *entities.ProxyGroup, eventBus *events.EventBus) (Task, error) {
	walmartTask := Task{
		Task: base.Task{
			Task:       task,
			Profile:    profile,
			ProxyGroup: proxyGroup,
			EventBus:   eventBus,
		},
	}
	if proxyGroup != nil {
		walmartTask.Task.Proxy = util.RandomLeastUsedProxy(proxyGroup.Proxies)
	} else {
		walmartTask.Task.Proxy = nil
	}
	return walmartTask, nil
}

// RefreshPX3 refreshes the px3 cookie every 4 minutes since it expires every 5 minutes
func (task *Task) RefreshPX3() {
	quit := make(chan bool)
	defer func() {
		quit <- true
	}()

	cancellationToken := &util.CancellationToken{Cancel: false}
	go func() {
		for {
			select {
			case <-quit:
				return
			default:
				needToStop := task.CheckForStop()
				if needToStop {
					cancellationToken.Cancel = true
					return
				}
			}
			time.Sleep(common.MS_TO_WAIT)
		}
	}()

	retry := true
	for retry {
		retry = task.RefreshPX3Helper(cancellationToken)
		time.Sleep(common.MS_TO_WAIT)
	}
}

func (task *Task) RefreshPX3Helper(cancellationToken *util.CancellationToken) bool {
	for {
		if cancellationToken.Cancel {
			return false
		}
		if task.PXValues.RefreshAt == 0 || time.Now().Unix() > task.PXValues.RefreshAt {
			pxValues, cancelled, err := SetPXCookie(task.Task.Proxy, &task.Task.Client, cancellationToken)
			if cancelled {
				return false
			}

			if err != nil {
				log.Println("Error setting px cookie for task: " + err.Error())
				return true
			}
			task.PXValues = pxValues
			task.PXValues.RefreshAt = time.Now().Unix() + 240
		}
		time.Sleep(common.MS_TO_WAIT)
	}
}

// PublishEvent wraps the EventBus's PublishTaskEvent function
func (task *Task) PublishEvent(status enums.TaskStatus, eventType enums.TaskEventType, statusPercentage int) {
	if status == enums.TaskIdle || !task.Task.StopFlag {
		task.Task.Task.SetTaskStatus(status)
		task.Task.EventBus.PublishTaskEvent(status, statusPercentage, eventType, nil, task.Task.Task.ID)
	}
}

// CheckForStop checks the stop flag and stops the monitor if it's true
func (task *Task) CheckForStop() bool {
	if task.Task.StopFlag && !task.Task.DontPublishEvents {
		task.PublishEvent(enums.TaskIdle, enums.TaskStop, 0)
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
//		7. SetCreditCard
//		8. SetPaymentInfo
//		9. PlaceOrder
func (task *Task) RunTask() {
	// If the function panics due to a runtime error, recover from it
	defer func() {
		if r := recover(); r != nil {
			task.PublishEvent(fmt.Sprintf(enums.TaskFailed, r), enums.TaskFail, 0)
		} else {
			if !task.Task.StopFlag &&
				!strings.Contains(task.Task.Task.TaskStatus, strings.ReplaceAll(enums.TaskIdle, " %s", "")) &&
				!strings.Contains(task.Task.Task.TaskStatus, strings.ReplaceAll(enums.CheckingOutFailure, " %s", "")) &&
				!strings.Contains(task.Task.Task.TaskStatus, strings.ReplaceAll(enums.CardDeclined, " %s", "")) &&
				!strings.Contains(task.Task.Task.TaskStatus, strings.ReplaceAll(enums.CheckingOutSuccess, " %s", "")) &&
				!strings.Contains(task.Task.Task.TaskStatus, strings.ReplaceAll(enums.TaskFailed, " %s", "")) {
				task.PublishEvent(enums.TaskIdle, enums.TaskStop, 0)
			}
		}
		task.Task.StopFlag = true
	}()
	task.StockData = WalmartInStockData{}
	task.Task.HasStockData = false

	if task.Task.Task.TaskDelay == 0 {
		task.Task.Task.TaskDelay = 2000
	}
	if task.Task.Task.TaskQty <= 0 {
		task.Task.Task.TaskQty = 1
	}

	err := task.Task.CreateClient(task.Task.Proxy)
	if err != nil {
		return
	}

	task.PublishEvent(enums.SettingUp, enums.TaskStart, 5)
	go task.RefreshPX3()
	for task.PXValues.RefreshAt == 0 {
		needToStop := task.CheckForStop()
		if needToStop {
			return
		}
		time.Sleep(common.MS_TO_WAIT)
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
	task.PublishEvent(enums.WaitingForMonitor, enums.TaskUpdate, 20)
	needToStop := task.WaitForMonitor()
	if needToStop {
		return
	}

	startTime := time.Now()

	// @Tehnic: The endpoint that you are monitoring with automatically adds it to the cart so you should somehow pass the
	// cookies/client to here and then completely cut out the AddToCart request, otherwise using a faster endpoint to monitor would be better.
	// 2. AddToCart
	task.PublishEvent(enums.AddingToCart, enums.TaskUpdate, 30)
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
		task.Task.EventBus.PublishTaskEvent(enums.EncryptingCardInfo, -1, enums.TaskUpdate, cardInfo, task.Task.Task.ID)
	}()

	// 3. GetCartInfo
	task.PublishEvent(enums.GettingCartInfo, enums.TaskUpdate, 50)
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
	task.PublishEvent(enums.SettingShippingInfo, enums.TaskUpdate, 60)
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
	task.PublishEvent(enums.GettingBillingInfo, enums.TaskUpdate, 70)
	needToStop = task.WaitForEncryptedPaymentInfo()
	if needToStop {
		return
	}

	// * @silent: The piHash that this SetCreditCard is returning isn't needed but it may help with cancels in the future so it will take some testing during beta,
	// * but for now we should just comment it out

	// 7. SetCreditCard
	task.PublishEvent(enums.SettingBillingInfo, enums.TaskUpdate, 80)
	/* setCreditCard := false
	for !setCreditCard {
		needToStop := task.CheckForStop()
		if needToStop {
			return
		}
		setCreditCard = task.SetCreditCard()
		if !setCreditCard {
			time.Sleep(time.Duration(task.Task.Task.TaskDelay) * time.Millisecond)
		}
	} */

	// 8. SetPaymentInfo
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

	// 9. PlaceOrder
	task.PublishEvent(enums.CheckingOut, enums.TaskUpdate, 90)
	placedOrder := false
	var retries int
	status := enums.OrderStatusFailed
	for !placedOrder {
		needToStop := task.CheckForStop()
		if needToStop {
			return
		}
		if status == enums.OrderStatusDeclined || retries > common.MAX_RETRIES {
			break
		}
		placedOrder, status = task.PlaceOrder()
		if !placedOrder {
			retries++
			time.Sleep(time.Duration(task.Task.Task.TaskDelay) * time.Millisecond)
		}
	}

	endTime := time.Now()

	log.Println("STARTED AT: " + startTime.String())
	log.Println("  ENDED AT: " + endTime.String())
	log.Println("TIME TO CHECK OUT: ", endTime.Sub(startTime).Milliseconds())

	switch status {
	case enums.OrderStatusSuccess:
		task.PublishEvent(enums.CheckingOutSuccess, enums.TaskComplete, 100)
	case enums.OrderStatusDeclined:
		task.PublishEvent(enums.CardDeclined, enums.TaskComplete, 100)
	case enums.OrderStatusFailed:
		task.PublishEvent(fmt.Sprintf(enums.CheckingOutFailure, "Unknown error"), enums.TaskComplete, 100)
	}

	quantity := task.Task.Task.TaskQty
	if quantity > task.StockData.MaxQty {
		quantity = task.StockData.MaxQty
	}

	go util.ProcessCheckout(&util.ProcessCheckoutInfo{
		BaseTask:     task.Task,
		Success:      placedOrder,
		Status:       status,
		Content:      "",
		Embeds:       task.CreateWalmartEmbed(status, task.StockData.ImageURL),
		ItemName:     task.StockData.ProductName,
		ImageURL:     task.StockData.ImageURL,
		Sku:          task.StockData.SKU,
		Retailer:     enums.Walmart,
		Price:        task.StockData.Price,
		Quantity:     quantity,
		MsToCheckout: time.Since(startTime).Milliseconds(),
	})
}

// WaitForMonitor waits until the Monitor has sent the info to the task to continue
func (task *Task) WaitForMonitor() bool {

	for {
		needToStop := task.CheckForStop()
		if needToStop {
			return true
		}
		if task.StockData.OfferID != "" && task.StockData.SKU != "" {
			task.Task.HasStockData = true
			return false
		}
		time.Sleep(common.MS_TO_WAIT)
	}
}

func (task *Task) HandlePXCap(resp *http.Response, redirectURL string) bool {
	quit := make(chan bool)
	defer func() {
		quit <- true
		if r := recover(); r != nil {
			task.HandlePXCap(resp, redirectURL)
		}
	}()

	cancellationToken := util.CancellationToken{Cancel: false}
	go func() {
		for {
			select {
			case <-quit:
				return
			default:
				needToStop := task.CheckForStop()
				if needToStop {
					cancellationToken.Cancel = true
					return
				}
			}
			time.Sleep(common.MS_TO_WAIT)
		}
	}()

	task.PublishEvent(enums.BypassingPX, enums.TaskUpdate, -1)
	captchaURL := resp.Request.URL.String()
	if redirectURL != "" {
		captchaURL = BaseEndpoint + redirectURL[1:]
	}
	err := SetPXCapCookie(strings.ReplaceAll(captchaURL, "affil.", ""), &task.PXValues, task.Task.Proxy, &task.Task.Client, &cancellationToken)
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
	u, _ := url.Parse("https://www.walmart.com/")
	task.Task.Client.Jar.SetCookies(u, []*http.Cookie{{
		Name:     "com.wm.reflector",
		Value:    fmt.Sprintf(`"reflectorid:0000000000000000000000@lastupd:%d000@firstcreate:%d000"`, time.Now().Add(-10*time.Minute).Unix(), time.Now().Add(-20*24*time.Hour).Unix()),
		Path:     "/",
		Domain:   ".walmart.com",
		Expires:  time.Now().Add(10 * 365 * 24 * time.Hour),
		SameSite: http.SameSiteStrictMode,
	}})

	resp, _, err := util.MakeRequest(&util.Request{
		Client: task.Task.Client,
		Method: "GET",
		URL:    BlockedToBaseEndpoint,
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
		log.Println("Setup request 2 error: " + err.Error())
	}
	if strings.Contains(resp.Request.URL.String(), "blocked") {
		handled := task.HandlePXCap(resp, BaseEndpoint)
		return handled
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
	phaseStr, err := util.FindInString(body, "PIE.phase = ", ";")
	if err != nil || phaseStr == "" {
		return pieValues
	}
	phase, err := strconv.Atoi(phaseStr)
	if err != nil {
		return pieValues
	}

	pieValues = PIEValues{
		L:     L,
		E:     E,
		K:     K,
		KeyID: KeyID,
		Phase: phase,
	}

	return pieValues
}

// AddToCart sends a POST request to the AddToCartEndpoint with an AddToCartRequest body
func (task *Task) AddToCart() bool {
	addToCartResponse := AddToCartResponse{}
	quantity := task.Task.Task.TaskQty
	if quantity > task.StockData.MaxQty {
		quantity = task.StockData.MaxQty
	}

	data := AddToCartRequest{
		OfferID:               task.StockData.OfferID,
		Quantity:              quantity,
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
			{"referer", AddToCartReferer + "ip/" + task.StockData.SKU + "/sellers"},
		},
		RequestBodyStruct:  data,
		ResponseBodyStruct: &addToCartResponse,
	})
	if strings.Contains(resp.Request.URL.String(), "blocked") || (addToCartResponse.RedirectURL != "" && strings.Contains(addToCartResponse.RedirectURL, "blocked")) {
		handled := task.HandlePXCap(resp, addToCartResponse.RedirectURL)
		if handled {
			task.PublishEvent(enums.AddingToCart, enums.TaskUpdate, -1)
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
			task.PublishEvent(enums.GettingCartInfo, enums.TaskUpdate, -1)
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
			task.PublishEvent(enums.SettingCartInfo, enums.TaskUpdate, -1)
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
			{"content-length", fmt.Sprint(len(dataStr))},
			{"pragma", "no-cache"},
			{"cache-control", "no-cache"},
			{"accept", "application/json, text/javascript, */*; q=0.01"},
			{"inkiru_precedence", "false"},
			{"wm_cvv_in_session", "true"},
			{"user-agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/91.0.4472.77 Safari/537.36"},
			{"wm_vertical_id", "0"},
			{"content-type", "application/json"},
			{"origin", "https://www.walmart.com"},
			{"sec-fetch-site", "same-origin"},
			{"sec-fetch-mode", "cors"},
			{"sec-fetch-dest", "empty"},
			{"referer", SetShippingInfoReferer},
			{"accept-encoding", "gzip, deflate, br"},
			{"accept-language", "en-US,en;q=0.9"},
		},
		RequestBodyStruct:  data,
		ResponseBodyStruct: &setShippingInfoResponse,
	})
	if strings.Contains(resp.Request.URL.String(), "blocked") || (setShippingInfoResponse.RedirectURL != "" && strings.Contains(setShippingInfoResponse.RedirectURL, "blocked")) {
		handled := task.HandlePXCap(resp, setShippingInfoResponse.RedirectURL)
		if handled {
			task.PublishEvent(enums.SettingShippingInfo, enums.TaskUpdate, -1)
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
		time.Sleep(common.MS_TO_WAIT)
	}
}

// SetCreditCard sets the CreditCard and also returns the PiHash needed for SetPaymentInfo
func (task *Task) SetCreditCard() (bool, bool) {

	if !common.ValidCardType([]byte(task.Task.Profile.CreditCard.CardNumber), task.Task.Task.TaskRetailer) {
		return false, true
	}

	cardType := util.GetCardType([]byte(task.Task.Profile.CreditCard.CardNumber), task.Task.Task.TaskRetailer)

	setCreditCardResponse := SetCreditCardResponse{}
	data := Payment{
		EncryptedPan:   task.CardInfo.EncryptedPan,
		EncryptedCvv:   task.CardInfo.EncryptedCvv,
		IntegrityCheck: task.CardInfo.IntegrityCheck,
		KeyId:          task.CardInfo.KeyId,
		Phase:          task.CardInfo.Phase,
		State:          task.Task.Profile.BillingAddress.StateCode,
		PostalCode:     task.Task.Profile.BillingAddress.ZipCode,
		AddressLineOne: task.Task.Profile.BillingAddress.Address1,
		AddressLineTwo: task.Task.Profile.BillingAddress.Address2,
		City:           task.Task.Profile.BillingAddress.City,
		AddressType:    "RESIDENTIAL",
		FirstName:      task.Task.Profile.BillingAddress.FirstName,
		LastName:       task.Task.Profile.BillingAddress.LastName,
		ExpiryMonth:    task.Task.Profile.CreditCard.ExpMonth,
		ExpiryYear:     task.Task.Profile.CreditCard.ExpYear,
		Phone:          task.Task.Profile.PhoneNumber,
		CardType:       cardType,
		IsGuest:        true, // No login yet
	}
	dataStr, err := json.Marshal(data)
	if err != nil {
		log.Println("SetCreditCard Request Error: " + err.Error())
		return false, false
	}
	resp, _, err := util.MakeRequest(&util.Request{
		Client: task.Task.Client,
		Method: "POST",
		URL:    SetCreditCardEndpoint,
		RawHeaders: [][2]string{
			{"content-length", fmt.Sprint(len(dataStr))},
			{"pragma", "no-cache"},
			{"cache-control", "no-cache"},
			{"sec-ch-ua", `" Not;A Brand";v="99", "Google Chrome";v="91", "Chromium";v="91"`},
			{"accept", "application/json"},
			{"sec-ch-ua-mobile", "?0"},
			{"user-agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/91.0.4472.77 Safari/537.36"},
			{"content-type", "application/json"},
			{"origin", "https://www.walmart.com"},
			{"sec-fetch-site", "same-origin"},
			{"sec-fetch-mode", "cors"},
			{"sec-fetch-dest", "empty"},
			{"referer", SetCreditCardReferer},
			{"accept-encoding", "gzip, deflate, br"},
			{"accept-language", "en-US,en;q=0.9"},
		},
		RequestBodyStruct:  data,
		ResponseBodyStruct: &setCreditCardResponse,
	})

	if strings.Contains(resp.Request.URL.String(), "blocked") || (setCreditCardResponse.RedirectURL != "" && strings.Contains(setCreditCardResponse.RedirectURL, "blocked")) {
		handled := task.HandlePXCap(resp, setCreditCardResponse.RedirectURL)
		if handled {
			task.PublishEvent(enums.SettingBillingInfo, enums.TaskUpdate, -1)
		}
	}
	if err != nil {
		log.Println("SetCreditCard Request Error: " + err.Error())
		return false, false
	}

	switch resp.StatusCode {
	case 200:
		task.CardInfo.PiHash = setCreditCardResponse.PiHash
		task.CardInfo.PaymentType = setCreditCardResponse.PaymentType
	case 404:
		log.Printf("Not Found: %v\n", resp.Status)
	default:
		log.Printf("Unknown Code: %v\n", resp.StatusCode)
	}
	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		return true, false
	}
	return false, false
}

// SetPaymentInfo sets the payment info to prepare for placing an order
func (task *Task) SetPaymentInfo() (bool, bool) {

	if !common.ValidCardType([]byte(task.Task.Profile.CreditCard.CardNumber), task.Task.Task.TaskRetailer) {
		return false, true
	}

	cardType := util.GetCardType([]byte(task.Task.Profile.CreditCard.CardNumber), task.Task.Task.TaskRetailer)

	task.CardInfo.PaymentType = "CREDITCARD"
	setPaymentInfoResponse := SetPaymentInfoResponse{}
	data := SubmitPaymentRequest{
		[]SubmitPayment{{
			PaymentType:    task.CardInfo.PaymentType,
			CardType:       cardType,
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
		return false, false
	}
	resp, _, err := util.MakeRequest(&util.Request{
		Client: task.Task.Client,
		Method: "POST",
		URL:    SetPaymentInfoEndpoint,
		RawHeaders: [][2]string{
			{"content-length", fmt.Sprint(len(dataStr))},
			{"pragma", "no-cache"},
			{"cache-control", "no-cache"},
			{"sec-ch-ua", `" Not;A Brand";v="99", "Google Chrome";v="91", "Chromium";v="91"`},
			{"inkiru_precedence", "false"},
			{"sec-ch-ua-mobile", "?0"},
			{"user-agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/91.0.4472.77 Safari/537.36"},
			{"content-type", "application/json"},
			{"accept", "application/json"},
			{"wm_cvv_in_session", "true"},
			{"wm_vertical_id", "0"},
			{"origin", "https://www.walmart.com"},
			{"sec-fetch-site", "same-origin"},
			{"sec-fetch-mode", "cors"},
			{"sec-fetch-dest", "empty"},
			{"referer", SetPaymentInfoReferer},
			{"sec-fetch-user", "?1"},
			{"upgrade-insecure-requests", "1"},
			{"accept-encoding", "gzip, deflate, br"},
			{"accept-language", "en-US,en;q=0.9"},
		},
		RequestBodyStruct:  data,
		ResponseBodyStruct: &setPaymentInfoResponse,
	})
	if strings.Contains(resp.Request.URL.String(), "blocked") || (setPaymentInfoResponse.RedirectURL != "" && strings.Contains(setPaymentInfoResponse.RedirectURL, "blocked")) {
		handled := task.HandlePXCap(resp, setPaymentInfoResponse.RedirectURL)
		if handled {
			task.PublishEvent(enums.SettingBillingInfo, enums.TaskUpdate, -1)
		}
	}
	if err != nil {
		log.Println("SetPaymentInfo Request Error: " + err.Error())
		return false, false
	}

	switch resp.StatusCode {
	case 404:
		log.Printf("Not Found: %v\n", resp.Status)
	default:
		log.Printf("Unknown Code: %v\n", resp.StatusCode)
	}
	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		return true, false
	}
	return false, false
}

// PlaceOrder completes the checkout process
func (task *Task) PlaceOrder() (bool, enums.OrderStatus) {
	status := enums.OrderStatusFailed
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
		return false, status
	}
	resp, _, err := util.MakeRequest(&util.Request{
		Client: task.Task.Client,
		Method: "PUT",
		URL:    PlaceOrderEndpoint,
		RawHeaders: [][2]string{
			{"content-length", fmt.Sprint(len(dataStr))},
			{"pragma", "no-cache"},
			{"cache-control", "no-cache"},
			{"sec-ch-ua", `" Not;A Brand";v="99", "Google Chrome";v="91", "Chromium";v="91"`},
			{"accept", "application/json, text/javascript, */*; q=0.01"},
			{"inkiru_precedence", "false"},
			{"wm_cvv_in_session", "true"},
			{"user-agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/91.0.4472.77 Safari/537.36"},
			{"wm_vertical_id", "0"},
			{"content-type", "application/json"},
			{"origin", "https://www.walmart.com"},
			{"sec-fetch-site", "same-origin"},
			{"sec-fetch-mode", "cors"},
			{"sec-fetch-dest", "empty"},
			{"referer", PlaceOrderReferer},
			{"accept-encoding", "gzip, deflate, br"},
			{"accept-language", "en-US,en;q=0.9"},
			{"sec-ch-ua-mobile", "?0"},
		},
		RequestBodyStruct:  data,
		ResponseBodyStruct: &placeOrderResponse,
	})
	if strings.Contains(resp.Request.URL.String(), "blocked") || (placeOrderResponse.RedirectURL != "" && strings.Contains(placeOrderResponse.RedirectURL, "blocked")) {
		handled := task.HandlePXCap(resp, placeOrderResponse.RedirectURL)
		if handled {
			task.PublishEvent(enums.CheckingOut, enums.TaskUpdate, -1)
		}
	}
	if err != nil {
		log.Println("PlaceOrder Request Error: " + err.Error())
		return false, status
	}
	var success bool
	switch resp.StatusCode {
	case 400:
		status = enums.OrderStatusDeclined
	case 404:
		status = enums.OrderStatusFailed
		log.Printf("Not Found: %v\n", resp.Status)
	default:
		status = enums.OrderStatusFailed
		log.Printf("Unknown Code: %v\n", resp.StatusCode)
	}
	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		success = true
		status = enums.OrderStatusSuccess
	}

	return success, status
}
