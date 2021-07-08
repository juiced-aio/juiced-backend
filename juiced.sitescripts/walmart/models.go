package walmart

import (
	"backend.juicedbot.io/juiced.infrastructure/common/enums"
	"backend.juicedbot.io/juiced.sitescripts/base"
	"backend.juicedbot.io/juiced.sitescripts/util"
)

// Endpoints
const (
	BaseEndpoint          = "https://www.walmart.com/"
	BlockedToBaseEndpoint = "https://www.walmart.com/blocked?url=Lw=="

	PIEEndpoint = "https://securedataweb.walmart.com/pie/v1/wmcom_us_vtg_pie/getkey.js?bust="
	PIEReferer  = "https://www.walmart.com/"

	AddToCartEndpoint = "https://www.walmart.com/api/v3/cart/guest/:CID/items"
	AddToCartReferer  = "https://www.walmart.com/"

	GetCartInfoEndpoint = "https://www.walmart.com/api/checkout/v3/contract?page=CHECKOUT_VIEW"
	GetCartInfoReferer  = "https://www.walmart.com/checkout/"

	SetPcidEndpoint = "https://www.walmart.com/api/checkout/v3/contract/:PCID"
	SetPcidReferer  = "https://www.walmart.com/checkout/"

	SetShippingInfoEndpoint = "https://www.walmart.com/api/checkout/v3/contract/:PCID/shipping-address"
	SetShippingInfoReferer  = "https://www.walmart.com/checkout/"

	SetCreditCardEndpoint = "https://www.walmart.com/api/checkout-customer/:CID/credit-card"
	SetCreditCardReferer  = "https://www.walmart.com/checkout/"

	SetPaymentInfoEndpoint = "https://www.walmart.com/api/checkout/v3/contract/:PCID/payment"
	SetPaymentInfoReferer  = "https://www.walmart.com/checkout/"

	PlaceOrderEndpoint = "https://www.walmart.com/api/checkout/v3/contract/:PCID/order"
	PlaceOrderReferer  = "https://www.walmart.com/checkout/"

	MonitorEndpoint = "https://www.walmart.com/ip/%s/sellers"
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
	Phase          int    `json:"phase"`
	PiHash         string `json:"piHash"`
	PaymentType    string `json:"paymentType"`
}

type EncryptCardInfo struct {
	CardNumber string    `json:"cardNumber"`
	CardCVV    string    `json:"cardCVV"`
	PIEValues  PIEValues `json:"PIE"`
}

type PIEValues struct {
	L     int    `json:"L"`
	E     int    `json:"E"`
	K     string `json:"K"`
	KeyID string `json:"key_id"`
	Phase int    `json:"phase"`
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
	Phase          int    `json:"phase"`
}

// used in SetCreditCard
type Payment struct {
	EncryptedPan   string `json:"encryptedPan"`
	EncryptedCvv   string `json:"encryptedCvv"`
	IntegrityCheck string `json:"integrityCheck"`
	KeyId          string `json:"keyId"`
	Phase          int    `json:"phase"`
	State          string `json:"state"`
	PostalCode     string `json:"postalCode"`
	AddressLineOne string `json:"addressLineOne"`
	AddressLineTwo string `json:"addressLineTwo"`
	City           string `json:"city"`
	AddressType    string `json:"addressType"`
	FirstName      string `json:"firstName"`
	LastName       string `json:"lastName"`
	ExpiryMonth    string `json:"expiryMonth"`
	ExpiryYear     string `json:"expiryYear"`
	Phone          string `json:"phone"`
	CardType       string `json:"cardType"`
	IsGuest        bool   `json:"isGuest"`
}

//used in SetPaymentInfo
type SubmitPayment struct {
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
	Phase          int    `json:"phase"`
	PiHash         string `json:"piHash"`
}

//Used in GetCartInfo
type StoreList struct {
	Id string `json:"id"`
}

//Part of the GetCartInfoResponse response
type Summary struct {
	SubTotal      float64 `json:"subTotal"`
	GrandTotal    float64 `json:"grandTotal"`
	QuantityTotal int     `json:"quantityTotal"`
}
