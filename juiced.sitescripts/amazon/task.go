package amazon

import (
	"errors"
	"fmt"
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
	"github.com/anaskhan96/soup"
	"github.com/go-rod/rod"
	"github.com/go-rod/rod/lib/launcher"
	"github.com/go-rod/rod/lib/launcher/flags"
	"github.com/go-rod/stealth"
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

// CreateAmazonTask takes a Task entity and turns it into a Amazon Task
func CreateAmazonTask(task *entities.Task, profile entities.Profile, proxy entities.Proxy, eventBus *events.EventBus, loginType enums.LoginType, email, password string) (Task, error) {
	amazonTask := Task{}
	client, err := util.CreateClient(proxy)
	if err != nil {
		return amazonTask, err
	}
	amazonTask = Task{
		Task: base.Task{
			Task:     task,
			Profile:  profile,
			Proxy:    proxy,
			EventBus: eventBus,
			Client:   client,
		},
		AccountInfo: AccountInfo{
			Email:     email,
			Password:  password,
			LoginType: loginType,
		},
	}
	return amazonTask, err
}

func (task *Task) RunTask() {
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
	// Sending the account info to the monitor
	Accounts <- AccChan{task.Task.Task.TaskGroupID, task.Task.Client, task.AccountInfo}

	task.PublishEvent(enums.WaitingForMonitor, enums.TaskUpdate)
	// 2. WaitForMonitor
	needToStop := task.WaitForMonitor()
	if needToStop {
		return
	}

	switch task.CheckoutInfo.MonitorType {
	case enums.SlowSKUMonitor:
		task.PublishEvent(enums.AddingToCart, enums.TaskUpdate)
		// 3. AddToCart
		addedToCart := false
		for !addedToCart {
			var retries int
			needToStop := task.CheckForStop()
			if needToStop {
				return
			}
			addedToCart = task.AddToCart()
			if !addedToCart && retries < 5 {
				retries++
				time.Sleep(time.Duration(task.Task.Task.TaskDelay) * time.Millisecond)
			} else {
				return
			}
		}
		task.PublishEvent(enums.CheckingOut, enums.TaskUpdate)
		// 4. PlaceOrder
		placedOrder := false
		for !placedOrder {
			var retries int
			needToStop := task.CheckForStop()
			if needToStop {
				return
			}
			placedOrder = task.PlaceOrder()
			if !placedOrder && retries < 5 {
				retries++
				time.Sleep(time.Duration(task.Task.Task.TaskDelay) * time.Millisecond)
			} else {
				return
			}
		}
	case enums.FastSKUMonitor:
		task.PublishEvent(enums.CheckingOut, enums.TaskUpdate)
		// 3. PlaceOrder
		placedOrder := false
		for !placedOrder {
			var retries int
			needToStop := task.CheckForStop()
			if needToStop {
				return
			}
			placedOrder = task.PlaceOrder()
			if !placedOrder && retries < 5 {
				retries++
				time.Sleep(time.Duration(task.Task.Task.TaskDelay) * time.Millisecond)
			} else {
				return
			}
		}
	}
}

// Logs in based on what LoginType the user chooses
func (task *Task) Login() bool {
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
		return false
	}

	doc := soup.HTMLParse(body)
	task.AccountInfo.SavedAddressID = doc.Find("input", "name", "dropdown-selection").Attrs()["value"]

	fmt.Println(task.AccountInfo.SavedAddressID)

	task.AccountInfo.SessionID, err = util.FindInString(body, `ue_sid = '`, `'`)
	if err != nil {
		return false
	}
	fmt.Println(task.AccountInfo.SessionID)

	amzCookies, err := page.Cookies([]string{BaseEndpoint})
	if err != nil {
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
	return true
}

// Requests login using !Help's API
// Once we get an api key I will make sure this works, honestly it shouldn't
// work at all. This is how it was from when I first made it and it worked then so.
func (task *Task) requestsLogin() bool {
	resp, err := util.MakeRequest(&util.Request{
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

	defer resp.Body.Close()

	body := util.ReadBody(resp)
	doc := soup.HTMLParse(body)
	return_To := doc.Find("input", "name", "openid.return_to").Attrs()["value"]

	prevRID := doc.Find("input", "name", "prevRID").Attrs()["value"]

	workflowState := doc.Find("input", "name", "workflowState").Attrs()["value"]

	appActionToken := doc.Find("input", "name", "appActionToken").Attrs()["value"]

	var tempMeta Login

	resp, err = util.MakeRequest(&util.Request{
		Client:             task.Task.Client,
		Method:             "POST",
		URL:                "https://botbypass.com/metadata_api/metadata1_page_1?email=" + task.AccountInfo.Email + "&passwordLength=" + fmt.Sprint(len(task.AccountInfo.Password)) + "&apiKey=" + MetaData1APIKey,
		ResponseBodyStruct: tempMeta,
	})
	if err != nil {
		return false
	}

	defer resp.Body.Close()

	params := util.CreateParams(map[string]string{
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

	resp, err = util.MakeRequest(&util.Request{
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

	defer resp.Body.Close()

	resp, err = util.MakeRequest(&util.Request{
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

	defer resp.Body.Close()

	body = util.ReadBody(resp)
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
		if task.CheckoutInfo.MonitorType != emptyString {
			return false
		}
		// @silent: Why sleeping here?
		time.Sleep(time.Duration(task.Task.Task.TaskDelay) * time.Millisecond)
	}
}

// Takes the task OfferID, ASIN, and SavedAddressID then tries adding that item to the cart
func (task *Task) AddToCart() bool {
	currentEndpoint := AmazonEndpoints[util.RandomNumberInt(0, 2)]
	form := url.Values{
		"isAsync":         {"1"},
		"addressID":       {task.AccountInfo.SavedAddressID},
		"asin.1":          {task.TaskInfo.ASIN},
		"offerListing.1":  {task.TaskInfo.OfferID},
		"quantity.1":      {"1"},
		"forcePlaceOrder": {"Place+this+duplicate+order"},
	}

	resp, err := util.MakeRequest(&util.Request{
		Client: task.Task.Client,
		Method: "POST",
		URL:    currentEndpoint + "/checkout/turbo-initiate?ref_=dp_start-bbf_1_glance_buyNow_2-1&referrer=detail&pipelineType=turbo&clientId=retailwebsite&weblab=RCX_CHECKOUT_TURBO_DESKTOP_PRIME_87783&temporaryAddToCart=1",
		RawHeaders: [][2]string{
			{"content-length", fmt.Sprint(len(form.Encode()))},
			{"sec-ch-ua", `" Not A;Brand";v="99", "Chromium";v="90", "Google Chrome";v="90"`},
			{"x-amz-checkout-entry-referer-url", currentEndpoint + fmt.Sprintf(MonitorEndpoints[util.RandomNumberInt(0, len(MonitorEndpoints))], task.TaskInfo.ASIN) + util.Randomizer("&pldnSite=1")},
			{"x-amz-turbo-checkout-dp-url", currentEndpoint + fmt.Sprintf(MonitorEndpoints[util.RandomNumberInt(0, len(MonitorEndpoints))], task.TaskInfo.ASIN) + util.Randomizer("&pldnSite=1")},
			{"rtt", "100"},
			{"sec-ch-ua-mobile", "?0"},
			{"user-agent", task.CheckoutInfo.UA},
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
			{"referer", currentEndpoint + fmt.Sprintf(MonitorEndpoints[util.RandomNumberInt(0, len(MonitorEndpoints))], task.TaskInfo.ASIN) + util.Randomizer("&pldnSite=1")},
			{"accept-encoding", "gzip, deflate, br"},
			{"accept-language", "en-US,en;q=0.9"},
		},
		Data: []byte(form.Encode()),
	})
	if err != nil {
		fmt.Println(err)
	}

	defer resp.Body.Close()

	switch resp.StatusCode {
	case 200:
		body := util.ReadBody(resp)
		doc := soup.HTMLParse(body)

		err := doc.Find("input", "name", "anti-csrftoken-a2z").Error
		if err != nil {
			return false
		}
		task.CheckoutInfo.AntiCsrf = doc.Find("input", "name", "anti-csrftoken-a2z").Attrs()["value"]

		task.CheckoutInfo.PID, err = util.FindInString(body, `currentPurchaseId":"`, `"`)
		if err != nil {
			fmt.Println("Could not find PID")
			return false
		}
		task.CheckoutInfo.RID, err = util.FindInString(body, `var ue_id = '`, `'`)
		if err != nil {
			fmt.Println("Could not find RID")
			return false
		}
		images := doc.FindAll("img")
		for _, source := range images {
			if strings.Contains(source.Attrs()["src"], "https://m.media-amazon.com") {
				task.CheckoutInfo.ImageURL = source.Attrs()["src"]
			}
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
func (task *Task) PlaceOrder() bool {

	currentEndpoint := AmazonEndpoints[util.RandomNumberInt(0, 2)]
	form := url.Values{
		"x-amz-checkout-csrf-token": {task.AccountInfo.SessionID},
		"ref_":                      {"chk_spc_placeOrder"},
		"referrer":                  {"spc"},
		"pid":                       {task.CheckoutInfo.PID},
		"pipelineType":              {"turbo"},
		"clientId":                  {"retailwebsite"},
		"temporaryAddToCart":        {"1"},
		"hostPage":                  {"detail"},
		"weblab":                    {"RCX_CHECKOUT_TURBO_DESKTOP_PRIME_87783"},
		"isClientTimeBased":         {"1"},
		"forcePlaceOrder":           {"Place+this+duplicate+order"},
	}

	resp, err := util.MakeRequest(&util.Request{
		Client: task.Task.Client,
		Method: "POST",
		URL:    fmt.Sprintf(CheckoutEndpoint, task.CheckoutInfo.RID, fmt.Sprint(time.Now().UnixNano())[0:13], task.CheckoutInfo.PID),
		RawHeaders: [][2]string{
			{"content-length", fmt.Sprint(len(form.Encode()))},
			{"sec-ch-ua", `" Not A;Brand";v="99", "Chromium";v="90", "Google Chrome";v="90"`},
			{"x-amz-checkout-entry-referer-url", currentEndpoint + "/gp/product/" + task.TaskInfo.ASIN + "/ref=crt_ewc_title_oth_4?ie=UTF8&psc=1&smid=ATVPDKIKX0DER"},
			{"anti-csrftoken-a2z", task.CheckoutInfo.AntiCsrf},
			{"sec-ch-ua-mobile", "?0"},
			{"user-agent", task.CheckoutInfo.UA},
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
	if err != nil {
		fmt.Println(err)
	}

	defer resp.Body.Close()

	var status enums.OrderStatus
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

	_, user, err := queries.GetUserInfo()
	if err != nil {
		fmt.Println("Could not get user info")
		return false
	}
	sec.DiscordWebhook(success, "", task.CreateAmazonEmbed(status, task.CheckoutInfo.ImageURL), user)

	return success
}
