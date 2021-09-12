package stores_test

import (
	"testing"
)

func TestProfileGroupStore(t *testing.T) {
	t.Run("GetAllProfileGroups", func(t *testing.T) {
		t.Run("GetAllProfileGroups returns empty array upon first startup", func(t *testing.T) {

		})
		t.Run("GetAllProfileGroups returns ProfileGroup in correct order after adding them", func(t *testing.T) {

		})
		t.Run("Each ProfileGroup returned by GetAllProfileGroups has a unique pointer", func(t *testing.T) {

		})
		t.Run("Each Profile within each ProfileGroup returned by GetAllProfileGroups has a unique pointer", func(t *testing.T) {

		})
	})

	t.Run("GetProfileGroup", func(t *testing.T) {
		t.Run("GetProfileGroup returns ProfileGroupNotFoundError for invalid profileGroupID", func(t *testing.T) {

		})
		t.Run("GetProfileGroup returns correct ProfileGroup", func(t *testing.T) {

		})
		t.Run("Each Profile in the ProfileGroup returned by GetProfileGroup has a unique pointer", func(t *testing.T) {

		})
		t.Run("GetProfileGroup returns the same Profile pointer if it is called twice on the same profileGroupID", func(t *testing.T) {

		})
		t.Run("GetProfileGroup returns different Profile pointers if it is called on two different profileGroupID", func(t *testing.T) {

		})
	})

	t.Run("CreateProfileGroup", func(t *testing.T) {
		t.Run("CreateProfileGroup returns ProfileGroupAlreadyExistsError if a ProfileGroup with the given name already exists", func(t *testing.T) {

		})
		t.Run("CreateProfileGroup returns correct ProfileGroup", func(t *testing.T) {

		})
		t.Run("CreateProfileGroup adds the ProfileGroup to the profileGroupStore", func(t *testing.T) {

		})
		t.Run("CreateProfileGroup adds the ProfileGroup to the database", func(t *testing.T) {

		})
		t.Run("Each Profile in the ProfileGroup returned by CreateProfileGroup has the same pointer as the corresponding Profile in the profileStore", func(t *testing.T) {

		})
	})

	t.Run("UpdateProfileGroup", func(t *testing.T) {
		t.Run("UpdateProfileGroup returns ProfileGroupNotFoundError for invalid profileGroupID", func(t *testing.T) {

		})
		t.Run("UpdateProfileGroup updates the ProfileGroup's values as expected", func(t *testing.T) {

		})
		t.Run("UpdateProfileGroup doesn't change the ProfileGroup's pointer", func(t *testing.T) {

		})
		t.Run("Each Profile that was in the ProfileGroup updated by UpdateProfileGroup has the same pointer as it did before the ProfileGroup was updated", func(t *testing.T) {

		})
	})

	t.Run("RemoveProfileGroup", func(t *testing.T) {
		t.Run("RemoveProfileGroup returns ProfileGroupNotFoundError for invalid profileGroupID", func(t *testing.T) {

		})
		t.Run("RemoveProfileGroup removes the ProfileGroup from the profileGroupStore", func(t *testing.T) {

		})
		t.Run("RemoveProfileGroup removes the ProfileGroup from the database", func(t *testing.T) {

		})
		t.Run("RemoveProfileGroup removes the ProfileGroup's Proxies from the database", func(t *testing.T) {

		})
		t.Run("RemoveProfileGroup removes the ProfileGroupID for every Profile the deleted ProfileGroup contained in the profileGroupStore", func(t *testing.T) {

		})
		t.Run("RemoveProfileGroup removes the ProfileGroupID from every Profile the deleted ProfileGroup contained in the database", func(t *testing.T) {

		})
	})

	t.Run("CloneProfileGroup", func(t *testing.T) {
		t.Run("CloneProfileGroup returns ProfileGroupNotFoundError for invalid profileGroupID", func(t *testing.T) {

		})
		t.Run("CloneProfileGroup returns correct ProfileGroup", func(t *testing.T) {

		})
		t.Run("CloneProfileGroup adds the new ProfileGroup to the proxyGroupStore", func(t *testing.T) {

		})
		t.Run("CloneProfileGroup adds the new ProfileGroup to the database", func(t *testing.T) {

		})
		t.Run("CloneProfileGroup returns a different pointer than the ProfileGroup it is cloning", func(t *testing.T) {

		})
		t.Run("CloneProfileGroup returns the same pointer for each Profile in the new ProfileGroup as the corresponding Profile in the ProfileGroup it is cloning", func(t *testing.T) {

		})
	})

	t.Run("AddProfilesToGroup", func(t *testing.T) {
		t.Run("AddProfilesToGroup returns ProfileGroupNotFoundError for invalid profileGroupID", func(t *testing.T) {

		})
		t.Run("AddProfilesToGroup updates the ProfileGroup's Profiles and ProfileIDs in the profileStore", func(t *testing.T) {

		})
		t.Run("AddProfilesToGroup updates the ProfileGroup's ProfileIDsJoined in the database", func(t *testing.T) {

		})
		t.Run("AddProfilesToGroup doesn't change the ProfileGroup's pointer", func(t *testing.T) {

		})
	})

	t.Run("RemoveProfilesFromGroup", func(t *testing.T) {
		t.Run("RemoveGroupIDsFromProfile returns ProfileGroupNotFoundError for invalid profileGroupID", func(t *testing.T) {

		})
		t.Run("RemoveGroupIDsFromProfile updates the ProfileGroup's Profiles and ProfileIDs in the profileStore", func(t *testing.T) {

		})
		t.Run("RemoveGroupIDsFromProfile updates the ProfileGroup's ProfileIDsJoined in the database", func(t *testing.T) {

		})
		t.Run("RemoveGroupIDsFromProfile doesn't change the ProfileGroup's pointer", func(t *testing.T) {

		})
	})
}
