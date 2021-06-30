package gamestop

import (
	"net/url"
	"time"

	"backend.juicedbot.io/juiced.client/http"
	"backend.juicedbot.io/juiced.infrastructure/common/entities"
	"backend.juicedbot.io/juiced.infrastructure/common/enums"
	"backend.juicedbot.io/juiced.sitescripts/base"
)

const (
	BaseEndpoint          = "https://www.gamestop.com"
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

var ParsedBase, _ = url.Parse(BaseEndpoint)

type AddHeadersFunction func(*http.Request, ...string)

// Monitor info
type Monitor struct {
	Monitor         base.Monitor
	SKUsSentToTask  []string
	RunningMonitors []string
	OutOfStockSKUs  []string
	SKUs            []string
	InStock         []GamestopInStockData
	SKUWithInfo     map[string]entities.GamestopSingleMonitorInfo
}

type GamestopInStockData struct {
	SKU        string
	Price      int
	ItemName   string
	PID        string
	ImageURL   string
	ProductURL string
}

var DefaultRawHeaders = [][2]string{
	{"sec-ch-ua", `" Not A;Brand";v="99", "Chromium";v="90", "Google Chrome";v="90"`},
	{"sec-ch-ua-mobile", "?0"},
	{"upgrade-insecure-requests", "1"},
	{"user-agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/90.0.4430.212 Safari/537.36"},
	{"accept", "text/html,application/xhtml+xml,application/xml;q=0.9,image/avif,image/webp,image/apng,*/*;q=0.8,application/signed-exchange;v=b3;q=0.9"},
	{"sec-fetch-site", "none"},
	{"sec-fetch-mode", "navigate"},
	{"sec-fetch-user", "?1"},
	{"sec-fetch-dest", "document"},
	{"accept-encoding", "gzip, deflate, br"},
	{"accept-language", "en-US,en;q=0.9"},
}

type Task struct {
	Task         base.Task
	TaskType     enums.TaskType
	CheckoutInfo CheckoutInfo
	AccountInfo  AccountInfo
}
type AccountInfo struct {
	Email    string
	Password string
}

type CheckoutInfo struct {
	SKUInStock           string
	PID                  string
	Price                int
	ShipmentUUID         string
	OriginalShipmentUUID string
	CSRF                 string
	ProductURL           string
	ImageURL             string
	ItemName             string
}

type LoginResponse struct {
	Action                string                `json:"action"`
	Querystring           string                `json:"queryString"`
	Locale                string                `json:"locale"`
	Authenticatedcustomer Authenticatedcustomer `json:"authenticatedCustomer"`
	Gtmsessiondata        Gtmsessiondata        `json:"gtmSessionData"`
	Success               bool                  `json:"success"`
	Redirecturl           string                `json:"redirectUrl"`
	Loginstatus           Loginstatus           `json:"loginStatus"`
}
type Authenticatedcustomer struct {
}
type User struct {
	Usertype    string `json:"userType"`
	Loginstatus string `json:"loginStatus"`
	Custkey     string `json:"custKey"`
	Membertype  string `json:"memberType"`
}
type Gtmsessiondata struct {
	User  User   `json:"user"`
	Event string `json:"event"`
}
type Customer struct {
}
type Loginstatus struct {
	Success                         bool        `json:"success"`
	Errormessage                    interface{} `json:"errorMessage"`
	Needstobeloggedin               bool        `json:"needsToBeLoggedIn"`
	Needslegacypasswordreset        bool        `json:"needsLegacyPasswordReset"`
	Userexistsinsfcc                bool        `json:"userExistsInSFCC"`
	Userexistsinauthservice         bool        `json:"userExistsInAuthService"`
	Accountlocked                   bool        `json:"accountLocked"`
	Openplayermodal                 bool        `json:"openPlayerModal"`
	Isstoreaccountwithmembership    bool        `json:"isStoreAccountWithMembership"`
	Isstoreaccountwithoutmembership bool        `json:"isStoreAccountWithoutMembership"`
	Enableplayermodalform           bool        `json:"enablePlayerModalForm"`
	Customer                        Customer    `json:"customer"`
}

type MonitorResponse struct {
	Product   Product         `json:"product"`
	Gtmdata   Gtmdata         `json:"gtmData"`
	Mccevents [][]interface{} `json:"__mccEvents"`
}
type Product struct {
	ID                 string `json:"id"`
	Selectedproducturl string `json:"selectedProductUrl"`
}

type Productinfo struct {
	SKU          string `json:"sku"`
	Productid    string `json:"productID"`
	Name         string `json:"name"`
	Availability string `json:"availability"`
}
type Price struct {
	Sellingprice string `json:"sellingPrice"`
	Baseprice    string `json:"basePrice"`
	Currency     string `json:"currency"`
}
type Gtmdata struct {
	Productinfo Productinfo `json:"productInfo"`
	Price       Price       `json:"price"`
}

type AddToCartResponse struct {
	Message string `json:"message"`
}

type AkamaiResponse struct {
	Sensordata string `json:"sensorData"`
	Pixel      string `json:"pixel"`
}

type SensorRequest struct {
	SensorData string `json:"sensor_data"`
}

type PlaceOrderResponse struct {
	Action       string          `json:"action"`
	Querystring  string          `json:"queryString"`
	Locale       string          `json:"locale"`
	Error        bool            `json:"error"`
	Errormessage string          `json:"errorMessage"`
	Mccevents    [][]interface{} `json:"__mccEvents"`
}

// Discord webhook details
type DiscordWebhook struct {
	Content interface{} `json:"content"`
	Embeds  []Embed     `json:"embeds"`
}

type Embed struct {
	Title     string    `json:"title"`
	Color     int       `json:"color"`
	Fields    []Field   `json:"fields"`
	Footer    Footer    `json:"footer"`
	Timestamp time.Time `json:"timestamp"`
	Thumbnail Thumbnail `json:"thumbnail"`
}

type Field struct {
	Name   string `json:"name"`
	Value  string `json:"value"`
	Inline bool   `json:"inline,omitempty"`
}
type Footer struct {
	Text    string `json:"text"`
	IconURL string `json:"icon_url"`
}

type Thumbnail struct {
	URL string `json:"url"`
}
