package topps

import (
	"net/url"

	"backend.juicedbot.io/juiced.infrastructure/common/entities"
	"backend.juicedbot.io/juiced.infrastructure/common/enums"
	"backend.juicedbot.io/juiced.sitescripts/base"
	"backend.juicedbot.io/juiced.sitescripts/hawk-go"
)

const (
	BaseEndpoint                      = "https://www.topps.com"
	BaseLoginEndpoint                 = "https://www.topps.com/customer/account/login"
	LoginEndpoint                     = "https://www.topps.com/customer/account/loginPost"
	MonitorEndpoint                   = "https://www.topps.com/%v.html"
	GetCartInfoEndpoint               = "https://www.topps.com/customer/section/load/?sections=cart%2Cdirectory-data%2Cammessages%2Cgtm%2Cmessages&force_new_section_timestamp=true&_="
	CheckoutEndpoint                  = "https://www.topps.com/checkout"
	SubmitShippingInfoEndpoint        = "https://www.topps.com/rest/default/V1/guest-carts/%v/shipping-information"
	AccountSubmitShippingInfoEndpoint = "https://www.topps.com/rest/default/V1/carts/mine/shipping-information"
	GetCardTokenEndpoint              = "https://payments.braintree-api.com/graphql"
	AnalyticsEndpoint                 = "https://client-analytics.braintreegateway.com/x5fsg388mmc9s2hz"
	PlaceOrderEndpoint                = "https://www.topps.com/rest/default/V1/guest-carts/%v/payment-information"
	AccountPlaceOrderEndpoint         = "https://www.topps.com/rest/default/V1/carts/mine/payment-information"
)

var ParsedBase, _ = url.Parse(BaseEndpoint)

// Monitor info
type Monitor struct {
	Monitor         base.Monitor
	ItemsSentToTask []string
	RunningMonitors []string
	OutOfStockItems []string
	Items           []string
	InStock         []ToppsInStockData
	ItemWithInfo    map[string]entities.ToppsSingleMonitorInfo
}

type Task struct {
	Task        base.Task
	TaskType    enums.TaskType
	AccountInfo AccountInfo
	TaskInfo    TaskInfo
	StockData   ToppsInStockData
}

type Acc struct {
	GroupID     string
	Scraper     hawk.Scraper
	AccountInfo AccountInfo
}

type AccountInfo struct {
	Email    string
	Password string
}

type TaskInfo struct {
	RelicID    string
	CheckoutID string
	CardToken  string
	AuthToken  string
}

type ToppsInStockData struct {
	SKU         string
	FormKey     string
	OptionName  string
	OptionID    string
	ImageURL    string
	ProductName string
	Item        string
	ItemURL     string
	AddURL      string
	Price       float64
}

type Regions struct {
	RegionID []RegionID `json:"region_id"`
}

type RegionID struct {
	Title     string `json:"title"`
	Value     string `json:"value"`
	Label     string `json:"label"`
	CountryID string `json:"country_id,omitempty"`
}

type Option struct {
	ID    string
	Price float64
}

type GetCartInfoResponse struct {
	Cart Cart `json:"cart"`
}

type DataMage struct {
	PaypalBraintreeJsApplepayImplementationsShortcut PaypalBraintreeJsApplepayImplementationsShortcut `json:"PayPal_Braintree/js/applepay/implementations/shortcut"`
}
type PaypalBraintreeJsApplepayImplementationsShortcut struct {
	ID               string `json:"id"`
	Clienttoken      string `json:"clientToken"`
	Quoteid          string `json:"quoteId"`
	Displayname      string `json:"displayName"`
	Actionsuccess    string `json:"actionSuccess"`
	Grandtotalamount string `json:"grandTotalAmount"`
	Isloggedin       bool   `json:"isLoggedIn"`
	Storecode        string `json:"storeCode"`
}

type Cart struct {
	ExtraActions string `json:"extra_actions"`
}

type BrainTreeConfig struct {
	Authorizationfingerprint string `json:"authorizationFingerprint"`
}

type SubmitShippingInfoRequest struct {
	Addressinformation Addressinformation `json:"addressInformation"`
}

type ShippingAddress struct {
	Countryid         string   `json:"countryId"`
	Regionid          string   `json:"regionId"`
	Regioncode        string   `json:"regionCode"`
	Region            string   `json:"region"`
	Street            []string `json:"street"`
	Company           string   `json:"company"`
	Telephone         string   `json:"telephone"`
	Fax               string   `json:"fax"`
	Postcode          string   `json:"postcode"`
	City              string   `json:"city"`
	Firstname         string   `json:"firstname"`
	Lastname          string   `json:"lastname"`
	SaveInAddressBook int      `json:"save_in_address_book"`
}
type BillingAddress struct {
	Countryid         string      `json:"countryId"`
	Regionid          string      `json:"regionId"`
	Regioncode        string      `json:"regionCode"`
	Region            string      `json:"region"`
	Street            []string    `json:"street"`
	Company           string      `json:"company"`
	Telephone         string      `json:"telephone"`
	Fax               string      `json:"fax"`
	Postcode          string      `json:"postcode"`
	City              string      `json:"city"`
	Firstname         string      `json:"firstname"`
	Lastname          string      `json:"lastname"`
	SaveInAddressBook int         `json:"save_in_address_book"`
	Saveinaddressbook interface{} `json:"saveInAddressBook"`
}
type ExtensionAttributes struct {
}
type Addressinformation struct {
	ShippingAddress     ShippingAddress     `json:"shipping_address"`
	BillingAddress      BillingAddress      `json:"billing_address"`
	ShippingMethodCode  string              `json:"shipping_method_code"`
	ShippingCarrierCode string              `json:"shipping_carrier_code"`
	ExtensionAttributes ExtensionAttributes `json:"extension_attributes"`
}

type SetPaymentInfoRequest struct {
	Cartid         string         `json:"cartId"`
	Billingaddress Billingaddress `json:"billingAddress"`
	Paymentmethod  Paymentmethod  `json:"paymentMethod"`
	Email          string         `json:"email"`
}

type TotalsInformationRequest struct {
	Addressinformation Addressinformation2 `json:"addressInformation"`
}

type AdvancedConditions struct {
	PaymentMethod         string   `json:"payment_method"`
	City                  string   `json:"city"`
	ShippingAddressLine   []string `json:"shipping_address_line"`
	BillingAddressCountry string   `json:"billing_address_country"`
	Currency              string   `json:"currency"`
}
type ExtensionAttributes2 struct {
	AdvancedConditions AdvancedConditions `json:"advanced_conditions"`
}

type Addressinformation2 struct {
	Address             Address `json:"address"`
	ShippingMethodCode  string  `json:"shipping_method_code"`
	ShippingCarrierCode string  `json:"shipping_carrier_code"`
}

type Address struct {
	Countryid           string               `json:"countryId"`
	Region              string               `json:"region"`
	Regionid            string               `json:"regionId"`
	Postcode            string               `json:"postcode"`
	City                string               `json:"city"`
	ExtensionAttributes ExtensionAttributes2 `json:"extension_attributes"`
}

type JwtJson struct {
	Exp      int64    `json:"exp"`
	Jti      string   `json:"jti"`
	Sub      string   `json:"sub"`
	Iss      string   `json:"iss"`
	Merchant Merchant `json:"merchant"`
	Rights   []string `json:"rights"`
	Scope    []string `json:"scope"`
	Options  Options2 `json:"options"`
}
type Merchant struct {
	PublicID            string `json:"public_id"`
	VerifyCardByDefault bool   `json:"verify_card_by_default"`
}
type Options2 struct {
	MerchantAccountID string `json:"merchant_account_id"`
}

type GetCardTokenRequest struct {
	Clientsdkmetadata Clientsdkmetadata `json:"clientSdkMetadata"`
	Query             string            `json:"query"`
	Variables         Variables         `json:"variables"`
	Operationname     string            `json:"operationName"`
}
type Clientsdkmetadata struct {
	Source      string `json:"source"`
	Integration string `json:"integration"`
	Sessionid   string `json:"sessionId"`
}
type Creditcard struct {
	Number          string `json:"number"`
	Expirationmonth string `json:"expirationMonth"`
	Expirationyear  string `json:"expirationYear"`
	Cvv             string `json:"cvv"`
}
type Options struct {
	Validate bool `json:"validate"`
}
type Input struct {
	Creditcard Creditcard `json:"creditCard"`
	Options    Options    `json:"options"`
}
type Variables struct {
	Input Input `json:"input"`
}

type AnalyticsRequest struct {
	Analytics                []Analytics `json:"analytics"`
	Braintreelibraryversion  string      `json:"braintreeLibraryVersion"`
	Meta                     Meta        `json:"_meta"`
	Authorizationfingerprint string      `json:"authorizationFingerprint"`
}
type Analytics struct {
	Kind      string `json:"kind"`
	Isasync   bool   `json:"isAsync"`
	Timestamp int64  `json:"timestamp"`
}
type Meta struct {
	Merchantappid   string `json:"merchantAppId"`
	Platform        string `json:"platform"`
	Sdkversion      string `json:"sdkVersion"`
	Source          string `json:"source"`
	Integration     string `json:"integration"`
	Integrationtype string `json:"integrationType"`
	Sessionid       string `json:"sessionId"`
}

type GetCardTokenResponse struct {
	Data Data `json:"data"`
}
type Tokenizecreditcard struct {
	Token string `json:"token"`
}
type Data struct {
	Tokenizecreditcard Tokenizecreditcard `json:"tokenizeCreditCard"`
}

type PlaceOrderRequest struct {
	Cartid         string         `json:"cartId"`
	Billingaddress Billingaddress `json:"billingAddress"`
	Paymentmethod  Paymentmethod  `json:"paymentMethod"`
	Email          string         `json:"email"`
}
type Billingaddress struct {
	Countryid         string      `json:"countryId"`
	Regionid          string      `json:"regionId"`
	Regioncode        string      `json:"regionCode"`
	Region            string      `json:"region"`
	Street            []string    `json:"street"`
	Company           string      `json:"company"`
	Telephone         string      `json:"telephone"`
	Fax               string      `json:"fax"`
	Postcode          string      `json:"postcode"`
	City              string      `json:"city"`
	Firstname         string      `json:"firstname"`
	Lastname          string      `json:"lastname"`
	Saveinaddressbook interface{} `json:"saveInAddressBook"`
}
type AdditionalData struct {
	PaymentMethodNonce string `json:"payment_method_nonce,omitempty"`
	DeviceData         string `json:"device_data,omitempty"`
	AmgdprAgreement    string `json:"amgdpr_agreement"`
}
type Paymentmethod struct {
	Method         string         `json:"method"`
	AdditionalData AdditionalData `json:"additional_data"`
}

type PlaceOrderResponse struct {
	Message string `json:"message"`
}
