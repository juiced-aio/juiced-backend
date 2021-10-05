package disney

import (
	"backend.juicedbot.io/juiced.infrastructure/entities"
	"backend.juicedbot.io/juiced.infrastructure/enums"
)

const (
	BaseEndpoint    = "https://www.shopdisney.com"
	ProductEndpoint = "https://www.shopdisney.com/%s.html"

	AkamaiEndpoint = "https://www.shopdisney.com/-8EtYaOo6sO7SoYZJZMI/DEp3h2VJ9JLi/PztFCls8Ag/QT5yR/ktqCA4"

	FirstLoginEndpoint  = "https://registerdisney.go.com/jgc/v6/client/DCP-DISNEYSTORE.WEB-PROD/guest/login?langPref=en-US"
	SecondLoginEndpoint = "https://www.shopdisney.com/ocapi/cc/login"

	AddToCartEndpoint = "https://www.shopdisney.com/on/demandware.store/Sites-shopDisney-Site/default/Cart-AddProduct"

	GetCartInfoEndpoint = "https://www.shopdisney.com/my-bag?validateBasket=1"
	GetCartInfoReferer  = "https://www.shopdisney.com/my-bag"

	ValidateCheckoutEndpoint = "https://www.shopdisney.com/ocapi/cc/checkout?validateCheckout=1"
	ValidateCheckoutReferer  = "https://www.shopdisney.com/my-bag"

	CheckoutEndpoint = "https://www.shopdisney.com/checkout"

	ValidateAddressEndpoint = "https://www.shopdisney.com/on/demandware.store/Sites-shopDisney-Site/default/CheckoutAddressServices-ValidateAddress"

	SubmitShippingInfoEndpoint = "https://www.shopdisney.com/on/demandware.store/Sites-shopDisney-Site/default/CheckoutShippingServices-SubmitShipping"
	SubmitShippingInfoReferer  = "https://www.shopdisney.com/checkout?stage=shipping"

	EstablishAppSessionEndpoint = "https://www.shopdisney.com/ocapi/cc/establishappsession"
	EstablishAppSessionReferer  = "https://www.shopdisney.com/checkout?stage=shipping"

	GetPaysheetAEEndpoint = "https://paymentsheet.wdprapps.disney.com/api/v1/checkout/%v"

	GetCardTokenEndpoint = "https://paymentsheet.wdprapps.disney.com/api/v1/card/token"
	GetCardTokenReferer  = "https://paymentsheet.wdprapps.disney.com/embedded/web?fromOrigin=https:%2F%2Fwww.shopdisney.com"

	ValidateBasketEndpoint = "https://www.shopdisney.com/on/demandware.store/Sites-shopDisney-Site/default/Checkout-ValidateBasket"

	PlaceOrderEndpoint = "https://paymentsheet.wdprapps.disney.com/api/v3/process/%v"

	MonitorEndpoint  = "https://www.shopdisney.com/on/demandware.store/Sites-shopDisney-Site/default/Product-Variation?pid=%s&Quantity=1&format=ajax"
	MonitorEndpoint2 = "https://www.shopdisney.com/on/demandware.store/Sites-shopDisney-Site/default/Product-Variation?pid=%s&Quantity=1&format=ajax&dwvar_%s_color=%s&dwvar_%s_size=%s"
)

// Errors
const (
	EmptyCSRFValuesError = "empty csrf input value(s)"
)

type SKUMonitor struct {
	Input       entities.MonitorInput
	DisneyInput MonitorInput
	BaseMonitor *entities.BaseMonitor

	ItemName string
	ImageURL string
	Price    float64
}

type MonitorInput struct {
	Size  string
	Color string
}

type DisneyProductInfo struct {
	VID           string
	Size          string
	Color         string
	QuantityLimit int
	IsPreOrder    bool
	IsBackOrder   bool
}

type Task struct {
	Input       entities.TaskInput
	DisneyInput TaskInput
	BaseTask    *entities.BaseTask

	ShipmentUUID   string
	ShippingMethod string
	CSRF           string
	Total          float64
	PaymentData    PaymentData
}

type TaskInput struct {
	Email    string
	Password string
	TaskType enums.TaskType
}

type SizeInfo struct {
	VID  string
	Size string
}
