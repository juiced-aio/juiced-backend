package pokemoncenter

type GetCartInfoResponse struct {
	Lineitems []Lineitems `json:"_lineitems"`
}

type Lineitems struct {
	Element []Element `json:"_element"`
}

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

type CheckoutResponse struct {
	Messages []Messages    `json:"messages"`
	Links    []interface{} `json:"links"`
}
type Data struct {
	Cause string `json:"cause"`
}
type Messages struct {
	Type         string `json:"type"`
	ID           string `json:"id"`
	DebugMessage string `json:"debug-message"`
	Data         Data   `json:"data"`
}
