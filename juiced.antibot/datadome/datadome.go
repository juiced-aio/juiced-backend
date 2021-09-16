package datadome

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"strings"
	"time"

	"backend.juicedbot.io/juiced.client/http"
	"backend.juicedbot.io/juiced.infrastructure/captcha"
	"backend.juicedbot.io/juiced.infrastructure/entities"
	"backend.juicedbot.io/juiced.infrastructure/enums"
	"backend.juicedbot.io/juiced.infrastructure/util"
	u "backend.juicedbot.io/juiced.sitescripts/util"
)

func HandleDatadomeMonitor(monitor *entities.BaseMonitor, retailer enums.Retailer, baseURL *url.URL, referer, parentURL, domain string, body string) error {
	currentStatus := monitor.Status
	monitor.PublishEvent(enums.WaitingForCaptchaMonitor, enums.MonitorUpdate, nil)
	quit := make(chan bool)
	defer func() {
		quit <- true
	}()

	cancellationToken := &util.CancellationToken{Cancel: false}
	go func() {
		for {
			select {
			case <-quit:
				return
			default:
				needToStop := monitor.CheckForStop()
				if needToStop {
					cancellationToken.Cancel = true
					return
				}
			}
			time.Sleep(util.MS_TO_WAIT)
		}
	}()

	datadomeStr, err := util.FindInString(body, "<script>var dd=", "}")
	if err != nil {
		return err
	}
	datadomeStr += "}"
	datadomeStr = strings.ReplaceAll(datadomeStr, "'", "\"")

	datadomeInfo := DatadomeInfo{}
	err = json.Unmarshal([]byte(datadomeStr), &datadomeInfo)
	if err != nil {
		return err
	}

	cookies := monitor.Client.Jar.Cookies(baseURL)
	for _, cookie := range cookies {
		if cookie.Name == "datadome" {
			datadomeInfo.CID = cookie.Value
		}
	}

	if datadomeInfo.CID == "" {
		return errors.New("datadome cookie doesn't exist")
	}

	err = SetDatadomeCookie(datadomeInfo, monitor.Client, monitor.Proxy, retailer, baseURL, referer, parentURL, domain, cancellationToken)
	if err == nil {
		monitor.PublishEvent(currentStatus, enums.MonitorUpdate, nil)
	}
	return err
}

func HandleDatadomeTask(task *entities.BaseTask, retailer enums.Retailer, baseURL *url.URL, referer, parentURL, domain string, body string) error {
	currentStatus := task.Status
	task.PublishEvent(enums.WaitingForCaptcha, 0, enums.TaskUpdate)
	quit := make(chan bool)
	defer func() {
		quit <- true
	}()

	cancellationToken := &util.CancellationToken{Cancel: false}
	go func() {
		for {
			select {
			case <-quit:
				return
			default:
				needToStop := task.CheckForStop()
				if needToStop {
					cancellationToken.Cancel = true
					return
				}
			}
			time.Sleep(util.MS_TO_WAIT)
		}
	}()

	datadomeStr, err := util.FindInString(body, "<script>var dd=", "}")
	if err != nil {
		return err
	}
	datadomeStr += "}"
	datadomeStr = strings.ReplaceAll(datadomeStr, "'", "\"")

	datadomeInfo := DatadomeInfo{}
	err = json.Unmarshal([]byte(datadomeStr), &datadomeInfo)
	if err != nil {
		return err
	}

	cookies := task.Client.Jar.Cookies(baseURL)
	for _, cookie := range cookies {
		if cookie.Name == "datadome" {
			datadomeInfo.CID = cookie.Value
		}
	}

	if datadomeInfo.CID == "" {
		return errors.New("datadome cookie doesn't exist")
	}

	err = SetDatadomeCookie(datadomeInfo, task.Client, task.Proxy, retailer, baseURL, referer, parentURL, domain, cancellationToken)
	if err == nil {
		task.PublishEvent(currentStatus, 0, enums.TaskUpdate)
	}
	return err
}

func SetDatadomeCookie(datadomeInfo DatadomeInfo, client *http.Client, proxy *entities.Proxy, retailer enums.Retailer, baseURL *url.URL, referer, parentURL, domain string, cancellationToken *util.CancellationToken) error {
	params := util.CreateParams(map[string]string{
		"initialCid": datadomeInfo.InitialCID,
		"hash":       datadomeInfo.Hash,
		"cid":        datadomeInfo.CID,
		"t":          datadomeInfo.T,
		"referer":    referer,
		"s":          fmt.Sprint(datadomeInfo.S),
	})
	proxy_ := entities.Proxy{}
	if proxy != nil && proxy.Host != "localhost" {
		proxy_ = *proxy
	}
	token, err := captcha.RequestCaptchaToken(enums.ReCaptchaV2, retailer, DatadomeEndpoint+params, "", .7, proxy_)
	if err != nil {
		return err
	}
	for token == nil {
		if cancellationToken.Cancel {
			return nil
		}
		token = captcha.PollCaptchaTokens(enums.ReCaptchaV2, retailer, DatadomeEndpoint+params, proxy_)
		time.Sleep(1 * time.Second / 10)
	}

	tokenInfo, ok := token.(entities.ReCaptchaToken)
	if !ok {
		return errors.New("token is not ReCaptchaToken")
	}

	params = util.CreateParams(map[string]string{
		"icid":                 datadomeInfo.InitialCID,
		"hash":                 datadomeInfo.Hash,
		"cid":                  datadomeInfo.CID,
		"t":                    datadomeInfo.T,
		"referer":              referer,
		"s":                    fmt.Sprint(datadomeInfo.S),
		"parent_url":           parentURL,
		"ua":                   "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/90.0.4430.212 Safari/537.36",
		"ccid":                 "null",
		"x-forwarded-for":      "",
		"captchaChallenge":     "",
		"g-recaptcha-response": tokenInfo.Token,
	})

	resp, body, err := u.MakeRequest(&u.Request{
		Client: client,
		Method: "GET",
		URL:    DatadomeChallengeEndpoint + params,
		RawHeaders: [][2]string{
			{"sec-ch-ua", "\" Not A;Brand\";v=\"99\", \"Chromium\";v=\"90\", \"Google Chrome\";v=\"90\""},
			{"accept", "*/*"},
			{"x-requested-with", "XMLHttpRequest"},
			{"sec-ch-ua-mobile", "?0"},
			{"user-agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/90.0.4430.212 Safari/537.36"},
			{"sec-fetch-site", "same-origin"},
			{"sec-fetch-mode", "cors"},
			{"sec-fetch-dest", "empty"},
			{"referer", referer},
			{"accept-encoding", "gzip, deflate, br"},
			{"accept-language", "en-US,en;q=0.9"},
		},
	})
	if err != nil {
		return err
	}

	if resp.StatusCode != 200 {
		return errors.New("not 200: " + fmt.Sprint(resp.StatusCode))
	}

	type DatadomeCookie struct {
		Cookie string `json:"cookie"`
	}

	cookie := DatadomeCookie{}

	err = json.Unmarshal([]byte(body), &cookie)
	if err != nil {
		return err
	}

	if !strings.Contains(cookie.Cookie, "datadome=") ||
		!strings.Contains(cookie.Cookie, "; ") {
		return errors.New("bad cookie: " + cookie.Cookie)
	}
	datadomeCookieValue, err := util.FindInString(cookie.Cookie, "datadome=", "; ")
	if err != nil {
		return err
	}

	datadomeCookie := &http.Cookie{
		Name:     "datadome",
		Value:    datadomeCookieValue,
		Path:     "/",
		Domain:   domain,
		MaxAge:   31536000,
		Secure:   true,
		SameSite: http.SameSiteLaxMode,
	}
	oldCookies := client.Jar.Cookies(baseURL)
	newCookies := []*http.Cookie{}
	for _, oldCookie := range oldCookies {
		if oldCookie.Name != "datadome" {
			newCookies = append(newCookies, oldCookie)
		}
	}
	newCookies = append(newCookies, datadomeCookie)
	client.Jar.SetCookies(baseURL, newCookies)
	return nil
}
