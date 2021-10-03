package boxlunch

import (
	"backend.juicedbot.io/juiced.infrastructure/entities"
)

// Endpoints
const (
	BaseEndpoint    = "https://www.boxlunch.com/"
	ProductEndpoint = "https://www.boxlunch.com/product/%s"

	AddToCartEndpoint = "https://www.boxlunch.com/on/demandware.store/Sites-boxlunch-Site/default/Cart-AddProduct?format=ajax"
	AddToCartReferer  = "https://www.boxlunch.com/product/"

	GetCheckoutEndpoint = "https://www.boxlunch.com/cart?"
	GetCheckoutReferer  = "https://www.boxlunch.com/cart"

	ProceedToCheckoutEndpoint = "https://www.boxlunch.com/cart?dwcont="
	ProceedToCheckoutReferer  = "https://www.boxlunch.com/cart?"

	GuestCheckoutEndpoint = "https://www.boxlunch.com/cart?dwcont="
	GuestCheckoutReferer  = "https://www.boxlunch.com/cart?dwcont="

	SubmitShippingEndpoint = "https://www.boxlunch.com/cart?dwcont="
	SubmitShippingReferer  = "https://www.boxlunch.com/cart?dwcont="

	UseOrigAddressEndpoint = "https://www.boxlunch.com/cart?dwcont="
	UseOrigAddressReferer  = "https://www.boxlunch.com/cart?dwcont="

	SubmitPaymentInfoEndpoint = "https://www.boxlunch.com/cart?dwcont="
	SubmitPaymentInfoReferer  = "https://www.boxlunch.com/cart?dwcont="

	SubmitOrderEndpoint = "https://www.boxlunch.com/orderconfirmation"
	SubmitOrderReferer  = "https://www.boxlunch.com/cart?dwcont="

	MonitorEndpoint  = "https://www.boxlunch.com/on/demandware.store/Sites-boxlunch-Site/default/Product-Variation?pid=%s&Quantity=1&format=ajax"
	MonitorEndpoint2 = "https://www.boxlunch.com/on/demandware.store/Sites-boxlunch-Site/default/Product-Variation?pid=%s&Quantity=1&format=ajax&dwvar_%s_color=%v&dwvar_%s_size=%v"
)

type SKUMonitor struct {
	Input         entities.MonitorInput
	BoxLunchInput MonitorInput
	BaseMonitor   *entities.BaseMonitor

	ItemName string
	ImageURL string
	Price    float64
}

type MonitorInput struct {
	Size  string
	Color string
}

type BoxLunchProductInfo struct {
	SizePID string
	Size    string
	Color   string
}

type Task struct {
	Input         entities.TaskInput
	BoxLunchInput TaskInput
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
