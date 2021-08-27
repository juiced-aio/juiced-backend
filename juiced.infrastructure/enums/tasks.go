package enums

type TaskStatus = string

const (
	TaskIdle   TaskStatus = "Idle"
	TaskFailed TaskStatus = "Fatal error: %s"

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

type LoginType = string

const (
	LoginTypeBROWSER  LoginType = "BROWSER"
	LoginTypeREQUESTS LoginType = "REQUESTS"
)

type ShippingType = string

const (
	ShippingTypeSAVED ShippingType = "SAVED"
	ShippingTypeNEW   ShippingType = "NEW"
)

type PaymentType = string

const (
	PaymentTypeSAVED PaymentType = "SAVED"
	PaymentTypeNEW   PaymentType = "NEW"
)

type CheckoutType = string

const (
	CheckoutTypeSHIP   CheckoutType = "SHIP"
	CheckoutTypePICKUP CheckoutType = "PICKUP"
	CheckoutTypeEITHER CheckoutType = "EITHER"
)

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
