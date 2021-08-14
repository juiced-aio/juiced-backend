package newegg

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/url"

	"backend.juicedbot.io/juiced.client/http"
	"backend.juicedbot.io/juiced.sitescripts/util"
)

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

func (task *Task) PlaceOrder() bool {
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
		return false
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
		return false
	}

	if placeOrderResponse.Result != "Success" {
		return false
	}

	task.TaskInfo.VBVToken = placeOrderResponse.Vbvdata.Jwttoken
	task.TaskInfo.CardBin = placeOrderResponse.Vbvdata.Cardbin

	return true
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
