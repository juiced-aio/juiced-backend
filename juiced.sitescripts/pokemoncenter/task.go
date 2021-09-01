package pokemoncenter

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/url"
	"regexp"
	"strings"
	"sync"
	"time"

	"backend.juicedbot.io/juiced.client/http"
	"backend.juicedbot.io/juiced.infrastructure/common"
	"backend.juicedbot.io/juiced.infrastructure/common/entities"
	"backend.juicedbot.io/juiced.infrastructure/common/enums"
	"backend.juicedbot.io/juiced.infrastructure/common/events"
	"backend.juicedbot.io/juiced.sitescripts/base"
	"backend.juicedbot.io/juiced.sitescripts/util"
)

func CreatePokemonCenterTask(task *entities.Task, profile entities.Profile, proxyGroup *entities.ProxyGroup, eventBus *events.EventBus, email, password string, taskType enums.TaskType, addressType enums.AddressType) (Task, error) {
	pokemonCenterTask := Task{
		Task: base.Task{
			Task:       task,
			Profile:    profile,
			ProxyGroup: proxyGroup,
			EventBus:   eventBus,
		},
		AccountInfo: AccountInfo{
			Email:    email,
			Password: password,
		},
		TaskType:    taskType,
		AddressType: addressType,
	}
	if proxyGroup != nil {
		pokemonCenterTask.Task.Proxy = util.RandomLeastUsedProxy(proxyGroup.Proxies)
	} else {
		pokemonCenterTask.Task.Proxy = nil
	}
	return pokemonCenterTask, nil
}

func (task *Task) PublishEvent(status enums.TaskStatus, eventType enums.TaskEventType, statusPercentage int) {
	if status == enums.TaskIdle || !task.Task.StopFlag {
		task.Task.Task.SetTaskStatus(status)
		task.Task.EventBus.PublishTaskEvent(status, statusPercentage, eventType, nil, task.Task.Task.ID)
	}
}

func (task *Task) CheckForStop() bool {
	if task.Task.StopFlag && !task.Task.DontPublishEvents {
		task.PublishEvent(enums.TaskIdle, enums.TaskStop, 0)
		return true
	}
	return false
}

func (task *Task) RunTask() {
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
	task.StockData = PokemonCenterInStockData{}
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

	// 1. Login/LoginGuest
	if task.TaskType == enums.TaskTypeAccount {
		task.PublishEvent(enums.LoggingIn, enums.TaskUpdate, 10)
		if success, _ := task.RunUntilSuccessful(task.BecomeGuest, common.MAX_RETRIES); !success {
			return
		}
		if success, _ := task.RunUntilSuccessful(task.Login, common.MAX_RETRIES); !success {
			return
		}
	} else {
		task.PublishEvent(enums.SettingUp, enums.TaskUpdate, 10)
		if success, _ := task.RunUntilSuccessful(task.LoginGuest, common.MAX_RETRIES); !success {
			return
		}
	}

	// So clearing the cart is only needed for accounts of course but I noticed that it hits datadome everytime so
	// I thought it would be a good idea to force getting the cookie since it lasts throughout the task and will speed up
	// checkout times.
	if success, _ := task.RunUntilSuccessful(task.ClearCart, common.MAX_RETRIES); !success {
		return
	}

	// 2. RefreshLogin (in background)
	task.RefreshAt = time.Now().Unix() + 1800
	go task.RefreshLogin()

	// 3. Encrypt card details
	task.PublishEvent(enums.EncryptingCardInfo, enums.TaskUpdate, 15)
	if success, _ := task.RunUntilSuccessful(task.RetrieveEncryptedCardDetails, common.MAX_RETRIES); !success {
		return
	}

	// 4. WaitForMonitor
	task.PublishEvent(enums.WaitingForMonitor, enums.TaskUpdate, 20)
	needToStop := task.WaitForMonitor()
	if needToStop {
		return
	}

	task.Task.StartTime = time.Now()

	// 5. AddToCart
	task.PublishEvent(enums.AddingToCart, enums.TaskUpdate, 30)
	if success, _ := task.RunUntilSuccessful(task.AddToCart, -1); !success {
		return
	}

	// 6. Submit email address
	if task.TaskType == enums.TaskTypeGuest {
		task.PublishEvent(enums.SettingEmailAddress, enums.TaskUpdate, 60)
		if success, _ := task.RunUntilSuccessful(task.SubmitEmailAddress, common.MAX_RETRIES); !success {
			return
		}
	}

	// 7. Submit address details
	task.PublishEvent(enums.SettingShippingInfo, enums.TaskUpdate, 70)

	if task.AddressType == enums.AddressTypeNEW {
		if success, _ := task.RunUntilSuccessful(task.SubmitAddressDetails, common.MAX_RETRIES); !success {
			return
		}
	} else {
		// if success, _ := task.RunUntilSuccessful(task.SubmitAddressDetails, common.MAX_RETRIES); !success {
		// 	return
		// }
	}

	// 8. Submit payment details
	task.PublishEvent(enums.SettingBillingInfo, enums.TaskUpdate, 80)
	if success, _ := task.RunUntilSuccessful(task.SubmitPaymentDetails, common.MAX_RETRIES); !success {
		return
	}

	// 9. Checkout
	task.PublishEvent(enums.CheckingOut, enums.TaskUpdate, 90)
	success, status := task.RunUntilSuccessful(task.Checkout, common.MAX_RETRIES)

	task.Task.EndTime = time.Now()

	log.Println("STARTED AT: " + task.Task.StartTime.String())
	log.Println("  ENDED AT: " + task.Task.EndTime.String())
	log.Println("TIME TO CHECK OUT: ", task.Task.EndTime.Sub(task.Task.StartTime).Milliseconds())

	switch status {
	case enums.OrderStatusSuccess:
		task.PublishEvent(enums.CheckingOutSuccess, enums.TaskComplete, 100)
	case enums.OrderStatusDeclined:
		success = false
		task.PublishEvent(enums.CardDeclined, enums.TaskComplete, 100)
	case enums.OrderStatusFailed:
		task.PublishEvent(fmt.Sprintf(enums.CheckingOutFailure, "Unknown error"), enums.TaskComplete, 100)
	}

	quantity := task.Task.Task.TaskQty
	if task.StockData.MaxQuantity != 0 && quantity > task.StockData.MaxQuantity {
		quantity = task.StockData.MaxQuantity
	}

	log.Println(status)

	go util.ProcessCheckout(&util.ProcessCheckoutInfo{
		BaseTask:     task.Task,
		Success:      success,
		Status:       status,
		Embeds:       task.CreatePokemonCenterEmbed(status, task.StockData.ImageURL),
		Content:      "",
		ItemName:     task.StockData.ItemName,
		ImageURL:     task.StockData.ImageURL,
		Sku:          task.StockData.SKU,
		Retailer:     enums.PokemonCenter,
		Price:        task.StockData.Price,
		Quantity:     quantity,
		MsToCheckout: time.Since(task.Task.StartTime).Milliseconds(),
	})
}

func (task *Task) BecomeGuest() (bool, string) {
	resp, body, err := util.MakeRequest(&util.Request{
		Client: task.Task.Client,
		Method: "GET",
		URL:    BaseEndpoint,
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
	case 403:
		task.HandleDatadome(body, 10)
		return false, fmt.Sprintf(enums.SettingUpFailure, errors.New("hit datadome"))
	case 200:
	default:
		return false, fmt.Sprintf(enums.SettingUpFailure, errors.New("non 200 response"))
	}

	return true, enums.SettingUpSuccess
}

func (task *Task) Login() (bool, string) {
	loginResponse := LoginResponse{}

	params := url.Values{}
	params.Add("username", task.AccountInfo.Email)
	params.Add("password", task.AccountInfo.Password)
	params.Add("grant_type", "password")
	params.Add("role", "REGISTERED")
	params.Add("scope", "pokemon")

	// First request fails because no datadome cookie
	hasCookie := false
	for _, cookie := range task.Task.Client.Jar.Cookies(BaseURL) {
		if cookie.Name == "datadome" {
			hasCookie = true
		}
	}

	if !hasCookie {
		task.Task.Client.Jar.SetCookies(BaseURL, []*http.Cookie{{
			Name:     "datadome",
			Value:    common.RandString(6),
			Path:     "/",
			Domain:   ".pokemoncenter.com",
			Secure:   true,
			SameSite: http.SameSiteLaxMode,
		}})
	}

	resp, body, err := util.MakeRequest(&util.Request{
		Client: task.Task.Client,
		Method: "POST",
		URL:    LoginEndpoint,
		RawHeaders: [][2]string{
			{"content-length", fmt.Sprint(len(params.Encode()))},
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
		Data: []byte(params.Encode()),
	})
	if err != nil {
		return false, fmt.Sprintf(enums.LoginFailure, err.Error())
	}

	switch resp.StatusCode {
	case 200:
		err = json.Unmarshal([]byte(body), &loginResponse)
		if err != nil {
			return false, fmt.Sprintf(enums.LoginFailure, err.Error())
		}

		task.AccessToken = loginResponse.AccessToken
		return true, enums.LoginSuccess
	case 401:
		return false, fmt.Sprintf(enums.LoginFailure, err.Error())
	case 403:
		task.HandleDatadome(body, 10)
		return false, fmt.Sprintf(enums.LoginFailure, errors.New("hit datadome"))
	}

	return false, fmt.Sprintf(enums.LoginFailure, UnknownError)
}

func (task *Task) ClearCart() (bool, string) {
	var getCartInfoResponse GetCartInfoResponse
	resp, body, err := util.MakeRequest(&util.Request{
		Client: task.Task.Client,
		Method: "GET",
		URL:    CartEndpoint + "?format=nodatalinks",
		RawHeaders: [][2]string{
			{"sec-ch-ua", "\" Not A;Brand\";v=\"99\", \"Chromium\";v=\"90\", \"Google Chrome\";v=\"90\""},
			{"accept", "*/*"},
			{"x-requested-with", "XMLHttpRequest"},
			{"sec-ch-ua-mobile", "?0"},
			{"user-agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/90.0.4430.212 Safari/537.36"},
			{"x-store-scope", "pokemon"},
			{"origin", BaseEndpoint},
			{"sec-fetch-site", "same-origin"},
			{"sec-fetch-mode", "cors"},
			{"sec-fetch-dest", "empty"},
			{"referer", CartRefererEndpoint},
			{"accept-encoding", "gzip, deflate, br"},
			{"accept-language", "en-US,en;q=0.9"},
		},
	})
	if err != nil {
		return false, fmt.Sprintf(enums.SettingUpFailure, err.Error())
	}

	switch resp.StatusCode {
	case 200:
		err = json.Unmarshal([]byte(body), &getCartInfoResponse)
		if err != nil {
			return false, fmt.Sprintf(enums.SettingUpFailure, err.Error())
		}
	case 403:
		task.HandleDatadome(body, 10)
		return false, fmt.Sprintf(enums.SettingUpFailure, errors.New("hit datadome"))
	default:
		return false, fmt.Sprintf(enums.SettingUpFailure, errors.New("non 200 status "+resp.Status))
	}

	wg := sync.WaitGroup{}
	wg.Add(len(getCartInfoResponse.Lineitems))
	for _, lineItem := range getCartInfoResponse.Lineitems {
		go func(itemUrl string) {
			defer wg.Done()
			resp, _, err := util.MakeRequest(&util.Request{
				Client: task.Task.Client,
				Method: "POST",
				URL:    CartEndpoint + "?type=product&format=nodatalinks",
				RawHeaders: [][2]string{
					{"sec-ch-ua", "\" Not A;Brand\";v=\"99\", \"Chromium\";v=\"90\", \"Google Chrome\";v=\"90\""},
					{"accept", "*/*"},
					{"x-requested-with", "XMLHttpRequest"},
					{"sec-ch-ua-mobile", "?0"},
					{"user-agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/90.0.4430.212 Safari/537.36"},
					{"x-store-scope", "pokemon"},
					{"content-type", "application/json"},
					{"origin", BaseEndpoint},
					{"sec-fetch-site", "same-origin"},
					{"sec-fetch-mode", "cors"},
					{"sec-fetch-dest", "empty"},
					{"referer", CartRefererEndpoint},
					{"accept-encoding", "gzip, deflate, br"},
					{"accept-language", "en-US,en;q=0.9"},
				},

				Data: []byte(`{"productURI": "` + itemUrl + `","quantity": 0}`),
			})
			if err != nil || resp.StatusCode >= 200 && resp.StatusCode < 300 {
				return
			}
		}(lineItem.Element[0].Self.URI)
	}
	wg.Wait()

	return true, ""
}

func (task *Task) LoginGuest() (bool, string) {
	resp, body, err := util.MakeRequest(&util.Request{
		Client: task.Task.Client,
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

		accessToken := AccessToken{}
		json.Unmarshal([]byte(match[0]), &accessToken)
		task.AccessToken = accessToken.AccessToken
		return true, enums.SettingUpSuccess
	case 403:
		task.HandleDatadome(body, 10)
		return false, fmt.Sprintf(enums.SettingUpFailure, errors.New("hit datadome"))
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
			if task.TaskType == enums.TaskTypeAccount {
				if success, _ := task.RunUntilSuccessful(task.Login, common.MAX_RETRIES); !success {
					return
				}
			} else {
				if success, _ := task.RunUntilSuccessful(task.LoginGuest, common.MAX_RETRIES); !success {
					return
				}
			}
			task.RefreshAt = time.Now().Unix() + 1800
		}
		time.Sleep(common.MS_TO_WAIT)
	}
}

func (task *Task) RetrieveEncryptedCardDetails() (bool, string) {
	var err error

	// 1. Retrieve public key for encryption
	if success, _ := task.RunUntilSuccessful(task.RetrievePublicKey, common.MAX_RETRIES); !success {
		return false, fmt.Sprintf(enums.EncryptingCardInfoFailure, RetrieveCyberSourcePublicKeyError)
	}

	// 2. Encrypt using CyberSourceV2 encryption
	task.CyberSecureInfo.PublicToken, err = CyberSourceV2(task.CyberSecureInfo.PublicKey, task.Task.Profile.CreditCard)
	if task.CyberSecureInfo.PublicToken == "" || err != nil {
		errorMessage := CyberSourceEncryptionError
		if err != nil {
			errorMessage = err.Error()
		}
		return false, fmt.Sprintf(enums.EncryptingCardInfoFailure, errorMessage)
	}

	// 3. Retrieve CyberSourceV2 Token
	if success, _ := task.RunUntilSuccessful(task.RetrieveToken, common.MAX_RETRIES); !success {
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

	return true, ""
}

func (task *Task) RetrievePublicKey() (bool, string) {
	paymentKeyResponse := PaymentKeyResponse{}

	resp, _, err := util.MakeRequest(&util.Request{
		Client: task.Task.Client,
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
	resp, body, err := util.MakeRequest(&util.Request{
		Client: task.Task.Client,
		Method: "POST",
		URL:    CyberSourceTokenEndpoint,
		RawHeaders: [][2]string{
			{"content-length", fmt.Sprint(len(task.CyberSecureInfo.PublicToken))},
			{"sec-ch-ua", "\" Not A;Brand\";v=\"99\", \"Chromium\";v=\"90\", \"Google Chrome\";v=\"90\""},
			{"accept", "*/*"},
			{"x-requested-with", "XMLHttpRequest"},
			{"sec-ch-ua-mobile", "?0"},
			{"user-agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/90.0.4430.212 Safari/537.36"},
			{"content-type", "application/jwt; charset=UTF-8"},
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
	case 201:
		task.CyberSecureInfo.Privatekey = body
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
			task.Task.HasStockData = true
			return false
		}
		time.Sleep(common.MS_TO_WAIT)
	}
}

func (task *Task) AddToCart() (bool, string) {
	quantity := task.Task.Task.TaskQty
	if task.StockData.MaxQuantity != 0 && quantity > task.StockData.MaxQuantity {
		quantity = task.StockData.MaxQuantity
	}

	addToCartRequest := AddToCartRequest{
		ProductUri:    task.StockData.AddToCartForm,
		Quantity:      quantity,
		Configuration: struct{}{},
	}
	addToCartResponse := AddToCartResponse{}

	addToCartRequestBytes, err := json.Marshal(addToCartRequest)
	if err != nil {
		return false, fmt.Sprintf(enums.AddingToCartFailure, err.Error())
	}

	resp, body, err := util.MakeRequest(&util.Request{
		Client: task.Task.Client,
		Method: "POST",
		URL:    AddToCartEndpoint,
		RawHeaders: [][2]string{
			{"content-length", fmt.Sprint(len(addToCartRequestBytes))},
			{"sec-ch-ua", "\" Not A;Brand\";v=\"99\", \"Chromium\";v=\"90\", \"Google Chrome\";v=\"90\""},
			{"accept", "*/*"},
			{"x-requested-with", "XMLHttpRequest"},
			{"sec-ch-ua-mobile", "?0"},
			{"user-agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/90.0.4430.212 Safari/537.36"},
			{"content-type", "application/json; charset=UTF-8"},
			{"origin", BaseEndpoint},
			{"sec-fetch-site", "same-origin"},
			{"sec-fetch-mode", "cors"},
			{"sec-fetch-dest", "empty"},
			{"referer", fmt.Sprintf(AddToCartRefererEndpoint, task.StockData.SKU) + "/"},
			{"accept-encoding", "gzip, deflate, br"},
			{"accept-language", "en-US,en;q=0.9"},
		},
		Data: addToCartRequestBytes,
	})
	if err != nil {
		return false, fmt.Sprintf(enums.AddingToCartFailure, err.Error())
	}

	switch resp.StatusCode {
	case 200:
		err = json.Unmarshal([]byte(body), &addToCartResponse)
		if err != nil {
			return false, fmt.Sprintf(enums.LoginFailure, err.Error())
		}

		if addToCartResponse.Quantity != quantity {
			return false, fmt.Sprintf(enums.AddingToCartFailure, fmt.Sprintf(AddToCartQuantityError, quantity, addToCartResponse.Quantity))
		} else {
			return true, enums.AddingToCartSuccess
		}
	case 201:
		err = json.Unmarshal([]byte(body), &addToCartResponse)
		if err != nil {
			return false, fmt.Sprintf(enums.LoginFailure, err.Error())
		}

		if addToCartResponse.Quantity != quantity {
			return false, fmt.Sprintf(enums.AddingToCartFailure, fmt.Sprintf(AddToCartQuantityError, quantity, addToCartResponse.Quantity))
		} else {
			return true, enums.AddingToCartSuccess
		}

	case 403:
		task.HandleDatadome(body, 30)
		return false, fmt.Sprintf(enums.AddingToCartFailure, errors.New("hit datadome"))
	}

	return false, fmt.Sprintf(enums.AddingToCartFailure, UnknownError)
}

func (task *Task) SubmitEmailAddress() (bool, string) {
	emailRequest := EmailRequest{
		Email: task.Task.Profile.Email,
	}

	emailBytes, err := json.Marshal(emailRequest)
	if err != nil {
		return false, fmt.Sprintf(enums.SettingEmailAddressFailure, err.Error())
	}
	resp, _, err := util.MakeRequest(&util.Request{
		Client: task.Task.Client,
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
	submitAddressRequest := SubmitAddressRequest{
		Billing: Address{
			FamilyName:      task.Task.Profile.BillingAddress.LastName,
			GivenName:       task.Task.Profile.BillingAddress.FirstName,
			StreetAddress:   task.Task.Profile.BillingAddress.Address1,
			ExtendedAddress: task.Task.Profile.BillingAddress.Address2,
			Locality:        task.Task.Profile.BillingAddress.City,
			Region:          task.Task.Profile.BillingAddress.StateCode,
			PostalCode:      task.Task.Profile.BillingAddress.ZipCode,
			CountryName:     task.Task.Profile.BillingAddress.CountryCode,
			PhoneNumber:     task.Task.Profile.PhoneNumber,
		},
		Shipping: Address{
			FamilyName:      task.Task.Profile.ShippingAddress.LastName,
			GivenName:       task.Task.Profile.ShippingAddress.FirstName,
			StreetAddress:   task.Task.Profile.ShippingAddress.Address1,
			ExtendedAddress: task.Task.Profile.ShippingAddress.Address2,
			Locality:        task.Task.Profile.ShippingAddress.City,
			Region:          task.Task.Profile.ShippingAddress.StateCode,
			PostalCode:      task.Task.Profile.ShippingAddress.ZipCode,
			CountryName:     task.Task.Profile.ShippingAddress.CountryCode,
			PhoneNumber:     task.Task.Profile.PhoneNumber,
		},
	}

	submitAddressRequestBytes, err := json.Marshal(submitAddressRequest)
	if err != nil {
		return false, fmt.Sprintf(enums.SettingShippingInfoFailure, err.Error())
	}
	resp, _, err := util.MakeRequest(&util.Request{
		Client: task.Task.Client,
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
		PaymentDisplay: fmt.Sprintf("%s %s/%s", task.Task.Profile.CreditCard.CardType, task.Task.Profile.CreditCard.ExpMonth, task.Task.Profile.CreditCard.ExpYear),
		PaymentKey:     task.CyberSecureInfo.PublicKey,
		PaymentToken:   task.CyberSecureInfo.JtiToken,
	}
	submitPaymentResponse := SubmitPaymentResponse{}

	paymentDetailsBytes, err := json.Marshal(submitPaymentRequest)
	if err != nil {
		return false, fmt.Sprintf(enums.SettingBillingInfoFailure, err.Error())
	}
	resp, _, err := util.MakeRequest(&util.Request{
		Client: task.Task.Client,
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
		},
		RequestBodyStruct:  &submitPaymentRequest,
		ResponseBodyStruct: &submitPaymentResponse,
	})
	if err != nil {
		return false, fmt.Sprintf(enums.SettingBillingInfoFailure, err.Error())
	}

	switch resp.StatusCode {
	case 200:
		task.CheckoutInfo.CheckoutUri = submitPaymentResponse.Self.Uri
		return true, enums.SettingBillingInfoSuccess
	case 201:
		task.CheckoutInfo.CheckoutUri = submitPaymentResponse.Self.Uri
		return true, enums.SettingBillingInfoSuccess

	}

	return false, fmt.Sprintf(enums.SettingBillingInfoFailure, UnknownError)
}

func (task *Task) Checkout() (bool, string) {
	checkoutDetailsRequest := CheckoutDetailsRequest{PurchaseFrom: strings.Replace(task.CheckoutInfo.CheckoutUri, "paymentmethods", "purchases", 1) + "/form"}

	submitAddressRequestBytes, err := json.Marshal(checkoutDetailsRequest)
	if err != nil {
		return false, fmt.Sprintf(enums.CheckingOutFailure, err.Error())
	}
	var checkoutResponse CheckoutResponse
	resp, body, err := util.MakeRequest(&util.Request{
		Client: task.Task.Client,
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
		},
		RequestBodyStruct:  &checkoutDetailsRequest,
		ResponseBodyStruct: &checkoutResponse,
	})
	if err != nil {
		log.Println(err.Error())
		return false, fmt.Sprintf(enums.CheckingOutFailure, err.Error())
	}
	log.Println(string(body))

	switch resp.StatusCode {
	case 200:
		return true, enums.OrderStatusSuccess
	case 409:
		switch checkoutResponse.Messages[0].ID {
		case "payment.declined":
			return true, enums.OrderStatusDeclined
		default:
			return false, enums.OrderStatusFailed
		}

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
			FamilyName:      task.Task.Profile.BillingAddress.LastName,
			GivenName:       task.Task.Profile.BillingAddress.FirstName,
			StreetAddress:   task.Task.Profile.BillingAddress.Address1,
			ExtendedAddress: task.Task.Profile.BillingAddress.Address2,
			Locality:        task.Task.Profile.BillingAddress.City,
			Region:          task.Task.Profile.BillingAddress.StateCode,
			PostalCode:      task.Task.Profile.BillingAddress.ZipCode,
			CountryName:     "US",
			PhoneNumber:     task.Task.Profile.PhoneNumber,
		},
		Shipping: Address{
			FamilyName:      task.Task.Profile.ShippingAddress.LastName,
			GivenName:       task.Task.Profile.ShippingAddress.FirstName,
			StreetAddress:   task.Task.Profile.ShippingAddress.Address1,
			ExtendedAddress: task.Task.Profile.ShippingAddress.Address2,
			Locality:        task.Task.Profile.ShippingAddress.City,
			Region:          task.Task.Profile.ShippingAddress.StateCode,
			PostalCode:      task.Task.Profile.ShippingAddress.ZipCode,
			CountryName:     "US",
			PhoneNumber:     task.Task.Profile.PhoneNumber,
		},
	}

	//json marshal this for content length.
	submitAddressRequestBytes, err := json.Marshal(submitAddressRequest)
	if err != nil {
		log.Fatal("Marshal payload failed with error " + err.Error())
	}
	resp, _, err := util.MakeRequest(&util.Request{
		Client: task.Task.Client,
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
