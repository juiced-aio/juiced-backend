package entities

import (
	"encoding/json"

	"github.com/google/uuid"
)

// UnmarshalJSON is a Card's redefinition of the default UnmarshalJSON function
func (card *Card) UnmarshalJSON(data []byte) error {
	type CardAlias Card
	temp := &CardAlias{ID: uuid.New().String()}

	err := json.Unmarshal(data, temp)
	if err != nil {
		return err
	}
	*card = Card(*temp)
	return nil
}

// Card is a class that holds details about a Profile's payment method
type Card struct {
	ID             string `json:"ID" db:"ID"`
	ProfileID      string `json:"profileID" db:"profileID"`
	CardholderName string `json:"cardHolderName" db:"cardHolderName"`
	CardNumber     string `json:"cardNumber" db:"cardNumber"`
	ExpMonth       string `json:"expMonth" db:"expMonth"`
	ExpYear        string `json:"expYear" db:"expYear"`
	CVV            string `json:"cvv" db:"cvv"`
	CardType       string `json:"cardType" db:"cardType"`
}

// UnmarshalJSON is a Address's redefinition of the default UnmarshalJSON function
func (address *Address) UnmarshalJSON(data []byte) error {
	type AddressAlias Address
	temp := &AddressAlias{ID: uuid.New().String()}

	err := json.Unmarshal(data, temp)
	if err != nil {
		return err
	}
	*address = Address(*temp)
	return nil
}

// Address is a class that holds details about a Profile's address
type Address struct {
	ID          string `json:"ID" db:"ID"`
	ProfileID   string `json:"profileID" db:"profileID"`
	FirstName   string `json:"firstName" db:"firstName"`
	LastName    string `json:"lastName" db:"lastName"`
	Address1    string `json:"address1" db:"address1"`
	Address2    string `json:"address2" db:"address2"`
	City        string `json:"city" db:"city"`
	ZipCode     string `json:"zipCode" db:"zipCode"`
	StateCode   string `json:"stateCode" db:"stateCode"`
	CountryCode string `json:"countryCode" db:"countryCode"`
}

// Profile is a class that holds details about a single profile
type Profile struct {
	ID                    string   `json:"ID" db:"ID"`
	ProfileGroupIDs       []string `json:"profileGroupIDs" db:"profileGroupIDs"`
	ProfileGroupIDsJoined string   `json:"profileGroupIDsJoined" db:"profileGroupIDsJoined"`
	Name                  string   `json:"name" db:"name"`
	Email                 string   `json:"email" db:"email"`
	PhoneNumber           string   `json:"phoneNumber" db:"phoneNumber"`
	ShippingAddress       *Address `json:"shippingAddress"`
	BillingAddress        *Address `json:"billingAddress"`
	CreditCard            *Card    `json:"creditCard"`
	CreationDate          int64    `json:"creationDate" db:"creationDate"`
}
