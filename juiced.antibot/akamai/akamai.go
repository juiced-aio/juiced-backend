package akamai

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"strings"
	"time"

	"backend.juicedbot.io/juiced.infrastructure/entities"
	"backend.juicedbot.io/juiced.infrastructure/staticstores"
	u "backend.juicedbot.io/juiced.infrastructure/util"
	sec "backend.juicedbot.io/juiced.security/auth/util"
	"backend.juicedbot.io/juiced.sitescripts/util"
)

func HandleAkamaiMonitor(monitor *entities.BaseMonitor, referer, baseEndpoint, akamaiEndpoint string, baseURL *url.URL) error {
	quit := make(chan bool)
	defer func() {
		quit <- true
	}()

	cancellationToken := &u.CancellationToken{Cancel: false}
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
			time.Sleep(u.MS_TO_WAIT)
		}
	}()

	abckCookie := ""

	for _, cookie := range monitor.Client.Jar.Cookies(baseURL) {
		if cookie.Name == "_abck" {
			abckCookie = cookie.Value
		}
	}

	if abckCookie == "" {
		resp, _, err := util.MakeRequest(&util.Request{
			Client: monitor.Client,
			Method: "GET",
			URL:    akamaiEndpoint,
			RawHeaders: [][2]string{
				{"sec-ch-ua", "\" Not A;Brand\";v=\"99\", \"Chromium\";v=\"90\", \"Google Chrome\";v=\"90\""},
				{"sec-ch-ua-mobile", "?0"},
				{"user-agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/90.0.4430.212 Safari/537.36"},
				{"content-type", "text/plain;charset=UTF-8"},
				{"accept", "*/*"},
				{"origin", baseEndpoint},
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
			return errors.New(resp.Status)
		}
	}

	if cancellationToken.Cancel {
		return nil
	}

	userInfo := staticstores.GetUserInfo()

	genResponse, _, err := sec.ExperimentalAkamai(referer, "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/92.0.4515.107 Safari/537.36", abckCookie, 0, 0, 0, 0, userInfo)
	if err != nil {
		return err
	}

	if cancellationToken.Cancel {
		return nil
	}

	sensorRequest := SensorRequest{
		SensorData: genResponse.SensorData,
	}
	data, err := json.Marshal(sensorRequest)
	if err != nil {
		return err
	}
	sensorResponse := SensorResponse{}

	resp, _, err := util.MakeRequest(&util.Request{
		Client: monitor.Client,
		Method: "POST",
		URL:    akamaiEndpoint,
		RawHeaders: [][2]string{
			{"content-length", fmt.Sprint(len(data))},
			{"sec-ch-ua", "\" Not A;Brand\";v=\"99\", \"Chromium\";v=\"90\", \"Google Chrome\";v=\"90\""},
			{"sec-ch-ua-mobile", "?0"},
			{"user-agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/92.0.4515.107 Safari/537.36"},
			{"content-type", "text/plain;charset=UTF-8"},
			{"accept", "*/*"},
			{"origin", baseEndpoint},
			{"sec-fetch-site", "same-origin"},
			{"sec-fetch-mode", "cors"},
			{"sec-fetch-dest", "empty"},
			{"referer", referer},
			{"accept-encoding", "gzip, deflate, br"},
			{"accept-language", "en-US,en;q=0.9"},
		},
		Data:               data,
		ResponseBodyStruct: &sensorResponse,
	})
	if cancellationToken.Cancel {
		return nil
	}
	if err != nil {
		return err
	}
	if !sensorResponse.Success || resp.StatusCode != 201 {
		return fmt.Errorf("bad sensor: %s", resp.Status)
	}

	for _, cookie := range monitor.Client.Jar.Cookies(baseURL) {
		if cookie.Name == "_abck" {
			abckCookie = cookie.Value
		}
	}

	genResponse, _, err = sec.ExperimentalAkamai(referer, "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/92.0.4515.107 Safari/537.36", abckCookie, 1, genResponse.SavedD3, genResponse.SavedStartTS, genResponse.DeviceNum, userInfo)
	if err != nil {
		return err
	}
	if cancellationToken.Cancel {
		return nil
	}

	sensorRequest = SensorRequest{
		SensorData: genResponse.SensorData,
	}
	data, err = json.Marshal(sensorRequest)
	if err != nil {
		return err
	}
	sensorResponse = SensorResponse{}

	resp, _, err = util.MakeRequest(&util.Request{
		Client: monitor.Client,
		Method: "POST",
		URL:    akamaiEndpoint,
		RawHeaders: [][2]string{
			{"content-length", fmt.Sprint(len(data))},
			{"sec-ch-ua", "\" Not A;Brand\";v=\"99\", \"Chromium\";v=\"90\", \"Google Chrome\";v=\"90\""},
			{"sec-ch-ua-mobile", "?0"},
			{"user-agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/90.0.4430.212 Safari/537.36"},
			{"content-type", "text/plain;charset=UTF-8"},
			{"accept", "*/*"},
			{"origin", baseEndpoint},
			{"sec-fetch-site", "same-origin"},
			{"sec-fetch-mode", "cors"},
			{"sec-fetch-dest", "empty"},
			{"referer", referer},
			{"accept-encoding", "gzip, deflate, br"},
			{"accept-language", "en-US,en;q=0.9"},
		},
		Data:               data,
		ResponseBodyStruct: &sensorResponse,
	})
	if cancellationToken.Cancel {
		return nil
	}
	if err != nil {
		return err
	}
	if !sensorResponse.Success || resp.StatusCode != 201 {
		return fmt.Errorf("bad sensor: %s", resp.Status)
	}

	if strings.Contains(baseURL.Host, "gamestop") {
		if len(abckCookie) > 488 {
			return nil
		}
	} else {
		for _, cookie := range monitor.Client.Jar.Cookies(baseURL) {
			if cookie.Name == "_abck" {
				validator, _ := u.FindInString(cookie.Value, "~", "~")
				if validator == "-1" {
					if cancellationToken.Cancel {
						return nil
					}
					return HandleAkamaiMonitor(monitor, referer, baseEndpoint, akamaiEndpoint, baseURL)
				}
			}
		}
	}
	return nil
}

func HandleAkamaiTask(task *entities.BaseTask, referer, baseEndpoint, akamaiEndpoint string, baseURL *url.URL) error {
	quit := make(chan bool)
	defer func() {
		quit <- true
	}()

	cancellationToken := &u.CancellationToken{Cancel: false}
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
			time.Sleep(u.MS_TO_WAIT)
		}
	}()

	abckCookie := ""

	for _, cookie := range task.Client.Jar.Cookies(baseURL) {
		if cookie.Name == "_abck" {
			abckCookie = cookie.Value
		}
	}

	if abckCookie == "" {
		resp, _, err := util.MakeRequest(&util.Request{
			Client: task.Client,
			Method: "GET",
			URL:    akamaiEndpoint,
			RawHeaders: [][2]string{
				{"sec-ch-ua", "\" Not A;Brand\";v=\"99\", \"Chromium\";v=\"90\", \"Google Chrome\";v=\"90\""},
				{"sec-ch-ua-mobile", "?0"},
				{"user-agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/90.0.4430.212 Safari/537.36"},
				{"content-type", "text/plain;charset=UTF-8"},
				{"accept", "*/*"},
				{"origin", baseEndpoint},
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
			return errors.New(resp.Status)
		}
	}

	if cancellationToken.Cancel {
		return nil
	}

	userInfo := staticstores.GetUserInfo()

	genResponse, _, err := sec.ExperimentalAkamai(referer, "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/92.0.4515.107 Safari/537.36", abckCookie, 0, 0, 0, 0, userInfo)
	if err != nil {
		return err
	}

	if cancellationToken.Cancel {
		return nil
	}

	sensorRequest := SensorRequest{
		SensorData: genResponse.SensorData,
	}
	data, err := json.Marshal(sensorRequest)
	if err != nil {
		return err
	}
	sensorResponse := SensorResponse{}

	resp, _, err := util.MakeRequest(&util.Request{
		Client: task.Client,
		Method: "POST",
		URL:    akamaiEndpoint,
		RawHeaders: [][2]string{
			{"content-length", fmt.Sprint(len(data))},
			{"sec-ch-ua", "\" Not A;Brand\";v=\"99\", \"Chromium\";v=\"90\", \"Google Chrome\";v=\"90\""},
			{"sec-ch-ua-mobile", "?0"},
			{"user-agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/92.0.4515.107 Safari/537.36"},
			{"content-type", "text/plain;charset=UTF-8"},
			{"accept", "*/*"},
			{"origin", baseEndpoint},
			{"sec-fetch-site", "same-origin"},
			{"sec-fetch-mode", "cors"},
			{"sec-fetch-dest", "empty"},
			{"referer", referer},
			{"accept-encoding", "gzip, deflate, br"},
			{"accept-language", "en-US,en;q=0.9"},
		},
		Data:               data,
		ResponseBodyStruct: &sensorResponse,
	})
	if cancellationToken.Cancel {
		return nil
	}
	if err != nil {
		return err
	}
	if !sensorResponse.Success || resp.StatusCode != 201 {
		return fmt.Errorf("bad sensor: %s", resp.Status)
	}

	for _, cookie := range task.Client.Jar.Cookies(baseURL) {
		if cookie.Name == "_abck" {
			abckCookie = cookie.Value
		}
	}

	genResponse, _, err = sec.ExperimentalAkamai(referer, "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/92.0.4515.107 Safari/537.36", abckCookie, 1, genResponse.SavedD3, genResponse.SavedStartTS, genResponse.DeviceNum, userInfo)
	if err != nil {
		return err
	}
	if cancellationToken.Cancel {
		return nil
	}

	sensorRequest = SensorRequest{
		SensorData: genResponse.SensorData,
	}
	data, err = json.Marshal(sensorRequest)
	if err != nil {
		return err
	}
	sensorResponse = SensorResponse{}

	resp, _, err = util.MakeRequest(&util.Request{
		Client: task.Client,
		Method: "POST",
		URL:    akamaiEndpoint,
		RawHeaders: [][2]string{
			{"content-length", fmt.Sprint(len(data))},
			{"sec-ch-ua", "\" Not A;Brand\";v=\"99\", \"Chromium\";v=\"90\", \"Google Chrome\";v=\"90\""},
			{"sec-ch-ua-mobile", "?0"},
			{"user-agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/90.0.4430.212 Safari/537.36"},
			{"content-type", "text/plain;charset=UTF-8"},
			{"accept", "*/*"},
			{"origin", baseEndpoint},
			{"sec-fetch-site", "same-origin"},
			{"sec-fetch-mode", "cors"},
			{"sec-fetch-dest", "empty"},
			{"referer", referer},
			{"accept-encoding", "gzip, deflate, br"},
			{"accept-language", "en-US,en;q=0.9"},
		},
		Data:               data,
		ResponseBodyStruct: &sensorResponse,
	})
	if cancellationToken.Cancel {
		return nil
	}
	if err != nil {
		return err
	}
	if !sensorResponse.Success || resp.StatusCode != 201 {
		return fmt.Errorf("bad sensor: %s", resp.Status)
	}

	if strings.Contains(baseURL.Host, "gamestop") {
		if len(abckCookie) > 488 {
			return nil
		}
	} else {
		for _, cookie := range task.Client.Jar.Cookies(baseURL) {
			if cookie.Name == "_abck" {
				validator, _ := u.FindInString(cookie.Value, "~", "~")
				if validator == "-1" {
					if cancellationToken.Cancel {
						return nil
					}
					return HandleAkamaiTask(task, referer, baseEndpoint, akamaiEndpoint, baseURL)
				}
			}
		}
	}
	return nil
}
