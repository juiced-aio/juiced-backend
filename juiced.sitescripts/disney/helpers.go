package disney

import (
	"fmt"
	"log"
	"math/rand"
	"time"

	"backend.juicedbot.io/juiced.client/http"
	"backend.juicedbot.io/juiced.infrastructure/common/enums"
	sec "backend.juicedbot.io/juiced.security/auth/util"
	"backend.juicedbot.io/juiced.sitescripts/util"
)

func BecomeGuest(client *http.Client) bool {
	resp, _, err := util.MakeRequest(&util.Request{
		Client: *client,
		Method: "GET",
		URL:    BaseEndpoint,
		RawHeaders: http.RawHeader{
			{"sec-ch-ua", `" Not;A Brand";v="99", "Google Chrome";v="91", "Chromium";v="91"`},
			{"sec-ch-ua-mobile", `?0`},
			{"upgrade-insecure-requests", `1`},
			{"user-agent", `Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/91.0.4472.124 Safari/537.36`},
			{"accept", `text/html,application/xhtml+xml,application/xml;q=0.9,image/avif,image/webp,image/apng,*/*;q=0.8,application/signed-exchange;v=b3;q=0.9`},
			{"sec-fetch-site", `none`},
			{"sec-fetch-mode", `navigate`},
			{"sec-fetch-user", `?1`},
			{"sec-fetch-dest", `document`},
			{"accept-encoding", `gzip, deflate, br`},
			{"accept-language", `en-US,en;q=0.9`},
		},
	})
	if err != nil || resp.StatusCode != 200 {
		if err != nil {
			log.Println(err.Error())
		}
		return false
	}

	return true
}

func RandomString(selection string, length int) (r string) {
	for i := 0; i < length; i++ {
		r += string(selection[rand.Intn(len(selection))])
	}
	return
}

// Creates a embed for the DiscordWebhook function
func (task *Task) CreateDisneyEmbed(status enums.OrderStatus, imageURL string) []sec.DiscordEmbed {
	embeds := []sec.DiscordEmbed{
		{
			Fields: []sec.DiscordField{
				{
					Name:   "Site:",
					Value:  "Disney",
					Inline: true,
				},
				{
					Name:   "Price:",
					Value:  "$" + fmt.Sprint(task.TaskInfo.Total),
					Inline: true,
				},
				{
					Name:   "Product SKU:",
					Value:  fmt.Sprintf("[%v](%v)", task.StockData.PID, task.StockData.ItemURL),
					Inline: true,
				},
				{
					Name:  "Product Name:",
					Value: task.StockData.ProductName,
				},
				{
					Name:  "Task Type:",
					Value: string(task.TaskType),
				},
				{
					Name:  "Proxy:",
					Value: "||" + " " + util.ProxyCleaner(*task.Task.Proxy) + " " + "||",
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
