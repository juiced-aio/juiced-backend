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
	AccessToken string   `json:"access_token"`
	TokenType   string   `json:"token_type"`
	ExpiresIn   int64    `json:"expires_in"`
	Scope       string   `json:"scope"`
	Role        string   `json:"role"`
	Roles       []string `json:"roles"`
	FamilyName  string   `json:"familyName"`
	GivenName   string   `json:"givenName"`
}
