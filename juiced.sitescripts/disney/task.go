package disney

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"log"
	"net/url"
	"strings"
	"time"

	"backend.juicedbot.io/juiced.client/http"
	"backend.juicedbot.io/juiced.infrastructure/common"
	"backend.juicedbot.io/juiced.infrastructure/common/captcha"
	"backend.juicedbot.io/juiced.infrastructure/common/entities"
	"backend.juicedbot.io/juiced.infrastructure/common/enums"
	"backend.juicedbot.io/juiced.infrastructure/common/events"
	"backend.juicedbot.io/juiced.sitescripts/base"
	"backend.juicedbot.io/juiced.sitescripts/util"
	"github.com/google/uuid"
)

// CreateDisneyTask takes a Task entity and turns it into a Disney Task
func CreateDisneyTask(task *entities.Task, profile entities.Profile, proxyGroup *entities.ProxyGroup, eventBus *events.EventBus, taskType enums.TaskType, email, password string) (Task, error) {
	disneyTask := Task{
		Task: base.Task{
			Task:       task,
			Profile:    profile,
			ProxyGroup: proxyGroup,
			EventBus:   eventBus,
		},
		TaskType: taskType,
		AccountInfo: AccountInfo{
			Email:    email,
			Password: password,
		},
	}
	if proxyGroup != nil {
		disneyTask.Task.Proxy = util.RandomLeastUsedProxy(proxyGroup.Proxies)
	}
	return disneyTask, nil
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
	if task.Task.StopFlag {
		task.PublishEvent(enums.TaskIdle, enums.TaskStop, 0)
		return true
	}
	return false
}

// RunTask is the script driver that calls all the individual requests
// Function order:
// 		1. Login / Become a guest
// 		2. WaitForMonitor
// 		3. AddtoCart
// 		4. GetCheckoutInfo
// 		5. ValidateCheckout
//		6. SubmitShippingInfo
// 		7. EstablishAppSession
// 		8. GetPaysheetAE
// 		9. GetCardToken
// 		10. PlaceOrder
func (task *Task) RunTask() {
	// If the function panics due to a runtime error, recover from it
	defer func() {
		if r := recover(); r != nil {
			task.PublishEvent(fmt.Sprintf(enums.TaskFailed, r), enums.TaskFail, 0)
		} else {
			if !strings.Contains(task.Task.Task.TaskStatus, strings.ReplaceAll(enums.TaskIdle, " %s", "")) &&
				!strings.Contains(task.Task.Task.TaskStatus, strings.ReplaceAll(enums.CheckingOutFailure, " %s", "")) &&
				!strings.Contains(task.Task.Task.TaskStatus, strings.ReplaceAll(enums.CardDeclined, " %s", "")) &&
				!strings.Contains(task.Task.Task.TaskStatus, strings.ReplaceAll(enums.CheckedOut, " %s", "")) &&
				!strings.Contains(task.Task.Task.TaskStatus, strings.ReplaceAll(enums.TaskFailed, " %s", "")) {
				task.PublishEvent(enums.TaskIdle, enums.TaskStop, 0)
			}
		}
		task.Task.StopFlag = true
	}()
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

	// 1. Login / Become a guest
	sessionMade := false
	for !sessionMade {
		needToStop := task.CheckForStop()
		if needToStop {
			return
		}
		switch task.TaskType {
		case enums.TaskTypeAccount:
			if task.Task.Task.TaskStatus != enums.LoggingIn {
				task.PublishEvent(enums.LoggingIn, enums.TaskStart, 10)
			}
			sessionMade = task.Login()

		case enums.TaskTypeGuest:
			if task.Task.Task.TaskStatus != enums.SettingUp {
				task.PublishEvent(enums.SettingUp, enums.TaskStart, 10)
			}
			sessionMade = BecomeGuest(&task.Task.Client)
		}

		if !sessionMade {
			time.Sleep(time.Duration(task.Task.Task.TaskDelay) * time.Millisecond)
		}
	}

	newAbck := false
	for !newAbck {
		needToStop := task.CheckForStop()
		if needToStop {
			return
		}
		err := util.NewAbck(&task.Task.Client, BaseEndpoint, BaseEndpoint, AkamaiEndpoint)
		if err == nil {
			newAbck = true
		}
		if !newAbck {
			time.Sleep(time.Duration(task.Task.Task.TaskDelay) * time.Millisecond)
		}
	}

	task.PublishEvent(enums.WaitingForMonitor, enums.TaskUpdate, 15)
	// 2. WaitForMonitor
	needToStop := task.WaitForMonitor()
	if needToStop {
		return
	}

	task.PublishEvent(enums.AddingToCart, enums.TaskUpdate, 20)
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

	startTime := time.Now()

	task.PublishEvent(enums.GettingCartInfo, enums.TaskUpdate, 40)
	// 4. GetCheckoutInfo
	gotCheckoutInfo := false
	for !gotCheckoutInfo {
		needToStop := task.CheckForStop()
		if needToStop {
			return
		}
		gotCheckoutInfo = task.GetCheckoutInfo()
		if !gotCheckoutInfo {
			time.Sleep(time.Duration(task.Task.Task.TaskDelay) * time.Millisecond)
		}
	}

	task.PublishEvent(enums.SettingCartInfo, enums.TaskUpdate, 50)
	// 5. ValidateCheckout
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

	task.PublishEvent(enums.SettingShippingInfo, enums.TaskUpdate, 60)
	// 6. SubmitShippingInfo
	submittedShippingInfo := false
	for !submittedShippingInfo {
		needToStop := task.CheckForStop()
		if needToStop {
			return
		}
		submittedShippingInfo = task.SubmitShippingInfo()
		if !submittedShippingInfo {
			time.Sleep(time.Duration(task.Task.Task.TaskDelay) * time.Millisecond)
		}
	}

	task.PublishEvent(enums.GettingBillingInfo, enums.TaskUpdate, 65)
	// 7. EstablishAppSession
	establishedAppSession := false
	for !establishedAppSession {
		needToStop := task.CheckForStop()
		if needToStop {
			return
		}
		establishedAppSession = task.EstablishAppSession()
		if !establishedAppSession {
			time.Sleep(time.Duration(task.Task.Task.TaskDelay) * time.Millisecond)
		}
	}

	task.PublishEvent(enums.SettingBillingInfo, enums.TaskUpdate, 70)
	// 8. GetPaysheetAE
	gotPaymentAE := false
	for !gotPaymentAE {
		needToStop := task.CheckForStop()
		if needToStop {
			return
		}
		gotPaymentAE = task.GetPaysheetAE()
		if !gotPaymentAE {
			time.Sleep(time.Duration(task.Task.Task.TaskDelay) * time.Millisecond)
		}
	}

	task.PublishEvent(enums.GettingOrderInfo, enums.TaskUpdate, 80)
	// 9. GetCardToken
	gotCardToken := false
	for !gotCardToken {
		needToStop := task.CheckForStop()
		if needToStop {
			return
		}
		gotCardToken = task.GetCardToken()
		if !gotCardToken {
			time.Sleep(time.Duration(task.Task.Task.TaskDelay) * time.Millisecond)
		}
	}

	task.PublishEvent(enums.CheckingOut, enums.TaskUpdate, 90)
	// 10. PlaceOrder
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
		task.PublishEvent(enums.CheckedOut, enums.TaskComplete, 100)
	case enums.OrderStatusDeclined:
		task.PublishEvent(enums.CardDeclined, enums.TaskComplete, 100)
	case enums.OrderStatusFailed:
		task.PublishEvent(enums.CheckoutFailed, enums.TaskComplete, 100)
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

	for _, cookie := range task.Task.Client.Jar.Cookies(ParsedBase) {
		if cookie.Name == "_abck" {
			validator, _ := util.FindInString(cookie.Value, "~", "~")
			if validator == "-1" {
				err := util.NewAbck(&task.Task.Client, task.StockData.ItemURL, BaseEndpoint, AkamaiEndpoint)
				if err != nil {
					return false
				}
			}
		}
	}

	resp, _, err = util.MakeRequest(&util.Request{
		Client: task.Task.Client,
		Method: "OPTIONS",
		URL:    "https://registerdisney.go.com/jgc/v6/client/DCP-DISNEYSTORE.WEB-PROD/api-key?langPref=en-US",
		RawHeaders: http.RawHeader{
			{"accept", `*/*`},
			{"access-control-request-method", `POST`},
			{"access-control-request-headers", `cache-control,content-type,conversation-id,correlation-id,expires,pragma`},
			{"origin", `https://cdn.registerdisney.go.com`},
			{"user-agent", `Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/92.0.4515.107 Safari/537.36`},
			{"sec-fetch-mode", `cors`},
			{"sec-fetch-site", `same-site`},
			{"sec-fetch-dest", `empty`},
			{"referer", `https://cdn.registerdisney.go.com/`},
			{"accept-encoding", `gzip, deflate, br`},
			{"accept-language", `en-US,en;q=0.9`},
		},
	})
	if err != nil || resp.StatusCode != 200 {
		return false
	}

	correlationID := uuid.New().String()
	conversationId := uuid.New().String()
	currentTime := time.Now().UTC()

	resp, _, err = util.MakeRequest(&util.Request{
		Client: task.Task.Client,
		Method: "POST",
		URL:    "https://registerdisney.go.com/jgc/v6/client/DCP-DISNEYSTORE.WEB-PROD/api-key?langPref=en-US",
		RawHeaders: http.RawHeader{
			{"content-length", `4`},
			{"sec-ch-ua", `" Not;A Brand";v="99", "Google Chrome";v="91", "Chromium";v="91"`},
			{"pragma", `no-cache`},
			{"correlation-id", correlationID},
			{"sec-ch-ua-mobile", `?0`},
			{"user-agent", `Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/91.0.4472.164 Safari/537.36`},
			{"content-type", `application/json`},
			{"cache-control", `no-cache`},
			{"conversation-id", conversationId},
			{"expires", `-1`},
			{"accept", `*/*`},
			{"origin", `https://cdn.registerdisney.go.com`},
			{"sec-fetch-site", `same-site`},
			{"sec-fetch-mode", `cors`},
			{"sec-fetch-dest", `empty`},
			{"referer", `https://cdn.registerdisney.go.com/`},
			{"accept-encoding", `gzip, deflate, br`},
			{"accept-language", `en-US,en;q=0.9`},
		},
		Data: []byte("null"),
	})
	if err != nil || resp.StatusCode != 200 {
		return false
	}
	apiKey1 := resp.Header.Get("api-key")
	if apiKey1 == "" {
		return false
	}

	loginRequest := LoginRequest{
		Loginvalue: task.AccountInfo.Email,
		Password:   task.AccountInfo.Password,
	}

	proxy := entities.Proxy{}
	if task.Task.Proxy != nil {
		proxy = *task.Task.Proxy
	}
	token, err := captcha.RequestCaptchaToken(enums.ReCaptchaV3, enums.Disney, BaseEndpoint+"/", "login", 0.9, proxy)
	if err != nil {
		return false
	}
	for token == nil {
		needToStop := task.CheckForStop()
		if needToStop {
			return false
		}
		token = captcha.PollCaptchaTokens(enums.ReCaptchaV3, enums.Disney, BaseEndpoint+"/", proxy)
		time.Sleep(1 * time.Second / 10)
	}
	tokenInfo, ok := token.(entities.ReCaptchaToken)
	if !ok {
		return false
	}

	data, err := json.Marshal(loginRequest)
	if err != nil {
		return false
	}
	loginResponse := LoginResponse{}
	resp, _, err = util.MakeRequest(&util.Request{
		Client: task.Task.Client,
		Method: "POST",
		URL:    FirstLoginEndpoint,
		RawHeaders: http.RawHeader{
			{"sec-ch-ua", `"Chromium";v="92", " Not A;Brand";v="99", "Google Chrome";v="92"`},
			{"pragma", `no-cache`},
			{"correlation-id", correlationID},
			{"sec-ch-ua-mobile", `?0`},
			{"content-type", `application/json`},
			{"user-agent", `Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/92.0.4515.107 Safari/537.36`},
			{"cache-control", `no-cache`},
			{"conversation-id", conversationId},
			{"device-id", `null`},
			{"g-recaptcha-token", tokenInfo.Token},
			{"expires", `-1`},
			{"accept", `*/*`},
			{"origin", `https://cdn.registerdisney.go.com`},
			{"sec-fetch-site", `same-site`},
			{"sec-fetch-mode", `cors`},
			{"sec-fetch-dest", `empty`},
			{"referer", `https://cdn.registerdisney.go.com/`},
			{"accept-encoding", `gzip, deflate, br`},
			{"accept-language", `en-US,en;q=0.9`},
			{"authorization", `APIKEY ` + apiKey1},
			{"content-length", fmt.Sprint(len(data))},
		},
		Data:               data,
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
			task.Task.HasStockData = true
			return false
		}
		time.Sleep(common.MS_TO_WAIT)
	}
}

func (task *Task) AddToCart() bool {
	for _, cookie := range task.Task.Client.Jar.Cookies(ParsedBase) {
		if cookie.Name == "_abck" {
			validator, _ := util.FindInString(cookie.Value, "~", "~")
			if validator == "-1" {
				err := util.NewAbck(&task.Task.Client, task.StockData.ItemURL, BaseEndpoint, AkamaiEndpoint)
				if err != nil {
					return false
				}
			}
		}
	}

	data := []byte(common.CreateParams(map[string]string{
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

func (task *Task) GetCheckoutInfo() bool {
	getCheckoutInfoResponse := GetCheckoutInfoResponse{}
	resp, _, err := util.MakeRequest(&util.Request{
		Client: task.Task.Client,
		Method: "GET",
		URL:    GetCheckoutInfoEndpoint,
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
		ResponseBodyStruct: &getCheckoutInfoResponse,
	})
	if err != nil || resp.StatusCode != 200 {
		fmt.Println(err)
		return false
	}

	task.TaskInfo.ShipmentUUID = getCheckoutInfoResponse.Items[0].ShipmentUUID
	task.TaskInfo.ShippingMethod = getCheckoutInfoResponse.Shipments[0].SelectedShippingMethod
	task.TaskInfo.CsrfToken = getCheckoutInfoResponse.Csrf.Token
	return true
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
	data := []byte(common.CreateParams(map[string]string{
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
	default:
		status = enums.OrderStatusSuccess
		success = true
	}

	go util.ProcessCheckout(&util.ProcessCheckoutInfo{
		BaseTask:     task.Task,
		Success:      success,
		Status:       status,
		Content:      "",
		Embeds:       task.CreateDisneyEmbed(status, task.StockData.ImageURL),
		ItemName:     task.StockData.ProductName,
		ImageURL:     task.StockData.ImageURL,
		Sku:          task.StockData.PID,
		Retailer:     enums.Disney,
		Price:        task.TaskInfo.Total,
		Quantity:     task.Task.Task.TaskQty,
		MsToCheckout: time.Since(startTime).Milliseconds(),
	})

	return success, true, status
}
