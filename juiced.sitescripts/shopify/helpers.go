package shopify

import (
	"fmt"
	"strings"
	"time"

	"backend.juicedbot.io/juiced.client/http"
	"backend.juicedbot.io/juiced.infrastructure/common"
	"backend.juicedbot.io/juiced.infrastructure/common/enums"
	sec "backend.juicedbot.io/juiced.security/auth/util"
	"backend.juicedbot.io/juiced.sitescripts/util"
)

func SplitCard(card string, cardType string) string {
	var cardSplit string
	switch cardType {
	case "AMEX":
		cardSplit = strings.Join([]string{card[:4], card[4:10], card[10:15]}, " ")
	case "Diners":
		cardSplit = strings.Join([]string{card[:4], card[4:10], card[10:14]}, " ")
	default:
		cardSplit = strings.Join([]string{card[:4], card[4:8], card[8:12], card[12:16]}, " ")
	}
	return cardSplit
}

// Creates a embed for the DiscordWebhook function
func (task *Task) CreateShopifyEmbed(status enums.OrderStatus, imageURL string) []sec.DiscordEmbed {
	fields := []sec.DiscordField{
		{
			Name:   "Retailer:",
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
			Value:  fmt.Sprintf("[%v](%v%v)", task.SiteURL, task.VariantID, task.TaskInfo.ItemURL),
			Inline: true,
		},
		{
			Name:  "Product Name:",
			Value: task.TaskInfo.Name,
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

func BecomeGuest(client http.Client, siteURL, sitePassword string) bool {
	if sitePassword == "" {
		resp, _, err := util.MakeRequest(&util.Request{
			Client: client,
			Method: "GET",
			URL:    siteURL,
			RawHeaders: http.RawHeader{
				{"sec-ch-ua", `" Not;A Brand";v="99", "Google Chrome";v="91", "Chromium";v="91"`},
				{"accept", "application/json, text/javascript, */*; q=0.01"},
				{"x-requested-with", "XMLHttpRequest"},
				{"sec-ch-ua-mobile", "?0"},
				{"user-agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/91.0.4472.106 Safari/537.36"},
				{"content-type", "application/x-www-form-urlencoded; charset=UTF-8"},
				{"sec-fetch-site", "same-origin"},
				{"sec-fetch-site", "cors"},
				{"sec-fetch-dest", "empty"},
				{"accept-encoding", "gzip, deflate"},
				{"accept-language", "en-US,en;q=0.9"},
			},
		})
		if err != nil || resp.StatusCode != 200 {
			return false
		}
	} else {
		paramsString := common.CreateParams(map[string]string{
			"form_type": "storefront_password",
			"utf8":      "✓",
			"password":  sitePassword,
		})

		resp, _, err := util.MakeRequest(&util.Request{
			Client: client,
			Method: "POST",
			URL:    siteURL,
			RawHeaders: http.RawHeader{
				{"content-length", fmt.Sprint(len(paramsString))},
				{"sec-ch-ua", `" Not;A Brand";v="99", "Google Chrome";v="91", "Chromium";v="91"`},
				{"accept", "application/json, text/javascript, */*; q=0.01"},
				{"x-requested-with", "XMLHttpRequest"},
				{"sec-ch-ua-mobile", "?0"},
				{"user-agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/91.0.4472.106 Safari/537.36"},
				{"content-type", "application/x-www-form-urlencoded; charset=UTF-8"},
				{"sec-fetch-site", "same-origin"},
				{"sec-fetch-site", "cors"},
				{"sec-fetch-dest", "empty"},
				{"accept-encoding", "gzip, deflate"},
				{"accept-language", "en-US,en;q=0.9"},
			},
			Data: []byte(paramsString),
		})
		if err != nil || resp.StatusCode != 200 {
			return false
		}
	}

	return true
}
