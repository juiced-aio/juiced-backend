package hottopic

import (
	"backend.juicedbot.io/juiced.infrastructure/common/entities"
	"backend.juicedbot.io/juiced.sitescripts/base"
)

// Endpoints
const (
	BaseEndpoint = "https://www.hottopic.com/"

	AddToCartEndpoint = "https://www.hottopic.com/on/demandware.store/Sites-hottopic-Site/default/Cart-AddProduct?format=ajax"
	AddToCartReferer  = "https://www.hottopic.com/product/"

	GetCheckoutEndpoint = "https://www.hottopic.com/cart?"
	GetCheckoutReferer  = "https://www.hottopic.com/cart"

	ProceedToCheckoutEndpoint = "https://www.hottopic.com/cart?dwcont="
	ProceedToCheckoutReferer  = "https://www.hottopic.com/cart?"

	GuestCheckoutEndpoint = "https://www.hottopic.com/cart?dwcont="
	GuestCheckoutReferer  = "https://www.hottopic.com/cart?dwcont="

	SubmitShippingEndpoint = "https://www.hottopic.com/cart?dwcont="
	SubmitShippingReferer  = "https://www.hottopic.com/cart?dwcont="

	UseOrigAddressEndpoint = "https://www.hottopic.com/cart?dwcont="
	UseOrigAddressReferer  = "https://www.hottopic.com/cart?dwcont="

	SubmitPaymentInfoEndpoint = "https://www.hottopic.com/cart?dwcont="
	SubmitPaymentInfoReferer  = "https://www.hottopic.com/cart?dwcont="

	SubmitOrderEndpoint = "https://www.hottopic.com/orderconfirmation"
	SubmitOrderReferer  = "https://www.hottopic.com/cart?dwcont="

	MonitorEndpoint = "https://www.hottopic.com/on/demandware.store/Sites-hottopic-Site/default/Product-Variation?pid=%s&Quantity=1&format=ajax"
)

// Monitor info
type Monitor struct {
	Monitor         base.Monitor
	RunningMonitors []string
	Pids            []string
	InStock         []HotTopicInStockData
	PidWithInfo     map[string]entities.HottopicSingleMonitorInfo
}

type HotTopicInStockData struct {
	PID         string
	SizePID     string
	Color       string
	ProductName string
	ImageURL    string
	Price       int
}

// Task info
type Task struct {
	Task      base.Task
	PID       string
	SizePID   string
	Color     string
	Inseam    string
	Dwcont    string
	OldDwcont string
	SecureKey string
}
