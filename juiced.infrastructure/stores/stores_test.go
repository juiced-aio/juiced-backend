package stores_test

import (
	"log"
	"os"
	"testing"
	"time"

	"backend.juicedbot.io/juiced.infrastructure/database"
	"backend.juicedbot.io/juiced.infrastructure/entities"
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
	var err error
	proxyGroup1 := entities.ProxyGroup{
		Name: "1",
		Proxies: []*entities.Proxy{
			{
				Host:     "127.0.0.1",
				Port:     "5000",
				Username: "username",
				Password: "password",
			},
			{
				Host: "127.0.0.1",
				Port: "5001",
			},
		},
	}
	proxyGroup2 := entities.ProxyGroup{
		Name: "2",
		Proxies: []*entities.Proxy{
			{
				Host:     "127.0.0.1",
				Port:     "5000",
				Username: "username",
				Password: "password",
			},
		},
	}
	proxyGroup3 := entities.ProxyGroup{
		Name: "3",
		Proxies: []*entities.Proxy{
			{
				Host:     "127.0.0.1",
				Port:     "5000",
				Username: "username",
				Password: "password",
			},
		},
	}
	var proxyGroup1Ptr *entities.ProxyGroup
	var proxyGroup2Ptr *entities.ProxyGroup
	var proxyGroup3Ptr *entities.ProxyGroup

	t.Run("GetAllProxyGroups", func(t *testing.T) {
		t.Run("GetAllProxyGroups returns empty array upon first startup", func(t *testing.T) {
			proxyGroups := stores.GetAllProxyGroups()
			if len(proxyGroups) != 0 {
				t.Fatal("Array is not empty")
			}
		})
		t.Run("GetAllProxyGroups returns ProxyGroups in correct order after adding them", func(t *testing.T) {
			proxyGroup1Ptr, err = stores.CreateProxyGroup(proxyGroup1)
			if err != nil {
				t.Fatalf("CreateProxyGroup 1 failed: %v\n", err)
			}
			time.Sleep(2 * time.Second)
			proxyGroup2Ptr, err = stores.CreateProxyGroup(proxyGroup2)
			if err != nil {
				t.Fatalf("CreateProxyGroup 2 failed: %v\n", err)
			}
			proxyGroups := stores.GetAllProxyGroups()
			if len(proxyGroups) != 2 {
				t.Fatalf("Array has %d elements (should have 2)\n", len(proxyGroups))
			}
			if proxyGroups[0].CreationDate >= proxyGroups[1].CreationDate {
				t.Fatalf("First ProxyGroup has CreationDate (%d) after second ProxyGroup (%d)\n", proxyGroups[0].CreationDate, proxyGroups[1].CreationDate)
			}
			if proxyGroups[0] != proxyGroup1Ptr {
				t.Fatalf("First ProxyGroup does not have same pointer (%p) as stored proxyGroup1Ptr (%p)\n", proxyGroups[0], proxyGroup1Ptr)
			}
			if proxyGroups[1] != proxyGroup2Ptr {
				t.Fatalf("Second ProxyGroup does not have same pointer (%p) as stored proxyGroup2Ptr (%p)\n", proxyGroups[0], proxyGroup1Ptr)
			}
		})
		t.Run("Each ProxyGroup returned by GetAllProxyGroups has a unique pointer", func(t *testing.T) {
			proxyGroups := stores.GetAllProxyGroups()
			if len(proxyGroups) != 2 {
				t.Fatalf("Array has %d elements (should have 2)\n", len(proxyGroups))
			}
			if proxyGroups[0] == proxyGroups[1] {
				t.Fatalf("First ProxyGroup has same pointer (%p) as second ProxyGroup (%p)\n", proxyGroups[0], proxyGroups[1])
			}
		})
		t.Run("Each Proxy within each ProxyGroup returned by GetAllProxyGroups has a unique pointer", func(t *testing.T) {
			proxyGroups := stores.GetAllProxyGroups()
			if len(proxyGroups) != 2 {
				t.Fatalf("Array has %d elements (should have 2)\n", len(proxyGroups))
			}
			proxies := append(proxyGroups[0].Proxies, proxyGroups[1].Proxies...)
			visited := make(map[*entities.Proxy]bool)
			for _, proxy := range proxies {
				if visited[proxy] {
					t.Fatalf("Two proxies have the same pointer (%p)\n", proxy)
				} else {
					visited[proxy] = true
				}
			}
		})
	})

	t.Run("GetProxyGroup", func(t *testing.T) {
		t.Run("GetProxyGroup returns ProxyGroupNotFoundError for invalid proxyGroupID", func(t *testing.T) {
			_, err := stores.GetProxyGroup("INVALID_GROUP_ID")
			if err == nil {
				t.Fatal("GetProxyGroup did not return an error on an invalid proxyGroupID\n")
			}
			if _, ok := err.(*stores.ProxyGroupNotFoundError); !ok {
				t.Fatalf("GetProxyGroup did not return a ProxyGroupNotFoundError (actual error: %v)\n", err)
			}
		})
		t.Run("GetProxyGroup returns correct ProxyGroup", func(t *testing.T) {
			proxyGroupPtr, err := stores.GetProxyGroup(proxyGroup1Ptr.GroupID)
			if err != nil {
				t.Fatalf("GetProxyGroup returned an error on a valid proxyGroupID: %v\n", err)
			}
			if proxyGroupPtr != proxyGroup1Ptr {
				t.Fatalf("ProxyGroup returned by GetProxyGroup does not have same pointer (%p) as stored proxyGroup1Ptr (%p)\n", proxyGroupPtr, proxyGroup1Ptr)
			}
			if len(proxyGroupPtr.Proxies) != len(proxyGroup1Ptr.Proxies) {
				t.Fatalf("ProxyGroup returned by GetProxyGroup does not have same number of proxies (%d) as stored proxyGroup1Ptr (%d)\n", len(proxyGroupPtr.Proxies), len(proxyGroup1Ptr.Proxies))
			}
			for i := 0; i < len(proxyGroupPtr.Proxies); i++ {
				if proxyGroupPtr.Proxies[i] != proxyGroup1Ptr.Proxies[i] {
					t.Fatalf("Proxy #%d in ProxyGroup returned by GetProxyGroup does not have same pointer (%p) as Proxy #%d in stored proxyGroup1Ptr (%p)\n", i, proxyGroupPtr.Proxies[i], i, proxyGroup1Ptr.Proxies[i])
				}
			}
		})
		t.Run("Each Proxy in the ProxyGroup returned by GetProxyGroup has a unique pointer", func(t *testing.T) {
			proxyGroupPtr, err := stores.GetProxyGroup(proxyGroup1Ptr.GroupID)
			if err != nil {
				t.Fatalf("GetProxyGroup returned an error: %v\n", err)
			}
			visited := make(map[*entities.Proxy]bool)
			for _, proxy := range proxyGroupPtr.Proxies {
				if visited[proxy] {
					t.Fatalf("Two proxies have the same pointer (%p)\n", proxy)
				} else {
					visited[proxy] = true
				}
			}
		})
		t.Run("GetProxyGroup returns the same ProxyGroup pointer if it is called twice on the same proxyGroupID", func(t *testing.T) {
			proxyGroupPtrA, err := stores.GetProxyGroup(proxyGroup1Ptr.GroupID)
			if err != nil {
				t.Fatalf("GetProxyGroup returned an error: %v\n", err)
			}
			proxyGroupPtrB, err := stores.GetProxyGroup(proxyGroup1Ptr.GroupID)
			if err != nil {
				t.Fatalf("GetProxyGroup returned an error: %v\n", err)
			}
			if proxyGroupPtrA != proxyGroupPtrB {
				t.Fatalf("The two ProxyGroups returned by GetProxyGroup don't have the same pointer (%p, %p)\n", proxyGroupPtrA, proxyGroupPtrB)
			}
			if len(proxyGroupPtrA.Proxies) != len(proxyGroupPtrB.Proxies) {
				t.Fatalf("The two ProxyGroups returned by GetProxyGroup don't have same number of proxies (%d, %d)\n", len(proxyGroupPtrA.Proxies), len(proxyGroupPtrB.Proxies))
			}
			for i := 0; i < len(proxyGroupPtrA.Proxies); i++ {
				if proxyGroupPtrA.Proxies[i] != proxyGroupPtrB.Proxies[i] {
					t.Fatalf("Proxy #%d in the two ProxyGroups returned by GetProxyGroup don't have same pointer (%p, %p)\n", i, proxyGroupPtrA.Proxies[i], proxyGroupPtrB.Proxies[i])
				}
			}
		})
		t.Run("GetProxyGroup returns unique ProxyGroup pointers if it is called on two different proxyGroupIDs", func(t *testing.T) {
			proxyGroupPtr1, err := stores.GetProxyGroup(proxyGroup1Ptr.GroupID)
			if err != nil {
				t.Fatalf("GetProxyGroup returned an error: %v\n", err)
			}
			proxyGroupPtr2, err := stores.GetProxyGroup(proxyGroup2Ptr.GroupID)
			if err != nil {
				t.Fatalf("GetProxyGroup returned an error: %v\n", err)
			}
			if proxyGroupPtr1 == proxyGroupPtr2 {
				t.Fatalf("The two ProxyGroups returned by GetProxyGroup have the same pointer (%p, %p)\n", proxyGroupPtr1, proxyGroupPtr2)
			}
		})
	})

	t.Run("CreateProxyGroup", func(t *testing.T) {
		t.Run("CreateProxyGroup returns ProxyGroupAlreadyExistsError if a ProxyGroup with the given name already exists", func(t *testing.T) {
			_, err := stores.CreateProxyGroup(proxyGroup1)
			if err == nil {
				t.Fatal("CreateProxyGroup should have failed due to duplicate name but succeeded\n", err)
			}
			if _, ok := err.(*stores.ProxyGroupAlreadyExistsError); !ok {
				t.Fatalf("CreateProxyGroup did not return a ProxyGroupAlreadyExistsError (actual error: %v)\n", err)
			}
		})
		t.Run("CreateProxyGroup returns correct ProxyGroup", func(t *testing.T) {
			proxyGroup3Ptr, err = stores.CreateProxyGroup(proxyGroup3)
			if err != nil {
				t.Fatalf("CreateProxyGroup failed: %v\n", err)
			}
			if proxyGroup3Ptr.Name != proxyGroup3.Name {
				t.Fatalf("Created ProxyGroup does not have same name (%s) as stored proxyGroup3 (%s)\n", proxyGroup3Ptr.Name, proxyGroup3.Name)
			}
			if len(proxyGroup3Ptr.Proxies) != len(proxyGroup3.Proxies) {
				t.Fatalf("Created ProxyGroup does not have same number of proxies (%d) as stored proxyGroup3 (%d)\n", len(proxyGroup3Ptr.Proxies), len(proxyGroup3.Proxies))
			}
			for i := 0; i < len(proxyGroup3.Proxies); i++ {
				if proxyGroup3Ptr.Proxies[i] != proxyGroup3.Proxies[i] {
					t.Fatalf("Proxy #%d in the ProxyGroup returned by CreateProxyGroup does not have same pointer (%p) as Proxy #%d in stored proxyGroup3 (%p)\n", i, proxyGroup3Ptr.Proxies[i], i, proxyGroup3.Proxies[i])
				}
			}
			if proxyGroup3Ptr.GroupID == "" {
				t.Fatal("Created ProxyGroup has an empty GroupID\n")
			}
			if proxyGroup3Ptr.CreationDate == 0 {
				t.Fatal("Created ProxyGroup has CreationDate = 0\n")
			}
			for i := 0; i < len(proxyGroup3Ptr.Proxies); i++ {
				if proxyGroup3Ptr.Proxies[i].ID == "" {
					t.Fatalf("Proxy #%d in the created ProxyGroup has an empty ID\n", i)
				}
				if proxyGroup3Ptr.Proxies[i].ProxyGroupID == "" {
					t.Fatalf("Proxy #%d in the created ProxyGroup has an empty ProxyGroupID\n", i)
				}
				if proxyGroup3Ptr.Proxies[i].CreationDate == 0 {
					t.Fatalf("Proxy #%d in the created ProxyGroup has CreationDate = 0\n", i)
				}
			}
		})
		t.Run("CreateProxyGroup adds the ProxyGroup to the proxyGroupStore", func(t *testing.T) {
			proxyGroup3Store, err := stores.GetProxyGroup(proxyGroup3Ptr.GroupID)
			if err != nil {
				t.Fatalf("stores.GetProxyGroup failed: %v\n", err)
			}
			if proxyGroup3Store != proxyGroup3Ptr {
				t.Fatalf("ProxyGroup returned by GetProxyGroup does not have same pointer (%p) as stored proxyGroup3Ptr (%p)\n", proxyGroup3Store, proxyGroup3Ptr)
			}
		})
		t.Run("CreateProxyGroup adds the ProxyGroup to the database", func(t *testing.T) {
			proxyGroup3Database, err := database.GetProxyGroup(proxyGroup3Ptr.GroupID)
			if err != nil {
				t.Fatalf("database.GetProxyGroup failed: %v\n", err)
			}
			if proxyGroup3Database.GroupID != proxyGroup3Ptr.GroupID {
				t.Fatalf("ProxyGroup returned by database.GetProxyGroup does not have same GroupID (%s) as stored proxyGroup3Ptr (%s)\n", proxyGroup3Database.GroupID, proxyGroup3Ptr.GroupID)
			}
			if proxyGroup3Database.Name != proxyGroup3Ptr.Name {
				t.Fatalf("ProxyGroup returned by database.GetProxyGroup does not have same Name (%s) as stored proxyGroup3Ptr (%s)\n", proxyGroup3Database.Name, proxyGroup3Ptr.Name)
			}
			if proxyGroup3Database.CreationDate != proxyGroup3Ptr.CreationDate {
				t.Fatalf("ProxyGroup returned by database.GetProxyGroup does not have same CreationDate (%d) as stored proxyGroup3Ptr (%d)\n", proxyGroup3Database.CreationDate, proxyGroup3Ptr.CreationDate)
			}
		})
	})

	proxyGroupUpdate := entities.ProxyGroup{
		Name: "UPDATE",
		Proxies: []*entities.Proxy{
			{ // Same as proxyGroup2
				Host:     "127.0.0.1",
				Port:     "5000",
				Username: "username",
				Password: "password",
			},
			{ // New
				Host:     "127.0.0.1",
				Port:     "5001",
				Username: "username",
				Password: "password",
			},
		},
	}
	var proxyGroupUpdatePtr *entities.ProxyGroup
	t.Run("UpdateProxyGroup", func(t *testing.T) {
		t.Run("UpdateProxyGroup returns ProxyGroupNotFoundError for invalid proxyGroupID", func(t *testing.T) {
			_, err := stores.UpdateProxyGroup("INVALID_GROUP_ID", proxyGroupUpdate)
			if err == nil {
				t.Fatal("GetProxyGroup did not return an error on an invalid proxyGroupID\n")
			}
			if _, ok := err.(*stores.ProxyGroupNotFoundError); !ok {
				t.Fatalf("GetProxyGroup did not return a ProxyGroupNotFoundError (actual error: %v)\n", err)
			}
		})
		t.Run("UpdateProxyGroup updates the ProxyGroup's values as expected", func(t *testing.T) {
			proxyGroupUpdatePtr, err = stores.UpdateProxyGroup(proxyGroup2Ptr.GroupID, proxyGroupUpdate)
			if err != nil {
				t.Fatalf("UpdateProxyGroup returned an error: %v\n", err)
			}
			if proxyGroupUpdatePtr.Name != proxyGroupUpdate.Name {
				t.Fatalf("ProxyGroup returned by UpdateProxyGroup does not have same Name (%s) as proxyGroupUpdate (%s)\n", proxyGroupUpdatePtr.Name, proxyGroupUpdate.Name)
			}
			if len(proxyGroupUpdatePtr.Proxies) != len(proxyGroupUpdate.Proxies) {
				t.Fatalf("ProxyGroup returned by UpdateProxyGroup does not have same number of proxies (%d) as proxyGroupUpdate (%d)\n", len(proxyGroupUpdatePtr.Proxies), len(proxyGroupUpdate.Proxies))
			}
			for i := 0; i < len(proxyGroup3.Proxies); i++ {
				if proxyGroupUpdatePtr.Proxies[i] != proxyGroupUpdate.Proxies[i] {
					t.Fatalf("Proxy #%d in the ProxyGroup returned by UpdateProxyGroup does not have same pointer (%p) as Proxy #%d in proxyGroupUpdate (%p)\n", i, proxyGroupUpdatePtr.Proxies[i], i, proxyGroupUpdate.Proxies[i])
				}
			}
		})
		t.Run("UpdateProxyGroup doesn't change the ProxyGroup's pointer", func(t *testing.T) {
			if proxyGroupUpdatePtr != proxyGroup2Ptr {
				t.Fatalf("ProxyGroup returned by UpdateProxyGroup does not have same pointer (%p) as stored proxyGroup2Ptr (%p)\n", proxyGroupUpdatePtr, proxyGroup2Ptr)
			}
			if proxyGroupUpdatePtr.GroupID != proxyGroup2Ptr.GroupID {
				t.Fatalf("ProxyGroup returned by UpdateProxyGroup does not have same GroupID (%s) as stored proxyGroup2Ptr (%s)\n", proxyGroupUpdatePtr.GroupID, proxyGroup2Ptr.GroupID)
			}
			if proxyGroupUpdatePtr.CreationDate-proxyGroup2Ptr.CreationDate != 0 {
				t.Fatalf("ProxyGroup returned by UpdateProxyGroup does not have same CreationDate (%d) as stored proxyGroup2Ptr (%d)\n", proxyGroupUpdatePtr.CreationDate, proxyGroup2Ptr.CreationDate)
			}
			if proxyGroupUpdatePtr.Name != proxyGroup2Ptr.Name {
				t.Fatalf("ProxyGroup returned by UpdateProxyGroup does not have same Name (%s) as stored proxyGroup2Ptr (%s)\n", proxyGroupUpdatePtr.Name, proxyGroup2Ptr.Name)
			}
			if len(proxyGroupUpdatePtr.Proxies) != len(proxyGroup2Ptr.Proxies) {
				t.Fatalf("ProxyGroup returned by UpdateProxyGroup does not have same number of proxies (%d) as stored proxyGroup2Ptr (%d)\n", len(proxyGroupUpdatePtr.Proxies), len(proxyGroup2Ptr.Proxies))
			}
			for i := 0; i < len(proxyGroup3.Proxies); i++ {
				if proxyGroupUpdatePtr.Proxies[i] != proxyGroup2Ptr.Proxies[i] {
					t.Fatalf("Proxy #%d in the ProxyGroup returned by UpdateProxyGroup does not have same pointer (%p) as Proxy #%d in stored proxyGroup2Ptr (%p)\n", i, proxyGroupUpdatePtr.Proxies[i], i, proxyGroup2Ptr.Proxies[i])
				}
			}
		})
		t.Run("Each Proxy that still exists in the ProxyGroup updated by UpdateProxyGroup has the same pointer as it did before the ProxyGroup was updated", func(t *testing.T) {
			// TODO
		})
		t.Run("For each Task that uses a Proxy that no longer exists, UpdateProxyGroup replaces the Task's Proxy", func(t *testing.T) {
			// TODO
		})
		t.Run("For each Monitor that uses a Proxy that no longer exists, UpdateProxyGroup replaces the Monitor's Proxy", func(t *testing.T) {
			// TODO
		})
		t.Run("Tasks that use a ProxyGroup that has been updated by UpdateProxyGroup have access to the latest ProxyGroup values", func(t *testing.T) {
			// TODO
		})
		t.Run("Monitors that use a ProxyGroup that has been updated by UpdateProxyGroup have access to the latest ProxyGroup values", func(t *testing.T) {
			// TODO
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
			// TODO
		})
		t.Run("RemoveProxyGroup removes the Proxy and ProxyGroup from every Monitor that uses the deleted ProxyGroup", func(t *testing.T) {
			// TODO
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
