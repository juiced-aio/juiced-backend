package enums

import "fmt"

type TaskStatus = string

const (
	TaskIdle               TaskStatus = "Idle"
	TaskStarted            TaskStatus = "Started"
	TaskFailed             TaskStatus = "Fatal error: %s"
	TaskMaxRetriesExceeded TaskStatus = "Task failed (max retries exceeded): %s"

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

	GettingCartInfo        TaskStatus = "Getting cart info"
	GettingCartInfoSuccess TaskStatus = "Got cart info"
	GettingCartInfoFailure TaskStatus = "Error getting cart info: %s"

	SettingCartInfo        TaskStatus = "Setting cart info"
	SettingCartInfoSuccess TaskStatus = "Set cart info"
	SettingCartInfoFailure TaskStatus = "Error setting cart info: %s"

	HandlingQueue        TaskStatus = "Handling queue"
	HandlingQueueSuccess TaskStatus = "Handled queue"
	HandlingQueueFailure TaskStatus = "Error handling queue: %s"

	GettingOrderInfo        TaskStatus = "Getting order info"
	GettingOrderInfoSuccess TaskStatus = "Got order info"
	GettingOrderInfoFailure TaskStatus = "Error getting order info: %s"

	SettingEmailAddress        TaskStatus = "Setting email address"
	SettingEmailAddressSuccess TaskStatus = "Set email address"
	SettingEmailAddressFailure TaskStatus = "Error setting email address: %s"

	SettingShippingInfo        TaskStatus = "Setting shipping info"
	SettingShippingInfoSuccess TaskStatus = "Set shipping info"
	SettingShippingInfoFailure TaskStatus = "Error setting shipping info: %s"

	GettingBillingInfo        TaskStatus = "Getting billing info"
	GettingBillingInfoSuccess TaskStatus = "Got billing info"
	GettingBillingInfoFailure TaskStatus = "Error getting billing info: %s"

	SettingBillingInfo        TaskStatus = "Setting billing info"
	SettingBillingInfoSuccess TaskStatus = "Set billing info"
	SettingBillingInfoFailure TaskStatus = "Error setting billing info: %s"

	SettingOrderInfo        TaskStatus = "Setting order info"
	SettingOrderInfoSuccess TaskStatus = "Set order info"
	SettingOrderInfoFailure TaskStatus = "Error setting order info: %s"

	CheckingOut        TaskStatus = "Checking out"
	CheckingOutSuccess TaskStatus = "Checked out!"
	CheckingOutFailure TaskStatus = "Error checking out: %s"
	CardDeclined       TaskStatus = "Error checking out: card declined"

	WaitingForLogin     TaskStatus = "Waiting for login cookies"
	WaitingForMonitor   TaskStatus = "Waiting for monitor"
	WaitingForCaptcha   TaskStatus = "Waiting for Captcha"
	BypassingPX         TaskStatus = "Bypassing PX"
	GettingShippingInfo TaskStatus = "Getting shipping info"
	CheckoutFailed      TaskStatus = "Checkout failed"
	CheckedOut          TaskStatus = "Checked out!"
)

const (
	UnknownError         = "unknown error (status code: %d)"
	CaptchaRequiredError = "captcha required"
	BadCaptchaTokenError = "bad captcha token"
)

type TaskEventType = string

const (
	TaskStart    TaskEventType = "TaskStart"
	TaskUpdate   TaskEventType = "TaskUpdate"
	TaskFail     TaskEventType = "TaskFail"
	TaskStop     TaskEventType = "TaskStop"
	TaskComplete TaskEventType = "TaskComplete"
)

type LoginType = string

const (
	LoginTypeBROWSER  LoginType = "BROWSER"
	LoginTypeREQUESTS LoginType = "REQUESTS"
)

type AddressType = string

const (
	AddressTypeSAVED AddressType = "SAVED"
	AddressTypeNEW   AddressType = "NEW"
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

type InvalidInputTypeError struct {
	Field    string
	ShouldBe string
}

func (e *InvalidInputTypeError) Error() string {
	return fmt.Sprintf("unsupported input type for field %s (should be %s)", e.Field, e.ShouldBe)
}

type EmptyInputFieldError struct {
	Field string
}

func (e *EmptyInputFieldError) Error() string {
	return fmt.Sprintf("input for required field %s is empty", e.Field)
}

type InvalidAddressTypeError struct {
}

func (e *InvalidAddressTypeError) Error() string {
	return "cannot use saved address in guest mode"
}
