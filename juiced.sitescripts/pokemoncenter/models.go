package pokemoncenter

import (
	"backend.juicedbot.io/juiced.infrastructure/common/entities"
	"backend.juicedbot.io/juiced.sitescripts/base"
)

// Endpoints
const (
	BaseEndpoint = "https://www.pokemoncenter.com/"

	MonitorEndpoint                     = "https://www.pokemoncenter.com/product/%s"
	AddToCartEndpoint                   = "https://www.pokemoncenter.com/tpci-ecommweb-api/cart?type=product&format=zoom.nodatalinks"
	AddToCartRefererEndpoint            = "https://www.pokemoncenter.com/product/%s"
	SubmitAddressEndpoint               = "https://www.pokemoncenter.com/tpci-ecommweb-api/address?format=zoom.nodatalinks"
	SubmitAddresRefererEndpoint         = ""
	SubmitAddressValidateEndpoint       = "https://www.pokemoncenter.com/tpci-ecommweb-api/address/validate"
	SubmitAddresValidateRefererEndpoint = ""
	SubmitPaymentDetailsEndpoint        = "https://www.pokemoncenter.com/tpci-ecommweb-api/payment?microform=true&format=zoom.nodatalinks"
	SubmitEmailEndpoint                 = "https://www.pokemoncenter.com/tpci-ecommweb-api/email?format=zoom.nodatalinks"
	CheckoutEndpoint                    = "https://www.pokemoncenter.com/tpci-ecommweb-api/order?format=zoom.nodatalinks"
	PaymentKeyEndpoint                  = "https://www.pokemoncenter.com/tpci-ecommweb-api/payment/key?microform=true&locale=en-US"
	GetAuthKeyEndpoint                  = "https://www.pokemoncenter.com/tpci-ecommweb-api/cart?format=zoom.nodatalinks"
	CyberSourceTokenEndpoint            = "https://flex.cybersource.com/flex/v2/tokens"
	LoginEndpoint                       = "https://www.pokemoncenter.com/tpci-ecommweb-api/auth?format=zoom.nodatalinks"
)

// Monitor info
type Monitor struct {
	Monitor         base.Monitor
	SKUsSentToTask  []string
	RunningMonitors []string
	OutOfStockSKUs  []string
	SKUs            []string
	InStock         []PokemonCenterInStockData
	SKUWithInfo     map[string]entities.PokemonCenterSingleMonitorInfo
}

type PokemonCenterInStockData struct {
	SKU           string
	Price         float64
	ItemName      string
	AddToCartForm string
	ImageURL      string
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

// Task info
type Task struct {
	Task            base.Task
	CheckoutInfo    CheckoutInfo
	CyberSecureInfo CyberSecureInfo
	AccessToken     string
}

type CyberSecureInfo struct {
	PublicKey   string
	PublicToken string
	Privatekey  string
	JtiToken    string
}

type CheckoutInfo struct {
	AddToCartForm string
	ImageURL      string
	Price         float64
	ItemName      string
	SKU           string
	CheckoutUri   string
}

type AddToCartRequest struct {
	ProductUri    string `json:"productUri"`
	Quantity      int    `json:"quantity"`
	Configuration string `json:"configuration"`
}

type AddToCartResponse struct {
	Type     string `json:"type"`
	Quantity int    `json:"quantity"`
}

type SubmitAddressRequest struct {
	Shipping Address `json:"shipping"`
	Billing  Address `json:"billing"`
}

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

type Email struct {
	Email string `json:"email"`
}

type PaymentKeyResponse struct {
	KeyId string `json:"keyId"`
}

type PaymentToken struct {
	Jti string `json:"jti"`
}

type Encrypt struct {
	Flx struct {
		Jwk struct {
			Kty string
			E   string
			Use string
			N   string
			Kid string
		}
	}
}

type RSA struct {
	Kty string `json:"kty"`
	E   string `json:"e"`
	Use string `json:"use"`
	Kid string `json:"kid"`
	N   string `json:"n"`
}

type Header__ struct {
	Kid string `json:"kid"`
	Jwk RSA    `json:"jwk"`
}

type EncryptedObject struct {
	Context string `json:"context"`
	Index   int    `json:"index"`
	Data    Card   `json:"data"`
}

type Card struct {
	SecurityCode string `json:"securityCode"`
	Number       string `json:"number"`
	Type         string `json:"type"`
	ExpMonth     string `json:"expMonth"`
	ExpYear      string `json:"expYear"`
}

type CheckoutDetailsRequest struct {
	PurchaseFrom string `json:"purchaseForm"`
}

type PaymentDetails struct {
	PaymentDisplay string `json:"paymentDisplay"`
	PaymentKey     string `json:"paymentKey"`
	PaymentToken   string `json:"paymentToken"`
}

type SubmitPaymentResponse struct {
	Self struct {
		Uri string
	}
}

type AccessToken struct {
	Access_token string
}
