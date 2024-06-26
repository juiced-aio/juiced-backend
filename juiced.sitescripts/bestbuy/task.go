package bestbuy

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"log"
	"strconv"
	"strings"
	"time"

	"backend.juicedbot.io/juiced.client/http"
	"backend.juicedbot.io/juiced.infrastructure/common"
	"backend.juicedbot.io/juiced.infrastructure/common/entities"
	"backend.juicedbot.io/juiced.infrastructure/common/enums"
	"backend.juicedbot.io/juiced.infrastructure/common/events"
	"backend.juicedbot.io/juiced.sitescripts/base"
	"backend.juicedbot.io/juiced.sitescripts/util"
	"github.com/anaskhan96/soup"

	browser "github.com/eddycjy/fake-useragent"
	"github.com/google/uuid"
)

// CreateBestbuyTask takes a Task entity and turns it into a Bestbuy Task
func CreateBestbuyTask(task *entities.Task, profile entities.Profile, proxyGroup *entities.ProxyGroup, eventBus *events.EventBus, taskType enums.TaskType, locationID, email, password string) (Task, error) {
	bestbuyTask := Task{
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
		TaskType:   taskType,
		LocationID: locationID,
	}
	if proxyGroup != nil {
		bestbuyTask.Task.Proxy = util.RandomLeastUsedProxy(proxyGroup.Proxies)
	} else {
		bestbuyTask.Task.Proxy = nil
	}
	return bestbuyTask, nil
}

// PublishEvent wraps the EventBus's PublishTaskEvent function
func (task *Task) PublishEvent(status enums.TaskStatus, eventType enums.TaskEventType, statusPercentage int) {
	if status == enums.TaskIdle || !task.Task.StopFlag {
		task.Task.Task.SetTaskStatus(status)
		task.Task.EventBus.PublishTaskEvent(status, statusPercentage, eventType, nil, task.Task.Task.ID)
	}
}

// CheckForStop checks the stop flag and stops the monitor if it's true
func (task *Task) CheckForStop() bool {
	if task.Task.StopFlag && !task.Task.DontPublishEvents {
		task.PublishEvent(enums.TaskIdle, enums.TaskStop, 0)
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
	// If the function panics due to a runtime error, recover from it
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
	task.StockData = BestbuyInStockData{}
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

	// 1. Login / Become a guest
	sessionMade := false
	for !sessionMade {
		needToStop := task.CheckForStop()
		if needToStop {
			return
		}
		switch task.TaskType {
		case enums.TaskTypeAccount:
			task.PublishEvent(enums.LoggingIn, enums.TaskStart, 10)
			sessionMade = task.Login()
			if sessionMade {
				task.ClearCart()
			}
		case enums.TaskTypeGuest:
			task.PublishEvent(enums.SettingUp, enums.TaskStart, 10)
			sessionMade = BecomeGuest(task.Task.Client)
		}

		if !sessionMade {
			time.Sleep(time.Duration(task.Task.Task.TaskDelay) * time.Millisecond)
		}
	}

	task.PublishEvent(enums.WaitingForMonitor, enums.TaskUpdate, 20)
	// 2. WaitForMonitor
	needToStop := task.WaitForMonitor()
	if needToStop {
		return
	}

	task.PublishEvent(enums.AddingToCart, enums.TaskUpdate, 30)
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
	needToStop = task.CheckForStop()
	if needToStop {
		return
	}

	task.PublishEvent(enums.GettingCartInfo, enums.TaskUpdate, 35)

	startTime := time.Now()
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

	task.PublishEvent(enums.SettingShippingInfo, enums.TaskUpdate, 60)
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

	task.PublishEvent(enums.SettingBillingInfo, enums.TaskUpdate, 75)
	// 6. SetPaymentInfo
	setPaymentInfo := false
	doNotRetry := false
	for !setPaymentInfo {
		needToStop := task.CheckForStop()
		if needToStop || doNotRetry {
			return
		}
		setPaymentInfo, doNotRetry = task.SetPaymentInfo()
		if !setPaymentInfo {
			time.Sleep(time.Duration(task.Task.Task.TaskDelay) * time.Millisecond)
		}
	}

	task.PublishEvent(enums.CheckingOut, enums.TaskUpdate, 90)
	// 7. PlaceOrder
	placedOrder := false
	var retries int
	status := enums.OrderStatusFailed
	for !placedOrder {
		needToStop := task.CheckForStop()
		if needToStop {
			return
		}
		if status != enums.OrderStatusFailed || retries > common.MAX_RETRIES {
			break
		}
		placedOrder, status = task.PlaceOrder()
		if !placedOrder {
			retries++
			time.Sleep(time.Duration(task.Task.Task.TaskDelay) * time.Millisecond)
		}
	}

	endTime := time.Now()

	log.Println("STARTED AT: " + startTime.String())
	log.Println("  ENDED AT: " + endTime.String())
	log.Println("TIME TO CHECK OUT: ", endTime.Sub(startTime).Milliseconds())

	switch status {
	case enums.OrderStatusSuccess:
		task.PublishEvent(enums.CheckingOutSuccess, enums.TaskComplete, 100)
	case enums.OrderStatusDeclined:
		task.PublishEvent(enums.CardDeclined, enums.TaskComplete, 100)
	case enums.OrderStatusFailed:
		task.PublishEvent(fmt.Sprintf(enums.CheckingOutFailure, "Unknown error"), enums.TaskComplete, 100)
	default:
		task.PublishEvent(fmt.Sprintf(enums.CheckingOutFailure, status), enums.TaskComplete, 100)
	}

	quantity := task.Task.Task.TaskQty
	if task.StockData.MaxQuantity != 0 && quantity > task.StockData.MaxQuantity {
		quantity = task.StockData.MaxQuantity
	}

	go util.ProcessCheckout(&util.ProcessCheckoutInfo{
		BaseTask:     task.Task,
		Success:      placedOrder,
		Status:       status,
		Content:      "",
		Embeds:       task.CreateBestbuyEmbed(status, task.StockData.ImageURL),
		ItemName:     task.StockData.ProductName,
		ImageURL:     task.StockData.ImageURL,
		Sku:          task.StockData.SKU,
		Retailer:     enums.BestBuy,
		Price:        float64(task.StockData.Price),
		Quantity:     quantity,
		MsToCheckout: time.Since(startTime).Milliseconds(),
	})

}

// Login logs the task's client into the account specified
func (task *Task) Login() bool {
	resp, _, err := util.MakeRequest(&util.Request{
		Client:     task.Task.Client,
		Method:     "GET",
		URL:        BaseEndpoint,
		RawHeaders: DefaultRawHeaders,
	})
	if err != nil || resp.StatusCode != 200 {
		fmt.Println(err.Error())
		return false
	}

	err = util.NewAbck(&task.Task.Client, BaseEndpoint+"/", BaseEndpoint, AkamaiEndpoint)
	if err != nil {
		return false
	}

	resp, body, err := util.MakeRequest(&util.Request{
		Client: task.Task.Client,
		Method: "GET",
		URL:    LoginPageEndpoint,
		RawHeaders: [][2]string{
			{"pragma", "no-cache"},
			{"cache-control", "no-cache"},
			{"upgrade-insecure-requests", "1"},
			{"user-agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/91.0.4472.124 Safari/537.36"},
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
		fmt.Println(err.Error())
		return false
	}

	tmxURL := "https://tmx.bestbuy.com/jx2u3dtlvr835clj.js?%v=ummqowa2&%v=%v"
	var ZPLANK string
	for _, cookie := range task.Task.Client.Jar.Cookies(ParsedBase) {
		if cookie.Name == "ZPLANK" {
			ZPLANK = cookie.Value
		}
	}
	_, _, err = util.MakeRequest(&util.Request{
		Client: task.Task.Client,
		Method: "GET",
		URL:    fmt.Sprintf(tmxURL, common.RandString(16), common.RandString(16), ZPLANK),
	})
	if err != nil || resp.StatusCode != 200 {
		fmt.Println(err.Error())
		return false
	}

	// Getting all the json data
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

	var userAgent = `{"userAgent":"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/91.0.4472.124 Safari/537.36"}`
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
	{"token":"` + signinJson.Token + `","activity":"1:user-activity-2016-09:` + encryptedActivity + `","loginMethod":"UID_PASSWORD","flowOptions":"000000000000000","alpha":"` + correctData["alpha"] + `","Salmon":"FA7F2","encryptedEmail":"1:email-2017-01:` + encryptedEmail + `","` + correctData["pass"] + `":"` + password + `","info":"1:user-activity-2016-09:` + encryptedInfo + `","` + signinJson.Emailfieldname + `":"` + email + `"}
	`))

	err = util.NewAbck(&task.Task.Client, LoginPageEndpoint+"/", BaseEndpoint, AkamaiEndpoint)
	if err != nil {
		return false
	}
	var loginResponse LoginResponse
	resp, _, err = util.MakeRequest(&util.Request{
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
			{"user-agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/91.0.4472.124 Safari/537.36"},
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
		fmt.Println(err.Error())
		return false
	}

	resp, _, err = util.MakeRequest(&util.Request{
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
		fmt.Println(err.Error())
		return false
	}

	return true
}

func (task *Task) ClearCart() bool {
	var clearCartResponse ClearCartResponse
	resp, _, err := util.MakeRequest(&util.Request{
		Client:             task.Task.Client,
		Method:             "GET",
		URL:                CartInfoEndpoint,
		RawHeaders:         DefaultRawHeaders,
		ResponseBodyStruct: &clearCartResponse,
	})
	ok := util.HandleErrors(err, util.RequestDoError)
	if !ok || resp.StatusCode != 200 {
		return false
	}

	for _, lineItem := range clearCartResponse.Cart.Lineitems {
		resp, _, err := util.MakeRequest(&util.Request{
			Client:     task.Task.Client,
			Method:     "DELETE",
			URL:        BaseEndpoint + fmt.Sprintf("/cart/item/%v", lineItem.ID),
			RawHeaders: DefaultRawHeaders,
		})
		ok := util.HandleErrors(err, util.RequestDoError)
		if !ok || resp.StatusCode != 200 {
			return false
		}
	}

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
			task.Task.HasStockData = true
			return false
		}
		time.Sleep(common.MS_TO_WAIT)
	}
}

// AddToCart adds the item to cart and also handles a queue if there is one
func (task *Task) AddToCart() bool {
	addToCartRequest := AddToCartRequest{
		Items: []Items{},
	}
	quantity := task.Task.Task.TaskQty
	if task.StockData.MaxQuantity != 0 && quantity > task.StockData.MaxQuantity {
		quantity = task.StockData.MaxQuantity
	}
	for i := 0; i < quantity; i++ {
		addToCartRequest.Items = append(addToCartRequest.Items, Items{Skuid: task.StockData.SKU})
	}

	data, _ := json.Marshal(addToCartRequest)
	addToCartResponse := AddToCartResponse{}
	var handled bool

	for !handled {
		needToStop := task.CheckForStop()
		if needToStop {
			return true
		}
		for _, cookie := range task.Task.Client.Jar.Cookies(ParsedBase) {
			if cookie.Name == "_abck" {
				validator, _ := util.FindInString(cookie.Value, "~", "~")
				if validator == "-1" {
					err := util.NewAbck(&task.Task.Client, fmt.Sprintf("https://www.bestbuy.com/site/%v.p?skuId=%v", task.StockData.SKU, task.StockData.SKU), BaseEndpoint, AkamaiEndpoint)
					if err != nil {
						return false
					}
				}
			}
		}

		resp, _, err := util.MakeRequest(&util.Request{
			Client: task.Task.Client,
			Method: "POST",
			URL:    AddToCartEndpoint,
			RawHeaders: [][2]string{
				{"content-length", fmt.Sprint(len(data))},
				{"sec-ch-ua", "\" Not A;Brand\";v=\"99\", \"Chromium\";v=\"90\", \"Google Chrome\";v=\"90\""},
				{"accept", "application/json"},
				{"sec-ch-ua-mobile", "?0"},
				{"user-agent", browser.Chrome()},
				{"content-type", "application/json; charset=UTF-8"},
				{"origin", BaseEndpoint},
				{"sec-fetch-site", "same-origin"},
				{"sec-fetch-mode", "cors"},
				{"sec-fetch-dest", "empty"},
				{"referer", fmt.Sprintf("https://www.bestbuy.com/site/%v.p?skuId=%v", task.StockData.SKU, task.StockData.SKU)},
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

		switch resp.StatusCode {
		case 200:
			handled = true
		case 400:
			switch addToCartResponse.Errorsummary.Errorcode {
			case "ITEM_MAX_QUANTITY_EXCEEDED":
				quantity, err := common.FindInString(addToCartResponse.Errorsummary.Message, "limited to ", " per customer")
				if err != nil {
					return false
				}
				task.StockData.MaxQuantity, err = strconv.Atoi(quantity)
				if err != nil {
					return false
				}

				return false
			case "CONSTRAINED_ITEM":
				handled = task.HandleQueue(resp, data)
			}

		case 500:
			if task.TaskType == enums.TaskTypeGuest {
				time.Sleep(time.Duration(task.Task.Task.TaskDelay) * time.Millisecond)
			} else {
				time.Sleep(3 * time.Second)
			}
		}

	}
	return true
}

func (task *Task) HandleQueue(resp *http.Response, data []byte) (handled bool) {
	a2TransactionCode := resp.Header.Get("a2ctransactioncode")
	a2TransactionID := resp.Header.Get("a2ctransactionreferenceid")
	times, err := CheckTime(a2TransactionCode)
	if err != nil {
		fmt.Println(err.Error())
		return
	}
	fmt.Println(times)
	if times < 6 {
		// @silent: I added these just because, I think the users will like something like this though
		task.PublishEvent("Queued for "+fmt.Sprint(int(times))+" minutes, Joining Queue", enums.TaskUpdate, 40)
		for _, cookie := range task.Task.Client.Jar.Cookies(ParsedBase) {
			if cookie.Name == "_abck" {
				validator, _ := util.FindInString(cookie.Value, "~", "~")
				if validator == "-1" {
					err := util.NewAbck(&task.Task.Client, fmt.Sprintf("https://www.bestbuy.com/site/%v.p?skuId=%v", task.StockData.SKU, task.StockData.SKU), BaseEndpoint, AkamaiEndpoint)
					if err != nil {
						return
					}
				}
			}
		}
		fmt.Println("Joining Queue")
		queueChan := make(chan bool)
		go func() {
			time.Sleep(time.Duration(times*60000) * time.Millisecond)
			queueChan <- true
		}()
		go func() {
			for {
				queueChan <- false
				time.Sleep(common.MS_TO_WAIT)
			}
		}()
		for {
			qaf := <-queueChan
			if qaf {
				break
			}
			needToStop := task.CheckForStop()
			if needToStop {
				return
			}
		}
		task.PublishEvent("Queue is up", enums.TaskUpdate, 45)
		fmt.Println("Out of Queue")
		addToCartResponse := AddToCartResponse{}
		resp, _, err = util.MakeRequest(&util.Request{
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
				{"referer", fmt.Sprintf("https://www.bestbuy.com/site/%v.p?skuId=%v", task.StockData.SKU, task.StockData.SKU)},
				{"accept-encoding", "gzip, deflate, br"},
				{"accept-language", "en-US,en;q=0.9"},
			},
			Data:               data,
			ResponseBodyStruct: &addToCartResponse,
		})
		ok := util.HandleErrors(err, util.RequestDoError)
		if !ok {
			return
		}

		switch resp.StatusCode {
		case 200:
			handled = true
		case 400:
			switch addToCartResponse.Errorsummary.Errorcode {
			case "ITEM_MAX_QUANTITY_EXCEEDED":
				quantity, err := common.FindInString(addToCartResponse.Errorsummary.Message, "limited to ", " per customer")
				if err != nil {
					return
				}
				task.StockData.MaxQuantity, err = strconv.Atoi(quantity)
				if err != nil {
					return
				}

				return false
			case "CONSTRAINED_ITEM":
				return task.HandleQueue(resp, data)
			}
		}
		return
	} else {
		task.PublishEvent("Queued for "+fmt.Sprint(int(times))+" minutes, Retrying", enums.TaskUpdate, 35)
		//	As a guest you do not ever get blocked adding to cart, but while logged in you will get blocked
		if task.TaskType == enums.TaskTypeGuest {
			time.Sleep(time.Duration(task.Task.Task.TaskDelay) * time.Millisecond)
		} else {
			time.Sleep(3 * time.Second)
		}
		return task.AddToCart()
	}

}

// Checkout goes to the checkout page and gets the required information for the rest of the checkout process
func (task *Task) Checkout() bool {
	resp, body, err := util.MakeRequest(&util.Request{
		Client:     task.Task.Client,
		Method:     "GET",
		URL:        CheckoutEndpoint,
		RawHeaders: DefaultRawHeaders,
	})
	ok := util.HandleErrors(err, util.RequestDoError)
	if !ok {
		return false
	}

	switch resp.StatusCode {
	case 200:
		rawOrderData, err := util.FindInString(body, `var orderData = `, `;`)
		if err != nil {
			return false
		}

		orderData := OrderData{}

		err = json.Unmarshal([]byte(rawOrderData), &orderData)
		if err != nil {
			return false
		}

		var itemIDs []string
		for _, item := range orderData.Items {
			itemIDs = append(itemIDs, item.ID)
		}

		if len(orderData.Items) > 0 {
			task.CheckoutInfo.ID = orderData.ID
			task.CheckoutInfo.ItemIDs = append(task.CheckoutInfo.ItemIDs, itemIDs...)
			task.CheckoutInfo.PaymentID = orderData.Payment.ID
			task.CheckoutInfo.OrderID = orderData.Customerorderid
			return true
		}

	}

	return false
}

// SetShippingInfo sets the shipping info in checkout
func (task *Task) SetShippingInfo() bool {

	for _, cookie := range task.Task.Client.Jar.Cookies(ParsedBase) {
		if cookie.Name == "_abck" {
			validator, _ := util.FindInString(cookie.Value, "~", "~")
			if validator == "-1" {
				err := util.NewAbck(&task.Task.Client, BaseShippingEndpoint, BaseEndpoint, AkamaiEndpoint)
				if err != nil {
					return false
				}
			}
		}
	}

	var shipOrPickupRequest ShipOrPickupRequest

	for _, itemID := range task.CheckoutInfo.ItemIDs {
		if task.LocationID == "" {
			shipOrPickupRequest = append(shipOrPickupRequest, ShipOrPickup{
				ID: itemID,
				Selectedfulfillment: Selectedfulfillment1{
					Shipping: Shipping1{},
				},
			})
		} else {
			shipOrPickupRequest = append(shipOrPickupRequest, ShipOrPickup{
				ID:                   itemID,
				StoreFulfillmentType: "InStore",
				Type:                 "DEFAULT",
				Selectedfulfillment: Selectedfulfillment1{
					InStorePickup: InStorePickup{
						PickupStoreID:         task.LocationID,
						DisplayDateType:       "IN_HAND",
						IsAvailableAtLocation: true,
						IsSTSAvailable:        false,
					},
				},
			})
		}
	}

	data, err := json.Marshal(shipOrPickupRequest)
	ok := util.HandleErrors(err, util.RequestMarshalBodyError)
	if !ok {
		log.Println(571)
		log.Println(err.Error())
		return false
	}

	resp, _, err := util.MakeRequest(&util.Request{
		Client: task.Task.Client,
		Method: "PATCH",
		URL:    fmt.Sprintf(OrderEndpoint, task.CheckoutInfo.ID) + "/items",
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
		Data:              data,
		RequestBodyStruct: shipOrPickupRequest,
	})
	ok = util.HandleErrors(err, util.RequestDoError)
	if !ok || resp.StatusCode != 200 {
		log.Println(604)
		log.Println(err.Error())
		return false
	}

	setShippingRequest := SetShippingRequest{
		Phonenumber:     task.Task.Profile.PhoneNumber,
		Smsnotifynumber: "",
		Smsoptin:        false,
		Emailaddress:    task.Task.Profile.Email,
		Items:           []ShipItems{},
	}
	for _, itemID := range task.CheckoutInfo.ItemIDs {
		setShippingRequest.Items = append(setShippingRequest.Items, ShipItems{
			ID:   itemID,
			Type: "DEFAULT",
			Selectedfulfillment: Selectedfulfillment{
				Shipping: Shipping{
					Address: ShipAddress{
						Country:             task.Task.Profile.ShippingAddress.CountryCode,
						Savetoprofile:       false,
						Street2:             task.Task.Profile.ShippingAddress.Address2,
						Useaddressasbilling: false,
						Middleinitial:       "",
						Lastname:            task.Task.Profile.ShippingAddress.LastName,
						Street:              task.Task.Profile.ShippingAddress.Address1,
						City:                task.Task.Profile.ShippingAddress.City,
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
		})
	}
	data, err = json.Marshal(setShippingRequest)
	ok = util.HandleErrors(err, util.RequestMarshalBodyError)
	if !ok {
		log.Println(571)
		log.Println(err.Error())
		return false
	}

	setShippingResponse := UniversalOrderResponse{}

	resp, _, err = util.MakeRequest(&util.Request{
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
		log.Println(604)
		log.Println(err.Error())
		return false
	}

	if resp.StatusCode != 200 {
		log.Println(606)
		log.Println(resp.StatusCode)
		err = util.NewAbck(&task.Task.Client, BaseShippingEndpoint, BaseEndpoint, AkamaiEndpoint)
		if err != nil {
			return false
		}
		return false
	}

	if len(setShippingResponse.Errors) > 0 {
		switch setShippingResponse.Errors[0].Errorcode {
		case "standardizationError":
			log.Println("Bad shipping details")
			return false
		}
	}

	resp, _, err = util.MakeRequest(&util.Request{
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
		log.Println(642)
		log.Println(err.Error())
		return false
	}

	switch resp.StatusCode {
	case 200:
		return true
	default:
		err = util.NewAbck(&task.Task.Client, BaseShippingEndpoint, BaseEndpoint, AkamaiEndpoint)
		if err != nil {
			return false
		}
		return false
	}

}

// SetPaymentInfo sets the payment info in checkout
func (task *Task) SetPaymentInfo() (bool, bool) {

	if !common.ValidCardType([]byte(task.Task.Profile.CreditCard.CardNumber), task.Task.Task.TaskRetailer) {
		return false, true
	}

	cardType := util.GetCardType([]byte(task.Task.Profile.CreditCard.CardNumber), task.Task.Task.TaskRetailer)

	for _, cookie := range task.Task.Client.Jar.Cookies(ParsedBase) {
		if cookie.Name == "_abck" {
			validator, _ := util.FindInString(cookie.Value, "~", "~")
			if validator == "-1" {
				err := util.NewAbck(&task.Task.Client, BasePaymentEndpoint, BaseEndpoint, AkamaiEndpoint)
				if err != nil {
					return false, false
				}
			}
		}
	}
	billing := Billingaddress{
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
			Binnumber:       task.Task.Profile.CreditCard.CardNumber[0:6],
			Ispwpregistered: false,
			Expmonth:        task.Task.Profile.CreditCard.ExpMonth,
			Expyear:         task.Task.Profile.CreditCard.ExpYear,
			Cvv:             task.Task.Profile.CreditCard.CVV,
			Orderid:         task.CheckoutInfo.OrderID,
			Savetoprofile:   false,
			Type:            cardType,
			International:   false,
			Virtualcard:     false,
		},
	})
	resp, _, err := util.MakeRequest(&util.Request{
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
		fmt.Println(err.Error())
	}

	if resp.StatusCode != 200 {
		util.NewAbck(&task.Task.Client, BasePaymentEndpoint, BaseEndpoint, AkamaiEndpoint)
		return false, false
	}

	for _, cookie := range task.Task.Client.Jar.Cookies(ParsedBase) {
		if cookie.Name == "_abck" {
			validator, _ := util.FindInString(cookie.Value, "~", "~")
			if validator == "-1" {
				err = util.NewAbck(&task.Task.Client, BasePaymentEndpoint, BaseEndpoint, AkamaiEndpoint)
				if err != nil {
					return false, false
				}
			}
		}
	}

	resp, _, err = util.MakeRequest(&util.Request{
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
		fmt.Println(err.Error())
	}

	if resp.StatusCode != 200 {
		util.NewAbck(&task.Task.Client, BasePaymentEndpoint, BaseEndpoint, AkamaiEndpoint)
		return false, false
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
	resp, _, err = util.MakeRequest(&util.Request{
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
		fmt.Println(err.Error())
	}

	switch resp.StatusCode {
	case 200:
		task.CheckoutInfo.ThreeDsID = prelookupResonse.Threedsreferenceid
		return true, false
	default:
		return false, false
	}
}

// PlaceOrder completes the checkout by placing the order then sends a webhook depending on if successfully checked out or not
func (task *Task) PlaceOrder() (bool, enums.OrderStatus) {
	status := enums.OrderStatusFailed
	for _, cookie := range task.Task.Client.Jar.Cookies(ParsedBase) {
		if cookie.Name == "_abck" {
			validator, _ := util.FindInString(cookie.Value, "~", "~")
			if validator == "-1" {
				err := util.NewAbck(&task.Task.Client, BasePaymentEndpoint, BaseEndpoint, AkamaiEndpoint)
				if err != nil {
					return false, status
				}
			}
		}
	}
	data, err := json.Marshal(PlaceOrderRequest{
		Orderid: task.CheckoutInfo.ID,
		Threedsecurestatus: Threedsecurestatus{
			Threedsreferenceid: task.CheckoutInfo.ThreeDsID,
		},
	})
	ok := util.HandleErrors(err, util.RequestMarshalBodyError)
	if !ok {
		return false, status
	}

	resp, _, err := util.MakeRequest(&util.Request{
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
		return false, status
	}

	if resp.StatusCode != 200 {
		return false, status
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
	placeOrderResponse := UniversalOrderResponse{}
	resp, _, err = util.MakeRequest(&util.Request{
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
		Data:               data,
		ResponseBodyStruct: &placeOrderResponse,
	})
	ok = util.HandleErrors(err, util.RequestDoError)
	if !ok {
		return false, status
	}

	var success bool

	switch resp.StatusCode {
	case 200:
		for _, cookie := range resp.Cookies() {
			if cookie.Name == "CartItemCount" {
				if cookie.Value == "0" {
					fmt.Println("Checked Out")
					status = enums.OrderStatusSuccess
					success = true
				}

			}
		}
	default:
		if len(placeOrderResponse.Errors) > 0 {
			switch placeOrderResponse.Errors[0].Errorcode {
			case "CC_AUTH_FAILURE":
				fmt.Println("Card declined")
				status = enums.OrderStatusDeclined
			case "ITEM_EXCEEDED_ORDER_LIMIT":
				fmt.Println("Order limit exceeded")
				status = "Item limit exceeded"
			case "ACCOUNT_CREATION_REQUIRED":
				fmt.Println("Account creation required")
				status = "Account Required"
			default:
				fmt.Println("Failed to Checkout", placeOrderResponse)
				status = enums.OrderStatusFailed
			}
		} else {
			fmt.Println("Failed to Checkout", placeOrderResponse)
			status = enums.OrderStatusFailed
		}
		success = false

	}

	return success, status
}
