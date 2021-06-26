package walmart

import (
	"backend.juicedbot.io/juiced.infrastructure/common/enums"
	"backend.juicedbot.io/juiced.sitescripts/base"
	"backend.juicedbot.io/juiced.sitescripts/util"
)

// Endpoints
const (
	BaseEndpoint = "https://www.walmart.com/"

	AddToCartEndpoint = "https://www.walmart.com/api/v3/cart/guest/:CID/items"
	AddToCartReferer  = "https://www.walmart.com/"

	GetCartInfoEndpoint = "https://www.walmart.com/api/checkout/v3/contract?page=CHECKOUT_VIEW"
	GetCartInfoReferer  = "https://www.walmart.com/checkout/"

	SetPcidEndpoint = "https://www.walmart.com/api/checkout/v3/contract/:PCID"
	SetPcidReferer  = "https://www.walmart.com/checkout/"

	SetShippingInfoEndpoint = "https://www.walmart.com/api/checkout/v3/contract/:PCID/shipping-address"
	SetShippingInfoReferer  = "https://www.walmart.com/checkout/"

	SetPaymentInfoEndpoint = "https://www.walmart.com/api/checkout/v3/contract/:PCID/payment"
	SetPaymentInfoReferer  = "https://www.walmart.com/checkout/"

	PlaceOrderEndpoint = "https://www.walmart.com/api/checkout/v3/contract/:PCID/order"
	PlaceOrderReferer  = "https://www.walmart.com/checkout/"

	MonitorEndpoint      = "https://affil.walmart.com/cart/buynow?items="
	PriceMonitorEndpoint = "https://www.walmart.com/ip/"
)

// Monitor info
type Monitor struct {
	Monitor        base.Monitor
	MonitorType    enums.MonitorType
	SKUs           []string
	InStockForShip []WalmartInStockData
	PXValues       util.PXValues
}

type WalmartInStockData struct {
	Sku     string `json:"sku"`
	OfferID string `json:"offerID"`
}

// Task info
type Task struct {
	Task     base.Task
	OfferID  string
	Sku      string
	CardInfo CardInfo
	PXValues util.PXValues
}

//Part of the Task struct
type CardInfo struct {
	EncryptedCvv   string `json:"encryptedCvv"`
	EncryptedPan   string `json:"encryptedPan"`
	IntegrityCheck string `json:"integrityCheck"`
	KeyId          string `json:"keyId"`
	Phase          string `json:"phase"`
	PiHash         string `json:"piHash"`
}

type EncryptCardInfo struct {
	CardNumber string `json:"cardNumber"`
	CardCVV    string `json:"cardCVV"`
}

//Used in AddToCart function
type Cart struct {
	ItemCount int `json:"itemCount"`
}

//Used in PlaceOrderRequest
type VoltagePayment struct {
	PaymentType    string `json:"paymentType"`
	EncryptedCvv   string `json:"encryptedCvv"`
	EncryptedPan   string `json:"encryptedPan"`
	IntegrityCheck string `json:"integrityCheck"`
	KeyId          string `json:"keyId"`
	Phase          string `json:"phase"`
}

//used in SetPaymentInfo
type Payment struct {
	PaymentType    string `json:"paymentType"`
	CardType       string `json:"cardType"`
	FirstName      string `json:"firstName"`
	LastName       string `json:"lastName"`
	AddressLineOne string `json:"addressLineOne"`
	AddressLineTwo string `json:"addressLineTwo"`
	City           string `json:"city"`
	State          string `json:"state"`
	PostalCode     string `json:"postalCode"`
	ExpiryMonth    string `json:"expiryMonth"`
	ExpiryYear     string `json:"expiryYear"`
	Email          string `json:"email"`
	Phone          string `json:"phone"`
	EncryptedPan   string `json:"encryptedPan"`
	EncryptedCvv   string `json:"encryptedCvv"`
	IntegrityCheck string `json:"integrityCheck"`
	KeyId          string `json:"keyId"`
	Phase          string `json:"phase"`
	PiHash         string `json:"piHash"`
}

//Used in GetCartInfo
type StoreList struct {
	Id string `json:"id"`
}

//Part of the GetCartInfoResponse response
type Summary struct {
	SubTotal      string `json:"subTotal"`
	GrandTotal    string `json:"grandTotal"`
	QuantityTotal string `json:"quantityTotal"`
}
