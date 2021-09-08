package pokemoncenter

import (
	"backend.juicedbot.io/juiced.infrastructure/entities"
	"backend.juicedbot.io/juiced.infrastructure/enums"
)

// Endpoints
const (
	BaseEndpoint = "https://www.pokemoncenter.com/"

	MonitorEndpoint = "https://www.pokemoncenter.com/product/%s"

	AddToCartEndpoint                   = "https://www.pokemoncenter.com/tpci-ecommweb-api/cart?type=product&format=zoom.nodatalinks"
	AddToCartRefererEndpoint            = "https://www.pokemoncenter.com/product/%s"
	SubmitAddressEndpoint               = "https://www.pokemoncenter.com/tpci-ecommweb-api/address?format=zoom.nodatalinks"
	SubmitAddresRefererEndpoint         = "https://www.pokemoncenter.com/checkout/address"
	SubmitAddressValidateEndpoint       = "https://www.pokemoncenter.com/tpci-ecommweb-api/address/validate"
	SubmitAddresValidateRefererEndpoint = "https://www.pokemoncenter.com/checkout/address"
	SubmitPaymentDetailsEndpoint        = "https://www.pokemoncenter.com/tpci-ecommweb-api/payment?microform=true&format=zoom.nodatalinks"
	SubmitPaymentDetailsRefererEndpoint = "https://www.pokemoncenter.com/checkout/payment"
	SubmitEmailEndpoint                 = "https://www.pokemoncenter.com/tpci-ecommweb-api/email?format=zoom.nodatalinks"
	SubmitEmailRefererEndpoint          = "https://www.pokemoncenter.com/checkout/address"
	CheckoutEndpoint                    = "https://www.pokemoncenter.com/tpci-ecommweb-api/order?format=zoom.nodatalinks"
	CheckoutRefererEndpoint             = "https://www.pokemoncenter.com/checkout/summary"
	PublicPaymentKeyEndpoint            = "https://www.pokemoncenter.com/tpci-ecommweb-api/payment/key?microform=true&locale=en-US"
	PublicPaymentKeyRefererEndpoint     = "https://www.pokemoncenter.com/"
	AuthKeyEndpoint                     = "https://www.pokemoncenter.com/tpci-ecommweb-api/cart?format=zoom.nodatalinks"
	AuthKeyRefererEndpoint              = "https://www.pokemoncenter.com/cart"
	CyberSourceTokenEndpoint            = "https://flex.cybersource.com/flex/v2/tokens"
	CyberSourceTokenRefererEndpoint     = "https://www.pokemoncenter.com/cart"
	LoginEndpoint                       = "https://www.pokemoncenter.com/tpci-ecommweb-api/auth?format=zoom.nodatalinks"
	LoginRefererEndpoint                = "https://www.pokemoncenter.com/cart"
)

// Errors
const (
	UnknownError = "unknown error"

	RetrieveCyberSourcePublicKeyError    = "could not retrieve CyberSource public key"
	CyberSourceEncryptionError           = "could not perform CyberSource encryption"
	RetrieveCyberSourceTokenError        = "could not retrieve CyberSource token"
	RetrieveCyberSourcePaymentTokenError = "could not retrieve CyberSource payment token"

	AddToCartBadInputError = "add to cart form value is not a string"
	AddToCartQuantityError = "expected quantity of %d but found %d"
)

type SKUMonitor struct {
	Input       entities.MonitorInput
	BaseMonitor *entities.BaseMonitor
}

type Task struct {
	Input              entities.TaskInput
	PokemonCenterInput TaskInput
	BaseTask           *entities.BaseTask
	CyberSecureInfo    CyberSecureInfo
	AccessToken        string
	CheckoutUri        string
}

type TaskInput struct {
	Email       string
	Password    string
	TaskType    enums.TaskType
	AddressType enums.AddressType
}

type CyberSecureInfo struct {
	PublicKey   string
	PublicToken string
	Privatekey  string
	JtiToken    string
}
