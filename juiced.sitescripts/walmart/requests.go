package walmart

// AddToCartRequest is the request sent by the AddToCart function
type AddToCartRequest struct {
	OfferID               string `json:"offerId"`
	Quantity              int    `json:"quantity"`
	ShipMethodDefaultRule string `json:"shipMethodDefaultRule"`
}

//GetCartInfoRequest is the request sent by the GetCartInfo function
type GetCartInfoRequest struct {
	StoreListIds  []StoreList `json:"storeList"`
	ZipCode       string      `json:"postalCode"`
	City          string      `json:"city"`
	State         string      `json:"state"`
	IsZipLocated  bool        `json:"isZipLocated"`
	Crt           string      `json:"crt:CRT"`
	CustomerId    string      `json:"customerId:CID"`
	CustomerType  string      `json:"customerType:type"`
	AffiliateInfo string      `json:"affiliateInfo:com.wm.reflector"`
}

//SetShippingInfoRequest is the request sent by the SetShippingInfo function
type SetShippingInfoRequest struct {
	AddressLineOne     string      `json:"addressLineOne"`
	City               string      `json:"city"`
	FirstName          string      `json:"firstName"`
	LastName           string      `json:"lastName"`
	Phone              string      `json:"phone"`
	Email              string      `json:"email"`
	MarketingEmailPref bool        `json:"marketingEmailPref"`
	PostalCode         string      `json:"postalCode"`
	State              string      `json:"state"`
	CountryCode        string      `json:"countryCode"`
	ChangedFields      []string    `json:"changedFields"`
	Storelist          []Storelist `json:"storeList"`
}
type Address struct {
	Postalcode string `json:"postalCode"`
	Address1   string `json:"address1"`
	City       string `json:"city"`
	State      string `json:"state"`
	Country    string `json:"country"`
}
type Storetype struct {
	ID          int    `json:"id"`
	Name        string `json:"name"`
	Displayname string `json:"displayName"`
}
type Customerinfo struct {
	Distance      float64 `json:"distance"`
	Ispreferred   bool    `json:"isPreferred"`
	Iswithinrange bool    `json:"isWithinRange"`
}
type Storelist struct {
	ID           int          `json:"id"`
	Address      Address      `json:"address"`
	Storetype    Storetype    `json:"storeType"`
	Customerinfo Customerinfo `json:"customerInfo"`
}

//SetPaymentInfoRequest is the request sent by the SetPaymentInfo function
type SetPaymentInfoRequest struct {
	PaymentType    string `json:"paymentType"`
	CardType       string `json:"cardType"`
	FirstName      string `json:"firstName"`
	LastName       string `json:"lastName"`
	AddressLineOne string `json:"addressLineOne"`
	AddressLineTwo string `json:"addressLineTwo"`
	City           string `json:"city"`
	State          string `json:"state"`
	PostalCode     string `json:"postalCode"`
	ExpiryMonth    string `json:"expiryMonth"`
	ExpiryYear     string `json:"expiryYear"`
	Email          string `json:"email"`
	Phone          string `json:"phone"`
	EncryptedPan   string `json:"encryptedPan"`
	EncryptedCvv   string `json:"encryptedCvv"`
	IntegrityCheck string `json:"integrityCheck"`
	KeyId          string `json:"keyId"`
	Phase          string `json:"phase"`
	PiHash         string `json:"piHash"`
}

//PlaceOrderRequest is the request sent by the PlaceOrder function
type PlaceOrderRequest struct {
	CvvInSession   bool             `json:"cvvInSession"`
	VoltagePayment []VoltagePayment `json:"voltagePayments"`
}

//SubmitPaymentRequest is the request sent by the Payments function
type SubmitPaymentRequest struct {
	Payments     []SubmitPayment `json:"payments"`
	CvvInSession bool            `json:"cvvInSession"`
}
