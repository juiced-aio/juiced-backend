package amazon

import (
	"fmt"

	"backend.juicedbot.io/juiced.sitescripts/util"
)

// Creates the webhook depending on whether successful or not
func (task *Task) CreateAmazonFields(success bool) []util.Field {
	// When monitoring in Fast mode there is no way to find the name so this field will be empty and the
	// webhook would fail to send. This makes it NaN if this is the case.
	if task.TaskInfo.ItemName == "" {
		task.TaskInfo.ItemName = "*NaN*"
	}
	return []util.Field{
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
			Value:  fmt.Sprintf("[%v](https://www.amazon.com/dp/%v)", task.TaskInfo.ASIN, task.TaskInfo.ASIN),
			Inline: true,
		},
		{
			Name:  "Product Name:",
			Value: task.TaskInfo.ItemName,
		},
		{
			Name:  "Mode:",
			Value: string(task.TaskInfo.MonitorType),
		},
		{
			Name:  "Proxy:",
			Value: "||" + " " + util.ProxyCleaner(task.Task.Proxy) + " " + "||",
		},
		{
			Name:  "Offer Listing ID:",
			Value: task.TaskInfo.OfferID,
		},
	}
}
