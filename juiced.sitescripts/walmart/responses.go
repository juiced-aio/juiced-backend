package walmart

// GetPXCookieResponse is returned by the GetPXCookie endpoint
type GetPXCookieResponse struct {
	SetId string
	Vid   string
	Uuid  string
}

// AddToCartResponse is returned by the AddToCart endpoint
type AddToCartResponse struct {
	Cart Cart `json:"cart"`
}

// GetCartInfoResponse is returned by the GetCartInfo endpoint
type GetCartInfoResponse struct {
	CartId  string  `json:"cartId"`
	Summary Summary `json:"summary"`
}

// PlaceOrderResponse is returned by the PlaceOrder endpoint
type PlaceOrderResponse struct {
	StatusCode string `json:"statusCode"`
	Code       string `json:"code"`
}
