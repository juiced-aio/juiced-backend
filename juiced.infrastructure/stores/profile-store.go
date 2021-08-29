package stores

import (
	"fmt"
	"time"

	"backend.juicedbot.io/juiced.infrastructure/database"
	"backend.juicedbot.io/juiced.infrastructure/entities"
	"github.com/google/uuid"
)

type ProfileStore struct {
	Profiles map[string]*entities.Profile
}

var profileStore ProfileStore

func (store *ProfileStore) Init() error {
	profiles, err := database.GetAllProfiles()
	if err != nil {
		return err
	}

	for _, profile := range profiles {
		store.Profiles[profile.ID] = &profile
	}

	return nil
}

type ProfileNotFoundError struct {
	ID string
}

func (e *ProfileNotFoundError) Error() string {
	return fmt.Sprintf("Profile with ID %s not found", e.ID)
}

type ProfileNotFoundByNameError struct {
	Name string
}

func (e *ProfileNotFoundByNameError) Error() string {
	return fmt.Sprintf("Profile with name %s not found", e.Name)
}

func GetAllProfiles() []*entities.Profile {
	profiles := []*entities.Profile{}
	for _, profile := range profileStore.Profiles {
		profiles = append(profiles, profile)
	}

	return profiles
}

func GetProfiles(profileIDs []string) []*entities.Profile {
	profiles := []*entities.Profile{}
	for _, profileID := range profileIDs {
		if profile, ok := profileStore.Profiles[profileID]; ok {
			profiles = append(profiles, profile)
		}
	}

	return profiles
}

func GetProfile(profileID string) (*entities.Profile, error) {
	profile, ok := profileStore.Profiles[profileID]
	if !ok {
		return nil, &ProfileNotFoundError{profileID}
	}

	return profile, nil
}

func GetProfileByName(name string) (*entities.Profile, error) {
	for _, profile := range profileStore.Profiles {
		if profile.Name == name {
			return profile, nil
		}
	}

	return nil, &ProfileNotFoundByNameError{name}
}

func CreateProfile(profile entities.Profile) (*entities.Profile, error) {
	if profile.ID == "" {
		profile.ID = uuid.New().String()
	}
	if profile.ShippingAddress.ID == "" {
		profile.ShippingAddress.ID = uuid.New().String()
	}
	if profile.BillingAddress.ID == "" {
		profile.BillingAddress.ID = uuid.New().String()
	}
	if profile.CreditCard.ID == "" {
		profile.CreditCard.ID = uuid.New().String()
	}
	if profile.CreationDate == 0 {
		profile.CreationDate = time.Now().Unix()
	}

	profile.ShippingAddress.ProfileID = profile.ID
	profile.BillingAddress.ProfileID = profile.ID
	profile.CreditCard.ProfileID = profile.ID

	err := database.CreateProfile(profile)

	profilePtr := &profile
	profileStore.Profiles[profile.ID] = profilePtr

	return profilePtr, err
}

func UpdateProfile(profileID string, newProfile entities.Profile) (*entities.Profile, error) {
	profile, err := GetProfile(profileID)
	if err != nil {
		return nil, err
	}

	profile.ProfileGroupIDs = newProfile.ProfileGroupIDs
	profile.Name = newProfile.Name
	profile.Email = newProfile.Email
	profile.PhoneNumber = newProfile.PhoneNumber

	profile.ShippingAddress.FirstName = newProfile.ShippingAddress.FirstName
	profile.ShippingAddress.LastName = newProfile.ShippingAddress.LastName
	profile.ShippingAddress.Address1 = newProfile.ShippingAddress.Address1
	profile.ShippingAddress.Address2 = newProfile.ShippingAddress.Address2
	profile.ShippingAddress.City = newProfile.ShippingAddress.City
	profile.ShippingAddress.ZipCode = newProfile.ShippingAddress.ZipCode
	profile.ShippingAddress.StateCode = newProfile.ShippingAddress.StateCode
	profile.ShippingAddress.CountryCode = newProfile.ShippingAddress.CountryCode

	profile.BillingAddress.FirstName = newProfile.BillingAddress.FirstName
	profile.BillingAddress.LastName = newProfile.BillingAddress.LastName
	profile.BillingAddress.Address1 = newProfile.BillingAddress.Address1
	profile.BillingAddress.Address2 = newProfile.BillingAddress.Address2
	profile.BillingAddress.City = newProfile.BillingAddress.City
	profile.BillingAddress.ZipCode = newProfile.BillingAddress.ZipCode
	profile.BillingAddress.StateCode = newProfile.BillingAddress.StateCode
	profile.BillingAddress.CountryCode = newProfile.BillingAddress.CountryCode

	profile.CreditCard.CardholderName = newProfile.CreditCard.CardholderName
	profile.CreditCard.CardNumber = newProfile.CreditCard.CardNumber
	profile.CreditCard.ExpMonth = newProfile.CreditCard.ExpMonth
	profile.CreditCard.ExpYear = newProfile.CreditCard.ExpYear
	profile.CreditCard.CVV = newProfile.CreditCard.CVV
	profile.CreditCard.CardType = newProfile.CreditCard.CardType

	return profile, nil
}
