package pokemoncenter

import (
	"backend.juicedbot.io/juiced.infrastructure/common/entities"
	"backend.juicedbot.io/juiced.infrastructure/common/enums"
	"backend.juicedbot.io/juiced.sitescripts/base"
)

// Endpoints
const (
	BaseEndpoint = "https://www.pokemoncenter.com/"

	MonitorEndpoint = "https://www.pokemoncenter.com/product/%s"

	DatadomeEndpoint          = "https://geo.captcha-delivery.com/captcha/?"
	DatadomeChallengeEndpoint = "https://geo.captcha-delivery.com/captcha/check?"

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

	AddToCartQuantityError = "expected quantity of %d but found %d"
)

// Monitor info
type Monitor struct {
	Monitor        base.Monitor
	SKUsSentToTask []string
	OutOfStockSKUs []string
	SKUs           []string
	InStock        []PokemonCenterInStockData
	SKUWithInfo    map[string]entities.PokemonCenterSingleMonitorInfo
}

type MonitorResponse struct {
	Props struct {
		IsServer     bool
		InitialState struct {
			Product struct {
				AddToCartForm string
				Availability  string
				Images        struct {
					Original string
				}
				ListPrice struct {
					Amount float64
				}
				Name string
			}
		}
		InitialProps struct {
		}
	}
}

//Populated when item comes into stock, then populates checkout info
type PokemonCenterInStockData struct {
	SKU             string
	Price           float64
	OutOfPriceRange bool
	ItemName        string
	AddToCartForm   string
	ImageURL        string
}

type Task struct {
	Task            base.Task
	AccountInfo     AccountInfo
	TaskType        enums.TaskType
	CheckoutInfo    CheckoutInfo
	StockData       PokemonCenterInStockData
	CyberSecureInfo CyberSecureInfo
	AccessToken     string
	RefreshAt       int64
}

type DatadomeInfo struct {
	InitialCID string `json:"cid"`
	CID        string `json:"-"`
	Hash       string `json:"hsh"`
	T          string `json:"t"`
	S          int    `json:"s"`
	Host       string `json:"host"`
}

//Info used for payment encryption
type CyberSecureInfo struct {
	PublicKey   string
	PublicToken string
	Privatekey  string
	JtiToken    string
}

//Info used when logging in
type AccountInfo struct {
	Email    string
	Password string
}

//Info used for checkout
type CheckoutInfo struct {
	CheckoutUri string
}

//Used to Retrieve the GuestAuthId
type AccessToken struct {
	AccessToken string
}

//Used in a couple of requests
type Address struct {
	FamilyName      string `json:"familyName"`
	GivenName       string `json:"givenName"`
	StreetAddress   string `json:"streetAddress"`
	ExtendedAddress string `json:"extendedAddress"`
	Locality        string `json:"locality"`
	Region          string `json:"region"`
	PostalCode      string `json:"postalCode"`
	CountryName     string `json:"countryName"`
	PhoneNumber     string `json:"phoneNumber"`
}
