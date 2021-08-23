package amazon

import (
	"context"
	"errors"
	"fmt"
	"log"
	"math/rand"
	"net/url"
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
	"github.com/go-rod/rod"
	"github.com/go-rod/rod/lib/launcher"
	"github.com/go-rod/rod/lib/launcher/flags"
	"github.com/go-rod/rod/lib/proto"
	"github.com/go-rod/stealth"
	cmap "github.com/orcaman/concurrent-map"
)

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

var AmazonAccountStore = cmap.New()

// CreateAmazonTask takes a Task entity and turns it into a Amazon Task
func CreateAmazonTask(task *entities.Task, profile entities.Profile, proxyGroup *entities.ProxyGroup, eventBus *events.EventBus, loginType enums.LoginType, email, password string) (Task, error) {
	amazonTask := Task{}

	amazonTask = Task{
		Task: base.Task{
			Task:       task,
			Profile:    profile,
			ProxyGroup: proxyGroup,
			EventBus:   eventBus,
		},
		AccountInfo: AccountInfo{
			Email:     email,
			Password:  password,
			LoginType: loginType,
		},
	}
	if proxyGroup != nil {
		amazonTask.Task.Proxy = util.RandomLeastUsedProxy(proxyGroup.Proxies)
	}
	return amazonTask, nil
}

func (task *Task) RunTask() {
	// If the function panics due to a runtime error, recover from it
	defer func() {
		recover()
		// TODO @silent: Let the UI know that a task failed
	}()

	if task.Task.Task.TaskDelay == 0 {
		task.Task.Task.TaskDelay = 2000
	}
	if task.Task.Task.TaskQty == 0 {
		task.Task.Task.TaskQty = 1
	}

	err := task.Task.CreateClient(task.Task.Proxy)
	if err != nil {
		return
	}

	// 1. Setup task
	task.PublishEvent(enums.SettingUp, enums.TaskStart)
	setup := task.Setup()
	if setup {
		return
	}

	// Adding the account to the pool
	var accounts = []Acc{{task.Task.Task.TaskGroupID, task.Task.Client, task.AccountInfo}}
	oldAccounts, _ := AccountPool.Get(task.Task.Task.TaskGroupID)
	if oldAccounts != nil {
		accounts = append(accounts, oldAccounts.([]Acc)...)
	}
	AccountPool.Set(task.Task.Task.TaskGroupID, accounts)

	task.PublishEvent(enums.WaitingForMonitor, enums.TaskUpdate)
	// 2. WaitForMonitor
	needToStop := task.WaitForMonitor()
	if needToStop {
		return
	}

	err = task.Task.UpdateProxy(task.Task.Proxy)
	if err != nil {
		task.PublishEvent(enums.TaskIdle, enums.TaskFail)
		return
	}
	status := enums.OrderStatusFailed
	if task.StockData.MonitorType == enums.SlowSKUMonitor {
		task.PublishEvent(enums.AddingToCart, enums.TaskUpdate)
		// 3. AddToCart
		addedToCart := false
		var retries int
		for !addedToCart {
			needToStop := task.CheckForStop()
			if needToStop || retries > 5 {
				return
			}
			addedToCart = task.AddToCart()
			if !addedToCart {
				retries++
				time.Sleep(time.Duration(task.Task.Task.TaskDelay) * time.Millisecond)
			}
		}
	} else {
		task.Task.Client = task.StockData.Client
	}

	startTime := time.Now()
	task.PublishEvent(enums.CheckingOut, enums.TaskUpdate)

	// 4. PlaceOrder
	placedOrder := false

	for !placedOrder {
		var retries int
		needToStop := task.CheckForStop()
		if needToStop {
			return
		}
		if status == enums.OrderStatusDeclined || retries > 5 {
			break
		}
		placedOrder, status = task.PlaceOrder(startTime)
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
		task.PublishEvent(enums.CardDeclined, enums.TaskComplete)
	case enums.OrderStatusFailed:
		task.PublishEvent(enums.CheckoutFailed, enums.TaskComplete)
	}
}

// Sets the client up by either logging in or waiting for another task to login that is using the same account
func (task *Task) Setup() bool {
	// Bad but quick solution to the multiple logins
	time.Sleep(time.Duration(rand.Intn(1000)) * time.Millisecond)
	if AmazonAccountStore.Has(task.AccountInfo.Email) {
		inMap := true
		for inMap {
			needToStop := task.CheckForStop()
			if needToStop {
				return true
			}
			// Error will be nil unless the item isn't in the map which we are already checking above
			value, ok := AmazonAccountStore.Get(task.AccountInfo.Email)
			if ok {
				acc, ok := value.(Acc)
				if ok {
					if len(task.Task.Client.Jar.Cookies(baseURL)) == 0 {
						task.Task.Client.Jar = acc.Client.Jar
					}
					task.AccountInfo = acc.AccountInfo
					break
				} else {
					if task.Task.Task.TaskStatus != enums.WaitingForLogin {
						task.PublishEvent(enums.WaitingForLogin, enums.TaskUpdate)
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
		task.PublishEvent(enums.LoggingIn, enums.TaskUpdate)
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

	}

	return false
}

// Logs in based on what LoginType the user chooses
func (task *Task) Login() bool {
	AmazonAccountStore.Set(task.AccountInfo.Email, false)
	switch task.AccountInfo.LoginType {
	case enums.LoginTypeBROWSER:
		return task.browserLogin()
	case enums.LoginTypeREQUESTS:
		return task.requestsLogin()
	default:
		return false
	}
}

// Browser login using Rod
func (task *Task) browserLogin() bool {
	cookies := make([]*http.Cookie, 0)

	var userPassProxy bool
	var username string
	var password string

	launcher_ := launcher.New()

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

	ctx, cancel := context.WithCancel(context.Background())
	browserWithCancel := browser.Context(ctx)

	go func() {
		// Wait until either the StopFlag is set to true or the BrowserComplete flag is set to true
		for !task.Task.StopFlag && !task.BrowserComplete {
			time.Sleep(common.MS_TO_WAIT)
		}
		// If the StopFlag being set to true is the one that caused us to break out of that for loop, then the browser is still running, so call cancel()
		if task.Task.StopFlag {
			AmazonAccountStore.Remove(task.AccountInfo.Email)
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
	page.MustNavigate(LoginEndpoint)
	page.MustWaitLoad()
	page.MustElement("#ap_email").MustWaitVisible().Input(task.AccountInfo.Email)
	time.Sleep(2 * time.Second)
	page.MustElement("#continue").MustWaitVisible().MustClick()
	page.MustElement("#ap_password").MustWaitVisible().Input(task.AccountInfo.Password)
	page.MustElementX(`//input[@name="rememberMe"]`).MustWaitVisible().MustClick()
	time.Sleep(2 * time.Second)
	page.MustElement("#signInSubmit").MustWaitVisible().MustClick()
	fmt.Println("Accept 2fa")
	page.MustElement("#auth-cnep-done-button").MustWaitVisible().MustClick()
	page.MustWaitLoad()
	page.MustNavigate(BaseEndpoint)
	page.MustNavigate(TestItemEndpoint)
	page.MustWaitLoad()
	body, err := page.HTML()
	if err != nil {
		AmazonAccountStore.Remove(task.AccountInfo.Email)
		return false
	}

	doc := soup.HTMLParse(body)
	item := doc.Find("input", "name", "dropdown-selection")
	var addressID string
	if item.Error == nil {
		addressID = item.Attrs()["value"]
	}

	if addressID == "" {
		AmazonAccountStore.Remove(task.AccountInfo.Email)
		return false
	}

	sid, err := util.FindInString(body, `ue_sid = '`, `'`)
	if err != nil {
		AmazonAccountStore.Remove(task.AccountInfo.Email)
		return false
	}

	amzCookies, err := page.Cookies([]string{BaseEndpoint})
	if err != nil {
		AmazonAccountStore.Remove(task.AccountInfo.Email)
		return false
	}
	for _, amzCookie := range amzCookies {
		httpCookie := &http.Cookie{
			Name:   amzCookie.Name,
			Value:  amzCookie.Value,
			Domain: amzCookie.Domain,
			Path:   amzCookie.Path,
		}
		amzCookie.Value = strings.ReplaceAll(amzCookie.Value, `"`, `"`)
		if !strings.Contains(amzCookie.Value, `"`) {
			cookies = append(cookies, httpCookie)

		}
	}
	task.Task.Client.Jar.SetCookies(baseURL, cookies)
	acc := Acc{
		GroupID: task.Task.Task.TaskGroupID,
		Client:  task.Task.Client,
		AccountInfo: AccountInfo{
			Email:          task.AccountInfo.Email,
			Password:       task.AccountInfo.Password,
			LoginType:      enums.LoginTypeBROWSER,
			SavedAddressID: addressID,
			SessionID:      sid,
		},
	}
	task.AccountInfo = acc.AccountInfo
	AmazonAccountStore.Set(task.AccountInfo.Email, acc)
	return true
}

// Requests login using !Help's API
// Once we get an api key I will make sure this works, honestly it shouldn't
// work at all. This is how it was from when I first made it and it worked then so.
func (task *Task) requestsLogin() bool {
	_, body, err := util.MakeRequest(&util.Request{
		Client: task.Task.Client,
		Method: "GET",
		URL:    LoginEndpoint,
		RawHeaders: [][2]string{
			{"accept", "text/html,application/xhtml+xml,application/xml;q=0.9,image/avif,image/webp,image/apng,*/*;q=0.8,application/signed-exchange;v=b3;q=0.9"},
			{"accept-encoding", "gzip, deflate, br"},
			{"accept-language", "en-US,en;q=0.9"},
			{"cache-control", "no-cache"},
			{"pragma", "no-cache"},
			{"referer", BaseEndpoint},
			{"sec-ch-ua", `" Not A;Brand";v="99", "Chromium";v="90", "Google Chrome";v="90"`},
			{"sec-ch-ua-mobile", "?0"},
			{"sec-fetch-dest", "document"},
			{"sec-fetch-mode", "navigate"},
			{"sec-fetch-site", "none"},
			{"sec-fetch-user", "?1"},
			{"upgrade-insecure-requests", "1"},
			{"user-agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/90.0.4430.93 Safari/537.36"},
		},
	})
	if err != nil {
		return false
	}

	doc := soup.HTMLParse(body)
	return_To := doc.Find("input", "name", "openid.return_to").Attrs()["value"]

	prevRID := doc.Find("input", "name", "prevRID").Attrs()["value"]

	workflowState := doc.Find("input", "name", "workflowState").Attrs()["value"]

	appActionToken := doc.Find("input", "name", "appActionToken").Attrs()["value"]

	var tempMeta Login

	_, _, err = util.MakeRequest(&util.Request{
		Client:             task.Task.Client,
		Method:             "POST",
		URL:                "https://botbypass.com/metadata_api/metadata1_page_1?email=" + task.AccountInfo.Email + "&passwordLength=" + fmt.Sprint(len(task.AccountInfo.Password)) + "&apiKey=" + MetaData1APIKey,
		ResponseBodyStruct: tempMeta,
	})
	if err != nil {
		return false
	}

	params := common.CreateParams(map[string]string{
		"appActionToken":   string(appActionToken),
		"appAction":        "IGNIN_PWD_COLLECT",
		"subPageType":      "SignInClaimCollect",
		"openid.return_to": string(return_To),
		"prevRID":          string(prevRID),
		"workflowState":    string(workflowState),
		"email":            task.AccountInfo.Email,
		"password":         "",
		"create":           "0",
		"metadata1":        tempMeta.Metadata1,
	})

	_, _, err = util.MakeRequest(&util.Request{
		Client: task.Task.Client,
		Method: "POST",
		URL:    SigninEndpoint,
		RawHeaders: [][2]string{
			{"accept", "text/html,application/xhtml+xml,application/xml;q=0.9,image/avif,image/webp,image/apng,*/*;q=0.8,application/signed-exchange;v=b3;q=0.9"},
			{"accept-encoding", "gzip, deflate, br"},
			{"accept-language", "en-US,en;q=0.9"},
			{"cache-control", "no-cache"},
			{"cookie"},
			{"content-length"},
			{"content-type", "application/x-www-form-urlencoded"},
			{"origin", BaseEndpoint},
			{"pragma", "no-cache"},
			{"referer", LoginEndpoint},
			{"sec-ch-ua", `" Not A;Brand";v="99", "Chromium";v="90", "Google Chrome";v="90"`},
			{"sec-ch-ua-mobile", "?0"},
			{"sec-fetch-dest", "document"},
			{"sec-fetch-mode", "navigate"},
			{"sec-fetch-site", "none"},
			{"sec-fetch-user", "?1"},
			{"upgrade-insecure-requests", "1"},
			{"user-agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/90.0.4430.93 Safari/537.36"},
		},
		Data: []byte(params),
	})
	if err != nil {
		return false
	}

	_, body, err = util.MakeRequest(&util.Request{
		Client: task.Task.Client,
		Method: "GET",
		URL:    TestItemEndpoint,
		RawHeaders: [][2]string{
			{"accept", "text/html,application/xhtml+xml,application/xml;q=0.9,image/avif,image/webp,image/apng,*/*;q=0.8,application/signed-exchange;v=b3;q=0.9"},
			{"accept-encoding", "gzip, deflate, br"},
			{"accept-language", "en-US,en;q=0.9"},
			{"cache-control", "no-cache"},
			{"cookie"},
			{"downlink", "10"},
			{"ect", "4g"},
			{"pragma", "no-cache"},
			{"rtt", "100"},
			{"sec-ch-ua", `" Not A;Brand";v="99", "Chromium";v="90", "Google Chrome";v="90"`},
			{"sec-ch-ua-mobile", "?0"},
			{"sec-fetch-dest", "document"},
			{"sec-fetch-mode", "navigate"},
			{"sec-fetch-site", "none"},
			{"sec-fetch-user", "?1"},
			{"upgrade-insecure-requests", "1"},
			{"user-agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/90.0.4430.93 Safari/537.36"},
		},
	})
	if err != nil {
		return false
	}

	doc = soup.HTMLParse(body)
	task.AccountInfo.SavedAddressID = doc.Find("input", "name", "dropdown-selection").Attrs()["value"]
	if task.AccountInfo.SavedAddressID == "" {
		err := errors.New("no addressID")
		ok := util.HandleErrors(err, util.LoginDetailsError)
		if !ok {
			return false
		}
	}

	task.AccountInfo.SessionID, err = util.FindInString(string(body), `ue_sid = '`, `'`)
	ok := util.HandleErrors(err, util.LoginDetailsError)

	// It does have an effect so I'm not sure why it gives a warning here
	return ok
}

// WaitForMonitor waits until the Monitor has sent the info to the task to continue
func (task *Task) WaitForMonitor() bool {
	for {
		needToStop := task.CheckForStop()
		if needToStop {
			return true
		}
		emptyString := ""
		if task.StockData.OfferID != emptyString {
			return false
		}
		// I see why now
		time.Sleep(common.MS_TO_WAIT)
	}
}

// Takes the task OfferID, ASIN, and SavedAddressID then tries adding that item to the cart
func (task *Task) AddToCart() bool {
	currentEndpoint := AmazonEndpoints[util.RandomNumberInt(0, 2)]
	form := url.Values{
		"isAsync":         {"1"},
		"addressID":       {task.AccountInfo.SavedAddressID},
		"asin.1":          {task.StockData.ASIN},
		"offerListing.1":  {task.StockData.OfferID},
		"quantity.1":      {fmt.Sprint(task.Task.Task.TaskQty)},
		"forcePlaceOrder": {"Place+this+duplicate+order"},
	}

	resp, body, err := util.MakeRequest(&util.Request{
		Client: task.Task.Client,
		Method: "POST",
		URL:    currentEndpoint + "/checkout/turbo-initiate?ref_=dp_start-bbf_1_glance_buyNow_2-1&referrer=detail&pipelineType=turbo&clientId=retailwebsite&weblab=RCX_CHECKOUT_TURBO_DESKTOP_PRIME_87783&temporaryAddToCart=1",
		RawHeaders: [][2]string{
			{"content-length", fmt.Sprint(len(form.Encode()))},
			{"sec-ch-ua", `" Not A;Brand";v="99", "Chromium";v="90", "Google Chrome";v="90"`},
			{"x-amz-checkout-entry-referer-url", currentEndpoint + fmt.Sprintf(MonitorEndpoints[util.RandomNumberInt(0, len(MonitorEndpoints))], task.StockData.ASIN) + util.Randomizer("&pldnSite=1")},
			{"x-amz-turbo-checkout-dp-url", currentEndpoint + fmt.Sprintf(MonitorEndpoints[util.RandomNumberInt(0, len(MonitorEndpoints))], task.StockData.ASIN) + util.Randomizer("&pldnSite=1")},
			{"rtt", "100"},
			{"sec-ch-ua-mobile", "?0"},
			{"user-agent", task.StockData.UA},
			{"content-type", "application/x-www-form-urlencoded"},
			{"x-amz-support-custom-signin", "1"},
			{"accept", "*/*"},
			{"x-requested-with", "XMLHttpRequest"},
			{"x-amz-checkout-csrf-token", task.AccountInfo.SessionID},
			{"downlink", "10"},
			{"ect", "4g"},
			{"origin", BaseEndpoint},
			{"sec-fetch-site", "none"},
			{"sec-fetch-mode", "navigate"},
			{"sec-fetch-dest", "document"},
			{"referer", currentEndpoint + fmt.Sprintf(MonitorEndpoints[util.RandomNumberInt(0, len(MonitorEndpoints))], task.StockData.ASIN) + util.Randomizer("&pldnSite=1")},
			{"accept-encoding", "gzip, deflate, br"},
			{"accept-language", "en-US,en;q=0.9"},
		},
		Data: []byte(form.Encode()),
	})
	if err != nil {
		return false
	}

	switch resp.StatusCode {
	case 200:
		doc := soup.HTMLParse(body)

		err = doc.Find("input", "name", "anti-csrftoken-a2z").Error
		if err != nil {
			return false
		}
		task.StockData.AntiCsrf = doc.Find("input", "name", "anti-csrftoken-a2z").Attrs()["value"]

		task.StockData.PID, err = util.FindInString(body, `currentPurchaseId":"`, `"`)
		if err != nil {
			fmt.Println("Could not find PID")
			return false
		}
		task.StockData.RID, err = util.FindInString(body, `var ue_id = '`, `'`)
		if err != nil {
			fmt.Println("Could not find RID")
			return false
		}

		return true
	case 503:
		fmt.Println("Dogs of Amazon (503)")
		return false
	case 403:
		fmt.Println("SessionID expired")
		return false
	default:
		fmt.Printf("Unkown Code: %v", resp.StatusCode)
		return false

	}
}

// Places the order
func (task *Task) PlaceOrder(startTime time.Time) (bool, enums.OrderStatus) {
	status := enums.OrderStatusFailed
	currentEndpoint := AmazonEndpoints[util.RandomNumberInt(0, 2)]
	form := url.Values{
		"x-amz-checkout-csrf-token": {task.AccountInfo.SessionID},
		"ref_":                      {"chk_spc_placeOrder"},
		"referrer":                  {"spc"},
		"pid":                       {task.StockData.PID},
		"pipelineType":              {"turbo"},
		"clientId":                  {"retailwebsite"},
		"temporaryAddToCart":        {"1"},
		"hostPage":                  {"detail"},
		"weblab":                    {"RCX_CHECKOUT_TURBO_DESKTOP_PRIME_87783"},
		"isClientTimeBased":         {"1"},
		"forcePlaceOrder":           {"Place+this+duplicate+order"},
	}

	resp, _, err := util.MakeRequest(&util.Request{
		Client: task.Task.Client,
		Method: "POST",
		URL:    fmt.Sprintf(CheckoutEndpoint, task.StockData.RID, fmt.Sprint(time.Now().UnixNano())[0:13], task.StockData.PID),
		RawHeaders: [][2]string{
			{"content-length", fmt.Sprint(len(form.Encode()))},
			{"sec-ch-ua", `" Not A;Brand";v="99", "Chromium";v="90", "Google Chrome";v="90"`},
			{"x-amz-checkout-entry-referer-url", currentEndpoint + "/gp/product/" + task.StockData.ASIN + "/ref=crt_ewc_title_oth_4?ie=UTF8&psc=1&smid=ATVPDKIKX0DER"},
			{"anti-csrftoken-a2z", task.StockData.AntiCsrf},
			{"sec-ch-ua-mobile", "?0"},
			{"user-agent", task.StockData.UA},
			{"content-type", "application/x-www-form-urlencoded"},
			{"accept", "*/*"},
			{"x-requested-with", "XMLHttpRequest"},
			{"origin", BaseEndpoint},
			{"sec-fetch-site", "same-origin"},
			{"sec-fetch-mode", "cors"},
			{"sec-fetch-dest", "empty"},
			{"referer", LoginEndpoint},
			{"accept-encoding", "gzip, deflate, br"},
			{"accept-language", "en-US,en;q=0.9"},
		},
		Data: []byte(form.Encode()),
	})
	ok := util.HandleErrors(err, util.RequestDoError)
	if !ok {
		return false, status
	}

	var success bool
	switch resp.StatusCode {
	case 200:
		orderStatus := resp.Header.Get("x-amz-turbo-checkout-page-type")
		switch orderStatus {
		case "thankyou":
			task.PublishEvent(enums.CheckedOut, enums.TaskComplete)
			status = enums.OrderStatusSuccess
			success = true
		default:
			task.PublishEvent(enums.CheckoutFailed, enums.TaskComplete)
			status = enums.OrderStatusFailed
			success = false
		}

	case 503:
		task.PublishEvent(enums.CheckoutFailed, enums.TaskComplete)
		status = enums.OrderStatusFailed
		success = false
	default:
		task.PublishEvent(enums.CheckoutFailed, enums.TaskComplete)
		status = enums.OrderStatusFailed
		success = false
	}

	go util.ProcessCheckout(&util.ProcessCheckoutInfo{
		BaseTask:     task.Task,
		Success:      success,
		Status:       status,
		Content:      "",
		Embeds:       task.CreateAmazonEmbed(status, task.StockData.ImageURL),
		ItemName:     task.StockData.ItemName,
		ImageURL:     task.StockData.ImageURL,
		Sku:          task.StockData.ASIN,
		Retailer:     enums.Amazon,
		Price:        float64(task.StockData.Price),
		Quantity:     1,
		MsToCheckout: time.Since(startTime).Milliseconds(),
	})

	return success, status
}
