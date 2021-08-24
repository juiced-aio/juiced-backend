package pokemoncenter

import (
	"context"
	"crypto/rsa"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"strings"

	"backend.juicedbot.io/juiced.client/http"
	"backend.juicedbot.io/juiced.infrastructure/common/entities"
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
