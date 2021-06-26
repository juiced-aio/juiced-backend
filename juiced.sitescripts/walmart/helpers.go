package walmart

import (
	"errors"
	"fmt"
	"log"
	"net/url"
	"strings"
	"time"

	"backend.juicedbot.io/juiced.client/http"
	"backend.juicedbot.io/juiced.infrastructure/common/captcha"
	"backend.juicedbot.io/juiced.infrastructure/common/entities"
	"backend.juicedbot.io/juiced.infrastructure/common/enums"
	"backend.juicedbot.io/juiced.sitescripts/util"

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

func SetPXCookie(proxy entities.Proxy, client *http.Client) (util.PXValues, error) {
	px3, pxValues, err := util.GetPXCookie("walmart", proxy)
	if err != nil {
		fmt.Println("Error getting PX cookie: " + err.Error())
		return pxValues, err
	}
	cookie := &http.Cookie{
		Name:   "_px3",
		Value:  px3,
		Path:   "/",
		Domain: ".walmart.com",
	}
	u, err := url.Parse("https://walmart.com/") // This should never error, but just to be safe let's handle the error
	if err != nil {
		fmt.Println("Error parsing https://walmart.com/ to set PX cookie: " + err.Error())
		return pxValues, err
	}
	client.Jar.SetCookies(u, []*http.Cookie{cookie})
	return pxValues, nil
}

func SetPXCapCookie(captchaURL string, pxValues *util.PXValues, proxy entities.Proxy, client *http.Client) error {
	log.Println("Requesting Captcha token")
	token, err := captcha.RequestCaptchaToken(enums.ReCaptchaV2, enums.Walmart, captchaURL, proxy)
	if err != nil {
		return err
	}

	if token == nil {
		log.Println("No tokens available, waiting until one is available")
	}

	for token == nil {
		token, err = captcha.PollCaptchaTokens(enums.ReCaptchaV2, enums.Walmart, captchaURL, proxy)
		if err != nil {
			fmt.Println("Error getting ReCaptcha v2 Token: " + err.Error())
			return err
		}
		time.Sleep(1 * time.Second / 10)
	}
	log.Println("Received Captcha token")
	tokenString, ok := token.(string)
	if !ok {
		err = errors.New("token is not a string")
		fmt.Println("Error getting ReCaptcha v2 Token: " + err.Error())
		return err
	}
	px3, err := util.GetPXCapCookie("walmart", pxValues.SetID, pxValues.VID, pxValues.UUID, tokenString, proxy)
	if err != nil {
		fmt.Println("Error getting PXCap cookie: " + err.Error())
		return err
	}

	cookie := &http.Cookie{
		Name:   "_px3",
		Value:  px3,
		Path:   "/",
		Domain: ".walmart.com",
	}
	u, err := url.Parse("https://walmart.com/") // This should never error, but just to be safe let's handle the error
	if err != nil {
		fmt.Println("Error parsing https://walmart.com/ to set PXCap cookie: " + err.Error())
		return err
	}
	client.Jar.SetCookies(u, []*http.Cookie{cookie})
	return nil
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
