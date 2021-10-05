package disney

// Monitor
type MonitorResponse struct {
	Product  ProductResponse  `json:"product"`
	ATCState ATCStateResponse `json:"addToCartButtonState"`
}

type ProductResponse struct {
	ID            string               `json:"id"`
	ProductName   string               `json:"productName"`
	ProductType   string               `json:"productType"`
	ProductURL    string               `json:"selectedProductUrl"`
	Images        ImagesResponse       `json:"images"`
	Price         PriceResponse        `json:"price"`
	Availability  AvailabilityResponse `json:"availability"`
	Available     bool                 `json:"available"`
	QuantityLimit int                  `json:"quantityLimit"`
	Variants      []VariantResponse    `json:"variationAttributes"`
}

type ImagesResponse struct {
	LargeImages   []ImageResponse `json:"large"`
	SmallImages   []ImageResponse `json:"small"`
	HighResImages []ImageResponse `json:"highRes"`
}

type ImageResponse struct {
	ImageURL string `json:"url"`
}

type PriceResponse struct {
	Sales SalesResponse `json:"sales"`
}

type SalesResponse struct {
	Price string `json:"decimalPrice"`
}

type AvailabilityResponse struct {
	IsPreOrder  bool `json:"isPreOrder"`
	IsBackOrder bool `json:"isBackOrder"`
}

type ATCStateResponse struct {
	IsDisabled bool `json:"isDisabled"`
	IsSoldOut  bool `json:"isSoldOut"`
}

type VariantResponse struct {
	Attribute string                 `json:"attributeId"`
	Values    []VariantValueResponse `json:"values"`
}

type VariantValueResponse struct {
	Value      string `json:"value"`
	Selectable bool   `json:"selectable"`
}

// Login
type LoginResponse struct {
	Data  Data        `json:"data"`
	Error interface{} `json:"error"`
}
type Profile struct {
	Swid string `json:"swid"`
}
type Token struct {
	AccessToken             string      `json:"access_token"`
	RefreshToken            string      `json:"refresh_token"`
	Swid                    string      `json:"swid"`
	TTL                     int         `json:"ttl"`
	RefreshTTL              int         `json:"refresh_ttl"`
	HighTrustExpiresIn      int         `json:"high_trust_expires_in"`
	InitialGrantInChainTime int64       `json:"initial_grant_in_chain_time"`
	Iat                     int64       `json:"iat"`
	Exp                     int64       `json:"exp"`
	RefreshExp              int64       `json:"refresh_exp"`
	HighTrustExp            int64       `json:"high_trust_exp"`
	Scope                   interface{} `json:"scope"`
	Sso                     interface{} `json:"sso"`
	IDToken                 interface{} `json:"id_token"`
	Authenticator           string      `json:"authenticator"`
	LoginValue              interface{} `json:"loginValue"`
	ClickbackType           interface{} `json:"clickbackType"`
	SessionTransferKey      string      `json:"sessionTransferKey"`
	Created                 string      `json:"created"`
	LastChecked             string      `json:"lastChecked"`
	Expires                 string      `json:"expires"`
	RefreshExpires          string      `json:"refresh_expires"`
	BlueCookie              interface{} `json:"blue_cookie"`
}
type Data struct {
	Profile Profile `json:"profile"`
	Token   Token   `json:"token"`
}

// AddToCart
type AddToCartResponse struct {
	Message string `json:"message"`
}

// GetCartInfo
type GetCartInfoResponse struct {
	Csrf      Csrf        `json:"csrf"`
	Items     []Items     `json:"items"`
	Shipments []Shipments `json:"shipments"`
}
type Csrf struct {
	Token string `json:"token"`
}
type Items struct {
	ShipmentUUID string `json:"shipmentUUID"`
}
type Shipments struct {
	SelectedShippingMethod string `json:"selectedShippingMethod"`
}

// EstablishAppSession
type EstablishAppSessionResponse struct {
	PaymentData PaymentData `json:"paymentData"`
	Order       Order       `json:"order"`
}

type Order struct {
	Totals Totals `json:"totals"`
}

type Totals struct {
	GrandTotalValue GrandTotalValue `json:"grandTotalValue"`
}

type GrandTotalValue struct {
	Value float64 `json:"value"`
}

type PaymentData struct {
	Config   Config `json:"config"`
	CardInfo GetCardTokenResponse
}

type Config struct {
	Client                string `json:"client"`
	Domain                string `json:"domain"`
	Lang                  string `json:"lang"`
	Oauth                 string `json:"oauth"`
	Session               string `json:"session"`
	ConversationID        string `json:"conversationId"`
	Hmac                  string `json:"hmac"`
	HmacVersion           int    `json:"hmacVersion"`
	InspireFont           bool   `json:"inspireFont"`
	AllowWideLayout       bool   `json:"allowWideLayout"`
	ShowErrorIcons        bool   `json:"showErrorIcons"`
	ShowRequiredAsterisks bool   `json:"showRequiredAsterisks"`
	CSSOverride           string `json:"cssOverride"`
	EnableCSSOverride     bool   `json:"enableCssOverride"`
	PaymentAE             string
}

// GetCardToken
type GetCardTokenResponse struct {
	Token           string `json:"token"`
	Tokenvault      string `json:"tokenVault"`
	Issuer          string `json:"issuer"`
	Masked          string `json:"masked"`
	Type            string `json:"type"`
	IsWalletDefault bool   `json:"isWalletDefault"`
	CardProductName string `json:"cardProductName"`
	IsDisneyVisa    bool   `json:"isDisneyVisa"`
}

// PlaceOrder
type PlaceOrderResponse struct {
	PaidAmount        int     `json:"paidAmount"`
	UnpaidAmount      float64 `json:"unpaidAmount"`
	PaidThisRequest   int     `json:"paidThisRequest"`
	SuggestedErrorKey string  `json:"suggestedErrorKey"`
	Errors            Errors  `json:"errors"`
}
type Errors struct{}
