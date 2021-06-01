package walmart

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"strings"
	"time"

	"backend.juicedbot.io/m/v2/juiced.client/http"
	"backend.juicedbot.io/m/v2/juiced.infrastructure/common/events"
	"backend.juicedbot.io/m/v2/juiced.sitescripts/util"

	"github.com/anaskhan96/soup"
)

// AddWalmartHeaders adds Walmart-specific headers to the request
func AddWalmartHeaders(request *http.Request, referer ...string) {
	util.AddBaseHeaders(request)
	request.Header.Set("Accept", "application/json")
	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("Origin", "https://www.walmart.com")
	// omitcsrfjwt: true
	// omitcorrelationid: true
	// credentials: include
	// TODO: Header order
	if len(referer) != 0 {
		request.Header.Set("Referer", referer[0])
	}
}

// GetPXCookie retrieves a PX cookie from the PX API main endpoint
// TODO: Handle errors
func GetPXCookie(client *http.Client, proxy string) GetPXCookieResponse {
	data := GetPXCookieRequest{
		Proxy: proxy,
	}

	getPXCookieResponse := GetPXCookieResponse{}

	payload, _ := json.Marshal(data)
	request, _ := http.NewRequest("POST", "http://104.131.175.97:3000/px", bytes.NewReader(payload))
	SetPXAPIHeaders(request)

	resp, err := client.Do(request)
	if err != nil {
		log.Println(err.Error())
		return getPXCookieResponse
	}
	defer resp.Body.Close()
	body, _ := ioutil.ReadAll(resp.Body)
	s := strings.Split(string(body), "8===D(.)")
	if len(s) != 5 {
		log.Println(s)
	}
	getPXCookieResponse = GetPXCookieResponse{
		SetId: s[2],
		Vid:   s[3],
		Uuid:  s[4],
	}
	httpCookie := &http.Cookie{
		Name:   "_px3",
		Value:  s[0],
		Domain: ".walmart.com",
		Path:   "/",
	}
	client.Jar.SetCookies(BaseURL, append(client.Jar.Cookies(BaseURL), httpCookie))
	return getPXCookieResponse
}

// GetPXCapCookie retrieves a PX cookie from the PX API's captcha endpoint
// TODO: Handle errors
func GetPXCapCookie(client *http.Client, proxy, captchaToken string, pxInfo GetPXCookieResponse) string {
	data := GetPXCapCookieRequest{
		SetId: pxInfo.SetId,
		Vid:   pxInfo.Vid,
		Uuid:  pxInfo.Uuid,
		Proxy: proxy,
		Token: captchaToken,
	}

	payload, _ := json.Marshal(data)
	request, _ := http.NewRequest("POST", "http://104.131.175.97:3000/pxcap", bytes.NewReader(payload))
	SetPXAPIHeaders(request)

	resp, err := client.Do(request)
	if err != nil {
		log.Println(err.Error())
		return ""
	}
	defer resp.Body.Close()
	body, _ := ioutil.ReadAll(resp.Body)
	fmt.Println(string(body))
	return string(body)
}

// SetPXAPIHeaders is a helper for the PX API requests
func SetPXAPIHeaders(request *http.Request) {
	key := "hfju783habdneywu7163j2uc809dhe7e"
	iv := "aj87dhe64k9dhwu6"
	id := "juiced"
	site := "walmart"
	watch := fmt.Sprint(time.Now().Unix())
	auth := "51A16ABA26FEBD7D369D6DA7D53B4" + watch

	request.Header.Add("Content-Type", "application/json")
	request.Header.Add("x-anti-antibot-id", Aes256(id, key, iv))
	request.Header.Add("x-anti-antibot-site", Aes256(site, key, iv))
	request.Header.Add("x-anti-antibot-watch", Aes256(watch, key, iv))
	request.Header.Add("x-anti-antibot-auth", Aes256(auth, key, iv))
}

// Aes256 is a helper for the PX API encryption
// TODO: Handle errors
func Aes256(plaintext string, key string, iv string) string {
	bKey := []byte(key)
	bIV := []byte(iv)
	bPlaintext := PKCS5Padding([]byte(plaintext), aes.BlockSize, len(plaintext))
	block, _ := aes.NewCipher(bKey)
	ciphertext := make([]byte, len(bPlaintext))
	mode := cipher.NewCBCEncrypter(block, bIV)
	mode.CryptBlocks(ciphertext, bPlaintext)
	return base64.StdEncoding.EncodeToString(ciphertext)
}

// PKCS5Padding is a helper for the PX API encryption
// TODO: Handle errors
func PKCS5Padding(ciphertext []byte, blockSize int, after int) []byte {
	padding := (blockSize - len(ciphertext)%blockSize)
	padtext := bytes.Repeat([]byte{byte(padding)}, padding)
	return append(ciphertext, padtext...)
}

//Converts a list of in-stock skus to a WarlmartSingleStockData structure.
func ConvertSkuListToWalmartSingleStock(skuCodes []string) []events.WalmartSingleStockData {
	inStock := events.WalmartSingleStockData{}
	inStockForShip := make([]events.WalmartSingleStockData, 0)

	for i := 0; i < len(skuCodes); i++ {
		inStock.Sku = skuCodes[i]
		inStockForShip = append(inStockForShip, inStock)
	}

	return inStockForShip
}

//Parses the response from the monitor and retrieves the Sku codes, then returns as an list of strings
func ParseInstockSku(resp soup.Root) []string {
	inStockForShip := make([]string, 0)

	formattedSkuList := strings.Replace(resp.Find("a", "class", "btn-compact").Attrs()["href"], "buynow?items=", "", -1)
	listOfSkuSplit := strings.Split(formattedSkuList, ",")
	for _, element := range listOfSkuSplit {
		inStockForShip = append(inStockForShip, strings.Split(element, "|")[0])
	}

	return inStockForShip
}

//Checks if the URL in the monitors response is present, to indicate if any items are in-stock.
func UrlExistsInResponse(resp soup.Root) bool {
	val := resp.Find("a", "class", "btn-compact")
	if val.Error == nil {
		return true
	} else {
		return false
	}
}
