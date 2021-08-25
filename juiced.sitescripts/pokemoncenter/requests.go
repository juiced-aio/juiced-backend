package pokemoncenter

type AddToCartRequest struct {
	ProductUri    string `json:"productUri"`
	Quantity      int    `json:"quantity"`
	Configuration string `json:"configuration"`
}

type SubmitAddressRequest struct {
	Shipping Address `json:"shipping"`
	Billing  Address `json:"billing"`
}

type SubmitPaymentRequest struct {
	PaymentDisplay string `json:"paymentDisplay"`
	PaymentKey     string `json:"paymentKey"`
	PaymentToken   string `json:"paymentToken"`
}

type CheckoutDetailsRequest struct {
	PurchaseFrom string `json:"purchaseForm"`
}

type EmailRequest struct {
	Email string `json:"email"`
}
