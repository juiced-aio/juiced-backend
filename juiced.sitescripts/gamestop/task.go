package gamestop

import (
	"fmt"
	"log"
	"net/url"
	"strings"

	"backend.juicedbot.io/juiced.antibot/akamai"
	"backend.juicedbot.io/juiced.infrastructure/entities"
	"backend.juicedbot.io/juiced.infrastructure/enums"
	u "backend.juicedbot.io/juiced.infrastructure/util"
	"backend.juicedbot.io/juiced.sitescripts/util"

	"github.com/anaskhan96/soup"
	browser "github.com/eddycjy/fake-useragent"
)

const MAX_RETRIES = 5

func CreateTask(input entities.TaskInput, baseTask *entities.BaseTask) (entities.RetailerTask, error) {
	gameStopTaskInput, err := ValidateTaskInput(input.SiteSpecificInfo)
	if err != nil {
		return nil, err
	}
	return &Task{
		Input:         input,
		GameStopInput: gameStopTaskInput,
		BaseTask:      baseTask,
	}, nil
}

func (task *Task) GetSetupFunctions() []entities.TaskFunction {
	setupTaskFunctions := []entities.TaskFunction{}

	if task.GameStopInput.TaskType == enums.TaskTypeAccount {
		setupTaskFunctions = append(setupTaskFunctions, []entities.TaskFunction{
			// 1. Login
			{
				Function:         task.Login,
				StatusBegin:      enums.LoggingIn,
				StatusPercentage: 30,
				MaxRetries:       MAX_RETRIES,
			},
			// TODO: 2. Refresh Login
			// {
			// 	Function:        task.Login,
			// 	StatusBegin:     enums.LoggingIn,
			// 	MaxRetries:      MAX_RETRIES,
			// 	RefreshFunction: true,
			// 	RefreshEvery:    1800,
			// },
		}...)
	} else {
		setupTaskFunctions = append(setupTaskFunctions, []entities.TaskFunction{
			// 1. LoginGuest
			{
				Function:         task.LoginGuest,
				StatusBegin:      enums.SettingUp,
				StatusPercentage: 30,
				MaxRetries:       MAX_RETRIES,
			},
			// TODO: 2. Refresh LoginGuest
			// {
			// 	Function:        task.LoginGuest,
			// 	StatusBegin:     enums.SettingUp,
			// 	MaxRetries:      MAX_RETRIES,
			// 	RefreshFunction: true,
			// 	RefreshEvery:    1800,
			// },
		}...)
	}

	return setupTaskFunctions
}

func (task *Task) GetMainFunctions() []entities.TaskFunction {
	mainTaskFunctions := []entities.TaskFunction{
		// 3. AddToCart
		{
			Function:         task.AddToCart,
			StatusBegin:      enums.AddingToCart,
			StatusPercentage: 55,
			MsBetweenRetries: task.Input.DelayMS,
		},
		// 4. GetCartInfo
		{
			Function:         task.GetCartInfo,
			StatusBegin:      enums.GettingCartInfo,
			StatusPercentage: 65,
			MaxRetries:       MAX_RETRIES,
		},
		// 5. SubmitShippingDetails
		{
			Function:         task.SubmitShippingDetails,
			StatusBegin:      enums.SettingShippingInfo,
			StatusPercentage: 75,
			MaxRetries:       MAX_RETRIES,
		},
		// 6. SubmitPaymentDetails
		{
			Function:         task.SubmitPaymentDetails,
			StatusBegin:      enums.SettingBillingInfo,
			StatusPercentage: 85,
			MaxRetries:       MAX_RETRIES,
		},
		// 7. SubmitOrder
		{
			Function:         task.SubmitOrder,
			StatusBegin:      enums.CheckingOut,
			StatusPercentage: 95,
			MaxRetries:       MAX_RETRIES,
			CheckoutFunction: true,
		},
	}
	return mainTaskFunctions
}

func (task *Task) Login() (bool, string) {
	resp, body, err := util.MakeRequest(&util.Request{
		Client:     task.BaseTask.Client,
		Method:     "GET",
		URL:        BaseLoginEndpoint,
		RawHeaders: util.DefaultRawHeaders,
	})
	if err != nil {
		log.Println(err.Error())
		// return false, fmt.Sprintf(enums.LoginFailure, err.Error())
	}
	switch resp.StatusCode {
	case 200:
	default:
		return false, fmt.Sprintf(enums.LoginFailure, fmt.Sprintf(enums.UnknownError, resp.StatusCode))
	}

	csrf, err := u.FindInString(body, `name="csrf_token" value="`, `"`)
	if err != nil {
		return false, fmt.Sprintf(enums.LoginFailure, err.Error())
	}

	err = akamai.HandleAkamaiTask(task.BaseTask, BaseEndpoint+"/", BaseEndpoint, AkamaiEndpoint, BaseURL)
	if err != nil {
		return false, fmt.Sprintf(enums.LoginFailure, err.Error())
	}

	loginResponse := LoginResponse{}
	form := url.Values{
		"loginEmail":         {task.GameStopInput.Email},
		"loginPassword":      {task.GameStopInput.Password},
		"loginRememberMe":    {"true"},
		"userTimezoneOffset": {"420"},
		"csrf_token":         {csrf},
	}
	_, _, err = util.MakeRequest(&util.Request{
		Client: task.BaseTask.Client,
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
		log.Println(err.Error())
		// return false, fmt.Sprintf(enums.LoginFailure, err.Error())
	}

	if !loginResponse.LoginStatus.Success {
		return false, fmt.Sprintf(enums.LoginFailure, loginResponse.LoginStatus.ErrorMessage)
	}
	_, _, err = util.MakeRequest(&util.Request{
		Client: task.BaseTask.Client,
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
		log.Println(err.Error())
		// return false, fmt.Sprintf(enums.LoginFailure, err.Error())
	}

	switch resp.StatusCode {
	case 200:
		return true, enums.LoginSuccess

	case 403:
		err = akamai.HandleAkamaiTask(task.BaseTask, BaseEndpoint+"/", BaseEndpoint, AkamaiEndpoint, BaseURL)
		if err != nil {
			return false, fmt.Sprintf(enums.LoginFailure, err.Error())
		}
		return false, ""
	}

	return false, fmt.Sprintf(enums.LoginFailure, fmt.Sprintf(enums.UnknownError, resp.StatusCode))
}

func (task *Task) LoginGuest() (bool, string) {
	resp, _, err := util.MakeRequest(&util.Request{
		Client: task.BaseTask.Client,
		Method: "GET",
		URL:    BaseEndpoint,
		RawHeaders: [][2]string{
			{"upgrade-insecure-requests", "1"},
			{"user-agent", browser.Chrome()},
			{"accept", "text/html,application/xhtml+xml,application/xml;q=0.9,image/avif,image/webp,image/apng,*/*;q=0.8,application/signed-exchange;v=b3;q=0.9"},
			{"sec-fetch-site", "none"},
			{"sec-fetch-mode", "navigate"},
			{"sec-fetch-user", "?1"},
			{"sec-fetch-dest", "document"},
			{"sec-ch-ua", `" Not A;Brand";v="99", "Chromium";v="90", "Google Chrome";v="90"`},
			{"sec-ch-ua-mobile", "?0"},
			{"accept-encoding", "gzip, deflate, br"},
			{"accept-language", "en-US,en;q=0.9"},
		},
	})
	if err != nil {
		log.Println(err.Error())
		// return false, fmt.Sprintf(enums.SettingUpFailure, err.Error())
	}

	switch resp.StatusCode {
	case 200:
	default:
		return false, fmt.Sprintf(enums.SettingUpFailure, fmt.Sprintf(enums.UnknownError, resp.StatusCode))
	}

	err = akamai.HandleAkamaiTask(task.BaseTask, BaseEndpoint+"/", BaseEndpoint, AkamaiEndpoint, BaseURL)
	if err != nil {
		return false, fmt.Sprintf(enums.LoginFailure, err.Error())
	}

	return true, enums.SettingUpSuccess
}

func (task *Task) AddToCart() (bool, string) {
	quantity := task.Input.Quantity
	if maxQuantity, ok := task.BaseTask.ProductInfo.SiteSpecificInfo["MaxQuantity"].(int); ok && maxQuantity != 0 && quantity > maxQuantity {
		quantity = maxQuantity
	}

	pid := task.BaseTask.ProductInfo.SKU
	if vid, ok := task.BaseTask.ProductInfo.SiteSpecificInfo["VID"].(string); ok && vid != "" {
		pid = vid
	}

	addToCartResponse := AddToCartResponse{}
	form := url.Values{
		"pid":            {pid},
		"quantity":       {fmt.Sprint(quantity)},
		"upsellID":       {""},
		"purPROID":       {""},
		"options":        {"[]"},
		"isSDD":          {"false"},
		"deliveryOption": {"home"},
		"pageSpecified":  {"PDP"},
		"recommTitle":    {""},
	}
	if task.CaptchaToken != "" {
		form.Add("g-recaptcha-response", task.CaptchaToken)
	}
	resp, _, err := util.MakeRequest(&util.Request{
		Client: task.BaseTask.Client,
		Method: "POST",
		URL:    fmt.Sprintf(AddToCartEndpoint, pid),
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
			{"referer", task.BaseTask.ProductInfo.ItemURL},
			{"accept-encoding", "gzip, deflate, br"},
			{"accept-language", "en-US,en;q=0.9"},
		},
		Data:               []byte(form.Encode()),
		ResponseBodyStruct: &addToCartResponse,
	})
	if err != nil {
		log.Println(err.Error())
		// return false, fmt.Sprintf(enums.AddingToCartFailure, err.Error())
	}

	switch resp.StatusCode {
	case 200:
		if addToCartResponse.QuantityTotal == 0 {
			if addToCartResponse.CaptchaProtected {
				task.BaseTask.PublishEvent(enums.WaitingForCaptcha, 50, enums.TaskUpdate)
				task.CaptchaProtected = true

				token, err := util.RequestCaptchaToken(task.BaseTask, enums.ReCaptchaV2, enums.GameStop, task.BaseTask.ProductInfo.ItemURL, "atc", 0.8)
				if err != nil {
					return false, fmt.Sprintf(enums.AddingToCartFailure, err.Error())
				}

				task.CaptchaToken = token

				needToStop := task.BaseTask.CheckForStop()
				if needToStop {
					return false, ""
				}

				task.BaseTask.PublishEvent(enums.AddingToCart, 50, enums.TaskUpdate)
				return false, ""
			}
			return false, fmt.Sprintf(enums.AddingToCartFailure, addToCartResponse.Message)
		} else {
			return true, enums.AddingToCartSuccess
		}

	case 403:
		err = akamai.HandleAkamaiTask(task.BaseTask, BaseEndpoint+"/", BaseEndpoint, AkamaiEndpoint, BaseURL)
		if err != nil {
			return false, fmt.Sprintf(enums.AddingToCartFailure, err.Error())
		}
		return false, ""
	}

	return false, fmt.Sprintf(enums.AddingToCartFailure, fmt.Sprintf(enums.UnknownError, resp.StatusCode))
}

func (task *Task) GetCartInfo() (bool, string) {
	resp, body, err := util.MakeRequest(&util.Request{
		Client: task.BaseTask.Client,
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
		log.Println(err.Error())
		// return false, fmt.Sprintf(enums.AddingToCartFailure, err.Error())
	}

	switch resp.StatusCode {
	case 200:
		doc := soup.HTMLParse(string(body))
		shipmentUUIDInput := doc.Find("input", "name", "shipmentUUID")
		if shipmentUUIDInput.Error != nil {
			return false, fmt.Sprintf(enums.GettingCartInfoFailure, shipmentUUIDInput.Error.Error())
		}
		task.ShipmentUUID = shipmentUUIDInput.Attrs()["value"]
		originalShipmentUUIDInput := doc.Find("input", "name", "originalShipmentUUID")
		if originalShipmentUUIDInput.Error != nil {
			return false, fmt.Sprintf(enums.GettingCartInfoFailure, originalShipmentUUIDInput.Error.Error())
		}
		task.OriginalShipmentUUID = originalShipmentUUIDInput.Attrs()["value"]
		csrfInput := doc.Find("input", "name", "csrf_token")
		if csrfInput.Error != nil {
			return false, fmt.Sprintf(enums.GettingCartInfoFailure, csrfInput.Error.Error())
		}
		task.CSRF = csrfInput.Attrs()["value"]
		if task.ShipmentUUID != "" && task.OriginalShipmentUUID != "" && task.CSRF != "" {
			return true, enums.GettingCartInfoSuccess
		}
		return false, fmt.Sprintf(enums.GettingCartInfoFailure, EmptyCSRFValuesError)

	case 403:
		err = akamai.HandleAkamaiTask(task.BaseTask, BaseEndpoint+"/", BaseEndpoint, AkamaiEndpoint, BaseURL)
		if err != nil {
			return false, fmt.Sprintf(enums.GettingCartInfoFailure, err.Error())
		}
		return false, ""
	}

	return false, fmt.Sprintf(enums.GettingCartInfoFailure, fmt.Sprintf(enums.UnknownError, resp.StatusCode))
}

func (task *Task) SubmitShippingDetails() (bool, string) {
	form := url.Values{
		"originalShipmentUUID":             {task.OriginalShipmentUUID},
		"shipmentUUID":                     {task.ShipmentUUID},
		"shippingAddressUpdateLinkClicked": {"false"},
		"dwfrm_shipping_shippingAddress_addressFields_firstName":        {task.BaseTask.Profile.ShippingAddress.FirstName},
		"dwfrm_shipping_shippingAddress_addressFields_lastName":         {task.BaseTask.Profile.ShippingAddress.LastName},
		"dwfrm_shipping_shippingAddress_addressFields_address1":         {task.BaseTask.Profile.ShippingAddress.Address1},
		"dwfrm_shipping_shippingAddress_addressFields_address2":         {task.BaseTask.Profile.ShippingAddress.Address2},
		"dwfrm_shipping_shippingAddress_addressFields_country":          {task.BaseTask.Profile.ShippingAddress.CountryCode},
		"dwfrm_shipping_shippingAddress_addressFields_postalCode":       {task.BaseTask.Profile.ShippingAddress.ZipCode},
		"dwfrm_shipping_shippingAddress_addressFields_city":             {task.BaseTask.Profile.ShippingAddress.City},
		"dwfrm_shipping_shippingAddress_addressFields_states_stateCode": {task.BaseTask.Profile.ShippingAddress.StateCode},
		"dwfrm_billing_email_emailAddress":                              {task.BaseTask.Profile.Email},
		"dwfrm_shipping_shippingAddress_addressFields_phone":            {task.BaseTask.Profile.PhoneNumber},
		"dwfrm_shipping_shippingAddress_shippingMethodID":               {"16"},
		"csrf_token": {task.CSRF},
	}
	resp, _, err := util.MakeRequest(&util.Request{
		Client: task.BaseTask.Client,
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
		log.Println(err.Error())
		// return false, fmt.Sprintf(enums.AddingToCartFailure, err.Error())
	}

	switch resp.StatusCode {
	case 200:
		return true, enums.SettingShippingInfoSuccess

	case 403:
		err = akamai.HandleAkamaiTask(task.BaseTask, BaseEndpoint+"/", BaseEndpoint, AkamaiEndpoint, BaseURL)
		if err != nil {
			return false, fmt.Sprintf(enums.SettingShippingInfoFailure, err.Error())
		}
		return false, ""
	}

	return false, fmt.Sprintf(enums.SettingShippingInfoFailure, fmt.Sprintf(enums.UnknownError, resp.StatusCode))
}

func (task *Task) SubmitPaymentDetails() (bool, string) {
	err := akamai.HandleAkamaiTask(task.BaseTask, PaymentEndpoint+"/", BaseEndpoint, AkamaiEndpoint, BaseURL)
	if err != nil {
		return false, fmt.Sprintf(enums.SettingBillingInfoFailure, err.Error())
	}

	cardType := util.DetectRetailerCardType([]byte(task.BaseTask.Profile.CreditCard.CardNumber), enums.GameStop)

	expMonth := task.BaseTask.Profile.CreditCard.ExpMonth
	if task.BaseTask.Profile.CreditCard.ExpMonth[0:1] == "0" {
		expMonth = task.BaseTask.Profile.CreditCard.ExpMonth[1:]
	}
	form := url.Values{
		"dwfrm_giftCard_balance_accountNumber":               {""},
		"dwfrm_giftCard_balance_pinNumber":                   {""},
		"g-recaptcha-response":                               {""},
		"dwfrm_billing_paymentMethod":                        {"CREDIT_CARD"},
		"dwfrm_billing_creditCardFields_cardType":            {cardType},
		"dwfrm_billing_creditCardFields_cardNumber":          {task.BaseTask.Profile.CreditCard.CardNumber},
		"dwfrm_billing_creditCardFields_expirationMonth":     {expMonth},
		"dwfrm_billing_creditCardFields_expirationYear":      {task.BaseTask.Profile.CreditCard.ExpYear},
		"dwfrm_billing_creditCardFields_securityCode":        {task.BaseTask.Profile.CreditCard.CVV},
		"addressSelector":                                    {"new"},
		"dwfrm_billing_addressFields_firstName":              {task.BaseTask.Profile.BillingAddress.FirstName},
		"dwfrm_billing_addressFields_lastName":               {task.BaseTask.Profile.BillingAddress.LastName},
		"dwfrm_billing_addressFields_address1":               {strings.ToUpper(task.BaseTask.Profile.BillingAddress.Address1)},
		"dwfrm_billing_addressFields_address2":               {strings.ToUpper(task.BaseTask.Profile.BillingAddress.Address2)},
		"dwfrm_billing_addressFields_country":                {task.BaseTask.Profile.BillingAddress.CountryCode},
		"dwfrm_billing_addressFields_postalCode":             {task.BaseTask.Profile.BillingAddress.ZipCode},
		"dwfrm_billing_addressFields_city":                   {strings.ToUpper(task.BaseTask.Profile.BillingAddress.City)},
		"dwfrm_billing_addressFields_states_stateCode":       {task.BaseTask.Profile.BillingAddress.StateCode},
		"dwfrm_billing_email_emailAddress":                   {task.BaseTask.Profile.Email},
		"dwfrm_billing_addressFields_phone":                  {task.BaseTask.Profile.PhoneNumber},
		"dwfrm_billing_purCreditCardFields_purAccountNumber": {""},
		"flexPay":              {"nonFlexPayment"},
		"flexPayImgUrls":       {""},
		"flexPayEnable":        {"true"},
		"flexPayEnableQuadPay": {""},
		"flexPayEnableKlarna":  {""},
		"flexPayEnableSezzle":  {""},
		"csrf_token":           {task.CSRF},
		"flexpay":              {"nonFlexPayment"},
		"dwfrm_billing_accertify_accertifyDeviceFingerprint": {""},
	}

	resp, _, err := util.MakeRequest(&util.Request{
		Client: task.BaseTask.Client,
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
		log.Println(err.Error())
		// return false, fmt.Sprintf(enums.SettingBillingInfoFailure, err.Error())
	}

	switch resp.StatusCode {
	case 200:
		return true, enums.SettingBillingInfoSuccess

	case 403:
		err = akamai.HandleAkamaiTask(task.BaseTask, BaseEndpoint+"/", BaseEndpoint, AkamaiEndpoint, BaseURL)
		if err != nil {
			return false, fmt.Sprintf(enums.SettingBillingInfoFailure, err.Error())
		}
		return false, ""
	}

	return false, fmt.Sprintf(enums.SettingBillingInfoFailure, fmt.Sprintf(enums.UnknownError, resp.StatusCode))
}

func (task *Task) SubmitOrder() (bool, string) {
	placeOrderResponse := PlaceOrderResponse{}
	form := url.Values{
		"klarnaOrderId":              {""},
		"accertifyDeviceFingerprint": {""},
	}

	if task.CaptchaProtected {
		task.BaseTask.PublishEvent(enums.WaitingForCaptcha, 95, enums.TaskUpdate)
		task.CaptchaProtected = true

		token, err := util.RequestCaptchaToken(task.BaseTask, enums.ReCaptchaV2, enums.GameStop, CheckoutEndpoint+"/", "checkout", 0.8)
		if err != nil {
			return false, fmt.Sprintf(enums.CheckingOutFailure, err.Error())
		}
		form.Add("g-recaptcha-response", token)

		needToStop := task.BaseTask.CheckForStop()
		if needToStop {
			return false, ""
		}

		task.BaseTask.PublishEvent(enums.CheckingOut, 95, enums.TaskUpdate)
	}

	resp, _, err := util.MakeRequest(&util.Request{
		Client: task.BaseTask.Client,
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
	// ok := util.HandleErrors(err, util.RequestDoError)
	// if !ok {
	// 	return false, status
	// }
	if err != nil {
		log.Println(err.Error())
		// return false, fmt.Sprintf(enums.CheckingOutFailure, err.Error())
	}

	if placeOrderResponse.Error &&
		(strings.Contains(placeOrderResponse.ErrorMessage, "another form of payment") ||
			strings.Contains(placeOrderResponse.ErrorMessage, "The payment you submitted is not valid")) {
		return false, enums.CardDeclined
	}

	switch resp.StatusCode {
	case 200:
		if placeOrderResponse.Error {
			return false, fmt.Sprintf(enums.CheckingOutFailure, placeOrderResponse.ErrorMessage)
		} else {
			return true, enums.CheckingOutSuccess
		}

	case 403:
		err = akamai.HandleAkamaiTask(task.BaseTask, BaseEndpoint+"/", BaseEndpoint, AkamaiEndpoint, BaseURL)
		if err != nil {
			return false, fmt.Sprintf(enums.CheckingOutFailure, err.Error())
		}
		return false, ""
	}

	return false, fmt.Sprintf(enums.CheckingOutFailure, fmt.Sprintf(enums.UnknownError, resp.StatusCode))
}
