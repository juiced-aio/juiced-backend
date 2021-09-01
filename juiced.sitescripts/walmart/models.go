package walmart

import (
	"backend.juicedbot.io/juiced.infrastructure/common/entities"
	"backend.juicedbot.io/juiced.infrastructure/common/enums"
	"backend.juicedbot.io/juiced.sitescripts/base"
	"backend.juicedbot.io/juiced.sitescripts/util"
)

// Endpoints
const (
	BaseEndpoint          = "https://www.walmart.com"
	BlockedToBaseEndpoint = "https://www.walmart.com/blocked?url=Lw=="

	LoginEndpoint = "https://www.walmart.com/account/electrode/api/signin?ref=domain"

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

	MonitorEndpoint = "https://www.walmart.com/terra-firma/item/%v"
)

// Monitor info
type Monitor struct {
	Monitor        base.Monitor
	IDs            []string
	InStockForShip []WalmartInStockData
	IDWithInfo     map[string]entities.WalmartSingleMonitorInfo
	PXValues       util.PXValues
}

type WalmartInStockData struct {
	SKU         string
	OfferID     string
	ProductName string
	ImageURL    string
	Price       float64
	MaxQty      int
}

// Task info
type Task struct {
	Task        base.Task
	TaskType    enums.TaskType
	StockData   WalmartInStockData
	CardInfo    CardInfo
	AccountInfo AccountInfo
	PXValues    util.PXValues
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

type AccountInfo struct {
	Email    string
	Password string
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

type Items struct {
	OfferID              string  `json:"offerId"`
	Name                 string  `json:"name"`
	Price                float64 `json:"price"`
	AvailableQuantity    int     `json:"availableQuantity"`
	USItemID             string  `json:"USItemId"`
	Seller               Seller  `json:"seller"`
	MaxItemCountPerOrder float64 `json:"maxItemCountPerOrder"`
	Assets               Assets  `json:"assets"`
}

type Seller struct {
	ID string `json:"id"`
}

type Primary struct {
	Num100 string `json:"100"`
}
type Assets struct {
	Primary []Primary `json:"primary"`
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

// Used in GetSkuStock
type MonitorResponse struct {
	Payload Payload `json:"payload"`
}

type Payload struct {
	PrimaryProduct string      `json:"primaryProduct"`
	Products       interface{} `json:"products"`
	Offers         interface{} `json:"offers"`
	Images         interface{} `json:"images"`
}

type Product struct {
	ProductAttributes ProductAttributes `json:"productAttributes"`
}

type ProductAttributes struct {
	ProductName string `json:"productName"`
}

type Offer struct {
	Sellerid            string              `json:"sellerId"`
	Pricesinfo          Pricesinfo          `json:"pricesInfo"`
	Productavailability Productavailability `json:"productAvailability"`
	OfferInfo           OfferInfo           `json:"offerInfo"`
}

type Current struct {
	Price float64 `json:"price"`
}

type Pricemap struct {
	Current Current `json:"CURRENT"`
}

type Pricesinfo struct {
	Pricemap Pricemap `json:"priceMap"`
}

type Productavailability struct {
	Availabilitystatus string `json:"availabilityStatus"`
}

type OfferInfo struct {
	OfferID         string          `json:"offerId"`
	QuantityOptions QuantityOptions `json:"quantityOptions"`
}

type QuantityOptions struct {
	OrderLimit int `json:"orderLimit"`
}

type Image struct {
	Type          string        `json:"type"`
	Rank          int           `json:"rank"`
	AssetSizeUrls AssetSizeUrls `json:"assetSizeUrls"`
}

type AssetSizeUrls struct {
	Default string `json:"DEFAULT"`
}

//
