package shopify

import (
	"backend.juicedbot.io/juiced.infrastructure/entities"
	"backend.juicedbot.io/juiced.infrastructure/enums"
)

// Endpoints
const (
	ClearCartEndpoint     = "/cart/clear"
	ProductsEndpoint      = "/products.json"
	SearchEndpoint        = "/search/suggest.json?q=%v&resources[type]=product"
	AddToCartEndpoint     = "/cart/add.js"
	CartEndpoint          = "/cart"
	CheckoutEndpoint      = "/checkout"
	ShippingRatesEndpoint = "/cart/shipping_rates.json?shipping_address%%5Bzip%%5D=%v&shipping_address%%5Bcountry%%5D=%v&shipping_address%%5Bprovince%%5D=%v"
	CreditIDEndpoint      = "https://deposit.us.shopifycs.com/sessions"
	CreditIDReferer       = "https://checkout.shopifycs.com"
)

// Errors
const (
	InvalidQueueTokenError        = "missing or invalid queue token"
	NoShippingRatesAvailableError = "no valid shipping rates for this address"
)

type SKUMonitor struct {
	Input        entities.MonitorInput
	ShopifyInput MonitorInput
	BaseMonitor  *entities.BaseMonitor
}

type URLMonitor struct {
	Input        entities.MonitorInput
	ShopifyInput MonitorInput
	BaseMonitor  *entities.BaseMonitor
}

type MonitorInput struct {
	ShopifyRetailer enums.ShopifyRetailer
	SiteURL         string
	SitePassword    string
}

type ShopifyProductInfo struct {
	VID string
}

type Task struct {
	Input        entities.TaskInput
	ShopifyInput TaskInput
	BaseTask     *entities.BaseTask

	Preloaded      bool
	HasCheckpoint  bool
	HasQueue       bool
	AuthToken      string
	CheckoutURL    string
	PaymentGateway string
	CreditID       string
	OrderTotal     string
}

type TaskInput struct {
	ShopifyRetailer enums.ShopifyRetailer
	SiteURL         string
	SitePassword    string
	CouponCode      string
	Email           string
	Password        string
	TaskType        enums.TaskType
}
