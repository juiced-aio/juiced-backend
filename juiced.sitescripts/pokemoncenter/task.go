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

	"backend.juicedbot.io/juiced.infrastructure/common/entities"
	"backend.juicedbot.io/juiced.infrastructure/common/enums"
	"backend.juicedbot.io/juiced.infrastructure/common/events"
	"backend.juicedbot.io/juiced.infrastructure/queries"
	"backend.juicedbot.io/juiced.sitescripts/base"
	"backend.juicedbot.io/juiced.sitescripts/util"
)

// CreatePokemonCenterTask takes a Task entity and turns it into a PokemonCenter Task
func CreatePokemonCenterTask(task *entities.Task, profile entities.Profile, proxy entities.Proxy, eventBus *events.EventBus) (Task, error) {
	pokemonCenterTask := Task{}

	pokemonCenterTask = Task{
		Task: base.Task{
			Task:     task,
			Profile:  profile,
			Proxy:    proxy,
			EventBus: eventBus,
		},
	}
	return pokemonCenterTask, nil
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
func (task *Task) RunTask() {
	// If the function panics due to a runtime error, recover from it
	defer func() {
		recover()
		task.Task.StopFlag = true
		task.PublishEvent(enums.TaskIdle, enums.TaskFail)
	}()

	UseAccountLogin := false //this needs to come from the front end user selection somewhere
	task.RefreshAt = 0
	//set to 0 so we refresh now
	go task.RefreshLogin(UseAccountLogin)
	//Refresh login data using login or guest

	//Setup card encryption details for later
	for isSuccess, needtostop := task.RunUntilSuccessful(task.RetrieveEncryptedCardDetails()); !isSuccess || needtostop; {
		if needtostop {
			return
		}
	}

	// WaitForMonitor
	needToStop := task.WaitForMonitor()
	if needToStop {
		return
	}

	startTime := time.Now()

	// 1. AddToCart
	for isSuccess, needtostop := task.RunUntilSuccessful(task.AddToCart()); !isSuccess || needtostop; {
		if needtostop {
			return
		}
	}

	// 2. Submit email details
	for isSuccess, needtostop := task.RunUntilSuccessful(task.SubmitEmailAddress(UseAccountLogin)); !isSuccess || needtostop; {
		if needtostop {
			return
		}
	}

	// 3. Submit address details
	for isSuccess, needtostop := task.RunUntilSuccessful(task.SubmitAddressDetails(UseAccountLogin)); !isSuccess || needtostop; {
		if needtostop {
			return
		}
	}

	// 4. SubmitPaymentInfo
	for isSuccess, needtostop := task.RunUntilSuccessful(task.SubmitPaymentDetails()); !isSuccess || needtostop; {
		if needtostop {
			return
		}
	}

	// 5. Checkout
	for isSuccess, needtostop := task.RunUntilSuccessful(task.Checkout(startTime)); !isSuccess || needtostop; {
		if needtostop {
			return
		}
	}

	endTime := time.Now()

	log.Println("STARTED AT: " + startTime.String())
	log.Println("  ENDED AT: " + endTime.String())
	log.Println("TIME TO CHECK OUT: ", endTime.Sub(startTime).Milliseconds())

	task.PublishEvent(enums.CheckedOut, enums.TaskComplete)
}

func (task *Task) RetrieveEncryptedCardDetails() (bool, string) {
	task.PublishEvent(enums.EncryptingCardInfo, enums.TaskUpdate)
	card := Card{SecurityCode: task.Task.Profile.CreditCard.CVV, Number: task.Task.Profile.CreditCard.CardNumber, ExpMonth: task.Task.Profile.CreditCard.ExpMonth, ExpYear: task.Task.Profile.CreditCard.ExpYear}

	//check card details have been put in
	if card.SecurityCode == "" || card.Number == "" || card.ExpYear == "" || card.ExpMonth == "" {
		return false, enums.CardDetailsMissing
	}

	// Set public key for payment encryption
	for isSuccess, needtostop := task.RunUntilSuccessful(task.RetrievePublicKey()); !isSuccess || needtostop; {
		if needtostop {
			return false, enums.EncryptingCardInfoFailure
		}
	}

	//set card details ready to encrypt

	// Now we have the public payment key, encrypt using CyberSecure encrpytion
	task.CyberSecureInfo.PublicToken = CyberSourceV2(task.CyberSecureInfo.PublicKey, card)
	if task.CyberSecureInfo.PublicToken == "" {
		return false, enums.EncryptingCardInfoFailure
	}
	// Now we post this key to cyber secure to retrieve the token
	for isSuccess, needtostop := task.RunUntilSuccessful(task.RetrieveToken()); !isSuccess || needtostop; {
		if needtostop {
			return false, enums.SettingBillingInfoFailure
		}
	}

	// Now that we have the token we can retrieve the JTI token from this.
	task.CyberSecureInfo.JtiToken = retrievePaymentToken(task.CyberSecureInfo.Privatekey)
	if task.CyberSecureInfo.JtiToken == "" {
		return false, enums.EncryptingCardInfoFailure
	}

	//If we reach this point everything should be working as intended return true.
	return true, enums.EncryptingCardInfoSuccess
}

// WaitForMonitor waits until the Monitor has sent the info to the task to continue
func (task *Task) WaitForMonitor() bool {
	task.PublishEvent(enums.WaitingForMonitor, enums.TaskStart)
	for {
		needToStop := task.CheckForStop()
		if needToStop {
			return true
		}
		if task.CheckoutInfo.AddToCartForm != "" {
			return false
		}
		time.Sleep(1 * time.Millisecond)
	}
}

// Login and retrieve access code for auth cookie
func (task *Task) Login() (bool, string) {
	loginResponse := LoginResponse{}

	params := url.Values{}
	params.Add("username", "anthonyreeder123@gmail.com") //needs to come from front end
	params.Add("password", "pass")                       //needs to come from front end
	params.Add("grant_type", "password")                 //hardcode
	params.Add("role", "REGISTERED")                     //hardcode
	params.Add("scope", "pokemon")                       //hardcode

	resp, _, err := util.MakeRequest(&util.Request{
		Client: task.Task.Client,
		Method: "POST",
		URL:    fmt.Sprintf(LoginEndpoint),
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
			{"referer", fmt.Sprintf(AddToCartRefererEndpoint, task.CheckoutInfo.SKU)},
			{"accept-encoding", "gzip, deflate, br"},
			{"accept-language", "en-US,en;q=0.9"},
		},
		ResponseBodyStruct: loginResponse,
		Data:               []byte(params.Encode()),
	})
	if err != nil {
		fmt.Println(err.Error())
	}

	switch resp.StatusCode {
	case 200:
		task.AccessToken = loginResponse.Access_token
		return true, enums.LoginSuccess
	case 401:
		//Wrong user/password
		task.PublishEvent(enums.LoginFailed, enums.TaskFail)
	}
	return false, enums.LoginFailed
}

func (task *Task) RefreshLogin(useAccountLogin bool) {
	task.PublishEvent(enums.LoggingIn, enums.TaskUpdate)

	defer func() {
		if recover() != nil {
			task.RefreshLogin(useAccountLogin)
		}
	}()

	for {
		if task.RefreshAt == 0 || time.Now().Unix() > task.RefreshAt {
			if useAccountLogin {
				for isSuccess, needtostop := task.RunUntilSuccessful(task.Login()); !isSuccess || needtostop; {
					if needtostop {
						return
					}
				}
			} else {
				for isSuccess, needtostop := task.RunUntilSuccessful(task.LoginGuest()); !isSuccess || needtostop; {
					if needtostop {
						return
					}
				}
			}
			task.RefreshAt = time.Now().Unix() + 1800
		}
	}
}

// Add product to cart passed from monitor via checkoutinfo
func (task *Task) AddToCart() (bool, string) {
	task.PublishEvent(enums.AddingToCart, enums.TaskUpdate)

	//Setup request using data passed from 'Instock' data to the tasks 'Checkout data' (Done in monitor-store)
	addToCartRequest := AddToCartRequest{ProductUri: task.CheckoutInfo.AddToCartForm, Quantity: 1, Configuration: ""}
	//Empty Response for the response
	addToCartResponse := AddToCartResponse{}

	//json marshal this for content length.
	addToCartRequestBytes, err := json.Marshal(addToCartRequest)
	if err != nil {
		log.Fatal("Marshal payload failed with error " + err.Error())
	}

	//setup request
	resp, _, err := util.MakeRequest(&util.Request{
		Client: task.Task.Client,
		Method: "POST",
		URL:    fmt.Sprintf(AddToCartEndpoint),
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
			{"referer", fmt.Sprintf(AddToCartRefererEndpoint, task.CheckoutInfo.SKU)},
			{"accept-encoding", "gzip, deflate, br"},
			{"accept-language", "en-US,en;q=0.9"},
			{"Cookie", "auth={\"access_token\":\"" + task.AccessToken + "\",\"token_type\":\"bearer\",\"expires_in\":604799,\"scope\":\"pokemon\",\"role\":\"PUBLIC\",\"roles\":[\"PUBLIC\"]}"},
		},
		ResponseBodyStruct: &addToCartResponse,
		RequestBodyStruct:  &addToCartRequest,
	})
	if err != nil {
		fmt.Println(err.Error())
	}

	switch resp.StatusCode {
	case 200:
		if addToCartResponse.Type == "carts.line-item" {
			//we must check quantity as if logged in it could have previously stored items.
			if addToCartResponse.Quantity != 1 {
				return false, fmt.Sprint(enums.AddingToCartQuantityFailure, addToCartResponse.Quantity)
			} else {
				//instock
				return true, enums.AddingToCartSuccess
			}
		}
	}
	//If we reached this point we are out of stock or another issue.
	return false, enums.AddingToCartFailure
}

// Submit email address
func (task *Task) SubmitEmailAddress(useAccountLogin bool) (bool, string) {
	if useAccountLogin {
		return true, ""
	}
	task.PublishEvent(enums.SettingEmailAddress, enums.TaskUpdate)

	emailRequest := EmailRequest{
		Email: task.Task.Profile.Email,
	}

	emailBytes, err := json.Marshal(emailRequest)
	if err != nil {
		log.Fatal("Marshal payload failed with error " + err.Error())
	}
	resp, _, err := util.MakeRequest(&util.Request{
		Client: task.Task.Client,
		Method: "POST",
		URL:    fmt.Sprintf(SubmitEmailEndpoint),
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
			{"referer", fmt.Sprintf(SubmitAddresValidateRefererEndpoint)}, //double check this endpoint
			{"accept-encoding", "gzip, deflate, br"},
			{"accept-language", "en-US,en;q=0.9"},
			{"Cookie", "auth={\"access_token\":\"" + task.AccessToken + "\",\"token_type\":\"bearer\",\"expires_in\":604799,\"scope\":\"pokemon\",\"role\":\"PUBLIC\",\"roles\":[\"PUBLIC\"]}"},
		},
		RequestBodyStruct: &emailRequest,
	})
	if err != nil {
		fmt.Println(err.Error())
	}

	switch resp.StatusCode {
	case 200:
		//read response for errors etc...
		return true, enums.SettingEmailAddressSuccess
	}

	return false, enums.SettingEmailAddressFailure
}

// Validate address details (not needed but might be useful if there are problems later)
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
		URL:    fmt.Sprintf(SubmitAddressValidateEndpoint),
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
			{"referer", fmt.Sprintf(SubmitAddresValidateRefererEndpoint)}, //double check this endpoint
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

// Submit address details
func (task *Task) SubmitAddressDetails(useAccountLogin bool) (bool, string) {
	if useAccountLogin {
		return true, ""
	}

	task.PublishEvent(enums.SettingShippingInfo, enums.TaskUpdate)

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

	submitAddressRequestBytes, err := json.Marshal(submitAddressRequest)
	if err != nil {
		log.Fatal("Marshal payload failed with error " + err.Error())
	}
	resp, _, err := util.MakeRequest(&util.Request{
		Client: task.Task.Client,
		Method: "POST",
		URL:    fmt.Sprintf(SubmitAddressEndpoint),
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
			{"referer", fmt.Sprintf(SubmitAddresRefererEndpoint)}, //double check this endpoint
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
		return true, enums.SettingShippingInfoSuccess
	}

	return false, enums.SettingShippingInfoFailure
}

// Submit payment details
func (task *Task) SubmitPaymentDetails() (bool, string) {
	task.PublishEvent(enums.SettingBillingInfo, enums.TaskUpdate)

	//Payment display example: "Visa 02/2026"
	submitPaymentRequest := SubmitPaymentRequest{PaymentDisplay: task.Task.Profile.CreditCard.CardType + task.Task.Profile.CreditCard.ExpMonth + "/" + task.Task.Profile.CreditCard.ExpYear, PaymentKey: task.CyberSecureInfo.PublicKey, PaymentToken: task.CyberSecureInfo.JtiToken}
	submitPaymentResponse := SubmitPaymentResponse{}

	paymentDetailsBytes, err := json.Marshal(submitPaymentRequest)
	if err != nil {
		log.Fatal("Marshal payload failed with error " + err.Error())
	}
	resp, _, err := util.MakeRequest(&util.Request{
		Client: task.Task.Client,
		Method: "POST",
		URL:    fmt.Sprintf(SubmitPaymentDetailsEndpoint),
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
			{"referer", fmt.Sprintf(SubmitAddresRefererEndpoint)}, //double check this endpoint
			{"accept-encoding", "gzip, deflate, br"},
			{"accept-language", "en-US,en;q=0.9"},
			{"Cookie", "auth={\"access_token\":\"" + task.AccessToken + "\",\"token_type\":\"bearer\",\"expires_in\":604799,\"scope\":\"pokemon\",\"role\":\"PUBLIC\",\"roles\":[\"PUBLIC\"]}"},
		},
		RequestBodyStruct:  &submitPaymentRequest,
		ResponseBodyStruct: &submitPaymentResponse,
	})
	if err != nil {
		fmt.Println(err.Error())
	}

	switch resp.StatusCode {
	case 200:
		task.CheckoutInfo.CheckoutUri = submitPaymentResponse.Self.Uri
		return true, enums.SettingBillingInfoSuccess
	}

	return false, enums.SettingBillingInfoFailure
}

// Checkout - self explanitory
func (task *Task) Checkout(startTime time.Time) (bool, string) {
	task.PublishEvent(enums.CheckingOut, enums.TaskUpdate)

	checkoutDetailsRequest := CheckoutDetailsRequest{PurchaseFrom: strings.Replace(task.CheckoutInfo.CheckoutUri, "paymentmethods", "purchases", 1) + "/form"}

	submitAddressRequestBytes, err := json.Marshal(checkoutDetailsRequest)
	if err != nil {
		log.Fatal("Marshal payload failed with error " + err.Error())
	}
	resp, _, err := util.MakeRequest(&util.Request{
		Client: task.Task.Client,
		Method: "POST",
		URL:    fmt.Sprintf(CheckoutEndpoint),
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
			{"referer", fmt.Sprintf(SubmitAddresRefererEndpoint)}, //double check this endpoint
			{"accept-encoding", "gzip, deflate, br"},
			{"accept-language", "en-US,en;q=0.9"},
			{"Cookie", "auth={\"access_token\":\"" + task.AccessToken + "\",\"token_type\":\"bearer\",\"expires_in\":604799,\"scope\":\"pokemon\",\"role\":\"PUBLIC\",\"roles\":[\"PUBLIC\"]}"},
		},
		RequestBodyStruct: &checkoutDetailsRequest,
	})
	if err != nil {
		fmt.Println(err.Error())
	}

	switch resp.StatusCode {
	case 200:

		_, user, err := queries.GetUserInfo()
		if err != nil {
			return false, enums.GetUserFailure
		}

		util.ProcessCheckout(util.ProcessCheckoutInfo{
			BaseTask:     task.Task,
			Success:      true,
			Content:      "",
			Embeds:       task.CreatePokemonCenterEmbed(enums.OrderStatusSuccess, "https://media.discordapp.net/attachments/849430464036077598/855979506204278804/Icon_1.png?width=457&height=467"),
			UserInfo:     user,
			ItemName:     task.CheckoutInfo.ItemName,
			Sku:          task.CheckoutInfo.SKU,
			Retailer:     enums.PokemonCenter,
			Price:        task.CheckoutInfo.Price,
			Quantity:     1,
			MsToCheckout: time.Since(startTime).Milliseconds(),
		})
		return true, enums.CheckedOut
	}

	return false, enums.CheckoutFailed
}

// Retrieve public key
func (task *Task) RetrievePublicKey() (bool, string) {
	paymentKeyResponse := PaymentKeyResponse{}

	resp, _, err := util.MakeRequest(&util.Request{
		Client: task.Task.Client,
		Method: "GET",
		URL:    fmt.Sprintf(PublicPaymentKeyEndpoint),
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
			{"referer", fmt.Sprintf(SubmitAddresRefererEndpoint)}, //double check this endpoint
			{"accept-encoding", "gzip, deflate, br"},
			{"accept-language", "en-US,en;q=0.9"},
			{"Cookie", "auth={\"access_token\":\"" + task.AccessToken + "\",\"token_type\":\"bearer\",\"expires_in\":604799,\"scope\":\"pokemon\",\"role\":\"PUBLIC\",\"roles\":[\"PUBLIC\"]}"},
		},
		ResponseBodyStruct: &paymentKeyResponse,
	})
	if err != nil {
		fmt.Println(err.Error())
	}

	switch resp.StatusCode {
	case 200:
		task.CyberSecureInfo.PublicKey = paymentKeyResponse.KeyId
		return true, ""
	}

	return false, ""
}

// When using guest account retrieves the auth ID generated when you go on cart
func (task *Task) LoginGuest() (bool, string) {
	resp, _, err := util.MakeRequest(&util.Request{
		Client: task.Task.Client,
		Method: "GET",
		URL:    fmt.Sprintf(AuthKeyEndpoint),
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
			{"referer", fmt.Sprintf(SubmitAddresRefererEndpoint)}, //double check this endpoint
			{"accept-encoding", "gzip, deflate, br"},
			{"accept-language", "en-US,en;q=0.9"},
		},
	})
	if err != nil {
		fmt.Println(err.Error())
	}

	switch resp.StatusCode {
	case 200:
		rawHeader := resp.Header.Get("Set-Cookie")
		re := regexp.MustCompile("({)(.*?)(})")
		match := re.FindStringSubmatch(rawHeader)
		fmt.Println(match[0])

		accessToken := AccessToken{}
		json.Unmarshal([]byte(match[0]), &accessToken)
		task.AccessToken = accessToken.Access_token
		return true, enums.LoginSuccess

		//add captcha support here
	}
	//if we reach here then login has failed, we can read the response if we want specifics.
	return false, enums.LoginFailed
}

// Uses encrypted public key to get the JTI Token
func (task *Task) RetrieveToken() (bool, string) {
	resp, _, err := util.MakeRequest(&util.Request{
		Client: task.Task.Client,
		Method: "POST",
		URL:    fmt.Sprintf(CyberSourceTokenEndpoint),
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
			{"referer", fmt.Sprintf(SubmitAddresRefererEndpoint)}, //double check this endpoint
			{"accept-encoding", "gzip, deflate, br"},
			{"accept-language", "en-US,en;q=0.9"},
		},
		Data: []byte(task.CyberSecureInfo.PublicToken),
	})
	if err != nil {
		fmt.Println(err.Error())
	}

	switch resp.StatusCode {
	case 200:
		body, _ := ioutil.ReadAll(resp.Body)
		task.CyberSecureInfo.Privatekey = string(body)
		return true, ""
	}

	return false, ""
}
