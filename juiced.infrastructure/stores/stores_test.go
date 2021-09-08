package stores_test

import (
	"log"
	"os"
	"testing"

	"backend.juicedbot.io/juiced.infrastructure/database"
	"backend.juicedbot.io/juiced.infrastructure/staticstores"
	"backend.juicedbot.io/juiced.infrastructure/stores"
)

func TestMain(m *testing.M) {
	err := database.InitTestDatabase()
	if err == nil {
		err = staticstores.InitStores()
		if err == nil {
			err = stores.InitStores()
			if err == nil {
				os.Exit(m.Run())
			} else {
				log.Println("Error initializing stores: " + err.Error())
			}
		} else {
			log.Println("Error initializing static stores: " + err.Error())
		}
	} else {
		log.Println("Error initializing test database: " + err.Error())
	}
	os.Exit(1)
}

func TestAccountsStore(t *testing.T) {
	t.Run("GetAllAccounts", func(t *testing.T) {

	})

	t.Run("GetAccounts", func(t *testing.T) {

	})

	t.Run("GetAccount", func(t *testing.T) {

	})

	t.Run("GetAccountByEmail", func(t *testing.T) {

	})

	t.Run("CreateAccount", func(t *testing.T) {

	})

	t.Run("UpdateAccount", func(t *testing.T) {

	})

	t.Run("RemoveAccount", func(t *testing.T) {

	})

	t.Run("AccessAccountCookies", func(t *testing.T) {

	})

	t.Run("AccountLogin", func(t *testing.T) {

	})

	t.Run("AccountIsLoggedIn", func(t *testing.T) {

	})
}

func TestProxyGroupStore(t *testing.T) {
	t.Run("GetAllProxyGroups", func(t *testing.T) {
		t.Run("GetAllProxyGroups returns empty array upon first startup", func(t *testing.T) {

		})
		t.Run("GetAllProxyGroups returns ProxyGroups in correct order after adding them", func(t *testing.T) {

		})
		t.Run("Each ProxyGroup returned by GetAllProxyGroups has a unique pointer", func(t *testing.T) {

		})
		t.Run("Each Proxy within each ProxyGroup returned by GetAllProxyGroups has a unique pointer", func(t *testing.T) {

		})
	})

	t.Run("GetProxyGroup", func(t *testing.T) {
		t.Run("GetProxyGroup returns ProxyGroupNotFoundError for invalid proxyGroupID", func(t *testing.T) {

		})
		t.Run("GetProxyGroup returns correct ProxyGroup", func(t *testing.T) {

		})
		t.Run("Each Proxy in the ProxyGroup returned by GetProxyGroup has a unique pointer", func(t *testing.T) {

		})
		t.Run("GetProxyGroup returns the same ProxyGroup pointer if it is called twice on the same proxyGroupID", func(t *testing.T) {

		})
		t.Run("GetProxyGroup returns unique ProxyGroup pointers if it is called on two different proxyGroupIDs", func(t *testing.T) {

		})
	})

	t.Run("CreateProxyGroup", func(t *testing.T) {
		t.Run("CreateProxyGroup returns ProxyGroupAlreadyExistsError if a ProxyGroup with the given name already exists", func(t *testing.T) {

		})
		t.Run("CreateProxyGroup returns correct ProxyGroup", func(t *testing.T) {

		})
		t.Run("CreateProxyGroup adds the ProxyGroup to the proxyGroupStore", func(t *testing.T) {

		})
		t.Run("CreateProxyGroup adds the ProxyGroup to the database", func(t *testing.T) {

		})
	})

	t.Run("UpdateProxyGroup", func(t *testing.T) {
		t.Run("UpdateProxyGroup returns ProxyGroupNotFoundError for invalid proxyGroupID", func(t *testing.T) {

		})
		t.Run("UpdateProxyGroup updates the ProxyGroup's values as expected", func(t *testing.T) {

		})
		t.Run("UpdateProxyGroup doesn't change the ProxyGroup's pointer", func(t *testing.T) {

		})
		t.Run("Each Proxy that still exists in the ProxyGroup updated by UpdateProxyGroup has the same pointer as it did before the ProxyGroup was updated", func(t *testing.T) {

		})
		t.Run("For each Task that uses a Proxy that no longer exists, UpdateProxyGroup replaces the Task's Proxy", func(t *testing.T) {

		})
		t.Run("For each Monitor that uses a Proxy that no longer exists, UpdateProxyGroup replaces the Monitor's Proxy", func(t *testing.T) {

		})
		t.Run("Tasks that use a ProxyGroup that has been updated by UpdateProxyGroup have access to the latest ProxyGroup values", func(t *testing.T) {

		})
		t.Run("Monitors that use a ProxyGroup that has been updated by UpdateProxyGroup have access to the latest ProxyGroup values", func(t *testing.T) {

		})
	})

	t.Run("RemoveProxyGroup", func(t *testing.T) {
		t.Run("RemoveProxyGroup returns ProxyGroupNotFoundError for invalid proxyGroupID", func(t *testing.T) {

		})
		t.Run("RemoveProxyGroup removes the ProxyGroup from the proxyGroupStore", func(t *testing.T) {

		})
		t.Run("RemoveProxyGroup removes the ProxyGroup from the database", func(t *testing.T) {

		})
		t.Run("RemoveProxyGroup removes the ProxyGroup's Proxies from the database", func(t *testing.T) {

		})
		t.Run("RemoveProxyGroup removes the Proxy and ProxyGroup from every Task that uses the deleted ProxyGroup", func(t *testing.T) {

		})
		t.Run("RemoveProxyGroup removes the Proxy and ProxyGroup from every Monitor that uses the deleted ProxyGroup", func(t *testing.T) {

		})
	})

	t.Run("CloneProxyGroup", func(t *testing.T) {
		t.Run("CloneProxyGroup returns ProxyGroupNotFoundError for invalid proxyGroupID", func(t *testing.T) {

		})
		t.Run("CloneProxyGroup returns correct ProxyGroup", func(t *testing.T) {

		})
		t.Run("CloneProxyGroup adds the new ProxyGroup to the proxyGroupStore", func(t *testing.T) {

		})
		t.Run("CloneProxyGroup adds the new ProxyGroup to the database", func(t *testing.T) {

		})
		t.Run("CloneProxyGroup returns a different pointer than the ProxyGroup it is cloning", func(t *testing.T) {

		})
		t.Run("CloneProxyGroup returns a different pointer for each Proxy in the new ProxyGroup than for the corresponding Proxy in the ProxyGroup it is cloning", func(t *testing.T) {

		})
	})
}

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

func TestTaskStore(t *testing.T) {

}

func TestTaskGroupStore(t *testing.T) {

}
