package util

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"regexp"
	"strings"

	"backend.juicedbot.io/juiced.client/client"
	"backend.juicedbot.io/juiced.client/http"
	"backend.juicedbot.io/juiced.client/http/cookiejar"
	"backend.juicedbot.io/juiced.client/utls"
	"backend.juicedbot.io/juiced.infrastructure/entities"
)

func AddBaseHeaders(request *http.Request) {
	request.UserAgent()
	request.Header.Set("Connection", "keep-alive")
	request.Header.Set("Sec-Ch-Ua", "\" Not A;Brand\";v=\"99\", \"Chromium\";v=\"90\", \"Google Chrome\";v=\"90\"")
	request.Header.Set("X-Application-Name", "web")
	request.Header.Set("Sec-Ch-Ua-Mobile", "?0")
	request.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/90.0.4430.93 Safari/537.36")
	request.Header.Set("Sec-Fetch-Site", "same-site")
	request.Header.Set("Sec-Fetch-Mode", "cors")
	request.Header.Set("Sec-Fetch-Dest", "empty")
	request.Header.Set("Accept-Language", "en-US,en;q=0.9")
}

func CreateClient(proxy ...*entities.Proxy) (http.Client, error) {
	var cClient http.Client
	var err error
	if len(proxy) > 0 {
		if proxy[0] != nil {
			proxy[0].AddCount()
			cClient, err = client.NewClient(utls.HelloChrome_90, entities.ProxyCleaner(*proxy[0]))
			if err != nil {
				return cClient, err
			}
		} else {
			cClient, _ = client.NewClient(utls.HelloChrome_90)
		}
	} else {
		cClient, _ = client.NewClient(utls.HelloChrome_90)
	}
	cookieJar, err := cookiejar.New(nil)
	if err != nil {
		return cClient, err
	}
	cClient.Jar = cookieJar
	return cClient, err
}

func HandleErrors(err error, errorType ErrorType) bool {
	if err != nil {
		switch errorType {
		case RequestMarshalBodyError:

		case RequestCreateError:

		case RequestDoError:

		case RequestReadBodyError:

		case RequestUnmarshalBodyError:

		case ShippingTypeError:

		}
		log.Println(err.Error())
		return false
	}
	return true
}

func MakeRequest(requestInfo *Request) (*http.Response, string, error) {
	var payload io.Reader

	if os.Getenv("JUICED_LOG") == "LOG" {
		log.Println("\n================\nSTART " + requestInfo.Method + " " + requestInfo.URL)
	}
	if requestInfo.Data != nil {
		if os.Getenv("JUICED_LOG") == "LOG" {
			log.Println("REQUEST BODY: " + string(requestInfo.Data) + "\n")
		}
		payload = bytes.NewBuffer(requestInfo.Data)
	} else {
		payload = nil
	}
	if requestInfo.RequestBodyStruct != nil {
		data, err := json.Marshal(requestInfo.RequestBodyStruct)
		ok := HandleErrors(err, RequestMarshalBodyError)
		if !ok {
			return nil, "", err
		}
		if os.Getenv("JUICED_LOG") == "LOG" {
			log.Println("REQUEST BODY: " + string(data) + "\n")
		}
		payload = bytes.NewBuffer(data)
	}
	request, err := http.NewRequest(requestInfo.Method, requestInfo.URL, payload)
	ok := HandleErrors(err, RequestCreateError)
	if !ok {
		return nil, "", err
	}

	if requestInfo.Headers != nil {
		request.Header = requestInfo.Headers
	}
	if requestInfo.RawHeaders != nil {
		request.RawHeader = requestInfo.RawHeaders
	}
	if requestInfo.AddHeadersFunction != nil {
		requestInfo.AddHeadersFunction(request, requestInfo.Referer)
	}

	if os.Getenv("JUICED_LOG") == "LOG" {
		log.Println("REQUEST HEADERS:")
		for header, values := range request.Header {
			log.Println(header + ": " + strings.Join(values, ","))
		}
		for _, header := range request.RawHeader {
			log.Println(header[0] + ": " + header[1])
		}
		log.Println()
	}

	var response *http.Response
	if requestInfo.Client.Transport != nil {
		response, err = requestInfo.Client.Do(request)
	} else {
		response, err = requestInfo.Scraper.Do(request)
	}
	ok = HandleErrors(err, RequestDoError)
	if !ok {
		return response, "", err
	}

	defer response.Body.Close()

	body, err := ioutil.ReadAll(response.Body)
	if os.Getenv("JUICED_LOG") == "LOG" {
		log.Println("RESPONSE STATUS CODE: " + fmt.Sprint(response.StatusCode))
		if response.Header.Get("Content-Type") != "text/html;charset=UTF-8" {
			log.Println("RESPONSE BODY: " + string(body) + "\n")
		}
		log.Println("RESPONSE HEADERS:")
		for header, values := range response.Header {
			log.Println(header + ": " + strings.Join(values, ","))
		}
		log.Println()
		log.Println("END " + requestInfo.URL + "\n================\n")
	}
	ok = HandleErrors(err, RequestReadBodyError)
	if !ok {
		return response, "", err
	}

	if requestInfo.ResponseBodyStruct != nil {
		err = json.Unmarshal(body, requestInfo.ResponseBodyStruct)
		if err != nil {
			return response, "", err
		}
	}

	newBody := strings.ReplaceAll(string(body), "\n", "")
	newBody = strings.ReplaceAll(newBody, "\t", "")

	return response, newBody, nil
}

func DetectCardType(cardNumber []byte) string {
	matched, _ := regexp.Match(`^4`, cardNumber)
	if matched {
		return "Visa"
	}

	matched, _ = regexp.Match(`^(5[1-5][0-9]{14}|2(22[1-9][0-9]{12}|2[3-9][0-9]{13}|[3-6][0-9]{14}|7[0-1][0-9]{13}|720[0-9]{12}))$`, cardNumber)
	if matched {
		return "Mastercard"
	}

	matched, _ = regexp.Match(`^3[47]`, cardNumber)
	if matched {
		return "AMEX"
	}

	matched, _ = regexp.Match(`^(6011|622(12[6-9]|1[3-9][0-9]|[2-8][0-9]{2}|9[0-1][0-9]|92[0-5]|64[4-9])|65)`, cardNumber)
	if matched {
		return "Discover"
	}

	matched, _ = regexp.Match(`^36`, cardNumber)
	if matched {
		return "Diners"
	}

	matched, _ = regexp.Match(`^30[0-5]`, cardNumber)
	if matched {
		return "Diners - Carte Blanche"
	}

	matched, _ = regexp.Match(`^35(2[89]|[3-8][0-9])`, cardNumber)
	if matched {
		return "JCB"
	}

	matched, _ = regexp.Match(`^(4026|417500|4508|4844|491(3|7))`, cardNumber)
	if matched {
		return "Visa Electron"
	}

	return ""
}
