package newegg

import (
	"encoding/base64"
	"fmt"
	"math/rand"
	"strings"
	"time"

	"backend.juicedbot.io/juiced.client/http"
	"backend.juicedbot.io/juiced.infrastructure/common"
	"backend.juicedbot.io/juiced.infrastructure/common/enums"
	sec "backend.juicedbot.io/juiced.security/auth/util"
	"backend.juicedbot.io/juiced.sitescripts/util"
)

func BecomeGuest(client http.Client) bool {
	resp, _, err := util.MakeRequest(&util.Request{
		Client: client,
		Method: "GET",
		URL:    BaseEndpoint,
		RawHeaders: http.RawHeader{
			{"sec-ch-ua", `"Chromium";v="92", " Not A;Brand";v="99", "Google Chrome";v="92"`},
			{"sec-ch-ua-mobile", `?0`},
			{"upgrade-insecure-requests", `1`},
			{"user-agent", `Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/92.0.4515.131 Safari/537.36`},
			{"accept", `text/html,application/xhtml+xml,application/xml;q=0.9,image/avif,image/webp,image/apng,*/*;q=0.8,application/signed-exchange;v=b3;q=0.9`},
			{"sec-fetch-site", `none`},
			{"sec-fetch-mode", `navigate`},
			{"sec-fetch-user", `?1`},
			{"sec-fetch-dest", `document`},
			{"accept-encoding", `gzip, deflate, br`},
			{"accept-language", `en-US,en;q=0.9`},
		},
	})
	if resp.StatusCode != 200 || err != nil {
		return false
	}

	return true
}

func CreateExtras() (string, string) {
	nonceBytes := make([]byte, 16)
	_, err := rand.Read(nonceBytes)
	if err != nil {
		return "", ""
	}

	newSign := make([]byte, 16)
	_, err = rand.Read(newSign)
	if err != nil {
		return "", ""
	}

	params := common.CreateParams(map[string]string{
		"timestamp": fmt.Sprint(time.Now().Unix()),
		"nonce":     strings.ToLower(fmt.Sprintf("%X", nonceBytes)),
		"appId":     "107630",
	})

	return params, base64.StdEncoding.EncodeToString([]byte(strings.ToLower(fmt.Sprintf("%X", newSign))))

}

// Creates a embed for the DiscordWebhook function
func (task *Task) CreateNeweggEmbed(status enums.OrderStatus, imageURL string) []sec.DiscordEmbed {
	fields := []sec.DiscordField{
		{
			Name:   "Retailer:",
			Value:  "Newegg",
			Inline: true,
		},
		{
			Name:   "Price:",
			Value:  "$" + fmt.Sprint(task.StockData.Price),
			Inline: true,
		},
		{
			Name:   "Product SKU:",
			Value:  fmt.Sprintf("[%v](%v)", task.StockData.SKU, task.StockData.ItemURL),
			Inline: true,
		},
		{
			Name:  "Product Name:",
			Value: task.StockData.ProductName,
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
