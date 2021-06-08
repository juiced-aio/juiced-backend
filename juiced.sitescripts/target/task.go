package target

import (
	"encoding/json"
	"fmt"
	"log"
	"net/url"
	"strings"
	"time"

	"backend.juicedbot.io/juiced.client/http"
	"backend.juicedbot.io/juiced.infrastructure/common/entities"
	"backend.juicedbot.io/juiced.infrastructure/common/enums"
	"backend.juicedbot.io/juiced.infrastructure/common/events"
	"backend.juicedbot.io/juiced.infrastructure/queries"
	sec "backend.juicedbot.io/juiced.security/auth/util"
	"backend.juicedbot.io/juiced.sitescripts/base"
	"backend.juicedbot.io/juiced.sitescripts/util"
	"github.com/dgrijalva/jwt-go"
	"github.com/go-rod/rod"
	"github.com/go-rod/rod/lib/launcher"
	"github.com/go-rod/rod/lib/launcher/flags"
	"github.com/go-rod/stealth"
)

// TODO @silent: Handle proxies
// TODO @silent: Handle errors
// TODO @silent: Allow multiple tasks to use cookies from one account
// TODO @silent: Code for steps is repetitive, should abstract
// TODO @Humphreyyyy: Handle mid-checkout sellouts (at some point)
// 		TODO @silent: Mid-checkout sellout errors may have to propagate back up to the monitor

// CreateTargetTask takes a Task entity and turns it into a Target Task
func CreateTargetTask(task *entities.Task, profile entities.Profile, proxy entities.Proxy, eventBus *events.EventBus, checkoutType enums.CheckoutType, email, password string, paymentType enums.PaymentType) (Task, error) {
	targetTask := Task{}
	client, err := util.CreateClient(proxy)
	if err != nil {
		return targetTask, err
	}
	targetTask = Task{
		Task: base.Task{
			Task:     task,
			Profile:  profile,
			Proxy:    proxy,
			EventBus: eventBus,
			Client:   client,
		},
		CheckoutType: checkoutType,
		AccountInfo: AccountInfo{
			Email:          email,
			Password:       password,
			PaymentType:    paymentType,
			DefaultCardCVV: profile.CreditCard.CVV,
		},
	}
	return targetTask, err
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
		recover()
		// TODO @silent: Let the UI know that a task failed
	}()

	task.PublishEvent(enums.LoggingIn, enums.TaskStart)
	// 1. Login
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

	task.PublishEvent(enums.WaitingForMonitor, enums.TaskUpdate)
	// 3. WaitForMonitor
	needToStop := task.WaitForMonitor()
	if needToStop {
		return
	}

	startTime := time.Now()

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
	var status enums.OrderStatus
	for !placedOrder {
		needToStop := task.CheckForStop()
		if needToStop || status == enums.OrderStatusDeclined {
			return
		}
		status, placedOrder = task.PlaceOrder()
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

// Login logs the user in and sets the task's cookies for the logged in user
// TODO @silent: Handle stop flag within Login function
func (task *Task) Login() bool {
	cookies := make([]*http.Cookie, 0)

	u := launcher.New().
		Set(flags.Flag("headless")).
		// Delete(flags.Flag("--headless")).
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
		Set(flags.Flag("use-mock-keychain")).
		MustLaunch()

	browser := rod.New().ControlURL(u).MustConnect()

	defer browser.MustClose()

	page := stealth.MustPage(browser)
	page.MustNavigate(LoginEndpoint)
	page.MustElement("#username").MustWaitVisible().Input(task.AccountInfo.Email)
	page.MustElement("#password").MustWaitVisible().Input(task.AccountInfo.Password)
	page.MustElementX(`//*[contains(@class, 'sc-hMqMXs ysAUA')]`).MustWaitVisible().MustClick()
	page.MustElement("#login").MustWaitVisible().MustClick()
	page.MustElement("#account").MustWaitVisible().MustClick()
	page.MustElement("#accountNav-signIn").MustWaitVisible().MustClick()
	page.MustWaitLoad()
	page.MustNavigate(BaseEndpoint)
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
	return true
}

// RefreshLogin refreshes the login tokens so that the user can remain logged in
func (task *Task) RefreshLogin() {
	// If the function panics due to a runtime error, recover and restart it
	defer func() {
		recover()
		task.RefreshLogin()
	}()

	for {
		success := true
		if task.AccountInfo.Refresh == 0 || time.Now().Unix() > task.AccountInfo.Refresh {
			refreshLoginResponse := RefreshLoginResponse{}

			resp, _, err := util.MakeRequest(&util.Request{
				Client:             task.Task.Client,
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
			case 200:
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
		if task.TCIN != "" {
			return false
		}
	}
}

// AddToCart sends a post request to the AddToCartEndpoint with a body determined by the CheckoutType
func (task *Task) AddToCart() bool {
	var data []byte
	var err error
	tcinWithType := strings.Split(task.TCIN, "|")

	task.TCINType = tcinWithType[1]
	task.TCIN = tcinWithType[0]

	shipReq, err := json.Marshal(AddToCartPickupRequest{
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
	ok := util.HandleErrors(err, util.RequestMarshalBodyError)
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
	case 200:
		task.AccountInfo.CartID = addToCartResponse.CartID
		return true
	default:
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
	case 200:
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
	endpoint := SetPaymentInfoNEWEndpoint
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
func (task *Task) PlaceOrder() (enums.OrderStatus, bool) {
	var status enums.OrderStatus
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
		return status, false
	}

	// TODO: Handle various responses
	log.Println(placeOrderResponse.Message)
	var success bool
	switch resp.StatusCode {
	case 200:
		status = enums.OrderStatusSuccess
		success = true
	default:
		switch placeOrderResponse.Code {
		case "PAYMENT_DECLINED_EXCEPTION":
			status = enums.OrderStatusDeclined
			success = false
		default:
			status = enums.OrderStatusFailed
			success = false
		}
	}
	_, user, err := queries.GetUserInfo()
	if err != nil {
		fmt.Println("Could not get user info")
		return status, success
	}
	sec.DiscordWebhook(success, "", task.CreateTargetEmbed(status, task.AccountInfo.CartInfo.CartItems[0].ItemAttributes.ImagePath), user)
	return status, success
}
