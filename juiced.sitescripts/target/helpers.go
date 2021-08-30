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
	fields := []sec.DiscordField{
		{
			Name:   "Retailer:",
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
			Name:  "Profile:",
			Value: "||" + " " + task.Task.Profile.Name + " " + "||",
		},
	}

	if task.Task.Proxy != nil {
		fields = append(fields, sec.DiscordField{
			Name:  "Proxy:",
			Value: "||" + " " + util.ProxyCleaner(task.Task.Proxy) + " " + "||",
		})
	}

	embeds := []sec.DiscordEmbed{
		{
			Fields: fields,
			Footer: sec.DiscordFooter{
				Text:    "Juiced",
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
