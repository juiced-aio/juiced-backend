package enums

// MonitorStatus is a list of possible statuses that a Monitor can have
type MonitorStatus = string

// Idle --> WaitingForProductData --> WaitingForInStock* --> SendingProductInfoToTasks --> WaitingForOutOfStock --> WaitingForInStock --> ...
const (
	MonitorIdle               MonitorStatus = "Idle"
	SettingUpMonitor          MonitorStatus = "Setting up"
	BypassingPXMonitor        MonitorStatus = "Bypassing PX"
	WaitingForProductData     MonitorStatus = "Searching"
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
	TaskIdle            TaskStatus = "Idle"
	LoggingIn           TaskStatus = "Logging in"
	WaitingForLogin     TaskStatus = "Waiting for login cookies"
	SettingUp           TaskStatus = "Setting up task"
	WaitingForMonitor   TaskStatus = "Waiting for monitor"
	WaitingForCaptcha   TaskStatus = "Waiting for Captcha"
	BypassingPX         TaskStatus = "Bypassing PX"
	AddingToCart        TaskStatus = "Adding to cart"
	GettingCartInfo     TaskStatus = "Getting cart info"
	SettingCartInfo     TaskStatus = "Setting cart info"
	GettingShippingInfo TaskStatus = "Getting shipping info"
	SettingShippingInfo TaskStatus = "Setting shipping info"
	EncryptingCardInfo  TaskStatus = "Encrypting card details"
	GettingBillingInfo  TaskStatus = "Getting billing info"
	SettingBillingInfo  TaskStatus = "Setting billing info"
	GettingOrderInfo    TaskStatus = "Getting order info"
	SettingOrderInfo    TaskStatus = "Setting order info"
	CheckingOut         TaskStatus = "Checking out"
	CheckedOut          TaskStatus = "Checked out!"
	CardDeclined        TaskStatus = "Card declined"
	CheckoutFailed      TaskStatus = "Checkout failed"
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
	Target      Retailer = "Target"
	Walmart     Retailer = "Walmart"
	Amazon      Retailer = "Amazon"
	BestBuy     Retailer = "BestBuy"
	Disney      Retailer = "Disney"
	GameStop    Retailer = "GameStop"
	Shopify     Retailer = "Shopify"
	BigCartel   Retailer = "BigCartel"
	SquareSpace Retailer = "SquareSpace"
	HotTopic    Retailer = "HotTopic"
	BoxLunch    Retailer = "BoxLunch"
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
