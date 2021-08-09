package bigcartel

import (
	"fmt"
	"time"

	"backend.juicedbot.io/juiced.infrastructure/common/enums"
	sec "backend.juicedbot.io/juiced.security/auth/util"
	"backend.juicedbot.io/juiced.sitescripts/util"
)

// Creates a embed for the DiscordWebhook function
func (task *Task) CreateBigCartelEmbed(status enums.OrderStatus, imageURL string) []sec.DiscordEmbed {
	embeds := []sec.DiscordEmbed{
		{
			Fields: []sec.DiscordField{
				{
					Name:   "Site:",
					Value:  task.BigCartelRetailer,
					Inline: true,
				},
				{
					Name:   "Price:",
					Value:  "$" + fmt.Sprint(task.InStockData.ItemPrice),
					Inline: true,
				},
				{
					Name:   "Product SKU:",
					Value:  fmt.Sprintf("[%v](%v%v)", task.SiteURL, task.InStockData.Sku, task.InStockData.ImageURL),
					Inline: true,
				},
				{
					Name:  "Product Name:",
					Value: task.InStockData.ItemName,
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

//Improves readability on RunTask
func (task *Task) RunUntilSuccessful(runTaskResult bool, status string) (bool, bool) {
	needToStop := task.CheckForStop()
	// -1 retry = unlimited amount of retries.
	if needToStop || task.Retry > task.MaxRetry {
		task.Task.StopFlag = true //if retry is over the limit we want to set our stop flag.
		return true, true
	}
	if !runTaskResult { //We have failed the task
		if status != "" { //Check if we need to publish event
			task.PublishEvent(fmt.Sprint(status, " Retry: ", task.Retry), enums.TaskUpdate) //if failure then also send back retry number
		}
		if task.Retry >= 0 {
			task.Retry++ //increment our retry.
		}
		time.Sleep(time.Duration(task.Task.Task.TaskDelay) * time.Millisecond)
		return false, false
	} else { //We have completed the task succesfully
		if status != "" { //Check if we need to publish event
			task.PublishEvent(status, enums.TaskUpdate) //If success then just publish the status
		}
		//If we want we can reset retry amount here, depending on if we want the retries global or per task.
		task.Retry = 0
	}

	return true, false //Finally if we have reached this point then task was a success.
}
