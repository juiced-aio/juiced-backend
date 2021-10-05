package shopify

import (
	"fmt"
	"log"

	"backend.juicedbot.io/juiced.client/http"
	"backend.juicedbot.io/juiced.infrastructure/entities"
	"backend.juicedbot.io/juiced.infrastructure/enums"
	u "backend.juicedbot.io/juiced.infrastructure/util"
	"backend.juicedbot.io/juiced.sitescripts/util"
)

const MAX_RETRIES = 5

func CreateTask(input entities.TaskInput, baseTask *entities.BaseTask) (entities.RetailerTask, error) {
	shopifyTaskInput, err := ValidateTaskInput(input.SiteSpecificInfo)
	if err != nil {
		return nil, err
	}
	return &Task{
		Input:        input,
		ShopifyInput: shopifyTaskInput,
		BaseTask:     baseTask,
	}, nil
}

func (task *Task) GetSetupFunctions() []entities.TaskFunction {
	setupTaskFunctions := []entities.TaskFunction{}

	if task.ShopifyInput.SitePassword != "" {
		setupTaskFunctions = append(setupTaskFunctions, []entities.TaskFunction{
			// 1. SubmitPassword
			{
				Function:         task.SubmitPassword,
				StatusBegin:      enums.SettingUp,
				StatusPercentage: 20,
				MaxRetries:       MAX_RETRIES,
			},
		}...)
	}

	if task.ShopifyInput.TaskType == enums.TaskTypeAccount {
		if task.ShopifyInput.ShopifyRetailer == enums.GenericShopify {
			setupTaskFunctions = append(setupTaskFunctions, []entities.TaskFunction{
				// 2. Login
				{
					Function:         task.Login,
					StatusBegin:      enums.LoggingIn,
					StatusPercentage: 30,
					MaxRetries:       MAX_RETRIES,
				},
				// TODO: 3. Refresh Login
				// {
				// 	Function:        task.Login,
				// 	StatusBegin:     enums.LoggingIn,
				// 	MaxRetries:      MAX_RETRIES,
				// 	RefreshFunction: true,
				// 	RefreshEvery:    1800,
				// },
			}...)
		} else {
			switch task.ShopifyInput.ShopifyRetailer {
			case enums.MattelCreations:
				setupTaskFunctions = append(setupTaskFunctions, []entities.TaskFunction{
					// 2. Login
					{
						Function:         task.MattelCreationsLogin,
						StatusBegin:      enums.LoggingIn,
						StatusPercentage: 30,
						MaxRetries:       MAX_RETRIES,
					},
					// TODO: 3. Refresh Login
					// {
					// 	Function:        task.Login,
					// 	StatusBegin:     enums.LoggingIn,
					// 	MaxRetries:      MAX_RETRIES,
					// 	RefreshFunction: true,
					// 	RefreshEvery:    1800,
					// },
				}...)
			default:
				setupTaskFunctions = append(setupTaskFunctions, []entities.TaskFunction{
					// 2. Login
					{
						Function:         task.Login,
						StatusBegin:      enums.LoggingIn,
						StatusPercentage: 30,
						MaxRetries:       MAX_RETRIES,
					},
					// TODO: 3. Refresh Login
					// {
					// 	Function:        task.Login,
					// 	StatusBegin:     enums.LoggingIn,
					// 	MaxRetries:      MAX_RETRIES,
					// 	RefreshFunction: true,
					// 	RefreshEvery:    1800,
					// },
				}...)
			}
		}
	} else {
		if task.ShopifyInput.SitePassword == "" {
			setupTaskFunctions = append(setupTaskFunctions, []entities.TaskFunction{
				// 2. LoginGuest
				{
					Function:         task.LoginGuest,
					StatusBegin:      enums.SettingUp,
					StatusPercentage: 30,
					MaxRetries:       MAX_RETRIES,
				},
				// TODO: 3. Refresh LoginGuest
				// {
				// 	Function:        task.LoginGuest,
				// 	StatusBegin:     enums.SettingUp,
				// 	MaxRetries:      MAX_RETRIES,
				// 	RefreshFunction: true,
				// 	RefreshEvery:    1800,
				// },
			}...)
		}
	}
	setupTaskFunctions = append(setupTaskFunctions, []entities.TaskFunction{
		// 4. Preload
		{
			Function:         task.Preload,
			StatusBegin:      enums.SettingUp,
			StatusPercentage: 40,
			MaxRetries:       MAX_RETRIES,
		},
	}...)

	return setupTaskFunctions
}

func (task *Task) GetMainFunctions() []entities.TaskFunction {
	mainTaskFunctions := []entities.TaskFunction{
		// 5. AddToCart
		{
			Function:         task.AddToCart,
			StatusBegin:      enums.AddingToCart,
			StatusPercentage: 55,
			MsBetweenRetries: task.Input.DelayMS,
		},
		// 6. GoToCheckout
		{
			Function:         task.GoToCheckout,
			StatusBegin:      enums.GettingCartInfo,
			StatusPercentage: 65,
			MaxRetries:       MAX_RETRIES,
		},
		// 7. HandleQueue
		{
			Function:         task.HandleQueue,
			StatusBegin:      enums.HandlingQueue,
			StatusPercentage: 70,
			MaxRetries:       MAX_RETRIES,
		},
		// 8. SubmitShippingDetails
		{
			Function:         task.SubmitShippingDetails,
			StatusBegin:      enums.SettingShippingInfo,
			StatusPercentage: 75,
			MaxRetries:       MAX_RETRIES,
		},
		// 9. SubmitShippingRate
		{
			Function:         task.SubmitShippingRate,
			StatusBegin:      enums.SettingShippingInfo,
			StatusPercentage: 80,
			MaxRetries:       MAX_RETRIES,
		},
		// 10. GetCreditID
		{
			Function:         task.GetCreditID,
			StatusBegin:      enums.GettingBillingInfo,
			StatusPercentage: 85,
			MaxRetries:       MAX_RETRIES,
		},
		// 11. SubmitBillingDetails
		{
			Function:         task.SubmitBillingDetails,
			StatusBegin:      enums.SettingBillingInfo,
			StatusPercentage: 90,
			MaxRetries:       MAX_RETRIES,
		},
		// 12. SubmitOrder
		{
			Function:         task.SubmitOrder,
			StatusBegin:      enums.CheckingOut,
			StatusPercentage: 95,
			MaxRetries:       MAX_RETRIES,
			CheckoutFunction: true,
		},
	}
	return mainTaskFunctions
}

func (task *Task) Login() (bool, string) {
	return true, ""
}

func (task *Task) LoginGuest() (bool, string) {
	resp, _, err := util.MakeRequest(&util.Request{
		Client: task.BaseTask.Client,
		Method: "GET",
		URL:    task.ShopifyInput.SiteURL,
		RawHeaders: http.RawHeader{
			{"sec-ch-ua", `" Not;A Brand";v="99", "Google Chrome";v="91", "Chromium";v="91"`},
			{"accept", "application/json, text/javascript, */*; q=0.01"},
			{"x-requested-with", "XMLHttpRequest"},
			{"sec-ch-ua-mobile", "?0"},
			{"user-agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/91.0.4472.106 Safari/537.36"},
			{"content-type", "application/x-www-form-urlencoded; charset=UTF-8"},
			{"sec-fetch-site", "same-origin"},
			{"sec-fetch-site", "cors"},
			{"sec-fetch-dest", "empty"},
			{"accept-encoding", "gzip, deflate"},
			{"accept-language", "en-US,en;q=0.9"},
		},
	})
	if err != nil {
		log.Println(err.Error())
		// return false, fmt.Sprintf(enums.SettingUpFailure, err.Error())
	}
	switch resp.StatusCode {
	case 200:
		return true, enums.SettingUpSuccess
	}
	return false, fmt.Sprintf(enums.SettingUpFailure, fmt.Sprintf(enums.UnknownError, resp.StatusCode))
}

func (task *Task) SubmitPassword() (bool, string) {
	params := u.CreateParams(map[string]string{
		"form_type": "storefront_password",
		"utf8":      "✓",
		"password":  task.ShopifyInput.SitePassword,
	})
	resp, _, err := util.MakeRequest(&util.Request{
		Client: task.BaseTask.Client,
		Method: "POST",
		URL:    task.ShopifyInput.SiteURL,
		RawHeaders: http.RawHeader{
			{"content-length", fmt.Sprint(len(params))},
			{"sec-ch-ua", `" Not;A Brand";v="99", "Google Chrome";v="91", "Chromium";v="91"`},
			{"accept", "application/json, text/javascript, */*; q=0.01"},
			{"x-requested-with", "XMLHttpRequest"},
			{"sec-ch-ua-mobile", "?0"},
			{"user-agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/91.0.4472.106 Safari/537.36"},
			{"content-type", "application/x-www-form-urlencoded; charset=UTF-8"},
			{"sec-fetch-site", "same-origin"},
			{"sec-fetch-site", "cors"},
			{"sec-fetch-dest", "empty"},
			{"accept-encoding", "gzip, deflate"},
			{"accept-language", "en-US,en;q=0.9"},
		},
		Data: []byte(params),
	})
	if err != nil {
		log.Println(err.Error())
		// return false, fmt.Sprintf(enums.SettingUpFailure, err.Error())
	}
	switch resp.StatusCode {
	case 200:
		return true, enums.SettingUpSuccess
	}
	return false, fmt.Sprintf(enums.SettingUpFailure, fmt.Sprintf(enums.UnknownError, resp.StatusCode))
}

func (task *Task) Preload() (bool, string) {
	return true, ""
}

func (task *Task) AddToCart() (bool, string) {
	return true, ""
}

func (task *Task) GoToCheckout() (bool, string) {
	return true, ""
}

func (task *Task) HandleQueue() (bool, string) {
	return true, ""
}

func (task *Task) SubmitShippingDetails() (bool, string) {
	return true, ""
}

func (task *Task) SubmitShippingRate() (bool, string) {
	return true, ""
}

func (task *Task) GetCreditID() (bool, string) {
	return true, ""
}

func (task *Task) SubmitBillingDetails() (bool, string) {
	return true, ""
}

func (task *Task) SubmitOrder() (bool, string) {
	return true, ""
}
