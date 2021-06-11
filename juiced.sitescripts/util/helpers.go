package util

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"math/rand"
	"net/url"
	"os"
	"regexp"
	"strings"
	"time"

	"backend.juicedbot.io/juiced.client/http"
	"backend.juicedbot.io/juiced.client/http/cookiejar"
	sec "backend.juicedbot.io/juiced.security/auth/util"
	"backend.juicedbot.io/juiced.sitescripts/base"

	cclient "backend.juicedbot.io/juiced.client/client"
	"backend.juicedbot.io/juiced.infrastructure/commands"
	"backend.juicedbot.io/juiced.infrastructure/common/entities"
	"backend.juicedbot.io/juiced.infrastructure/queries"
	tls "github.com/Titanium-ctrl/utls"
)

// CreateClient creates an HTTP client
func CreateClient(proxy ...entities.Proxy) (http.Client, error) {
	var client http.Client
	var err error
	if len(proxy) > 0 {
		client, err = cclient.NewClient(tls.HelloChrome_83, ProxyCleaner(proxy[0]))
		if err != nil {
			return client, err
		}
	} else {
		client, err = cclient.NewClient(tls.HelloChrome_83)
		if err != nil {
			return client, err
		}
	}
	cookieJar, err := cookiejar.New(nil)
	if err != nil {
		return client, err
	}
	client.Jar = cookieJar
	return client, err
}

// Adds base headers to the request
func AddBaseHeaders(request *http.Request) {
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

// HandleErrors handles errors
// TODO: Actually handle errors
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

// Request wraps the main functions of a request
// (marshal JSON, create request, execute request, read body, unmarshal JSON)
// while handling errors throughout
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
	response, err := requestInfo.Client.Do(request)
	ok = HandleErrors(err, RequestDoError)
	if !ok {
		return response, "", err
	}

	defer response.Body.Close()

	body, err := ioutil.ReadAll(response.Body)
	if os.Getenv("JUICED_LOG") == "LOG" {
		log.Println("RESPONSE BODY: " + string(body) + "\n")
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

// SendDiscordWebhook sends checkout information to the Discord Webhook
func SendDiscordWebhook(discordWebhook string, success bool, fields []Field, imageURL string) bool {
	client := http.Client{
		Transport: &http.Transport{},
	}
	response, _, err := MakeRequest(&Request{
		Client: client,
		Method: "POST",
		URL:    discordWebhook,
		AddHeadersFunction: func(request *http.Request, e ...string) {
			request.Header.Set("content-type", "application/json")
		},
		RequestBodyStruct: CreateDiscordWebhook(success, fields, imageURL),
	})
	if err != nil {
		return false
	}
	return response.StatusCode >= 200 && response.StatusCode < 300
}

// CreateDiscordWebhook creates a DiscordWebhook struct
func CreateDiscordWebhook(success bool, fields []Field, imageURL string) DiscordWebhook {
	webhook := DiscordWebhook{
		Content: nil,
		Embeds: []Embed{
			{
				Fields: fields,
				Footer: Footer{
					Text:    "Juiced AIO",
					IconURL: "https://cdn.discordapp.com/icons/688572290488991757/b684ee4e3cfb661d32afc48f24776e60.png?size=128",
				},
				Timestamp: time.Now(),
			},
		},
	}
	switch success {
	case true:
		webhook.Embeds[0].Title = ":tangerine: Checkout! :tangerine:"
		webhook.Embeds[0].Color = 16742912
		webhook.Embeds[0].Thumbnail = Thumbnail{
			URL: imageURL,
		}
	case false:
		webhook.Embeds[0].Title = ":lemon: Failed to Place Order :lemon:"
		webhook.Embeds[0].Color = 16766464
	}
	return webhook
}

// CreateParams turns a string->string map into a URL parameter string
func CreateParams(paramsLong map[string]string) string {
	params := url.Values{}
	for key, value := range paramsLong {
		params.Add(key, value)
	}
	return params.Encode()
}

// TernaryOperator is a make-shift ternary operator since Golang doesn't have one out of the box
func TernaryOperator(condition bool, trueOutcome interface{}, falseOutcome interface{}) interface{} {
	if condition {
		return trueOutcome
	}
	return falseOutcome
}

func ProxyCleaner(proxyDirty entities.Proxy) string {
	if proxyDirty.Host == "" {
		return ""
	}
	if proxyDirty.Username == "" && proxyDirty.Password == "" {
		return fmt.Sprintf("http://%s:%s", proxyDirty.Host, proxyDirty.Port)
	} else {
		return fmt.Sprintf("http://%s:%s@%s:%s", proxyDirty.Username, proxyDirty.Password, proxyDirty.Host, proxyDirty.Port)
	}

}

func FindInString(str string, start string, end string) (string, error) {
	comp := regexp.MustCompile(fmt.Sprintf("%v(.*?)%v", start, end))
	comp.MatchString(str)

	o := comp.FindStringSubmatch(str)
	if len(o) == 0 {
		return "", errors.New("string not found")
	}
	parsed := o[1]
	if parsed == "" {
		return parsed, errors.New("string not found")
	}

	return parsed, nil

}

func RandomNumberInt(min int, max int) int64 {
	rand.Seed(time.Now().UnixNano())
	a := int64(rand.Intn(max-min) + min)
	return a
}

func Randomizer(s string) string {
	nums := RandomNumberInt(0, 100)
	if nums < 50 {
		return s
	}
	return ""

}

// Returns true if it finds the string x in the slice s
func InSlice(s []string, x string) bool {
	for _, i := range s {
		if i == x {
			return true
		}
	}
	return false
}

// Removes the string x from the slice s
func RemoveFromSlice(s []string, x string) []string {
	var position int
	for i, r := range s {
		if r == x {
			position = i
		}
	}
	if position == 0 {
		return s
	}
	s[position] = s[len(s)-1]

	return s[:len(s)-1]
}

// Function to generate valid abck cookies using an API
func NewAbck(abckClient *http.Client, location string, BaseEndpoint, AkamaiEndpoint string) error {
	var ParsedBase, _ = url.Parse(BaseEndpoint)
	ver := "1.7"
	if ParsedBase.Host == "www.gamestop.com" {
		ver = "1.69"
	}
	_, user, err := queries.GetUserInfo()
	if err != nil {
		return err
	}

	resp, _, err := MakeRequest(&Request{
		Client: *abckClient,
		Method: "GET",
		URL:    AkamaiEndpoint,
		RawHeaders: [][2]string{
			{"sec-ch-ua", "\" Not A;Brand\";v=\"99\", \"Chromium\";v=\"90\", \"Google Chrome\";v=\"90\""},
			{"sec-ch-ua-mobile", "?0"},
			{"user-agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/90.0.4430.212 Safari/537.36"},
			{"content-type", "text/plain;charset=UTF-8"},
			{"accept", "*/*"},
			{"origin", BaseEndpoint},
			{"sec-fetch-site", "same-origin"},
			{"sec-fetch-mode", "cors"},
			{"sec-fetch-dest", "empty"},
			{"referer", location},
			{"accept-encoding", "gzip, deflate, br"},
			{"accept-language", "en-US,en;q=0.9"},
		},
	})
	if err != nil {
		return err
	}

	var abckCookie string
	var genResponse sec.AkamaiAPIResponse
	for _, cookie := range abckClient.Jar.Cookies(ParsedBase) {
		if cookie.Name == "_abck" {
			abckCookie = cookie.Value
		}
	}

	genResponse, _, err = sec.Akamai(location, "true", "true", "false", "false", abckCookie, AkamaiEndpoint, ver, "true", "", "", "true", user)
	if err != nil {
		return err
	}

	sensorRequest := SensorRequest{
		SensorData: genResponse.SensorData,
	}

	data, _ := json.Marshal(sensorRequest)
	resp, _, err = MakeRequest(&Request{
		Client: *abckClient,
		Method: "POST",
		URL:    AkamaiEndpoint,
		RawHeaders: [][2]string{
			{"content-length", fmt.Sprint(len(data))},
			{"sec-ch-ua", "\" Not A;Brand\";v=\"99\", \"Chromium\";v=\"90\", \"Google Chrome\";v=\"90\""},
			{"sec-ch-ua-mobile", "?0"},
			{"user-agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/90.0.4430.212 Safari/537.36"},
			{"content-type", "text/plain;charset=UTF-8"},
			{"accept", "*/*"},
			{"origin", BaseEndpoint},
			{"sec-fetch-site", "same-origin"},
			{"sec-fetch-mode", "cors"},
			{"sec-fetch-dest", "empty"},
			{"referer", location},
			{"accept-encoding", "gzip, deflate, br"},
			{"accept-language", "en-US,en;q=0.9"},
		},
		Data: data,
	})
	if err != nil {
		return err
	}

	for _, cookie := range abckClient.Jar.Cookies(ParsedBase) {
		if cookie.Name == "_abck" {
			abckCookie = cookie.Value
		}
	}

	genResponse, _, err = sec.Akamai(location, "true", "false", "false", "false", abckCookie, AkamaiEndpoint, ver, "false", "", "", "true", user)
	if err != nil {
		return err
	}

	sensorRequest = SensorRequest{
		SensorData: genResponse.SensorData,
	}
	data, _ = json.Marshal(sensorRequest)

	resp, _, err = MakeRequest(&Request{
		Client: *abckClient,
		Method: "POST",
		URL:    AkamaiEndpoint,
		RawHeaders: [][2]string{
			{"content-length", fmt.Sprint(len(data))},
			{"sec-ch-ua", "\" Not A;Brand\";v=\"99\", \"Chromium\";v=\"90\", \"Google Chrome\";v=\"90\""},
			{"sec-ch-ua-mobile", "?0"},
			{"user-agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/90.0.4430.212 Safari/537.36"},
			{"content-type", "text/plain;charset=UTF-8"},
			{"accept", "*/*"},
			{"origin", BaseEndpoint},
			{"sec-fetch-site", "same-origin"},
			{"sec-fetch-mode", "cors"},
			{"sec-fetch-dest", "empty"},
			{"referer", location},
			{"accept-encoding", "gzip, deflate, br"},
			{"accept-language", "en-US,en;q=0.9"},
		},
		Data: data,
	})
	if err != nil {
		return err
	}

	for _, cookie := range abckClient.Jar.Cookies(ParsedBase) {
		if cookie.Name == "_abck" {
			abckCookie = cookie.Value
		}
	}

	genResponse, _, err = sec.Akamai(location, "true", "false", "false", "false", abckCookie, AkamaiEndpoint, ver, "false", "", "", "true", user)
	if err != nil {
		return err
	}

	sensorRequest = SensorRequest{
		SensorData: genResponse.SensorData,
	}

	data, err = json.Marshal(sensorRequest)
	if err != nil {
		return err
	}

	resp, _, err = MakeRequest(&Request{
		Client: *abckClient,
		Method: "POST",
		URL:    AkamaiEndpoint,
		RawHeaders: [][2]string{
			{"content-length", fmt.Sprint(len(data))},
			{"sec-ch-ua", "\" Not A;Brand\";v=\"99\", \"Chromium\";v=\"90\", \"Google Chrome\";v=\"90\""},
			{"sec-ch-ua-mobile", "?0"},
			{"user-agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/90.0.4430.212 Safari/537.36"},
			{"content-type", "text/plain;charset=UTF-8"},
			{"accept", "*/*"},
			{"origin", BaseEndpoint},
			{"sec-fetch-site", "same-origin"},
			{"sec-fetch-mode", "cors"},
			{"sec-fetch-dest", "empty"},
			{"referer", location},
			{"accept-encoding", "gzip, deflate, br"},
			{"accept-language", "en-US,en;q=0.9"},
		},
		Data: data,
	})
	if err != nil {
		return err
	}

	switch resp.StatusCode {
	case 201:
		for _, cookie := range abckClient.Jar.Cookies(ParsedBase) {
			if cookie.Name == "_abck" {
				fmt.Println(cookie.Value)
				validator, _ := FindInString(cookie.Value, "~", "~")
				if validator == "-1" {
					NewAbck(abckClient, location, BaseEndpoint, AkamaiEndpoint)
				}

			}
		}
		return nil
	}
	return errors.New(resp.Status)
}

// Processes each checkout by sending a webhook and logging the checkout
func ProcessCheckout(pci ProcessCheckoutInfo) {
	sec.DiscordWebhook(pci.Success, pci.Content, pci.Embeds, pci.UserInfo)
	SendCheckout(pci.BaseTask, pci.ItemName, pci.Sku, pci.Price, pci.Quantity)
}

// Logs the checkout
func SendCheckout(task base.Task, itemName string, sku string, price int, quantity int) {
	commands.CreateCheckout(entities.Checkout{
		ItemName:    itemName,
		SKU:         sku,
		Price:       price,
		Quantity:    quantity,
		Retailer:    task.Task.TaskRetailer,
		ProfileName: task.Profile.Name,
		Time:        time.Now(),
	})
}
