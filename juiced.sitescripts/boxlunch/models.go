package boxlunch

import (
	"backend.juicedbot.io/juiced.infrastructure/common/entities"
	"backend.juicedbot.io/juiced.sitescripts/base"
)

// Endpoints
const (
	BaseEndpoint = "https://www.boxlunch.com/"

	AddToCartEndpoint = "https://www.boxlunch.com/on/demandware.store/Sites-boxlunch-Site/default/Cart-AddProduct?format=ajax"
	AddToCartReferer  = "https://www.boxlunch.com/product/"

	GetCheckoutEndpoint = "https://www.boxlunch.com/cart?"
	GetCheckoutReferer  = "https://www.boxlunch.com/cart"

	ProceedToCheckoutEndpoint = "https://www.boxlunch.com/cart?"
	ProceedToCheckoutReferer  = "https://www.boxlunch.com/cart?"

	GuestCheckoutEndpoint = "https://www.boxlunch.com/cart?"
	GuestCheckoutReferer  = "https://www.boxlunch.com/cart?"

	SubmitShippingEndpoint = "https://www.boxlunch.com/cart?"
	SubmitShippingReferer  = "https://www.boxlunch.com/cart?"

	UseOrigAddressEndpoint = "https://www.boxlunch.com/cart?"
	UseOrigAddressReferer  = "https://www.boxlunch.com/cart?"

	SubmitPaymentInfoEndpoint = "https://www.boxlunch.com/cart?"
	SubmitPaymentInfoReferer  = "https://www.boxlunch.com/cart?"

	SubmitOrderEndpoint = "https://www.boxlunch.com/orderconfirmation"
	SubmitOrderReferer  = "https://www.boxlunch.com/cart?"

	MonitorEndpoint = "https://www.boxlunch.com/on/demandware.store/Sites-boxlunch-Site/default/Product-Variation?pid=%s&Quantity=1&format=ajax"
)

// Monitor info
type Monitor struct {
	Monitor         base.Monitor
	RunningMonitors []string
	Pids            []string
	InStock         []BoxlunchInStockData
	PidWithInfo     map[string]entities.BoxlunchSingleMonitorInfo
}

// Task info
type Task struct {
	Task      base.Task
	StockData BoxlunchInStockData
	Inseam    string
	Dwcont    string
	OldDwcont string
	SecureKey string
}

type BoxlunchInStockData struct {
	PID         string
	Price       int
	SizePID     string
	Size        string
	Color       string
	ProductName string
	ImageURL    string
}
