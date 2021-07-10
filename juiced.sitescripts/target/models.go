package target

import (
	"time"

	"github.com/dgrijalva/jwt-go"

	"backend.juicedbot.io/juiced.client/http"
	"backend.juicedbot.io/juiced.infrastructure/common/entities"
	"backend.juicedbot.io/juiced.infrastructure/common/enums"
	"backend.juicedbot.io/juiced.sitescripts/base"
	cmap "github.com/orcaman/concurrent-map"
)

// Endpoints
const (
	BaseEndpoint                = "https://www.target.com"
	GetTCINStockEndpoint        = "https://redsky.target.com/redsky_aggregations/v1/web/plp_fulfillment_v1?"
	GetTCINStockHost            = "redsky.target.com"
	GetTCINStockReferer         = "https://www.target.com/"
	TCINInfoEndpoint            = "https://redsky.target.com/redsky_aggregations/v1/web/pdp_client_v1?"
	TCINInfoReferer             = "https://www.target.com/p/-/A-"
	LoginEndpoint               = "https://gsp.target.com/gsp/authentications/v1/auth_codes?client_id=ecom-web-1.0.0&state=1619237851891&redirect_uri=https%3A%2F%2Fwww.target.com%2F&assurance_level=M"
	RefreshLoginEndpoint        = "https://gsp.target.com/gsp/oauth_tokens/v2/client_tokens"
	RefreshLoginReferer         = "https://www.target.com/"
	AddToCartEndpoint           = "https://carts.target.com/web_checkouts/v1/cart_items?field_groups=CART%2CCART_ITEMS%2CSUMMARY&key=feaf228eb2777fd3eee0fd5192ae7107d6224b39"
	AddToCartReferer            = "https://www.target.com/p/-/A-"
	GetCartInfoEndpoint         = "https://carts.target.com/web_checkouts/v1/pre_checkout?field_groups=PAYMENT_INSTRUCTIONS%2CCART_ITEMS&key=feaf228eb2777fd3eee0fd5192ae7107d6224b39"
	GetCartInfoReferer          = "https://www.target.com/co-review?precheckout=true"
	SetShippingInfoEndpoint     = "https://carts.target.com/web_checkouts/v1/cart_shipping_addresses/%s?field_groups=ADDRESSES%%2CPAYMENT_INSTRUCTIONS&key=feaf228eb2777fd3eee0fd5192ae7107d6224b39"
	SetShippingInfoReferer      = "https://www.target.com/co-delivery"
	SetPaymentInfoSAVEDEndpoint = "https://carts.target.com/checkout_payments/v1/payment_instructions/%s/?key=feaf228eb2777fd3eee0fd5192ae7107d6224b39"
	SetPaymentInfoNEWEndpoint   = "https://carts.target.com/checkout_payments/v1/payment_instructions/%s?key=feaf228eb2777fd3eee0fd5192ae7107d6224b39"
	SetPaymentInfoSAVEDReferer  = "https://www.target.com/co-review"
	SetPaymentInfoNEWReferer    = "https://www.target.com/co-payment"
	PlaceOrderEndpoint          = "https://carts.target.com/web_checkouts/v1/checkout?field_groups=ADDRESSES%2CCART%2CCART_ITEMS%2CDELIVERY_WINDOWS%2CPAYMENT_INSTRUCTIONS%2CPICKUP_INSTRUCTIONS%2CPROMOTION_CODES%2CSUMMARY&key=feaf228eb2777fd3eee0fd5192ae7107d6224b39"
	PlaceOrderReferer           = "https://www.target.com/co-payment"
	TargetCancelMethodEndpoint  = "https://api.target.com/rum_analytics/v1"
	TargetCancelMethodReferer   = "https://www.target.com/assets/commerce/3e5063091c6d0decffbe.worker.js"
)

// Monitor info
type Monitor struct {
	Monitor          base.Monitor
	MonitorType      enums.MonitorType
	TCINs            []string
	StoreID          string
	TCINsWithInfo    map[string]entities.TargetSingleMonitorInfo
	InStockForShip   cmap.ConcurrentMap
	InStockForPickup cmap.ConcurrentMap
}

// Task info
type Task struct {
	Task         base.Task
	CheckoutType enums.CheckoutType
	AccountInfo  AccountInfo
	InStockData  SingleStockData
	TCIN         string
	TCINType     enums.CheckoutType
}

type TargetStockData struct {
	InStockForShip      []SingleStockData
	OutOfStockForShip   []SingleStockData
	InStockForPickup    []SingleStockData
	OutOfStockForPickup []SingleStockData
}

type SingleStockData struct {
	TCIN            string
	TCINType        enums.CheckoutType
	ProductName     string
	ProductImageURL string
}

// Used in SetPaymentInfo function
type CVV struct {
	CVV string `json:"cvv"`
}

type AccountInfo struct {
	Email          string
	Password       string
	ShippingType   enums.ShippingType
	PaymentType    enums.PaymentType
	DefaultCardCVV string
	CartID         string
	StoreID        string
	Cookies        []*http.Cookie
	CartInfo       GetCartInfoResponse
	Refresh        int64
}
type CardDetails struct {
	CardName    string `json:"card_name"`
	CardNumber  string `json:"card_number"`
	CVV         string `json:"cvv"`
	ExpiryMonth string `json:"expiry_month"`
	ExpiryYear  string `json:"expiry_year"`
}
type BillingAddress struct {
	AddressLine1 string `json:"address_line1"`
	City         string `json:"city"`
	FirstName    string `json:"first_name"`
	LastName     string `json:"last_name"`
	Phone        string `json:"phone"`
	State        string `json:"state"`
	ZipCode      string `json:"zip_code"`
	Country      string `json:"country"`
}

// Used in the Login function
type LoginJWT struct {
	Sub string  `json:"sub"`
	Iss string  `json:"iss"`
	Exp int     `json:"exp"`
	Iat int     `json:"iat"`
	Jti string  `json:"jti"`
	Sky string  `json:"sky"`
	Sut string  `json:"sut"`
	Did string  `json:"did"`
	Eid string  `json:"eid"`
	Gss float64 `json:"gss"`
	Sco string  `json:"sco"`
	Cli string  `json:"cli"`
	Asl string  `json:"asl"`
	jwt.StandardClaims
}

// Used in RefreshLogin function
type ClientCredential struct {
	ClientID string `json:"client_id"`
}
type DeviceInfo struct {
	UserAgent               string `json:"user_agent"`
	Language                string `json:"language"`
	ColorDepth              string `json:"color_depth"`
	DeviceMemory            string `json:"device_memory"`
	PixelRatio              string `json:"pixel_ratio"`
	HardwareConcurrency     string `json:"hardware_concurrency"`
	Resolution              string `json:"resolution"`
	AvailableResolution     string `json:"available_resolution"`
	TimezoneOffset          string `json:"timezone_offset"`
	SessionStorage          string `json:"session_storage"`
	LocalStorage            string `json:"local_storage"`
	IndexedDb               string `json:"indexed_db"`
	AddBehavior             string `json:"add_behavior"`
	OpenDatabase            string `json:"open_database"`
	CPUClass                string `json:"cpu_class"`
	NavigatorPlatform       string `json:"navigator_platform"`
	DoNotTrack              string `json:"do_not_track"`
	RegularPlugins          string `json:"regular_plugins"`
	Adblock                 string `json:"adblock"`
	HasLiedLanguages        string `json:"has_lied_languages"`
	HasLiedResolution       string `json:"has_lied_resolution"`
	HasLiedOs               string `json:"has_lied_os"`
	HasLiedBrowser          string `json:"has_lied_browser"`
	TouchSupport            string `json:"touch_support"`
	JsFonts                 string `json:"js_fonts"`
	NavigatorVendor         string `json:"navigator_vendor"`
	NavigatorAppName        string `json:"navigator_app_name"`
	NavigatorAppCodeName    string `json:"navigator_app_code_name"`
	NavigatorAppVersion     string `json:"navigator_app_version"`
	NavigatorLanguages      string `json:"navigator_languages"`
	NavigatorCookiesEnabled string `json:"navigator_cookies_enabled"`
	NavigatorJavaEnabled    string `json:"navigator_java_enabled"`
	VisitorID               string `json:"visitor_id"`
	TealeafID               string `json:"tealeaf_id"`
	WebglVendor             string `json:"webgl_vendor"`
	BrowserName             string `json:"browser_name"`
	BrowserVersion          string `json:"browser_version"`
	CPUArchitecture         string `json:"cpu_architecture"`
	DeviceVendor            string `json:"device_vendor"`
	DeviceModel             string `json:"device_model"`
	DeviceType              string `json:"device_type"`
	EngineName              string `json:"engine_name"`
	EngineVersion           string `json:"engine_version"`
	OsName                  string `json:"os_name"`
	OsVersion               string `json:"os_version"`
}

// Used in Monitor function by StockResponse
type Data struct {
	ProductSummaries []ProductSummaries `json:"product_summaries"`
}
type Errors struct {
	Message    string      `json:"message"`
	Locations  []Locations `json:"locations"`
	Extensions Extensions  `json:"extensions"`
}
type Extensions struct {
	Classification string `json:"classification"`
}
type Locations struct {
	Line   int `json:"line"`
	Column int `json:"column"`
}
type ProductSummaries struct {
	Typename    string      `json:"__typename"`
	TCIN        string      `json:"tcin"`
	Fulfillment Fulfillment `json:"fulfillment"`
}

// Used in GetCartInfo function by GetCartInfoResponse

type ShippingOptions struct {
	AvailabilityStatus         string     `json:"availability_status"`
	LoyaltyAvailabilityStatus  string     `json:"loyalty_availability_status"`
	AvailableToPromiseQuantity float64    `json:"available_to_promise_quantity"`
	MinimumOrderQuantity       float64    `json:"minimum_order_quantity"`
	Services                   []Services `json:"services"`
}

type Addresses struct {
	ProfileAddressID string `json:"profile_address_id"`
	AddressLine1     string `json:"address_line1"`
	City             string `json:"city"`
	Country          string `json:"country"`
	FirstName        string `json:"first_name"`
	LastName         string `json:"last_name"`
	Phone            string `json:"phone,omitempty"`
	State            string `json:"state"`
	AddressType      string `json:"address_type"`
	ZipCode          string `json:"zip_code"`
	AddressID        string `json:"address_id,omitempty"`
	Email            string `json:"email,omitempty"`
	Mobile           string `json:"mobile,omitempty"`
	Selected         bool   `json:"selected,omitempty"`
	PoBoxAddress     bool   `json:"po_box_address,omitempty"`
	DefaultAddress   bool   `json:"default_address,omitempty"`
}

type CartItems struct {
	Tcin                         string              `json:"tcin"`
	Dpci                         string              `json:"dpci"`
	FreeGiftStatus               string              `json:"free_gift_status"`
	CartID                       string              `json:"cart_id"`
	TotalCartItemQuantity        int                 `json:"total_cart_item_quantity"`
	CartItemID                   string              `json:"cart_item_id"`
	CartItemType                 string              `json:"cart_item_type"`
	ReturnPolicy                 []ReturnPolicy      `json:"return_policy"`
	StoreFrontDetails            []StoreFrontDetails `json:"store_front_details"`
	Quantity                     int                 `json:"quantity"`
	UnitPrice                    float64             `json:"unit_price"`
	ListPrice                    float64             `json:"list_price"`
	CurrentPrice                 float64             `json:"current_price"`
	Msrp                         float64             `json:"msrp"`
	Eyebrow                      string              `json:"eyebrow"`
	LimitedQuantity              bool                `json:"limited_quantity"`
	AvailableQuantityInInventory int                 `json:"available_quantity_in_inventory"`
	Handling                     Handling            `json:"handling"`
	Fulfillment                  Fulfillment         `json:"fulfillment"`
	ItemAttributes               ItemAttributes      `json:"item_attributes"`
	ItemSummary                  ItemSummary         `json:"item_summary"`
	Subscription                 Subscription        `json:"subscription"`
	InventoryInfo                InventoryInfo       `json:"inventory_info"`
	ItemIndicators               ItemIndicators      `json:"item_indicators"`
	Variation                    Variation           `json:"variation"`
	CreatedAt                    time.Time           `json:"created_at"`
	UpdatedAt                    time.Time           `json:"updated_at"`
	AddOnThreshold               int                 `json:"add_on_threshold"`
	ItemAddChannel               string              `json:"item_add_channel"`
}

type Fulfillment struct {
	Type                              string            `json:"type"`
	AddressID                         string            `json:"address_id"`
	LocationData                      LocationData      `json:"location_data"`
	ShipMethod                        string            `json:"ship_method"`
	Price                             Price             `json:"price"`
	EstimatedArrivalStartDate         string            `json:"estimated_arrival_start_date"`
	EstimatedArrivalEndDate           string            `json:"estimated_arrival_end_date"`
	ShipMethodServiceLevelDescription string            `json:"ship_method_service_level_description"`
	ProductID                         string            `json:"product_id"`
	IsOutOfStockInAllStoreLocations   bool              `json:"is_out_of_stock_in_all_store_locations"`
	ShippingOptions                   ShippingOptions   `json:"shipping_options"`
	StoreOptions                      []StoreOptions    `json:"store_options"`
	ScheduledDelivery                 ScheduledDelivery `json:"scheduled_delivery"`
}
type GuestLocation struct {
	ZipCode   string `json:"zip_code"`
	State     string `json:"state"`
	Country   string `json:"country"`
	Latitude  string `json:"latitude"`
	Longitude string `json:"longitude"`
}
type GuestProfile struct {
	Phone            string        `json:"phone"`
	Email            string        `json:"email"`
	CreateDate       string        `json:"create_date"`
	FirstName        string        `json:"first_name"`
	LastName         string        `json:"last_name"`
	CartwheelID      string        `json:"cartwheel_id"`
	PreferredStoreID string        `json:"preferred_store_id"`
	Memberships      []Memberships `json:"memberships"`
}
type Handling struct{}
type Indicators struct {
	HasDigitalDeliveryItem           bool `json:"has_digital_delivery_item"`
	HasGiftOptionApplied             bool `json:"has_gift_option_applied"`
	HasRegistryItem                  bool `json:"has_registry_item"`
	HasScheduledDeliveryItem         bool `json:"has_scheduled_delivery_item"`
	HasShipToStoreItem               bool `json:"has_ship_to_store_item"`
	HasStorePickupItem               bool `json:"has_store_pickup_item"`
	HasSubscriptionItem              bool `json:"has_subscription_item"`
	HasShippingRequiredItem          bool `json:"has_shipping_required_item"`
	HasPaymentApplied                bool `json:"has_payment_applied"`
	HasPaymentSatisfied              bool `json:"has_payment_satisfied"`
	HasOnlyGiftcardItems             bool `json:"has_only_giftcard_items"`
	HasAddressAssociatedAll          bool `json:"has_address_associated_all"`
	HasOnlyGiftcardTender            bool `json:"has_only_giftcard_tender"`
	HasPaypalTenderEnabled           bool `json:"has_paypal_tender_enabled"`
	HasApplepayTenderEnabled         bool `json:"has_applepay_tender_enabled"`
	HasGiftcardTenderEnabled         bool `json:"has_giftcard_tender_enabled"`
	HasThirdpartyTenderEnabled       bool `json:"has_thirdparty_tender_enabled"`
	HasTargetTenderEnabled           bool `json:"has_target_tender_enabled"`
	HasTargetDebitCardTenderEnabled  bool `json:"has_target_debit_card_tender_enabled"`
	HasStoreFrontItem                bool `json:"has_store_front_item"`
	IsCreditCardCompareRequired      bool `json:"is_credit_card_compare_required"`
	HasEstimatedDeliveryDateModified bool `json:"has_estimated_delivery_date_modified"`
	HasCircleApplied                 bool `json:"has_circle_applied"`
	HasPickupWindowItem              bool `json:"has_pickup_window_item"`
	HasCurbsideItem                  bool `json:"has_curbside_item"`
	HasAdultBeverageItem             bool `json:"has_adult_beverage_item"`
	HasAdultBeveragePickup           bool `json:"has_adult_beverage_pickup"`
	HasAdultBeverageShipt            bool `json:"has_adult_beverage_shipt"`
	IsAddressRequiredForDigitalTax   bool `json:"is_address_required_for_digital_tax"`
	AllowsAlcohol                    bool `json:"allows_alcohol"`
}
type InventoryInfo struct {
	AvailabilityStatus               string `json:"availability_status"`
	LocationType                     string `json:"location_type"`
	DemandType                       string `json:"demand_type"`
	OnlineAvailableToPromiseQuantity int    `json:"online_available_to_promise_quantity"`
	StoresAvailableToPromiseQuantity int    `json:"stores_available_to_promise_quantity"`
}
type InStoreOnly struct {
	AvailabilityStatus string `json:"availability_status"`
}
type ItemAttributes struct {
	Description         string `json:"description"`
	ImagePath           string `json:"image_path"`
	MaxPurchaseQuantity int    `json:"max_purchase_quantity"`
	ProductSubtypeName  string `json:"product_subtype_name"`
	ProductTypeName     string `json:"product_type_name"`
}
type ItemIndicators struct {
	IsStorePickupEligible         bool `json:"is_store_pickup_eligible"`
	IsShipToStoreEligible         bool `json:"is_ship_to_store_eligible"`
	IsGiftWrapEligible            bool `json:"is_gift_wrap_eligible"`
	IsGiftMessageEligible         bool `json:"is_gift_message_eligible"`
	IsGiftOptionsEligible         bool `json:"is_gift_options_eligible"`
	IsThirdPartyGiftcard          bool `json:"is_third_party_giftcard"`
	IsShippingInOriginalContainer bool `json:"is_shipping_in_original_container"`
	IsLoyaltyPurchaseEligible     bool `json:"is_loyalty_purchase_eligible"`
	IsAddOn                       bool `json:"is_add_on"`
	IsFreeGift                    bool `json:"is_free_gift"`
	IsSubscription                bool `json:"is_subscription"`
	IsSubscriptionEligible        bool `json:"is_subscription_eligible"`
	IsComingSoon                  bool `json:"is_coming_soon"`
	IsBulky                       bool `json:"is_bulky"`
	IsMarketplace                 bool `json:"is_marketplace"`
	IsPreOrder                    bool `json:"is_pre_order"`
	IsRetailTicketed              bool `json:"is_retail_ticketed"`
	IsPickupWindowItem            bool `json:"is_pickup_window_item"`
	IsAffirmEligible              bool `json:"is_affirm_eligible"`
}
type ItemSummary struct {
	ShippingDiscount    float64 `json:"shipping_discount"`
	ShippingCharge      float64 `json:"shipping_charge"`
	TotalDiscount       float64 `json:"total_discount"`
	PartialDiscount     float64 `json:"partial_discount"`
	TotalProduct        float64 `json:"total_product"`
	TaxAmount           float64 `json:"tax_amount"`
	Price               float64 `json:"price"`
	ShippingTax         float64 `json:"shipping_tax"`
	HandlingFee         float64 `json:"handling_fee"`
	SurchargeFee        float64 `json:"surcharge_fee"`
	HandlingFeeTax      float64 `json:"handling_fee_tax"`
	HandlingFeeDiscount float64 `json:"handling_fee_discount"`
	GiftWrapFee         float64 `json:"gift_wrap_fee"`
	GiftWrapTax         float64 `json:"gift_wrap_tax"`
}
type LocationData struct {
	ZipCode string `json:"zip_code"`
	State   string `json:"state"`
}
type Memberships struct {
	Name     string `json:"name"`
	Type     string `json:"type"`
	Selected bool   `json:"selected"`
}
type OrderPickup struct {
	PickupDate         string `json:"pickup_date"`
	GuestPickSLA       int    `json:"guest_pick_sla"`
	AvailabilityStatus string `json:"availability_status"`
}
type PaymentInstructions struct {
	Amount               float64   `json:"amount"`
	CardName             string    `json:"card_name"`
	CreatedAt            time.Time `json:"created_at"`
	EncryptionSystem     string    `json:"encryption_system"`
	ExpiryMonth          string    `json:"expiry_month"`
	ExpiryYear           string    `json:"expiry_year"`
	CardNumber           string    `json:"card_number"`
	CardType             string    `json:"card_type"`
	CardSubtype          string    `json:"card_subtype"`
	Status               string    `json:"status"`
	UpdatedAt            time.Time `json:"updated_at"`
	WalletID             string    `json:"wallet_id"`
	PaymentInstructionID string    `json:"payment_instruction_id"`
	PaymentVerified      bool      `json:"payment_verified"`
	IsCvvRequired        bool      `json:"is_cvv_required"`
	IsPinPlccRequired    bool      `json:"is_pin_plcc_required"`
}
type Price struct {
	ShippingPrice float64 `json:"shipping_price"`
	CurrentRetail float64 `json:"current_retail"`
}
type RedCardBanner struct {
	OrderTotal           string `json:"order_total"`
	DiscountedOrderTotal string `json:"discounted_order_total"`
}
type ReturnPolicy struct {
	ID                int    `json:"id"`
	Type              string `json:"type"`
	DayCount          int    `json:"day_count"`
	UserType          string `json:"user_type"`
	PermanentPolicyID int    `json:"permanent_policy_id"`
	StartDate         string `json:"start_date"`
	EndDate           string `json:"end_date"`
	EndTime           string `json:"end_time"`
}
type ScheduledDelivery struct {
	AvailabilityStatus string `json:"availability_status"`
}
type ScheduledDeliveryWindows struct{}
type Services struct {
	ShippingMethodID               string    `json:"shipping_method_id"`
	MinDeliveryDate                string    `json:"min_delivery_date"`
	MaxDeliveryDate                string    `json:"max_delivery_date"`
	IsTwoDayShipping               bool      `json:"is_two_day_shipping"`
	IsBaseShippingMethod           bool      `json:"is_base_shipping_method"`
	ServiceLevelDescription        string    `json:"service_level_description"`
	ShippingMethodShortDescription string    `json:"shipping_method_short_description"`
	Cutoff                         time.Time `json:"cutoff"`
}

type StoreFrontDetails struct {
	ID         string `json:"id"`
	SellerID   string `json:"seller_id"`
	VendorName string `json:"vendor_name"`
}
type StoreOptions struct {
	LocationName            string      `json:"location_name"`
	LocationAddress         string      `json:"location_address"`
	LocationID              string      `json:"location_id"`
	SearchResponseStoreType string      `json:"search_response_store_type"`
	OrderPickup             OrderPickup `json:"order_pickup"`
	InStoreOnly             InStoreOnly `json:"in_store_only"`
}
type Subscription struct{}
type Summary struct {
	TotalDiscounts           int           `json:"total_discounts"`
	TotalGiftWrapAmount      int           `json:"total_gift_wrap_amount"`
	TotalProductAmount       float64       `json:"total_product_amount"`
	TotalShippingDiscount    float64       `json:"total_shipping_discount"`
	TotalShippingAmount      float64       `json:"total_shipping_amount"`
	TotalShippingTax         int           `json:"total_shipping_tax"`
	TotalTax                 float64       `json:"total_tax"`
	TotalSurchargeFee        int           `json:"total_surcharge_fee"`
	TotalHandlingFee         int           `json:"total_handling_fee"`
	TotalHandlingDiscount    int           `json:"total_handling_discount"`
	TotalHandlingFeeTax      int           `json:"total_handling_fee_tax"`
	GrandTotal               float64       `json:"grand_total"`
	BalanceDue               int           `json:"balance_due"`
	ItemsQuantity            int           `json:"items_quantity"`
	TotalAuthorizationAmount float64       `json:"total_authorization_amount"`
	TotalBagFee              int           `json:"total_bag_fee"`
	RedCardBanner            RedCardBanner `json:"red_card_banner"`
}
type Variants struct {
	Name  string `json:"name"`
	Value string `json:"value"`
}
type Variation struct {
	Theme    string     `json:"theme"`
	Variants []Variants `json:"variants"`
}

// Used in AddToCart function
type CartItem struct {
	TCIN          string `json:"tcin"`
	Quantity      int    `json:"quantity"`
	ItemChannelID string `json:"item_channel_id"`
}

type CartFulfillment struct {
	Type       enums.CheckoutType `json:"type"`
	LocationID string             `json:"location_id"`
	ShipMethod string             `json:"ship_method"`
}

type ProductDescription struct {
	Title string `json:"title"`
}

type Images struct {
	PrimaryImageURL string `json:"primary_image_url"`
}

type Enrichment struct {
	Images Images `json:"images"`
}

type Item struct {
	Enrichment         Enrichment         `json:"enrichment"`
	ProductDescription ProductDescription `json:"product_description"`
}
type Product struct {
	Price Price `json:"price"`
	Item  Item  `json:"item"`
}

type PriceData struct {
	Product Product `json:"product"`
}

type Orders struct {
	ChannelID              string                `json:"channel_id"`
	OrderID                string                `json:"order_id"`
	TestCart               bool                  `json:"test_cart"`
	ReferenceID            string                `json:"reference_id"`
	GuestID                string                `json:"guest_id"`
	CartState              string                `json:"cart_state"`
	CartType               string                `json:"cart_type"`
	TrackingEmail          string                `json:"tracking_email"`
	GuestType              string                `json:"guest_type"`
	GuestLocation          GuestLocation         `json:"guest_location"`
	GuestProfile           GuestProfile          `json:"guest_profile"`
	ShoppingContext        string                `json:"shopping_context"`
	ShoppingLocationID     string                `json:"shopping_location_id"`
	SubstitutionPreference string                `json:"substitution_preference"`
	Summary                Summary               `json:"summary"`
	CartItems              []CartItems           `json:"cart_items"`
	Addresses              []Addresses           `json:"addresses"`
	PaymentInstructions    []PaymentInstructions `json:"payment_instructions"`
	Indicators             Indicators            `json:"indicators"`
	ParentCartNumber       string                `json:"parent_cart_number"`
}

type UserAgent struct {
	DeviceFormFactor string `json:"device_form_factor"`
	Name             string `json:"name"`
	Network          string `json:"network"`
	Original         string `json:"original"`
}
type CartIndicators struct {
	HasShippingRequiredItem         string `json:"has_shipping_required_item"`
	HasPaymentApplied               string `json:"has_payment_applied"`
	HasPaymentSatisfied             string `json:"has_payment_satisfied"`
	HasAddressAssociatedAll         string `json:"has_address_associated_all"`
	HasPaypalTenderEnabled          string `json:"has_paypal_tender_enabled"`
	HasApplepayTenderEnabled        string `json:"has_applepay_tender_enabled"`
	HasGiftcardTenderEnabled        string `json:"has_giftcard_tender_enabled"`
	HasThirdpartyTenderEnabled      string `json:"has_thirdparty_tender_enabled"`
	HasTargetTenderEnabled          string `json:"has_target_tender_enabled"`
	HasTargetDebitCardTenderEnabled string `json:"has_target_debit_card_tender_enabled"`
}
type Addtocart struct {
}
type M struct {
	CartIndicators    CartIndicators `json:"cart_indicators"`
	Cartitemsquantity string         `json:"cartItemsQuantity"`
	ReferenceID       string         `json:"reference_id"`
	CartState         string         `json:"cart_state"`
	GuestType         string         `json:"guest_type"`
	Tealeafakasid     string         `json:"TealeafAkaSid"`
	Addtocart         Addtocart      `json:"addToCart"`
	Converted         string         `json:"converted"`
}
type User struct {
	ID string `json:"id"`
}
type Client struct {
	User User `json:"user"`
}
type Event struct {
	Action string `json:"action"`
}
type Labels struct {
	Application string `json:"application"`
	BlossomID   string `json:"blossom_id"`
	Cluster     string `json:"cluster"`
}
type Packages struct {
	BuildVersion string `json:"build_version"`
}
type URL struct {
	Domain string `json:"domain"`
	Path   string `json:"path"`
}
type Text struct {
	Num3 string `json:"3"`
	Num4 string `json:"4"`
}
type Custom struct {
	Text Text `json:"text"`
}
type Tgt struct {
	CartID string `json:"cart_id"`
	Custom Custom `json:"custom"`
}
type Metrics struct {
	E              string   `json:"e"`
	M              M        `json:"m"`
	Client         Client   `json:"client"`
	Event          Event    `json:"event"`
	Labels         Labels   `json:"labels"`
	Packages       Packages `json:"packages"`
	LogDestination string   `json:"log_destination"`
	URL            URL      `json:"url"`
	Tgt            Tgt      `json:"tgt"`
}
type Records struct {
	Appid     string    `json:"appId"`
	Useragent string    `json:"userAgent"`
	Network   string    `json:"network"`
	B         string    `json:"b"`
	D         string    `json:"d"`
	Z         string    `json:"z"`
	N         string    `json:"n"`
	V         string    `json:"v"`
	Vi        string    `json:"vi"`
	T         int64     `json:"t"`
	UserAgent UserAgent `json:"user_agent"`
	Metrics   []Metrics `json:"metrics"`
}
