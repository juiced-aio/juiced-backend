package entities

import (
	"encoding/json"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// UnmarshalJSON is a Card's redefinition of the default UnmarshalJSON function
func (card *Card) UnmarshalJSON(data []byte) error {
	type CardAlias Card
	temp := &CardAlias{ID: primitive.NewObjectID()}

	err := json.Unmarshal(data, temp)
	if err != nil {
		return err
	}
	*card = Card(*temp)
	return nil
}

// Card is a class that holds details about a Profile's payment method
type Card struct {
	ID             primitive.ObjectID `json:"ID"`
	CardholderName string             `json:"cardHolderName"`
	CardNumber     string             `json:"cardNumber"`
	ExpMonth       string             `json:"expMonth"`
	ExpYear        string             `json:"expYear"`
	CVV            string             `json:"cvv"`
	CardType       string             `json:"cardType"`
}

// SetID updates the Card's ID
func (creditCard *Card) SetID(ID primitive.ObjectID) {
	creditCard.ID = ID
}

// UnmarshalJSON is a Address's redefinition of the default UnmarshalJSON function
func (address *Address) UnmarshalJSON(data []byte) error {
	type AddressAlias Address
	temp := &AddressAlias{ID: primitive.NewObjectID()}

	err := json.Unmarshal(data, temp)
	if err != nil {
		return err
	}
	*address = Address(*temp)
	return nil
}

// SetID updates the Address's ID
func (address *Address) SetID(ID primitive.ObjectID) {
	address.ID = ID
}

// Address is a class that holds details about a Profile's address
type Address struct {
	ID          primitive.ObjectID `json:"ID"`
	FirstName   string             `json:"firstName"`
	LastName    string             `json:"lastName"`
	Address1    string             `json:"address1"`
	Address2    string             `json:"address2"`
	City        string             `json:"city"`
	ZipCode     string             `json:"zipCode"`
	StateCode   string             `json:"stateCode"`
	CountryCode string             `json:"countryCode"`
}

// Profile is a class that holds details about a single profile
type Profile struct {
	ID              primitive.ObjectID `json:"ID"`
	Name            string             `json:"name"`
	Email           string             `json:"email"`
	PhoneNumber     string             `json:"phoneNumber"`
	ShippingAddress Address            `json:"shippingAddress"`
	BillingAddress  Address            `json:"billingAddress"`
	CreditCard      Card               `json:"creditCard"`
}

// Profiles is a class that holds multiple Profiles
type Profiles struct {
	Profiles []Profile `json:"profiles"`
}

// SetID updates the Profile's ID
func (profile *Profile) SetID(ID primitive.ObjectID) {
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
	GroupID  primitive.ObjectID `json:"groupID"`
	Name     string             `json:"name"`
	Profiles []Profile          `json:"profiles"`
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
	GroupID    primitive.ObjectID   `json:"groupID"`
	Name       string               `json:"name"`
	ProfileIDs []primitive.ObjectID `json:"profileIDs"`
}

// AddProfileIDsToGroup adds the given ProfileIDs to the ProfileGroup
func (profileGroup *ProfileGroup) AddProfileIDsToGroup(profileIDsToAdd []primitive.ObjectID) {
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
func (profileGroup *ProfileGroup) RemoveProfileIDsFromGroup(profileIDsToRemove []primitive.ObjectID) {
	profileIDs := []primitive.ObjectID{}
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
func (profileGroup *ProfileGroup) SetProfileIDs(profileIDs []primitive.ObjectID) {
	profileGroup.ProfileIDs = profileIDs
}

// SetName updates the ProfileGroup's Name
func (profileGroup *ProfileGroup) SetName(name string) {
	profileGroup.Name = name
}

// SetGroupID updates the ProfileGroup's GroupID
func (profileGroup *ProfileGroup) SetGroupID(GroupID primitive.ObjectID) {
	profileGroup.GroupID = GroupID
}

// ParseProfileGroup returns a ProfileGroup object parsed from a JSON bytes array
func ParseProfileGroup(profileGroup *ProfileGroup, data []byte) error {
	err := json.Unmarshal(data, &profileGroup)
	return err
}
