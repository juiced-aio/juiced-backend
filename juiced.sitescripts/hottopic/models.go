package hottopic

import (
	"backend.juicedbot.io/juiced.infrastructure/entities"
)

// Endpoints
const (
	BaseEndpoint    = "https://www.hottopic.com/"
	ProductEndpoint = "https://www.hottopic.com/product/%s"

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

	MonitorEndpoint  = "https://www.hottopic.com/on/demandware.store/Sites-hottopic-Site/default/Product-Variation?pid=%s&Quantity=1&format=ajax"
	MonitorEndpoint2 = "https://www.hottopic.com/on/demandware.store/Sites-hottopic-Site/default/Product-Variation?pid=%s&Quantity=1&format=ajax&dwvar_%s_color=%v&dwvar_%s_size=%v"
)

type SKUMonitor struct {
	Input         entities.MonitorInput
	HotTopicInput MonitorInput
	BaseMonitor   *entities.BaseMonitor

	ItemName string
	ImageURL string
	Price    float64
}

type MonitorInput struct {
	Size  string
	Color string
}

type HotTopicProductInfo struct {
	SizePID string
	Size    string
	Color   string
}

type Task struct {
	Input         entities.TaskInput
	HotTopicInput TaskInput
	BaseTask      *entities.BaseTask

	PrevDwcont string
	Dwcont     string
	SecureKey  string
}

type TaskInput struct{}

type SizeInfo struct {
	SizePID string
	Size    string
}
