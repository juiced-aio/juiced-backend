package topps

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"html"
	"log"
	"math/rand"
	"net/url"
	"time"

	"backend.juicedbot.io/juiced.client/http"
	"backend.juicedbot.io/juiced.infrastructure/common"
	"backend.juicedbot.io/juiced.infrastructure/common/captcha"
	"backend.juicedbot.io/juiced.infrastructure/common/entities"
	"backend.juicedbot.io/juiced.infrastructure/common/enums"
	"backend.juicedbot.io/juiced.infrastructure/common/events"
	"backend.juicedbot.io/juiced.sitescripts/base"
	"backend.juicedbot.io/juiced.sitescripts/hawk-go"
	"backend.juicedbot.io/juiced.sitescripts/util"
	"github.com/anaskhan96/soup"
	"github.com/google/uuid"
	cmap "github.com/orcaman/concurrent-map"
)

var ToppsAccountStore = cmap.New()

// CreateToppsTask takes a Task entity and turns it into a Topps Task
func CreateToppsTask(task *entities.Task, profile entities.Profile, proxyGroup *entities.ProxyGroup, eventBus *events.EventBus, taskType enums.TaskType, email, password string) (Task, error) {
	toppsTask := Task{}

	toppsTask = Task{
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
		toppsTask.Task.Proxy = util.RandomLeastUsedProxy(proxyGroup.Proxies)
	}
	return toppsTask, nil
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
// 		1. Setup task
// 		2. WaitForMonitor
// 		3. AddToCart
// 		4. GetCartInfo
//		5. SubmitShippingInfo
// 		6. GetCardToken
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
	if task.Task.Task.TaskQty <= 0 {
		task.Task.Task.TaskQty = 1
	}

	err := task.Task.CreateClient(task.Task.Proxy)
	if err != nil {
		return
	}
	task.Task.Scraper = hawk.Init(task.Task.Client, common.HAWK_KEY, false)

	// 1. Setup task
	task.PublishEvent(enums.SettingUp, enums.TaskStart)
	setup := task.Setup()
	if setup {
		return
	}

	// Adding the account to the pool
	var accounts = []Acc{{task.Task.Task.TaskGroupID, task.Task.Scraper, task.AccountInfo}}
	oldAccounts, _ := AccountPool.Get(task.Task.Task.TaskGroupID)
	if oldAccounts != nil {
		accounts = append(accounts, oldAccounts.([]Acc)...)
	}
	AccountPool.Set(task.Task.Task.TaskGroupID, accounts)

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
	// 4. GetCartInfo
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

	task.PublishEvent(enums.SettingShippingInfo, enums.TaskUpdate)
	// 5. SubmitShippingInfo
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

	task.PublishEvent(enums.SettingBillingInfo, enums.TaskUpdate)
	// 6. GetCardToken
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

// Sets the client up by either logging in or waiting for another task to login that is using the same account
func (task *Task) Setup() bool {
	if task.TaskType == enums.TaskTypeGuest {
		return BecomeGuest(task.Task.Scraper)
	}
	// Bad but quick solution to the multiple logins
	time.Sleep(time.Duration(rand.Intn(1000)) * time.Millisecond)
	if ToppsAccountStore.Has(task.AccountInfo.Email) {
		inMap := true
		for inMap {
			needToStop := task.CheckForStop()
			if needToStop {
				return true
			}
			// Error will be nil unless the item isn't in the map which we are already checking above
			value, ok := ToppsAccountStore.Get(task.AccountInfo.Email)
			if ok {
				acc, ok := value.(Acc)
				if ok {
					if len(task.Task.Scraper.Client.Jar.Cookies(ParsedBase)) == 0 {
						task.Task.Scraper.Client.Jar = acc.Scraper.Client.Jar
					}
					task.AccountInfo = acc.AccountInfo
					break
				} else {
					if task.Task.Task.TaskStatus != enums.WaitingForLogin {
						task.PublishEvent(enums.WaitingForLogin, enums.TaskUpdate)
					}
					time.Sleep(common.MS_TO_WAIT)
				}
			} else {
				inMap = false
				return task.Setup()
			}
		}
	} else {
		// Login
		task.PublishEvent(enums.LoggingIn, enums.TaskUpdate)
		loggedIn := false
		for !loggedIn {
			needToStop := task.CheckForStop()
			if needToStop {
				return true
			}
			loggedIn = task.Login()
			if !loggedIn {
				time.Sleep(time.Duration(task.Task.Task.TaskDelay) * time.Millisecond)
			}
		}

	}

	return false
}

// Haven't tested the login yet
func (task *Task) Login() bool {
	defer func() {
		if recover() != nil {
			ToppsAccountStore.Remove(task.AccountInfo.Email)
		}
	}()

	ToppsAccountStore.Set(task.AccountInfo.Email, false)

	resp, body, err := util.MakeRequest(&util.Request{
		Scraper: task.Task.Scraper,
		Method:  "GET",
		URL:     BaseLoginEndpoint,
		RawHeaders: http.RawHeader{
			{"upgrade-insecure-requests", `1`},
			{"user-agent", `Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/92.0.4515.131 Safari/537.36`},
			{"accept", `text/html,application/xhtml+xml,application/xml;q=0.9,image/avif,image/webp,image/apng,*/*;q=0.8,application/signed-exchange;v=b3;q=0.9`},
			{"sec-fetch-site", `same-origin`},
			{"sec-fetch-mode", `navigate`},
			{"sec-fetch-user", `?1`},
			{"sec-fetch-dest", `document`},
			{"sec-ch-ua", `"Chromium";v="92", " Not A;Brand";v="99", "Google Chrome";v="92"`},
			{"sec-ch-ua-mobile", `?0`},
			{"accept-encoding", `gzip, deflate, br`},
			{"accept-language", `en-US,en;q=0.9`},
		},
	})
	if resp.StatusCode != 200 || err != nil {
		ToppsAccountStore.Remove(task.AccountInfo.Email)
		return false
	}

	doc := soup.HTMLParse(body)
	elem := doc.Find("input", "name", "form_key")
	if elem.Error != nil {
		ToppsAccountStore.Remove(task.AccountInfo.Email)
		return false
	}
	formKey := elem.Attrs()["value"]

	proxy := entities.Proxy{}
	if task.Task.Proxy != nil {
		proxy = *task.Task.Proxy
	}
	token, err := captcha.RequestCaptchaToken(enums.ReCaptchaV2, enums.Topps, BaseLoginEndpoint+"/", "login", 0.7, proxy)
	if err != nil {
		ToppsAccountStore.Remove(task.AccountInfo.Email)
		return false
	}
	for token == nil {
		token = captcha.PollCaptchaTokens(enums.ReCaptchaV2, enums.Topps, BaseEndpoint+"/", proxy)
		time.Sleep(common.MS_TO_WAIT)
	}
	tokenInfo, ok := token.(entities.ReCaptchaToken)
	if !ok {
		ToppsAccountStore.Remove(task.AccountInfo.Email)
		return false
	}

	payload := common.CreateParams(map[string]string{
		"form_key":                           formKey,
		"login[username]":                    task.AccountInfo.Email,
		"login[password]":                    task.AccountInfo.Password,
		"amgdpr_agreement[privacy_checkbox]": `1`,
		"am-gdpr-checkboxes-from":            `registration`,
		"g-recaptcha-response":               tokenInfo.Token,
		"recaptcha-validate-":                ``,
		"send":                               ``,
	})

	resp, _, err = util.MakeRequest(&util.Request{
		Scraper: task.Task.Scraper,
		Method:  "POST",
		URL:     LoginEndpoint,
		RawHeaders: http.RawHeader{
			{"cache-control", `max-age=0`},
			{"sec-ch-ua", `"Chromium";v="92", " Not A;Brand";v="99", "Google Chrome";v="92"`},
			{"sec-ch-ua-mobile", `?0`},
			{"upgrade-insecure-requests", `1`},
			{"origin", BaseEndpoint},
			{"content-type", `application/x-www-form-urlencoded`},
			{"user-agent", `Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/92.0.4515.131 Safari/537.36`},
			{"accept", `text/html,application/xhtml+xml,application/xml;q=0.9,image/avif,image/webp,image/apng,*/*;q=0.8,application/signed-exchange;v=b3;q=0.9`},
			{"sec-fetch-site", `same-origin`},
			{"sec-fetch-mode", `navigate`},
			{"sec-fetch-user", `?1`},
			{"sec-fetch-dest", `document`},
			{"referer", BaseLoginEndpoint + "/"},
			{"accept-encoding", `gzip, deflate, br`},
			{"accept-language", `en-US,en;q=0.9`},
		},

		Data: []byte(payload),
	})
	if resp.StatusCode != 302 || err != nil {
		ToppsAccountStore.Remove(task.AccountInfo.Email)
		return false
	}

	acc := Acc{
		GroupID: task.Task.Task.TaskGroupID,
		Scraper: task.Task.Scraper,
		AccountInfo: AccountInfo{
			Email:    task.AccountInfo.Email,
			Password: task.AccountInfo.Password,
		},
	}

	ToppsAccountStore.Set(task.AccountInfo.Email, acc)

	return true
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

// Adds the item to the cart
func (task *Task) AddToCart() bool {
	data := task.CreateMultipartForm()

	urlParsed, _ := url.Parse(task.StockData.AddURL)
	task.Task.Scraper.Client.Jar.SetCookies(urlParsed, []*http.Cookie{
		{
			Name:  "form_key",
			Value: task.StockData.FormKey,
		},
	})

	resp, _, err := util.MakeRequest(&util.Request{
		Scraper: task.Task.Scraper,
		Method:  "POST",
		URL:     task.StockData.AddURL,
		RawHeaders: http.RawHeader{
			{"pragma", `no-cache`},
			{"cache-control", `no-cache`},
			{"sec-ch-ua", `"Chromium";v="92", " Not A;Brand";v="99", "Google Chrome";v="92"`},
			{"accept", `application/json, text/javascript, */*; q=0.01`},
			{"x-requested-with", `XMLHttpRequest`},
			{"sec-ch-ua-mobile", `?0`},
			{"user-agent", `Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/92.0.4515.131 Safari/537.36`},
			{"content-type", `multipart/form-data; boundary=----WebKitFormBoundaryIL7E2YKB4G0gjFhb`},
			{"origin", BaseEndpoint},
			{"sec-fetch-site", `same-origin`},
			{"sec-fetch-mode", `cors`},
			{"sec-fetch-dest", `empty`},
			{"referer", task.StockData.ItemURL + ".html"},
			{"accept-encoding", `gzip, deflate`},
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

// Gets the cart info which includes the checkoutID and auth token for the card requests
func (task *Task) GetCartInfo() bool {
	var getCartInfoResponse GetCartInfoResponse
	resp, _, err := util.MakeRequest(&util.Request{
		Scraper: task.Task.Scraper,
		Method:  "GET",
		URL:     GetCartInfoEndpoint + fmt.Sprint(time.Now().UnixNano()),
		RawHeaders: http.RawHeader{
			{"sec-ch-ua", `"Chromium";v="92", " Not A;Brand";v="99", "Google Chrome";v="92"`},
			{"accept", `application/json, text/javascript, */*; q=0.01`},
			{"x-newrelic-id", `Vg8HUFNVDhAHVVdXBwYAVFE=`},
			{"x-requested-with", `XMLHttpRequest`},
			{"sec-ch-ua-mobile", `?0`},
			{"upgrade-insecure-requests", `1`},
			{"user-agent", `Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/92.0.4515.131 Safari/537.36`},
			{"sec-fetch-site", `none`},
			{"sec-fetch-mode", `navigate`},
			{"sec-fetch-user", `?1`},
			{"sec-fetch-dest", `document`},
			{"referer", task.StockData.ItemURL + ".html"},
			{"accept-encoding", `gzip, deflate`},
			{"accept-language", `en-US,en;q=0.9`},
		},
		ResponseBodyStruct: &getCartInfoResponse,
	})
	if err != nil || resp.StatusCode != 200 {
		fmt.Println(err)

		return false
	}
	data, err := common.FindInString(getCartInfoResponse.Cart.ExtraActions, `data-mage-init=\"`, `\"`)
	if err != nil {
		return false
	}

	var dataMage DataMage
	dataDecoded := html.UnescapeString(data)

	err = json.Unmarshal([]byte(dataDecoded), &dataMage)
	if err != nil {
		return false
	}

	task.TaskInfo.CheckoutID = dataMage.PaypalBraintreeJsApplepayImplementationsShortcut.Quoteid

	decodedClientToken, err := base64.StdEncoding.DecodeString(dataMage.PaypalBraintreeJsApplepayImplementationsShortcut.Clienttoken)
	if err != nil {
		fmt.Println(err)
		return false
	}

	jsonMap := make(map[string]interface{})
	err = json.Unmarshal(decodedClientToken, &jsonMap)
	if err != nil {
		fmt.Println(err)
		return false
	}

	task.TaskInfo.AuthToken = jsonMap["authorizationFingerprint"].(string)

	return err == nil
}

// Submitting shipping info to the guest endpoint or the account endpoint depending on the TaskType
func (task *Task) SubmitShippingInfo() bool {
	currentEndpoint := fmt.Sprintf(SubmitShippingInfoEndpoint, task.TaskInfo.CheckoutID)
	if task.TaskType == enums.TaskTypeAccount {
		currentEndpoint = AccountSubmitShippingInfoEndpoint
	}
	submitShippingInfoRequest := SubmitShippingInfoRequest{
		Addressinformation: Addressinformation{
			ShippingAddress: ShippingAddress{
				Countryid:         task.Task.Profile.ShippingAddress.CountryCode,
				Regionid:          regionMap[stateMap[task.Task.Profile.ShippingAddress.StateCode]].Value,
				Regioncode:        task.Task.Profile.ShippingAddress.StateCode,
				Region:            stateMap[task.Task.Profile.ShippingAddress.StateCode],
				Street:            []string{task.Task.Profile.ShippingAddress.Address1, task.Task.Profile.ShippingAddress.Address2},
				Company:           "",
				Telephone:         task.Task.Profile.PhoneNumber,
				Fax:               "",
				Postcode:          task.Task.Profile.ShippingAddress.ZipCode,
				City:              task.Task.Profile.ShippingAddress.City,
				Firstname:         task.Task.Profile.ShippingAddress.FirstName,
				Lastname:          task.Task.Profile.ShippingAddress.LastName,
				SaveInAddressBook: 0,
			},
			BillingAddress: BillingAddress{
				Countryid:         task.Task.Profile.BillingAddress.CountryCode,
				Regionid:          regionMap[stateMap[task.Task.Profile.BillingAddress.StateCode]].Value,
				Regioncode:        task.Task.Profile.BillingAddress.StateCode,
				Region:            stateMap[task.Task.Profile.BillingAddress.StateCode],
				Street:            []string{task.Task.Profile.BillingAddress.Address1, task.Task.Profile.BillingAddress.Address2},
				Company:           "",
				Telephone:         task.Task.Profile.PhoneNumber,
				Fax:               "",
				Postcode:          task.Task.Profile.BillingAddress.ZipCode,
				City:              task.Task.Profile.BillingAddress.City,
				Firstname:         task.Task.Profile.BillingAddress.FirstName,
				Lastname:          task.Task.Profile.BillingAddress.LastName,
				SaveInAddressBook: 0,
			},
			ShippingMethodCode:  "amstrates1",
			ShippingCarrierCode: "amstrates",
			ExtensionAttributes: ExtensionAttributes{},
		},
	}
	resp, _, err := util.MakeRequest(&util.Request{
		Scraper: task.Task.Scraper,
		Method:  "POST",
		URL:     currentEndpoint,
		RawHeaders: http.RawHeader{
			{"sec-ch-ua", `"Chromium";v="92", " Not A;Brand";v="99", "Google Chrome";v="92"`},
			{"accept", `*/*`},
			{"x-requested-with", `XMLHttpRequest`},
			{"sec-ch-ua-mobile", `?0`},
			{"user-agent", `Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/92.0.4515.131 Safari/537.36`},
			{"content-type", `application/json`},
			{"origin", BaseEndpoint},
			{"sec-fetch-site", `same-origin`},
			{"sec-fetch-mode", `cors`},
			{"sec-fetch-dest", `empty`},
			{"referer", CheckoutEndpoint + "/"},
			{"accept-encoding", `gzip, deflate, br`},
			{"accept-language", `en-US,en;q=0.9`},
		},
		RequestBodyStruct: submitShippingInfoRequest,
	})
	if err != nil || resp.StatusCode != 200 {
		fmt.Println(err)
		return false
	}

	return true
}

// Getting the card token required in the PlaceOrder function
func (task *Task) GetCardToken() bool {
	sessionID := uuid.New().String()

	resp, _, err := util.MakeRequest(&util.Request{
		Scraper: task.Task.Scraper,
		Method:  "POST",
		URL:     GetCardTokenEndpoint,
		Headers: http.Header{
			"Connection":        {`keep-alive`},
			"sec-ch-ua":         {`"Chromium";v="92", " Not A;Brand";v="99", "Google Chrome";v="92"`},
			"Authorization":     {`Bearer ` + task.TaskInfo.AuthToken},
			"sec-ch-ua-mobile":  {`?0`},
			"User-Agent":        {`Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/92.0.4515.131 Safari/537.36`},
			"Braintree-Version": {`2018-05-10`},
			"Content-Type":      {`application/json`},
			"Accept":            {`*/*`},
			"Origin":            {`https://assets.braintreegateway.com`},
			"Sec-Fetch-Site":    {`cross-site`},
			"Sec-Fetch-Mode":    {`cors`},
			"Sec-Fetch-Dest":    {`empty`},
			"Referer":           {`https://assets.braintreegateway.com/`},
			"Accept-Encoding":   {`deflate, br`},
			"Accept-Language":   {`en-US,en;q=0.9`},
			http.HeaderOrderKey: {"Connection", "Content-Length", "sec-ch-ua", "Authorization", "sec-ch-ua-mobile", "User-Agent", "Braintree-Version", "Content-Type", "Accept", "Origin", "Sec-Fetch-Site", "Sec-Fetch-Mode", "Sec-Fetch-Dest", "Referer", "Accept-Encoding", "Accept-Language"},
		},
		Data: []byte(fmt.Sprintf(`{"clientSdkMetadata": {"source": "client","integration": "custom","sessionId": "%v"},"query": "query ClientConfiguration {   clientConfiguration {     analyticsUrl     environment     merchantId     assetsUrl     clientApiUrl     creditCard {       supportedCardBrands       challenges       threeDSecureEnabled       threeDSecure {         cardinalAuthenticationJWT       }     }     applePayWeb {       countryCode       currencyCode       merchantIdentifier       supportedCardBrands     }     googlePay {       displayName       supportedCardBrands       environment       googleAuthorization       paypalClientId     }     ideal {       routeId       assetsUrl     }     kount {       merchantId     }     masterpass {       merchantCheckoutId       supportedCardBrands     }     paypal {       displayName       clientId       privacyUrl       userAgreementUrl       assetsUrl       environment       environmentNoNetwork       unvettedMerchant       braintreeClientId       billingAgreementsEnabled       merchantAccountId       currencyCode       payeeEmail     }     unionPay {       merchantAccountId     }     usBankAccount {       routeId       plaidPublicKey     }     venmo {       merchantId       accessToken       environment     }     visaCheckout {       apiKey       externalClientId       supportedCardBrands     }     braintreeApi {       accessToken       url     }     supportedFeatures   } }","operationName": "ClientConfiguration"}`, sessionID)),
	})
	if err != nil || resp.StatusCode != 200 {
		fmt.Println(err)
		return false
	}

	getCardTokenRequest := GetCardTokenRequest{
		Clientsdkmetadata: Clientsdkmetadata{
			Source:      "client",
			Integration: "custom",
			Sessionid:   sessionID,
		},
		Query: "mutation TokenizeCreditCard($input: TokenizeCreditCardInput!) {   tokenizeCreditCard(input: $input) {     token     creditCard {       bin       brandCode       last4       cardholderName       expirationMonth      expirationYear      binData {         prepaid         healthcare         debit         durbinRegulated         commercial         payroll         issuingBank         countryOfIssuance         productId       }     }   } }",
		Variables: Variables{
			Input: Input{
				Creditcard: Creditcard{
					Number:          task.Task.Profile.CreditCard.CardNumber,
					Expirationmonth: task.Task.Profile.CreditCard.ExpMonth,
					Expirationyear:  "20" + task.Task.Profile.CreditCard.ExpYear,
					Cvv:             task.Task.Profile.CreditCard.CVV,
				},
				Options: Options{},
			},
		},
		Operationname: "TokenizeCreditCard",
	}
	var getCardTokenResponse GetCardTokenResponse
	resp, _, err = util.MakeRequest(&util.Request{
		Scraper: task.Task.Scraper,
		Method:  "POST",
		URL:     GetCardTokenEndpoint,
		Headers: http.Header{
			"Connection":        {`keep-alive`},
			"sec-ch-ua":         {`"Chromium";v="92", " Not A;Brand";v="99", "Google Chrome";v="92"`},
			"Authorization":     {`Bearer ` + task.TaskInfo.AuthToken},
			"sec-ch-ua-mobile":  {`?0`},
			"User-Agent":        {`Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/92.0.4515.131 Safari/537.36`},
			"Braintree-Version": {`2018-05-10`},
			"Content-Type":      {`application/json`},
			"Accept":            {`*/*`},
			"Origin":            {`https://assets.braintreegateway.com`},
			"Sec-Fetch-Site":    {`cross-site`},
			"Sec-Fetch-Mode":    {`cors`},
			"Sec-Fetch-Dest":    {`empty`},
			"Referer":           {`https://assets.braintreegateway.com/`},
			"Accept-Encoding":   {`deflate, br`},
			"Accept-Language":   {`en-US,en;q=0.9`},
			http.HeaderOrderKey: {"Connection", "Content-Length", "sec-ch-ua", "Authorization", "sec-ch-ua-mobile", "User-Agent", "Braintree-Version", "Content-Type", "Accept", "Origin", "Sec-Fetch-Site", "Sec-Fetch-Mode", "Sec-Fetch-Dest", "Referer", "Accept-Encoding", "Accept-Language"},
		},
		RequestBodyStruct:  getCardTokenRequest,
		ResponseBodyStruct: &getCardTokenResponse,
	})
	if err != nil || resp.StatusCode != 200 {
		fmt.Println(err)
		return false
	}

	task.TaskInfo.CardToken = getCardTokenResponse.Data.Tokenizecreditcard.Token

	return true
}

// Placing the order using the CardToken from the GetCardToken function
func (task *Task) PlaceOrder(startTime time.Time) (bool, enums.OrderStatus) {
	status := enums.OrderStatusFailed

	currentEndpoint := fmt.Sprintf(PlaceOrderEndpoint, task.TaskInfo.CheckoutID)
	if task.TaskType == enums.TaskTypeAccount {
		currentEndpoint = AccountPlaceOrderEndpoint
	}

	placeOrderRequest := PlaceOrderRequest{
		Cartid: task.TaskInfo.CheckoutID,
		Billingaddress: Billingaddress{
			Countryid:  task.Task.Profile.BillingAddress.CountryCode,
			Regionid:   regionMap[stateMap[task.Task.Profile.BillingAddress.StateCode]].Value,
			Regioncode: task.Task.Profile.BillingAddress.StateCode,
			Region:     stateMap[task.Task.Profile.BillingAddress.StateCode],
			Street:     []string{task.Task.Profile.BillingAddress.Address1, task.Task.Profile.BillingAddress.Address2},
			Company:    "",
			Telephone:  task.Task.Profile.PhoneNumber,
			Fax:        "",
			Postcode:   task.Task.Profile.BillingAddress.ZipCode,
			City:       task.Task.Profile.BillingAddress.City,
			Firstname:  task.Task.Profile.BillingAddress.FirstName,
			Lastname:   task.Task.Profile.BillingAddress.LastName,
		},
		Paymentmethod: Paymentmethod{
			Method: "braintree",
			AdditionalData: AdditionalData{
				PaymentMethodNonce: task.TaskInfo.CardToken,
				AmgdprAgreement:    "{}",
			},
		},
		Email: task.Task.Profile.Email,
	}

	var placeOrderResponse PlaceOrderResponse
	data, _ := json.Marshal(placeOrderRequest)
	resp, _, err := util.MakeRequest(&util.Request{
		Scraper: task.Task.Scraper,
		Method:  "POST",
		URL:     currentEndpoint,
		RawHeaders: http.RawHeader{
			{"content-length", fmt.Sprint(len(data))},
			{"sec-ch-ua", `"Chromium";v="92", " Not A;Brand";v="99", "Google Chrome";v="92"`},
			{"accept", `*/*`},
			{"x-newrelic-id", `Vg8HUFNVDhAHVVdXBwYAVFE=`},
			{"x-requested-with", `XMLHttpRequest`},
			{"sec-ch-ua-mobile", `?0`},
			{"user-agent", `Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/92.0.4515.131 Safari/537.36`},
			{"content-type", `application/json`},
			{"origin", BaseEndpoint},
			{"sec-fetch-site", `same-origin`},
			{"sec-fetch-mode", `cors`},
			{"sec-fetch-dest", `empty`},
			{"referer", CheckoutEndpoint + "/"},
			{"accept-encoding", `gzip, deflate`},
			{"accept-language", `en-US,en;q=0.9`},
		},
		Data:               data,
		ResponseBodyStruct: &placeOrderResponse,
	})
	if err != nil || resp.StatusCode != 200 {
		fmt.Println(err)
		return false, status
	}

	var success bool
	// I do not know what successfully placing an order returns
	switch placeOrderResponse.Message {
	case "Your payment could not be taken. Please try again or use a different payment method. Do Not Honor":
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
		Embeds:       task.CreateToppsEmbed(status, task.StockData.ImageURL),
		ItemName:     task.StockData.ProductName,
		ImageURL:     task.StockData.ImageURL,
		Sku:          task.StockData.Item,
		Retailer:     enums.Topps,
		Price:        task.StockData.Price,
		Quantity:     task.Task.Task.TaskQty,
		MsToCheckout: time.Since(startTime).Milliseconds(),
	})

	return true, status
}
