package pokemoncenter

type AddToCartResponse struct {
	Type     string `json:"type"`
	Quantity int    `json:"quantity"`
}

type SubmitPaymentResponse struct {
	Self struct {
		Uri string
	}
}

type PaymentKeyResponse struct {
	KeyId string `json:"keyId"`
}

type LoginResponse struct {
	Access_token string `json:"access_token"`
	Token_type   string `json:"token_type"`
	Expires_in   string `json:"expires_in"`
	Scope        string `json:"scope"`
	Role         string `json:"role"`
	Roles        string `json:"roles"`
	FamilyName   string `json:"familyName"`
	GivenName    string `json:"givenName"`
}
