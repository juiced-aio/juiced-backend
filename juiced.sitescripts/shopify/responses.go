package shopify

import "time"

// Monitor
type MonitorResponse struct {
	Products []Products `json:"products"`
}
type Products struct {
	Variants []Variants `json:"variants"`
}
type Variants struct {
	ID int64 `json:"id"`
}

// AddToCart
type AddToCartResponse struct {
	ID                 int64  `json:"id"`
	Quantity           int    `json:"quantity"`
	VariantID          int64  `json:"variant_id"`
	Key                string `json:"key"`
	Title              string `json:"title"`
	Price              int    `json:"price"`
	SKU                string `json:"sku"`
	Vendor             string `json:"vendor"`
	ProductID          int64  `json:"product_id"`
	URL                string `json:"url"`
	Image              string `json:"image"`
	Handle             string `json:"handle"`
	ProductTitle       string `json:"product_title"`
	ProductDescription string `json:"product_description"`
}

// Queue
type PollResponse struct {
	Data Data `json:"data"`
}
type ProductVariantAvailability struct {
	ID        int64 `json:"id"`
	Available bool  `json:"available"`
}
type Poll struct {
	Token                      string                       `json:"token"`
	PollAfter                  time.Time                    `json:"pollAfter"`
	QueueEtaSeconds            int                          `json:"queueEtaSeconds"`
	ProductVariantAvailability []ProductVariantAvailability `json:"productVariantAvailability"`
	Typename                   string                       `json:"__typename"`
}
type Data struct {
	Poll Poll `json:"poll"`
}

// GetShippingRates
type GetShippingRatesResponse struct {
	ShippingRates []ShippingRates `json:"shipping_rates"`
}
type ShippingRates struct {
	Code   string `json:"code"`
	Price  string `json:"price"`
	Source string `json:"source"`
}

// GetCreditID
type GetCreditIDResponse struct {
	ID string `json:"id"`
}
