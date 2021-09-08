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

	"backend.juicedbot.io/juiced.encryption/cybersource"
	"backend.juicedbot.io/juiced.infrastructure/entities"
	"backend.juicedbot.io/juiced.infrastructure/enums"
	u "backend.juicedbot.io/juiced.infrastructure/util"
	"backend.juicedbot.io/juiced.sitescripts/util"
)

const MAX_RETRIES = 5

func CreateTask(input entities.TaskInput, baseTask *entities.BaseTask) (entities.RetailerTask, error) {
	pokemonCenterTaskInput, err := ValidateTaskInput(input.SiteSpecificInfo)
	if err != nil {
		return nil, err
	}
	return &Task{
		Input:              input,
		PokemonCenterInput: pokemonCenterTaskInput,
		BaseTask:           baseTask,
	}, nil
}

func (task *Task) GetSetupFunctions() []entities.TaskFunction {
	setupTaskFunctions := []entities.TaskFunction{}

	if task.PokemonCenterInput.TaskType == enums.TaskTypeAccount {
		setupTaskFunctions = append(setupTaskFunctions, []entities.TaskFunction{
			// 1. Login
			{
				Function:    task.Login,
				StatusBegin: enums.LoggingIn,
				MaxRetries:  MAX_RETRIES,
			},
			// 2. Refresh Login
			{
				Function:        task.Login,
				StatusBegin:     enums.LoggingIn,
				MaxRetries:      MAX_RETRIES,
				RefreshFunction: true,
				RefreshEvery:    1800,
			},
		}...)
	} else {
		setupTaskFunctions = append(setupTaskFunctions, []entities.TaskFunction{
			// 1. LoginGuest
			{
				Function:    task.LoginGuest,
				StatusBegin: enums.SettingUp,
				MaxRetries:  MAX_RETRIES,
			},
			// 2. Refresh LoginGuest
			{
				Function:        task.LoginGuest,
				StatusBegin:     enums.SettingUp,
				MaxRetries:      MAX_RETRIES,
				RefreshFunction: true,
				RefreshEvery:    1800,
			},
		}...)
	}

	setupTaskFunctions = append(setupTaskFunctions, []entities.TaskFunction{
		// 3. EncryptCardDetails
		// 		3a. RetrievePublicKey
		{
			Function:    task.RetrievePublicKey,
			StatusBegin: enums.EncryptingCardInfo,
			MaxRetries:  MAX_RETRIES,
		},
		//		3b. RetrievePrivateKey
		{
			Function:    task.RetrievePrivateKey,
			StatusBegin: "",
			MaxRetries:  MAX_RETRIES,
		},
		//		3c. RetrieveToken
		{
			Function:    task.RetrieveToken,
			StatusBegin: "",
			MaxRetries:  MAX_RETRIES,
		},
		//		3d. RetrieveJTI
		{
			Function:    task.RetrieveJTI,
			StatusBegin: "",
			MaxRetries:  MAX_RETRIES,
		},
	}...)
	return setupTaskFunctions
}

func (task *Task) GetMainFunctions() []entities.TaskFunction {
	mainTaskFunctions := []entities.TaskFunction{
		// 4. AddToCart
		{
			Function:         task.AddToCart,
			StatusBegin:      enums.WaitingForMonitor,
			MsBetweenRetries: int(u.MS_TO_WAIT),
		},
	}
	if task.PokemonCenterInput.TaskType == enums.TaskTypeGuest {
		// 5. SubmitEmailAddress
		mainTaskFunctions = append(mainTaskFunctions, entities.TaskFunction{
			Function:    task.SubmitEmailAddress,
			StatusBegin: enums.SettingEmailAddress,
			MaxRetries:  MAX_RETRIES,
		})
	}

	mainTaskFunctions = append(mainTaskFunctions, []entities.TaskFunction{
		// 7. SubmitAddressDetails
		{
			Function:    task.SubmitAddressDetails,
			StatusBegin: enums.SettingShippingInfo,
			MaxRetries:  MAX_RETRIES,
		},
		// 8. SubmitPaymentDetails
		{
			Function:    task.SubmitPaymentDetails,
			StatusBegin: enums.SettingBillingInfo,
			MaxRetries:  MAX_RETRIES,
		},
		// 9. Checkout
		{
			Function:         task.Checkout,
			StatusBegin:      enums.CheckingOut,
			MaxRetries:       MAX_RETRIES,
			CheckoutFunction: true,
		},
	}...)

	return mainTaskFunctions
}

func (task *Task) Login() (bool, string) {
	loginResponse := LoginResponse{}

	params := url.Values{}
	params.Add("username", task.PokemonCenterInput.Email)
	params.Add("password", task.PokemonCenterInput.Password)
	params.Add("grant_type", "password")
	params.Add("role", "REGISTERED")
	params.Add("scope", "pokemon")

	resp, _, err := util.MakeRequest(&util.Request{
		Client: task.BaseTask.Client,
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
		Client: task.BaseTask.Client,
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

func (task *Task) RetrievePublicKey() (bool, string) {
	paymentKeyResponse := PaymentKeyResponse{}

	resp, _, err := util.MakeRequest(&util.Request{
		Client: task.BaseTask.Client,
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

func (task *Task) RetrievePrivateKey() (bool, string) {
	var err error
	task.CyberSecureInfo.PublicToken, err = cybersource.RetrievePublicToken(task.CyberSecureInfo.PublicKey, task.BaseTask.Profile.CreditCard)
	if task.CyberSecureInfo.PublicToken == "" || err != nil {
		errorMessage := CyberSourceEncryptionError
		if err != nil {
			errorMessage = err.Error()
		}
		return false, fmt.Sprintf(enums.EncryptingCardInfoFailure, errorMessage)
	}
	return true, ""
}

func (task *Task) RetrieveToken() (bool, string) {
	resp, _, err := util.MakeRequest(&util.Request{
		Client: task.BaseTask.Client,
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

func (task *Task) RetrieveJTI() (bool, string) {
	var err error
	task.CyberSecureInfo.JtiToken, err = cybersource.RetrievePaymentToken(task.CyberSecureInfo.Privatekey)
	if task.CyberSecureInfo.JtiToken == "" || err != nil {
		errorMessage := RetrieveCyberSourcePaymentTokenError
		if err != nil {
			errorMessage = err.Error()
		}
		return false, fmt.Sprintf(enums.EncryptingCardInfoFailure, errorMessage)
	}

	return true, enums.EncryptingCardInfoSuccess
}

func (task *Task) AddToCart() (bool, string) {
	if _, ok := task.BaseTask.ProductInfo.SiteSpecificInfo["AddToCartForm"].(string); !ok {
		return false, fmt.Sprintf(enums.AddingToCartFailure, AddToCartBadInputError)
	}
	addToCartRequest := AddToCartRequest{
		ProductUri:    task.BaseTask.ProductInfo.SiteSpecificInfo["AddToCartForm"].(string),
		Quantity:      task.Input.Quantity,
		Configuration: "",
	}
	addToCartResponse := AddToCartResponse{}

	addToCartRequestBytes, err := json.Marshal(addToCartRequest)
	if err != nil {
		return false, fmt.Sprintf(enums.AddingToCartFailure, err.Error())
	}

	resp, _, err := util.MakeRequest(&util.Request{
		Client: task.BaseTask.Client,
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
			{"referer", fmt.Sprintf(AddToCartRefererEndpoint, task.BaseTask.ProductInfo.SKU)},
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
			if addToCartResponse.Quantity != task.Input.Quantity {
				return false, fmt.Sprintf(enums.AddingToCartFailure, fmt.Sprintf(AddToCartQuantityError, task.Input.Quantity, addToCartResponse.Quantity))
			} else {
				return true, enums.AddingToCartSuccess
			}
		}
	}

	return false, fmt.Sprintf(enums.AddingToCartFailure, UnknownError)
}

func (task *Task) SubmitEmailAddress() (bool, string) {
	emailRequest := EmailRequest{
		Email: task.BaseTask.Profile.Email,
	}

	emailBytes, err := json.Marshal(emailRequest)
	if err != nil {
		return false, fmt.Sprintf(enums.SettingEmailAddressFailure, err.Error())
	}
	resp, _, err := util.MakeRequest(&util.Request{
		Client: task.BaseTask.Client,
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
	if task.PokemonCenterInput.TaskType == enums.TaskTypeAccount {
		return true, ""
	}

	submitAddressRequest := SubmitAddressRequest{
		Billing: Address{
			FamilyName:      task.BaseTask.Profile.BillingAddress.LastName,
			GivenName:       task.BaseTask.Profile.BillingAddress.FirstName,
			StreetAddress:   task.BaseTask.Profile.BillingAddress.Address1,
			ExtendedAddress: task.BaseTask.Profile.BillingAddress.Address2,
			Locality:        task.BaseTask.Profile.BillingAddress.City,
			Region:          task.BaseTask.Profile.BillingAddress.StateCode,
			PostalCode:      task.BaseTask.Profile.BillingAddress.ZipCode,
			CountryName:     task.BaseTask.Profile.BillingAddress.CountryCode,
			PhoneNumber:     task.BaseTask.Profile.PhoneNumber,
		},
		Shipping: Address{
			FamilyName:      task.BaseTask.Profile.ShippingAddress.LastName,
			GivenName:       task.BaseTask.Profile.ShippingAddress.FirstName,
			StreetAddress:   task.BaseTask.Profile.ShippingAddress.Address1,
			ExtendedAddress: task.BaseTask.Profile.ShippingAddress.Address2,
			Locality:        task.BaseTask.Profile.ShippingAddress.City,
			Region:          task.BaseTask.Profile.ShippingAddress.StateCode,
			PostalCode:      task.BaseTask.Profile.ShippingAddress.ZipCode,
			CountryName:     task.BaseTask.Profile.ShippingAddress.CountryCode,
			PhoneNumber:     task.BaseTask.Profile.PhoneNumber,
		},
	}

	submitAddressRequestBytes, err := json.Marshal(submitAddressRequest)
	if err != nil {
		return false, fmt.Sprintf(enums.SettingShippingInfoFailure, err.Error())
	}
	resp, _, err := util.MakeRequest(&util.Request{
		Client: task.BaseTask.Client,
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
		PaymentDisplay: fmt.Sprintf("%s %s/%s", task.BaseTask.Profile.CreditCard.CardType, task.BaseTask.Profile.CreditCard.ExpMonth, task.BaseTask.Profile.CreditCard.ExpYear),
		PaymentKey:     task.CyberSecureInfo.PublicKey,
		PaymentToken:   task.CyberSecureInfo.JtiToken,
	}
	submitPaymentResponse := SubmitPaymentResponse{}

	paymentDetailsBytes, err := json.Marshal(submitPaymentRequest)
	if err != nil {
		return false, fmt.Sprintf(enums.SettingBillingInfoFailure, err.Error())
	}
	resp, _, err := util.MakeRequest(&util.Request{
		Client: task.BaseTask.Client,
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
		Client: task.BaseTask.Client,
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
			FamilyName:      task.BaseTask.Profile.BillingAddress.LastName,
			GivenName:       task.BaseTask.Profile.BillingAddress.FirstName,
			StreetAddress:   task.BaseTask.Profile.BillingAddress.Address1,
			ExtendedAddress: task.BaseTask.Profile.BillingAddress.Address2,
			Locality:        task.BaseTask.Profile.BillingAddress.City,
			Region:          task.BaseTask.Profile.BillingAddress.StateCode,
			PostalCode:      task.BaseTask.Profile.BillingAddress.ZipCode,
			CountryName:     task.BaseTask.Profile.BillingAddress.CountryCode,
			PhoneNumber:     task.BaseTask.Profile.PhoneNumber,
		},
		Shipping: Address{
			FamilyName:      task.BaseTask.Profile.ShippingAddress.LastName,
			GivenName:       task.BaseTask.Profile.ShippingAddress.FirstName,
			StreetAddress:   task.BaseTask.Profile.ShippingAddress.Address1,
			ExtendedAddress: task.BaseTask.Profile.ShippingAddress.Address2,
			Locality:        task.BaseTask.Profile.ShippingAddress.City,
			Region:          task.BaseTask.Profile.ShippingAddress.StateCode,
			PostalCode:      task.BaseTask.Profile.ShippingAddress.ZipCode,
			CountryName:     task.BaseTask.Profile.ShippingAddress.CountryCode,
			PhoneNumber:     task.BaseTask.Profile.PhoneNumber,
		},
	}

	//json marshal this for content length.
	submitAddressRequestBytes, err := json.Marshal(submitAddressRequest)
	if err != nil {
		log.Fatal("Marshal payload failed with error " + err.Error())
	}
	resp, _, err := util.MakeRequest(&util.Request{
		Client: task.BaseTask.Client,
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
