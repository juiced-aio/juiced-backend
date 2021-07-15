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
