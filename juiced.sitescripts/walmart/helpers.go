package walmart

import (
	"fmt"
	"net/url"
	"strings"
	"time"

	"backend.juicedbot.io/juiced.client/http"
	"backend.juicedbot.io/juiced.infrastructure/common/entities"
	"backend.juicedbot.io/juiced.infrastructure/common/events"
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

func SetPXCapCookie(captchaURL string, pxValues *util.PXValues, proxy entities.Proxy, client *http.Client) error {
	token := "" // In the future, this will be replaced with something like util.RequestCaptchaToken(captchaURL). For now, just leave it as a blank string (it won't work, but we just want the infrastructure so we can update it once we have Captcha ready).
	px3, err := util.GetPXCapCookie("walmart", pxValues.SetID, pxValues.VID, pxValues.UUID, token, proxy)
	if err != nil {
		fmt.Println("Error getting PXCap cookie: " + err.Error())
		return err
	}
	var cookies []*http.Cookie

	cookie := &http.Cookie{
		Name:   "_px3",
		Value:  px3,
		Path:   "/",
		Domain: ".walmart.com",
	}
	cookies = append(cookies, cookie)
	u, err := url.Parse("https://walmart.com/") // This should never error, but just to be safe let's handle the error
	if err != nil {
		fmt.Println("Error parsing https://walmart.com/ to set PXCap cookie: " + err.Error())
		return err
	}
	client.Jar.SetCookies(u, cookies)
	return nil
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

// RefreshPX3 refreshes the px3 cookie every 4 minutes since it expires every 5 minutes
func (task *Task) RefreshPX3() {
	// If the function panics due to a runtime error, recover and restart it
	defer func() {
		recover()
		task.RefreshPX3()
	}()

	for {
		if task.PXValues.RefreshAt == 0 || time.Now().Unix() > task.PXValues.RefreshAt {
			_, pxValues, err := util.GetPXCookie("walmart", task.Task.Proxy)

			if err != nil {
				return // Eventually we'll want to handle this. But if we run into errors and keep requesting cookies, we might send a TON of requests to our API, and I don't want them to get mad at us for sending too many.
			}
			task.PXValues = pxValues
		}
	}
}

func (monitor *Monitor) RefreshPX3() {
	// If the function panics due to a runtime error, recover and restart it
	defer func() {
		recover()
		monitor.RefreshPX3()
	}()

	for {
		if monitor.PXValues.RefreshAt == 0 || time.Now().Unix() > monitor.PXValues.RefreshAt {
			_, pxValues, err := util.GetPXCookie("walmart", monitor.Monitor.Proxy)

			if err != nil {
				return // Eventually we'll want to handle this. But if we run into errors and keep requesting cookies, we might send a TON of requests to our API, and I don't want them to get mad at us for sending too many.
			}
			monitor.PXValues = pxValues
		}
	}
}
