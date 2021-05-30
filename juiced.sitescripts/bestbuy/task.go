package bestbuy

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"strings"
	"time"

	"backend.juicedbot.io/m/v2/juiced.infrastructure/common/entities"
	"backend.juicedbot.io/m/v2/juiced.infrastructure/common/enums"
	"backend.juicedbot.io/m/v2/juiced.infrastructure/common/events"
	"backend.juicedbot.io/m/v2/juiced.sitescripts/base"
	"backend.juicedbot.io/m/v2/juiced.sitescripts/util"
	"github.com/anaskhan96/soup"
	"github.com/google/uuid"
)

// CreateBestbuyTask takes a Task entity and turns it into a Bestbuy Task
func CreateBestbuyTask(task *entities.Task, profile entities.Profile, proxy entities.Proxy, eventBus *events.EventBus, taskType enums.TaskType, email, password string) (Task, error) {
	bestbuyTask := Task{}
	client, err := util.CreateClient(proxy)
	if err != nil {
		return bestbuyTask, err
	}
	bestbuyTask = Task{
		Task: base.Task{
			Task:     task,
			Profile:  profile,
			Proxy:    proxy,
			EventBus: eventBus,
			Client:   client,
		},
		AccountInfo: AccountInfo{
			Email:    email,
			Password: password,
		},
		TaskType: taskType,
	}
	return bestbuyTask, err
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
// 		1. Login / Become a guest
// 		2. WaitForMonitor
// 		3. AddtoCart / Handle a queue
// 		4. Checkout
//		5. SetShippingInfo
// 		6. SetPaymentInfo
// 		7. PlaceOrder
func (task *Task) RunTask() {
	task.PublishEvent(enums.LoggingIn, enums.TaskStart)
	// 1. Login / Become a guest
	sessionMade := false
	for !sessionMade {
		needToStop := task.CheckForStop()
		if needToStop {
			return
		}
		switch task.TaskType {
		case enums.TaskTypeAccount:
			sessionMade = task.Login()
		case enums.TaskTypeGuest:
			sessionMade = BecomeGuest(&task.Task.Client)
		}

		if !sessionMade {
			time.Sleep(time.Duration(task.Task.Task.TaskDelay) * time.Millisecond)
		}
	}

	task.PublishEvent(enums.WaitingForMonitor, enums.TaskUpdate)
	// 2. WaitForMonitor
	needToStop := task.WaitForMonitor()
	if needToStop {
		return
	}

	startTime := time.Now()

	task.PublishEvent(enums.AddingToCart, enums.TaskUpdate)
	// 3. AddtoCart / Handle a queue
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
	// 4. Checkout
	gotCartInfo := false
	for !gotCartInfo {
		needToStop := task.CheckForStop()
		if needToStop {
			return
		}
		gotCartInfo = task.Checkout()
		if !gotCartInfo {
			time.Sleep(time.Duration(task.Task.Task.TaskDelay) * time.Millisecond)
		}
	}

	task.PublishEvent(enums.SettingShippingInfo, enums.TaskUpdate)
	// 5. SetShippingInfo

	setShippingInfo := false
	for !setShippingInfo {
		needToStop := task.CheckForStop()
		if needToStop {
			return
		}
		setShippingInfo = task.SetShippingInfo()
		if !setShippingInfo {
			time.Sleep(time.Duration(task.Task.Task.TaskDelay) * time.Millisecond)
		}
	}

	task.PublishEvent(enums.SettingBillingInfo, enums.TaskUpdate)
	// 6. SetPaymentInfo
	setPaymentInfo := false
	for !setPaymentInfo {
		needToStop := task.CheckForStop()
		if needToStop {
			return
		}
		setPaymentInfo = task.SetPaymentInfo()
		if !setPaymentInfo {
			time.Sleep(time.Duration(task.Task.Task.TaskDelay) * time.Millisecond)
		}
	}

	task.PublishEvent(enums.CheckingOut, enums.TaskUpdate)
	// 7. PlaceOrder
	placedOrder := false
	for !placedOrder {
		needToStop := task.CheckForStop()
		if needToStop {
			return
		}
		placedOrder = task.PlaceOrder()
		if !placedOrder {
			time.Sleep(time.Duration(task.Task.Task.TaskDelay) * time.Millisecond)
		}
	}

	endTime := time.Now()

	log.Println("STARTED AT: " + startTime.String())
	log.Println("  ENDED AT: " + endTime.String())
	log.Println("TIME TO CHECK OUT: " + endTime.Sub(startTime).String())

	task.PublishEvent(enums.CheckedOut, enums.TaskComplete)
}

// Function to generate valid abck cookies using an API
func (task *Task) NewAbck(abckClient *http.Client, location string) bool {
	resp, err := util.MakeRequest(&util.Request{
		Client: *abckClient,
		Method: "GET",
		URL:    AkamaiEndpoint,
		RawHeaders: [][2]string{
			{"sec-ch-ua", "\" Not A;Brand\";v=\"99\", \"Chromium\";v=\"90\", \"Google Chrome\";v=\"90\""},
			{"sec-ch-ua-mobile", "?0"},
			{"user-agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/90.0.4430.212 Safari/537.36"},
			{"content-type", "text/plain;charset=UTF-8"},
			{"accept", "*/*"},
			{"origin", BaseEndpoint},
			{"sec-fetch-site", "same-origin"},
			{"sec-fetch-mode", "cors"},
			{"sec-fetch-dest", "empty"},
			{"referer", location},
			{"accept-encoding", "gzip, deflate, br"},
			{"accept-language", "en-US,en;q=0.9"},
		},
	})
	if err != nil {
		fmt.Println(err)
	}
	defer resp.Body.Close()

	var abckCookie string
	var genResponse GenResponse
	for _, cookie := range abckClient.Jar.Cookies(ParsedBase) {
		if cookie.Name == "_abck" {
			abckCookie = cookie.Value
		}
	}
	form := url.Values{
		"authKey":        {"UFBfkndVOYYGZ58Uu8Mv49TrODCEiVE0gKpOAj679Z0dJoQThi9wGpoq6tpIRvrR"},
		"pageurl":        {location},
		"skipkact":       {"true"},
		"skipmact":       {"true"},
		"onblur":         {"false"},
		"onfocus":        {"false"},
		"abck":           {abckCookie},
		"sensordatalink": {AkamaiEndpoint},
		"ver":            {"1.7"},
		"firstpost":      {"true"},
		"pixelid":        {""},
		"pixelg":         {""},
		"json":           {"true"},
	}
	client, _ := util.CreateClient()
	resp, err = util.MakeRequest(&util.Request{
		Client: client,
		Method: "POST",
		URL:    GenEndpoint,
		RawHeaders: [][2]string{
			{"Content-Type", "application/x-www-form-urlencoded"},
			{"User-Agent", "Juiced/1.0"},
			{"Accept", "*/*"},
			{"Accept-Encoding", "gzip, deflate, br"},
			{"Connection", "keep-alive"},
			{"Content-Length", fmt.Sprint(len(form.Encode()))},
		},
		Data:               []byte(form.Encode()),
		ResponseBodyStruct: &genResponse,
	})
	if err != nil {
		fmt.Println(err)
	}
	defer resp.Body.Close()

	sensorRequest := SensorRequest{
		SensorData: genResponse.Sensordata,
	}

	data, _ := json.Marshal(sensorRequest)
	resp, err = util.MakeRequest(&util.Request{
		Client: *abckClient,
		Method: "POST",
		URL:    AkamaiEndpoint,
		RawHeaders: [][2]string{
			{"content-length", fmt.Sprint(len(data))},
			{"sec-ch-ua", "\" Not A;Brand\";v=\"99\", \"Chromium\";v=\"90\", \"Google Chrome\";v=\"90\""},
			{"sec-ch-ua-mobile", "?0"},
			{"user-agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/90.0.4430.212 Safari/537.36"},
			{"content-type", "text/plain;charset=UTF-8"},
			{"accept", "*/*"},
			{"origin", BaseEndpoint},
			{"sec-fetch-site", "same-origin"},
			{"sec-fetch-mode", "cors"},
			{"sec-fetch-dest", "empty"},
			{"referer", location},
			{"accept-encoding", "gzip, deflate, br"},
			{"accept-language", "en-US,en;q=0.9"},
		},
		Data: data,
	})
	if err != nil {
		fmt.Println(err)
	}
	defer resp.Body.Close()

	for _, cookie := range abckClient.Jar.Cookies(ParsedBase) {
		if cookie.Name == "_abck" {
			abckCookie = cookie.Value
		}
	}
	form = url.Values{
		"authKey":        {"UFBfkndVOYYGZ58Uu8Mv49TrODCEiVE0gKpOAj679Z0dJoQThi9wGpoq6tpIRvrR"},
		"pageurl":        {location},
		"skipkact":       {"true"},
		"skipmact":       {"false"},
		"onblur":         {"false"},
		"onfocus":        {"false"},
		"abck":           {abckCookie},
		"sensordatalink": {AkamaiEndpoint},
		"ver":            {"1.7"},
		"firstpost":      {"false"},
		"pixelid":        {""},
		"pixelg":         {""},
		"json":           {"true"},
	}

	resp, err = util.MakeRequest(&util.Request{
		Client: client,
		Method: "POST",
		URL:    GenEndpoint,
		RawHeaders: [][2]string{
			{"Content-Type", "application/x-www-form-urlencoded"},
			{"User-Agent", "Juiced/1.0"},
			{"Accept", "*/*"},
			{"Accept-Encoding", "gzip, deflate, br"},
			{"Connection", "keep-alive"},
			{"Content-Length", fmt.Sprint(len(form.Encode()))},
		},
		Data:               []byte(form.Encode()),
		ResponseBodyStruct: &genResponse,
	})
	if err != nil {
		fmt.Println(err)
	}
	defer resp.Body.Close()

	sensorRequest = SensorRequest{
		SensorData: genResponse.Sensordata,
	}
	data, _ = json.Marshal(sensorRequest)

	resp, err = util.MakeRequest(&util.Request{
		Client: *abckClient,
		Method: "POST",
		URL:    AkamaiEndpoint,
		RawHeaders: [][2]string{
			{"content-length", fmt.Sprint(len(data))},
			{"sec-ch-ua", "\" Not A;Brand\";v=\"99\", \"Chromium\";v=\"90\", \"Google Chrome\";v=\"90\""},
			{"sec-ch-ua-mobile", "?0"},
			{"user-agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/90.0.4430.212 Safari/537.36"},
			{"content-type", "text/plain;charset=UTF-8"},
			{"accept", "*/*"},
			{"origin", BaseEndpoint},
			{"sec-fetch-site", "same-origin"},
			{"sec-fetch-mode", "cors"},
			{"sec-fetch-dest", "empty"},
			{"referer", location},
			{"accept-encoding", "gzip, deflate, br"},
			{"accept-language", "en-US,en;q=0.9"},
		},
		Data: data,
	})
	if err != nil {
		fmt.Println(err)
	}
	defer resp.Body.Close()

	for _, cookie := range abckClient.Jar.Cookies(ParsedBase) {
		if cookie.Name == "_abck" {
			abckCookie = cookie.Value
		}
	}
	form = url.Values{
		"authKey":        {"UFBfkndVOYYGZ58Uu8Mv49TrODCEiVE0gKpOAj679Z0dJoQThi9wGpoq6tpIRvrR"},
		"pageurl":        {location},
		"skipkact":       {"true"},
		"skipmact":       {"false"},
		"onblur":         {"false"},
		"onfocus":        {"false"},
		"abck":           {abckCookie},
		"sensordatalink": {AkamaiEndpoint},
		"ver":            {"1.7"},
		"firstpost":      {"false"},
		"pixelid":        {""},
		"pixelg":         {""},
		"json":           {"true"},
	}

	resp, err = util.MakeRequest(&util.Request{
		Client: client,
		Method: "POST",
		URL:    GenEndpoint,
		RawHeaders: [][2]string{
			{"Content-Type", "application/x-www-form-urlencoded"},
			{"User-Agent", "Juiced/1.0"},
			{"Accept", "*/*"},
			{"Accept-Encoding", "gzip, deflate, br"},
			{"Connection", "keep-alive"},
			{"Content-Length", fmt.Sprint(len(form.Encode()))},
		},
		Data:               []byte(form.Encode()),
		ResponseBodyStruct: &genResponse,
	})
	if err != nil {
		fmt.Println(err)
	}
	defer resp.Body.Close()

	sensorRequest = SensorRequest{
		SensorData: genResponse.Sensordata,
	}
	data, _ = json.Marshal(sensorRequest)

	resp, err = util.MakeRequest(&util.Request{
		Client: *abckClient,
		Method: "POST",
		URL:    AkamaiEndpoint,
		RawHeaders: [][2]string{
			{"content-length", fmt.Sprint(len(data))},
			{"sec-ch-ua", "\" Not A;Brand\";v=\"99\", \"Chromium\";v=\"90\", \"Google Chrome\";v=\"90\""},
			{"sec-ch-ua-mobile", "?0"},
			{"user-agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/90.0.4430.212 Safari/537.36"},
			{"content-type", "text/plain;charset=UTF-8"},
			{"accept", "*/*"},
			{"origin", BaseEndpoint},
			{"sec-fetch-site", "same-origin"},
			{"sec-fetch-mode", "cors"},
			{"sec-fetch-dest", "empty"},
			{"referer", location},
			{"accept-encoding", "gzip, deflate, br"},
			{"accept-language", "en-US,en;q=0.9"},
		},
		Data: data,
	})
	if err != nil {
		fmt.Println(err)
	}
	defer resp.Body.Close()

	switch resp.StatusCode {
	case 201:
		for _, cookie := range abckClient.Jar.Cookies(ParsedBase) {
			if cookie.Name == "_abck" {
				fmt.Println(cookie.Value)
				validator, _ := util.FindInString(cookie.Value, "~", "~")
				if validator == "-1" {
					task.NewAbck(abckClient, location)
				}

			}
		}
		return true
	}
	return false
}

// Login logs the task's client into the account specified
func (task *Task) Login() bool {

	resp, err := util.MakeRequest(&util.Request{
		Client:     task.Task.Client,
		Method:     "GET",
		URL:        BaseEndpoint,
		RawHeaders: DefaultRawHeaders,
	})
	if err != nil || resp.StatusCode != 200 {
		fmt.Println(err)
		return false
	}
	defer resp.Body.Close()

	task.NewAbck(&task.Task.Client, BaseEndpoint+"/")

	resp, err = util.MakeRequest(&util.Request{
		Client: task.Task.Client,
		Method: "GET",
		URL:    LoginPageEndpoint,
		RawHeaders: [][2]string{
			{"pragma", "no-cache"},
			{"cache-control", "no-cache"},
			{"upgrade-insecure-requests", "1"},
			{"user-agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/90.0.4430.212 Safari/537.36"},
			{"accept", "text/html,application/xhtml+xml,application/xml;q=0.9,image/avif,image/webp,image/apng,*/*;q=0.8,application/signed-exchange;v=b3;q=0.9"},
			{"sec-fetch-site", "same-origin"},
			{"sec-fetch-mode", "navigate"},
			{"sec-fetch-user", "?1"},
			{"sec-fetch-dest", "document"},
			{"sec-ch-ua", "\" Not A;Brand\";v=\"99\", \"Chromium\";v=\"90\", \"Google Chrome\";v=\"90\""},
			{"sec-ch-ua-mobile", "?0"},
			{"referer", BaseEndpoint + "/"},
			{"accept-encoding", "gzip, deflate, br"},
			{"accept-language", "en-US,en;q=0.9"},
		},
	})
	if err != nil || resp.StatusCode != 200 {
		fmt.Println(err)
		return false
	}
	defer resp.Body.Close()

	// Getting all the json data
	body := util.ReadBody(resp)
	doc := soup.HTMLParse(body)
	signinData := doc.Find("script", "id", "signon-data").FullText()

	correctData := make(map[string]string)
	var signinJson *SigninJson

	json.Unmarshal([]byte(signinData), &signinJson)

	for i := range signinJson.Alpha {
		var tempReverse string
		for _, norm := range signinJson.Alpha[i] {
			tempReverse = string(norm) + tempReverse

		}
		tempDecode, _ := base64.StdEncoding.DecodeString(tempReverse)
		if strings.Contains(string(tempDecode), "_A_") {
			correctData["alpha"] = signinJson.Alpha[i]
		}

	}

	for i := range signinJson.Codelist {
		tempDecode, _ := base64.StdEncoding.DecodeString(signinJson.Codelist[i])
		if strings.Contains(string(tempDecode), "_X_") {
			correctData["pass"] = signinJson.Codelist[i]
		}
	}

	var userAgent = `{"userAgent":"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/90.0.4430.93 Safari/537.36"}`
	var email = task.AccountInfo.Email
	var password = task.AccountInfo.Password
	timestamp := time.Now().Format("2006-01-02T15:04:05.000Z")
	tempActivity := UserActivity{
		Mousemoved:         true,
		Keyboardused:       true,
		Fieldreceivedinput: true,
		Fieldreceivedfocus: true,
		Timestamp:          timestamp,
		Email:              email,
	}
	tempData, err := json.Marshal(tempActivity)
	ok := util.HandleErrors(err, util.RequestMarshalBodyError)
	if !ok {
		return false
	}

	encryptedActivity := encrypt(tempData, activityKey)
	encryptedInfo := encrypt([]byte(userAgent), activityKey)
	encryptedEmail := encrypt([]byte(email), emailKey)

	data := bytes.NewBuffer([]byte(`
	{"token":"` + signinJson.Token + `","activity":"1:user-activity-2016-09:` + encryptedActivity + `","loginMethod":"UID_PASSWORD","flowOptions":"000000010000000","alpha":"` + correctData["alpha"] + `","Salmon":"FA7F2","encryptedEmail":"1:email-2017-01:` + encryptedEmail + `","` + correctData["pass"] + `":"` + password + `","info":"1:user-activity-2016-09:` + encryptedInfo + `","` + signinJson.Emailfieldname + `":"` + email + `"}
	`))

	task.Task.Client.Jar.SetCookies(ParsedBase, []*http.Cookie{
		{Name: "ZPLANK", Value: "0e0a383f97f24e5ab11fef6269000a93"},
	})

	task.NewAbck(&task.Task.Client, LoginPageEndpoint+"/")
	var loginResponse LoginResponse
	resp, err = util.MakeRequest(&util.Request{
		Client: task.Task.Client,
		Method: "POST",
		URL:    LoginEndpoint,
		RawHeaders: [][2]string{
			{"content-length", fmt.Sprint(data.Len())},
			{"pragma", "no-cache"},
			{"cache-control", "no-cache"},
			{"sec-ch-ua", "\" Not A;Brand\";v=\"99\", \"Chromium\";v=\"90\", \"Google Chrome\";v=\"90\""},
			{"accept", "application/json"},
			{"sec-ch-ua-mobile", "?0"},
			{"user-agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/90.0.4430.212 Safari/537.36"},
			{"content-type", "application/json"},
			{"origin", BaseEndpoint},
			{"sec-fetch-site", "same-origin"},
			{"sec-fetch-mode", "cors"},
			{"sec-fetch-dest", "empty"},
			{"referer", "https://www.bestbuy.com/identity/signin?token=" + signinJson.Token},
			{"accept-encoding", "gzip, deflate, br"},
			{"accept-language", "en-US,en;q=0.9"},
		},
		Data:               data.Bytes(),
		ResponseBodyStruct: &loginResponse,
	})
	if err != nil || resp.StatusCode != 200 {
		fmt.Println(err)
		return false
	}
	defer resp.Body.Close()
	fmt.Println(loginResponse)
	resp, err = util.MakeRequest(&util.Request{
		Client: task.Task.Client,
		Method: "GET",
		URL:    BaseEndpoint,
		RawHeaders: [][2]string{
			{"sec-ch-ua", "\" Not A;Brand\";v=\"99\", \"Chromium\";v=\"90\", \"Google Chrome\";v=\"90\""},
			{"sec-ch-ua-mobile", "?0"},
			{"upgrade-insecure-requests", "1"},
			{"user-agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/90.0.4430.212 Safari/537.36"},
			{"accept", "text/html,application/xhtml+xml,application/xml;q=0.9,image/avif,image/webp,image/apng,*/*;q=0.8,application/signed-exchange;v=b3;q=0.9"},
			{"sec-fetch-site", "none"},
			{"sec-fetch-mode", "navigate"},
			{"sec-fetch-user", "?1"},
			{"sec-fetch-dest", "document"},
			{"referer", "https://www.bestbuy.com/identity/signin?token=" + signinJson.Token},
			{"accept-encoding", "gzip, deflate, br"},
			{"accept-language", "en-US,en;q=0.9"},
		},
	})
	if err != nil || resp.StatusCode != 200 {
		fmt.Println(err)
		return false
	}
	defer resp.Body.Close()

	return true
}

// WaitForMonitor waits until the Monitor has sent the info to the task to continue
func (task *Task) WaitForMonitor() bool {
	for {
		needToStop := task.CheckForStop()
		if needToStop {
			return true
		}
		if task.CheckoutInfo.SKUInStock != "" {
			return false
		}
	}
}

// AddToCart adds the item to cart and also handles a queue if there is one
func (task *Task) AddToCart() bool {
	var (
		a2TransactionCode string
		a2TransactionID   string
	)

	addToCartRequest := AddToCartRequest{
		Items: []Items{
			{Skuid: task.CheckoutInfo.SKUInStock},
		},
	}
	data, _ := json.Marshal(addToCartRequest)
	addToCartResponse := AddToCartResponse{}
	var handled bool
	for !handled {
		for _, cookie := range task.Task.Client.Jar.Cookies(ParsedBase) {
			if cookie.Name == "_abck" {
				validator, _ := util.FindInString(cookie.Value, "~", "~")
				if validator == "-1" {
					task.NewAbck(&task.Task.Client, fmt.Sprintf("https://www.bestbuy.com/site/%v.p?skuId=%v", task.CheckoutInfo.SKUInStock, task.CheckoutInfo.SKUInStock))
				}
			}
		}

		resp, err := util.MakeRequest(&util.Request{
			Client: task.Task.Client,
			Method: "POST",
			URL:    AddToCartEndpoint,
			RawHeaders: [][2]string{
				{"content-length", fmt.Sprint(len(data))},
				{"sec-ch-ua", "\" Not A;Brand\";v=\"99\", \"Chromium\";v=\"90\", \"Google Chrome\";v=\"90\""},
				{"accept", "application/json"},
				{"sec-ch-ua-mobile", "?0"},
				{"user-agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/90.0.4430.212 Safari/537.36"},
				{"content-type", "application/json; charset=UTF-8"},
				{"origin", BaseEndpoint},
				{"sec-fetch-site", "same-origin"},
				{"sec-fetch-mode", "cors"},
				{"sec-fetch-dest", "empty"},
				{"referer", fmt.Sprintf("https://www.bestbuy.com/site/%v.p?skuId=%v", task.CheckoutInfo.SKUInStock, task.CheckoutInfo.SKUInStock)},
				{"accept-encoding", "gzip, deflate, br"},
				{"accept-language", "en-US,en;q=0.9"},
			},
			Data:               data,
			ResponseBodyStruct: &addToCartResponse,
		})
		ok := util.HandleErrors(err, util.RequestDoError)
		if !ok {
			return false
		}

		defer resp.Body.Close()

		switch resp.StatusCode {
		case 200:
			handled = true
		case 400:
			a2TransactionCode = resp.Header.Get("a2ctransactioncode")
			a2TransactionID = resp.Header.Get("a2ctransactionreferenceid")
			times, err := CheckTime(a2TransactionCode)
			if err != nil {
				fmt.Println(err)
			}

			fmt.Println(times)
			if times < 5 {
				for _, cookie := range task.Task.Client.Jar.Cookies(ParsedBase) {
					if cookie.Name == "_abck" {
						validator, _ := util.FindInString(cookie.Value, "~", "~")
						if validator == "-1" {
							task.NewAbck(&task.Task.Client, fmt.Sprintf("https://www.bestbuy.com/site/%v.p?skuId=%v", task.CheckoutInfo.SKUInStock, task.CheckoutInfo.SKUInStock))
						}
					}
				}
				fmt.Println("Joining Queue")
				time.Sleep(time.Duration(times*60000) * time.Millisecond)
				fmt.Println("Out of Queue")
				addToCartResponse = AddToCartResponse{}
				resp, err := util.MakeRequest(&util.Request{
					Client: task.Task.Client,
					Method: "POST",
					URL:    AddToCartEndpoint,
					RawHeaders: [][2]string{
						{"content-length", fmt.Sprint(len(data))},
						{"sec-ch-ua", "\" Not A;Brand\";v=\"99\", \"Chromium\";v=\"90\", \"Google Chrome\";v=\"90\""},
						{"accept", "application/json"},
						{"a2ctransactioncode", a2TransactionCode},
						{"a2ctransactionreferenceid", a2TransactionID},
						{"sec-ch-ua-mobile", "?0"},
						{"user-agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/90.0.4430.212 Safari/537.36"},
						{"content-type", "application/json; charset=UTF-8"},
						{"origin", BaseEndpoint},
						{"sec-fetch-site", "same-origin"},
						{"sec-fetch-mode", "cors"},
						{"sec-fetch-dest", "empty"},
						{"referer", fmt.Sprintf("https://www.bestbuy.com/site/%v.p?skuId=%v", task.CheckoutInfo.SKUInStock, task.CheckoutInfo.SKUInStock)},
						{"accept-encoding", "gzip, deflate, br"},
						{"accept-language", "en-US,en;q=0.9"},
					},
					Data:               data,
					ResponseBodyStruct: &addToCartResponse,
				})
				ok := util.HandleErrors(err, util.RequestDoError)
				if !ok {
					return false
				}

				defer resp.Body.Close()

				switch resp.StatusCode {
				case 200:
					handled = true
				case 400:
					switch addToCartResponse.Errorsummary.Errorcode {
					case "ITEM_MAX_QUANTITY_EXCEEDED":
						handled = true
					case "CONSTRAINED_ITEM":
						fmt.Println("Requeued")
					}
				}
			} else {
				//	As a guest you do not ever get blocked adding to cart, but while logged in you will get blocked
				time.Sleep(3 * time.Second)
			}
		}

	}
	return true
}

// Checkout goes to the checkout page and gets the required information for the rest of the checkout process
func (task *Task) Checkout() bool {
	resp, err := util.MakeRequest(&util.Request{
		Client:     task.Task.Client,
		Method:     "GET",
		URL:        CheckoutEndpoint,
		RawHeaders: DefaultRawHeaders,
	})
	ok := util.HandleErrors(err, util.RequestDoError)
	if !ok {
		return false
	}

	defer resp.Body.Close()

	switch resp.StatusCode {
	case 200:
		body := util.ReadBody(resp)

		rawOrderData, err := util.FindInString(body, `var orderData = `, `;`)
		if err != nil {
			return false
		}

		orderData := OrderData{}

		json.Unmarshal([]byte(rawOrderData), &orderData)
		fmt.Println(orderData.Items)
		task.CheckoutInfo.ID = orderData.ID
		task.CheckoutInfo.ItemID = orderData.Items[0].ID
		task.CheckoutInfo.PaymentID = orderData.Payment.ID
		task.CheckoutInfo.OrderID = orderData.Customerorderid
		task.CheckoutInfo.ImageUrl = orderData.Items[0].Meta.Imageurl + ";canvasHeight=500;canvasWidth=500"
		task.CheckoutInfo.ItemName = orderData.Items[0].Meta.Shortlabel
		return true

	default:
		return false
	}
}

// SetShippingInfo sets the shipping info in checkout
func (task *Task) SetShippingInfo() bool {
	setShippingRequest := SetShippingRequest{
		Phonenumber:     task.Task.Profile.PhoneNumber,
		Smsnotifynumber: "",
		Smsoptin:        false,
		Emailaddress:    task.Task.Profile.Email,
		Items: []ShipItems{
			{
				ID:   task.CheckoutInfo.ItemID,
				Type: "DEFAULT",
				Selectedfulfillment: Selectedfulfillment{
					Shipping: Shipping{
						Address: ShipAddress{
							Country:             task.Task.Profile.ShippingAddress.CountryCode,
							Savetoprofile:       false,
							Street2:             strings.ToUpper(task.Task.Profile.ShippingAddress.Address2),
							Useaddressasbilling: task.Task.Profile.SSAB,
							Middleinitial:       "",
							Lastname:            task.Task.Profile.ShippingAddress.LastName,
							Street:              strings.ToUpper(task.Task.Profile.ShippingAddress.Address1),
							City:                strings.ToUpper(task.Task.Profile.ShippingAddress.City),
							Override:            false,
							Zipcode:             task.Task.Profile.ShippingAddress.ZipCode,
							State:               task.Task.Profile.ShippingAddress.StateCode,
							Firstname:           task.Task.Profile.ShippingAddress.FirstName,
							Iswishlistaddress:   false,
							Dayphonenumber:      task.Task.Profile.PhoneNumber,
							Type:                "RESIDENTIAL",
						},
					},
				},
				Giftmessageselected: false,
			},
		},
	}
	data, err := json.Marshal(setShippingRequest)
	ok := util.HandleErrors(err, util.RequestMarshalBodyError)
	if !ok {
		return false
	}

	setShippingResponse := SetShippingResponse{}

	resp, err := util.MakeRequest(&util.Request{
		Client: task.Task.Client,
		Method: "PATCH",
		URL:    fmt.Sprintf(OrderEndpoint, task.CheckoutInfo.ID) + "/",
		RawHeaders: [][2]string{
			{"content-length", fmt.Sprint(len(data))},
			{"pragma", "no-cache"},
			{"cache-control", "no-cache"},
			{"sec-ch-ua", "\" Not A;Brand\";v=\"99\", \"Chromium\";v=\"90\", \"Google Chrome\";v=\"90\""},
			{"accept", "application/com.bestbuy.order+json"},
			{"x-user-interface", "DotCom-Optimized"},
			{"sec-ch-ua-mobile", "?0"},
			{"user-agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/90.0.4430.212 Safari/537.36"},
			{"content-type", "application/json"},
			{"origin", BaseEndpoint},
			{"sec-fetch-site", "same-origin"},
			{"sec-fetch-mode", "cors"},
			{"sec-fetch-dest", "empty"},
			{"referer", BaseShippingEndpoint},
			{"accept-encoding", "gzip, deflate, br"},
			{"accept-language", "en-US,en;q=0.9"},
		},
		Data:               data,
		ResponseBodyStruct: &setShippingResponse,
	})
	ok = util.HandleErrors(err, util.RequestDoError)
	if !ok {
		return false
	}

	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return false
	}

	resp, err = util.MakeRequest(&util.Request{
		Client: task.Task.Client,
		Method: "POST",
		URL:    fmt.Sprintf(OrderEndpoint, task.CheckoutInfo.ID) + "/validate",
		RawHeaders: [][2]string{
			{"content-length", "0"},
			{"pragma", "no-cache"},
			{"cache-control", "no-cache"},
			{"sec-ch-ua", "\" Not A;Brand\";v=\"99\", \"Chromium\";v=\"90\", \"Google Chrome\";v=\"90\""},
			{"accept", "application/com.bestbuy.order+json"},
			{"x-user-interface", "DotCom-Optimized"},
			{"sec-ch-ua-mobile", "?0"},
			{"user-agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/90.0.4430.212 Safari/537.36"},
			{"content-type", "application/json"},
			{"origin", BaseEndpoint},
			{"sec-fetch-site", "same-origin"},
			{"sec-fetch-mode", "cors"},
			{"sec-fetch-dest", "empty"},
			{"referer", BaseShippingEndpoint},
			{"accept-encoding", "gzip, deflate, br"},
			{"accept-language", "en-US,en;q=0.9"},
		},
	})
	ok = util.HandleErrors(err, util.RequestDoError)
	if !ok {
		return false
	}

	switch resp.StatusCode {
	case 200:
		return true
	default:
		return false
	}
}

// SetPaymentInfo sets the payment info in checkout
func (task *Task) SetPaymentInfo() bool {
	task.NewAbck(&task.Task.Client, BasePaymentEndpoint)

	var billing Billingaddress
	if task.Task.Profile.SSAB {
		billing = Billingaddress{
			Country:             task.Task.Profile.ShippingAddress.CountryCode,
			Useaddressasbilling: true,
			Middleinitial:       "",
			Lastname:            task.Task.Profile.ShippingAddress.LastName,
			Iswishlistaddress:   false,
			City:                strings.ToUpper(task.Task.Profile.ShippingAddress.City),
			State:               task.Task.Profile.ShippingAddress.StateCode,
			Firstname:           task.Task.Profile.ShippingAddress.FirstName,
			Addressline1:        strings.ToUpper(task.Task.Profile.ShippingAddress.Address1),
			Addressline2:        strings.ToUpper(task.Task.Profile.ShippingAddress.Address2),
			Dayphone:            task.Task.Profile.PhoneNumber,
			Postalcode:          task.Task.Profile.ShippingAddress.ZipCode,
			Standardized:        false,
			Useroverridden:      false,
		}
	} else {
		billing = Billingaddress{
			Country:             task.Task.Profile.BillingAddress.CountryCode,
			Useaddressasbilling: true,
			Middleinitial:       "",
			Lastname:            task.Task.Profile.BillingAddress.LastName,
			Iswishlistaddress:   false,
			City:                strings.ToUpper(task.Task.Profile.BillingAddress.City),
			State:               task.Task.Profile.BillingAddress.StateCode,
			Firstname:           task.Task.Profile.BillingAddress.FirstName,
			Addressline1:        strings.ToUpper(task.Task.Profile.BillingAddress.Address1),
			Addressline2:        strings.ToUpper(task.Task.Profile.BillingAddress.Address2),
			Dayphone:            task.Task.Profile.PhoneNumber,
			Postalcode:          task.Task.Profile.BillingAddress.ZipCode,
			Standardized:        false,
			Useroverridden:      false,
		}
	}
	encryptedNumber := encrypt([]byte("00960001"+task.Task.Profile.CreditCard.CardNumber), paymentKey)
	data, _ := json.Marshal(SetPaymentRequest{
		Billingaddress: billing,
		Creditcard: Creditcard{
			Hascid:          false,
			Invalidcard:     false,
			Iscustomercard:  false,
			Isnewcard:       true,
			Isvisacheckout:  false,
			Govpurchasecard: false,
			Number:          encryptedNumber + ":3:735818052:" + task.Task.Profile.CreditCard.CardNumber,
			// This number is specific to the card type from here: https://www.bestbuy.com/checkout/global/binRange - But it just gives a range
			// Having it static either is fine or will only give declines, so I'll test with a working card soon and hopefully cancel it in time
			Binnumber:       "476771",
			Ispwpregistered: false,
			Expmonth:        task.Task.Profile.CreditCard.ExpMonth,
			Expyear:         task.Task.Profile.CreditCard.ExpYear,
			Cvv:             task.Task.Profile.CreditCard.CVV,
			Orderid:         task.CheckoutInfo.OrderID,
			Savetoprofile:   false,
			Type:            task.Task.Profile.CreditCard.CardType,
			International:   false,
			Virtualcard:     false,
		},
	})
	resp, err := util.MakeRequest(&util.Request{
		Client: task.Task.Client,
		Method: "PUT",
		URL:    fmt.Sprintf(PaymentEndpoint, task.CheckoutInfo.PaymentID),
		RawHeaders: [][2]string{
			{"content-length", fmt.Sprint(len(data))},
			{"pragma", "no-cache"},
			{"cache-control", "no-cache"},
			{"sec-ch-ua", "\" Not A;Brand\";v=\"99\", \"Chromium\";v=\"90\", \"Google Chrome\";v=\"90\""},
			{"accept", "application/json, text/javascript, */*; q=0.01"},
			{"x-client", "CHECKOUT"},
			{"x-context-id", task.CheckoutInfo.OrderID},
			{"sec-ch-ua-mobile", "?0"},
			{"user-agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/90.0.4430.212 Safari/537.36"},
			{"content-type", "application/json"},
			{"origin", BaseEndpoint},
			{"sec-fetch-site", "same-origin"},
			{"sec-fetch-mode", "cors"},
			{"sec-fetch-dest", "empty"},
			{"referer", BasePaymentEndpoint},
			{"accept-encoding", "gzip, deflate, br"},
			{"accept-language", "en-US,en;q=0.9"},
		},
		Data: data,
	})
	if err != nil {
		fmt.Println(err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return false
	}

	for _, cookie := range task.Task.Client.Jar.Cookies(ParsedBase) {
		if cookie.Name == "_abck" {
			validator, _ := util.FindInString(cookie.Value, "~", "~")
			if validator == "-1" {
				task.NewAbck(&task.Task.Client, BasePaymentEndpoint)
			}
		}
	}

	resp, err = util.MakeRequest(&util.Request{
		Client: task.Task.Client,
		Method: "POST",
		URL:    fmt.Sprintf(RefreshPaymentEndpoint, task.CheckoutInfo.ID),
		RawHeaders: [][2]string{
			{"content-length", "2"},
			{"pragma", "no-cache"},
			{"cache-control", "no-cache"},
			{"sec-ch-ua", "\" Not;A Brand\";v=\"99\", \"Google Chrome\";v=\"91\", \"Chromium\";v=\"91\""},
			{"accept", "application/json, text/javascript, */*; q=0.01"},
			{"x-user-interface", "DotCom-Optimized"},
			{"x-native-checkout-version", "__VERSION__"},
			{"sec-ch-ua-mobile", "?0"},
			{"user-agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/91.0.4472.77 Safari/537.36"},
			{"content-type", "application/json"},
			{"origin", BaseEndpoint},
			{"sec-fetch-site", "same-origin"},
			{"sec-fetch-mode", "cors"},
			{"sec-fetch-dest", "empty"},
			{"referer", BasePaymentEndpoint},
			{"accept-encoding", "gzip, deflate, br"},
			{"accept-language", "en-US,en;q=0.9"},
		},
		Data: []byte("{}"),
	})
	if err != nil {
		fmt.Println(err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return false
	}

	data, _ = json.Marshal(PrelookupRequest{
		Orderid: task.CheckoutInfo.OrderID,
		Browserinfo: Browserinfo{
			Javaenabled: false,
			Language:    "en-US",
			Useragent:   "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/91.0.4472.77 Safari/537.36",
			Height:      "1920",
			Width:       "1080",
			Timezone:    "420",
			Colordepth:  "24",
		},
	})

	var vt string
	for _, cookie := range task.Task.Client.Jar.Cookies(ParsedBase) {
		if cookie.Name == "vt" {
			vt = cookie.Value
		}
	}
	prelookupResonse := PrelookupResponse{}
	resp, err = util.MakeRequest(&util.Request{
		Client: task.Task.Client,
		Method: "POST",
		URL:    fmt.Sprintf(PrelookupEndpoint, task.CheckoutInfo.PaymentID),
		RawHeaders: [][2]string{
			{"content-length", fmt.Sprint(len(data))},
			{"pragma", "no-cache"},
			{"cache-control", "no-cache"},
			{"sec-ch-ua", "\" Not;A Brand\";v=\"99\", \"Google Chrome\";v=\"91\", \"Chromium\";v=\"91\""},
			{"vt", vt},
			{"x-context-id", task.CheckoutInfo.OrderID},
			{"sec-ch-ua-mobile", "?0"},
			{"user-agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/91.0.4472.77 Safari/537.36"},
			{"content-type", "application/json"},
			{"ut", "undefined"},
			{"x-client", "CHECKOUT"},
			{"x-request-id", uuid.New().String()},
			{"accept", "*/*"},
			{"origin", BaseEndpoint},
			{"sec-fetch-site", "same-origin"},
			{"sec-fetch-mode", "cors"},
			{"sec-fetch-dest", "empty"},
			{"referer", BasePaymentEndpoint},
			{"accept-encoding", "gzip, deflate, br"},
			{"accept-language", "en-US,en;q=0.9"},
		},
		Data:               data,
		ResponseBodyStruct: &prelookupResonse,
	})
	if err != nil {
		fmt.Println(err)
	}
	defer resp.Body.Close()

	switch resp.StatusCode {
	case 200:
		task.CheckoutInfo.ThreeDsID = prelookupResonse.Threedsreferenceid
		return true
	default:
		return false
	}
}

// PlaceOrder completes the checkout by placing the order then sends a webhook depending on if successfully checked out or not
func (task *Task) PlaceOrder() bool {
	data, err := json.Marshal(PlaceOrderRequest{
		Orderid: task.CheckoutInfo.ID,
		Threedsecurestatus: Threedsecurestatus{
			Threedsreferenceid: task.CheckoutInfo.ThreeDsID,
		},
	})
	ok := util.HandleErrors(err, util.RequestMarshalBodyError)
	if !ok {
		return false
	}

	resp, err := util.MakeRequest(&util.Request{
		Client: task.Task.Client,
		Method: "POST",
		URL:    PlaceOrderEndpoint,
		RawHeaders: [][2]string{
			{"content-length", fmt.Sprint(len(data))},
			{"pragma", "no-cache"},
			{"cache-control", "no-cache"},
			{"sec-ch-ua", "\" Not;A Brand\";v=\"99\", \"Google Chrome\";v=\"91\", \"Chromium\";v=\"91\""},
			{"accept", "application/json, text/javascript, */*; q=0.01"},
			{"x-user-interface", "DotCom-Optimized"},
			{"x-native-checkout-version", "__VERSION__"},
			{"sec-ch-ua-mobile", "?0"},
			{"user-agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/91.0.4472.77 Safari/537.36"},
			{"content-type", "application/json"},
			{"origin", BaseEndpoint},
			{"sec-fetch-site", "same-origin"},
			{"sec-fetch-mode", "cors"},
			{"sec-fetch-dest", "empty"},
			{"referer", BasePaymentEndpoint},
			{"accept-encoding", "gzip, deflate, br"},
			{"accept-language", "en-US,en;q=0.9"},
		},
		Data: data,
	})
	ok = util.HandleErrors(err, util.RequestDoError)
	if !ok {
		return false
	}

	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return false
	}

	data, _ = json.Marshal(Browserinfo{
		Javaenabled: false,
		Language:    "en-US",
		Useragent:   "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/91.0.4472.77 Safari/537.36",
		Height:      "1920",
		Width:       "1080",
		Timezone:    "420",
		Colordepth:  "24",
	})
	resp, err = util.MakeRequest(&util.Request{
		Client: task.Task.Client,
		Method: "POST",
		URL:    fmt.Sprintf(OrderEndpoint, task.CheckoutInfo.ID) + "/",
		RawHeaders: [][2]string{
			{"content-length", fmt.Sprint(len(data))},
			{"pragma", "no-cache"},
			{"cache-control", "no-cache"},
			{"sec-ch-ua", "\" Not;A Brand\";v=\"99\", \"Google Chrome\";v=\"91\", \"Chromium\";v=\"91\""},
			{"accept", "application/com.bestbuy.order+json"},
			{"x-user-interface", "DotCom-Optimized"},
			{"sec-ch-ua-mobile", "?0"},
			{"user-agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/91.0.4472.77 Safari/537.36"},
			{"content-type", "application/json"},
			{"origin", BaseEndpoint},
			{"sec-fetch-site", "same-origin"},
			{"sec-fetch-mode", "cors"},
			{"sec-fetch-dest", "empty"},
			{"referer", BasePaymentEndpoint},
			{"accept-encoding", "gzip, deflate, br"},
			{"accept-language", "en-US,en;q=0.9"},
		},
		Data: data,
	})
	ok = util.HandleErrors(err, util.RequestDoError)
	if !ok {
		return false
	}

	defer resp.Body.Close()

	var success bool
	switch resp.StatusCode {
	case 200:
		for _, cookie := range resp.Cookies() {
			if cookie.Name == "CartItemCount" {
				if cookie.Value == "0" {
					fmt.Println("Checked Out")
					success = true
				}

			}
		}
	default:
		fmt.Println("Failed to Checkout")
		success = false
	}

	util.SendDiscordWebhook(task.Task.DiscordWebhook, success, task.CreateBestbuyFields(success), task.CheckoutInfo.ImageUrl)
	return success
}
