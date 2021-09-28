package gamestop

import (
	"fmt"
	"log"
	"net/url"
	"time"

	"backend.juicedbot.io/juiced.antibot/akamai"
	"backend.juicedbot.io/juiced.infrastructure/captcha"
	"backend.juicedbot.io/juiced.infrastructure/entities"
	"backend.juicedbot.io/juiced.infrastructure/enums"
	u "backend.juicedbot.io/juiced.infrastructure/util"
	"backend.juicedbot.io/juiced.sitescripts/util"

	browser "github.com/eddycjy/fake-useragent"
)

const MAX_RETRIES = 5

func CreateTask(input entities.TaskInput, baseTask *entities.BaseTask) (entities.RetailerTask, error) {
	gameStopTaskInput, err := ValidateTaskInput(input.SiteSpecificInfo)
	if err != nil {
		return nil, err
	}
	return &Task{
		Input:         input,
		GameStopInput: gameStopTaskInput,
		BaseTask:      baseTask,
	}, nil
}

func (task *Task) GetSetupFunctions() []entities.TaskFunction {
	setupTaskFunctions := []entities.TaskFunction{}

	if task.GameStopInput.TaskType == enums.TaskTypeAccount {
		setupTaskFunctions = append(setupTaskFunctions, []entities.TaskFunction{
			// 1. Login
			{
				Function:         task.Login,
				StatusBegin:      enums.LoggingIn,
				StatusPercentage: 30,
				MaxRetries:       MAX_RETRIES,
			},
			// TODO: 2. Refresh Login
			// {
			// 	Function:        task.Login,
			// 	StatusBegin:     enums.LoggingIn,
			// 	MaxRetries:      MAX_RETRIES,
			// 	RefreshFunction: true,
			// 	RefreshEvery:    1800,
			// },
		}...)
	} else {
		setupTaskFunctions = append(setupTaskFunctions, []entities.TaskFunction{
			// 1. LoginGuest
			{
				Function:         task.LoginGuest,
				StatusBegin:      enums.SettingUp,
				StatusPercentage: 30,
				MaxRetries:       MAX_RETRIES,
			},
			// TODO: 2. Refresh LoginGuest
			// {
			// 	Function:        task.LoginGuest,
			// 	StatusBegin:     enums.SettingUp,
			// 	MaxRetries:      MAX_RETRIES,
			// 	RefreshFunction: true,
			// 	RefreshEvery:    1800,
			// },
		}...)
	}

	return setupTaskFunctions
}

func (task *Task) GetMainFunctions() []entities.TaskFunction {
	mainTaskFunctions := []entities.TaskFunction{
		// 3. AddToCart
		{
			Function:         task.AddToCart,
			StatusBegin:      enums.AddingToCart,
			StatusPercentage: 50,
			MsBetweenRetries: task.Input.DelayMS,
		},
	}
	return mainTaskFunctions
}

func (task *Task) Login() (bool, string) {
	resp, body, err := util.MakeRequest(&util.Request{
		Client:     task.BaseTask.Client,
		Method:     "GET",
		URL:        BaseLoginEndpoint,
		RawHeaders: util.DefaultRawHeaders,
	})
	if err != nil {
		return false, fmt.Sprintf(enums.LoginFailure, err.Error())
	}
	switch resp.StatusCode {
	case 200:
	default:
		return false, fmt.Sprintf(enums.LoginFailure, fmt.Sprintf(enums.UnknownError, resp.StatusCode))
	}

	csrf, err := u.FindInString(body, `name="csrf_token" value="`, `"`)
	if err != nil {
		return false, fmt.Sprintf(enums.LoginFailure, err.Error())
	}

	err = akamai.HandleAkamaiTask(task.BaseTask, BaseEndpoint+"/", BaseEndpoint, AkamaiEndpoint, BaseURL)
	if err != nil {
		return false, fmt.Sprintf(enums.LoginFailure, err.Error())
	}

	loginResponse := LoginResponse{}
	form := url.Values{
		"loginEmail":         {task.GameStopInput.Email},
		"loginPassword":      {task.GameStopInput.Password},
		"loginRememberMe":    {"true"},
		"userTimezoneOffset": {"420"},
		"csrf_token":         {csrf},
	}
	_, _, err = util.MakeRequest(&util.Request{
		Client: task.BaseTask.Client,
		Method: "POST",
		URL:    LoginEndpoint,
		RawHeaders: [][2]string{
			{"content-length", fmt.Sprint(len(form.Encode()))},
			{"pragma", "no-cache"},
			{"cache-control", "no-cache"},
			{"sec-ch-ua", "\" Not A;Brand\";v=\"99\", \"Chromium\";v=\"90\", \"Google Chrome\";v=\"90\""},
			{"accept", "application/json, text/javascript, */*; q=0.01"},
			{"x-requested-with", "XMLHttpRequest"},
			{"sec-ch-ua-mobile", "?0"},
			{"user-agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/90.0.4430.212 Safari/537.36"},
			{"content-type", "application/x-www-form-urlencoded; charset=UTF-8"},
			{"origin", BaseEndpoint},
			{"sec-fetch-site", "same-origin"},
			{"sec-fetch-mode", "cors"},
			{"sec-fetch-dest", "empty"},
			{"referer", BaseEndpoint + "/"},
			{"accept-encoding", "gzip, deflate, br"},
			{"accept-language", "en-US,en;q=0.9"},
		},
		Data:               []byte(form.Encode()),
		ResponseBodyStruct: &loginResponse,
	})
	if err != nil {
		log.Println(err.Error())
		// return false, fmt.Sprintf(enums.LoginFailure, err.Error())
	}

	if !loginResponse.LoginStatus.Success {
		return false, fmt.Sprintf(enums.LoginFailure, loginResponse.LoginStatus.ErrorMessage)
	}
	_, _, err = util.MakeRequest(&util.Request{
		Client: task.BaseTask.Client,
		Method: "GET",
		URL:    AccountEndpoint + "/",
		RawHeaders: [][2]string{
			{"pragma", "no-cache"},
			{"cache-control", "no-cache"},
			{"sec-ch-ua", "\" Not A;Brand\";v=\"99\", \"Chromium\";v=\"90\", \"Google Chrome\";v=\"90\""},
			{"sec-ch-ua-mobile", "?0"},
			{"upgrade-insecure-requests", "1"},
			{"user-agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/90.0.4430.212 Safari/537.36"},
			{"accept", "text/html,application/xhtml+xml,application/xml;q=0.9,image/avif,image/webp,image/apng,*/*;q=0.8,application/signed-exchange;v=b3;q=0.9"},
			{"sec-fetch-site", "same-origin"},
			{"sec-fetch-mode", "navigate"},
			{"sec-fetch-user", "?1"},
			{"sec-fetch-dest", "document"},
			{"referer", BaseEndpoint + "/"},
			{"accept-encoding", "gzip, deflate, br"},
			{"accept-language", "en-US,en;q=0.9"},
		},
	})
	if err != nil {
		log.Println(err.Error())
		// return false, fmt.Sprintf(enums.LoginFailure, err.Error())
	}

	switch resp.StatusCode {
	case 200:
		return true, enums.LoginSuccess

	case 403:
		err = akamai.HandleAkamaiTask(task.BaseTask, BaseEndpoint+"/", BaseEndpoint, AkamaiEndpoint, BaseURL)
		if err != nil {
			return false, fmt.Sprintf(enums.LoginFailure, err.Error())
		}
		return false, ""
	}

	return false, fmt.Sprintf(enums.LoginFailure, fmt.Sprintf(enums.UnknownError, resp.StatusCode))
}

func (task *Task) LoginGuest() (bool, string) {
	resp, _, err := util.MakeRequest(&util.Request{
		Client: task.BaseTask.Client,
		Method: "GET",
		URL:    BaseEndpoint,
		RawHeaders: [][2]string{
			{"upgrade-insecure-requests", "1"},
			{"user-agent", browser.Chrome()},
			{"accept", "text/html,application/xhtml+xml,application/xml;q=0.9,image/avif,image/webp,image/apng,*/*;q=0.8,application/signed-exchange;v=b3;q=0.9"},
			{"sec-fetch-site", "none"},
			{"sec-fetch-mode", "navigate"},
			{"sec-fetch-user", "?1"},
			{"sec-fetch-dest", "document"},
			{"sec-ch-ua", `" Not A;Brand";v="99", "Chromium";v="90", "Google Chrome";v="90"`},
			{"sec-ch-ua-mobile", "?0"},
			{"accept-encoding", "gzip, deflate, br"},
			{"accept-language", "en-US,en;q=0.9"},
		},
	})
	if err != nil {
		log.Println(err.Error())
		// return false, fmt.Sprintf(enums.SettingUpFailure, err.Error())
	}

	switch resp.StatusCode {
	case 200:
	default:
		return false, fmt.Sprintf(enums.SettingUpFailure, fmt.Sprintf(enums.UnknownError, resp.StatusCode))
	}

	err = akamai.HandleAkamaiTask(task.BaseTask, BaseEndpoint+"/", BaseEndpoint, AkamaiEndpoint, BaseURL)
	if err != nil {
		return false, fmt.Sprintf(enums.LoginFailure, err.Error())
	}

	return true, enums.SettingUpSuccess
}

func (task *Task) AddToCart() (bool, string) {
	quantity := task.Input.Quantity
	if maxQuantity, ok := task.BaseTask.ProductInfo.SiteSpecificInfo["MaxQuantity"].(int); ok && maxQuantity != 0 && quantity > maxQuantity {
		quantity = maxQuantity
	}

	pid := task.BaseTask.ProductInfo.SKU
	if vid, ok := task.BaseTask.ProductInfo.SiteSpecificInfo["VID"].(string); ok && vid != "" {
		pid = vid
	}

	addToCartResponse := AddToCartResponse{}
	form := url.Values{
		"pid":            {pid},
		"quantity":       {fmt.Sprint(quantity)},
		"upsellID":       {""},
		"purPROID":       {""},
		"options":        {"[]"},
		"isSDD":          {"false"},
		"deliveryOption": {"home"},
		"pageSpecified":  {"PDP"},
		"recommTitle":    {""},
	}
	if task.CaptchaToken != "" {
		form.Add("g-recaptcha-response", task.CaptchaToken)
	}
	resp, _, err := util.MakeRequest(&util.Request{
		Client: task.BaseTask.Client,
		Method: "POST",
		URL:    fmt.Sprintf(AddToCartEndpoint, task.BaseTask.ProductInfo.SKU),
		RawHeaders: [][2]string{
			{"content-length", fmt.Sprint(len(form.Encode()))},
			{"sec-ch-ua", "\" Not A;Brand\";v=\"99\", \"Chromium\";v=\"90\", \"Google Chrome\";v=\"90\""},
			{"accept", "*/*"},
			{"x-requested-with", "XMLHttpRequest"},
			{"sec-ch-ua-mobile", "?0"},
			{"user-agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/90.0.4430.212 Safari/537.36"},
			{"content-type", "application/x-www-form-urlencoded; charset=UTF-8"},
			{"origin", BaseEndpoint},
			{"sec-fetch-site", "same-origin"},
			{"sec-fetch-mode", "cors"},
			{"sec-fetch-dest", "empty"},
			{"referer", task.BaseTask.ProductInfo.ItemURL},
			{"accept-encoding", "gzip, deflate, br"},
			{"accept-language", "en-US,en;q=0.9"},
		},
		Data:               []byte(form.Encode()),
		ResponseBodyStruct: &addToCartResponse,
	})
	if err != nil {
		log.Println(err.Error())
		// return false, fmt.Sprintf(enums.AddingToCartFailure, err.Error())
	}

	switch resp.StatusCode {
	case 200:
		if addToCartResponse.QuantityTotal == 0 {
			if addToCartResponse.CaptchaProtected {
				task.BaseTask.PublishEvent(enums.WaitingForCaptcha, 50, enums.TaskUpdate)
				task.CaptchaProtected = true
				proxy := entities.Proxy{}
				if task.BaseTask.Proxy != nil {
					proxy = *task.BaseTask.Proxy
				}
				token, err := captcha.RequestCaptchaToken(enums.ReCaptchaV2, enums.GameStop, task.BaseTask.ProductInfo.ItemURL, "atc", 0.8, proxy)
				if err != nil {
					return false, fmt.Sprintf(enums.AddingToCartFailure, err.Error())
				}
				for token == nil {
					needToStop := task.BaseTask.CheckForStop()
					if needToStop {
						return false, ""
					}
					token = captcha.PollCaptchaTokens(enums.ReCaptchaV2, enums.GameStop, task.BaseTask.ProductInfo.ItemURL, proxy)
					time.Sleep(1 * time.Second / 10)
				}
				tokenInfo, ok := token.(entities.ReCaptchaToken)
				if !ok {
					return false, fmt.Sprintf(enums.AddingToCartFailure, enums.BadCaptchaTokenError)
				}
				task.CaptchaToken = tokenInfo.Token

				needToStop := task.BaseTask.CheckForStop()
				if needToStop {
					return false, ""
				}

				task.BaseTask.PublishEvent(enums.AddingToCart, 50, enums.TaskUpdate)
				return false, ""
			}
			return false, fmt.Sprintf(enums.AddingToCartFailure, addToCartResponse.Message)
		} else {
			return true, enums.AddingToCartSuccess
		}

	case 403:
		err = akamai.HandleAkamaiTask(task.BaseTask, BaseEndpoint+"/", BaseEndpoint, AkamaiEndpoint, BaseURL)
		if err != nil {
			return false, fmt.Sprintf(enums.LoginFailure, err.Error())
		}
		return false, ""
	}

	return false, fmt.Sprintf(enums.AddingToCartFailure, fmt.Sprintf(enums.UnknownError, resp.StatusCode))
}
