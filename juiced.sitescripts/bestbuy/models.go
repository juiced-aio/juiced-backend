package bestbuy

import (
	"net/http"
	"net/url"
	"time"

	"backend.juicedbot.io/m/v2/juiced.infrastructure/common/entities"
	"backend.juicedbot.io/m/v2/juiced.infrastructure/common/enums"
	"backend.juicedbot.io/m/v2/juiced.infrastructure/common/events"
	"backend.juicedbot.io/m/v2/juiced.sitescripts/base"
)

var (
	BaseEndpoint      = "https://www.bestbuy.com"
	GenEndpoint       = "https://notanapi.com/api/akamai/get/sensor-data"
	AkamaiEndpoint    = "https://www.bestbuy.com/Z43Qo-szvQDrezPFUWbI-oosQsM/9YOhShXz9OX1/D3ZjQkgC/EWdSfC5P/DlY"
	LoginPageEndpoint = "https://www.bestbuy.com/identity/global/signin"
	LoginEndpoint     = "https://www.bestbuy.com/identity/authenticate"
	//https://www.bestbuy.com/api/3.0/priceBlocks?skus=6439402%2c6439000
	MonitorEndpoint = "https://www.bestbuy.com/api/3.0/priceBlocks?skus=%v"
	//{"zipCode":null,"destinationZipCode":null,"showInStore":false,"showOnShelf":false,"additionalStores":null,"items":[{"sku": "6439299"}],"lookupInStoreQuantity":false,"consolidated":false,"locationId":null,"xboxAllAccess":false,"showOnlyOnShelf":false,"onlyBestBuyLocations":false}
	AddToCartEndpoint      = "https://www.bestbuy.com/cart/api/v1/addToCart"
	CartInfoEndpoint       = "https://www.bestbuy.com/cart/api/v1/fulfillment/ispu"
	CheckoutEndpoint       = "https://www.bestbuy.com/checkout/r/fast-track"
	BaseShippingEndpoint   = "https://www.bestbuy.com/checkout/r/fulfillment"
	OrderEndpoint          = "https://www.bestbuy.com/checkout/orders/%s"
	BasePaymentEndpoint    = "https://www.bestbuy.com/checkout/r/payment"
	PaymentEndpoint        = "https://www.bestbuy.com/payment/api/v1/payment/%s/creditCard"
	RefreshPaymentEndpoint = "https://www.bestbuy.com/checkout/orders/%s/paymentMethods/refreshPayment"
	PrelookupEndpoint      = "https://www.bestbuy.com/payment/api/v1/payment/%s/threeDSecure/preLookup"
	PlaceOrderEndpoint     = "https://www.bestbuy.com/checkout/api/1.0/paysecure/submitCardAuthentication"
)

var ParsedBase, _ = url.Parse(BaseEndpoint)

type AddHeadersFunction func(*http.Request, ...string)

// Monitor info
type Monitor struct {
	Monitor         base.Monitor
	SKUsSentToTask  []string
	RunningMonitors []string
	OutOfStockSKUs  []string
	SKUs            []string
	EventInfo       events.BestbuySingleStockData
	SKUWithInfo     map[string]entities.BestbuySingleMonitorInfo
}

type OrderStatus string

var (
	OrderStatusSuccess  OrderStatus = "SUCCESS"
	OrderStatusDeclined OrderStatus = "DECLINED"
	OrderStatusFailed   OrderStatus = "FAILED"
)

var DefaultRawHeaders = [][2]string{
	{"pragma", "no-cache"},
	{"cache-control", "no-cache"},
	{"sec-ch-ua", "\" Not A;Brand\";v=\"99\", \"Chromium\";v=\"90\", \"Google Chrome\";v=\"90\""},
	{"sec-ch-ua-mobile", "?0"},
	{"upgrade-insecure-requests", "1"},
	{"user-agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/90.0.4430.212 Safari/537.36"},
	{"accept", "text/html,application/xhtml+xml,application/xml;q=0.9,image/avif,image/webp,image/apng,*/*;q=0.8,application/signed-exchange;v=b3;q=0.9"},
	{"sec-fetch-site", "none"},
	{"sec-fetch-mode", "navigate"},
	{"sec-fetch-user", "?1"},
	{"sec-fetch-dest", "document"},
	{"accept-encoding", "gzip, deflate, br"},
	{"accept-language", "en-US,en;q=0.9"},
}

// Request parameters
type Request struct {
	Client             http.Client
	Method             string
	URL                string
	Headers            http.Header
	RawHeaders         [][2]string
	AddHeadersFunction AddHeadersFunction
	Referer            string
	Data               []byte
	RequestBodyStruct  interface{}
	ResponseBodyStruct interface{}
	Response           *http.Response
	RandOpt            string
}

type Task struct {
	Task         base.Task
	TaskType     enums.TaskType
	CheckoutInfo CheckoutInfo
	AccountInfo  AccountInfo
}

type CheckoutInfo struct {
	SKUInStock string
	ID         string
	ItemID     string
	PaymentID  string
	OrderID    string
	ThreeDsID  string
	ImageUrl   string
	Price      int
	ItemName   string
}

type AccountInfo struct {
	Email          string
	Password       string
	PaymentType    string
	DefaultCardCVV string
	CartID         string
	StoreID        string
	Cookies        []*http.Cookie
	Refresh        int64
}

type GenResponse struct {
	Sensordata string `json:"sensorData"`
	Pixel      string `json:"pixel"`
}

type SensorRequest struct {
	SensorData string `json:"sensor_data"`
}

type SigninJson struct {
	GoogleStepup                  bool        `json:"google.stepUp"`
	Mt                            interface{} `json:"mt"`
	Lastnamefieldname             string      `json:"lastNameFieldName"`
	AppleOauth2Flowredirecturi    string      `json:"apple.OAuth2FlowRedirectUri"`
	Verificationcodefieldname     string      `json:"verificationCodeFieldName"`
	Challengetype                 interface{} `json:"challengeType"`
	AppleAppid                    string      `json:"apple.appId"`
	Usepinplugin                  bool        `json:"usePinPlugin"`
	Removeheaderlinks             bool        `json:"removeHeaderLinks"`
	Enabledevicegrid              bool        `json:"enableDeviceGrid"`
	Enabletrustme                 bool        `json:"enableTrustMe"`
	Redirecturl                   interface{} `json:"redirectUrl"`
	Allowsguestflow               bool        `json:"allowsGuestFlow"`
	AppleState                    string      `json:"apple.state"`
	Reenterhash                   string      `json:"reenterHash"`
	GoogleAPITimeout              int         `json:"google.api.timeout"`
	Hidenotyou                    bool        `json:"hideNotYou"`
	AppleAutosignon               bool        `json:"apple.autoSignOn"`
	Contextid                     string      `json:"contextId"`
	Guesturl                      string      `json:"guestUrl"`
	Csitoken                      string      `json:"csiToken"`
	AppleUseoauth2Flow            bool        `json:"apple.useOAuth2Flow"`
	Smsenabled                    bool        `json:"smsEnabled"`
	Phonefieldname                string      `json:"phoneFieldName"`
	GoogleAutosignon              bool        `json:"google.autoSignOn"`
	Socialuseridfieldname         string      `json:"socialUserIdFieldName"`
	Allowscreateaccount           bool        `json:"allowsCreateAccount"`
	Urlredirecttimeout            int         `json:"urlRedirectTimeout"`
	Emailfieldname                string      `json:"emailFieldName"`
	Token                         string      `json:"token"`
	Analyticsreadyredirectwait    bool        `json:"analyticsReadyRedirectWait"`
	Analyticstimeout              int         `json:"analyticsTimeout"`
	Flowoptions                   string      `json:"flowOptions"`
	Socialaccesstokenfieldnames   []string    `json:"socialAccessTokenFieldNames"`
	Email                         interface{} `json:"email"`
	Enablecredentialsapi          bool        `json:"enableCredentialsApi"`
	Supportsmagiclinkonsigninpage bool        `json:"supportsMagicLinkOnSignInPage"`
	GoogleShowbutton              bool        `json:"google.showButton"`
	Enablebiometricauthentication bool        `json:"enableBiometricAuthentication"`
	Codelist                      []string    `json:"codeList"`
	Keyclassa                     string      `json:"keyClassA"`
	Alpha                         []string    `json:"alpha"`
	AppleScope                    string      `json:"apple.scope"`
	Zplankid                      string      `json:"zplankId"`
	AppleShowbutton               bool        `json:"apple.showButton"`
	GoogleScope                   string      `json:"google.scope"`
	Salmon                        string      `json:"Salmon"`
	GoogleAppid                   string      `json:"google.appId"`
	AppleAPITimeout               int         `json:"apple.api.timeout"`
	Allowemployeedelink           bool        `json:"allowEmployeeDelink"`
	Firstnamefieldname            string      `json:"firstNameFieldName"`
	Enablerememberme              bool        `json:"enableRememberMe"`
	Headers                       interface{} `json:"headers"`
	Showpopupmodalmessage         bool        `json:"showPopUpModalMessage"`
	Enablewebotp                  bool        `json:"enableWebOtp"`
	Hash                          string      `json:"hash"`
	Enablexgridheader             bool        `json:"enableXGridHeader"`
	Keyclassb                     string      `json:"keyClassB"`
	Passwordallowedlength         int         `json:"passwordAllowedLength"`
	Returnurl                     string      `json:"returnUrl"`
	Brand                         Brand       `json:"brand"`
	Disableheader                 bool        `json:"disableHeader"`
	Disablefooter                 bool        `json:"disableFooter"`
	Clientlogpercentages          []float64   `json:"clientLogPercentages"`
	Enabledynamictmx              bool        `json:"enableDynamicTmx"`
	Useensighten                  bool        `json:"useEnsighten"`
}
type MetalayerCreateaccount struct {
}
type MetalayerResetpassword struct {
}
type MetalayerSignin struct {
}
type Forgotpassword struct {
	Instructions string `json:"instructions"`
}
type Confirmsignon struct {
	Prompt           string `json:"prompt"`
	Valueproposition string `json:"valueProposition"`
	Buttoncopy       string `json:"buttonCopy"`
}
type Google struct {
	Prompt           string `json:"prompt"`
	Valueproposition string `json:"valueProposition"`
}
type Resetpassword struct {
	Prompt           string `json:"prompt"`
	Valueproposition string `json:"valueProposition"`
	Google           Google `json:"google"`
}
type Pin struct {
	Createprompt           string `json:"createPrompt"`
	Createvalueproposition string `json:"createValueProposition"`
	Usethemeheader         bool   `json:"useThemeHeader"`
	Usevalueproposition    string `json:"useValueProposition"`
	Disclaimertext         string `json:"disclaimerText"`
}
type Brand struct {
	Brand                                  string                 `json:"brand"`
	Createaccountsecondaryvalueproposition string                 `json:"createAccountSecondaryValueProposition"`
	Guestbuttoncopy                        string                 `json:"guestButtonCopy"`
	Guestcopyabovethebutton                string                 `json:"guestCopyAboveTheButton"`
	Guestcopybelowthebutton                string                 `json:"guestCopyBelowTheButton"`
	Guestprompt                            string                 `json:"guestPrompt"`
	MetalayerCreateaccount                 MetalayerCreateaccount `json:"metaLayer.createAccount"`
	MetalayerResetpassword                 MetalayerResetpassword `json:"metaLayer.resetPassword"`
	MetalayerSignin                        MetalayerSignin        `json:"metaLayer.signin"`
	Pagetitle                              string                 `json:"pageTitle"`
	Signinprompt                           string                 `json:"signinPrompt"`
	Signinvalueproposition                 string                 `json:"signinValueProposition"`
	Theme                                  string                 `json:"theme"`
	Tracer                                 string                 `json:"tracer"`
	Useparametersubstitution               bool                   `json:"useParameterSubstitution"`
	Createaccountpagetitle                 string                 `json:"createAccountPageTitle"`
	Createaccountprompt                    string                 `json:"createAccountPrompt"`
	Signinsubmitbuttoncopy                 string                 `json:"signinSubmitButtonCopy"`
	Createaccountsubmitbuttoncopy          string                 `json:"createAccountSubmitButtonCopy"`
	Skipcreateaccountbuttoncopy            string                 `json:"skipCreateAccountButtonCopy"`
	Disclaimertext                         string                 `json:"disclaimerText"`
	Signinlayout                           string                 `json:"signInLayout"`
	Createaccountcopyabovethebutton        string                 `json:"createAccountCopyAboveTheButton"`
	Createaccountcopybelowthebutton        string                 `json:"createAccountCopyBelowTheButton"`
	Createaccountbuttoncopy                string                 `json:"createAccountButtonCopy"`
	Returnurlcopy                          string                 `json:"returnUrlCopy"`
	Verifyownershippagetitle               string                 `json:"verifyOwnershipPageTitle"`
	Verifyownershipvalueproposition        string                 `json:"verifyOwnershipValueProposition"`
	Forgotpassword                         Forgotpassword         `json:"forgotPassword"`
	Confirmsignon                          Confirmsignon          `json:"confirmSignOn"`
	Resetpassword                          Resetpassword          `json:"resetPassword"`
	Pin                                    Pin                    `json:"pin"`
}

type UserActivity struct {
	Mousemoved         bool   `json:"mouseMoved"`
	Keyboardused       bool   `json:"keyboardUsed"`
	Fieldreceivedinput bool   `json:"fieldReceivedInput"`
	Fieldreceivedfocus bool   `json:"fieldReceivedFocus"`
	Timestamp          string `json:"timestamp"`
	Email              string `json:"email"`
}

type LoginResponse struct {
	Status         string        `json:"status"`
	Token          string        `json:"token"`
	Shouldredirect bool          `json:"shouldRedirect"`
	Redirecturl    string        `json:"redirectUrl"`
	Flowoptions    string        `json:"flowOptions"`
	Missingfields  []interface{} `json:"missingFields"`
}

type MonitorRequest struct {
	Zipcode               interface{} `json:"zipCode"`
	Destinationzipcode    interface{} `json:"destinationZipCode"`
	Showinstore           bool        `json:"showInStore"`
	Showonshelf           bool        `json:"showOnShelf"`
	Additionalstores      interface{} `json:"additionalStores"`
	Items                 []Items     `json:"items"`
	Lookupinstorequantity bool        `json:"lookupInStoreQuantity"`
	Consolidated          bool        `json:"consolidated"`
	Locationid            interface{} `json:"locationId"`
	Xboxallaccess         bool        `json:"xboxAllAccess"`
	Showonlyonshelf       bool        `json:"showOnlyOnShelf"`
	Onlybestbuylocations  bool        `json:"onlyBestBuyLocations"`
}
type Items struct {
	Sku   string `json:"sku"`
	Skuid string `json:"skuId"`
}

type MonitorResponse []struct {
	Sku Sku `json:"sku"`
}
type Attributes struct {
	Lowpriceguaranteedproduct bool `json:"lowPriceGuaranteedProduct"`
}

type Buttonstate struct {
	Buttonstate string `json:"buttonState"`
	Displaytext string `json:"displayText"`
	Skuid       string `json:"skuId"`
}
type Names struct {
	Short string `json:"short"`
}
type Pricedomain struct {
	Skuid               string  `json:"skuId"`
	Priceeventtype      string  `json:"priceEventType"`
	Regularprice        float64 `json:"regularPrice"`
	Currentprice        float64 `json:"currentPrice"`
	Customerprice       float64 `json:"customerPrice"`
	Totalsavings        float64 `json:"totalSavings"`
	Totalsavingspercent float64 `json:"totalSavingsPercent"`
	Ismap               bool    `json:"isMAP"`
	Currentasofdate     string  `json:"currentAsOfDate"`
}
type Price struct {
	Currentprice       float64     `json:"currentPrice"`
	Pricingtype        string      `json:"pricingType"`
	Smartpricerenabled bool        `json:"smartPricerEnabled"`
	Pricedomain        Pricedomain `json:"priceDomain"`
}
type Class struct {
	Displayname string `json:"displayName"`
	ID          string `json:"id"`
}
type Department struct {
	Displayname string `json:"displayName"`
	ID          string `json:"id"`
}
type Subclass struct {
	Displayname string `json:"displayName"`
	ID          string `json:"id"`
}
type Sku struct {
	Attributes       Attributes    `json:"attributes"`
	Brand            Brand         `json:"brand"`
	Buttonstate      Buttonstate   `json:"buttonState"`
	Condition        string        `json:"condition"`
	Names            Names         `json:"names"`
	Price            Price         `json:"price"`
	Producttype      string        `json:"productType"`
	Skuid            string        `json:"skuId"`
	URL              string        `json:"url"`
	Class            Class         `json:"class"`
	Department       Department    `json:"department"`
	Subclass         Subclass      `json:"subclass"`
	Inksubscriptions []interface{} `json:"inkSubscriptions"`
}

type AddToCartRequest struct {
	Items []Items `json:"items"`
}

type AddToCartResponse struct {
	Cartcount    int            `json:"cartCount"`
	Cartsubtotal float64        `json:"cartSubTotal"`
	Summaryitems []Summaryitems `json:"summaryItems"`
	Errorsummary Errorsummary   `json:"errorSummary"`
}
type Financeoption struct {
	Defaultplan                bool    `json:"defaultPlan"`
	Offerid                    string  `json:"offerId"`
	Financecodename            string  `json:"financeCodeName"`
	Financecode                int     `json:"financeCode"`
	Financeterm                int     `json:"financeTerm"`
	Rank                       int     `json:"rank"`
	Monthlypayment             float64 `json:"monthlyPayment"`
	Monthlypaymentincludingtax float64 `json:"monthlyPaymentIncludingTax"`
	Priority                   int     `json:"priority"`
	Plantype                   string  `json:"planType"`
	Rate                       int     `json:"rate"`
	Totalcost                  float64 `json:"totalCost"`
	Termsandconditions         string  `json:"termsAndConditions"`
	Financecodedesclong        string  `json:"financeCodeDescLong"`
}
type Pricedetails struct {
	Priceeventtype              string        `json:"priceEventType"`
	Regularpricemessagetype     interface{}   `json:"regularPriceMessageType"`
	Saleeventmessagetype        interface{}   `json:"saleEventMessageType"`
	Totalcurrentprice           float64       `json:"totalCurrentPrice"`
	Totalcustomerprice          float64       `json:"totalCustomerPrice"`
	Totalregularprice           float64       `json:"totalRegularPrice"`
	Totalinstantsavings         interface{}   `json:"totalInstantSavings"`
	Totalsavings                interface{}   `json:"totalSavings"`
	Unitcurrentprice            float64       `json:"unitCurrentPrice"`
	Unitregularprice            float64       `json:"unitRegularPrice"`
	Unitinstantsavings          interface{}   `json:"unitInstantSavings"`
	Unitopenboxsavings          interface{}   `json:"unitOpenBoxSavings"`
	Totalopenboxsavings         interface{}   `json:"totalOpenBoxSavings"`
	Totalopenboxprice           interface{}   `json:"totalOpenBoxPrice"`
	Unitopenboxprice            interface{}   `json:"unitOpenBoxPrice"`
	Contractregularprice        interface{}   `json:"contractRegularPrice"`
	Contractcurrentprice        interface{}   `json:"contractCurrentPrice"`
	Contractinstantsavings      interface{}   `json:"contractInstantSavings"`
	Nocontractregularprice      interface{}   `json:"noContractRegularPrice"`
	Nocontractcurrentprice      interface{}   `json:"noContractCurrentPrice"`
	Totalemployeeprice          interface{}   `json:"totalEmployeePrice"`
	Contracttotalsavingspercent interface{}   `json:"contractTotalSavingsPercent"`
	Bbydownpayment              interface{}   `json:"bbyDownPayment"`
	Billcreditamount            interface{}   `json:"billCreditAmount"`
	Gspunitprice                float64       `json:"gspUnitPrice"`
	Listprice                   float64       `json:"listPrice"`
	Icrcode                     interface{}   `json:"icrCode"`
	Offertargets                interface{}   `json:"offerTargets"`
	Offerqualifiers             interface{}   `json:"offerQualifiers"`
	Discountsapplied            interface{}   `json:"discountsApplied"`
	Isrequiredwithoffer         bool          `json:"isRequiredWithOffer"`
	Recurringchargeamount       interface{}   `json:"recurringChargeAmount"`
	Recurringchargeterm         interface{}   `json:"recurringChargeTerm"`
	Financeoption               Financeoption `json:"financeOption"`
	Preferredbadging            interface{}   `json:"preferredBadging"`
	Dealexpirationtimestamp     interface{}   `json:"dealExpirationTimeStamp"`
	Copunitprice                interface{}   `json:"copunitPrice"`
}
type Summaryitems struct {
	Skuid                string       `json:"skuId"`
	Isicr                bool         `json:"isICR"`
	Ishaccs              bool         `json:"isHaccs"`
	Ispaidmemberdiscount bool         `json:"isPaidMemberDiscount"`
	Lineid               string       `json:"lineId"`
	Quantity             int          `json:"quantity"`
	Price                float64      `json:"price"`
	Pricedetails         Pricedetails `json:"priceDetails"`
}

type Errorsummary struct {
	Errorcode string `json:"errorCode"`
	Message   string `json:"message"`
}

type OrderData struct {
	ID       string `json:"id"`
	Revision int    `json:"revision"`
	Meta     struct {
		Price struct {
			Producttotal                       string `json:"productTotal"`
			Cardmemberdiscount                 string `json:"cardMemberDiscount"`
			Totalsavings                       string `json:"totalSavings"`
			Totalpaidmemberdiscounts           string `json:"totalPaidMemberDiscounts"`
			Totalnonmemberdiscounts            string `json:"totalNonMemberDiscounts"`
			Rewardcert                         string `json:"rewardCert"`
			Promocodes                         string `json:"promoCodes"`
			Tax                                string `json:"tax"`
			Subtotal                           string `json:"subtotal"`
			Promocodesaftertax                 string `json:"promoCodesAfterTax"`
			Giftcards                          string `json:"giftCards"`
			Total                              string `json:"total"`
			Istaxestimate                      bool   `json:"isTaxEstimate"`
			Istaxcalculated                    bool   `json:"isTaxCalculated"`
			Fulfillmentincentiveamount         string `json:"fulfillmentIncentiveAmount"`
			Eligiblefulfillmentincentiveamount string `json:"eligibleFulfillmentIncentiveAmount"`
			Customerofferamount                string `json:"customerOfferAmount"`
			Leasedproducttotal                 string `json:"leasedProductTotal"`
			Nonleasedproducttotal              string `json:"nonLeasedProductTotal"`
			Nonleasedordertotal                string `json:"nonLeasedOrderTotal"`
		} `json:"price"`
		Hasmobilepackage                 bool `json:"hasMobilePackage"`
		Haswishlistitems                 bool `json:"hasWishListItems"`
		Istspeligible                    bool `json:"isTSPEligible"`
		Skipscheduling                   bool `json:"skipScheduling"`
		Isemployeediscountpolicyaccepted bool `json:"isEmployeeDiscountPolicyAccepted"`
		Istotaltechsupportapplied        bool `json:"isTotalTechSupportApplied"`
		Hasmonthlygspitem                bool `json:"hasMonthlyGspItem"`
		Primarypaymentrequired           bool `json:"primaryPaymentRequired"`
		Hasageverificationsku            bool `json:"hasAgeVerificationSku"`
		Hassmsnotifyerror                bool `json:"hasSmsNotifyError"`
		Mybestbuyaccount                 struct {
			Points            int    `json:"points"`
			Dollarsequivalent string `json:"dollarsEquivalent"`
		} `json:"myBestBuyAccount"`
		Ispwpeligible        bool     `json:"isPwpEligible"`
		Pwpeligiblecardtypes []string `json:"pwpEligibleCardTypes"`
		Fulfillments         struct {
			Shipping struct {
				Iseligibleforsamedayscheduling bool `json:"isEligibleForSameDayScheduling"`
			} `json:"shipping"`
			Ispu struct {
				Allitems struct {
					Availability bool `json:"availability"`
				} `json:"allItems"`
			} `json:"ispu"`
		} `json:"fulfillments"`
		Applecareplusdisclaimer            string `json:"applecarePlusDisclaimer"`
		Applecareplustheftdisclaimer       string `json:"applecarePlusTheftDisclaimer"`
		Applecareplustheftyearlydisclaimer string `json:"applecarePlusTheftYearlyDisclaimer"`
	} `json:"meta"`
	Items []struct {
		ID   string `json:"id"`
		Meta struct {
			Shortlabel            string   `json:"shortLabel"`
			Maxquantity           int      `json:"maxQuantity"`
			Availablefulfillments []string `json:"availableFulfillments"`
			Fulfillments          struct {
				Shipping struct {
					Availablelevelsofservice []struct {
						ID                 string `json:"id"`
						Level              string `json:"level"`
						Shipsbymin         string `json:"shipsByMin"`
						Shipsbymax         string `json:"shipsByMax"`
						Price              string `json:"price"`
						Signaturerequired  bool   `json:"signatureRequired"`
						Cutofftime         string `json:"cutOffTime"`
						Guaranteeddelivery bool   `json:"guaranteedDelivery"`
						Name               string `json:"name"`
						Dateindicator      string `json:"dateIndicator,omitempty"`
						Displaydatetype    string `json:"displayDateType"`
						Customerlosgroupid string `json:"customerLosGroupId"`
					} `json:"availableLevelsOfService"`
					Availableshiptolocationtypes []string `json:"availableShipToLocationTypes"`
				} `json:"shipping"`
				Instorepickup struct {
					Agentpickup bool `json:"agentPickup"`
					Store       struct {
						ID      string `json:"id"`
						Name    string `json:"name"`
						Address struct {
							Firstname      string `json:"firstName"`
							Lastname       string `json:"lastName"`
							Street         string `json:"street"`
							City           string `json:"city"`
							State          string `json:"state"`
							Zipcode        string `json:"zipcode"`
							Dayphonenumber string `json:"dayPhoneNumber"`
						} `json:"address"`
						Pickupdate          string   `json:"pickupDate"`
						Holduntildate       string   `json:"holdUntilDate"`
						Displaydatetype     string   `json:"displayDateType"`
						Constrainedquantity bool     `json:"constrainedQuantity"`
						Distance            float64  `json:"distance"`
						Pickuptypes         []string `json:"pickupTypes"`
						Iscurbsideavailable bool     `json:"isCurbsideAvailable"`
						Hour                struct {
							Date  string `json:"date"`
							Open  string `json:"open"`
							Close string `json:"close"`
						} `json:"hour"`
					} `json:"store"`
					Nearbystore struct {
						ID      string `json:"id"`
						Name    string `json:"name"`
						Type    string `json:"type"`
						Address struct {
							Firstname      string `json:"firstName"`
							Lastname       string `json:"lastName"`
							Street         string `json:"street"`
							City           string `json:"city"`
							State          string `json:"state"`
							Zipcode        string `json:"zipcode"`
							Dayphonenumber string `json:"dayPhoneNumber"`
						} `json:"address"`
						Pickupdate          string   `json:"pickupDate"`
						Holduntildate       string   `json:"holdUntilDate"`
						Displaydatetype     string   `json:"displayDateType"`
						Constrainedquantity bool     `json:"constrainedQuantity"`
						Distance            float64  `json:"distance"`
						Pickuptypes         []string `json:"pickupTypes"`
						Iscurbsideavailable bool     `json:"isCurbsideAvailable"`
					} `json:"nearbyStore"`
					Maximumdistancechecked float64 `json:"maximumDistanceChecked"`
					Numberofstoreschecked  int     `json:"numberOfStoresChecked"`
					Pickupdate             string  `json:"pickupDate"`
					Holduntildate          string  `json:"holdUntilDate"`
					Availabletoday         bool    `json:"availableToday"`
					Displaydatetype        string  `json:"displayDateType"`
					Isnearby               bool    `json:"isNearby"`
				} `json:"inStorePickup"`
			} `json:"fulfillments"`
			Fulfillmenteligibilities struct {
				Shippingeligible     bool `json:"shippingEligible"`
				Deliveryeligible     bool `json:"deliveryEligible"`
				Pickupeligible       bool `json:"pickupEligible"`
				Fedexpickupeligible  bool `json:"fedExPickupEligible"`
				Downloadeligible     bool `json:"downloadEligible"`
				Emaileligible        bool `json:"emailEligible"`
				Installationeligible bool `json:"installationEligible"`
				Directtoboxeligible  bool `json:"directToBoxEligible"`
				Ispueligible         bool `json:"ispuEligible"`
				Deliverable          bool `json:"deliverable"`
				Shippable            bool `json:"shippable"`
			} `json:"fulfillmentEligibilities"`
			Reservation struct {
				Reservable bool `json:"reservable"`
				Reserved   bool `json:"reserved"`
			} `json:"reservation"`
			Geeksquadinstallationeligiblity string `json:"geekSquadInstallationEligiblity"`
			Ishaulawayeligible              bool   `json:"isHaulAwayEligible"`
			Hasrequiredaccessories          bool   `json:"hasRequiredAccessories"`
			Storeclearance                  bool   `json:"storeClearance"`
			Storeid                         string `json:"storeId"`
			Skutype                         string `json:"skuType"`
			Skusubtype                      string `json:"skuSubType"`
			Price                           struct {
				Base    string `json:"base"`
				Savings struct {
					Total string `json:"total"`
				} `json:"savings"`
				Regular          string `json:"regular"`
				Unitregularprice string `json:"unitRegularPrice"`
			} `json:"price"`
			Isgiftable           bool   `json:"isGiftable"`
			Modelnumber          string `json:"modelNumber"`
			Ismonthlypaymentplan bool   `json:"isMonthlyPaymentPlan"`
			Iswishlistitem       bool   `json:"isWishListItem"`
			Category             struct {
				Department  string `json:"department"`
				Skuclass    string `json:"skuClass"`
				Skusubclass string `json:"skuSubClass"`
			} `json:"category"`
			Mustshiptogether               bool   `json:"mustShipTogether"`
			Isbonusitem                    bool   `json:"isBonusItem"`
			Isgiftwithpurchase             bool   `json:"isGiftWithPurchase"`
			Istspapplieditem               bool   `json:"isTSPAppliedItem"`
			Automaticrenewal               bool   `json:"automaticRenewal"`
			Istotaltechsupportapplied      bool   `json:"isTotalTechSupportApplied"`
			Isrelocationeligible           bool   `json:"isRelocationEligible"`
			Isgreatcallsubscriptionenabled bool   `json:"isGreatCallSubscriptionEnabled"`
			Imageurl                       string `json:"imageUrl"`
		} `json:"meta"`
		Type                 string `json:"type"`
		Quantity             int    `json:"quantity"`
		Sku                  string `json:"sku"`
		Giftmessageselected  bool   `json:"giftMessageSelected"`
		Isbbydigitalgiftcard bool   `json:"isBBYDigitalGiftCard"`
		Fulfillmenttype      string `json:"fulfillmentType"`
		Selectedfulfillment  struct {
			Instorepickup struct {
				Pickupstoreid   string `json:"pickupStoreId"`
				Giftreceipt     bool   `json:"giftReceipt"`
				Storeassist     bool   `json:"storeAssist"`
				Displaydatetype string `json:"displayDateType"`
			} `json:"inStorePickup"`
		} `json:"selectedFulfillment"`
		Creationtime                  string `json:"creationTime"`
		Geeksquadinstallationselected bool   `json:"geekSquadInstallationSelected"`
		Haulawayselected              bool   `json:"haulAwaySelected"`
		Issolutionpackageitem         bool   `json:"isSolutionPackageItem"`
		Relocationselected            bool   `json:"relocationSelected"`
		Sucoapplied                   bool   `json:"sucoApplied"`
		Isleasable                    bool   `json:"isLeasable"`
		Hasoverbox                    bool   `json:"hasOverBox"`
	} `json:"items"`
	Payment struct {
		ID string `json:"id"`
	} `json:"payment"`
	Paymentmethods struct {
		Displayrewardcertificate bool `json:"displayRewardCertificate"`
		Displaygiftcard          bool `json:"displayGiftCard"`
		Giftcardeligible         bool `json:"giftCardEligible"`
		Promocodeeligible        bool `json:"promoCodeEligible"`
		Displaypromotionalcode   bool `json:"displayPromotionalCode"`
		Displaycreditcard        bool `json:"displayCreditCard"`
		Displaypaypal            bool `json:"displayPayPal"`
		Isapplepayeligible       bool `json:"isApplePayEligible"`
		Iscvvrequired            bool `json:"isCvvRequired"`
		Financing                struct {
		} `json:"financing"`
	} `json:"paymentMethods"`
	Profileid    string `json:"profileId"`
	Guestprofile bool   `json:"guestProfile"`
	Guestorder   bool   `json:"guestOrder"`
	State        string `json:"state"`
	Taxexempt    struct {
		Govpurchasecard  bool `json:"govPurchaseCard"`
		Bbytaxexemptcard bool `json:"bbyTaxExemptCard"`
	} `json:"taxExempt"`
	Isageverified             bool   `json:"isAgeVerified"`
	Ispackagevalidated        bool   `json:"isPackageValidated"`
	Customerorderid           string `json:"customerOrderId"`
	Enteredby                 string `json:"enteredBy"`
	Isaccountcreationrequired bool   `json:"isAccountCreationRequired"`
	Isbusinesscustomer        bool   `json:"isBusinessCustomer"`
	Ispaidmember              bool   `json:"isPaidMember"`
	Config                    struct {
		Reactnativeeligible bool `json:"reactNativeEligible"`
	} `json:"config"`
	Preferences struct {
		Supportedspecialpaymentmethods string  `json:"supportedSpecialPaymentMethods"`
		Saleschannel                   string  `json:"salesChannel"`
		Ordertotal                     float64 `json:"orderTotal"`
		Skuid                          string  `json:"skuId"`
		Orderapplepayeligible          bool    `json:"orderApplePayEligible"`
		Paymentmethods                 []struct {
			Method string `json:"method"`
		} `json:"paymentMethods"`
	} `json:"preferences"`
	Isleaseorder bool `json:"isLeaseOrder"`
}

type ItemsRequest []struct {
	ID                  string                   `json:"id"`
	Selectedfulfillment ItemsSelectedfulfillment `json:"selectedFulfillment"`
}
type ItemsShipping struct {
}
type ItemsSelectedfulfillment struct {
	Shipping ItemsShipping `json:"shipping"`
}
type SetShippingRequest struct {
	Phonenumber     string      `json:"phoneNumber"`
	Smsnotifynumber string      `json:"smsNotifyNumber"`
	Smsoptin        bool        `json:"smsOptIn"`
	Emailaddress    string      `json:"emailAddress"`
	Items           []ShipItems `json:"items"`
}
type ShipAddress struct {
	Country             string `json:"country"`
	Savetoprofile       bool   `json:"saveToProfile"`
	Street2             string `json:"street2"`
	Useaddressasbilling bool   `json:"useAddressAsBilling"`
	Middleinitial       string `json:"middleInitial"`
	Lastname            string `json:"lastName"`
	Street              string `json:"street"`
	City                string `json:"city"`
	Override            bool   `json:"override"`
	Zipcode             string `json:"zipcode"`
	State               string `json:"state"`
	Firstname           string `json:"firstName"`
	Iswishlistaddress   bool   `json:"isWishListAddress"`
	Dayphonenumber      string `json:"dayPhoneNumber"`
	Type                string `json:"type"`
}
type Shipping struct {
	Address ShipAddress `json:"address"`
}
type Selectedfulfillment struct {
	Shipping Shipping `json:"shipping"`
}
type ShipItems struct {
	ID                  string              `json:"id"`
	Type                string              `json:"type"`
	Selectedfulfillment Selectedfulfillment `json:"selectedFulfillment"`
	Giftmessageselected bool                `json:"giftMessageSelected"`
}

type UniversalOrderResponse struct {
	ID       string `json:"id"`
	Revision int    `json:"revision"`
	Meta     struct {
		Price struct {
			Producttotal                       string `json:"productTotal"`
			Shipping                           string `json:"shipping"`
			Cardmemberdiscount                 string `json:"cardMemberDiscount"`
			Totalsavings                       string `json:"totalSavings"`
			Totalpaidmemberdiscounts           string `json:"totalPaidMemberDiscounts"`
			Totalnonmemberdiscounts            string `json:"totalNonMemberDiscounts"`
			Rewardcert                         string `json:"rewardCert"`
			Promocodes                         string `json:"promoCodes"`
			Tax                                string `json:"tax"`
			Subtotal                           string `json:"subtotal"`
			Promocodesaftertax                 string `json:"promoCodesAfterTax"`
			Giftcards                          string `json:"giftCards"`
			Total                              string `json:"total"`
			Istaxestimate                      bool   `json:"isTaxEstimate"`
			Istaxcalculated                    bool   `json:"isTaxCalculated"`
			Fulfillmentincentiveamount         string `json:"fulfillmentIncentiveAmount"`
			Eligiblefulfillmentincentiveamount string `json:"eligibleFulfillmentIncentiveAmount"`
			Customerofferamount                string `json:"customerOfferAmount"`
			Leasedproducttotal                 string `json:"leasedProductTotal"`
			Nonleasedproducttotal              string `json:"nonLeasedProductTotal"`
			Nonleasedordertotal                string `json:"nonLeasedOrderTotal"`
		} `json:"price"`
		Hasmobilepackage                 bool `json:"hasMobilePackage"`
		Haswishlistitems                 bool `json:"hasWishListItems"`
		Istspeligible                    bool `json:"isTSPEligible"`
		Isemployeediscountpolicyaccepted bool `json:"isEmployeeDiscountPolicyAccepted"`
		Istotaltechsupportapplied        bool `json:"isTotalTechSupportApplied"`
		Hasmonthlygspitem                bool `json:"hasMonthlyGspItem"`
		Primarypaymentrequired           bool `json:"primaryPaymentRequired"`
		Hasageverificationsku            bool `json:"hasAgeVerificationSku"`
		Hassmsnotifyerror                bool `json:"hasSmsNotifyError"`
		Mybestbuyaccount                 struct {
			Points            int    `json:"points"`
			Dollarsequivalent string `json:"dollarsEquivalent"`
		} `json:"myBestBuyAccount"`
		Ispwpeligible        bool     `json:"isPwpEligible"`
		Pwpeligiblecardtypes []string `json:"pwpEligibleCardTypes"`
		Fulfillments         struct {
			Shipping struct {
				Iseligibleforsamedayscheduling bool `json:"isEligibleForSameDayScheduling"`
			} `json:"shipping"`
		} `json:"fulfillments"`
		Applecareplusdisclaimer            string `json:"applecarePlusDisclaimer"`
		Applecareplustheftdisclaimer       string `json:"applecarePlusTheftDisclaimer"`
		Applecareplustheftyearlydisclaimer string `json:"applecarePlusTheftYearlyDisclaimer"`
	} `json:"meta"`
	Items []struct {
		ID   string `json:"id"`
		Meta struct {
			Shortlabel            string   `json:"shortLabel"`
			Maxquantity           int      `json:"maxQuantity"`
			Availablefulfillments []string `json:"availableFulfillments"`
			Fulfillments          struct {
				Shipping struct {
					Availablelevelsofservice []struct {
						ID                 string `json:"id"`
						Level              string `json:"level"`
						Shipsbymin         string `json:"shipsByMin"`
						Shipsbymax         string `json:"shipsByMax"`
						Price              string `json:"price"`
						Signaturerequired  bool   `json:"signatureRequired"`
						Cutofftime         string `json:"cutOffTime"`
						Guaranteeddelivery bool   `json:"guaranteedDelivery"`
						Name               string `json:"name"`
						Dateindicator      string `json:"dateIndicator"`
						Displaydatetype    string `json:"displayDateType"`
						Customerlosgroupid string `json:"customerLosGroupId"`
						Holidaymessage     bool   `json:"holidayMessage"`
					} `json:"availableLevelsOfService"`
					Availableshiptolocationtypes []string `json:"availableShipToLocationTypes"`
				} `json:"shipping"`
				Instorepickup struct {
					Agentpickup bool `json:"agentPickup"`
					Nearbystore struct {
						ID      string `json:"id"`
						Name    string `json:"name"`
						Type    string `json:"type"`
						Address struct {
							Firstname      string `json:"firstName"`
							Lastname       string `json:"lastName"`
							Street         string `json:"street"`
							City           string `json:"city"`
							State          string `json:"state"`
							Zipcode        string `json:"zipcode"`
							Dayphonenumber string `json:"dayPhoneNumber"`
						} `json:"address"`
						Constrainedquantity bool    `json:"constrainedQuantity"`
						Distance            float64 `json:"distance"`
					} `json:"nearbyStore"`
					Availabletoday bool `json:"availableToday"`
					Isnearby       bool `json:"isNearby"`
				} `json:"inStorePickup"`
			} `json:"fulfillments"`
			Fulfillmenteligibilities struct {
				Shippingeligible     bool `json:"shippingEligible"`
				Deliveryeligible     bool `json:"deliveryEligible"`
				Pickupeligible       bool `json:"pickupEligible"`
				Fedexpickupeligible  bool `json:"fedExPickupEligible"`
				Downloadeligible     bool `json:"downloadEligible"`
				Emaileligible        bool `json:"emailEligible"`
				Installationeligible bool `json:"installationEligible"`
				Directtoboxeligible  bool `json:"directToBoxEligible"`
				Ispueligible         bool `json:"ispuEligible"`
				Deliverable          bool `json:"deliverable"`
				Shippable            bool `json:"shippable"`
			} `json:"fulfillmentEligibilities"`
			Reservation struct {
				Reservable bool `json:"reservable"`
				Reserved   bool `json:"reserved"`
			} `json:"reservation"`
			Geeksquadinstallationeligiblity string `json:"geekSquadInstallationEligiblity"`
			Ishaulawayeligible              bool   `json:"isHaulAwayEligible"`
			Hasrequiredaccessories          bool   `json:"hasRequiredAccessories"`
			Storeclearance                  bool   `json:"storeClearance"`
			Skutype                         string `json:"skuType"`
			Skusubtype                      string `json:"skuSubType"`
			Price                           struct {
				Base    string `json:"base"`
				Savings struct {
					Total string `json:"total"`
				} `json:"savings"`
				Regular          string `json:"regular"`
				Unitregularprice string `json:"unitRegularPrice"`
				Shipping         string `json:"shipping"`
				Shippingdiscount string `json:"shippingDiscount"`
			} `json:"price"`
			Isgiftable           bool   `json:"isGiftable"`
			Modelnumber          string `json:"modelNumber"`
			Ismonthlypaymentplan bool   `json:"isMonthlyPaymentPlan"`
			Iswishlistitem       bool   `json:"isWishListItem"`
			Category             struct {
				Department  string `json:"department"`
				Skuclass    string `json:"skuClass"`
				Skusubclass string `json:"skuSubClass"`
			} `json:"category"`
			Mustshiptogether               bool   `json:"mustShipTogether"`
			Isbonusitem                    bool   `json:"isBonusItem"`
			Isgiftwithpurchase             bool   `json:"isGiftWithPurchase"`
			Istspapplieditem               bool   `json:"isTSPAppliedItem"`
			Automaticrenewal               bool   `json:"automaticRenewal"`
			Istotaltechsupportapplied      bool   `json:"isTotalTechSupportApplied"`
			Isrelocationeligible           bool   `json:"isRelocationEligible"`
			Isgreatcallsubscriptionenabled bool   `json:"isGreatCallSubscriptionEnabled"`
			Imageurl                       string `json:"imageUrl"`
		} `json:"meta"`
		Type                 string `json:"type"`
		Quantity             int    `json:"quantity"`
		Sku                  string `json:"sku"`
		Giftmessageselected  bool   `json:"giftMessageSelected"`
		Isbbydigitalgiftcard bool   `json:"isBBYDigitalGiftCard"`
		Fulfillmenttype      string `json:"fulfillmentType"`
		Selectedfulfillment  struct {
			Shipping struct {
				Levelofservice             string `json:"levelOfService"`
				Selectedcustomerlosgroupid string `json:"selectedCustomerLosGroupId"`
			} `json:"shipping"`
		} `json:"selectedFulfillment"`
		Creationtime                  string `json:"creationTime"`
		Geeksquadinstallationselected bool   `json:"geekSquadInstallationSelected"`
		Haulawayselected              bool   `json:"haulAwaySelected"`
		Issolutionpackageitem         bool   `json:"isSolutionPackageItem"`
		Relocationselected            bool   `json:"relocationSelected"`
		Sucoapplied                   bool   `json:"sucoApplied"`
		Isleasable                    bool   `json:"isLeasable"`
		Hasoverbox                    bool   `json:"hasOverBox"`
	} `json:"items"`
	Payment struct {
		ID string `json:"id"`
	} `json:"payment"`
	Paymentmethods struct {
		Displayrewardcertificate bool `json:"displayRewardCertificate"`
		Displaygiftcard          bool `json:"displayGiftCard"`
		Giftcardeligible         bool `json:"giftCardEligible"`
		Promocodeeligible        bool `json:"promoCodeEligible"`
		Displaypromotionalcode   bool `json:"displayPromotionalCode"`
		Displaycreditcard        bool `json:"displayCreditCard"`
		Displaypaypal            bool `json:"displayPayPal"`
		Isapplepayeligible       bool `json:"isApplePayEligible"`
		Iscvvrequired            bool `json:"isCvvRequired"`
		Financing                struct {
		} `json:"financing"`
	} `json:"paymentMethods"`
	Emailaddress string `json:"emailAddress"`
	Profileid    string `json:"profileId"`
	Guestprofile bool   `json:"guestProfile"`
	Guestorder   bool   `json:"guestOrder"`
	State        string `json:"state"`
	Errors       []struct {
		Errorcode    string `json:"errorCode"`
		Errormessage string `json:"errorMessage"`
		Element      struct {
			Type string `json:"type"`
			ID   string `json:"id"`
		} `json:"element"`
		Metadata struct {
			Type       string `json:"type"`
			Lineitemid string `json:"lineItemId"`
			Address    string `json:"address"`
		} `json:"metadata"`
		Monitoringcode string `json:"monitoringCode"`
	} `json:"errors"`
	Taxexempt struct {
		Govpurchasecard  bool `json:"govPurchaseCard"`
		Bbytaxexemptcard bool `json:"bbyTaxExemptCard"`
	} `json:"taxExempt"`
	Isageverified             bool   `json:"isAgeVerified"`
	Ispackagevalidated        bool   `json:"isPackageValidated"`
	Customerorderid           string `json:"customerOrderId"`
	Enteredby                 string `json:"enteredBy"`
	Isaccountcreationrequired bool   `json:"isAccountCreationRequired"`
	Isbusinesscustomer        bool   `json:"isBusinessCustomer"`
	Ispaidmember              bool   `json:"isPaidMember"`
	Config                    struct {
		Reactnativeeligible bool `json:"reactNativeEligible"`
	} `json:"config"`
	Preferences struct {
		Saleschannel          string  `json:"salesChannel"`
		Ordertotal            float64 `json:"orderTotal"`
		Skuid                 string  `json:"skuId"`
		Orderapplepayeligible bool    `json:"orderApplePayEligible"`
		Paymentmethods        []struct {
			Method string `json:"method"`
		} `json:"paymentMethods"`
	} `json:"preferences"`
	Isleaseorder bool `json:"isLeaseOrder"`
}

type Card struct {
	Type     string `json:"type"`
	Data     string `json:"data"`
	Isv2     bool   `json:"isV2"`
	Keyhost  string `json:"keyHost"`
	Keyclass string `json:"keyClass"`
}

type SetPaymentRequest struct {
	Billingaddress Billingaddress `json:"billingAddress"`
	Creditcard     Creditcard     `json:"creditCard"`
}
type Billingaddress struct {
	Country             string `json:"country"`
	Useaddressasbilling bool   `json:"useAddressAsBilling"`
	Middleinitial       string `json:"middleInitial"`
	Lastname            string `json:"lastName"`
	Iswishlistaddress   bool   `json:"isWishListAddress"`
	City                string `json:"city"`
	State               string `json:"state"`
	Firstname           string `json:"firstName"`
	Addressline1        string `json:"addressLine1"`
	Addressline2        string `json:"addressLine2"`
	Dayphone            string `json:"dayPhone"`
	Postalcode          string `json:"postalCode"`
	Standardized        bool   `json:"standardized"`
	Useroverridden      bool   `json:"userOverridden"`
}
type Creditcard struct {
	Hascid          bool   `json:"hasCID"`
	Invalidcard     bool   `json:"invalidCard"`
	Iscustomercard  bool   `json:"isCustomerCard"`
	Isnewcard       bool   `json:"isNewCard"`
	Isvisacheckout  bool   `json:"isVisaCheckout"`
	Govpurchasecard bool   `json:"govPurchaseCard"`
	Number          string `json:"number"`
	Binnumber       string `json:"binNumber"`
	Ispwpregistered bool   `json:"isPWPRegistered"`
	Expmonth        string `json:"expMonth"`
	Expyear         string `json:"expYear"`
	Cvv             string `json:"cvv"`
	Orderid         string `json:"orderId"`
	Savetoprofile   bool   `json:"saveToProfile"`
	Type            string `json:"type"`
	International   bool   `json:"international"`
	Virtualcard     bool   `json:"virtualCard"`
}

type PrelookupRequest struct {
	Orderid     string      `json:"orderId"`
	Browserinfo Browserinfo `json:"browserInfo"`
}
type Browserinfo struct {
	Javaenabled bool   `json:"javaEnabled"`
	Language    string `json:"language"`
	Useragent   string `json:"userAgent"`
	Height      string `json:"height"`
	Width       string `json:"width"`
	Timezone    string `json:"timeZone"`
	Colordepth  string `json:"colorDepth"`
}

type PrelookupResponse struct {
	Threedsreferenceid      string `json:"threeDSReferenceId"`
	Devicedatacollectionjwt string `json:"deviceDataCollectionJwt"`
	Devicedatacollectionurl string `json:"deviceDataCollectionUrl"`
	Paymentid               string `json:"paymentId"`
}

type PlaceOrderRequest struct {
	Orderid            string             `json:"orderId"`
	Threedsecurestatus Threedsecurestatus `json:"threeDSecureStatus"`
}
type Threedsecurestatus struct {
	Threedsreferenceid string `json:"threeDSReferenceId"`
}

type WebHook struct {
	Content interface{} `json:"content"`
	Embeds  []Embeds    `json:"embeds"`
}
type Fields struct {
	Name   string `json:"name"`
	Value  string `json:"value"`
	Inline bool   `json:"inline,omitempty"`
}
type Footer struct {
	Text    string `json:"text"`
	IconURL string `json:"icon_url"`
}
type Thumbnail struct {
	URL string `json:"url"`
}
type Embeds struct {
	Title     string    `json:"title"`
	Color     int       `json:"color"`
	Fields    []Fields  `json:"fields"`
	Footer    Footer    `json:"footer"`
	Timestamp time.Time `json:"timestamp"`
	Thumbnail Thumbnail `json:"thumbnail"`
}
