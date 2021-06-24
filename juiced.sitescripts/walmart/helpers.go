package walmart

import (
	"fmt"
	"strings"
	"time"

	"backend.juicedbot.io/juiced.client/http"
	"backend.juicedbot.io/juiced.infrastructure/common/enums"
	"backend.juicedbot.io/juiced.infrastructure/common/events"
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

// Creates a embed for the DiscordWebhook function
func (task *Task) CreateWalmartEmbed(status enums.OrderStatus) []sec.DiscordEmbed {
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
					Value:  "$" + fmt.Sprint(task.CheckoutInfo.ItemPrice),
					Inline: true,
				},
				{
					Name:   "Product SKU:",
					Value:  fmt.Sprintf("[%v](https://www.walmart.com/ip/%v)", task.Sku, task.Sku),
					Inline: true,
				},
				{
					Name:  "Product Name:",
					Value: task.CheckoutInfo.ItemName,
				},

				{
					Name:  "Proxy:",
					Value: "||" + " " + util.ProxyCleaner(task.Task.Proxy) + " " + "||",
				},
				{
					Name:  "Offer Listing ID:",
					Value: task.OfferID,
				},
			},
			Footer: sec.DiscordFooter{
				Text:    "Juiced AIO",
				IconURL: "https://cdn.discordapp.com/icons/688572290488991757/b684ee4e3cfb661d32afc48f24776e60.png?size=128",
			},
			Timestamp: time.Now(),
		},
	}

	switch status {
	case enums.OrderStatusSuccess:
		embeds[0].Title = ":tangerine: Checkout! :tangerine:"
		embeds[0].Color = 16742912
		embeds[0].Thumbnail = sec.DiscordThumbnail{
			URL: task.CheckoutInfo.ImageUrl,
		}
	case enums.OrderStatusDeclined:
		embeds[0].Title = ":lemon: Card Declined :lemon:"
		embeds[0].Color = 16766464
		embeds[0].Thumbnail = sec.DiscordThumbnail{
			URL: task.CheckoutInfo.ImageUrl,
		}
	case enums.OrderStatusFailed:
		embeds[0].Title = ":apple: Failed to Place Order :apple:"
		embeds[0].Color = 14495044
		embeds[0].Thumbnail = sec.DiscordThumbnail{
			URL: task.CheckoutInfo.ImageUrl,
		}
	}
	return embeds
}
