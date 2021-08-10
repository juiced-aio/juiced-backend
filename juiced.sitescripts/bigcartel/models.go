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
	Skus            []string
	SKUWithInfo     map[string]entities.BigCartelSingleMonitorInfo
	SiteURL         string
}

type Task struct {
	Task              base.Task
	InStockData       BigCartelInStockData
	BigCartelRetailer enums.BigCartelRetailer
	Step              Step
	Retry             int
	MaxRetry          int
	SiteURL           string
}

type BigCartelInStockData struct {
	Sku       string
	ItemName  string
	ImageURL  string
	ItemPrice float64
	StoreId   string
	CartToken string
	Key       string
	PaymentId string
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

type Item struct {
	Product_id        string `json:"product_id"`
	Product_option_id string `json:"product_option_id"`
	Product_name      string `json:"product_name"`
	Option_name       string `json:"option_name"`
	Full_name         string `json:"full_name"`
	Quantity          string `json:"quantity"`
	Price             string `json:"price"`
	Total             string `json:"total"`
	Total_cents       string `json:"total_cents"`
	Primary_image     struct {
		Url        string `json:"url"`
		Secure_url string `json:"secure_url"`
	} `json:"primary_image"`
}

type BigCartelRequestSubmitPaymentMethodResponse struct {
	Id              string
	Object          string
	Billing_details struct {
		Address struct {
			City        string
			Country     string
			Line1       string
			Line2       string
			Postal_code string
			State       string
		}
		Email string
		Name  string
		Phone string
	}
	Card struct {
		Brand  string
		Checks struct {
			Address_line1_check       string
			Address_postal_code_check string
			Cvc_check                 string
		}
		Country        string
		Exp_month      string
		Exp_year       string
		Funding        string
		Generated_from string
		Last4          string
		Networks       struct {
			Available []string
			Preferred string
		}
		Three_d_secure_usage struct {
			Supported bool
		}
		Wallet string
	}
	Created  string
	Customer string
	Livemode string
	Type     string
}

type PaymentSubmitRequest struct {
	Location string
}
type PaymentSubmitRequestFollowUp struct {
	Token  string
	Status string
	Errors string
}
