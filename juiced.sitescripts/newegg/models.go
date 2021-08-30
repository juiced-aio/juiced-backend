package newegg

import (
	"net/http"
	"net/url"

	"backend.juicedbot.io/juiced.infrastructure/common/entities"
	"backend.juicedbot.io/juiced.sitescripts/base"
)

const (
	BaseEndpoint               = "https://www.newegg.com"
	SecureBaseEndpoint         = "https://secure.newegg.com"
	MonitorEndpoint            = "https://www.newegg.com/product/api/ProductRealtime?ItemNumber=%v"
	AddToCartEndpoint          = "https://www.newegg.com/api/Add2Cart"
	CartEndpoint               = "https://secure.newegg.com/shop/cart"
	PrepareCheckoutEndpoint    = "https://secure.newegg.com/shop/api/CheckoutApi"
	AuthCheckoutEndpoint       = "https://secure.newegg.com/shop/api/AuthLoginApi"
	GuestCheckoutEndpoint      = "https://secure.newegg.com/shop/shoppingLogin?sessionId=%v"
	BaseCheckoutEndpoint       = "https://secure.newegg.com/shop/checkout?sessionId=%v&gf=1"
	SubmitShippingInfoEndpoint = "https://secure.newegg.com/shop/api/AddShippingApi"
	InitOrderEndpoint          = "https://secure.newegg.com/shop/api/InitOrderReviewApi"
	GetPaymentTokenEndpoint    = "https://secure.newegg.com/shop/api/NeweggPayAuth"
	SubmitPaymentInfoEndpoint  = "https://secure.newegg.com/payment/api/SaveCardEdit"
	GetCardTokenEndpoint       = "https://payments.braintree-api.com/graphql"
	PlaceOrderEndpoint         = "https://secure.newegg.com/shop/api/PlaceOrderApi"
	VerifyPaymentEndpoint      = "https://centinelapi.cardinalcommerce.com/V1/Order/JWT/Init"
	VerifyOrderEndpoint        = "https://secure.newegg.com/shop/api/VBVLookupApi"
)

var ParsedBase, _ = url.Parse(BaseEndpoint)

// Monitor info
type Monitor struct {
	Monitor        base.Monitor
	SKUsSentToTask []string
	OutOfStockSKUs []string
	SKUs           []string
	InStock        []NeweggInStockData
	SKUWithInfo    map[string]entities.NeweggSingleMonitorInfo
}

type Task struct {
	Task      base.Task
	TaskInfo  TaskInfo
	StockData NeweggInStockData
}

type NeweggInStockData struct {
	SKU             string
	ItemNumber      string
	ImageURL        string
	ProductName     string
	ItemURL         string
	MaxQuantity     int
	Price           float64
	OutOfPriceRange bool
}

type TaskInfo struct {
	SessionID    string
	PaymentToken string
	VBVToken     string
	CardBin      string
}

type Request struct {
	Client             http.Client
	Method             string
	URL                string
	Headers            http.Header
	RawHeaders         [][2]string
	Referer            string
	Data               []byte
	RequestBodyStruct  interface{}
	ResponseBodyStruct interface{}
	RandOpt            string
}

type MonitorResponse struct {
	MainItem MainItem `json:"MainItem"`
}

type MainItem struct {
	Description Description `json:"Description"`
	FinalPrice  float64     `json:"FinalPrice"`
	InStock     bool        `json:"Instock"`
	MaxQuantity int         `json:"LimitQuantity"`
	Item        string      `json:"Item"`
	Image       Image       `json:"Image"`
}

type Description struct {
	ProductName string `json:"ProductName"`
}

type Image struct {
	Normal Normal `json:"Normal"`
}

type Normal struct {
	ImageName string `json:"ImageName"`
}

type AddToCartRequest struct {
	Itemlist       []Itemlist `json:"ItemList"`
	Customernumber int        `json:"CustomerNumber"`
}
type Itemlist struct {
	Itemgroup     string      `json:"ItemGroup"`
	Itemnumber    string      `json:"ItemNumber"`
	Quantity      int         `json:"Quantity"`
	Optionalinfos interface{} `json:"OptionalInfos,omitempty"`
	Saletype      string      `json:"SaleType,omitempty"`
	ItemKey       string      `json:"ItemKey"`
}

type PrepareCheckoutRequest struct {
	Itemlist []Itemlist    `json:"ItemList"`
	Actions  []interface{} `json:"Actions"`
}

type AuthCheckoutRequest struct {
	Nextpage      string `json:"NextPage"`
	Sitetheme     string `json:"SiteTheme"`
	Sessionid     string `json:"SessionID"`
	Lastloginname string `json:"LastLoginName"`
	Callbackpage  string `json:"Callbackpage"`
}

type SubmitShippingInfoRequest struct {
	Detailinfo       Detailinfo `json:"DetailInfo"`
	Sessionid        string     `json:"SessionID"`
	Isguest          bool       `json:"IsGuest"`
	Isbillingaddress bool       `json:"IsBillingAddress"`
}
type Detailinfo struct {
	Transnumber       int           `json:"TransNumber"`
	Addresslabel      string        `json:"AddressLabel"`
	Contactwith       string        `json:"ContactWith"`
	Phone             string        `json:"Phone"`
	Fax               string        `json:"Fax"`
	Country           string        `json:"Country"`
	State             string        `json:"State"`
	City              string        `json:"City"`
	Address1          string        `json:"Address1"`
	Address2          string        `json:"Address2"`
	Zipcode           string        `json:"ZipCode"`
	Isdefault         bool          `json:"IsDefault"`
	Displaylines      []interface{} `json:"DisplayLines"`
	Email             string        `json:"Email"`
	Disableemail      bool          `json:"DisableEmail"`
	Companyname       string        `json:"CompanyName"`
	Languagecode      string        `json:"LanguageCode"`
	Isselected        bool          `json:"IsSelected"`
	Firstname         string        `json:"FirstName"`
	Lastname          string        `json:"LastName"`
	Addressverifymark string        `json:"AddressVerifyMark"`
	Phoneext          string        `json:"PhoneExt"`
	Saveaddress       bool          `json:"SaveAddress"`
	Editmodel         int           `json:"EditModel"`
}

type InitOrderRequest struct {
	Sessionid        string    `json:"SessionID"`
	Actions          []Actions `json:"Actions"`
	Enableasynctoken bool      `json:"EnableAsyncToken"`
}
type Actions struct {
	Actiontype  string `json:"ActionType"`
	Jsoncontent string `json:"JsonContent"`
}

type SubmitPaymentInfoRequest struct {
	Billing    Billing    `json:"Billing"`
	Creditcard Creditcard `json:"CreditCard"`
	Token      string     `json:"Token"`
}
type Billing struct {
	Address1    string `json:"Address1"`
	Address2    string `json:"Address2"`
	City        string `json:"City"`
	State       string `json:"State"`
	Zipcode     string `json:"ZipCode"`
	Contactwith string `json:"ContactWith"`
	Phone       string `json:"Phone"`
	Fax         string `json:"Fax"`
	Qashashcode string `json:"QASHashCode"`
}
type Creditcard struct {
	Paytermslabel    string `json:"PaytermsLabel"`
	Isdefault        bool   `json:"IsDefault"`
	Holdername       string `json:"HolderName"`
	Creditcardnumber string `json:"CreditCardNumber"`
	Expirationdate   string `json:"ExpirationDate"`
	Paytermscode     string `json:"PaytermsCode"`
	Donotsave        bool   `json:"DoNotSave"`
}

type PlaceOrderRequest struct {
	Sessionid               string          `json:"SessionID"`
	Isacceptnsccauth        bool            `json:"IsAcceptNSCCAuth"`
	Subscribenewsletterflag bool            `json:"SubscribeNewsletterFlag"`
	Creactaccount           bool            `json:"CreactAccount"`
	Password                string          `json:"Password"`
	Mobilesubscribe         Mobilesubscribe `json:"MobileSubscribe"`
	Languagecode            string          `json:"LanguageCode"`
	Cvv2                    string          `json:"Cvv2"`
}
type Mobilesubscribe struct {
}

type PlaceOrderResponse struct {
	Sonumbers            []int          `json:"SONumbers"`
	Addressverifyresults interface{}    `json:"AddressVerifyResults"`
	Notifyitemlist       interface{}    `json:"NotifyItemList"`
	Cardbin              interface{}    `json:"CardBin"`
	Needvbvverify        bool           `json:"NeedVBVVerify"`
	Jwttoken             interface{}    `json:"JwtToken"`
	Paymentcontent       Paymentcontent `json:"PaymentContent"`
	Worldpayinfo         interface{}    `json:"WorldPayInfo"`
	Vbvdata              Vbvdata        `json:"VBVData"`
	Flowersellerconfig   interface{}    `json:"FlowerSellerConfig"`
	Topmessagelist       interface{}    `json:"TopMessageList"`
	Result               string         `json:"Result"`
	Messageentitys       interface{}    `json:"MessageEntitys"`
}
type Paymentcontent struct {
	Transactionnumber   int         `json:"TransactionNumber"`
	Cvv2Confirmed       bool        `json:"Cvv2Confirmed"`
	Cardholder          interface{} `json:"CardHolder"`
	Cardnumber          string      `json:"CardNumber"`
	Realtype            string      `json:"RealType"`
	Cardexpiration      interface{} `json:"CardExpiration"`
	Paytermscode        string      `json:"PaytermsCode"`
	Virtualpaytermscode string      `json:"VirtualPaytermsCode"`
	Iscreditcard        bool        `json:"IsCreditCard"`
	Enabled             bool        `json:"Enabled"`
	Displaymessage      interface{} `json:"DisplayMessage"`
	Disablereason       interface{} `json:"DisableReason"`
	Runsimplecheckout   bool        `json:"RunSimpleCheckout"`
}
type Vbvdata struct {
	Jwttoken      string `json:"JwtToken"`
	Needvbvverify bool   `json:"NeedVBVVerify"`
	Cardbin       string `json:"CardBin"`
	Jslibrary     string `json:"JsLibrary"`
}

type VerifyPaymentRequest struct {
	Browserpayload    Browserpayload `json:"BrowserPayload"`
	Client            Client         `json:"Client"`
	Consumersessionid interface{}    `json:"ConsumerSessionId"`
	Serverjwt         string         `json:"ServerJWT"`
}
type Account struct {
	Accountnumber string `json:"AccountNumber"`
}
type Billingaddress struct {
}
type Shippingaddress struct {
}
type Consumer struct {
	Account         Account         `json:"Account"`
	Billingaddress  Billingaddress  `json:"BillingAddress"`
	Shippingaddress Shippingaddress `json:"ShippingAddress"`
}
type Orderdetails struct {
}
type Token struct {
}
type Authorization struct {
}
type Options struct {
}
type Ccaextension struct {
}
type Order struct {
	Consumer      Consumer      `json:"Consumer"`
	Orderdetails  Orderdetails  `json:"OrderDetails"`
	Cart          []interface{} `json:"Cart"`
	Token         Token         `json:"Token"`
	Authorization Authorization `json:"Authorization"`
	Options       Options       `json:"Options"`
	Ccaextension  Ccaextension  `json:"CCAExtension"`
}
type Supportsalternativepayments struct {
	Cca            bool `json:"cca"`
	Hostedfields   bool `json:"hostedFields"`
	Applepay       bool `json:"applepay"`
	Discoverwallet bool `json:"discoverwallet"`
	Wallet         bool `json:"wallet"`
	Paypal         bool `json:"paypal"`
	Visacheckout   bool `json:"visacheckout"`
}
type Browserpayload struct {
	Order                       Order                       `json:"Order"`
	Supportsalternativepayments Supportsalternativepayments `json:"SupportsAlternativePayments"`
}
type Client struct {
	Agent   string `json:"Agent"`
	Version string `json:"Version"`
}

type VerifyOrderRequest struct {
	Useragent     string `json:"UserAgent"`
	Browserheader string `json:"BrowserHeader"`
	Sessionid     string `json:"SessionID"`
}
