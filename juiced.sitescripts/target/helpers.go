package target

import (
	"fmt"
	"time"

	"backend.juicedbot.io/juiced.client/http"
	"backend.juicedbot.io/juiced.infrastructure/common/enums"
	sec "backend.juicedbot.io/juiced.security/auth/util"
	"backend.juicedbot.io/juiced.sitescripts/util"
)

// AddTargetHeaders adds Target-specific headers to the request
func AddTargetHeaders(request *http.Request, referer ...string) {
	util.AddBaseHeaders(request)
	request.Header.Set("Accept", "application/json")
	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("Origin", "https://www.target.com")
	if len(referer) != 0 {
		request.Header.Set("Referer", referer[0])
	}
}

// GetTCINStockRequestToMap maps a GetTCINStockRequest object to a map[string]string for converting to a query string
func GetTCINStockRequestToMap(getTCINStockRequest GetTCINStockRequest) map[string]string {
	return map[string]string{
		"key":                         getTCINStockRequest.Key,
		"tcins":                       getTCINStockRequest.TCINs,
		"store_id":                    getTCINStockRequest.StoreID,
		"scheduled_delivery_store_id": getTCINStockRequest.ScheduledDeliveryStoreID,
	}
}

// Creates a embed for the DiscordWebhook function
func (task *Task) CreateTargetEmbed(status enums.OrderStatus, imageURL string) []sec.DiscordEmbed {
	embeds := []sec.DiscordEmbed{
		{
			Fields: []sec.DiscordField{
				{
					Name:   "Site:",
					Value:  "Target",
					Inline: true,
				},
				{
					Name:   "Price:",
					Value:  "$" + fmt.Sprint(task.AccountInfo.CartInfo.CartItems[0].UnitPrice),
					Inline: true,
				},
				{
					Name:   "Product SKU:",
					Value:  fmt.Sprintf("[%v](https://www.target.com/p/-/A-%v)", task.TCIN, task.TCIN),
					Inline: true,
				},
				{
					Name:  "Product Name:",
					Value: task.AccountInfo.CartInfo.CartItems[0].ItemAttributes.Description,
				},
				{
					Name:  "Proxy:",
					Value: "||" + " " + util.ProxyCleaner(task.Task.Proxy) + " " + "||",
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
