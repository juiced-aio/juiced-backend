package captcha

import (
	"errors"
	"log"
	"strconv"
	"strings"
	"sync"

	"backend.juicedbot.io/juiced.infrastructure/common"
	"backend.juicedbot.io/juiced.infrastructure/common/entities"
	"backend.juicedbot.io/juiced.infrastructure/common/enums"
	"backend.juicedbot.io/juiced.infrastructure/common/events"
	"backend.juicedbot.io/juiced.infrastructure/queries"
	// Future sitescripts will be imported here
)

// CaptchaStore stores information about valid Captcha tokens
type CaptchaStore struct {
	ReCaptchaV2Tokens    map[enums.Retailer][]*entities.ReCaptchaToken
	ReCaptchaV3Tokens    map[enums.Retailer][]*entities.ReCaptchaToken
	HCaptchaTokens       map[enums.Retailer][]*entities.HCaptchaToken
	GeeTestCaptchaTokens map[enums.Retailer][]*entities.GeeTestCaptchaToken
	EventBus             *events.EventBus
}

var captchaStore *CaptchaStore

// InitCaptchaStore initializes the singleton instance of the Store
func InitCaptchaStore(eventBus *events.EventBus) {
	captchaStore = &CaptchaStore{
		ReCaptchaV2Tokens:    make(map[enums.Retailer][]*entities.ReCaptchaToken),
		ReCaptchaV3Tokens:    make(map[enums.Retailer][]*entities.ReCaptchaToken),
		HCaptchaTokens:       make(map[enums.Retailer][]*entities.HCaptchaToken),
		GeeTestCaptchaTokens: make(map[enums.Retailer][]*entities.GeeTestCaptchaToken),
		EventBus:             eventBus,
	}
}

// GetCaptchaStore returns the singleton instance of the CaptchaStore
func GetCaptchaStore() *CaptchaStore {
	return captchaStore
}

// RequestCaptchaToken returns a Captcha token from the Store, or requests one if none are available
func RequestCaptchaToken(captchaType enums.CaptchaType, retailer enums.Retailer, url string, proxy entities.Proxy, sitekey ...string) (interface{}, error) {
	var err error
	switch captchaType {
	case enums.ReCaptchaV2:
		tokens := captchaStore.ReCaptchaV2Tokens[retailer]
		for index, token := range tokens {
			if token.URL == url && token.Proxy.ID == proxy.ID {
				// If a valid token exists, remove it from the list of tokens and return it
				tokens[len(tokens)-1], tokens[index] = tokens[index], tokens[len(tokens)-1]
				captchaStore.ReCaptchaV2Tokens[retailer] = tokens[:len(tokens)-1]
				return *token, nil
			}
		}
		log.Println(sitekey)
		// Otherwise, request a token
		tempSitekey, ok := enums.ReCaptchaSitekeys[retailer]
		log.Println(tempSitekey)
		retailerSitekey := ""
		if !ok {
			// If the sitekey cannot be extracted from our list, the sitekey parameter is required
			if len(sitekey) == 0 {
				return nil, errors.New("sitekey is a required parameter for this retailer")
			} else {
				log.Println(sitekey[0])
				retailerSitekey = sitekey[0]
			}
		} else {
			retailerSitekey = string(tempSitekey)
		}
		log.Println(retailerSitekey)
		if retailerSitekey == "" {
			return nil, errors.New("sitekey is a required parameter for this retailer")
		}
		err = RequestReCaptchaV2Token(retailerSitekey, url, proxy, retailer)
	case enums.ReCaptchaV3:
		tokens := captchaStore.ReCaptchaV3Tokens[retailer]
		for index, token := range tokens {
			if token.URL == url && token.Proxy.ID == proxy.ID {
				// If a valid token exists, remove it from the list of tokens and return it
				tokens[len(tokens)-1], tokens[index] = tokens[index], tokens[len(tokens)-1]
				captchaStore.ReCaptchaV3Tokens[retailer] = tokens[:len(tokens)-1]
				return *token, nil
			}
		}
		// Otherwise, request a token
		tempSitekey, ok := enums.ReCaptchaSitekeys[retailer]
		retailerSitekey := ""
		if !ok {
			// If the sitekey cannot be extracted from our list, the sitekey parameter is required
			if len(sitekey) == 0 {
				return nil, errors.New("sitekey is a required parameter for this retailer")
			} else {
				retailerSitekey = sitekey[0]
			}
		} else {
			retailerSitekey = string(tempSitekey)
		}
		if retailerSitekey == "" {
			return nil, errors.New("sitekey is a required parameter for this retailer")
		}
		err = RequestReCaptchaV3Token(retailerSitekey, url, proxy, retailer)
	case enums.HCaptcha:
		tokens := captchaStore.HCaptchaTokens[retailer]
		for index, token := range tokens {
			if token.URL == url && token.Proxy.ID == proxy.ID {
				// If a valid token exists, remove it from the list of tokens and return it
				tokens[len(tokens)-1], tokens[index] = tokens[index], tokens[len(tokens)-1]
				captchaStore.HCaptchaTokens[retailer] = tokens[:len(tokens)-1]
				return *token, nil
			}
		}
		// Otherwise, request a token
		tempSitekey, ok := enums.HCaptchaSitekeys[retailer]
		retailerSitekey := ""
		if !ok {
			// If the sitekey cannot be extracted from our list, the sitekey parameter is required
			if len(sitekey) == 0 {
				return nil, errors.New("sitekey is a required parameter for this retailer")
			} else {
				retailerSitekey = sitekey[0]
			}
		} else {
			retailerSitekey = string(tempSitekey)
		}
		if retailerSitekey == "" {
			return nil, errors.New("sitekey is a required parameter for this retailer")
		}
		err = RequestHCaptchaToken(retailerSitekey, url, proxy, retailer)
	case enums.GeeTestCaptcha:
		// TODO @silent
	}
	// If none are available, return nil -- the Task requesting a captcha should poll this function frequently until successful
	return nil, err
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
		case settings.TwoCaptchaAPIKey:
			go func() {
				defer wg.Done()
				token, err := TwoCaptchaReq(settings.TwoCaptchaAPIKey, "https://2captcha.com/in.php?key="+settings.TwoCaptchaAPIKey+"&method=userrecaptcha&googlekey="+sitekey+"&pageurl="+url+"&proxy="+common.ProxyCleaner(proxy)+"&proxytype=http")
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
				captchaStore.ReCaptchaV2Tokens[retailer] = append(captchaStore.ReCaptchaV2Tokens[retailer], &entities.ReCaptchaToken{
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
					Type:          "RecaptchaV2Task",
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
				port, err := strconv.Atoi(proxy.Port)
				if err != nil {
					return
				}
				capMonsterResponse, err := CapMonsterReq(settings.CapMonsterAPIKey, CapMonsterTaskInfo{
					Type:          "NoCaptchaTask",
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
	err = KeyErrors(settings, keyError, enums.ReCaptchaV2)
	return err
}

// RequestReCaptchaV3Token requests a ReCaptchaV3 token from all available APIs and the frontend
func RequestReCaptchaV3Token(sitekey string, url string, proxy entities.Proxy, retailer enums.Retailer) error {
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
		case settings.TwoCaptchaAPIKey:
			go func() {
				defer wg.Done()
				token, err := TwoCaptchaReq(settings.TwoCaptchaAPIKey, "https://2captcha.com/in.php?key="+settings.TwoCaptchaAPIKey+"&method=userrecaptcha&googlekey="+sitekey+"&pageurl="+url+"&version=v3")
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
		case settings.TwoCaptchaAPIKey:
			go func() {
				defer wg.Done()

				token, err := TwoCaptchaReq(settings.TwoCaptchaAPIKey, "https://2captcha.com/in.php?key="+settings.TwoCaptchaAPIKey+"&method=hcaptcha&sitekey="+sitekey+"&pageurl="+url+"&proxy="+common.ProxyCleaner(proxy)+"&proxytype=http")
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
