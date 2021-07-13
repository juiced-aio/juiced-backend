package target

// GetTCINStockRequest is sent by the GetTCINStock function
type GetTCINStockRequest struct {
	Key                      string
	TCINs                    string
	StoreID                  string
	ScheduledDeliveryStoreID string
}

// RefreshLoginRequest is sent by the RefreshLogin function
type RefreshLoginRequest struct {
	GrantType        string           `json:"grant_type"`
	ClientCredential ClientCredential `json:"client_credential"`
	DeviceInfo       DeviceInfo       `json:"device_info"`
}

// AddToCartShipRequest is sent by the AddToCart (Ship) function
type AddToCartShipRequest struct {
	CartType        string   `json:"cart_type"`
	ChannelID       string   `json:"channel_id"`
	ShoppingContext string   `json:"shopping_context"`
	CartItem        CartItem `json:"cart_item"`
}

// AddToCartPickupRequest is sent by the AddToCart (Pickup) function
type AddToCartPickupRequest struct {
	CartType        string          `json:"cart_type"`
	ChannelID       string          `json:"channel_id"`
	ShoppingContext string          `json:"shopping_context"`
	CartItem        CartItem        `json:"cart_item"`
	Fulfillment     CartFulfillment `json:"fulfillment"`
}

// GetCartInfoRequest is sent by the GetCart function
type GetCartInfoRequest struct {
	CartType string `json:"cart_type"`
}

// SetShippingInfoRequest is sent by the SetShippingInfo function
type SetShippingInfoRequest struct {
	CartType         string                 `json:"cart_type"`
	Address          SetShippingInfoAddress `json:"address"`
	Selected         bool                   `json:"selected"`
	SaveToProfile    bool                   `json:"save_to_profile"`
	SkipVerification bool                   `json:"skip_verification"`
}
type SetShippingInfoAddress struct {
	AddressLine1  string `json:"address_line1"`
	AddressLine2  string `json:"address_line2"`
	AddressType   string `json:"address_type"`
	City          string `json:"city"`
	Country       string `json:"country"`
	FirstName     string `json:"first_name"`
	LastName      string `json:"last_name"`
	Mobile        string `json:"mobile"`
	SaveAsDefault bool   `json:"save_as_default"`
	State         string `json:"state"`
	ZipCode       string `json:"zip_code"`
}

// SetPaymentInfoSavedRequest is sent by the SetPaymentInfo (Saved) function
type SetPaymentInfoSavedRequest struct {
	CartID      string `json:"cart_id"`
	WalletMode  string `json:"wallet_mode"`
	PaymentType string `json:"payment_type"`
	CardDetails CVV    `json:"card_details"`
}

// SetPaymentInfoNewRequest is sent by the SetPaymentInfo (New) function
type SetPaymentInfoNewRequest struct {
	CartID         string         `json:"cart_id"`
	WalletMode     string         `json:"wallet_mode"`
	PaymentType    string         `json:"payment_type"`
	CardDetails    CardDetails    `json:"card_details"`
	BillingAddress BillingAddress `json:"billing_address"`
}

// PlaceOrderRequest is sent by the PlaceOrder function
type PlaceOrderRequest struct {
	CartType  string `json:"cart_type"`
	ChannelID int    `json:"channel_id"`
}

type TargetCancelMethodRequest struct {
	Records []Records `json:"records"`
}
