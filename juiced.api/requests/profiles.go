package requests

import (
	"backend.juicedbot.io/juiced.infrastructure/entities"
)

type ProfilesRequest struct {
	ProfileIDs []string `json:"profileIDs"`
}

type ImportProfilesRequest struct {
	GroupIDs []string `json:"groupIDs"`
	FilePath string   `json:"filepath"`
}

type ImportProfilesFileFormat struct {
	Profiles []entities.Profile `json:"profiles"`
}

type ExportProfilesRequest struct {
	ProfileIDs []string `json:"profileIDs"`
	FilePath   string   `json:"filepath"`
}

type ExportProfilesFileFormat struct {
	Profiles []ExportProfile `json:"profiles"`
}

type ExportAddress struct {
	FirstName   string `json:"firstName" db:"firstName"`
	LastName    string `json:"lastName" db:"lastName"`
	Address1    string `json:"address1" db:"address1"`
	Address2    string `json:"address2" db:"address2"`
	City        string `json:"city" db:"city"`
	ZipCode     string `json:"zipCode" db:"zipCode"`
	StateCode   string `json:"stateCode" db:"stateCode"`
	CountryCode string `json:"countryCode" db:"countryCode"`
}

type ExportCard struct {
	CardholderName string `json:"cardHolderName" db:"cardHolderName"`
	CardNumber     string `json:"cardNumber" db:"cardNumber"`
	ExpMonth       string `json:"expMonth" db:"expMonth"`
	ExpYear        string `json:"expYear" db:"expYear"`
	CVV            string `json:"cvv" db:"cvv"`
	CardType       string `json:"cardType" db:"cardType"`
}

type ExportProfile struct {
	Name            string        `json:"name" db:"name"`
	Email           string        `json:"email" db:"email"`
	PhoneNumber     string        `json:"phoneNumber" db:"phoneNumber"`
	ShippingAddress ExportAddress `json:"shippingAddress"`
	BillingAddress  ExportAddress `json:"billingAddress"`
	CreditCard      ExportCard    `json:"creditCard"`
}
