package pokemoncenter

import (
	"context"
	"crypto/rsa"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"log"
	"strings"
	"time"

	"backend.juicedbot.io/juiced.client/http"
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
func CyberSourceV2(keyId string, card entities.Card) (returnVal string) {
	key := strings.Split(keyId, ".")[1]

	decodedKeyBytes, _ := base64.StdEncoding.DecodeString(key)
	decodedKeyString := string(decodedKeyBytes)

	var encrypt Encrypt
	json.Unmarshal([]byte(decodedKeyString), &encrypt)

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

	card_ := new(Card)
	card_ = &Card{SecurityCode: card.CVV, Number: card.CardNumber, ExpMonth: card.ExpMonth, ExpYear: card.ExpYear}

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
		panic(err)
	}

	for it := set.Iterate(context.Background()); it.Next(context.Background()); {
		pair := it.Pair()
		key := pair.Value.(jwk.Key)

		var rawkey interface{}
		if err := key.Raw(&rawkey); err != nil {
			log.Printf("failed to create public key: %s", err)
			return
		}

		rsa___, ok := rawkey.(*rsa.PublicKey)

		if !ok {
			panic(fmt.Sprintf("expected ras key, got %T", rawkey))
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
			panic(err)
		}
		dumpMap("", headerMap)

		token__, err__ := jose.Encrypt(payload, jose.RSA_OAEP, jose.A256GCM, rsa___, jose.Headers(headerMap))
		if err__ != nil {
			fmt.Println(err__)
		}
		returnVal = token__

	}
	return returnVal
}

func retrievePaymentToken(keyId string) (jti string) {
	key := strings.Split(keyId, ".")[1]
	decodedKeyBytes, _ := base64.StdEncoding.DecodeString(key)
	decodedKeyString_ := string(decodedKeyBytes) + "}"
	fmt.Println(decodedKeyString_)
	var encrypt PaymentToken
	if err := json.Unmarshal([]byte(decodedKeyString_), &encrypt); err != nil {
		fmt.Println(err)
	}
	return encrypt.Jti
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

//Improves readability on RunTask
func (task *Task) RunUntilSuccessful(runTaskResult bool, status string) (bool, bool) {
	needToStop := task.CheckForStop()
	// -1 retry = unlimited amount of retries.
	if needToStop || task.Retry > task.MaxRetry {
		task.Task.StopFlag = true //if retry is over the limit we want to set our stop flag.
		return true, true
	}
	if !runTaskResult { //We have failed the task
		if status != "" { //Check if we need to publish event
			task.PublishEvent(fmt.Sprint(status, " Retry: ", task.Retry), enums.TaskUpdate) //if failure then also send back retry number
		}
		if task.Retry >= 0 {
			task.Retry++ //increment our retry.
		}
		time.Sleep(time.Duration(task.Task.Task.TaskDelay) * time.Millisecond)
		return false, false
	} else { //We have completed the task succesfully
		if status != "" { //Check if we need to publish event
			task.PublishEvent(status, enums.TaskUpdate) //If success then just publish the status
		}
		//If we want we can reset retry amount here, depending on if we want the retries global or per task.
		task.Retry = 0
	}

	return true, false //Finally if we have reached this point then task was a success.
}
