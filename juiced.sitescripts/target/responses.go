package target

// GetTCINStockResponse is returned by the GetTCINStock endpoint
type GetTCINStockResponse struct {
	Errors []Errors `json:"errors"`
	Data   Data     `json:"data"`
}

// GetTCINInfoResponse is returned by the TCINInfo endpoint
type GetTCINInfoResponse struct {
	Data PriceData `json:"data"`
}

// RefreshLoginResponse is returned by the RefreshLogin endpoint
type RefreshLoginResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	IDToken      string `json:"id_token"`
	ExpiresIn    int    `json:"expires_in"`
	TokenType    string `json:"token_type"`
}

// AddToCartResponse is returned by the AddToCart endpoint
type AddToCartResponse struct {
	Error        Error    `json:"Error"`
	CartID       string   `json:"cart_id"`
	CurrentPrice float64  `json:"current_price"`
	Alerts       []Alerts `json:"alerts"`
}

// GetCartInfoResponse is returned by the GetCartInfo endpoint
type GetCartInfoResponse struct {
	ChannelID           string                `json:"channel_id"`
	CartID              string                `json:"cart_id"`
	CartItems           []CartItems           `json:"cart_items"`
	Addresses           []Addresses           `json:"addresses"`
	PaymentInstructions []PaymentInstructions `json:"payment_instructions"`
}

type SetShippingInfoResponse struct {
	ChannelID           string                `json:"channel_id"`
	CartID              string                `json:"cart_id"`
	Addresses           []Addresses           `json:"addresses"`
	PaymentInstructions []PaymentInstructions `json:"payment_instructions"`
}

type SetPaymentInfoResponse struct {
	PaymentInstructionID     string  `json:"payment_instruction_id"`
	PaymentInstructionAmount float64 `json:"payment_instruction_amount"`
	RemainingBalance         float64 `json:"remaining_balance"`
}

type PlaceOrderResponse struct {
	Error   Error  `json:"Error"`
	Message string `json:"message"`
	Code    string `json:"code"`
}

type Error struct {
	Message string `json:"Message"`
	Detail  string `json:"Detail"`
}
