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
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"regexp"
	"strings"
	"time"

	"backend.juicedbot.io/m/v2/juiced.infrastructure/common/entities"
	cclient "github.com/IHaveNothingg/cclientwtf"
	tls "github.com/Titanium-ctrl/utls"
)

// CreateClient creates an HTTP client
func CreateClient(proxy entities.Proxy) (http.Client, error) {

	client, err := cclient.NewClient(tls.HelloChrome_83, ProxyCleaner(proxy))
	if err != nil {
		return client, err
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
	request.Header.Set("X-Application-Name", "web") // TODO
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
		log.Println(err)
		return false
	}
	return true
}

// Request wraps the main functions of a request
// (marshal JSON, create request, execute request, read body, unmarshal JSON)
// while handling errors throughout
func MakeRequest(requestInfo *Request) (*http.Response, error) {
	var payload io.Reader

	log.Println("START " + requestInfo.URL)
	if requestInfo.Data != nil {
		log.Println("REQUEST BODY: " + string(requestInfo.Data))
		payload = bytes.NewBuffer(requestInfo.Data)
	} else {
		payload = nil
	}
	if requestInfo.RequestBodyStruct != nil {
		data, err := json.Marshal(requestInfo.RequestBodyStruct)
		ok := HandleErrors(err, RequestMarshalBodyError)
		if !ok {
			return nil, err
		}
		log.Println("REQUEST BODY: " + string(data))
		payload = bytes.NewBuffer(data)
	}
	request, err := http.NewRequest(requestInfo.Method, requestInfo.URL, payload)
	ok := HandleErrors(err, RequestCreateError)
	if !ok {
		return nil, err
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

	log.Println("REQUEST HEADERS:")
	for header, values := range request.Header {
		log.Println(header + ": " + strings.Join(values, ","))
	}
	response, err := requestInfo.Client.Do(request)
	ok = HandleErrors(err, RequestDoError)
	if !ok {
		return response, err
	}

	//Removing defer resp.Body.Close() here because it was causing problems, now it's after every MakeRequest

	body, err := ioutil.ReadAll(response.Body)
	log.Println("RESPONSE BODY: " + string(body))
	log.Println("RESPONSE HEADERS:")
	for header, values := range response.Header {
		log.Println(header + ": " + strings.Join(values, ","))
	}
	log.Println("END " + requestInfo.URL)
	ok = HandleErrors(err, RequestReadBodyError)
	if !ok {
		return response, err
	}

	if requestInfo.ResponseBodyStruct != nil {
		json.Unmarshal(body, requestInfo.ResponseBodyStruct) // TODO
	}

	return response, nil // TODO @silent: Return body
}

// SendDiscordWebhook sends checkout information to the Discord Webhook
func SendDiscordWebhook(discordWebhook string, success bool, fields []Field, imageURL string) bool {
	client := http.Client{
		Transport: &http.Transport{},
	}
	response, err := MakeRequest(&Request{
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

func ReadBody(resp *http.Response) string {
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Println(err)
	}
	defer resp.Body.Close()
	newBody := strings.ReplaceAll(string(body), "\n", "")
	newBody = strings.ReplaceAll(newBody, "\t", "")
	return newBody

}
