package pokemoncenter

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/url"
	"regexp"
	"strings"
	"time"

	"backend.juicedbot.io/juiced.infrastructure/common"
	"backend.juicedbot.io/juiced.infrastructure/common/enums"
	"backend.juicedbot.io/juiced.sitescripts/util"
)

const MAX_RETRIES = 5

func (task *Task) PublishEvent(status enums.TaskStatus, eventType enums.TaskEventType) {
	task.TaskInfo.Task.SetTaskStatus(status)
	task.TaskInfo.EventBus.PublishTaskEvent(status, eventType, nil, task.TaskInfo.Task.ID)
}

func (task *Task) CheckForStop() bool {
	if task.TaskInfo.StopFlag {
		task.PublishEvent(enums.TaskIdle, enums.TaskStop)
		return true
	}
	return false
}

func (task *Task) RunTask() {
	defer func() {
		if r := recover(); r != nil {
			task.PublishEvent(fmt.Sprintf(enums.TaskFailed, r), enums.TaskFail)
		} else {
			task.PublishEvent(enums.TaskIdle, enums.TaskStop)
		}
		task.TaskInfo.StopFlag = true
	}()

	if task.TaskInfo.Task.TaskDelay == 0 {
		task.TaskInfo.Task.TaskDelay = 2000
	}
	if task.TaskInfo.Task.TaskQty <= 0 {
		task.TaskInfo.Task.TaskQty = 1
	}

	// 1. Login/LoginGuest
	if task.Input.TaskType == enums.TaskTypeAccount && task.Input.Email != "" && task.Input.Password != "" {
		task.PublishEvent(enums.LoggingIn, enums.TaskUpdate)
		if success, _ := task.RunUntilSuccessful(task.Login, MAX_RETRIES); !success {
			return
		}
	} else {
		task.PublishEvent(enums.SettingUp, enums.TaskUpdate)
		if success, _ := task.RunUntilSuccessful(task.LoginGuest, MAX_RETRIES); !success {
			return
		}
	}

	// 2. RefreshLogin (in background)
	task.RefreshAt = time.Now().Unix() + 1800
	go task.RefreshLogin()

	// 3. Encrypt card details
	task.PublishEvent(enums.EncryptingCardInfo, enums.TaskUpdate)
	if success, _ := task.RunUntilSuccessful(task.RetrieveEncryptedCardDetails, MAX_RETRIES); !success {
		return
	}

	// 4. WaitForMonitor
	task.PublishEvent(enums.WaitingForMonitor, enums.TaskUpdate)
	needToStop := task.WaitForMonitor()
	if needToStop {
		return
	}

	task.TaskInfo.StartTime = time.Now()

	// 5. AddToCart
	task.PublishEvent(enums.AddingToCart, enums.TaskUpdate)
	if success, _ := task.RunUntilSuccessful(task.AddToCart, -1); !success {
		return
	}

	// 6. Submit email address
	if task.Input.TaskType == enums.TaskTypeGuest {
		task.PublishEvent(enums.SettingEmailAddress, enums.TaskUpdate)
		if success, _ := task.RunUntilSuccessful(task.SubmitEmailAddress, MAX_RETRIES); !success {
			return
		}
	}

	// 7. Submit address details
	task.PublishEvent(enums.SettingShippingInfo, enums.TaskUpdate)
	if success, _ := task.RunUntilSuccessful(task.SubmitAddressDetails, MAX_RETRIES); !success {
		return
	}

	// 8. Submit payment details
	task.PublishEvent(enums.SettingBillingInfo, enums.TaskUpdate)
	if success, _ := task.RunUntilSuccessful(task.SubmitPaymentDetails, MAX_RETRIES); !success {
		return
	}

	// 9. Checkout
	task.PublishEvent(enums.CheckingOut, enums.TaskUpdate)
	success, status := task.RunUntilSuccessful(task.Checkout, MAX_RETRIES)

	go util.ProcessCheckout(&util.ProcessCheckoutInfo{
		TaskInfo:     task.TaskInfo,
		Success:      success,
		Status:       status,
		Embeds:       task.CreatePokemonCenterEmbed(status, task.StockData.ImageURL),
		Content:      "",
		ItemName:     task.StockData.ItemName,
		ImageURL:     task.StockData.ImageURL,
		Sku:          task.StockData.SKU,
		Retailer:     enums.PokemonCenter,
		Price:        task.StockData.Price,
		Quantity:     task.TaskInfo.Task.TaskQty,
		MsToCheckout: time.Since(task.TaskInfo.StartTime).Milliseconds(),
	})

	task.TaskInfo.EndTime = time.Now()

	log.Println("STARTED AT: " + task.TaskInfo.StartTime.String())
	log.Println("  ENDED AT: " + task.TaskInfo.EndTime.String())
	log.Println("TIME TO CHECK OUT: ", task.TaskInfo.EndTime.Sub(task.TaskInfo.StartTime).Milliseconds())
}

func (task *Task) Login() (bool, string) {
	loginResponse := LoginResponse{}

	params := url.Values{}
	params.Add("username", task.Input.Email)
	params.Add("password", task.Input.Password)
	params.Add("grant_type", "password")
	params.Add("role", "REGISTERED")
	params.Add("scope", "pokemon")

	resp, _, err := util.MakeRequest(&util.Request{
		Client: task.TaskInfo.Client,
		Method: "POST",
		URL:    LoginEndpoint,
		RawHeaders: [][2]string{
			{"content-length", fmt.Sprint(bytes.NewReader([]byte(params.Encode())).Size())},
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
			{"referer", LoginRefererEndpoint},
			{"accept-encoding", "gzip, deflate, br"},
			{"accept-language", "en-US,en;q=0.9"},
		},
		ResponseBodyStruct: loginResponse,
		Data:               []byte(params.Encode()),
	})
	if err != nil {
		return false, fmt.Sprintf(enums.LoginFailure, err.Error())
	}

	switch resp.StatusCode {
	case 200:
		task.AccessToken = loginResponse.AccessToken
		return true, enums.LoginSuccess
	case 401:
		return false, fmt.Sprintf(enums.LoginFailure, err.Error())
	}

	return false, fmt.Sprintf(enums.LoginFailure, UnknownError)
}

func (task *Task) LoginGuest() (bool, string) {
	resp, _, err := util.MakeRequest(&util.Request{
		Client: task.TaskInfo.Client,
		Method: "GET",
		URL:    AuthKeyEndpoint,
		RawHeaders: [][2]string{
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
			{"referer", AuthKeyRefererEndpoint},
			{"accept-encoding", "gzip, deflate, br"},
			{"accept-language", "en-US,en;q=0.9"},
		},
	})
	if err != nil {
		return false, fmt.Sprintf(enums.SettingUpFailure, err.Error())
	}

	switch resp.StatusCode {
	case 200:
		rawHeader := resp.Header.Get("Set-Cookie")
		re := regexp.MustCompile("({)(.*?)(})")
		match := re.FindStringSubmatch(rawHeader)
		fmt.Println(match[0])

		accessToken := AccessToken{}
		json.Unmarshal([]byte(match[0]), &accessToken)
		task.AccessToken = accessToken.AccessToken
		return true, enums.SettingUpSuccess
	}

	return false, fmt.Sprintf(enums.SettingUpFailure, UnknownError)
}

func (task *Task) RefreshLogin() {
	defer func() {
		if r := recover(); r != nil {
			task.RefreshLogin()
		}
	}()

	for {
		if task.RefreshAt == 0 || time.Now().Unix() > task.RefreshAt {
			if task.Input.TaskType == enums.TaskTypeAccount {
				if success, _ := task.RunUntilSuccessful(task.Login, MAX_RETRIES); !success {
					return
				}
			} else {
				if success, _ := task.RunUntilSuccessful(task.LoginGuest, MAX_RETRIES); !success {
					return
				}
			}
			task.RefreshAt = time.Now().Unix() + 1800
		}
		time.Sleep(time.Millisecond * common.MS_TO_WAIT)
	}
}

func (task *Task) RetrieveEncryptedCardDetails() (bool, string) {
	var err error

	// 1. Retrieve public key for encryption
	if success, _ := task.RunUntilSuccessful(task.RetrievePublicKey, MAX_RETRIES); !success {
		return false, fmt.Sprintf(enums.EncryptingCardInfoFailure, RetrieveCyberSourcePublicKeyError)
	}

	// 2. Encrypt using CyberSourceV2 encryption
	task.CyberSecureInfo.PublicToken, err = CyberSourceV2(task.CyberSecureInfo.PublicKey, task.TaskInfo.Profile.CreditCard)
	if task.CyberSecureInfo.PublicToken == "" || err != nil {
		errorMessage := CyberSourceEncryptionError
		if err != nil {
			errorMessage = err.Error()
		}
		return false, fmt.Sprintf(enums.EncryptingCardInfoFailure, errorMessage)
	}

	// 3. Retrieve CyberSourceV2 Token
	if success, _ := task.RunUntilSuccessful(task.RetrieveToken, MAX_RETRIES); !success {
		return false, fmt.Sprintf(enums.EncryptingCardInfoFailure, RetrieveCyberSourceTokenError)
	}

	// 4. Retrieve JTI from CyberSourceInfo
	task.CyberSecureInfo.JtiToken, err = retrievePaymentToken(task.CyberSecureInfo.Privatekey)
	if task.CyberSecureInfo.JtiToken == "" || err != nil {
		errorMessage := RetrieveCyberSourcePaymentTokenError
		if err != nil {
			errorMessage = err.Error()
		}
		return false, fmt.Sprintf(enums.EncryptingCardInfoFailure, errorMessage)
	}

	return true, enums.EncryptingCardInfoSuccess
}

func (task *Task) RetrievePublicKey() (bool, string) {
	paymentKeyResponse := PaymentKeyResponse{}

	resp, _, err := util.MakeRequest(&util.Request{
		Client: task.TaskInfo.Client,
		Method: "GET",
		URL:    PublicPaymentKeyEndpoint,
		RawHeaders: [][2]string{
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
			{"referer", PublicPaymentKeyRefererEndpoint},
			{"accept-encoding", "gzip, deflate, br"},
			{"accept-language", "en-US,en;q=0.9"},
			{"Cookie", "auth={\"access_token\":\"" + task.AccessToken + "\",\"token_type\":\"bearer\",\"expires_in\":604799,\"scope\":\"pokemon\",\"role\":\"PUBLIC\",\"roles\":[\"PUBLIC\"]}"},
		},
		ResponseBodyStruct: &paymentKeyResponse,
	})
	if err != nil {
		return false, fmt.Sprintf(enums.EncryptingCardInfoFailure, err.Error())
	}

	switch resp.StatusCode {
	case 200:
		task.CyberSecureInfo.PublicKey = paymentKeyResponse.KeyId
		return true, ""
	}

	return false, fmt.Sprintf(enums.EncryptingCardInfoFailure, UnknownError)
}

func (task *Task) RetrieveToken() (bool, string) {
	resp, _, err := util.MakeRequest(&util.Request{
		Client: task.TaskInfo.Client,
		Method: "POST",
		URL:    CyberSourceTokenEndpoint,
		RawHeaders: [][2]string{
			{"content-length", fmt.Sprint(bytes.NewReader([]byte(task.CyberSecureInfo.PublicToken)).Size())},
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
			{"referer", CyberSourceTokenRefererEndpoint},
			{"accept-encoding", "gzip, deflate, br"},
			{"accept-language", "en-US,en;q=0.9"},
		},
		Data: []byte(task.CyberSecureInfo.PublicToken),
	})
	if err != nil {
		return false, fmt.Sprintf(enums.EncryptingCardInfoFailure, err.Error())
	}

	switch resp.StatusCode {
	case 200:
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return false, fmt.Sprintf(enums.EncryptingCardInfoFailure, err.Error())
		}
		task.CyberSecureInfo.Privatekey = string(body)
		return true, ""
	}

	return false, fmt.Sprintf(enums.EncryptingCardInfoFailure, UnknownError)
}

func (task *Task) WaitForMonitor() bool {
	for {
		needToStop := task.CheckForStop()
		if needToStop {
			return true
		}
		if task.StockData.AddToCartForm != "" {
			return false
		}
		time.Sleep(time.Millisecond * common.MS_TO_WAIT)
	}
}

func (task *Task) AddToCart() (bool, string) {
	addToCartRequest := AddToCartRequest{
		ProductUri:    task.StockData.AddToCartForm,
		Quantity:      task.TaskInfo.Task.TaskQty,
		Configuration: "",
	}
	addToCartResponse := AddToCartResponse{}

	addToCartRequestBytes, err := json.Marshal(addToCartRequest)
	if err != nil {
		return false, fmt.Sprintf(enums.AddingToCartFailure, err.Error())
	}

	resp, _, err := util.MakeRequest(&util.Request{
		Client: task.TaskInfo.Client,
		Method: "POST",
		URL:    AddToCartEndpoint,
		RawHeaders: [][2]string{
			{"content-length", fmt.Sprint(bytes.NewReader(addToCartRequestBytes).Size())},
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
			{"referer", fmt.Sprintf(AddToCartRefererEndpoint, task.StockData.SKU)},
			{"accept-encoding", "gzip, deflate, br"},
			{"accept-language", "en-US,en;q=0.9"},
			{"Cookie", "auth={\"access_token\":\"" + task.AccessToken + "\",\"token_type\":\"bearer\",\"expires_in\":604799,\"scope\":\"pokemon\",\"role\":\"PUBLIC\",\"roles\":[\"PUBLIC\"]}"},
		},
		ResponseBodyStruct: &addToCartResponse,
		RequestBodyStruct:  &addToCartRequest,
	})
	if err != nil {
		return false, fmt.Sprintf(enums.AddingToCartFailure, err.Error())
	}

	switch resp.StatusCode {
	case 200:
		if addToCartResponse.Type == "carts.line-item" {
			if addToCartResponse.Quantity != task.TaskInfo.Task.TaskQty {
				return false, fmt.Sprintf(enums.AddingToCartFailure, fmt.Sprintf(AddToCartQuantityError, task.TaskInfo.Task.TaskQty, addToCartResponse.Quantity))
			} else {
				return true, enums.AddingToCartSuccess
			}
		}
	}

	return false, fmt.Sprintf(enums.AddingToCartFailure, UnknownError)
}

func (task *Task) SubmitEmailAddress() (bool, string) {
	emailRequest := EmailRequest{
		Email: task.TaskInfo.Profile.Email,
	}

	emailBytes, err := json.Marshal(emailRequest)
	if err != nil {
		return false, fmt.Sprintf(enums.SettingEmailAddressFailure, err.Error())
	}
	resp, _, err := util.MakeRequest(&util.Request{
		Client: task.TaskInfo.Client,
		Method: "POST",
		URL:    SubmitEmailEndpoint,
		RawHeaders: [][2]string{
			{"content-length", fmt.Sprint(bytes.NewReader(emailBytes).Size())},
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
			{"referer", SubmitEmailRefererEndpoint},
			{"accept-encoding", "gzip, deflate, br"},
			{"accept-language", "en-US,en;q=0.9"},
			{"Cookie", "auth={\"access_token\":\"" + task.AccessToken + "\",\"token_type\":\"bearer\",\"expires_in\":604799,\"scope\":\"pokemon\",\"role\":\"PUBLIC\",\"roles\":[\"PUBLIC\"]}"},
		},
		RequestBodyStruct: &emailRequest,
	})
	if err != nil {
		return false, fmt.Sprintf(enums.SettingEmailAddressFailure, err.Error())
	}

	switch resp.StatusCode {
	case 200:
		return true, enums.SettingEmailAddressSuccess
	}

	return false, fmt.Sprintf(enums.SettingEmailAddressFailure, UnknownError)
}

func (task *Task) SubmitAddressDetails() (bool, string) {
	if task.Input.TaskType == enums.TaskTypeAccount {
		return true, ""
	}

	submitAddressRequest := SubmitAddressRequest{
		Billing: Address{
			FamilyName:      task.TaskInfo.Profile.BillingAddress.LastName,
			GivenName:       task.TaskInfo.Profile.BillingAddress.FirstName,
			StreetAddress:   task.TaskInfo.Profile.BillingAddress.Address1,
			ExtendedAddress: task.TaskInfo.Profile.BillingAddress.Address2,
			Locality:        task.TaskInfo.Profile.BillingAddress.City,
			Region:          task.TaskInfo.Profile.BillingAddress.StateCode,
			PostalCode:      task.TaskInfo.Profile.BillingAddress.ZipCode,
			CountryName:     task.TaskInfo.Profile.BillingAddress.CountryCode,
			PhoneNumber:     task.TaskInfo.Profile.PhoneNumber,
		},
		Shipping: Address{
			FamilyName:      task.TaskInfo.Profile.ShippingAddress.LastName,
			GivenName:       task.TaskInfo.Profile.ShippingAddress.FirstName,
			StreetAddress:   task.TaskInfo.Profile.ShippingAddress.Address1,
			ExtendedAddress: task.TaskInfo.Profile.ShippingAddress.Address2,
			Locality:        task.TaskInfo.Profile.ShippingAddress.City,
			Region:          task.TaskInfo.Profile.ShippingAddress.StateCode,
			PostalCode:      task.TaskInfo.Profile.ShippingAddress.ZipCode,
			CountryName:     task.TaskInfo.Profile.ShippingAddress.CountryCode,
			PhoneNumber:     task.TaskInfo.Profile.PhoneNumber,
		},
	}

	submitAddressRequestBytes, err := json.Marshal(submitAddressRequest)
	if err != nil {
		return false, fmt.Sprintf(enums.SettingShippingInfoFailure, err.Error())
	}
	resp, _, err := util.MakeRequest(&util.Request{
		Client: task.TaskInfo.Client,
		Method: "POST",
		URL:    SubmitAddressEndpoint,
		RawHeaders: [][2]string{
			{"content-length", fmt.Sprint(bytes.NewReader(submitAddressRequestBytes).Size())},
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
			{"referer", SubmitAddresRefererEndpoint},
			{"accept-encoding", "gzip, deflate, br"},
			{"accept-language", "en-US,en;q=0.9"},
			{"Cookie", "auth={\"access_token\":\"" + task.AccessToken + "\",\"token_type\":\"bearer\",\"expires_in\":604799,\"scope\":\"pokemon\",\"role\":\"PUBLIC\",\"roles\":[\"PUBLIC\"]}"},
		},
		RequestBodyStruct: &submitAddressRequest,
	})
	if err != nil {
		return false, fmt.Sprintf(enums.SettingShippingInfoFailure, err.Error())
	}

	switch resp.StatusCode {
	case 200:
		return true, enums.SettingShippingInfoSuccess
	}

	return false, fmt.Sprintf(enums.SettingShippingInfoFailure, UnknownError)
}

func (task *Task) SubmitPaymentDetails() (bool, string) {
	submitPaymentRequest := SubmitPaymentRequest{
		PaymentDisplay: fmt.Sprintf("%s %s/%s", task.TaskInfo.Profile.CreditCard.CardType, task.TaskInfo.Profile.CreditCard.ExpMonth, task.TaskInfo.Profile.CreditCard.ExpYear),
		PaymentKey:     task.CyberSecureInfo.PublicKey,
		PaymentToken:   task.CyberSecureInfo.JtiToken,
	}
	submitPaymentResponse := SubmitPaymentResponse{}

	paymentDetailsBytes, err := json.Marshal(submitPaymentRequest)
	if err != nil {
		return false, fmt.Sprintf(enums.SettingBillingInfoFailure, err.Error())
	}
	resp, _, err := util.MakeRequest(&util.Request{
		Client: task.TaskInfo.Client,
		Method: "POST",
		URL:    SubmitPaymentDetailsEndpoint,
		RawHeaders: [][2]string{
			{"content-length", fmt.Sprint(bytes.NewReader(paymentDetailsBytes).Size())},
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
			{"referer", SubmitPaymentDetailsRefererEndpoint},
			{"accept-encoding", "gzip, deflate, br"},
			{"accept-language", "en-US,en;q=0.9"},
			{"Cookie", "auth={\"access_token\":\"" + task.AccessToken + "\",\"token_type\":\"bearer\",\"expires_in\":604799,\"scope\":\"pokemon\",\"role\":\"PUBLIC\",\"roles\":[\"PUBLIC\"]}"},
		},
		RequestBodyStruct:  &submitPaymentRequest,
		ResponseBodyStruct: &submitPaymentResponse,
	})
	if err != nil {
		return false, fmt.Sprintf(enums.SettingBillingInfoFailure, err.Error())
	}

	switch resp.StatusCode {
	case 200:
		task.CheckoutUri = submitPaymentResponse.Self.Uri
		return true, enums.SettingBillingInfoSuccess
	}

	return false, fmt.Sprintf(enums.SettingBillingInfoFailure, UnknownError)
}

func (task *Task) Checkout() (bool, string) {
	checkoutDetailsRequest := CheckoutDetailsRequest{PurchaseFrom: strings.Replace(task.CheckoutUri, "paymentmethods", "purchases", 1) + "/form"}

	submitAddressRequestBytes, err := json.Marshal(checkoutDetailsRequest)
	if err != nil {
		return false, fmt.Sprintf(enums.CheckingOutFailure, err.Error())
	}
	resp, _, err := util.MakeRequest(&util.Request{
		Client: task.TaskInfo.Client,
		Method: "POST",
		URL:    CheckoutEndpoint,
		RawHeaders: [][2]string{
			{"content-length", fmt.Sprint(bytes.NewReader(submitAddressRequestBytes).Size())},
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
			{"referer", CheckoutRefererEndpoint},
			{"accept-encoding", "gzip, deflate, br"},
			{"accept-language", "en-US,en;q=0.9"},
			{"Cookie", "auth={\"access_token\":\"" + task.AccessToken + "\",\"token_type\":\"bearer\",\"expires_in\":604799,\"scope\":\"pokemon\",\"role\":\"PUBLIC\",\"roles\":[\"PUBLIC\"]}"},
		},
		RequestBodyStruct: &checkoutDetailsRequest,
	})
	if err != nil {
		return false, fmt.Sprintf(enums.CheckingOutFailure, err.Error())
	}

	switch resp.StatusCode {
	case 200:
		return true, enums.CheckedOut
	}

	return false, fmt.Sprintf(enums.CheckingOutFailure, UnknownError)
}

////////////
// UNUSED //
////////////

// Validate address details
func (task *Task) SubmitAddressDetailsValidate() (bool, string) {
	submitAddressRequest := SubmitAddressRequest{
		Billing: Address{
			FamilyName:      task.TaskInfo.Profile.BillingAddress.LastName,
			GivenName:       task.TaskInfo.Profile.BillingAddress.FirstName,
			StreetAddress:   task.TaskInfo.Profile.BillingAddress.Address1,
			ExtendedAddress: task.TaskInfo.Profile.BillingAddress.Address2,
			Locality:        task.TaskInfo.Profile.BillingAddress.City,
			Region:          task.TaskInfo.Profile.BillingAddress.StateCode,
			PostalCode:      task.TaskInfo.Profile.BillingAddress.ZipCode,
			CountryName:     task.TaskInfo.Profile.BillingAddress.CountryCode,
			PhoneNumber:     task.TaskInfo.Profile.PhoneNumber,
		},
		Shipping: Address{
			FamilyName:      task.TaskInfo.Profile.ShippingAddress.LastName,
			GivenName:       task.TaskInfo.Profile.ShippingAddress.FirstName,
			StreetAddress:   task.TaskInfo.Profile.ShippingAddress.Address1,
			ExtendedAddress: task.TaskInfo.Profile.ShippingAddress.Address2,
			Locality:        task.TaskInfo.Profile.ShippingAddress.City,
			Region:          task.TaskInfo.Profile.ShippingAddress.StateCode,
			PostalCode:      task.TaskInfo.Profile.ShippingAddress.ZipCode,
			CountryName:     task.TaskInfo.Profile.ShippingAddress.CountryCode,
			PhoneNumber:     task.TaskInfo.Profile.PhoneNumber,
		},
	}

	//json marshal this for content length.
	submitAddressRequestBytes, err := json.Marshal(submitAddressRequest)
	if err != nil {
		log.Fatal("Marshal payload failed with error " + err.Error())
	}
	resp, _, err := util.MakeRequest(&util.Request{
		Client: task.TaskInfo.Client,
		Method: "POST",
		URL:    SubmitAddressValidateEndpoint,
		RawHeaders: [][2]string{
			{"content-length", fmt.Sprint(bytes.NewReader(submitAddressRequestBytes).Size())},
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
			{"referer", SubmitAddresValidateRefererEndpoint}, //double check this endpoint
			{"accept-encoding", "gzip, deflate, br"},
			{"accept-language", "en-US,en;q=0.9"},
			{"Cookie", "auth={\"access_token\":\"" + task.AccessToken + "\",\"token_type\":\"bearer\",\"expires_in\":604799,\"scope\":\"pokemon\",\"role\":\"PUBLIC\",\"roles\":[\"PUBLIC\"]}"},
		},
		RequestBodyStruct: &submitAddressRequest,
	})
	if err != nil {
		fmt.Println(err.Error())
	}

	switch resp.StatusCode {
	case 200:
		//read response for errors etc...
		return true, ""
	}

	return false, ""
}
