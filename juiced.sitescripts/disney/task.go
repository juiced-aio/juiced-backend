package disney

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"log"
	"net/url"
	"time"

	"backend.juicedbot.io/juiced.antibot/akamai"
	"backend.juicedbot.io/juiced.client/http"
	"backend.juicedbot.io/juiced.infrastructure/entities"
	"backend.juicedbot.io/juiced.infrastructure/enums"
	u "backend.juicedbot.io/juiced.infrastructure/util"
	"backend.juicedbot.io/juiced.sitescripts/util"

	browser "github.com/eddycjy/fake-useragent"
	"github.com/google/uuid"
)

const MAX_RETRIES = 5

func CreateTask(input entities.TaskInput, baseTask *entities.BaseTask) (entities.RetailerTask, error) {
	disneyTaskInput, err := ValidateTaskInput(input.SiteSpecificInfo)
	if err != nil {
		return nil, err
	}
	return &Task{
		Input:       input,
		DisneyInput: disneyTaskInput,
		BaseTask:    baseTask,
	}, nil
}

func (task *Task) GetSetupFunctions() []entities.TaskFunction {
	setupTaskFunctions := []entities.TaskFunction{}

	if task.DisneyInput.TaskType == enums.TaskTypeAccount {
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
			StatusPercentage: 60,
			MaxRetries:       MAX_RETRIES,
		},
		// 5. ValidateCheckout
		{
			Function:         task.ValidateCheckout,
			StatusBegin:      enums.SettingCartInfo,
			StatusPercentage: 65,
			MaxRetries:       MAX_RETRIES,
		},
		// 6. SubmitShippingDetails
		{
			Function:         task.SubmitShippingDetails,
			StatusBegin:      enums.SettingShippingInfo,
			StatusPercentage: 70,
			MaxRetries:       MAX_RETRIES,
		},
		// 7. EstablishAppSession
		{
			Function:         task.EstablishAppSession,
			StatusBegin:      enums.GettingBillingInfo,
			StatusPercentage: 75,
			MaxRetries:       MAX_RETRIES,
		},
		// 8. GetPaysheetAE
		{
			Function:         task.GetPaysheetAE,
			StatusBegin:      enums.SettingBillingInfo,
			StatusPercentage: 80,
			MaxRetries:       MAX_RETRIES,
		},
		// 9. GetCardToken
		{
			Function:         task.GetCardToken,
			StatusBegin:      enums.GettingOrderInfo,
			StatusPercentage: 85,
			MaxRetries:       MAX_RETRIES,
		},
		// 10. SubmitOrder
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
		URL:        BaseEndpoint,
		RawHeaders: util.DefaultRawHeaders,
	})
	if err != nil {
		return false, fmt.Sprintf(enums.LoginFailure, err.Error())
	}
	switch resp.StatusCode {
	case 200:
	default:
		return false, fmt.Sprintf(enums.LoginFailure, fmt.Sprintf(enums.UnknownError, resp.StatusCode))
	}

	clientID, err := u.FindInString(body, `"clientID": "`, `"`)
	if err != nil {
		return false, fmt.Sprintf(enums.LoginFailure, err.Error())
	}
	oidClientID, err := u.FindInString(body, `"oidClientID": "`, `"`)
	if err != nil {
		return false, fmt.Sprintf(enums.LoginFailure, err.Error())
	}

	if !IsValidAkamaiCookie(task.BaseTask.Client.Jar) {
		err = akamai.HandleAkamaiTask(task.BaseTask, BaseEndpoint+"/", BaseEndpoint, AkamaiEndpoint, BaseURL)
		if err != nil {
			return false, fmt.Sprintf(enums.LoginFailure, err.Error())
		}
	}

	resp, _, err = util.MakeRequest(&util.Request{
		Client: task.BaseTask.Client,
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
	if err != nil {
		log.Println(err.Error())
		if resp == nil {
			return false, fmt.Sprintf(enums.LoginFailure, err.Error())
		}
		// return false, fmt.Sprintf(enums.LoginFailure, err.Error())
	}

	correlationID := uuid.New().String()
	conversationId := uuid.New().String()
	currentTime := time.Now().UTC()

	resp, _, err = util.MakeRequest(&util.Request{
		Client: task.BaseTask.Client,
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
	if err != nil {
		log.Println(err.Error())
		if resp == nil {
			return false, fmt.Sprintf(enums.LoginFailure, err.Error())
		}
		// return false, fmt.Sprintf(enums.LoginFailure, err.Error())
	}
	apiKey1 := resp.Header.Get("api-key")
	if apiKey1 == "" {
		return false, fmt.Sprintf(enums.LoginFailure, "apiKey1")
	}

	loginRequest := LoginRequest{
		Loginvalue: task.DisneyInput.Email,
		Password:   task.DisneyInput.Password,
	}

	task.BaseTask.PublishEvent(enums.WaitingForCaptcha, 30, enums.TaskUpdate)
	token, err := util.RequestCaptchaToken(task.BaseTask, enums.Disney, enums.ReCaptchaV3, BaseEndpoint+"/", "login", 0.9)
	if err != nil {
		return false, fmt.Sprintf(enums.CheckingOutFailure, err.Error())
	}

	needToStop := task.BaseTask.CheckForStop()
	if needToStop {
		return false, ""
	}
	task.BaseTask.PublishEvent(enums.LoggingIn, 30, enums.TaskUpdate)

	data, err := json.Marshal(loginRequest)
	if err != nil {
		return false, fmt.Sprintf(enums.LoginFailure, "marshal")
	}
	loginResponse := LoginResponse{}
	resp, _, err = util.MakeRequest(&util.Request{
		Client: task.BaseTask.Client,
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
			{"g-recaptcha-token", token},
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
		log.Println(err.Error())
		if resp == nil {
			return false, fmt.Sprintf(enums.LoginFailure, err.Error())
		}
		// return false, fmt.Sprintf(enums.LoginFailure, err.Error())
	}
	apiKey2 := resp.Header.Get("api-key")
	if apiKey2 == "" {
		return false, fmt.Sprintf(enums.LoginFailure, "apiKey2")
	}
	deviceID := resp.Header.Get("device-id")
	if deviceID == "" {
		return false, fmt.Sprintf(enums.LoginFailure, "device-id")
	}

	switch resp.StatusCode {
	case 200:
	case 403:
		err = akamai.HandleAkamaiTask(task.BaseTask, BaseEndpoint+"/", BaseEndpoint, AkamaiEndpoint, BaseURL)
		if err != nil {
			return false, fmt.Sprintf(enums.LoginFailure, err.Error())
		}
		return false, ""
	default:
		return false, fmt.Sprintf(enums.LoginFailure, fmt.Sprintf(enums.UnknownError, resp.StatusCode))
	}

	loginResponse.Data.Token.Created = currentTime.Format("2006-01-02T15:04:05.000Z")
	loginResponse.Data.Token.LastChecked = currentTime.Format("2006-01-02T15:04:05.000Z")
	loginResponse.Data.Token.Expires = currentTime.Add(time.Hour * 24).Format("2006-01-02T15:04:05.000Z")
	loginResponse.Data.Token.RefreshExpires = currentTime.Add((time.Hour * 24) * 180).Format("2006-01-02T15:04:05.000Z")
	loginResponse.Data.Token.BlueCookie = nil
	jsonBytes, _ := json.Marshal(loginResponse.Data.Token)
	encryptedJson := base64.StdEncoding.EncodeToString(jsonBytes)

	parsedURL, _ := url.Parse(SecondLoginEndpoint)
	task.BaseTask.Client.Jar.SetCookies(parsedURL, []*http.Cookie{
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
		Client: task.BaseTask.Client,
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
	if err != nil {
		log.Println(err.Error())
		if resp == nil {
			return false, fmt.Sprintf(enums.LoginFailure, err.Error())
		}
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
		if resp == nil {
			return false, fmt.Sprintf(enums.SettingUpFailure, err.Error())
		}
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
	pid := task.BaseTask.ProductInfo.SKU
	if vid, ok := task.BaseTask.ProductInfo.SiteSpecificInfo["VID"].(string); ok && vid != task.BaseTask.ProductInfo.SKU {
		pid = vid
	}

	quantity := task.Input.Quantity
	if quantityLimit, ok := task.BaseTask.ProductInfo.SiteSpecificInfo["QuantityLimit"].(int); ok && quantityLimit > 0 && quantity > quantityLimit {
		quantity = quantityLimit
	}

	data := []byte(u.CreateParams(map[string]string{
		"pid":      pid,
		"quantity": fmt.Sprint(quantity),
	}))
	addToCartResponse := AddToCartResponse{}

	resp, _, err := util.MakeRequest(&util.Request{
		Client: task.BaseTask.Client,
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
			{"referer", task.BaseTask.ProductInfo.ItemURL},
			{"accept-encoding", "gzip, deflate"},
			{"accept-language", "en-US,en;q=0.9"},
		},
		Data:               data,
		ResponseBodyStruct: &addToCartResponse,
	})
	if err != nil {
		log.Println(err.Error())
		if resp == nil {
			return false, fmt.Sprintf(enums.AddingToCartFailure, err.Error())
		}
		// return false, fmt.Sprintf(enums.AddingToCartFailure, err.Error())
	}

	switch resp.StatusCode {
	case 200:
		if addToCartResponse.Message == "Product added to bag" {
			return true, enums.AddingToCartSuccess
		}
		return false, fmt.Sprintf(enums.AddingToCartFailure, addToCartResponse.Message)
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
	cartInfoResponse := GetCartInfoResponse{}
	resp, _, err := util.MakeRequest(&util.Request{
		Client: task.BaseTask.Client,
		Method: "GET",
		URL:    GetCartInfoEndpoint,
		RawHeaders: http.RawHeader{
			{"sec-ch-ua", `" Not;A Brand";v="99", "Google Chrome";v="91", "Chromium";v="91"`},
			{"accept", "application/json, text/javascript, */*; q=0.01"},
			{"x-requested-with", "XMLHttpRequest"},
			{"sec-ch-ua-mobile", "?0"},
			{"user-agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/91.0.4472.106 Safari/537.36"},
			{"sec-fetch-site", "same-origin"},
			{"sec-fetch-site", "cors"},
			{"sec-fetch-dest", "empty"},
			{"referer", GetCartInfoReferer},
			{"accept-encoding", "gzip, deflate"},
			{"accept-language", "en-US,en;q=0.9"},
		},
		ResponseBodyStruct: &cartInfoResponse,
	})
	if err != nil {
		log.Println(err.Error())
		if resp == nil {
			return false, fmt.Sprintf(enums.GettingCartInfoFailure, err.Error())
		}
		// return false, fmt.Sprintf(enums.GettingCartInfoFailure, err.Error())
	}

	switch resp.StatusCode {
	case 200:
		if len(cartInfoResponse.Items) == 0 || cartInfoResponse.Items[0].ShipmentUUID == "" ||
			len(cartInfoResponse.Shipments) == 0 || cartInfoResponse.Shipments[0].SelectedShippingMethod == "" ||
			cartInfoResponse.Csrf.Token == "" {
			return false, fmt.Sprintf(enums.GettingCartInfoFailure, EmptyCSRFValuesError)
		}
		task.ShipmentUUID = cartInfoResponse.Items[0].ShipmentUUID
		task.ShippingMethod = cartInfoResponse.Shipments[0].SelectedShippingMethod
		task.CSRF = cartInfoResponse.Csrf.Token
		return true, enums.GettingCartInfoSuccess
	case 403:
		err = akamai.HandleAkamaiTask(task.BaseTask, BaseEndpoint+"/", BaseEndpoint, AkamaiEndpoint, BaseURL)
		if err != nil {
			return false, fmt.Sprintf(enums.GettingCartInfoFailure, err.Error())
		}
		return false, ""
	}

	return false, fmt.Sprintf(enums.GettingCartInfoFailure, fmt.Sprintf(enums.UnknownError, resp.StatusCode))
}

func (task *Task) ValidateCheckout() (bool, string) {
	resp, _, err := util.MakeRequest(&util.Request{
		Client: task.BaseTask.Client,
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
			{"referer", ValidateCheckoutReferer},
			{"accept-encoding", "gzip, deflate"},
			{"accept-language", "en-US,en;q=0.9"},
		},
	})
	if err != nil {
		log.Println(err.Error())
		if resp == nil {
			return false, fmt.Sprintf(enums.SettingCartInfoFailure, err.Error())
		}
		// return false, fmt.Sprintf(enums.SettingCartInfoFailure, err.Error())
	}

	switch resp.StatusCode {
	case 200:
		return true, enums.SettingCartInfoSuccess
	case 403:
		err = akamai.HandleAkamaiTask(task.BaseTask, BaseEndpoint+"/", BaseEndpoint, AkamaiEndpoint, BaseURL)
		if err != nil {
			return false, fmt.Sprintf(enums.SettingCartInfoFailure, err.Error())
		}
		return false, ""
	}

	return false, fmt.Sprintf(enums.SettingCartInfoFailure, fmt.Sprintf(enums.UnknownError, resp.StatusCode))
}

func (task *Task) SubmitShippingDetails() (bool, string) {
	data := []byte(u.CreateParams(map[string]string{
		"originalShipmentUUID": task.ShipmentUUID,
		"shipmentUUID":         task.ShipmentUUID,
		"emojiValidation":      `(?:[\u2700-\u27bf]|(?:\ud83c[\udde6-\uddff]){2}|[\ud800-\udbff][\udc00-\udfff]|[\u0023-\u0039]\ufe0f?\u20e3|\u3299|\u3297|\u303d|\u3030|\u24c2|\ud83c[\udd70-\udd71]|\ud83c[\udd7e-\udd7f]|\ud83c\udd8e|\ud83c[\udd91-\udd9a]|\ud83c[\udde6-\uddff]|\ud83c[\ude01-\ude02]|\ud83c\ude1a|\ud83c\ude2f|\ud83c[\ude32-\ude3a]|\ud83c[\ude50-\ude51]|\u203c|\u2049|[\u25aa-\u25ab]|\u25b6|\u25c0|[\u25fb-\u25fe]|\u00a9|\u00ae|\u2122|\u2139|\ud83c\udc04|[\u2600-\u26FF]|\u2b05|\u2b06|\u2b07|\u2b1b|\u2b1c|\u2b50|\u2b55|\u231a|\u231b|\u2328|\u23cf|[\u23e9-\u23f3]|[\u23f8-\u23fa]|\ud83c\udccf|\u2934|\u2935|[\u2190-\u21ff]|\uFE0F)`,
		"shipping-countries":   `[{"id":"US","fields":{"firstName":{"maxLength":"15","mandatory":"true"},"lastName":{"maxLength":"25","mandatory":"true"},"address1":{"maxLength":"32","mandatory":"true"},"address2":{"maxLength":"32","mandatory":"false"},"city":{"maxLength":"25","mandatory":"true"},"state":{"maxLength":"","mandatory":"true"},"postalCode":{"maxLength":"10","mandatory":"true"},"country":{"maxLength":"","mandatory":"true"}}},{"id":"CA","fields":{"firstName":{"maxLength":"15","mandatory":"true"},"lastName":{"maxLength":"25","mandatory":"true"},"address1":{"maxLength":"32","mandatory":"true"},"address2":{"maxLength":"32","mandatory":"false"},"city":{"maxLength":"25","mandatory":"true"},"postalCode":{"maxLength":"10","mandatory":"true"},"country":{"maxLength":"","mandatory":"true"},"state":{"maxLength":"","mandatory":"true"}}},{"id":"default","fields":{"firstName":{"maxLength":"15","mandatory":"true"},"lastName":{"maxLength":"25","mandatory":"true"},"address1":{"maxLength":"32","mandatory":"true"},"address2":{"maxLength":"32","mandatory":"false"},"city":{"maxLength":"25","mandatory":"true"},"postalCode":{"maxLength":"10","mandatory":"false"},"country":{"maxLength":"","mandatory":"true"},"state":{"maxLength":"","mandatory":"false"}}}]`,
		"dwfrm_shipping_shippingAddress_addressFields_country":          task.BaseTask.Profile.ShippingAddress.CountryCode,
		"dwfrm_shipping_shippingAddress_addressFields_firstName":        task.BaseTask.Profile.ShippingAddress.FirstName,
		"dwfrm_shipping_shippingAddress_addressFields_lastName":         task.BaseTask.Profile.ShippingAddress.LastName,
		"dwfrm_shipping_shippingAddress_addressFields_address1":         task.BaseTask.Profile.ShippingAddress.Address1,
		"dwfrm_shipping_shippingAddress_addressFields_address2":         task.BaseTask.Profile.ShippingAddress.Address2,
		"dwfrm_shipping_shippingAddress_addressFields_postalCode":       task.BaseTask.Profile.ShippingAddress.ZipCode,
		"dwfrm_shipping_shippingAddress_addressFields_city":             task.BaseTask.Profile.ShippingAddress.City,
		"dwfrm_shipping_shippingAddress_addressFields_states_stateCode": task.BaseTask.Profile.ShippingAddress.StateCode,
		"dwfrm_shipping_shippingAddress_addressFields_phone":            task.BaseTask.Profile.PhoneNumber,
		"appSessionID":                         ``,
		"dwfrm_shipping_shippingAddress_email": task.BaseTask.Profile.Email,
		"shippingMethod":                       task.ShippingMethod,
		"csrf_token":                           task.CSRF,
	}))

	resp, _, err := util.MakeRequest(&util.Request{
		Client: task.BaseTask.Client,
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
			{"referer", SubmitShippingInfoReferer},
			{"accept-encoding", "gzip, deflate"},
			{"accept-language", "en-US,en;q=0.9"},
		},
		Data: data,
	})
	if err != nil {
		log.Println(err.Error())
		if resp == nil {
			return false, fmt.Sprintf(enums.SettingShippingInfoFailure, err.Error())
		}
		// return false, fmt.Sprintf(enums.SettingShippingInfoFailure, err.Error())
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

func (task *Task) EstablishAppSession() (bool, string) {
	establishAppSessionResponse := EstablishAppSessionResponse{}
	resp, _, err := util.MakeRequest(&util.Request{
		Client: task.BaseTask.Client,
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
			{"referer", EstablishAppSessionReferer},
			{"accept-encoding", `gzip, deflate, br`},
			{"accept-language", `en-US,en;q=0.9`},
		},
		ResponseBodyStruct: &establishAppSessionResponse,
	})
	if err != nil {
		fmt.Println(err.Error())
		if resp == nil {
			return false, fmt.Sprintf(enums.GettingBillingInfoFailure, err.Error())
		}
		// return false, fmt.Sprintf(enums.GettingBillingInfoFailure, err.Error())
	}

	switch resp.StatusCode {
	case 200:
		task.PaymentData = establishAppSessionResponse.PaymentData
		task.Total = establishAppSessionResponse.Order.Totals.GrandTotalValue.Value
		return true, enums.GettingBillingInfoSuccess
	case 403:
		err = akamai.HandleAkamaiTask(task.BaseTask, BaseEndpoint+"/", BaseEndpoint, AkamaiEndpoint, BaseURL)
		if err != nil {
			return false, fmt.Sprintf(enums.GettingBillingInfoFailure, err.Error())
		}
		return false, ""
	}

	return false, fmt.Sprintf(enums.GettingBillingInfoFailure, fmt.Sprintf(enums.UnknownError, resp.StatusCode))
}

func (task *Task) GetPaysheetAE() (bool, string) {
	resp, _, err := util.MakeRequest(&util.Request{
		Client: task.BaseTask.Client,
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
	if err != nil {
		fmt.Println(err.Error())
		if resp == nil {
			return false, fmt.Sprintf(enums.SettingBillingInfoFailure, err.Error())
		}
		// return false, fmt.Sprintf(enums.SettingBillingInfoFailure, err.Error())
	}

	switch resp.StatusCode {
	case 200:
		task.PaymentData.Config.PaymentAE = resp.Header.Get("x-disney-paysheet-ae")
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

func (task *Task) GetCardToken() (bool, string) {
	getCardTokenRequest := GetCardTokenRequest{
		Pan:  task.BaseTask.Profile.CreditCard.CardNumber,
		Type: "CREDIT_CARD",
	}
	getCardTokenResponse := GetCardTokenResponse{}
	resp, _, err := util.MakeRequest(&util.Request{
		Client: task.BaseTask.Client,
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
			{"referer", GetCardTokenReferer},
			{"accept-encoding", `gzip, deflate, br`},
		},
		RequestBodyStruct:  getCardTokenRequest,
		ResponseBodyStruct: &getCardTokenResponse,
	})
	if err != nil {
		fmt.Println(err.Error())
		if resp == nil {
			return false, fmt.Sprintf(enums.GettingOrderInfoFailure, err.Error())
		}
		// return false, fmt.Sprintf(enums.GettingOrderInfoFailure, err.Error())
	}

	switch resp.StatusCode {
	case 200:
		task.PaymentData.CardInfo = getCardTokenResponse
		return true, enums.GettingOrderInfoSuccess
	case 403:
		err = akamai.HandleAkamaiTask(task.BaseTask, BaseEndpoint+"/", BaseEndpoint, AkamaiEndpoint, BaseURL)
		if err != nil {
			return false, fmt.Sprintf(enums.GettingOrderInfoFailure, err.Error())
		}
		return false, ""
	}

	return false, fmt.Sprintf(enums.GettingOrderInfoFailure, fmt.Sprintf(enums.UnknownError, resp.StatusCode))
}

func (task *Task) SubmitOrder() (bool, string) {
	placeOrderRequest := PlaceOrderRequest{
		Cards: []Cards{
			{
				SecurityCode:      task.BaseTask.Profile.CreditCard.CVV,
				CardProductName:   task.PaymentData.CardInfo.CardProductName,
				IsDisneyVisa:      task.PaymentData.CardInfo.IsDisneyVisa,
				CardToken:         task.PaymentData.CardInfo.Token,
				CardDisplayNumber: task.PaymentData.CardInfo.Masked[len(task.PaymentData.CardInfo.Masked)-8:],
				Expiration:        task.BaseTask.Profile.CreditCard.ExpMonth + task.BaseTask.Profile.CreditCard.ExpYear,
				CardholderName:    task.BaseTask.Profile.CreditCard.CardholderName,
				ProfileName:       ProfileName{},
				ProfilePhone:      ProfilePhone{},
				BillingAddress: BillingAddress{
					Country:    task.BaseTask.Profile.BillingAddress.CountryCode,
					Line1:      task.BaseTask.Profile.BillingAddress.Address1,
					Line2:      task.BaseTask.Profile.BillingAddress.Address2,
					City:       task.BaseTask.Profile.BillingAddress.City,
					State:      task.BaseTask.Profile.BillingAddress.StateCode,
					PostalCode: task.BaseTask.Profile.BillingAddress.ZipCode,
				},
				SaveToProfile:   false,
				IsWalletDefault: false,
				IsWalletEdit:    false,
				Issuer:          task.PaymentData.CardInfo.Issuer,
				Type:            task.PaymentData.CardInfo.Type,
				IsStoredCard:    false,
				Amount:          task.Total,
			},
		},
		FraudParameters: []FraudParameters{
			{
				Group: "identification_info",
				Params: Params{
					EmailAddress: task.BaseTask.Profile.Email,
				},
			},
			{
				Group: "billing_info",
				Params: Params{
					DayPhone: task.BaseTask.Profile.PhoneNumber,
				},
			},
		},
	}
	data, _ := json.Marshal(placeOrderRequest)
	placeOrderResponse := PlaceOrderResponse{}
	resp, _, err := util.MakeRequest(&util.Request{
		Client: task.BaseTask.Client,
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
	if err != nil {
		fmt.Println(err.Error())
		if resp == nil {
			return false, fmt.Sprintf(enums.CheckingOutFailure, err.Error())
		}
		// return false, fmt.Sprintf(enums.CheckingOutFailure, err.Error())
	}

	if placeOrderResponse.SuggestedErrorKey == "d_credit_card" {
		return false, enums.CardDeclined
	}

	switch resp.StatusCode {
	case 200:
		return true, enums.CheckingOutSuccess
	case 403:
		err = akamai.HandleAkamaiTask(task.BaseTask, BaseEndpoint+"/", BaseEndpoint, AkamaiEndpoint, BaseURL)
		if err != nil {
			return false, fmt.Sprintf(enums.CheckingOutFailure, err.Error())
		}
		return false, ""
	}

	return false, fmt.Sprintf(enums.CheckingOutFailure, fmt.Sprintf(enums.UnknownError, resp.StatusCode))

}
