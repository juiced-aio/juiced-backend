package stores

import (
	"errors"

	"backend.juicedbot.io/m/v2/juiced.infrastructure/common/entities"
	"backend.juicedbot.io/m/v2/juiced.infrastructure/common/enums"
	"backend.juicedbot.io/m/v2/juiced.infrastructure/common/events"
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
		err = RequestReCaptchaV2Token(retailerSitekey, url, proxy)
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
		err = RequestReCaptchaV3Token(retailerSitekey, url, proxy)
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
		err = RequestHCaptchaToken(retailerSitekey, url, proxy)
	case enums.GeeTestCaptcha:
		// TODO @silent
	}
	// If none are available, return nil -- the Task requesting a captcha should poll this function frequently until successful
	return nil, err
}

// RequestReCaptchaV2Token requests a ReCaptchaV2 token from all available APIs and the frontend
func RequestReCaptchaV2Token(sitekey string, url string, proxy entities.Proxy) error {

	return nil
}

// RequestReCaptchaV3Token requests a ReCaptchaV3 token from all available APIs and the frontend
func RequestReCaptchaV3Token(sitekey string, url string, proxy entities.Proxy) error {

	return nil
}

// RequestHCaptchaToken requests a HCaptcha token from all available APIs and the frontend
func RequestHCaptchaToken(sitekey string, url string, proxy entities.Proxy) error {

	return nil
}

// RequestGeeTestCaptchaToken requests a GeeTestCaptcha token from all available APIs and the frontend
func RequestGeeTestCaptchaToken(sitekey string, url string, proxy entities.Proxy) error {

	return nil
}
