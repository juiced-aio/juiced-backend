package walmart

// GetPXCookieResponse is returned by the GetPXCookie endpoint
type GetPXCookieResponse struct {
	SetId string
	Vid   string
	Uuid  string
}

// AddToCartResponse is returned by the AddToCart endpoint
type AddToCartResponse struct {
	Cart  Cart `json:"cart"`
	Items []struct {
		Quantity int     `json:"quantity"`
		Price    float32 `json:"price"`
		Name     string  `json:"name"`
		Seller   struct {
			Name string
			Type string
			Id   string
		} `json:"seller"`
		Assets struct {
			Primary []struct {
				Num60  string `json:"60"`
				Num100 string `json:"100"`
			} `json:"primary"`
		} `json:"assets"`
	} `json:"items"`
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
