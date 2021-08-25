package enums

// MonitorStatus is a list of possible statuses that a Monitor can have
type MonitorStatus = string

// Idle --> WaitingForProductData --> WaitingForInStock* --> SendingProductInfoToTasks --> WaitingForOutOfStock --> WaitingForInStock --> ...
const (
	MonitorIdle               MonitorStatus = "Idle"
	SettingUpMonitor          MonitorStatus = "Setting up"
	BypassingPXMonitor        MonitorStatus = "Bypassing PX"
	WaitingForProductData     MonitorStatus = "Searching"
	ProxyBanned               MonitorStatus = "Proxy is banned"
	UnableToFindProduct       MonitorStatus = "Product not found"
	WaitingForInStock         MonitorStatus = "Out of stock"
	OutOfPriceRange           MonitorStatus = "Out of price range"
	SendingProductInfoToTasks MonitorStatus = "Sending to tasks"
	SentProductInfoToTasks    MonitorStatus = "Tasks in progress"
)

type MonitorEventType = string

const (
	MonitorStart    MonitorEventType = "MonitorStart"
	MonitorUpdate   MonitorEventType = "MonitorUpdate"
	MonitorFail     MonitorEventType = "MonitorFail"
	MonitorStop     MonitorEventType = "MonitorStop"
	MonitorComplete MonitorEventType = "MonitorComplete"
)

// TaskStatus is a list of possible statuses that a Task can have
type TaskStatus = string

// Idle --> LoggingIn* --> WaitingForMonitor --> AddingToCart --> ? --> CheckedOut
const (
	TaskIdle   TaskStatus = "Idle"
	TaskFailed TaskStatus = "FAIL: %s"

	SettingUp        TaskStatus = "Setting up task"
	SettingUpSuccess TaskStatus = "Set up task"
	SettingUpFailure TaskStatus = "Error setting up task: %s"

	LoggingIn    TaskStatus = "Logging in"
	LoginSuccess TaskStatus = "Logged in"
	LoginFailure TaskStatus = "Error logging in: %s"

	EncryptingCardInfo        TaskStatus = "Encrypting card details"
	EncryptingCardInfoSuccess TaskStatus = "Encrypted card details"
	EncryptingCardInfoFailure TaskStatus = "Error encrypting card details: %s"

	AddingToCart        TaskStatus = "Adding to cart"
	AddingToCartSuccess TaskStatus = "Added to cart"
	AddingToCartFailure TaskStatus = "Error adding to cart: %s"

	SettingEmailAddress        TaskStatus = "Setting email address"
	SettingEmailAddressSuccess TaskStatus = "Set email address"
	SettingEmailAddressFailure TaskStatus = "Error setting email address: %s"

	SettingShippingInfo        TaskStatus = "Setting shipping info"
	SettingShippingInfoSuccess TaskStatus = "Set shipping info"
	SettingShippingInfoFailure TaskStatus = "Error setting shipping info: %s"

	SettingBillingInfo        TaskStatus = "Setting billing info"
	SettingBillingInfoSuccess TaskStatus = "Set billing info"
	SettingBillingInfoFailure TaskStatus = "Error setting billing info: %s"

	CheckingOut        TaskStatus = "Checking out"
	CheckingOutSuccess TaskStatus = "Checked out!"
	CheckingOutFailure TaskStatus = "Error checking out: %s"
	CardDeclined       TaskStatus = "Error checking out: card declined"

	WaitingForLogin     TaskStatus = "Waiting for login cookies"
	WaitingForMonitor   TaskStatus = "Waiting for monitor"
	WaitingForCaptcha   TaskStatus = "Waiting for Captcha"
	BypassingPX         TaskStatus = "Bypassing PX"
	GettingCartInfo     TaskStatus = "Getting cart info"
	SettingCartInfo     TaskStatus = "Setting cart info"
	GettingShippingInfo TaskStatus = "Getting shipping info"
	GettingBillingInfo  TaskStatus = "Getting billing info"
	GettingOrderInfo    TaskStatus = "Getting order info"
	SettingOrderInfo    TaskStatus = "Setting order info"
	CheckoutFailed      TaskStatus = "Checkout failed"
	CheckedOut          TaskStatus = "Checked out!"
)

type TaskEventType = string

const (
	TaskStart    TaskEventType = "TaskStart"
	TaskUpdate   TaskEventType = "TaskUpdate"
	TaskStop     TaskEventType = "TaskStop"
	TaskFail     TaskEventType = "TaskFail"
	TaskComplete TaskEventType = "TaskComplete"
)

// Retailer is a list of possible websites that Tasks and Monitors can run for
type Retailer = string

// Choose one per Task/Monitor
const (
	Amazon        Retailer = "Amazon"
	BestBuy       Retailer = "BestBuy"
	BigCartel     Retailer = "BigCartel"
	BoxLunch      Retailer = "BoxLunch"
	Disney        Retailer = "Disney"
	GameStop      Retailer = "GameStop"
	HotTopic      Retailer = "HotTopic"
	Newegg        Retailer = "Newegg"
	PokemonCenter Retailer = "PokemonCenter"
	Shopify       Retailer = "Shopify"
	SquareSpace   Retailer = "SquareSpace"
	Target        Retailer = "Target"
	Topps         Retailer = "Topps"
	Walmart       Retailer = "Walmart"
)

type ShopifyRetailer = string

const (
	GenericShopify ShopifyRetailer = "GenericShopify"
	HotWheels      ShopifyRetailer = "HotWheels"
)

var ShopifySiteURLs = map[ShopifyRetailer]string{
	HotWheels: "https://creations.mattel.com",
}

// MonitorType is used to choose which monitoring method to use (SKU / URL / keywords)
type MonitorType = string

const (
	SKUMonitor     MonitorType = "SKU_MONITOR"
	FastSKUMonitor MonitorType = "FAST_SKU_MONITOR"
	SlowSKUMonitor MonitorType = "SLOW_SKU_MONITOR"
	URLMonitor     MonitorType = "URL_MONITOR"
	KeywordMonitor MonitorType = "KEYWORD_MONITOR"
)

// LoginType is used to choose which login method to use (headless browser / requests)
type LoginType = string

const (
	LoginTypeBROWSER  LoginType = "BROWSER"
	LoginTypeREQUESTS LoginType = "REQUESTS"
)

// ShippingType is used to choose which shipping address to use (stored in account / profile)
type ShippingType = string

const (
	ShippingTypeSAVED ShippingType = "SAVED"
	ShippingTypeNEW   ShippingType = "NEW"
)

// PaymentType is used to choose which payment info to use (stored in account / profile)
type PaymentType = string

const (
	PaymentTypeSAVED PaymentType = "SAVED"
	PaymentTypeNEW   PaymentType = "NEW"
)

// AddressType is used to choose which address info to use (stored in account / profile)
type AddressType = string

const (
	AddressTypeSAVED AddressType = "SAVED"
	AddressTypeNEW   AddressType = "NEW"
)

// CheckoutType is used to choose how the user will receive the item (shipping / pickup)
type CheckoutType = string

const (
	CheckoutTypeSHIP   CheckoutType = "SHIP"
	CheckoutTypePICKUP CheckoutType = "PICKUP"
	CheckoutTypeEITHER CheckoutType = "EITHER"
)

// TaskType is used to choose how the task will checkout (account / guest)
type TaskType = string

const (
	TaskTypeAccount TaskType = "ACCOUNT"
	TaskTypeGuest   TaskType = "GUEST"
)

type OrderStatus = string

const (
	OrderStatusSuccess  OrderStatus = "SUCCESS"
	OrderStatusDeclined OrderStatus = "DECLINED"
	OrderStatusFailed   OrderStatus = "FAILED"
)
