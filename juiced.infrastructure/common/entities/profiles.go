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

// SetID updates the Card's ID
func (creditCard *Card) SetID(ID string) {
	creditCard.ID = ID
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

// SetID updates the Address's ID
func (address *Address) SetID(ID string) {
	address.ID = ID
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
	ShippingAddress       Address  `json:"shippingAddress"`
	BillingAddress        Address  `json:"billingAddress"`
	CreditCard            Card     `json:"creditCard"`
	CreationDate          int64    `json:"creationDate" db:"creationDate"`
}

// SetID updates the Profile's ID
func (profile *Profile) SetID(ID string) {
	profile.ID = ID
}

// SetName updates the Profile's Name
func (profile *Profile) SetName(name string) {
	profile.Name = name
}

// ParseProfile returns a Profile object parsed from a JSON bytes array
func ParseProfile(profile *Profile, data []byte) error {
	err := json.Unmarshal(data, &profile)
	return err
}

// ProfileGroupWithProfiles is a class that holds a list of Profiles
type ProfileGroupWithProfiles struct {
	GroupID  string    `json:"groupID" db:"groupID"`
	Name     string    `json:"name" db:"name"`
	Profiles []Profile `json:"profiles"`
}

// AddProfilesToGroup adds the given ProfileIDs to the ProfileGroupWithProfiles
func (profileGroupWithProfiles *ProfileGroupWithProfiles) AddProfilesToGroup(profilesToAdd []Profile) {
	profiles := profileGroupWithProfiles.Profiles
	for i := 0; i < len(profilesToAdd); i++ {
		alreadyInProfiles := false
		for j := 0; j < len(profiles); j++ {
			if profiles[j].ID == profilesToAdd[i].ID {
				alreadyInProfiles = true
			}
		}
		if !alreadyInProfiles {
			profiles = append(profiles, profilesToAdd[i])
		}
	}
	profileGroupWithProfiles.Profiles = profiles
}

// RemoveProfilesFromGroup removes the given Profiles from the ProfileGroupWithProfiles
func (profileGroupWithProfiles *ProfileGroupWithProfiles) RemoveProfilesFromGroup(profilesToRemove []Profile) {
	profiles := []Profile{}
	for i := 0; i < len(profileGroupWithProfiles.Profiles); i++ {
		inProfilesToRemove := false
		for j := 0; j < len(profilesToRemove); j++ {
			if profilesToRemove[j].ID == profileGroupWithProfiles.Profiles[i].ID {
				inProfilesToRemove = true
			}
		}
		if !inProfilesToRemove {
			profiles = append(profiles, profileGroupWithProfiles.Profiles[i])
		}
	}
	profileGroupWithProfiles.Profiles = profiles
}

// SetProfiles updates the ProfileGroup's Profiles
func (profileGroupWithProfiles *ProfileGroupWithProfiles) SetProfiles(profiles []Profile) {
	profileGroupWithProfiles.Profiles = profiles
}

// ProfileGroup is a class that holds a list of Profile IDs
type ProfileGroup struct {
	GroupID          string   `json:"groupID" db:"groupID"`
	Name             string   `json:"name" db:"name"`
	ProfileIDs       []string `json:"profileIDs"`
	ProfileIDsJoined string   `json:"profileIDsJoined" db:"profileIDsJoined"`
	CreationDate     int64    `json:"creationDate" db:"creationDate"`
}

// AddProfileIDsToGroup adds the given ProfileIDs to the ProfileGroup
func (profileGroup *ProfileGroup) AddProfileIDsToGroup(profileIDsToAdd []string) {
	profileIDs := profileGroup.ProfileIDs
	for i := 0; i < len(profileIDsToAdd); i++ {
		alreadyInProfiles := false
		for j := 0; j < len(profileIDs); j++ {
			if profileIDs[j] == profileIDsToAdd[i] {
				alreadyInProfiles = true
			}
		}
		if !alreadyInProfiles {
			profileIDs = append(profileIDs, profileIDsToAdd[i])
		}
	}
	profileGroup.ProfileIDs = profileIDs
}

// RemoveProfileIDsFromGroup removes the given Profiles from the ProfileGroup
func (profileGroup *ProfileGroup) RemoveProfileIDsFromGroup(profileIDsToRemove []string) {
	profileIDs := []string{}
	for i := 0; i < len(profileGroup.ProfileIDs); i++ {
		inProfilesToRemove := false
		for j := 0; j < len(profileIDsToRemove); j++ {
			if profileIDsToRemove[j] == profileGroup.ProfileIDs[i] {
				inProfilesToRemove = true
			}
		}
		if !inProfilesToRemove {
			profileIDs = append(profileIDs, profileGroup.ProfileIDs[i])
		}
	}
	profileGroup.ProfileIDs = profileIDs
}

// SetProfileIDs updates the ProfileGroup's ProfileIDs
func (profileGroup *ProfileGroup) SetProfileIDs(profileIDs []string) {
	profileGroup.ProfileIDs = profileIDs
}

// SetName updates the ProfileGroup's Name
func (profileGroup *ProfileGroup) SetName(name string) {
	profileGroup.Name = name
}

// SetGroupID updates the ProfileGroup's GroupID
func (profileGroup *ProfileGroup) SetGroupID(GroupID string) {
	profileGroup.GroupID = GroupID
}

// ParseProfileGroup returns a ProfileGroup object parsed from a JSON bytes array
func ParseProfileGroup(profileGroup *ProfileGroup, data []byte) error {
	err := json.Unmarshal(data, &profileGroup)
	return err
}
