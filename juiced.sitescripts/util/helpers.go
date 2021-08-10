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

	cclient "backend.juicedbot.io/juiced.client/client"
	"backend.juicedbot.io/juiced.client/http"
	"backend.juicedbot.io/juiced.client/http/cookiejar"
	utls "backend.juicedbot.io/juiced.client/utls"
	"backend.juicedbot.io/juiced.infrastructure/commands"
	"backend.juicedbot.io/juiced.infrastructure/common/entities"
	"backend.juicedbot.io/juiced.infrastructure/common/enums"
	"backend.juicedbot.io/juiced.infrastructure/queries"
	sec "backend.juicedbot.io/juiced.security/auth/util"
	"backend.juicedbot.io/juiced.sitescripts/base"
)

// CreateClient creates an HTTP client
func CreateClient(proxy ...entities.Proxy) (http.Client, error) {
	var client http.Client
	var err error
	if len(proxy) > 0 {
		client, err = cclient.NewClient(utls.HelloChrome_90, ProxyCleaner(proxy[0]))
		if err != nil {
			return client, err
		}
	} else {
		client, err = cclient.NewClient(utls.HelloChrome_90)
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

var hookChan = make(chan HookInfo)

func QueueWebhook(success bool, content string, embeds []Embed) {
	hookChan <- HookInfo{
		Success: success,
		Content: content,
		Embeds:  embeds,
	}
}

func DiscordWebhookQueue() {
	for {
		hook := <-hookChan
		settings, err := queries.GetSettings()
		if err != nil {
			return
		}
		var webhookURL string
		if hook.Success {
			webhookURL = settings.SuccessDiscordWebhook
		} else {
			webhookURL = settings.FailureDiscordWebhook
		}
		if webhookURL != "" {
			SendDiscordWebhook(webhookURL, hook.Embeds)
		}
		time.Sleep(2*time.Second + (time.Second / 2))
	}
}

// SendDiscordWebhook sends checkout information to the Discord Webhook
func SendDiscordWebhook(discordWebhook string, embeds []Embed) bool {
	client := http.Client{
		Transport: &http.Transport{},
	}
	response, body, err := MakeRequest(&Request{
		Client: client,
		Method: "POST",
		URL:    discordWebhook,
		AddHeadersFunction: func(request *http.Request, e ...string) {
			request.Header.Set("content-type", "application/json")
		},
		RequestBodyStruct: DiscordWebhook{
			Content: nil,
			Embeds:  embeds,
		},
	})
	if err != nil {
		return false
	}

	fmt.Println(string(body))
	return response.StatusCode >= 200 && response.StatusCode < 300
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

// Function to generate valid abck cookies using an API
func NewAbck(abckClient *http.Client, location string, BaseEndpoint, AkamaiEndpoint string) error {
	var ParsedBase, _ = url.Parse(BaseEndpoint)

	_, user, err := queries.GetUserInfo()
	if err != nil {
		return err
	}

	var abckCookie string
	var genResponse sec.ExperimentalAkamaiAPIResponse
	for _, cookie := range abckClient.Jar.Cookies(ParsedBase) {
		if cookie.Name == "_abck" {
			abckCookie = cookie.Value
		}
	}

	if abckCookie == "" {
		_, _, err := MakeRequest(&Request{
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
	}

	genResponse, _, err = sec.ExperimentalAkamai(location, "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/92.0.4515.107 Safari/537.36", abckCookie, 0, 0, 0, 0, user)
	if err != nil {
		return err
	}

	sensorRequest := SensorRequest{
		SensorData: genResponse.SensorData,
	}

	data, err := json.Marshal(sensorRequest)
	if err != nil {
		return err
	}

	sensorResponse := SensorResponse{}
	resp, _, err := MakeRequest(&Request{
		Client: *abckClient,
		Method: "POST",
		URL:    AkamaiEndpoint,
		RawHeaders: [][2]string{
			{"content-length", fmt.Sprint(len(data))},
			{"sec-ch-ua", "\" Not A;Brand\";v=\"99\", \"Chromium\";v=\"90\", \"Google Chrome\";v=\"90\""},
			{"sec-ch-ua-mobile", "?0"},
			{"user-agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/92.0.4515.107 Safari/537.36"},
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
		Data:               data,
		ResponseBodyStruct: &sensorResponse,
	})
	if err != nil {
		return err
	}
	if !sensorResponse.Success || resp.StatusCode != 201 {
		return errors.New("bad sensor")
	}

	for _, cookie := range abckClient.Jar.Cookies(ParsedBase) {
		if cookie.Name == "_abck" {
			abckCookie = cookie.Value
		}
	}

	genResponse, _, err = sec.ExperimentalAkamai(location, "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/92.0.4515.107 Safari/537.36", abckCookie, 1, genResponse.SavedD3, genResponse.SavedStartTS, genResponse.DeviceNum, user)
	if err != nil {
		return err
	}

	sensorRequest = SensorRequest{
		SensorData: genResponse.SensorData,
	}
	data, _ = json.Marshal(sensorRequest)

	sensorResponse = SensorResponse{}
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
		Data:               data,
		ResponseBodyStruct: &sensorResponse,
	})
	if err != nil {
		return err
	}
	if !sensorResponse.Success {
		return errors.New("bad sensor")
	}

	switch resp.StatusCode {
	case 201:
		if ParsedBase.Host == "www.gamestop.com" {
			if len(abckCookie) > 488 {
				return nil
			}
		} else {
			for _, cookie := range abckClient.Jar.Cookies(ParsedBase) {
				if cookie.Name == "_abck" {
					validator, _ := FindInString(cookie.Value, "~", "~")
					if validator == "-1" {
						NewAbck(abckClient, location, BaseEndpoint, AkamaiEndpoint)
					}

				}
			}
		}

		return nil
	}
	return errors.New(resp.Status)
}

func SecToUtil(secEmbeds []sec.DiscordEmbed) (embeds []Embed) {
	for _, secEmbed := range secEmbeds {
		tempEmbed := Embed{
			Title: secEmbed.Title,
			Color: secEmbed.Color,
			Footer: Footer{
				Text:    secEmbed.Footer.Text,
				IconURL: secEmbed.Footer.IconURL,
			},
			Timestamp: secEmbed.Timestamp,
			Thumbnail: Thumbnail{
				URL: secEmbed.Thumbnail.URL,
			},
		}
		for _, secField := range secEmbed.Fields {
			tempEmbed.Fields = append(tempEmbed.Fields, Field{
				Name:   secField.Name,
				Value:  secField.Value,
				Inline: secField.Inline,
			})
		}
		embeds = append(embeds, tempEmbed)

	}
	return
}

// Processes each checkout by sending a webhook and logging the checkout
func ProcessCheckout(pci ProcessCheckoutInfo) {
	_, user, err := queries.GetUserInfo()
	if err != nil {
		fmt.Println("Could not get user info")
		return
	}
	pci.UserInfo = user
	if pci.Status != enums.OrderStatusFailed {
		go sec.DiscordWebhook(pci.Success, pci.Content, pci.Embeds, pci.UserInfo)
	}
	if pci.Success {
		go sec.LogCheckout(pci.ItemName, pci.Sku, pci.Retailer, int(pci.Price), pci.Quantity, pci.UserInfo)
		go SendCheckout(pci.BaseTask, pci.ItemName, pci.ImageURL, pci.Sku, int(pci.Price), pci.Quantity, pci.MsToCheckout)
	}
	QueueWebhook(pci.Success, pci.Content, SecToUtil(pci.Embeds))
}

// Logs the checkout
func SendCheckout(task base.Task, itemName string, imageURL string, sku string, price int, quantity int, msToCheckout int64) {
	commands.CreateCheckout(entities.Checkout{
		ItemName:     itemName,
		ImageURL:     imageURL,
		SKU:          sku,
		Price:        price,
		Quantity:     quantity,
		Retailer:     task.Task.TaskRetailer,
		ProfileName:  task.Profile.Name,
		MsToCheckout: msToCheckout,
		Time:         time.Now().Unix(),
	})
}

func GetPXCookie(site string, proxy entities.Proxy, cancellationToken *CancellationToken) (string, PXValues, bool, error) {
	var pxValues PXValues

	_, userInfo, err := queries.GetUserInfo()
	if err != nil {
		return "", pxValues, false, err
	}

	pxResponse, _, err := sec.PX(site, ProxyCleaner(proxy), userInfo)
	if err != nil {
		return "", pxValues, false, err
	}

	if pxResponse.PX3 == "" || pxResponse.SetID == "" || pxResponse.UUID == "" || pxResponse.VID == "" {
		if cancellationToken.Cancel {
			return "", pxValues, true, err
		}
		return GetPXCookie(site, proxy, cancellationToken)
	}

	return pxResponse.PX3, PXValues{
		SetID: pxResponse.SetID,
		UUID:  pxResponse.UUID,
		VID:   pxResponse.VID,
	}, false, nil
}

func GetPXCapCookie(site, setID, vid, uuid, token string, proxy entities.Proxy, cancellationToken *CancellationToken) (string, bool, error) {
	var px3 string

	_, userInfo, err := queries.GetUserInfo()
	if err != nil {
		return px3, false, err
	}

	px3, _, err = sec.PXCap(site, ProxyCleaner(proxy), setID, vid, uuid, token, userInfo)
	if err != nil {
		return "", false, err
	}
	if px3 == "" {
		if cancellationToken.Cancel {
			return "", true, err
		}
		return GetPXCapCookie(site, setID, vid, uuid, token, proxy, cancellationToken)
	}
	return px3, false, nil
}

// Returns the value of a cookie with the given cookieName and url
func GetCookie(client http.Client, uri string, cookieName string) (string, error) {
	u, err := url.Parse(uri)
	if err != nil {
		return "", err
	}
	for _, cookie := range client.Jar.Cookies(u) {
		if cookie.Name == cookieName {
			return cookie.Value, nil
		}
	}
	return "", errors.New("no cookie with name: " + cookieName)
}

func GetCardType(cardNumber []byte, retailer enums.Retailer) string {
	matched, _ := regexp.Match(`^4`, cardNumber)
	if matched {
		switch retailer {
		case enums.Walmart:
			return "VISA"
		case enums.BestBuy:
			return "VISA"
		default:
			return "Visa"
		}
	}

	matched, _ = regexp.Match(`^(5[1-5][0-9]{14}|2(22[1-9][0-9]{12}|2[3-9][0-9]{13}|[3-6][0-9]{14}|7[0-1][0-9]{13}|720[0-9]{12}))$`, cardNumber)
	if matched {
		switch retailer {
		case enums.GameStop:
			return "MasterCard"
		case enums.Walmart:
			return "MASTERCARD"
		default:
			return "Mastercard"
		}
	}

	matched, _ = regexp.Match(`^3[47]`, cardNumber)
	if matched {
		switch retailer {

		default:
			return "AMEX"
		}
	}

	matched, _ = regexp.Match(`^(6011|622(12[6-9]|1[3-9][0-9]|[2-8][0-9]{2}|9[0-1][0-9]|92[0-5]|64[4-9])|65)`, cardNumber)
	if matched {
		switch retailer {
		case enums.Walmart:
			return "DISCOVER"
		default:
			return "Discover"
		}
	}

	matched, _ = regexp.Match(`^36`, cardNumber)
	if matched {
		switch retailer {
		case enums.GameStop:
		case enums.Walmart:
		default:
			return "Diners"
		}
	}

	matched, _ = regexp.Match(`^30[0-5]`, cardNumber)
	if matched {
		switch retailer {
		case enums.BestBuy:
		case enums.BoxLunch:
		case enums.GameStop:
			return "MasterCard"
		case enums.HotTopic:
		case enums.Walmart:
		default:
			return "Diners - Carte Blanche"
		}
	}

	matched, _ = regexp.Match(`^35(2[89]|[3-8][0-9])`, cardNumber)
	if matched {
		switch retailer {
		case enums.BoxLunch:
		case enums.GameStop:
			return "Unknown"
		case enums.HotTopic:
		case enums.Target:
		case enums.Walmart:
		default:
			return "JCB"
		}
	}

	matched, _ = regexp.Match(`^(4026|417500|4508|4844|491(3|7))`, cardNumber)
	if matched {
		switch retailer {
		default:
			return "Visa Electron"
		}
	}

	return ""
}
