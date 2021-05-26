package util

import (
	"bytes"
	"encoding/json"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"strings"
	"time"

	cclient "github.com/IHaveNothingg/cclientwtf"
	tls "github.com/refraction-networking/utls"
)

// CreateClient creates an HTTP client
func CreateClient() (http.Client, error) {
	client, err := cclient.NewClient(tls.HelloChrome_83)
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
	defer response.Body.Close()

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
