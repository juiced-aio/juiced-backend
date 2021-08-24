package base

import (
	"fmt"
	"strings"
	"time"

	"backend.juicedbot.io/juiced.infrastructure/common"
	"backend.juicedbot.io/juiced.infrastructure/common/enums"
	sec "backend.juicedbot.io/juiced.security/auth/util"
	"backend.juicedbot.io/juiced.sitescripts/util"
)

func (task *Task) PublishEvent(status enums.TaskStatus, eventType enums.TaskEventType) {
	taskInfo := task.GetTaskInfo()
	if taskInfo == nil {
		return
	}
	taskInfo.Task.SetTaskStatus(status)
	taskInfo.EventBus.PublishTaskEvent(status, eventType, nil, taskInfo.Task.ID)
}

func (task *Task) CheckForStop() bool {
	taskInfo := task.GetTaskInfo()
	if taskInfo == nil {
		return true
	}
	if taskInfo.StopFlag {
		task.PublishEvent(enums.TaskIdle, enums.TaskStop)
		return true
	}
	return false
}

// RunUntilSuccessful runs a single function until (a) it succeeds, (b) the task needs to stop, or (c) it fails too many times.
// 		Passing in 0 for maxRetries will retry the function indefinitely.
//		Returns true if the function was successful, false if the function failed (and the task should stop)
func (task *Task) RunUntilSuccessful(function util.TaskFunction) (bool, string) {
	taskInfo := task.GetTaskInfo()
	if taskInfo == nil {
		return false, ""
	}

	attempt := 1
	if function.MaxRetries == 0 {
		attempt = -1
	}

	var success bool
	var status string
	for success, status = task.RunUntilSuccessfulHelper(function.Function, attempt); !success; {
		needToStop := task.CheckForStop()
		if needToStop || attempt > function.MaxRetries {
			taskInfo.StopFlag = true
			return false, ""
		}
		if attempt >= 0 {
			attempt++
		}

		if function.MsBetweenRetries == 0 {
			time.Sleep(time.Duration(taskInfo.Task.TaskDelay) * time.Millisecond)
		} else {
			time.Sleep(time.Duration(function.MsBetweenRetries) * time.Millisecond)
		}
	}

	return true, status
}

func (task *Task) RunUntilSuccessfulHelper(fn func() (bool, string), attempt int) (bool, string) {
	success, status := fn()

	if !success {
		if attempt > 0 {
			if status != "" {
				task.PublishEvent(fmt.Sprint(fmt.Sprintf("(Attempt #%d) ", attempt), status), enums.TaskUpdate)
			}
		} else {
			if status != "" {
				task.PublishEvent(fmt.Sprint("(Retrying) ", status), enums.TaskUpdate)
			}
		}
		return false, status
	}

	if status != "" {
		task.PublishEvent(status, enums.TaskUpdate)
	}
	return true, status
}

func (task *Task) RefreshWrapper(function util.TaskFunction) {
	defer func() {
		if r := recover(); r != nil {
			task.RefreshWrapper(function)
		}
	}()

	for {
		if function.RefreshAt == 0 || time.Now().Unix() > function.RefreshAt {
			if success, _ := task.RunUntilSuccessful(function); !success {
				return
			}
			function.RefreshAt = time.Now().Unix() + 1800
		}
		time.Sleep(time.Millisecond * common.MS_TO_WAIT)
	}

}

func CreateDiscordEmbed(retailer string, status string, taskInfo *util.TaskInfo) []sec.DiscordEmbed {
	embeds := []sec.DiscordEmbed{
		{
			Fields: []sec.DiscordField{
				{
					Name:   "Retailer:",
					Value:  retailer,
					Inline: true,
				},
				{
					Name:   "Price:",
					Value:  "$" + fmt.Sprintf("%f", taskInfo.StockInfo.Price),
					Inline: true,
				},
				{
					Name:   "Product SKU:",
					Value:  taskInfo.StockInfo.SKU,
					Inline: true,
				},
				{
					Name:  "Product Name:",
					Value: taskInfo.StockInfo.ItemName,
				},
				{
					Name:  "Proxy:",
					Value: "||" + " " + util.ProxyCleaner(taskInfo.Proxy) + " " + "||",
				},
			},
			Footer: sec.DiscordFooter{
				Text:    "Juiced AIO",
				IconURL: "https://media.discordapp.net/attachments/849430464036077598/855979506204278804/Icon_1.png?width=128&height=128",
			},
			Timestamp: time.Now(),
		},
	}

	if strings.Contains(status, enums.OrderStatusSuccess) {
		embeds[0].Title = ":tangerine: Checkout! :tangerine:"
		embeds[0].Color = 16742912
		embeds[0].Thumbnail = sec.DiscordThumbnail{
			URL: taskInfo.StockInfo.ImageURL,
		}
	}
	if strings.Contains(status, enums.OrderStatusDeclined) {
		embeds[0].Title = ":lemon: Card Declined :lemon:"
		embeds[0].Color = 16766464
		embeds[0].Thumbnail = sec.DiscordThumbnail{
			URL: taskInfo.StockInfo.ImageURL,
		}
	}
	if strings.Contains(status, enums.OrderStatusFailed) {
		embeds[0].Title = ":apple: Failed to Place Order :apple:"
		embeds[0].Color = 14495044
		embeds[0].Thumbnail = sec.DiscordThumbnail{
			URL: taskInfo.StockInfo.ImageURL,
		}
	}

	return embeds
}
