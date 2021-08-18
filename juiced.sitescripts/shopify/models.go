package shopify

import (
	"time"

	"backend.juicedbot.io/juiced.client/http"
	"backend.juicedbot.io/juiced.infrastructure/common/entities"
	"backend.juicedbot.io/juiced.infrastructure/common/enums"
	"backend.juicedbot.io/juiced.sitescripts/base"
	cmap "github.com/orcaman/concurrent-map"
)

const (
	ClearCartEndpoint     = "/cart/clear"
	ProductsEndpoint      = "/products.json"
	SearchEndpoint        = "/search/suggest.json?q=%v&resources[type]=product"
	MonitorEndpoint       = "/products.json"
	AddToCartEndpoint     = "/cart/add.js"
	CartEndpoint          = "/cart"
	CheckoutEndpoint      = "/checkout"
	ShippingRatesEndpoint = "/cart/shipping_rates.json?shipping_address%%5Bzip%%5D=%v&shipping_address%%5Bcountry%%5D=%v&shipping_address%%5Bprovince%%5D=%v"
	CreditIDEndpoint      = "https://deposit.us.shopifycs.com/sessions"
)

type Step = int

const (
	SettingUp           Step = 0
	Preloading          Step = 1
	WaitingForMonitor   Step = 2
	AddingToCart        Step = 3
	SettingShippingInfo Step = 4
)

type Task struct {
	Task            base.Task
	Step            Step
	ShopifyRetailer enums.ShopifyRetailer
	SiteURL         string
	SitePassword    string
	CouponCode      string
	StockData       SingleStockData
	AccountInfo     AccountInfo
	TaskInfo        TaskInfo
	Client          http.Client
}

type ShopifyStockData struct {
	InStock    []SingleStockData
	OutOfStock []SingleStockData
}

type SingleStockData struct {
	VariantID string
	Price     float64
	ItemName  string
	ImageURL  string
}

// Monitor info
type Monitor struct {
	Monitor         base.Monitor
	Keywords        []string
	InStock         cmap.ConcurrentMap
	SiteURL         string
	SitePassword    string
	KeywordWithInfo map[string]entities.ShopifySingleMonitorInfo
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

type MonitorResponse struct {
	Products []Products `json:"products"`
}
type Variants struct {
	ID               int64       `json:"id"`
	Title            string      `json:"title"`
	Option1          string      `json:"option1"`
	Option2          interface{} `json:"option2"`
	Option3          interface{} `json:"option3"`
	Sku              string      `json:"sku"`
	RequiresShipping bool        `json:"requires_shipping"`
	Taxable          bool        `json:"taxable"`
	FeaturedImage    interface{} `json:"featured_image"`
	Available        bool        `json:"available"`
	Price            string      `json:"price"`
	Grams            int         `json:"grams"`
	CompareAtPrice   interface{} `json:"compare_at_price"`
	Position         int         `json:"position"`
	ProductID        int64       `json:"product_id"`
	CreatedAt        string      `json:"created_at"`
	UpdatedAt        string      `json:"updated_at"`
}
type Images struct {
	ID         int64         `json:"id"`
	CreatedAt  string        `json:"created_at"`
	Position   int           `json:"position"`
	UpdatedAt  string        `json:"updated_at"`
	ProductID  int64         `json:"product_id"`
	VariantIds []interface{} `json:"variant_ids"`
	Src        string        `json:"src"`
	Width      int           `json:"width"`
	Height     int           `json:"height"`
}
type Options struct {
	Name     string   `json:"name"`
	Position int      `json:"position"`
	Values   []string `json:"values"`
}
type Products struct {
	ID          int64      `json:"id"`
	Title       string     `json:"title"`
	Handle      string     `json:"handle"`
	BodyHTML    string     `json:"body_html"`
	PublishedAt string     `json:"published_at"`
	CreatedAt   string     `json:"created_at"`
	UpdatedAt   string     `json:"updated_at"`
	Vendor      string     `json:"vendor"`
	ProductType string     `json:"product_type"`
	Tags        []string   `json:"tags"`
	Variants    []Variants `json:"variants"`
	Images      []Images   `json:"images"`
	Options     []Options  `json:"options"`
}

type ProductsResponse struct {
	Products []Products `json:"products"`
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
