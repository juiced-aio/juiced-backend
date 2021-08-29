package target

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"math/rand"
	"net/url"
	"runtime/debug"
	"strings"
	"time"

	"backend.juicedbot.io/juiced.client/http/cookiejar"

	"backend.juicedbot.io/juiced.client/http"
	"backend.juicedbot.io/juiced.infrastructure/common"
	"backend.juicedbot.io/juiced.infrastructure/common/entities"
	"backend.juicedbot.io/juiced.infrastructure/common/enums"
	"backend.juicedbot.io/juiced.infrastructure/common/events"
	"backend.juicedbot.io/juiced.sitescripts/base"
	"backend.juicedbot.io/juiced.sitescripts/util"
	"github.com/dgrijalva/jwt-go"
	"github.com/go-rod/rod"
	"github.com/go-rod/rod/lib/launcher"
	"github.com/go-rod/rod/lib/launcher/flags"
	"github.com/go-rod/rod/lib/proto"
	"github.com/go-rod/stealth"
	cmap "github.com/orcaman/concurrent-map"
)

const MAX_RETRIES = 5

var TargetAccountStore = cmap.New()

// CreateTargetTask takes a Task entity and turns it into a Target Task
func CreateTargetTask(task *entities.Task, profile entities.Profile, proxyGroup *entities.ProxyGroup, eventBus *events.EventBus, email, password string, paymentType enums.PaymentType) (Task, error) {
	targetTask := Task{
		Task: base.Task{
			Task:       task,
			Profile:    profile,
			ProxyGroup: proxyGroup,
			EventBus:   eventBus,
		},
		AccountInfo: AccountInfo{
			Email:          email,
			Password:       password,
			PaymentType:    paymentType,
			DefaultCardCVV: profile.CreditCard.CVV,
		},
	}
	if proxyGroup != nil {
		targetTask.Task.Proxy = util.RandomLeastUsedProxy(proxyGroup.Proxies)
	}
	return targetTask, nil
}

var baseURL, _ = url.Parse(BaseEndpoint)

// PublishEvent wraps the EventBus's PublishTaskEvent function
func (task *Task) PublishEvent(status enums.TaskStatus, eventType enums.TaskEventType, statusPercentage int) {
	task.Task.Task.SetTaskStatus(status)
	task.Task.EventBus.PublishTaskEvent(status, statusPercentage, eventType, nil, task.Task.Task.ID)
}

// CheckForStop checks the stop flag and stops the monitor if it's true
func (task *Task) CheckForStop() bool {
	if task.Task.StopFlag {
		task.PublishEvent(enums.TaskIdle, enums.TaskStop, 0)
		return true
	}
	return false
}

// RunTask is the script driver that calls all the individual requests
// Function order:
// 		1. Setup
// 		2. WaitForMonitor
// 		3. AddToCart
// 		4. GetCartInfo
//		5. SetShippingInfo
// 		6. SetPaymentInfo
// 		7. PlaceOrder
func (task *Task) RunTask() {
	// If the function panics due to a runtime error, recover from it
	defer func() {
		if r := recover(); r != nil {
			task.PublishEvent(fmt.Sprintf(enums.TaskFailed, r), enums.TaskFail, 0)
		} else {
			if !strings.Contains(task.Task.Task.TaskStatus, strings.ReplaceAll(enums.TaskIdle, " %s", "")) &&
				!strings.Contains(task.Task.Task.TaskStatus, strings.ReplaceAll(enums.CheckingOutFailure, " %s", "")) &&
				!strings.Contains(task.Task.Task.TaskStatus, strings.ReplaceAll(enums.CardDeclined, " %s", "")) &&
				!strings.Contains(task.Task.Task.TaskStatus, strings.ReplaceAll(enums.CheckedOut, " %s", "")) &&
				!strings.Contains(task.Task.Task.TaskStatus, strings.ReplaceAll(enums.TaskFailed, " %s", "")) {
				task.PublishEvent(enums.TaskIdle, enums.TaskStop, 0)
			}
		}
		task.Task.StopFlag = true
	}()
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

	// 1. Setup task
	task.PublishEvent(enums.SettingUp, enums.TaskStart, 10)
	setup := task.Setup()
	if setup {
		return
	}

	needToStop := task.CheckForStop()
	if needToStop {
		return
	}

	task.PublishEvent(enums.WaitingForMonitor, enums.TaskUpdate, 20)
	// 2. WaitForMonitor
	needToStop = task.WaitForMonitor()
	if needToStop {
		return
	}

	task.PublishEvent(enums.AddingToCart, enums.TaskUpdate, 30)
	// 3. AddtoCart
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

	task.PublishEvent(enums.GettingCartInfo, enums.TaskUpdate, 40)
	startTime := time.Now()
	// 4. GetCartInfo
	gotCartInfo := false
	for !gotCartInfo {
		needToStop := task.CheckForStop()
		if needToStop {
			return
		}
		_, gotCartInfo = task.GetCartInfo()
		if !gotCartInfo {
			time.Sleep(time.Duration(task.Task.Task.TaskDelay) * time.Millisecond)
		}
	}

	task.PublishEvent(enums.SettingShippingInfo, enums.TaskUpdate, 70)
	// 5. SetShippingInfo
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

	task.PublishEvent(enums.SettingBillingInfo, enums.TaskUpdate, 80)
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
	dontRetry := false
	retries := 0
	status := enums.OrderStatusFailed
	for !placedOrder {
		needToStop := task.CheckForStop()
		if needToStop {
			return
		}
		if status == enums.OrderStatusDeclined || retries > MAX_RETRIES || dontRetry {
			break
		}
		placedOrder, status, dontRetry = task.PlaceOrder(startTime, retries)
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
		task.PublishEvent(enums.CheckedOut, enums.TaskComplete, 100)
	case enums.OrderStatusDeclined:
		task.PublishEvent(enums.CardDeclined, enums.TaskComplete, 100)
	case enums.OrderStatusFailed:
		task.PublishEvent(enums.CheckoutFailed, enums.TaskComplete, 100)
	}

}

// Sets the client up by either logging in or waiting for another task to login that is using the same account
func (task *Task) Setup() bool {
	// Bad but quick solution to the multiple logins
	time.Sleep(time.Duration(rand.Intn(1000)) * time.Millisecond)
	if TargetAccountStore.Has(task.AccountInfo.Email) {
		inMap := true
		for inMap {
			needToStop := task.CheckForStop()
			if needToStop {
				return true
			}
			// Error will be nil unless the item isn't in the map which we are already checking above
			value, ok := TargetAccountStore.Get(task.AccountInfo.Email)
			if ok {
				client, isClient := value.(http.Client)
				if isClient {
					if len(task.Task.Client.Jar.Cookies(baseURL)) == 0 {
						task.Task.Client.Jar = client.Jar
					}
					break
				} else {
					if task.Task.Task.TaskStatus != enums.WaitingForLogin {
						task.PublishEvent(enums.WaitingForLogin, enums.TaskUpdate, 15)
					}
					time.Sleep(common.MS_TO_WAIT)
				}
			} else {
				inMap = false
				return task.Setup()
			}
		}
	} else {
		// Login
		task.PublishEvent(enums.LoggingIn, enums.TaskUpdate, 15)
		loggedIn := false
		for !loggedIn {
			needToStop := task.CheckForStop()
			if needToStop {
				return true
			}
			loggedIn = task.Login()
			if !loggedIn {
				time.Sleep(time.Duration(task.Task.Task.TaskDelay) * time.Millisecond)
			}
		}

		// Refresh login in background
		go task.RefreshLogin()

		clearedCart := false
		for !clearedCart {
			needToStop := task.CheckForStop()
			if needToStop {
				return true
			}
			clearedCart = task.ClearCart()
			if !clearedCart {
				time.Sleep(time.Duration(task.Task.Task.TaskDelay) * time.Millisecond)
			}
		}

	}

	return false
}

// Login logs the user in and sets the task's cookies for the logged in user
// TODO @silent: Handle stop flag within Login function
func (task *Task) Login() bool {
	defer func() {
		if r := recover(); r != nil {
			log.Println(string(debug.Stack()))
			TargetAccountStore.Remove(task.AccountInfo.Email)
		}
	}()
	TargetAccountStore.Set(task.AccountInfo.Email, false)
	var userPassProxy bool
	var username string
	var password string
	cookies := make([]*http.Cookie, 0)

	launcher_ := launcher.New()

	fileInfos, err := ioutil.ReadDir(launcher.DefaultBrowserDir)
	if len(fileInfos) == 0 || err != nil {
		task.PublishEvent("Possibly downloading browser. Please wait patiently", enums.TaskUpdate, 15)
	}

	proxyCleaned := ""
	if task.Task.Proxy != nil {
		proxyCleaned = common.ProxyCleaner(*task.Task.Proxy)
	}
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

	u := launcher_.
		Set(flags.Flag("headless")).
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
		Set(flags.Flag("use-mock-keychain")).
		MustLaunch()

	browser := rod.New().ControlURL(u).MustConnect()

	ctx, cancel := context.WithCancel(context.Background())
	browserWithCancel := browser.Context(ctx)

	go func() {
		// Wait until either the StopFlag is set to true or the BrowserComplete flag is set to true
		for !task.Task.StopFlag && !task.BrowserComplete {
			time.Sleep(common.MS_TO_WAIT)
		}
		// If the StopFlag being set to true is the one that caused us to break out of that for loop, then the browser is still running, so call cancel()
		if task.Task.StopFlag {
			TargetAccountStore.Remove(task.AccountInfo.Email)
			browserWithCancel.MustClose()
			cancel()
		}
	}()

	browserWithCancel.MustIgnoreCertErrors(true)

	defer func() { browserWithCancel.MustClose() }()

	if userPassProxy {
		go browserWithCancel.MustHandleAuth(username, password)()
	}

	page := stealth.MustPage(browserWithCancel)
	page.MustSetUserAgent(&proto.NetworkSetUserAgentOverride{UserAgent: "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/92.0.4515.107 Safari/537.36"})
	loginPage := page.MustNavigate(LoginEndpoint)
	if loginPage != nil {
		loginPage.MustWaitLoad()
	} else {
		TargetAccountStore.Remove(task.AccountInfo.Email)
		return false
	}
	if strings.Contains(page.MustHTML(), "accessDenied-CheckVPN") {
		task.PublishEvent("Bad Proxy", enums.TaskFail, 0)
		TargetAccountStore.Remove(task.AccountInfo.Email)
		return false
	}

	usernameBox := page.MustElement("#username").MustWaitVisible()
	usernameBox.MustTap()
	for i := range task.AccountInfo.Email {
		usernameBox.Input(string(task.AccountInfo.Email[i]))
		time.Sleep(125 * time.Millisecond)
	}

	time.Sleep(1 * time.Second / 2)
	passwordBox := page.MustElement("#password").MustWaitVisible()
	passwordBox.MustTap()
	for i := range task.AccountInfo.Password {
		passwordBox.Input(string(task.AccountInfo.Password[i]))
		time.Sleep(125 * time.Millisecond)
	}
	time.Sleep(1 * time.Second / 2)

	checkbox, err := page.ElementX(`//*[contains(@class, 'nds-checkbox')]`)
	if err != nil {
		checkbox, err = page.ElementX(`//*[contains(@class, 'sc-hMqMXs ysAUA')]`)
		if err != nil {
			TargetAccountStore.Remove(task.AccountInfo.Email)
			return false
		}
	}
	checkbox.MustWaitVisible().MustClick()
	page.MustElement("#login").MustWaitVisible().MustClick().MustWaitLoad()

	time.Sleep(1 * time.Second / 2)
	if strings.Contains(page.MustHTML(), "That password is incorrect.") {
		task.PublishEvent("Incorrect password", enums.TaskFail, 0)
		return false
	} else if strings.Contains(page.MustHTML(), "Your account is locked") {
		task.PublishEvent("Account is locked", enums.TaskFail, 0)
		return false
	}
	page.MustElement("#account").MustWaitLoad()
	page.MustWaitLoad()
	page.MustNavigate(BaseEndpoint).MustWaitLoad()
	page.MustElement("#account").MustWaitLoad().MustWaitInteractable().MustClick()
	page.MustElement("#accountNav-signIn").MustWaitVisible().MustClick()
	page.MustWaitLoad()

	browserCookies, _ := page.Cookies([]string{BaseEndpoint})

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
	task.BrowserComplete = true

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

				// Maybe the problem he was having was the claims.Eid != his task.AccountInfo.Email, shouldn't be an issue removing it, oh this is probably the problem everyone was having
				if err != nil {
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
				loggedIn = task.Login()
				if !loggedIn {
					time.Sleep(time.Duration(task.Task.Task.TaskDelay) * time.Millisecond)
				}
			}
		}
	}
}

// ClearCart clears the account's cart before starting the task
func (task *Task) ClearCart() bool {
	cartInfo, ok := task.GetCartInfo()
	if !ok {
		return ok
	}
	for _, cartItem := range cartInfo.CartItems {
		resp, _, err := util.MakeRequest(&util.Request{
			Client: task.Task.Client,
			Method: "DELETE",
			URL:    fmt.Sprintf(ClearCartEndpoint, cartItem.CartItemID),
			RawHeaders: http.RawHeader{
				{"pragma", "no-cache"},
				{"cache-control", "no-cache"},
				{"sec-ch-ua", `Not;A Brand";v="99", "Google Chrome";v="91", "Chromium";v="91"`},
				{"accept", "application/json"},
				{"x-application-name", "web"},
				{"sec-ch-ua-mobile", "?0"},
				{"user-agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/91.0.4472.124 Safari/537.36"},
				{"content-type", "application/json"},
				{"origin", BaseEndpoint},
				{"sec-fetch-site", "same-site"},
				{"sec-fetch-mode", "cors"},
				{"sec-fetch-dest", "empty"},
				{"referer", ClearCartReferer},
				{"accept-encoding", "gzip, deflate, br"},
				{"accept-language", "en-US,en;q=0.9"},
			},
		})
		if err != nil || resp.StatusCode != 200 {
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
		if task.InStockData.TCIN != "" {
			task.TCINType = task.InStockData.TCINType
			task.TCIN = task.InStockData.TCIN
			task.Task.HasStockData = true
			return false
		}
		time.Sleep(common.MS_TO_WAIT)
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
		for _, alert := range addToCartResponse.Alerts {
			if alert.Code == "MAX_PURCHASE_LIMIT_EXCEEDED" {
				return true
			}
		}

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
			task.Task.Client.Jar.SetCookies(resp.Request.URL, newCookies)

		}
		return false
	}

}

// GetCartInfo returns the cart info needed for updating payment and placing an order
func (task *Task) GetCartInfo() (getCartInfoResponse GetCartInfoResponse, _ bool) {
	getCartInfoRequest := GetCartInfoRequest{
		CartType: "REGULAR",
	}

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
		return getCartInfoResponse, false
	}

	switch resp.StatusCode {
	case 201:
		task.AccountInfo.CartInfo = getCartInfoResponse
		if task.AccountInfo.CartID == "" {
			task.AccountInfo.CartID = getCartInfoResponse.CartID
		}
		return getCartInfoResponse, true
	default:
		return getCartInfoResponse, false
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
func (task *Task) SetPaymentInfo() (bool, bool) {
	if !common.ValidCardType([]byte(task.Task.Profile.CreditCard.CardNumber), task.Task.Task.TaskRetailer) {
		return false, true
	}

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
		return false, false
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
		return false, false
	}

	// TODO: Handle various responses
	// Not much to handle here

	if resp.StatusCode != 200 {
		return false, false
	}
	return true, false
}

// PlaceOrder completes the checkout process
func (task *Task) PlaceOrder(startTime time.Time, retries int) (bool, enums.OrderStatus, bool) {
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
		case "MISSING_CREDIT_CARD_CVV":
			success = false
			dontRetry = true
		default:
			status = enums.OrderStatusFailed
			success = false
		}
	}

	if success || status == enums.OrderStatusDeclined || retries >= MAX_RETRIES {
		go util.ProcessCheckout(&util.ProcessCheckoutInfo{
			BaseTask:     task.Task,
			Success:      success,
			Status:       status,
			Content:      "",
			Embeds:       task.CreateTargetEmbed(status, task.AccountInfo.CartInfo.CartItems[0].ItemAttributes.ImagePath),
			ItemName:     task.AccountInfo.CartInfo.CartItems[0].ItemAttributes.Description,
			ImageURL:     task.AccountInfo.CartInfo.CartItems[0].ItemAttributes.ImagePath,
			Sku:          task.TCIN,
			Retailer:     enums.Target,
			Price:        task.AccountInfo.CartInfo.CartItems[0].UnitPrice,
			Quantity:     task.Task.Task.TaskQty,
			MsToCheckout: time.Since(startTime).Milliseconds(),
		})
	}

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
