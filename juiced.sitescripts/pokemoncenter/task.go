package pokemoncenter

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/url"
	"time"

	"backend.juicedbot.io/juiced.infrastructure/common/entities"
	"backend.juicedbot.io/juiced.infrastructure/common/enums"
	"backend.juicedbot.io/juiced.infrastructure/common/events"
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
// Function order:
// 		1. WaitForMonitor
//		2. AddToCart
//		3. GetCartInfo
// 		4. SetPCID
//		5. SetShippingInfo
//		6. SetPaymentInfo
//		7. PlaceOrder
func (task *Task) RunTask() {
	// If the function panics due to a runtime error, recover from it
	defer func() {
		recover()
		task.Task.StopFlag = true
		task.PublishEvent(enums.TaskIdle, enums.TaskFail)
	}()

	task.PublishEvent(enums.WaitingForMonitor, enums.TaskStart)

	// 1. WaitForMonitor
	needToStop := task.WaitForMonitor()
	if needToStop {
		return
	}

	startTime := time.Now()

	//We will need to first generate an auth ID value,
	//We do this by logging in or by doing a 'get' on the cart and parsing the set-cookie header.
	UseAccountLogin := false //this needs to come from the front end user selection somewhere

	task.PublishEvent(enums.LoggingIn, enums.TaskUpdate)
	loggedIn := false
	needToStop = task.CheckForStop()
	if needToStop {
		return
	}
	if UseAccountLogin {
		loggedIn = task.Login()
	} else {
		loggedIn = task.GetAuthId()
	}
	if !loggedIn {
		time.Sleep(time.Duration(task.Task.Task.TaskDelay) * time.Millisecond)
	} else {
		//Set auth cookie, might do in task itself.
	}

	//now were logged in we can add item to cart
	// 2. AddToCart
	task.PublishEvent(enums.AddingToCart, enums.TaskUpdate)
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
	//If we are logged in we can skip address details else do this
	if !UseAccountLogin {
		// 3. GetCartInfo
		task.PublishEvent(enums.GettingCartInfo, enums.TaskUpdate)
		gotCartInfo := false
		for !gotCartInfo {
			needToStop := task.CheckForStop()
			if needToStop {
				return
			}
			gotCartInfo = task.SubmitAddressDetails()
			if !gotCartInfo {
				time.Sleep(time.Duration(task.Task.Task.TaskDelay) * time.Millisecond)
			}
		}
	}
	//Set the paymentKeyId for payment encryption
	task.PublishEvent(enums.GettingBillingInfo, enums.TaskUpdate)
	getPaymentKey := false
	for !getPaymentKey {
		needToStop := task.CheckForStop()
		if needToStop {
			return
		}
		getPaymentKey = task.GetPaymentKeyId()
		if !getPaymentKey {
			time.Sleep(time.Duration(task.Task.Task.TaskDelay) * time.Millisecond)
		}
	}

	//Now we have the public payment key, encrypt using CyberSecure encrpytion
	tempKey := "" //This will be the key ID returned from GetPaymentKeyId()
	UnlockedKey := CyberSourceV2(tempKey)
	//Now we post this key to cyber secure to retrieve the token

	task.PublishEvent(enums.SettingCartInfo, enums.TaskUpdate)
	retrieveKey := false
	for !retrieveKey {
		needToStop := task.CheckForStop()
		if needToStop {
			return
		}
		retrieveKey = task.RetrieveToken()
		if !retrieveKey {
			time.Sleep(time.Duration(task.Task.Task.TaskDelay) * time.Millisecond)
		}
	}
	tokenReturnedFromCyberSecure := "" // temp for imp
	//Now that we have the token we can retrieve the JTI token from this.
	jtiToken := retrievePaymentToken(tokenReturnedFromCyberSecure)

	//Now we have the public key and JTI Token we can submit the payment information

	// 4. SubmitPaymentInfo
	task.PublishEvent(enums.SettingCartInfo, enums.TaskUpdate)
	submitPayment := false
	for !submitPayment {
		needToStop := task.CheckForStop()
		if needToStop {
			return
		}
		submitPayment = task.SubmitPaymentDetails()
		if !submitPayment {
			time.Sleep(time.Duration(task.Task.Task.TaskDelay) * time.Millisecond)
		}
	}

	//now we can checkout
	// 5. checkout
	task.PublishEvent(enums.SettingShippingInfo, enums.TaskUpdate)
	checkout := false
	for !checkout {
		needToStop := task.CheckForStop()
		if needToStop {
			return
		}
		checkout = task.Checkout()
		if !checkout {
			time.Sleep(time.Duration(task.Task.Task.TaskDelay) * time.Millisecond)
		}
	}

	endTime := time.Now()

	log.Println("STARTED AT: " + startTime.String())
	log.Println("  ENDED AT: " + endTime.String())
	log.Println("TIME TO CHECK OUT: ", endTime.Sub(startTime).Milliseconds())

	task.PublishEvent(enums.CheckedOut, enums.TaskComplete)
}

// WaitForMonitor waits until the Monitor has sent the info to the task to continue
func (task *Task) WaitForMonitor() bool {
	for {
		needToStop := task.CheckForStop()
		if needToStop {
			return true
		}
	}
}

//Login and retrieve access code for auth cookie
func (task *Task) Login() bool {
	params := url.Values{}
	params.Add("username", "anthonyreeder123@gmail.com")
	params.Add("password", "pass")
	params.Add("grant_type", "password") //hardcode
	params.Add("role", "REGISTERED")     //hardcode
	params.Add("scope", "pokemon")       //hardcode

	resp, _, err := util.MakeRequest(&util.Request{
		Client: task.Task.Client,
		Method: "POST",
		URL:    fmt.Sprintf(AddToCartEndpoint),
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
		Data: []byte(params.Encode()),
	})
	if err != nil {
		fmt.Println(err.Error())
	}

	switch resp.StatusCode {
	case 200:
		return true
	}

	//If we reached this point we are out of stock.
	return false
}

//add product to cart passed from monitor via checkoutinfo
func (task *Task) AddToCart() bool {
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
				//If guest, remove cookies get new auth ID
				//If logged in, Empty cart or alert user
			} else {
				//instock
				return true
			}
		}
	}
	//If we reached this point we are out of stock.
	return false
}

//only needs doing on guest
func (task *Task) SubmitEmailAddress() bool {
	//This task isn't 'needed' pe se, IF there is an issue with the address formatting this will tell us
	//The submitAddressDetails will not and seemingly always gives the same response, even if it hasn't worked.
	email := Email{
		Email: "theemailaddress@email.com",
	}

	//json marshal this for content length.
	emailBytes, err := json.Marshal(email)
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
		},
		RequestBodyStruct: &email,
	})
	if err != nil {
		fmt.Println(err.Error())
	}

	switch resp.StatusCode {
	case 200:
		//read response for errors etc...
		return true
	}
	return false
}

//only needs doing on guest
func (task *Task) SubmitAddressDetailsValidate() bool {
	//This task isn't 'needed' pe se, IF there is an issue with the address formatting this will tell us
	//The submitAddressDetails will not and seemingly always gives the same response, even if it hasn't worked.
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
		},
		RequestBodyStruct: &submitAddressRequest,
	})
	if err != nil {
		fmt.Println(err.Error())
	}

	switch resp.StatusCode {
	case 200:
		//read response for errors etc...
		return true
	}
	return false
}

//only needs doing on guest
func (task *Task) SubmitAddressDetails() bool {
	//If user is logged in and has address details on teh account then we dont need to submit address details
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
		},
		RequestBodyStruct: &submitAddressRequest,
	})
	if err != nil {
		fmt.Println(err.Error())
	}

	switch resp.StatusCode {
	case 200:
		return true
	}
	return false
}

//GET request to retrieve paymentKey used for encryption
func (task *Task) GetPaymentKeyId() bool {
	paymentKeyResponse := PaymentKeyResponse{}

	resp, _, err := util.MakeRequest(&util.Request{
		Client: task.Task.Client,
		Method: "GET",
		URL:    fmt.Sprintf(PaymentKeyEndpoint),
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
		ResponseBodyStruct: &paymentKeyResponse,
	})
	if err != nil {
		fmt.Println(err.Error())
	}

	//need to do someting with payment key.

	switch resp.StatusCode {
	case 200:
		return true
	}
	return false
}

func (task *Task) GetAuthId() bool {
	paymentKeyResponse := PaymentKeyResponse{}

	resp, _, err := util.MakeRequest(&util.Request{
		Client: task.Task.Client,
		Method: "GET",
		URL:    fmt.Sprintf(PaymentKeyEndpoint),
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
		ResponseBodyStruct: &paymentKeyResponse,
	})
	if err != nil {
		fmt.Println(err.Error())
	}

	//need to do someting with payment key.

	switch resp.StatusCode {
	case 200:
		return true
	}
	return false
}

//Where we post encrypted payment details too
func (task *Task) SubmitPaymentDetails() bool {
	//populate using values passed from cyberSecure encryption
	paymentDetails := PaymentDetails{}

	//json marshal this for content length.
	paymentDetailsBytes, err := json.Marshal(paymentDetails)
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
		},
		RequestBodyStruct: &paymentDetails,
	})
	if err != nil {
		fmt.Println(err.Error())
	}

	switch resp.StatusCode {
	case 200:
		return true
	}
	return false
}

//Uses enlocked key to get the JTI token
func (task *Task) RetrieveToken() bool {
	temp := "Empty string" //get the encrypted key returned from cyberSecure and use here

	resp, _, err := util.MakeRequest(&util.Request{
		Client: task.Task.Client,
		Method: "POST",
		URL:    fmt.Sprintf(CyberSourceTokenEndpoint),
		RawHeaders: [][2]string{
			{"content-length", fmt.Sprint(bytes.NewReader([]byte(temp)).Size())},
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
		Data: []byte(temp),
	})
	if err != nil {
		fmt.Println(err.Error())
	}

	switch resp.StatusCode {
	case 200:
		return true
	}
	return false
}

//Checkout - self explanitory
func (task *Task) Checkout() bool {
	//If user is logged in and has address details on teh account then we dont need to submit address details

	//populate using values passed from submitPaymentDetails
	submitAddressRequest := CheckoutDetailsRequest{}

	//json marshal this for content length.
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
		},
		RequestBodyStruct: &submitAddressRequest,
	})
	if err != nil {
		fmt.Println(err.Error())
	}

	switch resp.StatusCode {
	case 200:
		return true
	}
	return false
}
