package disney

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
	"backend.juicedbot.io/juiced.infrastructure/queries"
	"backend.juicedbot.io/juiced.sitescripts/base"
	"backend.juicedbot.io/juiced.sitescripts/util"
	"github.com/google/uuid"
)

// CreateDisneyTask takes a Task entity and turns it into a Disney Task
func CreateDisneyTask(task *entities.Task, profile entities.Profile, proxy entities.Proxy, eventBus *events.EventBus, email, password string) (Task, error) {
	disneyTask := Task{}

	disneyTask = Task{
		Task: base.Task{
			Task:     task,
			Profile:  profile,
			Proxy:    proxy,
			EventBus: eventBus,
		},
		AccountInfo: AccountInfo{
			Email:    email,
			Password: password,
		},
	}
	return disneyTask, nil
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
// 		3. AddtoCart
// 		4. ValidateCheckout
//		5. SubmitShippingInfo
// 		6. EstablishAppSession
// 		7. GetPaysheetAE
// 		8. GetCardToken
// 		9. PlaceOrder
func (task *Task) RunTask() {
	// If the function panics due to a runtime error, recover from it
	defer func() {
		if recover() != nil {
			task.Task.StopFlag = true
			task.PublishEvent(enums.TaskIdle, enums.TaskFail)
		}
		task.PublishEvent(enums.TaskIdle, enums.TaskComplete)
	}()

	client, err := util.CreateClient(task.Task.Proxy)
	if err != nil {
		return
	}
	task.Task.Client = client

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
			sessionMade = BecomeGuest(task.Task.Client)
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
	// 3. AddtoCart
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
	needToStop = task.CheckForStop()
	if needToStop {
		return
	}

	task.PublishEvent(enums.GettingCartInfo, enums.TaskUpdate)

	startTime := time.Now()
	// 4. ValidateCheckout
	validatedCheckout := false
	for !validatedCheckout {
		needToStop := task.CheckForStop()
		if needToStop {
			return
		}
		validatedCheckout = task.ValidateCheckout()
		if !validatedCheckout {
			time.Sleep(time.Duration(task.Task.Task.TaskDelay) * time.Millisecond)
		}
	}

	// 5. SubmitShippingInfo
	submittedShippingInfo := false
	for !validatedCheckout {
		needToStop := task.CheckForStop()
		if needToStop {
			return
		}
		submittedShippingInfo = task.SubmitShippingInfo()
		if !submittedShippingInfo {
			time.Sleep(time.Duration(task.Task.Task.TaskDelay) * time.Millisecond)
		}
	}

	// 6. EstablishAppSession
	establishedAppSession := false
	for !validatedCheckout {
		needToStop := task.CheckForStop()
		if needToStop {
			return
		}
		establishedAppSession = task.EstablishAppSession()
		if !establishedAppSession {
			time.Sleep(time.Duration(task.Task.Task.TaskDelay) * time.Millisecond)
		}
	}

	// 7. GetPaysheetAE
	gotPaymentAE := false
	for !validatedCheckout {
		needToStop := task.CheckForStop()
		if needToStop {
			return
		}
		gotPaymentAE = task.GetPaysheetAE()
		if !gotPaymentAE {
			time.Sleep(time.Duration(task.Task.Task.TaskDelay) * time.Millisecond)
		}
	}

	// 8. GetCardToken
	gotCardToken := false
	for !validatedCheckout {
		needToStop := task.CheckForStop()
		if needToStop {
			return
		}
		gotCardToken = task.GetCardToken()
		if !gotCardToken {
			time.Sleep(time.Duration(task.Task.Task.TaskDelay) * time.Millisecond)
		}
	}

	task.PublishEvent(enums.CheckingOut, enums.TaskUpdate)
	// 9. PlaceOrder
	placedOrder := false
	doNotRetry := false
	status := enums.OrderStatusFailed
	for !placedOrder {
		needToStop := task.CheckForStop()
		if needToStop || doNotRetry {
			return
		}
		if status == enums.OrderStatusDeclined {
			break
		}
		placedOrder, doNotRetry, status = task.PlaceOrder(startTime)
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
		task.PublishEvent(enums.CheckoutFailed, enums.TaskComplete)
	case enums.OrderStatusFailed:
		task.PublishEvent(enums.CheckoutFailed, enums.TaskComplete)
	}

}

func (task *Task) Login() bool {
	resp, body, err := util.MakeRequest(&util.Request{
		Client: task.Task.Client,
		Method: "GET",
		URL:    BaseEndpoint,
		RawHeaders: http.RawHeader{
			{"sec-ch-ua", `" Not;A Brand";v="99", "Google Chrome";v="91", "Chromium";v="91"`},
			{"sec-ch-ua-mobile", `?0`},
			{"upgrade-insecure-requests", `1`},
			{"user-agent", `Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/91.0.4472.124 Safari/537.36`},
			{"accept", `text/html,application/xhtml+xml,application/xml;q=0.9,image/avif,image/webp,image/apng,*/*;q=0.8,application/signed-exchange;v=b3;q=0.9`},
			{"sec-fetch-site", `none`},
			{"sec-fetch-mode", `navigate`},
			{"sec-fetch-user", `?1`},
			{"sec-fetch-dest", `document`},
			{"accept-encoding", `gzip, deflate, br`},
			{"accept-language", `en-US,en;q=0.9`},
		},
	})
	if err != nil || resp.StatusCode != 200 {
		fmt.Println(err)
		return false
	}
	clientID, err := common.FindInString(body, `"clientID": "`, `"`)
	if err != nil {
		return false
	}
	oidClientID, err := common.FindInString(body, `"oidClientID": "`, `"`)
	if err != nil {
		return false
	}

	resp, err = task.Task.Client.Post("https://registerdisney.go.com/jgc/v6/client/DCP-DISNEYSTORE.WEB-PROD/api-key?langPref=en-US", "", nil)
	if err != nil || resp.StatusCode != 200 {
		return false
	}
	apiKey1 := resp.Header.Get("Api-Key")
	if apiKey1 == "" {
		return false
	}

	loginRequest := LoginRequest{
		Loginvalue: task.AccountInfo.Email,
		Password:   task.AccountInfo.Password,
	}

	correlationID := uuid.New().String()
	conversationId := uuid.New().String()
	currentTime := time.Now().UTC()

	loginResponse := LoginResponse{}
	resp, _, err = util.MakeRequest(&util.Request{
		Client: task.Task.Client,
		Method: "POST",
		URL:    FirstLoginEndpoint,
		RawHeaders: http.RawHeader{
			{"correlation-id", correlationID},
			{"sec-ch-ua", `" Not;A Brand";v="99", "Google Chrome";v="91", "Chromium";v="91"`},
			{"x-requested-with", "XMLHttpRequest"},
			{"sec-ch-ua-mobile", "?0"},
			{"authorization", "APIKEY " + apiKey1},
			{"content-type", "application/json"},
			{"user-agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/91.0.4472.106 Safari/537.36"},
			{"conversation-id", conversationId},
			{"device-id", "null"},
			{"g-recaptcha-token", ""},
			{"expires", "-1"},
			{"accept", "*/*"},
			{"origin", "https://cdn.registerdisney.go.com"},
			{"sec-fetch-site", "same-origin"},
			{"sec-fetch-site", "cors"},
			{"sec-fetch-dest", "empty"},
			{"referer", "https://cdn.registerdisney.go.com/"},
			{"accept-encoding", "gzip, deflate"},
			{"accept-language", "en-US,en;q=0.9"},
		},
		RequestBodyStruct:  loginRequest,
		ResponseBodyStruct: &loginResponse,
	})
	if err != nil {
		fmt.Println(err)
		return false
	}

	apiKey2 := resp.Header.Get("api-key")
	if apiKey2 == "" {
		return false
	}
	deviceID := resp.Header.Get("device-id")
	if deviceID == "" {
		return false
	}

	loginResponse.Data.Token.Created = currentTime.Format("2006-01-02T15:04:05.000Z")
	loginResponse.Data.Token.LastChecked = currentTime.Format("2006-01-02T15:04:05.000Z")
	loginResponse.Data.Token.Expires = currentTime.Add(time.Hour * 24).Format("2006-01-02T15:04:05.000Z")
	loginResponse.Data.Token.RefreshExpires = currentTime.Add((time.Hour * 24) * 180).Format("2006-01-02T15:04:05.000Z")
	loginResponse.Data.Token.BlueCookie = nil
	jsonBytes, _ := json.Marshal(loginResponse.Data.Token)
	encryptedJson := base64.StdEncoding.EncodeToString(jsonBytes)

	parsedURL, _ := url.Parse(SecondLoginEndpoint)
	task.Task.Client.Jar.SetCookies(parsedURL, []*http.Cookie{
		{
			Name:  "DCP-DISNEYSTORE.WEB-PROD.api",
			Value: apiKey2,
		},
		{
			Name:  "device_2a9b58d7",
			Value: deviceID,
		},
		{
			Name:  "DCP-DISNEYSTORE.WEB-PROD.ts",
			Value: time.Now().UTC().Format("2006-01-02T15:04:05.000Z"),
		},
		{
			Name:  "DCP-DISNEYSTORE.WEB-PROD.token",
			Value: "5=" + encryptedJson + "|",
		},
		{
			Name:  "DCP-DISNEYSTORE.WEB-PROD-ac",
			Value: "XUS",
		},
		{
			Name:  "DCP-DISNEYSTORE.WEB-PROD.idn",
			Value: "00" + RandomString("abcdef0123456789", 8),
		},
		{
			Name:  "SWID",
			Value: loginResponse.Data.Token.Swid,
		},
		{
			Name:  "SWID_NT",
			Value: "0",
		},
	})

	resp, _, err = util.MakeRequest(&util.Request{
		Client: task.Task.Client,
		Method: "POST",
		URL:    SecondLoginEndpoint,
		RawHeaders: http.RawHeader{
			{"content-length", `0`},
			{"pragma", `no-cache`},
			{"cache-control", `no-cache`},
			{"sec-ch-ua", `" Not;A Brand";v="99", "Google Chrome";v="91", "Chromium";v="91"`},
			{"sec-ch-ua-mobile", `?0`},
			{"user-agent", `Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/91.0.4472.124 Safari/537.36`},
			{"x-oid-client-id", oidClientID},
			{"accept", `*/*`},
			{"x-requested-with", `XMLHttpRequest`},
			{"access_token", loginResponse.Data.Token.AccessToken},
			{"x-cc-client-id", clientID},
			{"origin", BaseEndpoint},
			{"sec-fetch-site", `same-origin`},
			{"sec-fetch-mode", `cors`},
			{"sec-fetch-dest", `empty`},
			{"referer", BaseEndpoint + "/"},
			{"accept-encoding", `gzip, deflate, br`},
			{"accept-language", `en-US,en;q=0.9`},
		},
	})
	if err != nil || resp.StatusCode != 200 {
		fmt.Println(err)
		return false
	}

	return true
}

// WaitForMonitor waits until the Monitor has sent the info to the task to continue
func (task *Task) WaitForMonitor() bool {
	for {
		needToStop := task.CheckForStop()
		if needToStop {
			return true
		}
		if task.StockData.PID != "" {
			return false
		}
		time.Sleep(1 * time.Millisecond)
	}
}

func (task *Task) AddToCart() bool {
	data := []byte(util.CreateParams(map[string]string{
		"pid":      task.StockData.VID,
		"quantity": fmt.Sprint(task.Task.Task.TaskQty),
	}))
	addToCartResponse := AddToCartResponse{}
	_, _, err := util.MakeRequest(&util.Request{
		Client: task.Task.Client,
		Method: "POST",
		URL:    AddToCartEndpoint,
		RawHeaders: http.RawHeader{
			{"content-length", fmt.Sprint(len(data))},
			{"sec-ch-ua", `" Not;A Brand";v="99", "Google Chrome";v="91", "Chromium";v="91"`},
			{"accept", "application/json, text/javascript, */*; q=0.01"},
			{"x-requested-with", "XMLHttpRequest"},
			{"sec-ch-ua-mobile", "?0"},
			{"origin", BaseEndpoint},
			{"user-agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/91.0.4472.106 Safari/537.36"},
			{"content-type", "application/x-www-form-urlencoded; charset=UTF-8"},
			{"sec-fetch-site", "same-origin"},
			{"sec-fetch-site", "cors"},
			{"sec-fetch-dest", "empty"},
			{"referer", task.StockData.ItemURL},
			{"accept-encoding", "gzip, deflate"},
			{"accept-language", "en-US,en;q=0.9"},
		},
		Data:               data,
		ResponseBodyStruct: &addToCartResponse,
	})
	if err != nil {
		fmt.Println(err)
		return false
	}

	return addToCartResponse.Message == "Product added to bag"
}

func (task *Task) ValidateCheckout() bool {
	resp, _, err := util.MakeRequest(&util.Request{
		Client: task.Task.Client,
		Method: "GET",
		URL:    ValidateCheckoutEndpoint,
		RawHeaders: http.RawHeader{
			{"sec-ch-ua", `" Not;A Brand";v="99", "Google Chrome";v="91", "Chromium";v="91"`},
			{"accept", "application/json, text/javascript, */*; q=0.01"},
			{"x-requested-with", "XMLHttpRequest"},
			{"sec-ch-ua-mobile", "?0"},
			{"user-agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/91.0.4472.106 Safari/537.36"},
			{"sec-fetch-site", "same-origin"},
			{"sec-fetch-site", "cors"},
			{"sec-fetch-dest", "empty"},
			{"referer", CartEndpoint},
			{"accept-encoding", "gzip, deflate"},
			{"accept-language", "en-US,en;q=0.9"},
		},
	})
	if err != nil || resp.StatusCode != 200 {
		fmt.Println(err)
		return false
	}

	return true
}

func (task *Task) SubmitShippingInfo() bool {
	data := []byte(util.CreateParams(map[string]string{
		"originalShipmentUUID": task.TaskInfo.ShipmentUUID,
		"shipmentUUID":         task.TaskInfo.ShipmentUUID,
		"emojiValidation":      `(?:[\u2700-\u27bf]|(?:\ud83c[\udde6-\uddff]){2}|[\ud800-\udbff][\udc00-\udfff]|[\u0023-\u0039]\ufe0f?\u20e3|\u3299|\u3297|\u303d|\u3030|\u24c2|\ud83c[\udd70-\udd71]|\ud83c[\udd7e-\udd7f]|\ud83c\udd8e|\ud83c[\udd91-\udd9a]|\ud83c[\udde6-\uddff]|\ud83c[\ude01-\ude02]|\ud83c\ude1a|\ud83c\ude2f|\ud83c[\ude32-\ude3a]|\ud83c[\ude50-\ude51]|\u203c|\u2049|[\u25aa-\u25ab]|\u25b6|\u25c0|[\u25fb-\u25fe]|\u00a9|\u00ae|\u2122|\u2139|\ud83c\udc04|[\u2600-\u26FF]|\u2b05|\u2b06|\u2b07|\u2b1b|\u2b1c|\u2b50|\u2b55|\u231a|\u231b|\u2328|\u23cf|[\u23e9-\u23f3]|[\u23f8-\u23fa]|\ud83c\udccf|\u2934|\u2935|[\u2190-\u21ff]|\uFE0F)`,
		"shipping-countries":   `[{"id":"US","fields":{"firstName":{"maxLength":"15","mandatory":"true"},"lastName":{"maxLength":"25","mandatory":"true"},"address1":{"maxLength":"32","mandatory":"true"},"address2":{"maxLength":"32","mandatory":"false"},"city":{"maxLength":"25","mandatory":"true"},"state":{"maxLength":"","mandatory":"true"},"postalCode":{"maxLength":"10","mandatory":"true"},"country":{"maxLength":"","mandatory":"true"}}},{"id":"CA","fields":{"firstName":{"maxLength":"15","mandatory":"true"},"lastName":{"maxLength":"25","mandatory":"true"},"address1":{"maxLength":"32","mandatory":"true"},"address2":{"maxLength":"32","mandatory":"false"},"city":{"maxLength":"25","mandatory":"true"},"postalCode":{"maxLength":"10","mandatory":"true"},"country":{"maxLength":"","mandatory":"true"},"state":{"maxLength":"","mandatory":"true"}}},{"id":"default","fields":{"firstName":{"maxLength":"15","mandatory":"true"},"lastName":{"maxLength":"25","mandatory":"true"},"address1":{"maxLength":"32","mandatory":"true"},"address2":{"maxLength":"32","mandatory":"false"},"city":{"maxLength":"25","mandatory":"true"},"postalCode":{"maxLength":"10","mandatory":"false"},"country":{"maxLength":"","mandatory":"true"},"state":{"maxLength":"","mandatory":"false"}}}]`,
		"dwfrm_shipping_shippingAddress_addressFields_country":          task.Task.Profile.ShippingAddress.CountryCode,
		"dwfrm_shipping_shippingAddress_addressFields_firstName":        task.Task.Profile.ShippingAddress.FirstName,
		"dwfrm_shipping_shippingAddress_addressFields_lastName":         task.Task.Profile.ShippingAddress.LastName,
		"dwfrm_shipping_shippingAddress_addressFields_address1":         task.Task.Profile.ShippingAddress.Address1,
		"dwfrm_shipping_shippingAddress_addressFields_address2":         task.Task.Profile.ShippingAddress.Address2,
		"dwfrm_shipping_shippingAddress_addressFields_postalCode":       task.Task.Profile.ShippingAddress.ZipCode,
		"dwfrm_shipping_shippingAddress_addressFields_city":             task.Task.Profile.ShippingAddress.City,
		"dwfrm_shipping_shippingAddress_addressFields_states_stateCode": task.Task.Profile.ShippingAddress.StateCode,
		"dwfrm_shipping_shippingAddress_addressFields_phone":            task.Task.Profile.PhoneNumber,
		"appSessionID":                         ``,
		"dwfrm_shipping_shippingAddress_email": task.Task.Profile.Email,
		"shippingMethod":                       task.TaskInfo.ShippingMethod,
		"csrf_token":                           task.TaskInfo.CsrfToken,
	}))

	resp, _, err := util.MakeRequest(&util.Request{
		Client: task.Task.Client,
		Method: "POST",
		URL:    SubmitShippingInfoEndpoint,
		RawHeaders: http.RawHeader{
			{"content-length", fmt.Sprint(len(data))},
			{"sec-ch-ua", `" Not;A Brand";v="99", "Google Chrome";v="91", "Chromium";v="91"`},
			{"accept", "application/json, text/javascript, */*; q=0.01"},
			{"x-requested-with", "XMLHttpRequest"},
			{"sec-ch-ua-mobile", "?0"},
			{"origin", BaseEndpoint},
			{"user-agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/91.0.4472.106 Safari/537.36"},
			{"content-type", "application/x-www-form-urlencoded; charset=UTF-8"},
			{"sec-fetch-site", "same-origin"},
			{"sec-fetch-site", "cors"},
			{"sec-fetch-dest", "empty"},
			{"referer", CheckoutEndpoint + "?stage=shipping"},
			{"accept-encoding", "gzip, deflate"},
			{"accept-language", "en-US,en;q=0.9"},
		},
		Data: data,
	})
	if err != nil || resp.StatusCode != 200 {
		fmt.Println(err)
		return false
	}

	return true
}

func (task *Task) EstablishAppSession() bool {
	establishAppSessionResponse := EstablishAppSessionResponse{}
	resp, _, err := util.MakeRequest(&util.Request{
		Client: task.Task.Client,
		Method: "GET",
		URL:    EstablishAppSessionEndpoint,
		RawHeaders: http.RawHeader{
			{"sec-ch-ua", `" Not;A Brand";v="99", "Google Chrome";v="91", "Chromium";v="91"`},
			{"sec-ch-ua-mobile", `?0`},
			{"upgrade-insecure-requests", `1`},
			{"user-agent", `Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/91.0.4472.124 Safari/537.36`},
			{"accept", `text/html,application/xhtml+xml,application/xml;q=0.9,image/avif,image/webp,image/apng,*/*;q=0.8,application/signed-exchange;v=b3;q=0.9`},
			{"sec-fetch-site", `cross-site`},
			{"sec-fetch-mode", `navigate`},
			{"sec-fetch-dest", `iframe`},
			{"referer", CheckoutEndpoint + `?stage=shipping`},
			{"accept-encoding", `gzip, deflate, br`},
			{"accept-language", `en-US,en;q=0.9`},
		},
		ResponseBodyStruct: &establishAppSessionResponse,
	})
	if err != nil || resp.StatusCode != 200 {
		fmt.Println(err)
		return false
	}
	task.PaymentData = establishAppSessionResponse.PaymentData
	task.TaskInfo.Total = establishAppSessionResponse.Order.Totals.GrandTotalValue.Value

	return true
}

func (task *Task) GetPaysheetAE() bool {
	resp, _, err := util.MakeRequest(&util.Request{
		Client: task.Task.Client,
		Method: "GET",
		URL:    fmt.Sprintf(GetPaysheetAEEndpoint, task.PaymentData.Config.Session),
		RawHeaders: http.RawHeader{
			{"sec-ch-ua", `" Not;A Brand";v="99", "Google Chrome";v="91", "Chromium";v="91"`},
			{"x-disney-paysheet-client", task.PaymentData.Config.Client},
			{"accept-language", `en-us`},
			{"sec-ch-ua-mobile", `?0`},
			{"authorization", `BEARER ` + task.PaymentData.Config.Oauth},
			{"x-disney-paysheet-session", task.PaymentData.Config.Session},
			{"x-client-conversation-id", task.PaymentData.Config.ConversationID},
			{"accept", `application/json, text/plain, */*`},
			{"x-disney-paysheet-interface", `web-embedded`},
			{"user-agent", `Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/91.0.4472.124 Safari/537.36`},
			{"__h", task.PaymentData.Config.Hmac},
			{"sec-fetch-site", `same-origin`},
			{"sec-fetch-mode", `cors`},
			{"sec-fetch-dest", `empty`},
			{"referer", `https://paymentsheet.wdprapps.disney.com/embedded/web?fromOrigin=https:%2F%2Fwww.shopdisney.com`},
			{"accept-encoding", `gzip, deflate, br`},
		},
	})
	if err != nil || resp.StatusCode != 200 {
		fmt.Println(err)
		return false
	}

	task.PaymentData.Config.PaymentAE = resp.Header.Get("x-disney-paysheet-ae")

	return true
}

func (task *Task) GetCardToken() bool {
	getCardTokenRequest := GetCardTokenRequest{
		Pan:  task.Task.Profile.CreditCard.CardNumber,
		Type: "CREDIT_CARD",
	}
	getCardTokenResponse := GetCardTokenResponse{}
	resp, _, err := util.MakeRequest(&util.Request{
		Client: task.Task.Client,
		Method: "POST",
		URL:    GetCardTokenEndpoint,
		RawHeaders: http.RawHeader{
			{"pragma", `no-cache`},
			{"cache-control", `no-cache`},
			{"sec-ch-ua", `" Not;A Brand";v="99", "Google Chrome";v="91", "Chromium";v="91"`},
			{"x-disney-paysheet-ae", task.PaymentData.Config.PaymentAE},
			{"x-disney-paysheet-client", task.PaymentData.Config.Client},
			{"accept-language", `en-us`},
			{"sec-ch-ua-mobile", `?0`},
			{"authorization", `BEARER ` + task.PaymentData.Config.Oauth},
			{"x-disney-paysheet-purl", `ent`},
			{"content-type", `application/json`},
			{"x-client-conversation-id", task.PaymentData.Config.ConversationID},
			{"accept", `application/json, text/plain, */*`},
			{"x-disney-paysheet-interface", `web-embedded`},
			{"user-agent", `Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/91.0.4472.124 Safari/537.36`},
			{"x-disney-paysheet-session", task.PaymentData.Config.Session},
			{"origin", `https://paymentsheet.wdprapps.disney.com`},
			{"sec-fetch-site", `same-origin`},
			{"sec-fetch-mode", `cors`},
			{"sec-fetch-dest", `empty`},
			{"referer", `https://paymentsheet.wdprapps.disney.com/embedded/web?fromOrigin=https:%2F%2Fwww.shopdisney.com`},
			{"accept-encoding", `gzip, deflate, br`},
		},
		RequestBodyStruct:  getCardTokenRequest,
		ResponseBodyStruct: &getCardTokenResponse,
	})
	if err != nil || resp.StatusCode != 200 {
		fmt.Println(err)
		return false
	}

	task.PaymentData.CardInfo = getCardTokenResponse

	return true
}

// PlaceOrder
func (task *Task) PlaceOrder(startTime time.Time) (bool, bool, enums.OrderStatus) {
	status := enums.OrderStatusFailed
	if !common.ValidCardType([]byte(task.Task.Profile.CreditCard.CardNumber), task.Task.Task.TaskRetailer) {
		return false, true, status
	}

	PlaceOrderRequest := PlaceOrderRequest{
		Cards: []Cards{
			{
				Securitycode:      task.Task.Profile.CreditCard.CVV,
				Cardproductname:   task.PaymentData.CardInfo.Cardproductname,
				Isdisneyvisa:      task.PaymentData.CardInfo.Isdisneyvisa,
				Cardtoken:         task.PaymentData.CardInfo.Token,
				Carddisplaynumber: task.PaymentData.CardInfo.Masked[len(task.PaymentData.CardInfo.Masked)-8:],
				Expiration:        task.Task.Profile.CreditCard.ExpMonth + task.Task.Profile.CreditCard.ExpYear,
				Cardholdername:    task.Task.Profile.CreditCard.CardholderName,
				Profilename:       Profilename{},
				Profilephone:      Profilephone{},
				Billingaddress: Billingaddress{
					Country:    task.Task.Profile.BillingAddress.CountryCode,
					Line1:      task.Task.Profile.BillingAddress.Address1,
					Line2:      task.Task.Profile.BillingAddress.Address2,
					City:       task.Task.Profile.BillingAddress.City,
					State:      task.Task.Profile.BillingAddress.StateCode,
					Postalcode: task.Task.Profile.BillingAddress.ZipCode,
				},
				Savetoprofile:   false,
				Iswalletdefault: false,
				Iswalletedit:    false,
				Issuer:          task.PaymentData.CardInfo.Issuer,
				Type:            task.PaymentData.CardInfo.Type,
				Isstoredcard:    false,
				Amount:          task.TaskInfo.Total,
			},
		},
		Fraudparameters: []Fraudparameters{
			{
				Group: "identification_info",
				Params: Params{
					EmailAddress: task.Task.Profile.Email,
				},
			},
			{
				Group: "billing_info",
				Params: Params{
					DayPhone: task.Task.Profile.PhoneNumber,
				},
			},
		},
	}
	data, _ := json.Marshal(PlaceOrderRequest)
	placeOrderResponse := PlaceOrderResponse{}
	resp, _, err := util.MakeRequest(&util.Request{
		Client: task.Task.Client,
		Method: "POST",
		URL:    fmt.Sprintf(PlaceOrderEndpoint, task.PaymentData.Config.Session),
		RawHeaders: http.RawHeader{
			{"content-length", fmt.Sprint(len(data))},
			{"pragma", `no-cache`},
			{"cache-control", `no-cache`},
			{"sec-ch-ua", `" Not;A Brand";v="99", "Google Chrome";v="91", "Chromium";v="91"`},
			{"x-disney-paysheet-ae", task.PaymentData.Config.PaymentAE},
			{"x-disney-paysheet-client", task.PaymentData.Config.Client},
			{"accept-language", `en-us`},
			{"sec-ch-ua-mobile", `?0`},
			{"authorization", `BEARER ` + task.PaymentData.Config.Oauth},
			{"x-disney-paysheet-purl", `ent`},
			{"content-type", `application/json`},
			{"x-client-conversation-id", task.PaymentData.Config.ConversationID},
			{"accept", `application/json, text/plain, */*`},
			{"x-disney-paysheet-interface", `web-embedded`},
			{"user-agent", `Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/91.0.4472.124 Safari/537.36`},
			{"x-disney-paysheet-session", task.PaymentData.Config.Session},
			{"origin", `https://paymentsheet.wdprapps.disney.com`},
			{"sec-fetch-site", `same-origin`},
			{"sec-fetch-mode", `cors`},
			{"sec-fetch-dest", `empty`},
			{"referer", `https://paymentsheet.wdprapps.disney.com/embedded/web?fromOrigin=https:%2F%2Fwww.shopdisney.com`},
			{"accept-encoding", `gzip, deflate, br`},
		},
		Data:               data,
		ResponseBodyStruct: &placeOrderResponse,
	})
	if err != nil || resp.StatusCode != 200 {
		fmt.Println(err)
		return false, false, status
	}

	// Need to know what a successful order looks like
	var success bool
	switch placeOrderResponse.Suggestederrorkey {
	case "d_credit_card":
		status = enums.OrderStatusDeclined
		return false, false, status
	default:
		status = enums.OrderStatusSuccess
		success = true
	}

	_, user, err := queries.GetUserInfo()
	if err != nil {
		fmt.Println("Could not get user info")
		return false, false, status
	}

	util.ProcessCheckout(util.ProcessCheckoutInfo{
		BaseTask:     task.Task,
		Success:      success,
		Content:      "",
		Embeds:       task.CreateDisneyEmbed(status, task.StockData.ImageURL),
		UserInfo:     user,
		ItemName:     task.StockData.ProductName,
		Sku:          task.StockData.PID,
		Retailer:     enums.Disney,
		Price:        task.TaskInfo.Total,
		Quantity:     task.Task.Task.TaskQty,
		MsToCheckout: time.Since(startTime).Milliseconds(),
	})

	return true, false, status
}
