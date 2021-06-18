package walmart

import (
	"net/url"
	"strings"

	"backend.juicedbot.io/juiced.client/http"
	"backend.juicedbot.io/juiced.infrastructure/common/entities"
	"backend.juicedbot.io/juiced.infrastructure/common/events"
	"backend.juicedbot.io/juiced.infrastructure/queries"
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

//Returns a PX cookie using Auth/Security/PXCap functions
func GetPxCookie(_url string, proxy entities.Proxy) (*url.URL, []*http.Cookie) {
	var cookies []*http.Cookie

	//get user info
	_, userinfo, _ := queries.GetUserInfo()

	//make PX request
	resp, _, _ := sec.PX(_url, util.ProxyCleaner(proxy), userinfo)
	sec.PXCap(_url, util.ProxyCleaner(proxy), resp.SetID, resp.VID, resp.UUID, resp.PX3, userinfo)
	cookie := &http.Cookie{
		Name:   "_px3",
		Value:  "",
		Path:   "/",
		Domain: ".walmart.com",
	}
	cookies = append(cookies, cookie)
	u, _ := url.Parse(_url)
	return u, cookies
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
