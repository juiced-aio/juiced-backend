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
	"backend.juicedbot.io/juiced.infrastructure/common/enums"
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
func CyberSourceV2(keyId string, card Card) (returnVal string) {
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
	card_.SecurityCode = card.SecurityCode
	card_.Number = card.Number
	card_.Type = "001" //visa and 002 = mastercard
	card_.ExpMonth = card.ExpMonth
	card_.ExpYear = card.ExpYear

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
				"securityCode":"260",
				"number":"4767718212263745",
				"type":"001",
				"expirationMonth":"02",
				"expirationYear":"2026"
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

//Improves readability on RunTask
func (task *Task) RunUntilSuccessful(runTaskResult bool, status string) (bool, bool) {
	needToStop := task.CheckForStop()
	x := 5 //should come from front-end somewhere, unless we want to hard code a 'retry' amount.
	//If we want individual tasks to have different retry amouunts we can assign each task and pass in as a paramater.
	// -1 retry = unlimited amount of retries.
	if needToStop || task.Retry > x {
		task.Task.StopFlag = true //if retry is over the limit we want to set our stop flag.
		return true, true
	}
	if !runTaskResult {
		if status != "" {
			task.PublishEvent(fmt.Sprint(status, " Retry: ", task.Retry), enums.TaskUpdate) //if failure then also send back retry number
		}
		task.Retry++
		time.Sleep(time.Duration(task.Task.Task.TaskDelay) * time.Millisecond)
		return false, false
	} else {
		if status != "" {
			task.PublishEvent(status, enums.TaskUpdate) //If success then just publish the status
		}
	}

	return true, false
}
