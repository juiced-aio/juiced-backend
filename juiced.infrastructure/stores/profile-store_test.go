package stores_test

import (
	"testing"
	"time"

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
	var profile1Ptr *entities.Profile
	var profile2Ptr *entities.Profile
	// var profile3Ptr *entities.Profile

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

		})
		t.Run("CreateProfile returns correct Profile", func(t *testing.T) {

		})
		t.Run("CreateProfile adds the Profile to the profileStore", func(t *testing.T) {

		})
		t.Run("CreateProfile adds the Profile to the database", func(t *testing.T) {

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
