package captcha

import (
	"errors"
	"log"
	"strconv"
	"strings"
	"sync"
	"time"

	"backend.juicedbot.io/juiced.infrastructure/common"
	"backend.juicedbot.io/juiced.infrastructure/common/entities"
	"backend.juicedbot.io/juiced.infrastructure/common/enums"
	"backend.juicedbot.io/juiced.infrastructure/common/events"
	"backend.juicedbot.io/juiced.infrastructure/queries"
	"gitlab.com/aycd-inc/autosolve-clients/autosolve-client-go"
	// Future sitescripts will be imported here
)

// CaptchaStore stores information about valid Captcha tokens
type CaptchaStore struct {
	ReCaptchaV2Tokens     map[enums.Retailer][]*entities.ReCaptchaToken
	ReCaptchaV3Tokens     map[enums.Retailer][]*entities.ReCaptchaToken
	HCaptchaTokens        map[enums.Retailer][]*entities.HCaptchaToken
	GeeTestCaptchaTokens  map[enums.Retailer][]*entities.GeeTestCaptchaToken
	AYCDReCaptchaV2Tokens map[enums.Retailer][]*entities.ReCaptchaToken
	AYCDReCaptchaV3Tokens map[enums.Retailer][]*entities.ReCaptchaToken
	EventBus              *events.EventBus
}

var captchaStore *CaptchaStore

// InitCaptchaStore initializes the singleton instance of the Store
func InitCaptchaStore(eventBus *events.EventBus) {
	captchaStore = &CaptchaStore{
		ReCaptchaV2Tokens:     make(map[enums.Retailer][]*entities.ReCaptchaToken),
		ReCaptchaV3Tokens:     make(map[enums.Retailer][]*entities.ReCaptchaToken),
		HCaptchaTokens:        make(map[enums.Retailer][]*entities.HCaptchaToken),
		GeeTestCaptchaTokens:  make(map[enums.Retailer][]*entities.GeeTestCaptchaToken),
		AYCDReCaptchaV2Tokens: make(map[enums.Retailer][]*entities.ReCaptchaToken),
		AYCDReCaptchaV3Tokens: make(map[enums.Retailer][]*entities.ReCaptchaToken),
		EventBus:              eventBus,
	}
}

// GetCaptchaStore returns the singleton instance of the CaptchaStore
func GetCaptchaStore() *CaptchaStore {
	return captchaStore
}

// RequestCaptchaToken returns a Captcha token from the Store, or requests one if none are available
func RequestCaptchaToken(_requestCaptchaTokenInfo interface{}) (interface{}, error) {
	requestCaptchaTokenInfo := RequestCaptchaTokenInfo{}
	if info, ok := _requestCaptchaTokenInfo.(RequestCaptchaTokenInfo); ok {
		requestCaptchaTokenInfo = info
	} else {
		return _requestCaptchaTokenInfo, errors.New("Interfaced passed is not the correct struct")
	}

	var err error
	switch requestCaptchaTokenInfo.CaptchaType {
	case enums.ReCaptchaV2:
		tokens := captchaStore.ReCaptchaV2Tokens[requestCaptchaTokenInfo.Retailer]
		for index, token := range tokens {
			if token.URL == requestCaptchaTokenInfo.Url && token.Proxy.ID == requestCaptchaTokenInfo.Proxy.ID {
				// If a valid token exists, remove it from the list of tokens and return it
				tokens[len(tokens)-1], tokens[index] = tokens[index], tokens[len(tokens)-1]
				captchaStore.ReCaptchaV2Tokens[requestCaptchaTokenInfo.Retailer] = tokens[:len(tokens)-1]
				return *token, nil
			}
		}
		tokens = captchaStore.AYCDReCaptchaV2Tokens[requestCaptchaTokenInfo.Retailer]
		for index, token := range tokens {
			if token.URL == requestCaptchaTokenInfo.Url && token.Proxy.ID == requestCaptchaTokenInfo.Proxy.ID {
				// If a valid token exists, remove it from the list of tokens and return it
				tokens[len(tokens)-1], tokens[index] = tokens[index], tokens[len(tokens)-1]
				captchaStore.AYCDReCaptchaV2Tokens[requestCaptchaTokenInfo.Retailer] = tokens[:len(tokens)-1]
				return *token, nil
			}
		}
		// Otherwise, request a token
		tempSitekey, ok := enums.ReCaptchaSitekeys[requestCaptchaTokenInfo.Retailer]
		retailerSitekey := ""
		if !ok {
			// If the sitekey cannot be extracted from our list, the sitekey parameter is required
			if len(requestCaptchaTokenInfo.Sitekey) == 0 {
				return nil, errors.New("sitekey is a required parameter for this retailer")
			} else {
				retailerSitekey = requestCaptchaTokenInfo.Sitekey[0]
			}
		} else {
			retailerSitekey = string(tempSitekey)
		}
		if retailerSitekey == "" {
			return nil, errors.New("sitekey is a required parameter for this retailer")
		}
		go RequestReCaptchaV2Token(retailerSitekey, requestCaptchaTokenInfo.Url, requestCaptchaTokenInfo.Proxy, requestCaptchaTokenInfo.Retailer)
	case enums.ReCaptchaV3:
		tokens := captchaStore.ReCaptchaV3Tokens[requestCaptchaTokenInfo.Retailer]
		for index, token := range tokens {
			if token.URL == requestCaptchaTokenInfo.Url && token.Proxy.ID == requestCaptchaTokenInfo.Proxy.ID {
				// If a valid token exists, remove it from the list of tokens and return it
				tokens[len(tokens)-1], tokens[index] = tokens[index], tokens[len(tokens)-1]
				captchaStore.ReCaptchaV3Tokens[requestCaptchaTokenInfo.Retailer] = tokens[:len(tokens)-1]
				return *token, nil
			}
		}
		tokens = captchaStore.AYCDReCaptchaV3Tokens[requestCaptchaTokenInfo.Retailer]
		for index, token := range tokens {
			if token.URL == requestCaptchaTokenInfo.Url && token.Proxy.ID == requestCaptchaTokenInfo.Proxy.ID {
				// If a valid token exists, remove it from the list of tokens and return it
				tokens[len(tokens)-1], tokens[index] = tokens[index], tokens[len(tokens)-1]
				captchaStore.AYCDReCaptchaV3Tokens[requestCaptchaTokenInfo.Retailer] = tokens[:len(tokens)-1]
				return *token, nil
			}
		}
		// Otherwise, request a token
		tempSitekey, ok := enums.ReCaptchaSitekeys[requestCaptchaTokenInfo.Retailer]
		retailerSitekey := ""
		if !ok {
			// If the sitekey cannot be extracted from our list, the sitekey parameter is required
			if len(requestCaptchaTokenInfo.Sitekey) == 0 {
				return nil, errors.New("sitekey is a required parameter for this retailer")
			} else {
				retailerSitekey = requestCaptchaTokenInfo.Sitekey[0]
			}
		} else {
			retailerSitekey = string(tempSitekey)
		}
		if retailerSitekey == "" {
			return nil, errors.New("sitekey is a required parameter for this retailer")
		}
		go RequestReCaptchaV3Token(retailerSitekey, requestCaptchaTokenInfo.Url, requestCaptchaTokenInfo.Action, requestCaptchaTokenInfo.MinScore, requestCaptchaTokenInfo.Proxy, requestCaptchaTokenInfo.Retailer)
	case enums.HCaptcha:
		tokens := captchaStore.HCaptchaTokens[requestCaptchaTokenInfo.Retailer]
		for index, token := range tokens {
			if token.URL == requestCaptchaTokenInfo.Url && token.Proxy.ID == requestCaptchaTokenInfo.Proxy.ID {
				// If a valid token exists, remove it from the list of tokens and return it
				tokens[len(tokens)-1], tokens[index] = tokens[index], tokens[len(tokens)-1]
				captchaStore.HCaptchaTokens[requestCaptchaTokenInfo.Retailer] = tokens[:len(tokens)-1]
				return *token, nil
			}
		}
		// Otherwise, request a token
		tempSitekey, ok := enums.HCaptchaSitekeys[requestCaptchaTokenInfo.Retailer]
		retailerSitekey := ""
		if !ok {
			// If the sitekey cannot be extracted from our list, the sitekey parameter is required
			if len(requestCaptchaTokenInfo.Sitekey) == 0 {
				return nil, errors.New("sitekey is a required parameter for this retailer")
			} else {
				retailerSitekey = requestCaptchaTokenInfo.Sitekey[0]
			}
		} else {
			retailerSitekey = string(tempSitekey)
		}
		if retailerSitekey == "" {
			return nil, errors.New("sitekey is a required parameter for this retailer")
		}
		go RequestHCaptchaToken(retailerSitekey, requestCaptchaTokenInfo.Url, requestCaptchaTokenInfo.Proxy, requestCaptchaTokenInfo.Retailer)
	case enums.GeeTestCaptcha:
		// TODO @silent
	}
	// If none are available, return nil -- the Task requesting a captcha should poll this function frequently until successful
	return nil, err
}

// PollCaptchaTokens returns a Captcha token from the store if one is available
func PollCaptchaTokens(captchaType enums.CaptchaType, retailer enums.Retailer, url string, proxy entities.Proxy) interface{} {
	switch captchaType {
	case enums.ReCaptchaV2:
		tokens := captchaStore.ReCaptchaV2Tokens[retailer]
		for index, token := range tokens {
			if token.URL == url && token.Proxy.ID == proxy.ID {
				// If a valid token exists, remove it from the list of tokens and return it
				tokens[len(tokens)-1], tokens[index] = tokens[index], tokens[len(tokens)-1]
				captchaStore.ReCaptchaV2Tokens[retailer] = tokens[:len(tokens)-1]
				return *token
			}
		}
		tokens = captchaStore.AYCDReCaptchaV2Tokens[retailer]
		for index, token := range tokens {
			if token.URL == url && token.Proxy.ID == proxy.ID {
				// If a valid token exists, remove it from the list of tokens and return it
				tokens[len(tokens)-1], tokens[index] = tokens[index], tokens[len(tokens)-1]
				captchaStore.AYCDReCaptchaV2Tokens[retailer] = tokens[:len(tokens)-1]
				return *token
			}
		}
	case enums.ReCaptchaV3:
		tokens := captchaStore.ReCaptchaV3Tokens[retailer]
		for index, token := range tokens {
			if token.URL == url && token.Proxy.ID == proxy.ID {
				// If a valid token exists, remove it from the list of tokens and return it
				tokens[len(tokens)-1], tokens[index] = tokens[index], tokens[len(tokens)-1]
				captchaStore.ReCaptchaV3Tokens[retailer] = tokens[:len(tokens)-1]
				return *token
			}
		}
		tokens = captchaStore.AYCDReCaptchaV3Tokens[retailer]
		for index, token := range tokens {
			if token.URL == url && token.Proxy.ID == proxy.ID {
				// If a valid token exists, remove it from the list of tokens and return it
				tokens[len(tokens)-1], tokens[index] = tokens[index], tokens[len(tokens)-1]
				captchaStore.AYCDReCaptchaV3Tokens[retailer] = tokens[:len(tokens)-1]
				return *token
			}
		}
	case enums.HCaptcha:
		tokens := captchaStore.HCaptchaTokens[retailer]
		for index, token := range tokens {
			if token.URL == url && token.Proxy.ID == proxy.ID {
				// If a valid token exists, remove it from the list of tokens and return it
				tokens[len(tokens)-1], tokens[index] = tokens[index], tokens[len(tokens)-1]
				captchaStore.HCaptchaTokens[retailer] = tokens[:len(tokens)-1]
				return *token
			}
		}
	case enums.GeeTestCaptcha:
		// TODO @silent
	}
	// If none are available, return nil -- the Task requesting a captcha should poll this function frequently until successful
	return nil
}

// RequestReCaptchaV2Token requests a ReCaptchaV2 token from all available APIs and the frontend
func RequestReCaptchaV2Token(sitekey string, url string, proxy entities.Proxy, retailer enums.Retailer) error {
	settings, err := queries.GetSettings()
	if err != nil {
		return err
	}
	keyError := CaptchaSuccess
	captchaServices, captchaServiceCount := CheckServices(settings)
	wg := sync.WaitGroup{}
	wg.Add(captchaServiceCount)

	for _, service := range captchaServices {
		switch service {
		case enums.AYCD:
			go func() {
				defer wg.Done()
				startWait := time.Now()
				// If AYCD is Connecting or Reconnecting, we can wait for a bit (I have it at 30 seconds before the function says fuck it)
				for aycdStatus == autosolve.Connecting || aycdStatus == autosolve.Reconnecting {
					time.Sleep(1 * time.Second / 10)
					if time.Since(startWait) > 30*time.Second {
						return
					}
				}
				// Once it's either Connected or Disconnected, ensure that it's Connected
				if aycdStatus != autosolve.Connected {
					return
				}

				proxyStr := ""
				if proxy.Host != "" && proxy.Port != "" {
					proxyStr = proxy.Host + ":" + proxy.Port
				}
				if proxy.Username != "" && proxy.Password != "" {
					proxyStr += ":" + proxy.Username + ":" + proxy.Password
				}
				// Because we store tokens based on retailer (allowing any task to use the token as long as the URL and Proxy match), we can just pass the retailer in as the TaskID
				// We also need to differentiate between V2 and V3 tokens, so append that to the retailer
				taskInfo := autosolve.CaptchaTokenRequest{
					TaskId:  retailer + "|V2",
					Url:     url,
					SiteKey: sitekey,
					Version: autosolve.ReCaptchaV2Checkbox, // TODO @silent: Infrastructure for differing between invisible and checkbox
					Proxy:   proxyStr,
					// ProxyRequired: ?, // TODO @silent: Get some more info about this
				}
				autosolve.SendTokenRequest(taskInfo)
			}()
		case settings.TwoCaptchaAPIKey:
			go func() {
				defer wg.Done()
				log.Println("Requesting 2Captcha token")
				endpoint := "https://2captcha.com/in.php?key=" + settings.TwoCaptchaAPIKey + "&method=userrecaptcha&googlekey=" + sitekey + "&pageurl=" + url
				if proxy.Host != "" {
					endpoint += "&proxy=" + common.ProxyCleaner(proxy) + "&proxytype=http"
				}
				var token string
				token, err = TwoCaptchaReq(settings.TwoCaptchaAPIKey, endpoint)
				if err != nil {
					log.Println("Error retrieving ReCaptchaV2 from 2Captcha: " + err.Error())
					if err.Error() == "ERROR_KEY_DOES_NOT_EXIST" || err.Error() == "ERROR_WRONG_USER_KEY" {
						keyError = BadTwoCapKeyError
					} else if err.Error() == "ERROR_ZERO_BALANCE" {
						keyError = TwoCapZeroBalanceError
					} else {
						keyError = err.Error()
					}
					return
				}
				captchaStore.ReCaptchaV2Tokens[retailer] = append(captchaStore.ReCaptchaV2Tokens[retailer], &entities.ReCaptchaToken{
					URL:   url,
					Proxy: proxy,
					Token: token,
				})
			}()
		case settings.AntiCaptchaAPIKey:
			go func() {
				defer wg.Done()
				log.Println("Requesting AntiCaptcha token")
				port := 0
				proxyType := ""
				if proxy.Port != "" {
					proxyType = "http"
					port, err = strconv.Atoi(proxy.Port)
					if err != nil {
						log.Println(err.Error())
						return
					}
				}
				var antiCaptchaResponse AntiCaptchaResponse
				taskType := "RecaptchaV2Task"
				emptyString := ""
				if proxy.Host == emptyString {
					taskType = "RecaptchaV2TaskProxyless"
				}
				antiCaptchaResponse, err = AntiCaptchaReq(settings.AntiCaptchaAPIKey, AntiCaptchaTaskInfo{
					Type:          taskType,
					Websiteurl:    url,
					Websitekey:    sitekey,
					Proxytype:     proxyType,
					Proxyaddress:  proxy.Host,
					Proxyport:     port,
					Proxylogin:    proxy.Username,
					Proxypassword: proxy.Password,
					Useragent:     "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/90.0.4430.212 Safari/537.36",
				})
				if err != nil {
					log.Println("Error retrieving ReCaptchaV2 from AntiCaptcha: " + err.Error())
					if err.Error() == "ERROR_KEY_DOES_NOT_EXIST" {
						keyError = BadAntiCapKeyError
					} else if err.Error() == "ERROR_ZERO_BALANCE" {
						keyError = AntiCapZeroBalanceError
					} else {
						keyError = err.Error()
					}
					return
				}

				captchaStore.ReCaptchaV2Tokens[retailer] = append(captchaStore.ReCaptchaV2Tokens[retailer], &entities.ReCaptchaToken{
					URL:   url,
					Proxy: proxy,
					Token: antiCaptchaResponse.Solution.GRecaptchaResponse,
				})
			}()
		case settings.CapMonsterAPIKey:
			go func() {
				defer wg.Done()
				log.Println("Requesting CapMonster token")
				port := 0
				proxyType := ""
				if proxy.Port != "" {
					proxyType = "http"
					port, err = strconv.Atoi(proxy.Port)
					if err != nil {
						log.Println(err.Error())
						return
					}
				}
				var capMonsterResponse CapMonsterResponse
				capMonsterResponse, err = CapMonsterReq(settings.CapMonsterAPIKey, CapMonsterTaskInfo{
					Type:          "NoCaptchaTask",
					Websiteurl:    url,
					Websitekey:    sitekey,
					Proxytype:     proxyType,
					Proxyaddress:  proxy.Host,
					Proxyport:     port,
					Proxylogin:    proxy.Username,
					Proxypassword: proxy.Password,
					Useragent:     "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/90.0.4430.212 Safari/537.36",
				})
				if err != nil {
					log.Println("Error retrieving ReCaptchaV2 from CapMonster: " + err.Error())
					if err.Error() == "ERROR_KEY_DOES_NOT_EXIST" {
						keyError = BadCapMonKeyError
					} else if err.Error() == "ERROR_ZERO_BALANCE" {
						keyError = CapMonZeroBalanceError
					} else {
						keyError = err.Error()
					}
					return
				}

				captchaStore.ReCaptchaV2Tokens[retailer] = append(captchaStore.ReCaptchaV2Tokens[retailer], &entities.ReCaptchaToken{
					URL:   url,
					Proxy: proxy,
					Token: capMonsterResponse.Solution.GRecaptchaResponse,
				})
			}()
		}
	}
	wg.Wait()
	if err != nil {
		return err
	}
	err = KeyErrors(settings, keyError, enums.ReCaptchaV2)
	return err
}

// TODO @silent: Make changes to match v2 function
// RequestReCaptchaV3Token requests a ReCaptchaV3 token from all available APIs and the frontend
func RequestReCaptchaV3Token(sitekey, action, url string, minScore float32, proxy entities.Proxy, retailer enums.Retailer) error {
	settings, err := queries.GetSettings()
	if err != nil {
		return err
	}

	keyError := CaptchaSuccess
	captchaServices, captchaServiceCount := CheckServices(settings)
	wg := sync.WaitGroup{}
	wg.Add(captchaServiceCount)

	// None of the Captcha services support proxies on ReCaptchaV3
	for _, service := range captchaServices {
		switch service {
		case enums.AYCD:
			go func() {
				defer wg.Done()
				startWait := time.Now()
				// If AYCD is Connecting or Reconnecting, we can wait for a bit (I have it at 30 seconds before the function says fuck it)
				for aycdStatus == autosolve.Connecting || aycdStatus == autosolve.Reconnecting {
					time.Sleep(1 * time.Second / 10)
					if time.Now().Sub(startWait) > 30*time.Second {
						return
					}
				}
				// Once it's either Connected or Disconnected, ensure that it's Connected
				if aycdStatus != autosolve.Connected {
					return
				}

				proxyStr := ""
				if proxy.Host != "" && proxy.Port != "" {
					proxyStr = proxy.Host + ":" + proxy.Port
				}
				if proxy.Username != "" && proxy.Password != "" {
					proxyStr += ":" + proxy.Username + ":" + proxy.Password
				}

				// Because we store tokens based on retailer (allowing any task to use the token as long as the URL and Proxy match), we can just pass the retailer in as the TaskID
				// We also need to differentiate between V2 and V3 tokens, so append that to the retailer
				taskInfo := autosolve.CaptchaTokenRequest{
					TaskId:   retailer + "|V3",
					Url:      url,
					SiteKey:  sitekey,
					Version:  autosolve.ReCaptchaV3,
					Action:   action,
					MinScore: minScore,
					Proxy:    proxyStr,
					// ProxyRequired: ?, // TODO @silent: Get some more info about this
				}
				autosolve.SendTokenRequest(taskInfo)
			}()
		case settings.TwoCaptchaAPIKey:
			go func() {
				defer wg.Done()
				token, err := TwoCaptchaReq(settings.TwoCaptchaAPIKey, "https://2captcha.com/in.php?key="+settings.TwoCaptchaAPIKey+"&method=userrecaptcha&googlekey="+sitekey+"&pageurl="+url+"&version=v3")
				if err != nil {
					log.Println("Error retrieving ReCaptchaV3 from 2Captcha: " + err.Error())
					if err.Error() == "ERROR_KEY_DOES_NOT_EXIST" || err.Error() == "ERROR_WRONG_USER_KEY" {
						keyError = BadTwoCapKeyError
					} else if err.Error() == "ERROR_ZERO_BALANCE" {
						keyError = TwoCapZeroBalanceError
					} else {
						keyError = err.Error()
					}
					return
				}
				captchaStore.ReCaptchaV3Tokens[retailer] = append(captchaStore.ReCaptchaV3Tokens[retailer], &entities.ReCaptchaToken{
					URL:   url,
					Proxy: proxy,
					Token: token,
				})
			}()
		case settings.AntiCaptchaAPIKey:
			go func() {
				defer wg.Done()

				antiCaptchaResponse, err := AntiCaptchaReq(settings.AntiCaptchaAPIKey, AntiCaptchaTaskInfo{
					Type:         "RecaptchaV3TaskProxyless",
					Websiteurl:   url,
					Websitekey:   sitekey,
					MinScore:     0.7,
					IsEnterprise: false,
				})
				if err != nil {
					log.Println("Error retrieving ReCaptchaV3 from AntiCaptcha: " + err.Error())
					if err.Error() == "ERROR_KEY_DOES_NOT_EXIST" {
						keyError = BadAntiCapKeyError
					} else if err.Error() == "ERROR_ZERO_BALANCE" {
						keyError = AntiCapZeroBalanceError
					} else {
						keyError = err.Error()
					}
					return
				}

				captchaStore.ReCaptchaV3Tokens[retailer] = append(captchaStore.ReCaptchaV3Tokens[retailer], &entities.ReCaptchaToken{
					URL:   url,
					Proxy: proxy,
					Token: antiCaptchaResponse.Solution.GRecaptchaResponse,
				})
			}()
		case settings.CapMonsterAPIKey:
			go func() {
				defer wg.Done()

				capMonsterResponse, err := CapMonsterReq(settings.CapMonsterAPIKey, CapMonsterTaskInfo{
					Type:       "RecaptchaV3TaskProxyless",
					Websiteurl: url,
					Websitekey: sitekey,
					MinScore:   0.7,
				})
				if err != nil {
					log.Println("Error retrieving ReCaptchaV3 from CapMonster: " + err.Error())
					if err.Error() == "ERROR_KEY_DOES_NOT_EXIST" {
						keyError = BadCapMonKeyError
					} else if err.Error() == "ERROR_ZERO_BALANCE" {
						keyError = CapMonZeroBalanceError
					} else {
						keyError = err.Error()
					}
					return
				}

				captchaStore.ReCaptchaV3Tokens[retailer] = append(captchaStore.ReCaptchaV3Tokens[retailer], &entities.ReCaptchaToken{
					URL:   url,
					Proxy: proxy,
					Token: capMonsterResponse.Solution.GRecaptchaResponse,
				})
			}()
		}
	}
	wg.Wait()
	err = KeyErrors(settings, keyError, enums.ReCaptchaV3)
	if err != nil {
		return err
	}

	return nil
}

// TODO @silent: Make changes to match v2 function
// RequestHCaptchaToken requests a HCaptcha token from all available APIs and the frontend
func RequestHCaptchaToken(sitekey string, url string, proxy entities.Proxy, retailer enums.Retailer) error {
	settings, err := queries.GetSettings()
	if err != nil {
		return err
	}

	keyError := CaptchaSuccess
	captchaServices, captchaServiceCount := CheckServices(settings)
	wg := sync.WaitGroup{}
	wg.Add(captchaServiceCount)

	for _, service := range captchaServices {
		switch service {
		case enums.AYCD:
			go func() {
				defer wg.Done()
			}()
		case settings.TwoCaptchaAPIKey:
			go func() {
				defer wg.Done()

				token, err := TwoCaptchaReq(settings.TwoCaptchaAPIKey, "https://2captcha.com/in.php?key="+settings.TwoCaptchaAPIKey+"&method=hcaptcha&sitekey="+sitekey+"&pageurl="+url+"&proxy="+common.ProxyCleaner(proxy)+"&proxytype=http")
				if err != nil {
					log.Println("Error retrieving HCaptcha from 2Captcha: " + err.Error())
					if err.Error() == "ERROR_KEY_DOES_NOT_EXIST" || err.Error() == "ERROR_WRONG_USER_KEY" {
						keyError = BadTwoCapKeyError
					} else if err.Error() == "ERROR_ZERO_BALANCE" {
						keyError = TwoCapZeroBalanceError
					} else {
						keyError = err.Error()
					}
					return
				}
				captchaStore.HCaptchaTokens[retailer] = append(captchaStore.HCaptchaTokens[retailer], &entities.HCaptchaToken{
					URL:   url,
					Proxy: proxy,
					Token: token,
				})
			}()
		case settings.AntiCaptchaAPIKey:
			go func() {
				defer wg.Done()
				port, err := strconv.Atoi(proxy.Port)
				if err != nil {
					return
				}
				antiCaptchaResponse, err := AntiCaptchaReq(settings.AntiCaptchaAPIKey, AntiCaptchaTaskInfo{
					Type:          "HCaptchaTask",
					Websiteurl:    url,
					Websitekey:    sitekey,
					Proxytype:     "http",
					Proxyaddress:  proxy.Host,
					Proxyport:     port,
					Proxylogin:    proxy.Username,
					Proxypassword: proxy.Password,
					Useragent:     "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/90.0.4430.212 Safari/537.36",
				})
				if err != nil {
					log.Println("Error retrieving HCaptcha from AntiCaptcha: " + err.Error())
					if err.Error() == "ERROR_KEY_DOES_NOT_EXIST" {
						keyError = BadAntiCapKeyError
					} else if err.Error() == "ERROR_ZERO_BALANCE" {
						keyError = AntiCapZeroBalanceError
					} else {
						keyError = err.Error()
					}
					return
				}

				captchaStore.HCaptchaTokens[retailer] = append(captchaStore.HCaptchaTokens[retailer], &entities.HCaptchaToken{
					URL:   url,
					Proxy: proxy,
					Token: antiCaptchaResponse.Solution.GRecaptchaResponse,
				})
			}()
		case settings.CapMonsterAPIKey:
			go func() {
				defer wg.Done()
				port, err := strconv.Atoi(proxy.Port)
				if err != nil {
					return
				}
				capMonsterResponse, err := CapMonsterReq(settings.CapMonsterAPIKey, CapMonsterTaskInfo{
					Type:          "HCaptchaTask",
					Websiteurl:    url,
					Websitekey:    sitekey,
					Proxytype:     "http",
					Proxyaddress:  proxy.Host,
					Proxyport:     port,
					Proxylogin:    proxy.Username,
					Proxypassword: proxy.Password,
					Useragent:     "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/90.0.4430.212 Safari/537.36",
				})
				if err != nil {
					log.Println("Error retrieving HCaptcha from CapMonster: " + err.Error())
					if err.Error() == "ERROR_KEY_DOES_NOT_EXIST" {
						keyError = BadCapMonKeyError
					} else if err.Error() == "ERROR_ZERO_BALANCE" {
						keyError = CapMonZeroBalanceError
					} else {
						keyError = err.Error()
					}
					return
				}

				captchaStore.HCaptchaTokens[retailer] = append(captchaStore.HCaptchaTokens[retailer], &entities.HCaptchaToken{
					URL:   url,
					Proxy: proxy,
					Token: capMonsterResponse.Solution.GRecaptchaResponse,
				})
			}()
		}
	}
	wg.Wait()
	err = KeyErrors(settings, keyError, enums.HCaptcha)
	if err != nil {
		return err
	}

	return nil
}

// RequestGeeTestCaptchaToken requests a GeeTestCaptcha token from all available APIs and the frontend
func RequestGeeTestCaptchaToken(sitekey string, url string, challenge string, proxy entities.Proxy, retailer enums.Retailer) error {
	settings, err := queries.GetSettings()
	if err != nil {
		return err
	}

	keyError := CaptchaSuccess
	captchaServices, captchaServiceCount := CheckServices(settings)
	wg := sync.WaitGroup{}
	wg.Add(captchaServiceCount)

	for _, service := range captchaServices {
		switch service {
		case enums.AYCD:
			go func() {
				defer wg.Done()
			}()
		case settings.TwoCaptchaAPIKey:
			go func() {
				defer wg.Done()
				valuesJoined, err := TwoCaptchaReq(settings.TwoCaptchaAPIKey, "https://2captcha.com/in.php?key="+settings.TwoCaptchaAPIKey+"&method=geetest&gt="+sitekey+"&challenge="+challenge+"&pageurl="+url+"&version=v3")
				if err != nil {
					if err.Error() == "ERROR_KEY_DOES_NOT_EXIST" || err.Error() == "ERROR_WRONG_USER_KEY" {
						keyError = BadTwoCapKeyError
					} else if err.Error() == "ERROR_ZERO_BALANCE" {
						keyError = TwoCapZeroBalanceError
					} else {
						keyError = err.Error()
					}
					return
				}
				valuesSplit := strings.Split(valuesJoined, "|")
				captchaStore.GeeTestCaptchaTokens[retailer] = append(captchaStore.GeeTestCaptchaTokens[retailer], &entities.GeeTestCaptchaToken{
					URL:   url,
					Proxy: proxy,
					Token: entities.GeeTestCaptchaTokenValues{
						Challenge: valuesSplit[0],
						Vaildate:  valuesSplit[1],
						SecCode:   valuesSplit[2],
					},
				})
			}()
		case settings.AntiCaptchaAPIKey:
			go func() {
				defer wg.Done()
				port, err := strconv.Atoi(proxy.Port)
				if err != nil {
					return
				}
				antiCaptchaResponse, err := AntiCaptchaReq(settings.AntiCaptchaAPIKey, AntiCaptchaTaskInfo{
					Type:          "GeeTestTask",
					Websiteurl:    url,
					Gt:            sitekey,
					Challenge:     challenge,
					Proxytype:     "http",
					Proxyaddress:  proxy.Host,
					Proxyport:     port,
					Proxylogin:    proxy.Username,
					Proxypassword: proxy.Password,
					Useragent:     "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/90.0.4430.212 Safari/537.36",
				})
				if err != nil {
					if err.Error() == "ERROR_KEY_DOES_NOT_EXIST" {
						keyError = BadAntiCapKeyError
					} else if err.Error() == "ERROR_ZERO_BALANCE" {
						keyError = AntiCapZeroBalanceError
					} else {
						keyError = err.Error()
					}
					return
				}

				captchaStore.GeeTestCaptchaTokens[retailer] = append(captchaStore.GeeTestCaptchaTokens[retailer], &entities.GeeTestCaptchaToken{
					URL:   url,
					Proxy: proxy,
					Token: entities.GeeTestCaptchaTokenValues{
						Challenge: antiCaptchaResponse.Solution.Challenge,
						Vaildate:  antiCaptchaResponse.Solution.Validate,
						SecCode:   antiCaptchaResponse.Solution.Seccode,
					},
				})
			}()
		// CapMonster does not currently support GeeTest
		case settings.CapMonsterAPIKey:
			go func() { wg.Done() }()
		}
	}
	wg.Wait()
	err = KeyErrors(settings, keyError, enums.GeeTestCaptcha)
	if err != nil {
		return err
	}

	return nil
}
