package stores_test

import (
	"testing"
)

func TestProfileStore(t *testing.T) {
	t.Run("GetAllProfiles", func(t *testing.T) {
		t.Run("GetAllProfiles returns empty array upon first startup", func(t *testing.T) {

		})
		t.Run("GetAllProfiles returns Profiles in correct order after adding them", func(t *testing.T) {

		})
		t.Run("Each Profile returned by GetAllProfiles has a unique pointer", func(t *testing.T) {

		})
	})

	t.Run("GetProfiles", func(t *testing.T) {
		t.Run("GetProfiles returns Profiles in correct order", func(t *testing.T) {

		})
		t.Run("Each Profile returned by GetProfiles has a unique pointer", func(t *testing.T) {

		})
	})

	t.Run("GetProfile", func(t *testing.T) {
		t.Run("GetProfile returns ProfileNotFoundError for invalid profileID", func(t *testing.T) {

		})
		t.Run("GetProfile returns correct Profile", func(t *testing.T) {

		})
		t.Run("GetProfile returns the same Profile pointer if it is called twice on the same profileID", func(t *testing.T) {

		})
		t.Run("GetProfile returns unique Profile pointers if it is called on two different profileIDs", func(t *testing.T) {

		})
	})

	t.Run("GetProfileByName", func(t *testing.T) {
		t.Run("GetProfile returns ProfileNotFoundByNameError for invalid profile name", func(t *testing.T) {

		})
		t.Run("GetProfileByName returns correct Profile", func(t *testing.T) {

		})
		t.Run("GetProfileByName returns the same Profile pointer if it is called twice on the same profileID", func(t *testing.T) {

		})
		t.Run("GetProfileByName returns unique Profile pointers if it is called on two different profileIDs", func(t *testing.T) {

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
