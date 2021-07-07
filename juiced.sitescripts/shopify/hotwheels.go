package shopify

import (
	"encoding/json"
	"fmt"
	"time"

	"backend.juicedbot.io/juiced.client/http"
	"backend.juicedbot.io/juiced.infrastructure/common/enums"
	"backend.juicedbot.io/juiced.sitescripts/util"
)

const (
	HotWheelsBaseLoginEndpoint = "https://platform.mattel/shopify/login?shop_url=https%3A%2F%2Fhwc-flash-sale.mattel.com&return_to=/account/&client_id=hot-wheels-collectors"
	HotWheelsLoginEndpoint     = "https://login.platform.mattel/account/login"
)

type HotWheelsLoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

func (task *Task) HotWheelsLogin() {
	task.PublishEvent(enums.LoggingIn, enums.TaskUpdate)
	loggedIn := false
	for !loggedIn {
		needToStop := task.CheckForStop()
		if needToStop {
			return
		}
		loggedIn = task.HotWheelsLoginHelper()
		if !loggedIn {
			time.Sleep(time.Duration(task.Task.Task.TaskDelay) * time.Millisecond)
		}
	}
}

func (task *Task) HotWheelsLoginHelper() bool {
	resp, _, err := util.MakeRequest(&util.Request{
		Client: task.Client,
		Method: "GET",
		URL:    task.Task.Task.ShopifyTaskInfo.BaseURL,
		RawHeaders: http.RawHeader{
			{"sec-ch-ua", `" Not;A Brand";v="99", "Google Chrome";v="91", "Chromium";v="91"`},
			{"accept", "text/html, text/javascript, */*; q=0.01"},
			{"x-requested-with", "XMLHttpRequest"},
			{"sec-ch-ua-mobile", "?0"},
			{"user-agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/91.0.4472.106 Safari/537.36"},
			{"origin", task.Task.Task.ShopifyTaskInfo.BaseURL},
			{"sec-fetch-site", "same-origin"},
			{"sec-fetch-mode", "cors"},
			{"sec-fetch-dest", "empty"},
			{"referer", task.Task.Task.ShopifyTaskInfo.BaseURL + "/"},
			{"accept-encoding", "gzip, deflate"},
			{"accept-language", "en-US,en;q=0.9"},
		},
	})
	if err != nil {
		fmt.Println(err)
	}

	loginURL := resp.Request.URL.String()
	query := loginURL[36:]

	loginRequest := HotWheelsLoginRequest{
		Email:    task.AccountInfo.Email,
		Password: task.AccountInfo.Password,
	}
	data, _ := json.Marshal(loginRequest)
	resp, _, err = util.MakeRequest(&util.Request{
		Client: task.Client,
		Method: "POST",
		URL:    HotWheelsLoginEndpoint,
		RawHeaders: http.RawHeader{
			{"content-length", fmt.Sprint(len(data))},
			{"sec-ch-ua", `" Not;A Brand";v="99", "Google Chrome";v="91", "Chromium";v="91"`},
			{"accept", "application/json, text/javascript, */*; q=0.01"},
			{"x-requested-with", "XMLHttpRequest"},
			{"sec-ch-ua-mobile", "?0"},
			{"user-agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/91.0.4472.106 Safari/537.36"},
			{"content-type", "application/json; charset=UTF-8"},
			{"origin", task.Task.Task.ShopifyTaskInfo.BaseURL},
			{"sec-fetch-site", "same-origin"},
			{"sec-fetch-mode", "cors"},
			{"sec-fetch-dest", "empty"},
			{"referer", loginURL},
			{"accept-encoding", "gzip, deflate"},
			{"accept-language", "en-US,en;q=0.9"},
		},
		Data: data,
	})
	if err != nil {
		fmt.Println(err)
	}
	switch resp.StatusCode {
	case 307:
		fmt.Println("Redirected")

	}

	resp, _, err = util.MakeRequest(&util.Request{
		Client: task.Client,
		Method: "GET",
		URL:    "https://login.platform.mattel/oauth/auth?" + query,
		RawHeaders: http.RawHeader{
			{"sec-ch-ua", `" Not;A Brand";v="99", "Google Chrome";v="91", "Chromium";v="91"`},
			{"accept", "text/html, text/javascript, */*; q=0.01"},
			{"x-requested-with", "XMLHttpRequest"},
			{"sec-ch-ua-mobile", "?0"},
			{"user-agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/91.0.4472.106 Safari/537.36"},
			{"origin", task.Task.Task.ShopifyTaskInfo.BaseURL},
			{"sec-fetch-site", "same-origin"},
			{"sec-fetch-mode", "cors"},
			{"sec-fetch-dest", "empty"},
			{"referer", loginURL},
			{"accept-encoding", "gzip, deflate"},
			{"accept-language", "en-US,en;q=0.9"},
		},
	})
	if err != nil {
		fmt.Println(err)
	}

	return resp.StatusCode == 200
}
