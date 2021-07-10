package target

import (
	"encoding/json"
	"fmt"
	"log"
	"net/url"
	"strings"
	"time"

	"backend.juicedbot.io/juiced.client/http/cookiejar"

	"backend.juicedbot.io/juiced.client/http"
	"backend.juicedbot.io/juiced.infrastructure/common"
	"backend.juicedbot.io/juiced.infrastructure/common/entities"
	"backend.juicedbot.io/juiced.infrastructure/common/enums"
	"backend.juicedbot.io/juiced.infrastructure/common/events"
	"backend.juicedbot.io/juiced.infrastructure/queries"
	"backend.juicedbot.io/juiced.sitescripts/base"
	"backend.juicedbot.io/juiced.sitescripts/util"
	"github.com/dgrijalva/jwt-go"
	"github.com/go-rod/rod"
	"github.com/go-rod/rod/lib/launcher"
	"github.com/go-rod/rod/lib/launcher/flags"
	"github.com/go-rod/stealth"
	cmap "github.com/orcaman/concurrent-map"
)

// TODO @silent: Handle proxies
// TODO @silent: Handle errors
// TODO @silent: Allow multiple tasks to use cookies from one account
// TODO @silent: Code for steps is repetitive, should abstract
// TODO @Humphreyyyy: Handle mid-checkout sellouts (at some point)
// 		TODO @silent: Mid-checkout sellout errors may have to propagate back up to the monitor

var TargetAccountStore = cmap.New()

// CreateTargetTask takes a Task entity and turns it into a Target Task
func CreateTargetTask(task *entities.Task, profile entities.Profile, proxy entities.Proxy, eventBus *events.EventBus, checkoutType enums.CheckoutType, email, password string, storeID string, paymentType enums.PaymentType) (Task, error) {
	targetTask := Task{}

	targetTask = Task{
		Task: base.Task{
			Task:     task,
			Profile:  profile,
			Proxy:    proxy,
			EventBus: eventBus,
		},
		CheckoutType: checkoutType,
		AccountInfo: AccountInfo{
			Email:          email,
			Password:       password,
			PaymentType:    paymentType,
			DefaultCardCVV: profile.CreditCard.CVV,
			StoreID:        storeID,
		},
	}
	return targetTask, nil
}

var baseURL, _ = url.Parse(BaseEndpoint)

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
// 		1. Login
// 		2. RefreshLogin (in background)
// 		3. WaitForMonitor
// 		4. AddToCart
// 		5. GetCartInfo
//		6. SetShippingInfo
// 		7. SetPaymentInfo
// 		8. PlaceOrder
func (task *Task) RunTask() {
	// If the function panics due to a runtime error, recover from it
	defer func() {
		if recover() != nil {
			task.Task.StopFlag = true
			task.PublishEvent(enums.TaskIdle, enums.TaskFail)
		}
		task.PublishEvent(enums.TaskIdle, enums.TaskComplete)
	}()

	client, err := util.CreateClient(task.Task.Proxy)
	if err != nil {
		return
	}
	task.Task.Client = client

	// 1. Login/Find account
	task.PublishEvent(enums.SettingUp, enums.TaskStart)
	if TargetAccountStore.Has(task.AccountInfo.Email) {
		for {
			// Error will be nil unless the item isn't in the map which we are already checking above
			value, _ := TargetAccountStore.Get(task.AccountInfo.Email)
			client, isClient := value.(http.Client)
			if isClient {
				if len(task.Task.Client.Jar.Cookies(baseURL)) == 0 {
					task.Task.Client = client
				}
				break
			} else {
				if task.Task.Task.TaskStatus != enums.WaitingForLogin {
					task.PublishEvent(enums.WaitingForLogin, enums.TaskUpdate)
				}
				time.Sleep(1 * time.Millisecond)
			}
		}

	} else {
		task.PublishEvent(enums.LoggingIn, enums.TaskUpdate)
		loggedIn := false
		for !loggedIn {
			needToStop := task.CheckForStop()
			if needToStop {
				return
			}
			loggedIn = task.Login()
			if !loggedIn {
				time.Sleep(time.Duration(task.Task.Task.TaskDelay) * time.Millisecond)
			}
		}
		// 2. RefreshLogin (in background)
		go task.RefreshLogin()
	}

	task.PublishEvent(enums.WaitingForMonitor, enums.TaskUpdate)
	// 3. WaitForMonitor
	needToStop := task.WaitForMonitor()
	if needToStop {
		return
	}

	task.PublishEvent(enums.AddingToCart, enums.TaskUpdate)
	// 4. AddtoCart
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
	startTime := time.Now()
	// 5. GetCartInfo
	gotCartInfo := false
	for !gotCartInfo {
		needToStop := task.CheckForStop()
		if needToStop {
			return
		}
		gotCartInfo = task.GetCartInfo()
		if !gotCartInfo {
			time.Sleep(time.Duration(task.Task.Task.TaskDelay) * time.Millisecond)
		}
	}

	task.PublishEvent(enums.SettingShippingInfo, enums.TaskUpdate)
	// 8. SetShippingInfo
	if task.AccountInfo.ShippingType == enums.ShippingTypeNEW {
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
	}

	task.PublishEvent(enums.SettingBillingInfo, enums.TaskUpdate)
	// 7. SetPaymentInfo
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
	// 8. PlaceOrder
	placedOrder := false
	dontRetry := false
	maxRetries := 5
	retries := 0
	status := enums.OrderStatusFailed
	for !placedOrder {
		needToStop := task.CheckForStop()
		if needToStop {
			return
		}
		if status == enums.OrderStatusDeclined || retries > maxRetries || dontRetry {
			break
		}
		placedOrder, status, dontRetry = task.PlaceOrder(startTime)
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
		task.PublishEvent(enums.CheckedOut, enums.TaskComplete)
	case enums.OrderStatusDeclined:
		task.PublishEvent(enums.CheckoutFailed, enums.TaskComplete)
	case enums.OrderStatusFailed:
		task.PublishEvent(enums.CheckoutFailed, enums.TaskComplete)
	}

}

// Login logs the user in and sets the task's cookies for the logged in user
// TODO @silent: Handle stop flag within Login function
func (task *Task) Login() bool {
	defer func() {
		if recover() != nil {
			TargetAccountStore.Remove(task.AccountInfo.Email)
		}
	}()
	TargetAccountStore.Set(task.AccountInfo.Email, false)
	var userPassProxy bool
	var username string
	var password string
	cookies := make([]*http.Cookie, 0)

	launcher_ := launcher.New()

	proxyCleaned := common.ProxyCleaner(task.Task.Proxy)
	if proxyCleaned != "" {
		proxyURL := proxyCleaned[7:]

		if strings.Contains(proxyURL, "@") {
			proxySplit := strings.Split(proxyURL, "@")
			proxyURL = proxySplit[1]
			userPass := strings.Split(proxySplit[0], ":")
			username = userPass[0]
			password = userPass[1]
			userPassProxy = true
		}

		launcher_ = launcher_.Proxy(proxyURL)
	}

	u := launcher_.Set(flags.Flag("headless")).
		// @silent: I disabled headless because after logging in a bunch today it seems I'm getting flagged and can't login unless the browser isn't headless,
		// and I'm guessing the users will run into this too since they might run many tasks with the same login. It's up to you if you want to
		// keep it headless or not. I also was running some bad chrome-flags for a while so it might actually never happen again.
		Delete(flags.Flag("--headless")).
		Delete(flags.Flag("--enable-automation")).
		Delete(flags.Flag("--restore-on-startup")).
		Set(flags.Flag("disable-background-networking")).
		Set(flags.Flag("enable-features"), "NetworkService,NetworkServiceInProcess").
		Set(flags.Flag("disable-background-timer-throttling")).
		Set(flags.Flag("disable-backgrounding-occluded-windows")).
		Set(flags.Flag("disable-breakpad")).
		Set(flags.Flag("disable-client-side-phishing-detection")).
		Set(flags.Flag("disable-default-apps")).
		Set(flags.Flag("disable-dev-shm-usage")).
		Set(flags.Flag("disable-extensions")).
		Set(flags.Flag("disable-features"), "site-per-process,TranslateUI,BlinkGenPropertyTrees").
		Set(flags.Flag("disable-hang-monitor")).
		Set(flags.Flag("disable-ipc-flooding-protection")).
		Set(flags.Flag("disable-popup-blocking")).
		Set(flags.Flag("disable-prompt-on-repost")).
		Set(flags.Flag("disable-renderer-backgrounding")).
		Set(flags.Flag("disable-sync")).
		Set(flags.Flag("force-color-profile"), "srgb").
		Set(flags.Flag("metrics-recording-only")).
		Set(flags.Flag("safebrowsing-disable-auto-update")).
		Set(flags.Flag("password-store"), "basic").
		Delete(flags.Flag("--use-mock-keychain")).
		MustLaunch()

	browser := rod.New().ControlURL(u).MustConnect()

	browser.MustIgnoreCertErrors(true)

	defer browser.MustClose()

	if userPassProxy {
		go browser.MustHandleAuth(username, password)()
	}

	page := stealth.MustPage(browser)

	page.MustNavigate(LoginEndpoint).WaitLoad()

	if strings.Contains(page.MustHTML(), "accessDenied-CheckVPN") {
		task.PublishEvent("Bad Proxy", enums.TaskUpdate)
		TargetAccountStore.Remove(task.AccountInfo.Email)
		return false
	}

	page.MustElement("#username").MustWaitVisible().Input(task.AccountInfo.Email)
	page.MustElement("#password").MustWaitVisible().Input(task.AccountInfo.Password)
	page.MustElementX(`//*[contains(@class, 'sc-hMqMXs ysAUA')]`).MustWaitVisible().MustClick()
	page.MustElement("#login").MustWaitVisible().MustClick().MustWaitLoad()
	page.MustElement("#account").MustWaitLoad()
	page.MustWaitLoad()
	page.MustNavigate(BaseEndpoint).MustWaitLoad()
	page.MustElement("#account").MustWaitLoad().MustWaitInteractable().MustClick()
	page.MustElement("#accountNav-signIn").MustWaitVisible().MustClick()
	page.MustWaitLoad()

	startTimeout := time.Now().Unix()
	browserCookies, _ := page.Cookies([]string{BaseEndpoint})
	for validCookie := false; !validCookie; {
		for _, cookie := range browserCookies {
			if cookie.Name == "accessToken" {
				claims := &LoginJWT{}
				new(jwt.Parser).ParseUnverified(cookie.Value, claims)
				if claims.Eid == task.AccountInfo.Email {
					validCookie = true
				}
			}
		}
		if time.Now().Unix()-startTimeout > 30 {
			TargetAccountStore.Remove(task.AccountInfo.Email)
			return false
		}
	}

	for _, cookie := range browserCookies {
		httpCookie := &http.Cookie{
			Name:   cookie.Name,
			Value:  cookie.Value,
			Domain: cookie.Domain,
			Path:   cookie.Path,
		}
		cookie.Value = strings.ReplaceAll(cookie.Value, `"`, "\"")
		if !strings.Contains(cookie.Value, `"`) {
			cookies = append(cookies, httpCookie)
		}
	}
	task.AccountInfo.Cookies = cookies
	task.Task.Client.Jar.SetCookies(baseURL, cookies)
	TargetAccountStore.Set(task.AccountInfo.Email, task.Task.Client)

	return true
}

// RefreshLogin refreshes the login tokens so that the user can remain logged in
func (task *Task) RefreshLogin() {
	// If the function panics due to a runtime error, recover and restart it
	defer func() {
		if recover() != nil {
			task.RefreshLogin()
		}
	}()

	for {
		success := true
		if task.AccountInfo.Refresh == 0 || time.Now().Unix() > task.AccountInfo.Refresh {
			refreshLoginResponse := RefreshLoginResponse{}
			client, hasClient := TargetAccountStore.Get(task.AccountInfo.Email)
			if !hasClient {
				continue
			}
			resp, _, err := util.MakeRequest(&util.Request{
				Client:             client.(http.Client),
				Method:             "POST",
				URL:                RefreshLoginEndpoint,
				AddHeadersFunction: AddTargetHeaders,
				Referer:            RefreshLoginReferer,
				RequestBodyStruct:  AutoGend,
				ResponseBodyStruct: &refreshLoginResponse,
			})
			if err != nil {
				success = false
				break
			}

			switch resp.StatusCode {
			case 201:
				claims := &LoginJWT{}

				new(jwt.Parser).ParseUnverified(string(refreshLoginResponse.AccessToken), claims)

				if err != nil || claims.Eid != task.AccountInfo.Email {
					success = false
					break
				}

				task.AccountInfo.Refresh = time.Now().Unix() + int64(refreshLoginResponse.ExpiresIn) - 300 // Refresh 5 mins before it expires, just in case
			default:
				success = false
			}
		}

		if !success {
			loggedIn := false
			for !loggedIn {
				task.Login()
				time.Sleep(time.Duration(task.Task.Task.TaskDelay) * time.Millisecond)
			}
		}
	}
}

// WaitForMonitor waits until the Monitor has sent the info to the task to continue
func (task *Task) WaitForMonitor() bool {
	for {
		needToStop := task.CheckForStop()
		if needToStop {
			return true
		}
		if task.InStockData.TCIN != "" {
			task.TCINType = task.InStockData.TCINType
			task.TCIN = task.InStockData.TCIN
			return false
		}
		time.Sleep(1 * time.Millisecond)
	}
}

// AddToCart sends a post request to the AddToCartEndpoint with a body determined by the CheckoutType
func (task *Task) AddToCart() bool {
	var data []byte
	var err error

	shipReq, err := json.Marshal(AddToCartShipRequest{
		CartType:        "REGULAR",
		ChannelID:       "90",
		ShoppingContext: "DIGITAL",
		CartItem: CartItem{
			TCIN:          task.TCIN,
			Quantity:      task.Task.Task.TaskQty,
			ItemChannelID: "10",
		},
	})
	ok := util.HandleErrors(err, util.RequestMarshalBodyError)
	if !ok {
		return false
	}
	pickupReq, err := json.Marshal(AddToCartPickupRequest{
		CartType:        "REGULAR",
		ChannelID:       "10",
		ShoppingContext: "DIGITAL",
		CartItem: CartItem{
			TCIN:          task.TCIN,
			Quantity:      task.Task.Task.TaskQty,
			ItemChannelID: "90",
		},
		Fulfillment: CartFulfillment{
			Type:       enums.CheckoutTypePICKUP,
			LocationID: task.AccountInfo.StoreID,
			ShipMethod: "STORE_PICKUP",
		},
	})
	ok = util.HandleErrors(err, util.RequestMarshalBodyError)
	if !ok {
		return false
	}

	switch task.CheckoutType {
	case enums.CheckoutTypeSHIP:
		data = shipReq
	case enums.CheckoutTypePICKUP:
		data = pickupReq
	case enums.CheckoutTypeEITHER:
		switch task.TCINType {
		case enums.CheckoutTypeSHIP:
			data = shipReq
		case enums.CheckoutTypePICKUP:
			data = pickupReq
		}
	}

	addToCartResponse := AddToCartResponse{}

	resp, _, err := util.MakeRequest(&util.Request{
		Client:             task.Task.Client,
		Method:             "POST",
		URL:                AddToCartEndpoint,
		AddHeadersFunction: AddTargetHeaders,
		Referer:            AddToCartReferer + task.TCIN,
		Data:               data,
		ResponseBodyStruct: &addToCartResponse,
	})
	if err != nil {
		return false
	}

	switch resp.StatusCode {
	case 201:
		task.AccountInfo.CartID = addToCartResponse.CartID
		return true
	default:
		if addToCartResponse.Error.Message == "Too Many Requests" {
			var newCookies []*http.Cookie
			for _, cookie := range task.Task.Client.Jar.Cookies(baseURL) {
				if cookie.Name != "TealeafAkaSid" {
					newCookies = append(newCookies, cookie)
				}
			}

			jar, _ := cookiejar.New(nil)
			task.Task.Client.Jar = jar
			// I'm going to modify the cookiejar package soon
			task.Task.Client.Jar.SetCookies(baseURL, newCookies)

		}
		return false
	}

}

// GetCartInfo returns the cart info needed for updating payment and placing an order
func (task *Task) GetCartInfo() bool {
	getCartInfoRequest := GetCartInfoRequest{
		CartType: "REGULAR",
	}
	getCartInfoResponse := GetCartInfoResponse{}
	resp, _, err := util.MakeRequest(&util.Request{
		Client:             task.Task.Client,
		Method:             "POST",
		URL:                GetCartInfoEndpoint,
		AddHeadersFunction: AddTargetHeaders,
		Referer:            GetCartInfoReferer,
		RequestBodyStruct:  getCartInfoRequest,
		ResponseBodyStruct: &getCartInfoResponse,
	})
	if err != nil {
		return false
	}

	switch resp.StatusCode {
	case 201:
		task.AccountInfo.CartInfo = getCartInfoResponse
		return true
	default:
		return false
	}

}

// SetShippingInfo sets the shipping address or does nothing if using a saved address
func (task *Task) SetShippingInfo() bool {
	resp, _, err := util.MakeRequest(&util.Request{
		Client:             task.Task.Client,
		Method:             "PUT",
		URL:                fmt.Sprintf(SetShippingInfoEndpoint, task.AccountInfo.CartInfo.Addresses[1].AddressID),
		AddHeadersFunction: AddTargetHeaders,
		Referer:            SetShippingInfoReferer,
		RequestBodyStruct: SetShippingInfoRequest{
			CartType: "REGULAR",
			Address: SetShippingInfoAddress{
				AddressLine1:  task.Task.Profile.ShippingAddress.Address1,
				AddressLine2:  "",
				AddressType:   "SHIPPING",
				City:          task.Task.Profile.ShippingAddress.City,
				Country:       task.Task.Profile.ShippingAddress.CountryCode,
				FirstName:     task.Task.Profile.ShippingAddress.FirstName,
				LastName:      task.Task.Profile.ShippingAddress.LastName,
				Mobile:        task.Task.Profile.PhoneNumber,
				SaveAsDefault: false,
				State:         task.Task.Profile.ShippingAddress.StateCode,
				ZipCode:       task.Task.Profile.ShippingAddress.ZipCode,
			},
			Selected:         true,
			SaveToProfile:    false,
			SkipVerification: true,
		},
	})
	if err != nil {
		return err == nil
	}

	if resp.StatusCode != 200 {
		return false
	}

	return true

}

// SetPaymentInfo sets the payment info to prepare for placing an order
func (task *Task) SetPaymentInfo() bool {
	var data []byte
	var err error
	var endpoint string
	if task.AccountInfo.PaymentType == enums.PaymentTypeSAVED && len(task.AccountInfo.CartInfo.PaymentInstructions) > 0 {
		data, err = json.Marshal(SetPaymentInfoSavedRequest{
			CartID:      task.AccountInfo.CartID,
			WalletMode:  "NONE",
			PaymentType: "CARD",
			CardDetails: CVV{
				CVV: task.AccountInfo.DefaultCardCVV,
			},
		})
		endpoint = fmt.Sprintf(SetPaymentInfoSAVEDEndpoint, task.AccountInfo.CartInfo.PaymentInstructions[0].PaymentInstructionID)
	} else {
		data, err = json.Marshal(SetPaymentInfoNewRequest{
			CartID:      task.AccountInfo.CartID,
			WalletMode:  "ADD",
			PaymentType: "CARD",
			CardDetails: CardDetails{
				CardName:    task.Task.Profile.BillingAddress.FirstName + " " + task.Task.Profile.BillingAddress.LastName,
				CardNumber:  task.Task.Profile.CreditCard.CardNumber,
				CVV:         task.Task.Profile.CreditCard.CVV,
				ExpiryMonth: task.Task.Profile.CreditCard.ExpMonth,
				ExpiryYear:  task.Task.Profile.CreditCard.ExpYear,
			},
			BillingAddress: BillingAddress{
				AddressLine1: task.Task.Profile.BillingAddress.Address1,
				City:         task.Task.Profile.BillingAddress.City,
				FirstName:    task.Task.Profile.BillingAddress.FirstName,
				LastName:     task.Task.Profile.BillingAddress.LastName,
				Phone:        task.Task.Profile.PhoneNumber,
				State:        task.Task.Profile.BillingAddress.StateCode,
				ZipCode:      task.Task.Profile.BillingAddress.ZipCode,
				Country:      task.Task.Profile.BillingAddress.CountryCode,
			},
		})
		endpoint = fmt.Sprintf(SetPaymentInfoNEWEndpoint, task.AccountInfo.CartInfo.PaymentInstructions[0].PaymentInstructionID)
	}
	ok := util.HandleErrors(err, util.RequestMarshalBodyError)
	if !ok {
		return false
	}

	resp, _, err := util.MakeRequest(&util.Request{
		Client:             task.Task.Client,
		Method:             "PUT",
		URL:                endpoint,
		AddHeadersFunction: AddTargetHeaders,
		Referer:            util.TernaryOperator(task.AccountInfo.PaymentType == enums.PaymentTypeSAVED, SetPaymentInfoSAVEDReferer, SetPaymentInfoNEWReferer).(string),
		Data:               data,
	})
	if err != nil {
		return false
	}

	// TODO: Handle various responses
	// Not much to handle here

	if resp.StatusCode != 200 {
		return false
	}
	return true
}

// PlaceOrder completes the checkout process
func (task *Task) PlaceOrder(startTime time.Time) (bool, enums.OrderStatus, bool) {
	status := enums.OrderStatusFailed
	placeOrderRequest := PlaceOrderRequest{
		CartType:  "REGULAR",
		ChannelID: 10,
	}
	placeOrderResponse := PlaceOrderResponse{}
	resp, _, err := util.MakeRequest(&util.Request{
		Client:             task.Task.Client,
		Method:             "POST",
		URL:                PlaceOrderEndpoint,
		AddHeadersFunction: AddTargetHeaders,
		Referer:            PlaceOrderReferer,
		RequestBodyStruct:  placeOrderRequest,
		ResponseBodyStruct: &placeOrderResponse,
	})
	if err != nil {
		return false, status, false
	}

	if placeOrderResponse.Error.Message == "Too Many Requests" {
		return false, status, true
	}

	// TODO: Handle various responses
	log.Println(placeOrderResponse.Message)
	var success bool
	var dontRetry bool
	switch resp.StatusCode {
	case 200:
		status = enums.OrderStatusSuccess
		success = true
		go task.TargetCancelMethod(placeOrderResponse)

	default:
		switch placeOrderResponse.Code {
		case "PAYMENT_DECLINED_EXCEPTION":
			status = enums.OrderStatusDeclined
			success = false
		case "CART_LOCKED":
			// @silent: This is happens when the account is already currently checking out another item.
			// I can't really think of a way around this and I'm wasting too much time trying to.
			success = false
			dontRetry = true
		default:
			status = enums.OrderStatusFailed
			success = false
		}
	}
	_, user, err := queries.GetUserInfo()
	if err != nil {
		fmt.Println("Could not get user info")
		return success, status, false
	}

	go util.ProcessCheckout(util.ProcessCheckoutInfo{
		BaseTask:     task.Task,
		Success:      success,
		Content:      "",
		Embeds:       task.CreateTargetEmbed(status, task.AccountInfo.CartInfo.CartItems[0].ItemAttributes.ImagePath),
		UserInfo:     user,
		ItemName:     task.AccountInfo.CartInfo.CartItems[0].ItemAttributes.Description,
		Sku:          task.TCIN,
		Retailer:     enums.Target,
		Price:        int(task.AccountInfo.CartInfo.CartItems[0].UnitPrice),
		Quantity:     1,
		MsToCheckout: time.Since(startTime).Milliseconds(),
	})

	return success, status, dontRetry
}

// Function to stop Target orders from canceling
func (task *Task) TargetCancelMethod(placeOrderResponse PlaceOrderResponse) {
	zip := strings.Split(placeOrderResponse.Orders[0].Addresses[0].ZipCode, "-")[0]
	vID, err := util.GetCookie(task.Task.Client, BaseEndpoint, "visitorId")
	if err != nil {
		fmt.Println("no visitorId cookie")
		return
	}
	tealeafAkasID, err := util.GetCookie(task.Task.Client, BaseEndpoint, "TealeafAkaSid")
	if err != nil {
		fmt.Println("no TealeafAkaSid cookie")
		return
	}

	vi := common.RandString(13) + fmt.Sprint(time.Now().Unix())
	targetCancelMethodRequest := TargetCancelMethodRequest{
		Records: []Records{
			{
				Appid:     "adaptive",
				Useragent: "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/91.0.4472.124 Safari/537.36",
				Network:   "unknown",
				B:         "Chrome 91",
				D:         "desktop",
				Z:         zip,
				N:         "adp-node|sha=3efc12b1,number=85098,source=client,canary=false,webCluster=prod",
				V:         vID,
				Vi:        vi,
				T:         time.Now().Unix(),
				UserAgent: UserAgent{
					DeviceFormFactor: "desktop",
					Name:             "Chrome 91",
					Network:          "unknown",
					Original:         "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/91.0.4472.124 Safari/537.36",
				},
				Metrics: []Metrics{
					{
						E: "checkout_review.checkout_place_order_success",
						M: M{
							CartIndicators: CartIndicators{
								HasShippingRequiredItem:         fmt.Sprint(placeOrderResponse.Orders[0].Indicators.HasShippingRequiredItem),
								HasPaymentApplied:               fmt.Sprint(placeOrderResponse.Orders[0].Indicators.HasPaymentApplied),
								HasPaymentSatisfied:             fmt.Sprint(placeOrderResponse.Orders[0].Indicators.HasPaymentSatisfied),
								HasAddressAssociatedAll:         fmt.Sprint(placeOrderResponse.Orders[0].Indicators.HasAddressAssociatedAll),
								HasPaypalTenderEnabled:          fmt.Sprint(placeOrderResponse.Orders[0].Indicators.HasPaypalTenderEnabled),
								HasApplepayTenderEnabled:        fmt.Sprint(placeOrderResponse.Orders[0].Indicators.HasApplepayTenderEnabled),
								HasGiftcardTenderEnabled:        fmt.Sprint(placeOrderResponse.Orders[0].Indicators.HasGiftcardTenderEnabled),
								HasThirdpartyTenderEnabled:      fmt.Sprint(placeOrderResponse.Orders[0].Indicators.HasThirdpartyTenderEnabled),
								HasTargetTenderEnabled:          fmt.Sprint(placeOrderResponse.Orders[0].Indicators.HasTargetTenderEnabled),
								HasTargetDebitCardTenderEnabled: fmt.Sprint(placeOrderResponse.Orders[0].Indicators.HasTargetDebitCardTenderEnabled),
							},
							Cartitemsquantity: fmt.Sprint(placeOrderResponse.Orders[0].Summary.ItemsQuantity),
							ReferenceID:       placeOrderResponse.Orders[0].ReferenceID,
							CartState:         "PENDING",
							GuestType:         "REGISTERED",
							Tealeafakasid:     tealeafAkasID,
							Addtocart:         Addtocart{},
							Converted:         "true",
						},
						Client: Client{
							User: User{
								ID: vID,
							},
						},
						Event: Event{
							Action: "checkout_review.checkout_place_order_success",
						},
						Labels: Labels{
							Application: "adaptive",
							BlossomID:   "CI03024104",
							Cluster:     "prod",
						},
						Packages: Packages{
							BuildVersion: "adp-node|sha=3efc12b1,number=85098,source=client,canary=false,webCluster=prod",
						},
						LogDestination: "pipeline3",
						URL: URL{
							Domain: "https://www.target.com",
							Path:   "/co-review",
						},
						Tgt: Tgt{
							CartID: placeOrderResponse.Orders[0].OrderID,
							Custom: Custom{
								Text: Text{
									Num3: placeOrderResponse.Orders[0].ReferenceID,
									Num4: "PAYMENT_STEP_REDESIGN_ENABLED",
								},
							},
						},
					},
				},
			},
		},
	}
	_, _, err = util.MakeRequest(&util.Request{
		Client: task.Task.Client,
		Method: "POST",
		URL:    TargetCancelMethodEndpoint,
		RawHeaders: http.RawHeader{
			{"user-agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/91.0.4472.124 Safari/537.36"},
			{"x-api-key", "2db5ccdb386d0a40ca853e7c46bcebb16d6d41cc"},
			{"content-type", "application/json"},
			{"accept", "*/*"},
			{"origin", BaseEndpoint},
			{"sec-fetch-site", "same-site"},
			{"sec-fetch-mode", "cors"},
			{"sec-fetch-dest", "empty"},
			{"referer", TargetCancelMethodReferer},
			{"accept-encoding", "gzip, deflate, br"},
			{"accept-language", "en-US,en;q=0.9"},
		},
		RequestBodyStruct: targetCancelMethodRequest,
	})
	if err != nil {
		fmt.Println(err)
	}
}
