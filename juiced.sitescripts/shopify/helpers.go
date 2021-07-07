package shopify

import (
	"fmt"
	"time"

	"backend.juicedbot.io/juiced.infrastructure/common/enums"
	sec "backend.juicedbot.io/juiced.security/auth/util"
	"backend.juicedbot.io/juiced.sitescripts/util"
)

func SplitCard(card string) []string {
	return []string{card[:4], card[4:8], card[8:12], card[12:16]}
}

// Creates a embed for the DiscordWebhook function
func (task *Task) CreateShopifyEmbed(status enums.OrderStatus, imageURL string) []sec.DiscordEmbed {
	embeds := []sec.DiscordEmbed{
		{
			Fields: []sec.DiscordField{
				{
					Name:   "Site:",
					Value:  task.ShopifyRetailer,
					Inline: true,
				},
				{
					Name:   "Price:",
					Value:  "$" + fmt.Sprint(task.TaskInfo.Price),
					Inline: true,
				},
				{
					Name:   "Product SKU:",
					Value:  fmt.Sprintf("[%v](https://hotwheelscollectors.com%v)", task.SKU, task.TaskInfo.ItemURL),
					Inline: true,
				},
				{
					Name:  "Product Name:",
					Value: task.TaskInfo.Name,
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
