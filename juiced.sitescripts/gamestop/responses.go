package gamestop

type AkamaiResponse struct {
	Sensordata string `json:"sensorData"`
	Pixel      string `json:"pixel"`
}

type MonitorResponse struct {
	Product   Product         `json:"product"`
	GtmData   GtmData         `json:"gtmData"`
	MccEvents [][]interface{} `json:"__mccEvents"`
}
type Availability struct {
	ButtonText string `json:"buttonText"`
}

type Product struct {
	Availability       Availability `json:"availability"`
	Available          bool         `json:"available"`
	ID                 string       `json:"id"`
	SelectedProductURL string       `json:"selectedProductUrl"`
	MaxOrderQuantity   int          `json:"maxOrderQuantity"`
}
type ProductInfo struct {
	SKU          string `json:"sku"`
	ProductID    string `json:"productID"`
	Name         string `json:"name"`
	Availability string `json:"availability"`
}
type Price struct {
	SellingPrice string `json:"sellingPrice"`
	BasePrice    string `json:"basePrice"`
	Currency     string `json:"currency"`
}
type GtmData struct {
	ProductInfo ProductInfo `json:"productInfo"`
	Price       Price       `json:"price"`
}

type LoginResponse struct {
	Action                string                `json:"action"`
	QueryString           string                `json:"queryString"`
	Locale                string                `json:"locale"`
	AuthenticatedCustomer AuthenticatedCustomer `json:"authenticatedCustomer"`
	GtmSessionData        GtmSessionData        `json:"gtmSessionData"`
	Success               bool                  `json:"success"`
	RedirectURL           string                `json:"redirectUrl"`
	LoginStatus           LoginStatus           `json:"loginStatus"`
}
type AuthenticatedCustomer struct{}
type User struct {
	UserType    string `json:"userType"`
	LoginStatus string `json:"loginStatus"`
	CustKey     string `json:"custKey"`
	MemberType  string `json:"memberType"`
}
type GtmSessionData struct {
	User  User   `json:"user"`
	Event string `json:"event"`
}
type Customer struct{}
type LoginStatus struct {
	Success                         bool        `json:"success"`
	ErrorMessage                    interface{} `json:"errorMessage"`
	NeedsToBeLoggedIn               bool        `json:"needsToBeLoggedIn"`
	NeedsLegacyPasswordReset        bool        `json:"needsLegacyPasswordReset"`
	UserExistsInSFCC                bool        `json:"userExistsInSFCC"`
	UserExistsInAuthService         bool        `json:"userExistsInAuthService"`
	AccountLocked                   bool        `json:"accountLocked"`
	OpenPlayerModal                 bool        `json:"openPlayerModal"`
	IsStoreAccountWithMembership    bool        `json:"isStoreAccountWithMembership"`
	IsStoreAccountWithoutMembership bool        `json:"isStoreAccountWithoutMembership"`
	EnablePlayerModalForm           bool        `json:"enablePlayerModalForm"`
	Customer                        Customer    `json:"customer"`
}

type AddToCartResponse struct {
	Message          string `json:"message"`
	CaptchaProtected bool   `json:"protectedCaptchaCheckEnabled"`
	QuantityTotal    int    `json:"quantityTotal"`
}

type PlaceOrderResponse struct {
	Action       string          `json:"action"`
	Querystring  string          `json:"queryString"`
	Locale       string          `json:"locale"`
	Error        bool            `json:"error"`
	Errormessage string          `json:"errorMessage"`
	Mccevents    [][]interface{} `json:"__mccEvents"`
}
