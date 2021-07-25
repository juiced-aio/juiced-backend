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
	TaskIdle TaskStatus = "Idle"

	LoggingIn    TaskStatus = "Logging in"
	LoginSuccess TaskStatus = "Login success"
	LoginFailed  TaskStatus = "Login failed"

	EncryptingCardInfo        TaskStatus = "Encrypting card details"
	EncryptingCardInfoSuccess TaskStatus = "Encrypting card details success"
	EncryptingCardInfoFailure TaskStatus = "Failed to encrypt card details"
	CardDetailsMissing        TaskStatus = "Card information missing"

	AddingToCart                TaskStatus = "Adding to cart"
	AddingToCartFailure         TaskStatus = "Add cart failed retry"
	AddingToCartSuccess         TaskStatus = "Add cart success"
	AddingToCartQuantityFailure TaskStatus = "Expected quantity of 1 but found "

	SettingEmailAddress        TaskStatus = "Setting email address"
	SettingEmailAddressFailure TaskStatus = "Setting email failed"
	SettingEmailAddressSuccess TaskStatus = "Setting email address success"

	SettingShippingInfo        TaskStatus = "Setting shipping info"
	SettingShippingInfoFailure TaskStatus = "Setting shipping failed"
	SettingShippingInfoSuccess TaskStatus = "Setting shipping info success"

	SettingBillingInfo        TaskStatus = "Setting billing info"
	SettingBillingInfoFailure TaskStatus = "Setting billing info failed"
	SettingBillingInfoSuccess TaskStatus = "Setting billing info Success"

	CheckingOutFailure TaskStatus = "Checkout failed"

	GetUserFailure      TaskStatus = "Failed to get user information on checkout"
	WaitingForLogin     TaskStatus = "Waiting for login cookies"
	SettingUp           TaskStatus = "Setting up task"
	WaitingForMonitor   TaskStatus = "Waiting for monitor"
	WaitingForCaptcha   TaskStatus = "Waiting for Captcha"
	BypassingPX         TaskStatus = "Bypassing PX"
	GettingCartInfo     TaskStatus = "Getting cart info"
	SettingCartInfo     TaskStatus = "Setting cart info"
	GettingShippingInfo TaskStatus = "Getting shipping info"
	GettingBillingInfo  TaskStatus = "Getting billing info"
	GettingOrderInfo    TaskStatus = "Getting order info"
	SettingOrderInfo    TaskStatus = "Setting order info"
	CheckingOut         TaskStatus = "Checking out"
	CardDeclined        TaskStatus = "Card declined"
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
	Target        Retailer = "Target"
	Walmart       Retailer = "Walmart"
	Amazon        Retailer = "Amazon"
	BestBuy       Retailer = "BestBuy"
	Disney        Retailer = "Disney"
	GameStop      Retailer = "GameStop"
	Shopify       Retailer = "Shopify"
	BigCartel     Retailer = "BigCartel"
	SquareSpace   Retailer = "SquareSpace"
	HotTopic      Retailer = "HotTopic"
	BoxLunch      Retailer = "BoxLunch"
	PokemonCenter Retailer = "PokemonCenter"
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
