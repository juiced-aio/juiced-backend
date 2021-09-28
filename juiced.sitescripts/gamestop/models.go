package gamestop

import (
	"backend.juicedbot.io/juiced.infrastructure/entities"
	"backend.juicedbot.io/juiced.infrastructure/enums"
)

// Endpoints
const (
	BaseEndpoint          = "https://www.gamestop.com"
	BaseLoginEndpoint     = "https://www.gamestop.com/login"
	LoginEndpoint         = "https://www.gamestop.com/on/demandware.store/Sites-gamestop-us-Site/default/Account-Login"
	AccountEndpoint       = "https://www.gamestop.com/account"
	ProductEndpoint       = "https://www.gamestop.com/products/%s.html"
	MonitorEndpoint       = "https://www.gamestop.com/on/demandware.store/Sites-gamestop-us-Site/default/Product-Variation?pid=%s"
	GenEndpoint           = "https://notanapi.com/api/akamai/get/sensor-data"
	AkamaiEndpoint        = "https://www.gamestop.com/2qvO5_xmdvFlK/HTu3HSUrmgZO/X0/uOOtQf1E/ORolEQ4D/ChJ/jfhxDHXg"
	AddToCartEndpoint     = "https://www.gamestop.com/on/demandware.store/Sites-gamestop-us-Site/default/Cart-AddProduct?redesignFlag=true&productID=%s"
	CheckoutLoginEndpoint = "https://www.gamestop.com/checkout/login"
	CheckoutEndpoint      = "https://www.gamestop.com/spcheckout"
	ShippingInfoEndpoint  = "https://www.gamestop.com/on/demandware.store/Sites-gamestop-us-Site/default/CheckoutServices-Get"
	ShippingEndpoint      = "https://www.gamestop.com/on/demandware.store/Sites-gamestop-us-Site/default/CheckoutShippingServices-SubmitShipping"
	PaymentEndpoint       = "https://www.gamestop.com/on/demandware.store/Sites-gamestop-us-Site/default/CheckoutServices-SubmitPayment?isFromSPC=true"
	PlaceOrderEndpoint    = "https://www.gamestop.com/on/demandware.store/Sites-gamestop-us-Site/default/CheckoutServices-PlaceOrder"
)

type SKUMonitor struct {
	Input         entities.MonitorInput
	GameStopInput MonitorInput
	BaseMonitor   *entities.BaseMonitor
}

type MonitorInput struct {
	Size      string
	Color     string
	Condition string
}

type GameStopProductInfo struct {
	VID         string
	Size        string
	Color       string
	Condition   string
	MaxQuantity int
}

type Task struct {
	Input         entities.TaskInput
	GameStopInput TaskInput
	BaseTask      *entities.BaseTask

	CaptchaProtected     bool
	CaptchaToken         string
	ShipmentUUID         string
	OriginalShipmentUUID string
	CSRF                 string
}

type TaskInput struct {
	Email    string
	Password string
	TaskType enums.TaskType
}

type SizeInfo struct {
	SizePID string
	Size    string
}
