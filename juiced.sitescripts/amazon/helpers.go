package amazon

import (
	"fmt"
	"time"

	"backend.juicedbot.io/juiced.infrastructure/common/enums"
	sec "backend.juicedbot.io/juiced.security/auth/util"
	"backend.juicedbot.io/juiced.sitescripts/util"
)

// Creates a embed for the DiscordWebhook function
func (task *Task) CreateAmazonEmbed(status enums.OrderStatus, imageURL string) []sec.DiscordEmbed {
	// When monitoring in Fast mode there is no way to find the name so this field will be empty and the
	// webhook would fail to send. This makes it NaN if this is the case.
	if task.StockData.ItemName == "" {
		task.StockData.ItemName = "*NaN*"
	}
	embeds := []sec.DiscordEmbed{
		{
			Fields: []sec.DiscordField{
				{
					Name:   "Site:",
					Value:  "Amazon",
					Inline: true,
				},
				{
					Name:   "Price:",
					Value:  "$" + fmt.Sprint(task.CheckoutInfo.Price),
					Inline: true,
				},
				{
					Name:   "Product SKU:",
					Value:  fmt.Sprintf("[%v](https://www.amazon.com/dp/%v)", task.StockData.ASIN, task.StockData.ASIN),
					Inline: true,
				},
				{
					Name:  "Product Name:",
					Value: task.StockData.ItemName,
				},
				{
					Name:  "Mode:",
					Value: string(task.StockData.MonitorType),
				},
				{
					Name:  "Proxy:",
					Value: "||" + " " + util.ProxyCleaner(task.Task.Proxy) + " " + "||",
				},
				{
					Name:  "Offer Listing ID:",
					Value: task.StockData.OfferID,
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
