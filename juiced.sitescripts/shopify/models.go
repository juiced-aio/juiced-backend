package shopify

import (
	"time"

	"backend.juicedbot.io/juiced.client/http"
	"backend.juicedbot.io/juiced.infrastructure/common/entities"
	"backend.juicedbot.io/juiced.infrastructure/common/enums"

	"backend.juicedbot.io/juiced.sitescripts/base"
)

const (
	SearchEndpoint        = "/search/suggest.json?q=%v&resources[type]=product"
	AddToCartEndpoint     = "/cart/add.js"
	CartEndpoint          = "/cart"
	CheckoutEndpoint      = "/checkout"
	ShippingRatesEndpoint = "/cart/shipping_rates.json?shipping_address%%5Bzip%%5D=%v&shipping_address%%5Bcountry%%5D=%v&shipping_address%%5Bprovince%%5D=%v"
	CreditIDEndpoint      = "https://deposit.us.shopifycs.com/sessions"
)

type Step = int

const (
	SettingUp           Step = 0
	WaitingForMonitor   Step = 1
	AddingToCart        Step = 2
	SettingShippingInfo Step = 3
)

type Task struct {
	Task            base.Task
	Step            Step
	ShopifyRetailer enums.ShopifyRetailer
	SiteURL         string
	VariantID       string
	InStockData     ShopifyInStockData
	AccountInfo     AccountInfo
	TaskInfo        TaskInfo
	Client          http.Client
}

type ShopifyInStockData struct {
	VariantID string
	Price     float64
	ItemName  string
	ImageURL  string
}

// Monitor info
type Monitor struct {
	Monitor         base.Monitor
	SKUsSentToTask  []string
	RunningMonitors []string
	OutOfStockSKUs  []string
	VIDs            []string
	InStock         []ShopifyInStockData
	SiteURL         string
	SKUWithInfo     map[string]entities.ShopifySingleMonitorInfo
}

type AccountInfo struct {
	Email    string
	Password string
}

type TaskInfo struct {
	Image          string
	Name           string
	ItemURL        string
	Price          int
	AuthToken      string
	CheckoutURL    string
	ShippingRate   string
	PaymentGateway string
	CreditID       string
	OrderTotal     string
}

type Request struct {
	Client             http.Client
	Method             string
	URL                string
	Headers            http.Header
	RawHeaders         [][2]string
	Referer            string
	Data               []byte
	RequestBodyStruct  interface{}
	ResponseBodyStruct interface{}
	RandOpt            string
}

type AddToCartResponse struct {
	ID                 int64  `json:"id"`
	Quantity           int    `json:"quantity"`
	VariantID          int64  `json:"variant_id"`
	Key                string `json:"key"`
	Title              string `json:"title"`
	Price              int    `json:"price"`
	Sku                string `json:"sku"`
	Vendor             string `json:"vendor"`
	ProductID          int64  `json:"product_id"`
	URL                string `json:"url"`
	Image              string `json:"image"`
	Handle             string `json:"handle"`
	ProductTitle       string `json:"product_title"`
	ProductDescription string `json:"product_description"`
}

type PollRequest struct {
	Query     string    `json:"query"`
	Variables Variables `json:"variables"`
}
type Variables struct {
	Token string `json:"token"`
}

type PollResponse struct {
	Data Data `json:"data"`
}
type Productvariantavailability struct {
	ID        int64 `json:"id"`
	Available bool  `json:"available"`
}
type Poll struct {
	Token                      string                       `json:"token"`
	Pollafter                  time.Time                    `json:"pollAfter"`
	Queueetaseconds            int                          `json:"queueEtaSeconds"`
	Productvariantavailability []Productvariantavailability `json:"productVariantAvailability"`
	Typename                   string                       `json:"__typename"`
}
type Data struct {
	Poll Poll `json:"poll"`
}

type ShippingRatesResponse struct {
	ShippingRates []ShippingRates `json:"shipping_rates"`
}
type ShippingRates struct {
	Code   string `json:"code"`
	Price  string `json:"price"`
	Source string `json:"source"`
}

type CreditIDRequest struct {
	CreditCard          CreditCard `json:"credit_card"`
	PaymentSessionScope string     `json:"payment_session_scope"`
}
type CreditCard struct {
	Number            string `json:"number"`
	Name              string `json:"name"`
	Month             int    `json:"month"`
	Year              int    `json:"year"`
	VerificationValue string `json:"verification_value"`
}

type CreditIDResponse struct {
	ID string `json:"id"`
}
