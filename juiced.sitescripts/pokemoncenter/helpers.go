package pokemoncenter

import (
	"context"
	"crypto/rsa"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/url"
	"strings"
	"time"

	"backend.juicedbot.io/juiced.client/http"
	"backend.juicedbot.io/juiced.infrastructure/common"
	"backend.juicedbot.io/juiced.infrastructure/common/captcha"
	"backend.juicedbot.io/juiced.infrastructure/common/entities"
	"backend.juicedbot.io/juiced.infrastructure/common/enums"
	sec "backend.juicedbot.io/juiced.security/auth/util"
	"backend.juicedbot.io/juiced.sitescripts/util"
	jose "github.com/dvsekhvalnov/jose2go"
	"github.com/lestrrat-go/jwx/jwk"
)

// AddPokemonCenterHeaders adds PokemonCenter headers to the request
func AddPokemonCenterHeaders(request *http.Request, referer ...string) {
	util.AddBaseHeaders(request)
	request.Header.Set("Accept", "application/json")
	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("Origin", "https://www.pokemoncenter.com")
	// omitcsrfjwt: true
	// omitcorrelationid: true
	// credentials: include
	// TODO: Header order
	if len(referer) != 0 {
		request.Header.Set("Referer", referer[0])
	}
}

func SetDatadomeCookie(datadomeInfo DatadomeInfo, proxy *entities.Proxy, client *http.Client, cancellationToken *util.CancellationToken) error {
	params := common.CreateParams(map[string]string{
		"initialCid": datadomeInfo.InitialCID,
		"hash":       datadomeInfo.Hash,
		"cid":        datadomeInfo.CID,
		"t":          datadomeInfo.T,
		"referer":    "https://www.pokemoncenter.com/",
		"s":          fmt.Sprint(datadomeInfo.S),
	})
	proxy_ := entities.Proxy{}
	if proxy != nil {
		proxy_ = *proxy
	}
	token, err := captcha.RequestCaptchaToken(enums.ReCaptchaV2, enums.PokemonCenter, DatadomeEndpoint+params, "", 0, proxy_)
	if err != nil {
		return err
	}
	for token == nil {
		if cancellationToken.Cancel {
			return nil
		}
		token = captcha.PollCaptchaTokens(enums.ReCaptchaV2, enums.PokemonCenter, DatadomeEndpoint+params, proxy_)
		time.Sleep(1 * time.Second / 10)
	}

	tokenInfo, ok := token.(entities.ReCaptchaToken)
	if !ok {
		return errors.New("token is not ReCaptchaToken")
	}

	params = common.CreateParams(map[string]string{
		"icid":                 datadomeInfo.InitialCID,
		"hash":                 datadomeInfo.Hash,
		"cid":                  datadomeInfo.CID,
		"t":                    datadomeInfo.T,
		"referer":              "https://www.pokemoncenter.com/",
		"s":                    fmt.Sprint(datadomeInfo.S),
		"parent_url":           "https://www.pokemoncenter.com",
		"ua":                   "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/15.0 Safari/605.1.15",
		"ccid":                 "null",
		"x-forwarded-for":      "",
		"captchaChallenge":     "",
		"g-recaptcha-response": tokenInfo.Token,
	})

	resp, body, err := util.MakeRequest(&util.Request{
		Client: *client,
		Method: "GET",
		URL:    DatadomeChallengeEndpoint + params,
		RawHeaders: [][2]string{
			{"sec-ch-ua", "\" Not A;Brand\";v=\"99\", \"Chromium\";v=\"90\", \"Google Chrome\";v=\"90\""},
			{"accept", "*/*"},
			{"x-requested-with", "XMLHttpRequest"},
			{"sec-ch-ua-mobile", "?0"},
			{"user-agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/90.0.4430.212 Safari/537.36"},
			{"sec-fetch-site", "same-origin"},
			{"sec-fetch-mode", "cors"},
			{"sec-fetch-dest", "empty"},
			{"referer", BaseEndpoint},
			{"accept-encoding", "gzip, deflate, br"},
			{"accept-language", "en-US,en;q=0.9"},
		},
	})
	if err != nil {
		return err
	}

	if resp.StatusCode != 200 {
		return errors.New("not 200: " + fmt.Sprint(resp.StatusCode))
	}

	type DatadomeCookie struct {
		Cookie string `json:"cookie"`
	}

	cookie := DatadomeCookie{}

	err = json.Unmarshal([]byte(body), &cookie)
	if err != nil {
		return err
	}

	if !strings.Contains(cookie.Cookie, "datadome=") ||
		!strings.Contains(cookie.Cookie, "; ") {
		return errors.New("bad cookie: " + cookie.Cookie)
	}
	datadomeCookieValue, err := util.FindInString(cookie.Cookie, "datadome=", "; ")
	if err != nil {
		return err
	}

	datadomeCookie := &http.Cookie{
		Name:     "datadome",
		Value:    datadomeCookieValue,
		Path:     "/",
		Domain:   ".pokemoncenter.com",
		MaxAge:   31536000,
		Secure:   true,
		SameSite: http.SameSiteLaxMode,
	}
	u, err := url.Parse("https://www.pokemomcenter.com/") // This should never error, but just to be safe let's handle the error
	if err != nil {
		log.Println("Error parsing https://www.pokemomcenter.com/ to set Datadome cookie: " + err.Error())
		return err
	}
	oldCookies := client.Jar.Cookies(u)
	newCookies := []*http.Cookie{}
	for _, oldCookie := range oldCookies {
		if oldCookie.Name != "datadome" {
			newCookies = append(newCookies, oldCookie)
		}
	}
	newCookies = append(newCookies, datadomeCookie)
	log.Println(newCookies)
	client.Jar.SetCookies(u, newCookies)
	return nil
}

func dumpMap(space string, m map[string]interface{}) {
	for k, v := range m {
		if mv, ok := v.(map[string]interface{}); ok {
			fmt.Printf("{ \"%v\": \n", k)
			dumpMap(space+"\t", mv)
			fmt.Printf("}\n")
		} else {
			fmt.Printf("%v %v : %v\n", space, k, v)
		}
	}
}

//We could of made this task apart of 'task' and pulled details from here but as this will be moved later to security its best to pass these details in.
func CyberSourceV2(keyId string, card entities.Card) (string, error) {
	returnVal := ""
	key := strings.Split(keyId, ".")[1]

	decodedKeyBytes, err := base64.StdEncoding.DecodeString(key)
	if err != nil {
		return "", err
	}
	decodedKeyString := string(decodedKeyBytes)

	var encrypt Encrypt
	err = json.Unmarshal([]byte(decodedKeyString), &encrypt)
	if err != nil {
		return "", err
	}

	kid := string(encrypt.Flx.Jwk.Kid)

	e := string(encrypt.Flx.Jwk.E)
	n := string(encrypt.Flx.Jwk.N)
	kty := string(encrypt.Flx.Jwk.Kty)
	use := string(encrypt.Flx.Jwk.Use)

	rsa_ := new(RSA)
	rsa_.Kid = kid
	rsa_.E = e
	rsa_.Kty = kty
	rsa_.N = n
	rsa_.Use = use

	header_ := new(Header__)
	header_.Kid = kid
	header_.Jwk = *rsa_

	card_ := &Card{SecurityCode: card.CVV, Number: card.CardNumber, ExpMonth: card.ExpMonth, ExpYear: card.ExpYear}

	switch card.CardType { // https://developer.cybersource.com/library/documentation/dev_guides/Retail_SO_API/html/Topics/app_card_types.htm
	case "Visa":
		card_.Type = "001"
	case "Mastercard":
		card_.Type = "002"
	case "AMEX":
		card_.Type = "003"
	case "Discover":
		card_.Type = "004"
	case "Diners":
		card_.Type = "005"
	case "Diners - Carte Blanche":
		card_.Type = "006"
	case "JCB":
		card_.Type = "007"
	case "Visa Electron":
		card_.Type = "033"
	}

	encryptedObject_ := new(EncryptedObject)
	encryptedObject_.Context = keyId
	encryptedObject_.Index = 0
	encryptedObject_.Data = *card_

	jwkJSON := `{
		"keys": [ 
		  {
			"kty": "` + kty + `",
			"n": "` + n + `",
			"use": "` + use + `",
			"alg": "RSA-OAEP",
			"e": "` + e + `",
			"kid": "` + kid + `"
		  }
		]
	  }
	  `

	set, err := jwk.Parse([]byte(jwkJSON))
	if err != nil {
		return "", err
	}

	for it := set.Iterate(context.Background()); it.Next(context.Background()); {
		pair := it.Pair()
		key := pair.Value.(jwk.Key)

		var rawkey interface{}
		if err := key.Raw(&rawkey); err != nil {
			return "", err
		}

		rsa___, ok := rawkey.(*rsa.PublicKey)

		if !ok {
			return "", fmt.Errorf("expected rsa key, got %T", rawkey)
		}

		payload := `{
			"context": "` + keyId + `",
			"index": 0,
			"data":{
				"securityCode":"` + card_.SecurityCode + `",
				"number":"` + card_.Number + `",
				"type":"` + card_.Type + `",
				"expirationMonth":"` + card_.ExpMonth + `",
				"expirationYear":"` + card_.ExpYear + `"
			}
		}`

		h_map := `{
			"kid":"` + kid + `",
			"jwk":{
				"kty":"` + kty + `",
				"e":"` + e + `",
				"use":"` + use + `",
				"kid":"` + kid + `",
				"n":"` + n + `"
			}
		}`

		headerMap := make(map[string]interface{})
		err := json.Unmarshal([]byte(h_map), &headerMap)
		if err != nil {
			return "", err
		}
		dumpMap("", headerMap)

		token__, err := jose.Encrypt(payload, jose.RSA_OAEP, jose.A256GCM, rsa___, jose.Headers(headerMap))
		if err != nil {
			return "", err
		}
		returnVal = token__
	}

	return returnVal, nil
}

func retrievePaymentToken(keyId string) (string, error) {
	key := strings.Split(keyId, ".")[1]
	decodedKeyBytes, err := base64.StdEncoding.DecodeString(key)
	if err != nil {
		return "", err
	}
	decodedKeyString_ := string(decodedKeyBytes) + "}"
	fmt.Println(decodedKeyString_)
	var encrypt PaymentToken
	if err := json.Unmarshal([]byte(decodedKeyString_), &encrypt); err != nil {
		return "", err
	}
	return encrypt.Jti, nil
}

// Creates a embed for the DiscordWebhook function
func (task *Task) CreatePokemonCenterEmbed(status enums.OrderStatus, imageURL string) []sec.DiscordEmbed {
	embeds := []sec.DiscordEmbed{
		{
			Fields: []sec.DiscordField{
				{
					Name:   "Site:",
					Value:  "Pokemon Center",
					Inline: true,
				},
				{
					Name:   "Price:",
					Value:  "$" + fmt.Sprintf("%f", task.StockData.Price),
					Inline: true,
				},
				{
					Name:   "Product SKU:",
					Value:  task.StockData.SKU,
					Inline: true,
				},
				{
					Name:  "Product Name:",
					Value: task.StockData.ItemName,
				},
				{
					Name:  "Proxy:",
					Value: "||" + " " + util.ProxyCleaner(task.Task.Proxy) + " " + "||",
				},
			},
			Footer: sec.DiscordFooter{
				Text:    "Juiced AIO",
				IconURL: "https://media.discordapp.net/attachments/849430464036077598/855979506204278804/Icon_1.png?width=128&height=128",
			},
			Timestamp: time.Now(),
		},
	}

	switch status {
	case enums.OrderStatusSuccess:
		embeds[0].Title = ":tangerine: Checkout! :tangerine:"
		embeds[0].Color = 16742912
		embeds[0].Thumbnail = sec.DiscordThumbnail{
			URL: imageURL,
		}
	case enums.OrderStatusDeclined:
		embeds[0].Title = ":lemon: Card Declined :lemon:"
		embeds[0].Color = 16766464
		embeds[0].Thumbnail = sec.DiscordThumbnail{
			URL: imageURL,
		}
	case enums.OrderStatusFailed:
		embeds[0].Title = ":apple: Failed to Place Order :apple:"
		embeds[0].Color = 14495044
		embeds[0].Thumbnail = sec.DiscordThumbnail{
			URL: imageURL,
		}
	}

	return embeds
}

// RunUntilSuccessful runs a single function until (a) it succeeds, (b) the task needs to stop, or (c) it fails maxRetries times.
// 		Passing in -1 for maxRetries will retry the function indefinitely.
//		Returns true if the function was successful, false if the function failed (and the task should stop)
func (task *Task) RunUntilSuccessful(fn func() (bool, string), maxRetries int) (bool, string) {
	attempt := 1
	if maxRetries == -1 {
		attempt = -2
	}

	var success bool
	var status string
	for !success {
		success, status = task.RunUntilSuccessfulHelper(fn, attempt)
		needToStop := task.CheckForStop()
		if needToStop {
			if task.Task.Task.TaskStatus != "Idle" {
				task.PublishEvent(enums.TaskIdle, enums.TaskStop, 0)
			}
			return false, ""
		}
		if attempt >= maxRetries {
			if task.Task.Task.TaskStatus != "Idle" {
				task.PublishEvent(fmt.Sprintf(enums.TaskFailed, status), enums.TaskFail, 0)
			}
			task.Task.StopFlag = true
			return false, ""
		}
		if attempt >= 0 {
			attempt++
		}

		time.Sleep(time.Duration(task.Task.Task.TaskDelay) * time.Millisecond)
	}

	return true, status
}

func (task *Task) RunUntilSuccessfulHelper(fn func() (bool, string), attempt int) (bool, string) {
	success, status := fn()

	if !success {
		if attempt > 0 {
			if status != "" && !task.Task.StopFlag && task.Task.Task.TaskStatus != "Idle" {
				task.PublishEvent(fmt.Sprint(fmt.Sprintf("(Attempt #%d) ", attempt), status), enums.TaskUpdate, -1)
			}
		} else {
			if status != "" && !task.Task.StopFlag && task.Task.Task.TaskStatus != "Idle" && task.Task.Task.TaskStatus != fmt.Sprint("(Retrying) ", status) {
				task.PublishEvent(fmt.Sprint("(Retrying) ", status), enums.TaskUpdate, -1)
			}
		}
		return false, status
	}

	if status != "" {
		task.PublishEvent(status, enums.TaskUpdate, -1)
	}
	return true, status
}
