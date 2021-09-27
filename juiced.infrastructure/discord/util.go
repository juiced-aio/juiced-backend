package discord

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"backend.juicedbot.io/juiced.client/http"
	"backend.juicedbot.io/juiced.infrastructure/entities"
	"backend.juicedbot.io/juiced.infrastructure/enums"
	"backend.juicedbot.io/juiced.infrastructure/staticstores"
	"backend.juicedbot.io/juiced.sitescripts/util"
)

var hookChan = make(chan WebhookInfo)

func QueueWebhook(success bool, content string, embeds []DiscordEmbed) {
	hookChan <- WebhookInfo{
		Success: success,
		Content: content,
		Embeds:  embeds,
	}
}

func DiscordWebhookQueue() {
	for {
		hook := <-hookChan
		settings := staticstores.GetSettings()
		var webhookURL string
		if hook.Success {
			webhookURL = settings.SuccessDiscordWebhook
		} else {
			webhookURL = settings.FailureDiscordWebhook
		}
		if webhookURL != "" {
			SendDiscordWebhook(webhookURL, hook.Embeds)
		}
		time.Sleep(2*time.Second + (time.Second / 2))
	}
}

func SendDiscordWebhook(discordWebhook string, embeds []DiscordEmbed) error {
	client := &http.Client{
		Transport: &http.Transport{},
	}
	response, _, err := util.MakeRequest(&util.Request{
		Client: client,
		Method: "POST",
		URL:    discordWebhook,
		AddHeadersFunction: func(request *http.Request, e ...string) {
			request.Header.Set("content-type", "application/json")
		},
		RequestBodyStruct: DiscordWebhook{
			Content: nil,
			Embeds:  embeds,
		},
	})
	if err != nil {
		return err
	}

	if response.StatusCode >= 200 && response.StatusCode < 300 {
		return nil
	}
	return errors.New(response.Status)
}

func CreateDiscordEmbed(retailer, proxy, profile, status string, productInfo entities.ProductInfo) []DiscordEmbed {
	fields := []DiscordField{
		{
			Name:   "Retailer:",
			Value:  retailer,
			Inline: true,
		},
		{
			Name:   "Price:",
			Value:  "$" + fmt.Sprintf("%.2f", productInfo.Price),
			Inline: true,
		},
		{
			Name:   "Product SKU:",
			Value:  productInfo.SKU,
			Inline: true,
		},
		{
			Name:  "Product Name:",
			Value: productInfo.ItemName,
		},
		{
			Name:  "Profile:",
			Value: "||" + " " + profile + " " + "||",
		},
	}
	if proxy != "" {
		fields = append(fields, DiscordField{
			Name:  "Proxy:",
			Value: "||" + " " + proxy + " " + "||",
		})
	}
	embeds := []DiscordEmbed{
		{
			Fields: fields,
			Footer: DiscordFooter{
				Text:    "Juiced",
				IconURL: "https://media.discordapp.net/attachments/849430464036077598/855979506204278804/Icon_1.png?width=128&height=128",
			},
			Timestamp: time.Now(),
		},
	}

	if strings.Contains(status, enums.CheckedOut) {
		embeds[0].Title = ":tangerine: Checkout! :tangerine:"
		embeds[0].Color = 16742912
		embeds[0].Thumbnail = DiscordThumbnail{
			URL: productInfo.ImageURL,
		}
	} else if strings.Contains(status, enums.CardDeclined) {
		embeds[0].Title = ":lemon: Card Declined :lemon:"
		embeds[0].Color = 16766464
		embeds[0].Thumbnail = DiscordThumbnail{
			URL: productInfo.ImageURL,
		}
	} else {
		embeds[0].Title = ":apple: Failed to Place Order :apple:"
		embeds[0].Color = 14495044
		embeds[0].Thumbnail = DiscordThumbnail{
			URL: productInfo.ImageURL,
		}
	}

	return embeds
}
