package walmart

// GetPXCookieResponse is returned by the GetPXCookie endpoint
type GetPXCookieResponse struct {
	SetId string
	Vid   string
	Uuid  string
}

// AddToCartResponse is returned by the AddToCart endpoint
type AddToCartResponse struct {
	Cart        Cart   `json:"cart"`
	RedirectURL string `json:"redirectUrl"`
}

// GetCartInfoResponse is returned by the GetCartInfo endpoint
type GetCartInfoResponse struct {
	CartId      string  `json:"cartId"`
	Summary     Summary `json:"summary"`
	RedirectURL string  `json:"redirectUrl"`
}

// SetPCIDResponse is returned by the SetPCID endpoint
type SetPCIDResponse struct {
	RedirectURL string `json:"redirectUrl"`
}

// SetShippingInfoResponse is returned by the SetShippingInfo endpoint
type SetShippingInfoResponse struct {
	RedirectURL string `json:"redirectUrl"`
}

type SetCreditCardResponse struct {
	PiHash      string `json:"piHash"`
	PaymentType string `json:"paymentType"`
	RedirectURL string `json:"redirectUrl"`
}

// SetPaymentInfoResponse is returned by the SetPaymentInfo endpoint
type SetPaymentInfoResponse struct {
	RedirectURL string `json:"redirectUrl"`
}

// PlaceOrderResponse is returned by the PlaceOrder endpoint
type PlaceOrderResponse struct {
	StatusCode  int    `json:"statusCode"`
	Code        string `json:"code"`
	RedirectURL string `json:"redirectUrl"`
}
