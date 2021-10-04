package disney

// Login
type LoginRequest struct {
	Loginvalue string `json:"loginValue"`
	Password   string `json:"password"`
}

// GetCardToken
type GetCardTokenRequest struct {
	Pan  string `json:"pan"`
	Type string `json:"type"`
}

// PlaceOrder
type PlaceOrderRequest struct {
	Cards           []Cards           `json:"cards"`
	FraudParameters []FraudParameters `json:"fraudParameters"`
}
type ProfileName struct{}
type ProfilePhone struct{}
type BillingAddress struct {
	Country    string `json:"country"`
	Line1      string `json:"line1"`
	Line2      string `json:"line2"`
	City       string `json:"city"`
	State      string `json:"state"`
	PostalCode string `json:"postalCode"`
}
type Cards struct {
	SecurityCode      string         `json:"securityCode"`
	CardProductName   string         `json:"cardProductName"`
	IsDisneyVisa      bool           `json:"isDisneyVisa"`
	CardToken         string         `json:"cardToken"`
	CardDisplayNumber string         `json:"cardDisplayNumber"`
	Expiration        string         `json:"expiration"`
	CardholderName    string         `json:"cardholderName"`
	ProfileName       ProfileName    `json:"profileName"`
	ProfilePhone      ProfilePhone   `json:"profilePhone"`
	BillingAddress    BillingAddress `json:"billingAddress"`
	SaveToProfile     bool           `json:"saveToProfile"`
	IsWalletDefault   bool           `json:"isWalletDefault"`
	IsWalletEdit      bool           `json:"isWalletEdit"`
	Issuer            string         `json:"issuer"`
	Type              string         `json:"type"`
	IsStoredCard      bool           `json:"isStoredCard"`
	Amount            float64        `json:"amount"`
}
type Params struct {
	EmailAddress string `json:"email_address,omitempty"`
	DayPhone     string `json:"day_phone,omitempty"`
}
type FraudParameters struct {
	Group  string `json:"group"`
	Params Params `json:"params"`
}
