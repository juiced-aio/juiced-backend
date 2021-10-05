package shopify

type PollRequest struct {
	Query     string    `json:"query"`
	Variables Variables `json:"variables"`
}
type Variables struct {
	Token string `json:"token"`
}

type CreditIDRequest struct {
	CreditCard          CreditCard `json:"credit_card"`
	PaymentSessionScope string     `json:"payment_session_scope"`
}
type CreditCard struct {
	Number            string `json:"number"`
	Name              string `json:"name"`
	Month             int    `json:"month"`
	Year              int    `json:"year"`
	VerificationValue string `json:"verification_value"`
}
