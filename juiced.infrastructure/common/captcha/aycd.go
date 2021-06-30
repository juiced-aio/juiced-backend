package captcha

import (
	"errors"
	"strings"

	"backend.juicedbot.io/juiced.infrastructure/common/entities"
	"gitlab.com/aycd-inc/autosolve-clients/autosolve-client-go"
)

var JUICED_CLIENT_ID = "Juiced-AIO-4d1c6af4-6bc3-4443-beae-892f218d2c17"
var aycdStatus = autosolve.Disconnected

func InitAycd() error {
	// Establishes listener function to receive token responses
	var tokenListener autosolve.CaptchaTokenResponseListener = func(tokenResponse autosolve.CaptchaTokenResponse) {
		proxy := entities.Proxy{}
		var err error
		if tokenResponse.Request.Proxy != "" {
			proxy, err = ParseProxy(tokenResponse.Request.Proxy)
			if err != nil {
				// TODO @silent: Handle
				// (This shouldn't really fail, since we validate the Proxy entity that is passed into the autosolve Request initially)
				// (But still good to handle regardless)
			}
		}

		if !strings.Contains(tokenResponse.TaskId, "|") {
			// TODO @silent: Handle
			// (This shouldn't really happen, since we hardcode the TaskId in the Request to include this)
			// (But still good to handle regardless)
		}

		token := &entities.ReCaptchaToken{
			URL:   tokenResponse.Request.Url,
			Proxy: proxy,
			Token: tokenResponse.Token,
		}

		retailerAndReCaptchaType := strings.Split(tokenResponse.TaskId, "|")
		retailer := retailerAndReCaptchaType[0]
		reCaptchaType := retailerAndReCaptchaType[1]
		if reCaptchaType == "V2" {
			captchaStore.AYCDReCaptchaV2Tokens[retailer] = append(captchaStore.AYCDReCaptchaV2Tokens[retailer], token)
		} else {
			captchaStore.AYCDReCaptchaV3Tokens[retailer] = append(captchaStore.AYCDReCaptchaV3Tokens[retailer], token)
		}
	}

	// Establishes listener function to receive cancel token responses
	var tokenCancelListener autosolve.CaptchaTokenCancelResponseListener = func(cancelTokenResponse autosolve.CaptchaTokenCancelResponse) {
		// TODO @silent: Handle
	}

	// Establishes listener function to receive status updates from AutoSolve
	var statusListener autosolve.StatusListener = func(status autosolve.Status) {
		aycdStatus = status
	}

	// Establishes listener function to receive errors
	var errorListener autosolve.ErrorListener = func(err error) {
		// TODO @silent: Handle
	}

	return autosolve.Load(JUICED_CLIENT_ID, tokenListener, tokenCancelListener, statusListener, errorListener)
}

func ConnectToAycd(accessToken, apiKey string) error {
	result, err := autosolve.Connect(accessToken, apiKey)
	if err != nil {
		return err
	}

	switch result {
	case autosolve.Success:
		return nil
	case autosolve.InvalidClientId:
		return errors.New("invalid client key")
	case autosolve.InvalidAccessToken:
		return errors.New("invalid access token")
	case autosolve.InvalidApiKey:
		return errors.New("invalid api key")
	case autosolve.InvalidCredentials:
		return errors.New("invalid credentials")
	}

	return nil
}
