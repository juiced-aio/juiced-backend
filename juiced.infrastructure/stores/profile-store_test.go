package stores_test

import (
	"testing"
	"time"

	"backend.juicedbot.io/juiced.infrastructure/database"
	"backend.juicedbot.io/juiced.infrastructure/entities"
	"backend.juicedbot.io/juiced.infrastructure/stores"
)

func TestProfileStore(t *testing.T) {
	var err error
	address1 := entities.Address{
		FirstName:   "F",
		LastName:    "L",
		Address1:    "123 Main St",
		Address2:    "Apt 1A",
		City:        "Beverly Hills",
		ZipCode:     "90210",
		StateCode:   "CA",
		CountryCode: "US",
	}
	card1 := entities.Card{
		CardholderName: "F L",
		CardNumber:     "4242424242424242",
		ExpMonth:       "12",
		ExpYear:        "2022",
		CVV:            "123",
		CardType:       "VISA",
	}
	profile1 := entities.Profile{
		Name:            "1",
		Email:           "email@email.com",
		PhoneNumber:     "1231231234",
		ShippingAddress: address1,
		BillingAddress:  address1,
		CreditCard:      card1,
	}
	profile2 := entities.Profile{
		Name:            "2",
		Email:           "email@email.com",
		PhoneNumber:     "1231231234",
		ShippingAddress: address1,
		BillingAddress:  address1,
		CreditCard:      card1,
	}
	profile3 := entities.Profile{
		Name:            "3",
		Email:           "email@email.com",
		PhoneNumber:     "1231231234",
		ShippingAddress: address1,
		BillingAddress:  address1,
		CreditCard:      card1,
	}
	var profile1Ptr *entities.Profile
	var profile2Ptr *entities.Profile
	var profile3Ptr *entities.Profile

	t.Run("GetAllProfiles", func(t *testing.T) {
		t.Run("GetAllProfiles returns empty array upon first startup", func(t *testing.T) {
			profiles := stores.GetAllProfiles()
			if len(profiles) != 0 {
				t.Fatal("Array is not empty")
			}
		})
		t.Run("GetAllProfiles returns Profiles in correct order after adding them", func(t *testing.T) {
			profile1Ptr, err = stores.CreateProfile(profile1)
			if err != nil {
				t.Fatalf("CreateProfile 1 failed: %v\n", err)
			}
			time.Sleep(2 * time.Second)
			profile2Ptr, err = stores.CreateProfile(profile2)
			if err != nil {
				t.Fatalf("CreateProfile 2 failed: %v\n", err)
			}
			profiles := stores.GetAllProfiles()
			if len(profiles) != 2 {
				t.Fatalf("Array has %d elements (should have 2)\n", len(profiles))
			}
			if profiles[0].CreationDate >= profiles[1].CreationDate {
				t.Fatalf("First Profile has CreationDate (%d) after second Profile (%d)\n", profiles[0].CreationDate, profiles[1].CreationDate)
			}
			if profiles[0] != profile1Ptr {
				t.Fatalf("First Profile does not have same pointer (%p) as stored profile1Ptr (%p)\n", profiles[0], profile1Ptr)
			}
			if profiles[1] != profile2Ptr {
				t.Fatalf("Second Profile does not have same pointer (%p) as stored profile2Ptr (%p)\n", profiles[0], profile1Ptr)
			}
		})
		t.Run("Each Profile returned by GetAllProfiles has a unique pointer", func(t *testing.T) {
			profiles := stores.GetAllProfiles()
			if len(profiles) != 2 {
				t.Fatalf("Array has %d elements (should have 2)\n", len(profiles))
			}
			if profiles[0] == profiles[1] {
				t.Fatalf("First Profile has same pointer (%p) as second Profile (%p)\n", profiles[0], profiles[1])
			}
		})
	})

	t.Run("GetProfiles", func(t *testing.T) {
		t.Run("GetProfiles returns Profiles in correct order", func(t *testing.T) {
			profiles := stores.GetProfiles([]string{profile1Ptr.ID, profile2Ptr.ID})
			if len(profiles) != 2 {
				t.Fatalf("Array has %d elements (should have 2)\n", len(profiles))
			}
			if profiles[0].CreationDate >= profiles[1].CreationDate {
				t.Fatalf("First Profile has CreationDate (%d) after second Profile (%d)\n", profiles[0].CreationDate, profiles[1].CreationDate)
			}
			if profiles[0] != profile1Ptr {
				t.Fatalf("First Profile does not have same pointer (%p) as stored profile1Ptr (%p)\n", profiles[0], profile1Ptr)
			}
			if profiles[1] != profile2Ptr {
				t.Fatalf("Second Profile does not have same pointer (%p) as stored profile2Ptr (%p)\n", profiles[0], profile1Ptr)
			}
		})
		t.Run("Each Profile returned by GetProfiles has a unique pointer", func(t *testing.T) {

		})
	})

	t.Run("GetProfile", func(t *testing.T) {
		t.Run("GetProfile returns ProfileNotFoundError for invalid profileID", func(t *testing.T) {
			_, err := stores.GetProfile("INVALID_PROFILE_ID")
			if err == nil {
				t.Fatal("GetProfile did not return an error on an invalid profileID\n")
			}
			if _, ok := err.(*stores.ProfileNotFoundError); !ok {
				t.Fatalf("GetProfile did not return a ProfileNotFoundError (actual error: %v)\n", err)
			}
		})
		t.Run("GetProfile returns correct Profile", func(t *testing.T) {
			profilePtr, err := stores.GetProfile(profile1Ptr.ID)
			if err != nil {
				t.Fatalf("GetProfile returned an error on a valid profileID: %v\n", err)
			}
			if profilePtr != profile1Ptr {
				t.Fatalf("Profile returned by GetProfile does not have same pointer (%p) as stored profile1Ptr (%p)\n", profilePtr, profile1Ptr)
			}
		})
		t.Run("GetProfile returns the same Profile pointer if it is called twice on the same profileID", func(t *testing.T) {
			profilePtrA, err := stores.GetProfile(profile1Ptr.ID)
			if err != nil {
				t.Fatalf("GetProfile returned an error: %v\n", err)
			}
			profilePtrB, err := stores.GetProfile(profile1Ptr.ID)
			if err != nil {
				t.Fatalf("GetProfile returned an error: %v\n", err)
			}
			if profilePtrA != profilePtrB {
				t.Fatalf("The two Profiles returned by GetProfile don't have the same pointer (%p, %p)\n", profilePtrA, profilePtrB)
			}
		})
		t.Run("GetProfile returns unique Profile pointers if it is called on two different profileIDs", func(t *testing.T) {
			profilePtr1, err := stores.GetProfile(profile1Ptr.ID)
			if err != nil {
				t.Fatalf("GetProfile returned an error: %v\n", err)
			}
			profilePtr2, err := stores.GetProfile(profile2Ptr.ID)
			if err != nil {
				t.Fatalf("GetProfile returned an error: %v\n", err)
			}
			if profilePtr1 == profilePtr2 {
				t.Fatalf("The two Profiles returned by GetProfile have the same pointer (%p, %p)\n", profilePtr1, profilePtr2)
			}
		})
	})

	t.Run("GetProfileByName", func(t *testing.T) {
		t.Run("GetProfileByName returns ProfileNotFoundByNameError for invalid profile name", func(t *testing.T) {
			_, err := stores.GetProfileByName("INVALID_PROFILE_NAME")
			if err == nil {
				t.Fatal("GetProfileByName did not return an error on an invalid profileID\n")
			}
			if _, ok := err.(*stores.ProfileNotFoundByNameError); !ok {
				t.Fatalf("GetProfileByName did not return a ProfileNotFoundByNameError (actual error: %v)\n", err)
			}
		})
		t.Run("GetProfileByName returns correct Profile", func(t *testing.T) {
			profilePtr, err := stores.GetProfileByName(profile1Ptr.Name)
			if err != nil {
				t.Fatalf("GetProfileByName returned an error on a valid name: %v\n", err)
			}
			if profilePtr != profile1Ptr {
				t.Fatalf("Profile returned by GetProfileByName does not have same pointer (%p) as stored profile1Ptr (%p)\n", profilePtr, profile1Ptr)
			}
		})
		t.Run("GetProfileByName returns the same Profile pointer if it is called twice on the same profileID", func(t *testing.T) {
			profilePtrA, err := stores.GetProfileByName(profile1Ptr.Name)
			if err != nil {
				t.Fatalf("GetProfileByName returned an error: %v\n", err)
			}
			profilePtrB, err := stores.GetProfileByName(profile1Ptr.Name)
			if err != nil {
				t.Fatalf("GetProfileByName returned an error: %v\n", err)
			}
			if profilePtrA != profilePtrB {
				t.Fatalf("The two Profiles returned by GetProfileByName don't have the same pointer (%p, %p)\n", profilePtrA, profilePtrB)
			}
		})
		t.Run("GetProfileByName returns unique Profile pointers if it is called on two different profileIDs", func(t *testing.T) {
			profilePtr1, err := stores.GetProfileByName(profile1Ptr.Name)
			if err != nil {
				t.Fatalf("GetProfileByName returned an error: %v\n", err)
			}
			profilePtr2, err := stores.GetProfileByName(profile2Ptr.Name)
			if err != nil {
				t.Fatalf("GetProfileByName returned an error: %v\n", err)
			}
			if profilePtr1 == profilePtr2 {
				t.Fatalf("The two Profiles returned by GetProfileByName have the same pointer (%p, %p)\n", profilePtr1, profilePtr2)
			}
		})
	})

	t.Run("CreateProfile", func(t *testing.T) {
		t.Run("CreateProfile returns ProfileAlreadyExistsError if a Profile with the given name already exists", func(t *testing.T) {
			_, err := stores.CreateProfile(profile1)
			if err == nil {
				t.Fatal("CreateProfile should have failed due to duplicate name but succeeded\n", err)
			}
			if _, ok := err.(*stores.ProfileAlreadyExistsError); !ok {
				t.Fatalf("CreateProfile did not return a ProfileAlreadyExistsError (actual error: %v)\n", err)
			}
		})
		t.Run("CreateProfile returns correct Profile", func(t *testing.T) {
			profile3Ptr, err = stores.CreateProfile(profile3)
			if err != nil {
				t.Fatalf("CreateProfile failed: %v\n", err)
			}
			if profile3Ptr.Name != profile3.Name {
				t.Fatalf("Created Profile does not have same name (%s) as stored profile3 (%s)\n", profile3Ptr.Name, profile3.Name)
			}
			if profile3Ptr.Email != profile3.Email {
				t.Fatalf("Created Profile does not have same email (%s) as stored profile3 (%s)\n", profile3Ptr.Email, profile3.Email)
			}
			if profile3Ptr.PhoneNumber != profile3.PhoneNumber {
				t.Fatalf("Created Profile does not have same phone number (%s) as stored profile3 (%s)\n", profile3Ptr.PhoneNumber, profile3.PhoneNumber)
			}
			if profile3Ptr.ShippingAddress.FirstName != profile3.ShippingAddress.FirstName {
				t.Fatalf("Created Profile does not have same ShippingAddress.FirstName (%s) as stored profile3 (%s)\n", profile3Ptr.ShippingAddress.FirstName, profile3.ShippingAddress.FirstName)
			}
			if profile3Ptr.ShippingAddress.LastName != profile3.ShippingAddress.LastName {
				t.Fatalf("Created Profile does not have same ShippingAddress.LastName (%s) as stored profile3 (%s)\n", profile3Ptr.ShippingAddress.LastName, profile3.ShippingAddress.FirstName)
			}
			if profile3Ptr.ShippingAddress.Address1 != profile3.ShippingAddress.Address1 {
				t.Fatalf("Created Profile does not have same ShippingAddress.Address1 (%s) as stored profile3 (%s)\n", profile3Ptr.ShippingAddress.Address1, profile3.ShippingAddress.Address1)
			}
			if profile3Ptr.ShippingAddress.Address2 != profile3.ShippingAddress.Address2 {
				t.Fatalf("Created Profile does not have same ShippingAddress.Address2 (%s) as stored profile3 (%s)\n", profile3Ptr.ShippingAddress.Address2, profile3.ShippingAddress.Address2)
			}
			if profile3Ptr.ShippingAddress.City != profile3.ShippingAddress.City {
				t.Fatalf("Created Profile does not have same ShippingAddress.City (%s) as stored profile3 (%s)\n", profile3Ptr.ShippingAddress.City, profile3.ShippingAddress.City)
			}
			if profile3Ptr.ShippingAddress.ZipCode != profile3.ShippingAddress.ZipCode {
				t.Fatalf("Created Profile does not have same ShippingAddress.ZipCode (%s) as stored profile3 (%s)\n", profile3Ptr.ShippingAddress.ZipCode, profile3.ShippingAddress.ZipCode)
			}
			if profile3Ptr.ShippingAddress.StateCode != profile3.ShippingAddress.StateCode {
				t.Fatalf("Created Profile does not have same ShippingAddress.StateCode (%s) as stored profile3 (%s)\n", profile3Ptr.ShippingAddress.StateCode, profile3.ShippingAddress.StateCode)
			}
			if profile3Ptr.ShippingAddress.CountryCode != profile3.ShippingAddress.CountryCode {
				t.Fatalf("Created Profile does not have same ShippingAddress.CountryCode (%s) as stored profile3 (%s)\n", profile3Ptr.ShippingAddress.CountryCode, profile3.ShippingAddress.CountryCode)
			}
			if profile3Ptr.BillingAddress.FirstName != profile3.BillingAddress.FirstName {
				t.Fatalf("Created Profile does not have same BillingAddress.FirstName (%s) as stored profile3 (%s)\n", profile3Ptr.BillingAddress.FirstName, profile3.BillingAddress.FirstName)
			}
			if profile3Ptr.BillingAddress.LastName != profile3.BillingAddress.LastName {
				t.Fatalf("Created Profile does not have same BillingAddress.LastName (%s) as stored profile3 (%s)\n", profile3Ptr.BillingAddress.LastName, profile3.BillingAddress.FirstName)
			}
			if profile3Ptr.BillingAddress.Address1 != profile3.BillingAddress.Address1 {
				t.Fatalf("Created Profile does not have same BillingAddress.Address1 (%s) as stored profile3 (%s)\n", profile3Ptr.BillingAddress.Address1, profile3.BillingAddress.Address1)
			}
			if profile3Ptr.BillingAddress.Address2 != profile3.BillingAddress.Address2 {
				t.Fatalf("Created Profile does not have same BillingAddress.Address2 (%s) as stored profile3 (%s)\n", profile3Ptr.BillingAddress.Address2, profile3.BillingAddress.Address2)
			}
			if profile3Ptr.BillingAddress.City != profile3.BillingAddress.City {
				t.Fatalf("Created Profile does not have same BillingAddress.City (%s) as stored profile3 (%s)\n", profile3Ptr.BillingAddress.City, profile3.BillingAddress.City)
			}
			if profile3Ptr.BillingAddress.ZipCode != profile3.BillingAddress.ZipCode {
				t.Fatalf("Created Profile does not have same BillingAddress.ZipCode (%s) as stored profile3 (%s)\n", profile3Ptr.BillingAddress.ZipCode, profile3.BillingAddress.ZipCode)
			}
			if profile3Ptr.BillingAddress.StateCode != profile3.BillingAddress.StateCode {
				t.Fatalf("Created Profile does not have same BillingAddress.StateCode (%s) as stored profile3 (%s)\n", profile3Ptr.BillingAddress.StateCode, profile3.BillingAddress.StateCode)
			}
			if profile3Ptr.BillingAddress.CountryCode != profile3.BillingAddress.CountryCode {
				t.Fatalf("Created Profile does not have same BillingAddress.CountryCode (%s) as stored profile3 (%s)\n", profile3Ptr.BillingAddress.CountryCode, profile3.BillingAddress.CountryCode)
			}
			if profile3Ptr.CreditCard.CardholderName != profile3.CreditCard.CardholderName {
				t.Fatalf("Created Profile does not have same CreditCard.CardholderName (%s) as stored profile3 (%s)\n", profile3Ptr.CreditCard.CardholderName, profile3.CreditCard.CardholderName)
			}
			if profile3Ptr.CreditCard.CardNumber != profile3.CreditCard.CardNumber {
				t.Fatalf("Created Profile does not have same CreditCard.CardNumber (%s) as stored profile3 (%s)\n", profile3Ptr.CreditCard.CardNumber, profile3.CreditCard.CardNumber)
			}
			if profile3Ptr.CreditCard.ExpMonth != profile3.CreditCard.ExpMonth {
				t.Fatalf("Created Profile does not have same CreditCard.ExpMonth (%s) as stored profile3 (%s)\n", profile3Ptr.CreditCard.ExpMonth, profile3.CreditCard.ExpMonth)
			}
			if profile3Ptr.CreditCard.ExpYear != profile3.CreditCard.ExpYear {
				t.Fatalf("Created Profile does not have same CreditCard.ExpYear (%s) as stored profile3 (%s)\n", profile3Ptr.CreditCard.ExpYear, profile3.CreditCard.ExpYear)
			}
			if profile3Ptr.CreditCard.CVV != profile3.CreditCard.CVV {
				t.Fatalf("Created Profile does not have same CreditCard.CVV (%s) as stored profile3 (%s)\n", profile3Ptr.CreditCard.CVV, profile3.CreditCard.CVV)
			}
			if profile3Ptr.CreditCard.CardType != profile3.CreditCard.CardType {
				t.Fatalf("Created Profile does not have same CreditCard.CardType (%s) as stored profile3 (%s)\n", profile3Ptr.CreditCard.CardType, profile3.CreditCard.CardType)
			}
			if profile3Ptr.ID == "" {
				t.Fatal("Created Profile has an empty ID\n")
			}
			if profile3Ptr.CreationDate == 0 {
				t.Fatal("Created Profile has CreationDate = 0\n")
			}
			if profile3Ptr.ShippingAddress.ID == "" {
				t.Fatal("Created Profile.ShippingAddress has an empty ID\n")
			}
			if profile3Ptr.ShippingAddress.ProfileID != profile3Ptr.ID {
				t.Fatalf("Created Profile.ShippingAddress has different ProfileID (%s) than created Profile's ID (%s)\n", profile3Ptr.ShippingAddress.ProfileID, profile3Ptr.ID)
			}
			if profile3Ptr.BillingAddress.ID == "" {
				t.Fatal("Created Profile.BillingAddress has an empty ID\n")
			}
			if profile3Ptr.BillingAddress.ProfileID != profile3Ptr.ID {
				t.Fatalf("Created Profile.BillingAddress has different ProfileID (%s) than created Profile's ID (%s)\n", profile3Ptr.BillingAddress.ProfileID, profile3Ptr.ID)
			}
			if profile3Ptr.CreditCard.ID == "" {
				t.Fatal("Created Profile.CreditCard has an empty ID\n")
			}
			if profile3Ptr.CreditCard.ProfileID != profile3Ptr.ID {
				t.Fatalf("Created Profile.CreditCard has different ProfileID (%s) than created Profile's ID (%s)\n", profile3Ptr.CreditCard.ProfileID, profile3Ptr.ID)
			}
		})
		t.Run("CreateProfile adds the Profile to the profileStore", func(t *testing.T) {
			profile3Store, err := stores.GetProfile(profile3Ptr.ID)
			if err != nil {
				t.Fatalf("stores.GetProfile failed: %v\n", err)
			}
			if profile3Store != profile3Ptr {
				t.Fatalf("Profile returned by GetProfile does not have same pointer (%p) as stored profile3Ptr (%p)\n", profile3Store, profile3Ptr)
			}
		})
		t.Run("CreateProfile adds the Profile to the database", func(t *testing.T) {
			profile3Database, err := database.GetProfile(profile3Ptr.ID)
			if err != nil {
				t.Fatalf("database.GetProfile failed: %v\n", err)
			}
			if profile3Database.ID != profile3Ptr.ID {
				t.Fatalf("Profile returned by database.GetProfile does not have same ID (%s) as stored profile3Ptr (%s)\n", profile3Database.ID, profile3Ptr.ID)
			}
			if profile3Ptr.Name != profile3.Name {
				t.Fatalf("Profile returned by database.GetProfile does not have same name (%s) as stored profile3 (%s)\n", profile3Ptr.Name, profile3.Name)
			}
			if profile3Ptr.Email != profile3.Email {
				t.Fatalf("Profile returned by database.GetProfile does not have same email (%s) as stored profile3 (%s)\n", profile3Ptr.Email, profile3.Email)
			}
			if profile3Ptr.PhoneNumber != profile3.PhoneNumber {
				t.Fatalf("Profile returned by database.GetProfile does not have same phone number (%s) as stored profile3 (%s)\n", profile3Ptr.PhoneNumber, profile3.PhoneNumber)
			}
			if profile3Ptr.ShippingAddress.FirstName != profile3.ShippingAddress.FirstName {
				t.Fatalf("Profile returned by database.GetProfile does not have same ShippingAddress.FirstName (%s) as stored profile3 (%s)\n", profile3Ptr.ShippingAddress.FirstName, profile3.ShippingAddress.FirstName)
			}
			if profile3Ptr.ShippingAddress.LastName != profile3.ShippingAddress.LastName {
				t.Fatalf("Profile returned by database.GetProfile does not have same ShippingAddress.LastName (%s) as stored profile3 (%s)\n", profile3Ptr.ShippingAddress.LastName, profile3.ShippingAddress.FirstName)
			}
			if profile3Ptr.ShippingAddress.Address1 != profile3.ShippingAddress.Address1 {
				t.Fatalf("Profile returned by database.GetProfile does not have same ShippingAddress.Address1 (%s) as stored profile3 (%s)\n", profile3Ptr.ShippingAddress.Address1, profile3.ShippingAddress.Address1)
			}
			if profile3Ptr.ShippingAddress.Address2 != profile3.ShippingAddress.Address2 {
				t.Fatalf("Profile returned by database.GetProfile does not have same ShippingAddress.Address2 (%s) as stored profile3 (%s)\n", profile3Ptr.ShippingAddress.Address2, profile3.ShippingAddress.Address2)
			}
			if profile3Ptr.ShippingAddress.City != profile3.ShippingAddress.City {
				t.Fatalf("Profile returned by database.GetProfile does not have same ShippingAddress.City (%s) as stored profile3 (%s)\n", profile3Ptr.ShippingAddress.City, profile3.ShippingAddress.City)
			}
			if profile3Ptr.ShippingAddress.ZipCode != profile3.ShippingAddress.ZipCode {
				t.Fatalf("Profile returned by database.GetProfile does not have same ShippingAddress.ZipCode (%s) as stored profile3 (%s)\n", profile3Ptr.ShippingAddress.ZipCode, profile3.ShippingAddress.ZipCode)
			}
			if profile3Ptr.ShippingAddress.StateCode != profile3.ShippingAddress.StateCode {
				t.Fatalf("Profile returned by database.GetProfile does not have same ShippingAddress.StateCode (%s) as stored profile3 (%s)\n", profile3Ptr.ShippingAddress.StateCode, profile3.ShippingAddress.StateCode)
			}
			if profile3Ptr.ShippingAddress.CountryCode != profile3.ShippingAddress.CountryCode {
				t.Fatalf("Profile returned by database.GetProfile does not have same ShippingAddress.CountryCode (%s) as stored profile3 (%s)\n", profile3Ptr.ShippingAddress.CountryCode, profile3.ShippingAddress.CountryCode)
			}
			if profile3Ptr.BillingAddress.FirstName != profile3.BillingAddress.FirstName {
				t.Fatalf("Profile returned by database.GetProfile does not have same BillingAddress.FirstName (%s) as stored profile3 (%s)\n", profile3Ptr.BillingAddress.FirstName, profile3.BillingAddress.FirstName)
			}
			if profile3Ptr.BillingAddress.LastName != profile3.BillingAddress.LastName {
				t.Fatalf("Profile returned by database.GetProfile does not have same BillingAddress.LastName (%s) as stored profile3 (%s)\n", profile3Ptr.BillingAddress.LastName, profile3.BillingAddress.FirstName)
			}
			if profile3Ptr.BillingAddress.Address1 != profile3.BillingAddress.Address1 {
				t.Fatalf("Profile returned by database.GetProfile does not have same BillingAddress.Address1 (%s) as stored profile3 (%s)\n", profile3Ptr.BillingAddress.Address1, profile3.BillingAddress.Address1)
			}
			if profile3Ptr.BillingAddress.Address2 != profile3.BillingAddress.Address2 {
				t.Fatalf("Profile returned by database.GetProfile does not have same BillingAddress.Address2 (%s) as stored profile3 (%s)\n", profile3Ptr.BillingAddress.Address2, profile3.BillingAddress.Address2)
			}
			if profile3Ptr.BillingAddress.City != profile3.BillingAddress.City {
				t.Fatalf("Profile returned by database.GetProfile does not have same BillingAddress.City (%s) as stored profile3 (%s)\n", profile3Ptr.BillingAddress.City, profile3.BillingAddress.City)
			}
			if profile3Ptr.BillingAddress.ZipCode != profile3.BillingAddress.ZipCode {
				t.Fatalf("Profile returned by database.GetProfile does not have same BillingAddress.ZipCode (%s) as stored profile3 (%s)\n", profile3Ptr.BillingAddress.ZipCode, profile3.BillingAddress.ZipCode)
			}
			if profile3Ptr.BillingAddress.StateCode != profile3.BillingAddress.StateCode {
				t.Fatalf("Profile returned by database.GetProfile does not have same BillingAddress.StateCode (%s) as stored profile3 (%s)\n", profile3Ptr.BillingAddress.StateCode, profile3.BillingAddress.StateCode)
			}
			if profile3Ptr.BillingAddress.CountryCode != profile3.BillingAddress.CountryCode {
				t.Fatalf("Profile returned by database.GetProfile does not have same BillingAddress.CountryCode (%s) as stored profile3 (%s)\n", profile3Ptr.BillingAddress.CountryCode, profile3.BillingAddress.CountryCode)
			}
			if profile3Ptr.CreditCard.CardholderName != profile3.CreditCard.CardholderName {
				t.Fatalf("Profile returned by database.GetProfile does not have same CreditCard.CardholderName (%s) as stored profile3 (%s)\n", profile3Ptr.CreditCard.CardholderName, profile3.CreditCard.CardholderName)
			}
			if profile3Ptr.CreditCard.CardNumber != profile3.CreditCard.CardNumber {
				t.Fatalf("Profile returned by database.GetProfile does not have same CreditCard.CardNumber (%s) as stored profile3 (%s)\n", profile3Ptr.CreditCard.CardNumber, profile3.CreditCard.CardNumber)
			}
			if profile3Ptr.CreditCard.ExpMonth != profile3.CreditCard.ExpMonth {
				t.Fatalf("Profile returned by database.GetProfile does not have same CreditCard.ExpMonth (%s) as stored profile3 (%s)\n", profile3Ptr.CreditCard.ExpMonth, profile3.CreditCard.ExpMonth)
			}
			if profile3Ptr.CreditCard.ExpYear != profile3.CreditCard.ExpYear {
				t.Fatalf("Profile returned by database.GetProfile does not have same CreditCard.ExpYear (%s) as stored profile3 (%s)\n", profile3Ptr.CreditCard.ExpYear, profile3.CreditCard.ExpYear)
			}
			if profile3Ptr.CreditCard.CVV != profile3.CreditCard.CVV {
				t.Fatalf("Profile returned by database.GetProfile does not have same CreditCard.CVV (%s) as stored profile3 (%s)\n", profile3Ptr.CreditCard.CVV, profile3.CreditCard.CVV)
			}
			if profile3Ptr.CreditCard.CardType != profile3.CreditCard.CardType {
				t.Fatalf("Profile returned by database.GetProfile does not have same CreditCard.CardType (%s) as stored profile3 (%s)\n", profile3Ptr.CreditCard.CardType, profile3.CreditCard.CardType)
			}
			if profile3Ptr.ID == "" {
				t.Fatal("Profile returned by database.GetProfile has an empty ID\n")
			}
			if profile3Ptr.CreationDate == 0 {
				t.Fatal("Profile returned by database.GetProfile has CreationDate = 0\n")
			}
			if profile3Ptr.ShippingAddress.ID == "" {
				t.Fatal("Profile returned by database.GetProfile.ShippingAddress has an empty ID\n")
			}
			if profile3Ptr.ShippingAddress.ProfileID != profile3Ptr.ID {
				t.Fatalf("Profile returned by database.GetProfile.ShippingAddress has different ProfileID (%s) than created Profile's ID (%s)\n", profile3Ptr.ShippingAddress.ProfileID, profile3Ptr.ID)
			}
			if profile3Ptr.BillingAddress.ID == "" {
				t.Fatal("Profile returned by database.GetProfile.BillingAddress has an empty ID\n")
			}
			if profile3Ptr.BillingAddress.ProfileID != profile3Ptr.ID {
				t.Fatalf("Profile returned by database.GetProfile.BillingAddress has different ProfileID (%s) than created Profile's ID (%s)\n", profile3Ptr.BillingAddress.ProfileID, profile3Ptr.ID)
			}
			if profile3Ptr.CreditCard.ID == "" {
				t.Fatal("Profile returned by database.GetProfile.CreditCard has an empty ID\n")
			}
			if profile3Ptr.CreditCard.ProfileID != profile3Ptr.ID {
				t.Fatalf("Profile returned by database.GetProfile.CreditCard has different ProfileID (%s) than created Profile's ID (%s)\n", profile3Ptr.CreditCard.ProfileID, profile3Ptr.ID)
			}
		})
	})

	t.Run("UpdateProfile", func(t *testing.T) {
		t.Run("UpdateProfile returns ProfileNotFoundError for invalid profileID", func(t *testing.T) {

		})
		t.Run("UpdateProfile updates the Profile's values as expected", func(t *testing.T) {

		})
		t.Run("UpdateProfile doesn't change the Profile's pointer", func(t *testing.T) {

		})
		t.Run("Tasks that use a Profile that has been updated by UpdateProfile have access to the latest Profile values", func(t *testing.T) {

		})
		t.Run("Monitors that use a Profile that has been updated by UpdateProfile have access to the latest Profile values", func(t *testing.T) {

		})
		t.Run("ProfileGroups that have a Profile that has been updated by UpdateProfile have access to the latest Profile values", func(t *testing.T) {

		})
	})

	t.Run("RemoveProfile", func(t *testing.T) {
		t.Run("RemoveProfile returns ProfileNotFoundError for invalid profileID", func(t *testing.T) {

		})
		t.Run("RemoveProfile removes the Profile from the profileStore", func(t *testing.T) {

		})
		t.Run("RemoveProfile removes the Profile from the database", func(t *testing.T) {

		})
		t.Run("RemoveProfile removes every Task that uses the deleted Profile from the taskStore", func(t *testing.T) {

		})
		t.Run("RemoveProfile removes every Task that uses the deleted Profile from the database", func(t *testing.T) {

		})
		t.Run("RemoveProfile removes the Profile and ProfileID for every ProfileGroup the deleted Profile was in in the profileGroupStore", func(t *testing.T) {

		})
		t.Run("RemoveProfile removes the ProfileID from every ProfileGroup the deleted Profile was in in the database", func(t *testing.T) {

		})
	})

	t.Run("CloneProfile", func(t *testing.T) {
		t.Run("CloneProfile returns ProfileNotFoundError for invalid profileID", func(t *testing.T) {

		})
		t.Run("CloneProfile returns correct Profile", func(t *testing.T) {

		})
		t.Run("CloneProfile adds the new Profile to the profileStore", func(t *testing.T) {

		})
		t.Run("CloneProfile adds the new Profile to the database", func(t *testing.T) {

		})
		t.Run("CloneProfile returns a different pointer than the Profile it is cloning", func(t *testing.T) {

		})
	})

	t.Run("AddGroupIDsToProfile", func(t *testing.T) {
		t.Run("AddGroupIDsToProfile returns ProfileNotFoundError for invalid profileID", func(t *testing.T) {

		})
		t.Run("AddGroupIDsToProfile updates the Profile's ProfileGroupIDs in the profileStore", func(t *testing.T) {

		})
		t.Run("AddGroupIDsToProfile updates the Profile's ProfileGroupIDsJoined in the database", func(t *testing.T) {

		})
		t.Run("AddGroupIDToProfile doesn't change the Profile's pointer", func(t *testing.T) {

		})
	})

	t.Run("RemoveGroupIDsFromProfile", func(t *testing.T) {
		t.Run("RemoveGroupIDsFromProfile returns ProfileNotFoundError for invalid profileID", func(t *testing.T) {

		})
		t.Run("RemoveGroupIDsFromProfile updates the Profile's ProfileGroupIDs in the profileStore", func(t *testing.T) {

		})
		t.Run("RemoveGroupIDsFromProfile updates the Profile's ProfileGroupIDsJoined in the database", func(t *testing.T) {

		})
		t.Run("RemoveGroupIDsFromProfile doesn't change the Profile's pointer", func(t *testing.T) {

		})
	})
}
