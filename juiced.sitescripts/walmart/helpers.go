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
	sec "backend.juicedbot.io/juiced.security/auth/util"
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
	cookies := client.Jar.Cookies(u)
	cookies = append(cookies, cookie)
	client.Jar.SetCookies(u, cookies)

	return pxValues, nil
}

func SetPXCapCookie(captchaURL string, pxValues *util.PXValues, proxy entities.Proxy, client *http.Client) error {
	log.Println("Requesting Captcha token")
	token, err := captcha.RequestCaptchaToken(enums.ReCaptchaV2, enums.Walmart, captchaURL, "", 0, proxy)
	if err != nil {
		fmt.Println("Error getting ReCaptcha v2 Token: " + err.Error())
		return err
	}

	if token == nil {
		log.Println("No tokens available, waiting until one is available")
	}

	for token == nil {
		token = captcha.PollCaptchaTokens(enums.ReCaptchaV2, enums.Walmart, captchaURL, proxy)
		time.Sleep(1 * time.Second / 10)
	}
	log.Println("Received Captcha token")
	tokenInfo, ok := token.(entities.ReCaptchaToken)
	if !ok {
		log.Println(token)
		err = errors.New("token could not be parsed")
		log.Println("Error getting ReCaptcha v2 Token: " + err.Error())
		return err
	}
	px3, err := util.GetPXCapCookie("walmart", pxValues.SetID, pxValues.VID, pxValues.UUID, tokenInfo.Token, proxy)
	if err != nil {
		log.Println("Error getting PXCap cookie: " + err.Error())
		return err
	}

	log.Println("Retrieved PX3 cookie: " + px3)

	px3Cookie := &http.Cookie{
		Name:   "_px3",
		Value:  px3,
		Path:   "/",
		Domain: ".walmart.com",
	}
	u, err := url.Parse("https://www.walmart.com/") // This should never error, but just to be safe let's handle the error
	if err != nil {
		log.Println("Error parsing https://www.walmart.com/ to set PXCap cookie: " + err.Error())
		return err
	}
	oldCookies := client.Jar.Cookies(u)
	newCookies := []*http.Cookie{}
	for _, oldCookie := range oldCookies {
		if oldCookie.Name != "_px3" {
			newCookies = append(newCookies, oldCookie)
		}
	}
	newCookies = append(newCookies, px3Cookie)
	client.Jar.SetCookies(u, newCookies)

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

// Creates a embed for the DiscordWebhook function
func (task *Task) CreateWalmartEmbed(status enums.OrderStatus, imageURL string) []sec.DiscordEmbed {
	embeds := []sec.DiscordEmbed{
		{
			Fields: []sec.DiscordField{
				{
					Name:   "Site:",
					Value:  "Walmart",
					Inline: true,
				},
				{
					Name:   "Price:",
					Value:  "$" + fmt.Sprint(0), // TODO: @TeHNiC
					Inline: true,
				},
				{
					Name:   "Product SKU:",
					Value:  fmt.Sprintf("[%v](https://www.walmart.com/ip/%v)", task.Sku, task.Sku),
					Inline: true,
				},
				{
					Name:  "Product Name:",
					Value: "NaN", // TODO: @TeHNiC
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
func (task *Task) IsSuccessful(runTaskResult bool, event bool) (bool, bool) {
	needToStop := task.CheckForStop()
	if needToStop {
		return true, true
	}
	if !runTaskResult {
		time.Sleep(time.Duration(task.Task.Task.TaskDelay) * time.Millisecond)
		return false, false
	}

	return true, false
}
