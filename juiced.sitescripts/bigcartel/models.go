package bigcartel

import (
	"backend.juicedbot.io/juiced.infrastructure/common/entities"
	"backend.juicedbot.io/juiced.infrastructure/common/enums"
	"backend.juicedbot.io/juiced.sitescripts/base"
)

const (
	AddToCartEndpoint     = "/cart"
	AddressEmailEndpoint  = "/store/%s/carts/%s"
	PaymentMethodEndpoint = "https://api.bigcartel.com/store/%s/checkouts"
	PaymentInfoEndpoint   = "/store/%s/carts/%s"
	CheckoutEndpoint      = "/store/%s/carts/%s"
	GetStockEndpoint      = "https://redsky.target.com/redsky_aggregations/v1/web/plp_fulfillment_v1?"
)

type Step = int

const (
	SettingUp           Step = 0
	Preloading          Step = 1
	WaitingForMonitor   Step = 2
	AddingToCart        Step = 3
	SettingShippingInfo Step = 4
)

type Monitor struct {
	Monitor         base.Monitor
	RunningMonitors []string
	Pids            []string
	InStock         []BigCartelInStockData
	PidWithInfo     map[string]entities.BoxlunchSingleMonitorInfo
	SiteInfo        SiteInfo
	Skus            []string
	SKUWithInfo     map[string]entities.BigCartelSingleMonitorInfo
}

type Task struct {
	Task              base.Task
	InStockData       BigCartelInStockData
	SiteInfo          SiteInfo
	BigCartelRetailer enums.BigCartelRetailer
	Step              Step
	Retry             int
	MaxRetry          int
}

type BigCartelInStockData struct {
	Sku       string
	ItemName  string
	ImageURL  string
	ItemPrice float64
	StoreId   string
	CartToken string
}

type SiteInfo struct {
	BaseUrl  string
	Email    string
	Password string
	Sku      string
}

type AddToCartResponse struct {
}

type BigCartelRequestSubmitNameAndEmail struct {
	Buyer_email                 string `json:"buyer_email"`
	Buyer_first_name            string `json:"buyer_first_name"`
	Buyer_last_name             string `json:"buyer_last_name"`
	Buyer_opted_in_to_marketing bool   `json:"buyer_opted_in_to_marketing"`
	Buyer_phone_number          string `json:"buyer_phone_number"`
}

type BigCartelRequestSubmitAddress struct {
	Shipping_address_1             string `json:"shipping_address_1"`
	Shipping_address_2             string `json:"shipping_address_2"`
	Shipping_city                  string `json:"shipping_city"`
	Shipping_country_autofill_name string `json:"shipping_country_autofill_name"`
	Shipping_country_id            string `json:"shipping_country_id"`
	Shipping_state                 string `json:"shipping_state"`
	Shipping_zip                   string `json:"shipping_zip"`
}

type Payment struct {
	Stripe_payment_method_id string `json:"stripe_payment_method_id"`
}
