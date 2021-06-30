package captcha

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strings"
	"time"

	"backend.juicedbot.io/juiced.infrastructure/commands"
	"backend.juicedbot.io/juiced.infrastructure/common/entities"
	"backend.juicedbot.io/juiced.infrastructure/common/enums"
)

type CaptchaAPIError struct {
	CaptchaType enums.CaptchaType
	CaptchaAPI  enums.CaptchaAPI
	Err         error
}

func (r *CaptchaAPIError) Error() string {
	return fmt.Sprintf("%s encountered an error requesting a %s Captcha: %v", r.CaptchaType, r.CaptchaAPI, r.Err)
}

func KeyErrors(settings entities.Settings, keyError KeyError, captchaType enums.CaptchaType) error {
	if keyError == CaptchaSuccess {
		return nil
	}

	switch keyError {
	case BadTwoCapKeyError:
		settings.TwoCaptchaAPIKey = ""
		_, err := commands.UpdateSettings(settings)
		if err != nil {
			return err
		}
	case BadAntiCapKeyError:
		settings.AntiCaptchaAPIKey = ""
		_, err := commands.UpdateSettings(settings)
		if err != nil {
			return err
		}
	case BadCapMonKeyError:
		settings.CapMonsterAPIKey = ""
		_, err := commands.UpdateSettings(settings)
		if err != nil {
			return err
		}
	}

	return &CaptchaAPIError{
		CaptchaType: captchaType,
		CaptchaAPI:  enums.TwoCaptcha,
		Err:         errors.New(keyError),
	}
}

func CheckServices(settings entities.Settings) (captchaServices []string, captchaServiceCount int) {
	if settings.TwoCaptchaAPIKey != "" {
		captchaServices = append(captchaServices, settings.TwoCaptchaAPIKey)
		captchaServiceCount++
	}
	if settings.AntiCaptchaAPIKey != "" {
		captchaServices = append(captchaServices, settings.AntiCaptchaAPIKey)
		captchaServiceCount++
	}
	if settings.CapMonsterAPIKey != "" {
		captchaServices = append(captchaServices, settings.CapMonsterAPIKey)
		captchaServiceCount++
	}
	if settings.AYCDAccessToken != "" && settings.AYCDAPIKey != "" {
		captchaServices = append(captchaServices, settings.AYCDAccessToken+"|"+settings.AYCDAPIKey)
	}
	return
}

func TwoCaptchaReq(apiKey string, uri string) (string, error) {
	var token string
	resp, err := http.DefaultClient.Get(uri)
	if err != nil {
		log.Println("82: " + err.Error())
		return token, err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Println("89: " + err.Error())
		return token, err
	}

	splitResp := strings.Split(string(body), "|")

	if len(splitResp) == 1 || splitResp[0] != "OK" {
		log.Println("96: " + splitResp[0])
		// All possible errors: https://2captcha.com/2captcha-api#error_handling
		return token, errors.New(splitResp[0])
	}

	var ready bool

	for !ready {
		resp, err := http.DefaultClient.Get("https://2captcha.com/res.php?key=" + apiKey + "&action=get&id=" + splitResp[1])
		if err != nil {
			return token, err
		}
		defer resp.Body.Close()

		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return token, err
		}

		if strings.Contains(string(body), "challenge") {
			twoCaptchaGeeTestResponse := TwoCaptchaGeeTestResponse{}
			err = json.Unmarshal(body, &twoCaptchaGeeTestResponse)
			if err != nil {
				return token, err
			}
			token = twoCaptchaGeeTestResponse.Challenge + "|" + twoCaptchaGeeTestResponse.Validate + "|" + twoCaptchaGeeTestResponse.Seccode
		} else {
			splitResp := strings.Split(string(body), "|")
			if len(splitResp) == 1 {
				if splitResp[0] != "CAPCHA_NOT_READY" {
					return token, errors.New(splitResp[0])
				}
				time.Sleep(1 * time.Second)
			} else {
				if splitResp[0] != "OK" {
					return token, errors.New(splitResp[0])
				}
				token = splitResp[1]
				ready = true
			}
		}
	}
	return token, err
}

func AntiCaptchaReq(apiKey string, taskInfo AntiCaptchaTaskInfo) (AntiCaptchaResponse, error) {
	antiCaptchaResponse := AntiCaptchaResponse{}

	payloadBytes, err := json.Marshal(AntiCaptchaTask{
		Clientkey: apiKey,
		Task:      taskInfo,
	})
	if err != nil {
		log.Println("147: " + err.Error())
		return antiCaptchaResponse, err
	}
	data := bytes.NewReader(payloadBytes)

	req, err := http.NewRequest("POST", "https://api.anti-captcha.com/createTask", data)
	if err != nil {
		log.Println("154: " + err.Error())
		return antiCaptchaResponse, err
	}
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		log.Println("162: " + err.Error())
		return antiCaptchaResponse, err
	}
	defer resp.Body.Close()

	antiCaptchaStart := AntiCaptchaStart{}
	body, _ := ioutil.ReadAll(resp.Body)
	err = json.Unmarshal(body, &antiCaptchaStart)
	if err != nil {
		log.Println("171: " + err.Error())
		return antiCaptchaResponse, err
	}

	if antiCaptchaStart.Errorid != 0 {
		log.Println("176: " + antiCaptchaStart.ErrorCode)
		return antiCaptchaResponse, errors.New(antiCaptchaStart.ErrorCode)
	}
	payloadBytes, err = json.Marshal(AntiCaptchaRequest{
		Clientkey: apiKey,
		Taskid:    antiCaptchaStart.TaskID,
	})
	if err != nil {
		log.Println("184: " + err.Error())
		return antiCaptchaResponse, err
	}
	data = bytes.NewReader(payloadBytes)

	req, err = http.NewRequest("POST", "https://api.anti-captcha.com/getTaskResult", data)
	if err != nil {
		log.Println("191: " + err.Error())
		return antiCaptchaResponse, err
	}
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Content-Type", "application/json")

	var ready bool
	for !ready {
		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			return antiCaptchaResponse, err
		}

		body, _ := ioutil.ReadAll(resp.Body)
		err = json.Unmarshal(body, &antiCaptchaResponse)
		if err != nil {
			return antiCaptchaResponse, err
		}

		resp.Body.Close()
		if antiCaptchaResponse.Errorid != 0 {
			return antiCaptchaResponse, errors.New(antiCaptchaResponse.ErrorCode)
		}
		if antiCaptchaResponse.Status == "ready" {
			ready = true
		} else {
			antiCaptchaResponse = AntiCaptchaResponse{}
			time.Sleep(1 * time.Second)
		}
	}

	return antiCaptchaResponse, err
}

func CapMonsterReq(apiKey string, taskInfo CapMonsterTaskInfo) (CapMonsterResponse, error) {
	capMonsterResponse := CapMonsterResponse{}

	payloadBytes, err := json.Marshal(CapMonsterTask{
		Clientkey: apiKey,
		Task:      taskInfo,
	})
	if err != nil {
		log.Println("226: " + err.Error())
		return capMonsterResponse, err
	}
	data := bytes.NewReader(payloadBytes)

	req, err := http.NewRequest("POST", "https://api.capmonster.cloud/createTask", data)
	if err != nil {
		log.Println("233: " + err.Error())
		return capMonsterResponse, err
	}
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		log.Println("241: " + err.Error())
		return capMonsterResponse, err
	}
	defer resp.Body.Close()

	capMonsterStart := CapMonsterStart{}
	body, _ := ioutil.ReadAll(resp.Body)
	err = json.Unmarshal(body, &capMonsterStart)
	if err != nil {
		log.Println("250: " + err.Error())
		return capMonsterResponse, err
	}

	if capMonsterStart.Errorid != 0 {
		log.Println("255: " + capMonsterStart.Errorcode)
		return capMonsterResponse, errors.New(capMonsterStart.Errorcode)
	}
	payloadBytes, err = json.Marshal(AntiCaptchaRequest{
		Clientkey: apiKey,
		Taskid:    capMonsterStart.TaskID,
	})
	if err != nil {
		log.Println("263: " + err.Error())
		return capMonsterResponse, err
	}
	data = bytes.NewReader(payloadBytes)

	req, err = http.NewRequest("POST", "https://api.capmonster.cloud/getTaskResult ", data)
	if err != nil {
		log.Println("270: " + err.Error())
		return capMonsterResponse, err
	}
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Content-Type", "application/json")

	var ready bool
	for !ready {
		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			return capMonsterResponse, err
		}

		body, _ := ioutil.ReadAll(resp.Body)
		log.Println(body)
		err = json.Unmarshal(body, &capMonsterResponse)
		if err != nil {
			return capMonsterResponse, err
		}

		resp.Body.Close()
		if capMonsterResponse.Status == "ready" {
			ready = true
		} else {
			capMonsterResponse = CapMonsterResponse{}
			time.Sleep(1 * time.Second)
		}
	}

	return capMonsterResponse, err
}

func ParseProxy(proxyStr string) (entities.Proxy, error) {
	proxy := entities.Proxy{}

	proxySplit := strings.Split(proxyStr, ":")
	if len(proxySplit) != 2 && len(proxySplit) != 4 {
		return proxy, errors.New("bad proxy")
	}

	username := ""
	if len(proxySplit) == 4 {
		username = proxySplit[2]
	}
	password := ""
	if len(proxySplit) == 4 {
		password = proxySplit[3]
	}

	proxy = entities.Proxy{
		Host:     proxySplit[0],
		Port:     proxySplit[1],
		Username: username,
		Password: password,
	}

	return proxy, nil
}
