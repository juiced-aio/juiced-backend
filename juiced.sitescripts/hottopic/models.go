package hottopic

import (
	"backend.juicedbot.io/juiced.infrastructure/common/entities"
	"backend.juicedbot.io/juiced.infrastructure/common/events"
	"backend.juicedbot.io/juiced.sitescripts/base"
)

// Endpoints
const (
	BaseEndpoint = "https://www.hottopic.com/"

	AddToCartEndpoint = "https://www.hottopic.com/on/demandware.store/Sites-hottopic-Site/default/Cart-AddProduct?format=ajax"
	AddToCartReferer  = "https://www.hottopic.com/product/"

	GetCheckoutEndpoint = "https://www.hottopic.com/cart?"
	GetCheckoutReferer  = "https://www.hottopic.com/cart"

	ProceedToCheckoutEndpoint = "https://www.hottopic.com/cart?"
	ProceedToCheckoutReferer  = "https://www.hottopic.com/cart?"

	GuestCheckoutEndpoint = "https://www.hottopic.com/cart?"
	GuestCheckoutReferer  = "https://www.hottopic.com/cart?"

	SubmitShippingEndpoint = "https://www.hottopic.com/cart?"
	SubmitShippingReferer  = "https://www.hottopic.com/cart?"

	UseOrigAddressEndpoint = "https://www.hottopic.com/cart?"
	UseOrigAddressReferer  = "https://www.hottopic.com/cart?"

	SubmitPaymentInfoEndpoint = "https://www.hottopic.com/cart?"
	SubmitPaymentInfoReferer  = "https://www.hottopic.com/cart?"

	SubmitOrderEndpoint = "https://www.hottopic.com/orderconfirmation"
	SubmitOrderReferer  = "https://www.hottopic.com/cart?"

	MonitorEndpoint = "https://www.hottopic.com/on/demandware.store/Sites-hottopic-Site/default/Product-Variation?pid="
)

// Monitor info
type Monitor struct {
	Monitor         base.Monitor
	RunningMonitors []string
	Pids            []PidSingle
	EventInfo       events.HotTopicSingleStockData
	PidWithInfo     map[string]entities.HottopicSingleMonitorInfo
}

type PidSingle struct {
	Pid   string
	size  string
	color string
}

// Task info
type Task struct {
	Task      base.Task
	Pid       string
	Size      string
	Color     string
	Inseam    string
	Dwcont    string
	OldDwcont string
	SecureKey string
}
