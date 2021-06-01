package amazon

import (
	"net/url"
	"time"

	"backend.juicedbot.io/m/v2/juiced.client/http"
	"backend.juicedbot.io/m/v2/juiced.infrastructure/common/entities"
	"backend.juicedbot.io/m/v2/juiced.infrastructure/common/enums"
	"backend.juicedbot.io/m/v2/juiced.infrastructure/common/events"
	"backend.juicedbot.io/m/v2/juiced.sitescripts/base"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// Endpoints
const (
	BaseEndpoint     = "https://www.amazon.com/"
	LoginEndpoint    = "https://www.amazon.com/ap/signin?openid.pape.max_auth_age=900&openid.return_to=https%3A%2F%2Fwww.amazon.com%2Fap%2Fcnep%3Fie%3DUTF8%26orig_return_to%3Dhttps%253A%252F%252Fwww.amazon.com%252Fyour-account%26openid.assoc_handle%3Dusflex%26pageId%3Dusflex&openid.identity=http%3A%2F%2Fspecs.openid.net%2Fauth%2F2.0%2Fidentifier_select&openid.assoc_handle=usflex&openid.mode=checkid_setup&openid.ns.pape=http%3A%2F%2Fspecs.openid.net%2Fextensions%2Fpape%2F1.0&openid.claimed_id=http%3A%2F%2Fspecs.openid.net%2Fauth%2F2.0%2Fidentifier_select&openid.ns=http%3A%2F%2Fspecs.openid.net%2Fauth%2F2.0"
	SignoutEndpoint  = "https://amazon.com/gp/flex/sign-out.html?path=%2Fgp%2Fyourstore%2Fhome&signIn=1&useRedirectOnSuccess=1&action=sign-out&ref_=nav_AccountFlyout_signout"
	AdressesEndpoint = "https://www.amazon.com/a/addresses?ref_=ya_d_l_addr"
	// TestItemEndpoint will be used for testing and also is important for the login process to get the addressID + x-amz-checkout-csrf-token
	TestItemEndpoint = "https://www.amazon.com/Programming-Language-Addison-Wesley-Professional-Computing/dp/0134190440/ref=sr_1_4?dchild=1&keywords=golang&qid=1620785590&s=books&sr=1-4"
	SigninEndpoint   = "https://www.amazon.com/ap/signin"
	BuyNowEndpoint   = "https://www.amazon.com/checkout/turbo-initiate?ref_=dp_start-bbf_1_glance_buyNow_2-1&referrer=detail&pipelineType=turbo&clientId=retailwebsite&weblab=RCX_CHECKOUT_TURBO_DESKTOP_PRIME_87783&temporaryAddToCart=1"
	CheckoutEndpoint = "https://www.amazon.com/checkout/spc/place-order?ref_=chk_spc_placeOrder&_srcRID=%s&clientId=retailwebsite&pipelineType=turbo&cachebuster=%s&pid=%s"
)

// Endpoints for monitoring
var MonitorEndpoints = []string{
	"/dp/%s?m=ATVPDKIKX0DER",
	"/gp/mobile/udp/%s?m=ATVPDKIKX0DER",
	"/gp/product/%s/ref=ox_sc_act_title_1?smid=ATVPDKIKX0DER",
	"/portal-migration/aod?asin=%s&m=ATVPDKIKX0DER",
	"/gp/aod/ajax?asin=%s&m=ATVPDKIKX0DER",
}

var AmazonEndpoints = []string{"https://www.amazon.com", "https://smile.amazon.com"}

var baseURL, _ = url.Parse(BaseEndpoint)

// Monitor info
type Monitor struct {
	Monitor         base.Monitor
	RunningMonitors []string
	OutOfStockASINs []string
	ASINs           []string
	AccountClient   http.Client
	AddressID       string
	SessionID       string
	EventInfo       events.AmazonSingleStockData
	ASINWithInfo    map[string]entities.AmazonSingleMonitorInfo
}

type AccChan struct {
	GroupID     primitive.ObjectID
	Client      http.Client
	AccountInfo AccountInfo
}

type MonitorType string

const (
	MonitorTypeOFID  MonitorType = "OFID"
	MonitorTypeTURBO MonitorType = "TURBO"
)

type LoginType string

const (
	LoginTypeBROWSER  LoginType = "BROWSER"
	LoginTypeREQUESTS LoginType = "REQUESTS"
)

const MetaData1APIKey = ""

type AddHeadersFunction func(*http.Request, ...string)

type Task struct {
	Task         base.Task
	AccountInfo  AccountInfo
	TaskInfo     TaskInfo
	CheckoutInfo CheckoutInfo
}
type AccountInfo struct {
	Email          string
	Password       string
	LoginType      enums.LoginType
	SavedAddressID string
	SessionID      string
}

type TaskInfo struct {
	ASIN        string
	ItemName    string
	OfferID     string
	MonitorType MonitorType
}

type CheckoutInfo struct {
	Price       int
	AntiCsrf    string
	PID         string
	RID         string
	ImageURL    string
	UA          string
	MonitorType enums.MonitorType
}

type Login struct {
	Metadata1 string `json:"metadata1"`
}

type WebHook struct {
	Content interface{} `json:"content"`
	Embeds  []Embeds    `json:"embeds"`
}
type Fields struct {
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
type Embeds struct {
	Title     string    `json:"title"`
	Color     int       `json:"color"`
	Fields    []Fields  `json:"fields"`
	Footer    Footer    `json:"footer"`
	Timestamp time.Time `json:"timestamp"`
	Thumbnail Thumbnail `json:"thumbnail"`
}
