package disney

import (
	"net/url"

	"backend.juicedbot.io/juiced.infrastructure/common/entities"
	"backend.juicedbot.io/juiced.infrastructure/common/enums"
	"backend.juicedbot.io/juiced.sitescripts/base"
)

const (
	BaseEndpoint                = "https://www.shopdisney.com"
	AkamaiEndpoint              = "https://www.shopdisney.com/-8EtYaOo6sO7SoYZJZMI/DEp3h2VJ9JLi/PztFCls8Ag/QT5yR/ktqCA4"
	MonitorEndpoint             = "https://www.shopdisney.com/on/demandware.store/Sites-shopDisney-Site/default/Product-Variation"
	FirstLoginEndpoint          = "https://registerdisney.go.com/jgc/v6/client/DCP-DISNEYSTORE.WEB-PROD/guest/login?langPref=en-US"
	SecondLoginEndpoint         = "https://www.shopdisney.com/ocapi/cc/login"
	AddToCartEndpoint           = "https://www.shopdisney.com/on/demandware.store/Sites-shopDisney-Site/default/Cart-AddProduct"
	CartEndpoint                = "https://www.shopdisney.com/my-bag"
	GetCheckoutInfoEndpoint     = "https://www.shopdisney.com/my-bag?validateBasket=1"
	ValidateCheckoutEndpoint    = "https://www.shopdisney.com/ocapi/cc/checkout?validateCheckout=1"
	CheckoutEndpoint            = "https://www.shopdisney.com/checkout"
	ValidateAddressEndpoint     = "https://www.shopdisney.com/on/demandware.store/Sites-shopDisney-Site/default/CheckoutAddressServices-ValidateAddress"
	SubmitShippingInfoEndpoint  = "https://www.shopdisney.com/on/demandware.store/Sites-shopDisney-Site/default/CheckoutShippingServices-SubmitShipping"
	EstablishAppSessionEndpoint = "https://www.shopdisney.com/ocapi/cc/establishappsession"
	GetPaysheetAEEndpoint       = "https://paymentsheet.wdprapps.disney.com/api/v1/checkout/%v"
	GetCardTokenEndpoint        = "https://paymentsheet.wdprapps.disney.com/api/v1/card/token"
	ValidateBasketEndpoint      = "https://www.shopdisney.com/on/demandware.store/Sites-shopDisney-Site/default/Checkout-ValidateBasket"
	PlaceOrderEndpoint          = "https://paymentsheet.wdprapps.disney.com/api/v3/process/%v"
)

var ParsedBase, _ = url.Parse(BaseEndpoint)

// Task info
type Task struct {
	Task        base.Task
	TaskType    enums.TaskType
	StockData   DisneyInStockData
	AccountInfo AccountInfo
	PaymentData PaymentData
	TaskInfo    TaskInfo
}

// Monitor info
type Monitor struct {
	Monitor         base.Monitor
	RunningMonitors []string
	Pids            []string
	InStock         []DisneyInStockData
	PidWithInfo     map[string]entities.DisneySingleMonitorInfo
}

type DisneyInStockData struct {
	PID         string
	VID         string
	Size        string
	Color       string
	ProductName string
	ItemURL     string
	ImageURL    string
}

type AccountInfo struct {
	Email    string
	Password string
}

type TaskInfo struct {
	Token          string
	ShipmentUUID   string
	ShippingMethod string
	CsrfToken      string
	Total          float64
}

type LoginRequest struct {
	Loginvalue string `json:"loginValue"`
	Password   string `json:"password"`
}

type LoginResponse struct {
	Data  Data        `json:"data"`
	Error interface{} `json:"error"`
}
type Profile struct {
	Swid string `json:"swid"`
}

type Token struct {
	AccessToken             string      `json:"access_token"`
	RefreshToken            string      `json:"refresh_token"`
	Swid                    string      `json:"swid"`
	TTL                     int         `json:"ttl"`
	RefreshTTL              int         `json:"refresh_ttl"`
	HighTrustExpiresIn      int         `json:"high_trust_expires_in"`
	InitialGrantInChainTime int64       `json:"initial_grant_in_chain_time"`
	Iat                     int64       `json:"iat"`
	Exp                     int64       `json:"exp"`
	RefreshExp              int64       `json:"refresh_exp"`
	HighTrustExp            int64       `json:"high_trust_exp"`
	Scope                   interface{} `json:"scope"`
	Sso                     interface{} `json:"sso"`
	IDToken                 interface{} `json:"id_token"`
	Authenticator           string      `json:"authenticator"`
	Loginvalue              interface{} `json:"loginValue"`
	Clickbacktype           interface{} `json:"clickbackType"`
	Sessiontransferkey      string      `json:"sessionTransferKey"`
	Created                 string      `json:"created"`
	LastChecked             string      `json:"lastChecked"`
	Expires                 string      `json:"expires"`
	RefreshExpires          string      `json:"refresh_expires"`
	BlueCookie              interface{} `json:"blue_cookie"`
}
type Data struct {
	Profile Profile `json:"profile"`
	Token   Token   `json:"token"`
}

type MonitorResponse struct {
	Product Product `json:"product"`
}

type Product struct {
	ProductName        string `json:"productName"`
	ID                 string `json:"id"`
	Price              Price  `json:"price"`
	Images             Images `json:"images"`
	Available          bool   `json:"available"`
	SelectedProductUrl string `json:"selectedProductUrl"`
}

type Price struct {
	Sales Sales `json:"sales"`
}

type Sales struct {
	Value float64 `json:"value"`
}

type Images struct {
	Small   []Small   `json:"small"`
	HighRes []HighRes `json:"highRes"`
}

type Small struct {
	URL string `json:"url"`
}

type HighRes struct {
	URL string `json:"url"`
}

type AddToCartResponse struct {
	Message string `json:"message"`
}

type GetCheckoutInfoResponse struct {
	Csrf      Csrf        `json:"csrf"`
	Items     []Items     `json:"items"`
	Shipments []Shipments `json:"shipments"`
}

type Csrf struct {
	Token string `json:"token"`
}

type Items struct {
	ShipmentUUID string `json:"shipmentUUID"`
}

type Shipments struct {
	SelectedShippingMethod string `json:"selectedShippingMethod"`
}

type EstablishAppSessionResponse struct {
	PaymentData PaymentData `json:"paymentData"`
	Order       Order       `json:"order"`
}

type Order struct {
	Totals Totals `json:"totals"`
}

type Totals struct {
	GrandTotalValue GrandTotalValue `json:"grandTotalValue"`
}

type GrandTotalValue struct {
	Value float64 `json:"value"`
}
type PaymentData struct {
	Config   Config `json:"config"`
	CardInfo GetCardTokenResponse
}

type Config struct {
	Client                string `json:"client"`
	Domain                string `json:"domain"`
	Lang                  string `json:"lang"`
	Oauth                 string `json:"oauth"`
	Session               string `json:"session"`
	ConversationID        string `json:"conversationId"`
	Hmac                  string `json:"hmac"`
	HmacVersion           int    `json:"hmacVersion"`
	InspireFont           bool   `json:"inspireFont"`
	AllowWideLayout       bool   `json:"allowWideLayout"`
	ShowErrorIcons        bool   `json:"showErrorIcons"`
	ShowRequiredAsterisks bool   `json:"showRequiredAsterisks"`
	CSSOverride           string `json:"cssOverride"`
	EnableCSSOverride     bool   `json:"enableCssOverride"`
	PaymentAE             string
}

type GetCardTokenRequest struct {
	Pan  string `json:"pan"`
	Type string `json:"type"`
}

type GetCardTokenResponse struct {
	Token           string `json:"token"`
	Tokenvault      string `json:"tokenVault"`
	Issuer          string `json:"issuer"`
	Masked          string `json:"masked"`
	Type            string `json:"type"`
	Iswalletdefault bool   `json:"isWalletDefault"`
	Cardproductname string `json:"cardProductName"`
	Isdisneyvisa    bool   `json:"isDisneyVisa"`
}

// Everything akamai/sensor data related will be removed/replaced in the future once we have the akamai endpoint
type SensorDataRequest struct {
	BaseURL      string `json:"baseURL"`
	UserAgent    string `json:"userAgent"`
	PostIndx     int    `json:"postIndx"`
	Cookie       string `json:"cookie"`
	SavedD3      int64  `json:"savedD3"`
	SavedStartTs int64  `json:"savedStartTs"`
}

type SensorDataResponse struct {
	SensorData   string `json:"sensorData"`
	SavedD3      int64  `json:"savedD3"`
	SavedStartTs int64  `json:"savedStartTs"`
}

type SensorPostRequest struct {
	SensorData string `json:"sensor_data"`
}

type PlaceOrderRequest struct {
	Cards           []Cards           `json:"cards"`
	Fraudparameters []Fraudparameters `json:"fraudParameters"`
}
type Profilename struct {
}
type Profilephone struct {
}
type Billingaddress struct {
	Country    string `json:"country"`
	Line1      string `json:"line1"`
	Line2      string `json:"line2"`
	City       string `json:"city"`
	State      string `json:"state"`
	Postalcode string `json:"postalCode"`
}
type Cards struct {
	Securitycode      string         `json:"securityCode"`
	Cardproductname   string         `json:"cardProductName"`
	Isdisneyvisa      bool           `json:"isDisneyVisa"`
	Cardtoken         string         `json:"cardToken"`
	Carddisplaynumber string         `json:"cardDisplayNumber"`
	Expiration        string         `json:"expiration"`
	Cardholdername    string         `json:"cardholderName"`
	Profilename       Profilename    `json:"profileName"`
	Profilephone      Profilephone   `json:"profilePhone"`
	Billingaddress    Billingaddress `json:"billingAddress"`
	Savetoprofile     bool           `json:"saveToProfile"`
	Iswalletdefault   bool           `json:"isWalletDefault"`
	Iswalletedit      bool           `json:"isWalletEdit"`
	Issuer            string         `json:"issuer"`
	Type              string         `json:"type"`
	Isstoredcard      bool           `json:"isStoredCard"`
	Amount            float64        `json:"amount"`
}
type Params struct {
	EmailAddress string `json:"email_address,omitempty"`
	DayPhone     string `json:"day_phone,omitempty"`
}

type Fraudparameters struct {
	Group  string `json:"group"`
	Params Params `json:"params"`
}

type PlaceOrderResponse struct {
	Paidamount        int     `json:"paidAmount"`
	Unpaidamount      float64 `json:"unpaidAmount"`
	Paidthisrequest   int     `json:"paidThisRequest"`
	Suggestederrorkey string  `json:"suggestedErrorKey"`
	Errors            Errors  `json:"errors"`
}

type Errors struct {
}
