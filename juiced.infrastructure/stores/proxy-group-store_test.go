package stores_test

import (
	"testing"
	"time"

	"backend.juicedbot.io/juiced.infrastructure/database"
	"backend.juicedbot.io/juiced.infrastructure/entities"
	"backend.juicedbot.io/juiced.infrastructure/stores"
)

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
			if proxyGroupUpdatePtr.CreationDate != proxyGroup2Ptr.CreationDate {
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
		t.Run("UpdateProxyGroup updates the ProxyGroup in the proxyGroupStore", func(t *testing.T) {
			// TODO
		})
		t.Run("UpdateProxyGroup updates the ProxyGroup in the database", func(t *testing.T) {
			// TODO
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

	var proxyGroup1Removed entities.ProxyGroup
	t.Run("RemoveProxyGroup", func(t *testing.T) {
		t.Run("RemoveProxyGroup returns ProxyGroupNotFoundError for invalid proxyGroupID", func(t *testing.T) {
			_, err := stores.RemoveProxyGroup("INVALID_GROUP_ID")
			if err == nil {
				t.Fatal("RemoveProxyGroup did not return an error on an invalid proxyGroupID\n")
			}
			if _, ok := err.(*stores.ProxyGroupNotFoundError); !ok {
				t.Fatalf("RemoveProxyGroup did not return a ProxyGroupNotFoundError (actual error: %v)\n", err)
			}
		})
		t.Run("RemoveProxyGroup returns the correct ProxyGroup", func(t *testing.T) {
			proxyGroup1BeforeRemove := *proxyGroup1Ptr
			proxyGroup1Removed, err = stores.RemoveProxyGroup(proxyGroup1Ptr.GroupID)
			if err != nil {
				t.Fatalf("RemoveProxyGroup returned an error: %v\n", err)
			}
			if proxyGroup1BeforeRemove.GroupID != proxyGroup1Removed.GroupID {
				t.Fatalf("ProxyGroup returned by RemoveProxyGroup does not have same GroupID (%s) as proxyGroup1BeforeRemove (%s)\n", proxyGroup1Removed.GroupID, proxyGroup1BeforeRemove.GroupID)
			}
			if proxyGroup1BeforeRemove.CreationDate != proxyGroup1Removed.CreationDate {
				t.Fatalf("ProxyGroup returned by RemoveProxyGroup does not have same CreationDate (%d) as proxyGroup1BeforeRemove (%d)\n", proxyGroup1Removed.CreationDate, proxyGroup1BeforeRemove.CreationDate)
			}
			if proxyGroup1BeforeRemove.Name != proxyGroup1Removed.Name {
				t.Fatalf("ProxyGroup returned by RemoveProxyGroup does not have same Name (%s) as proxyGroup1BeforeRemove (%s)\n", proxyGroup1Removed.Name, proxyGroup1BeforeRemove.Name)
			}
			if len(proxyGroup1BeforeRemove.Proxies) != len(proxyGroup1Removed.Proxies) {
				t.Fatalf("ProxyGroup returned by RemoveProxyGroup does not have same number of proxies (%d) as proxyGroup1BeforeRemove (%d)\n", len(proxyGroup1Removed.Proxies), len(proxyGroup1BeforeRemove.Proxies))
			}
			for i := 0; i < len(proxyGroup3.Proxies); i++ {
				if proxyGroup1BeforeRemove.Proxies[i] != proxyGroup1Removed.Proxies[i] {
					t.Fatalf("Proxy #%d in the ProxyGroup returned by RemoveProxyGroup does not have same pointer (%p) as Proxy #%d in proxyGroup1BeforeRemove (%p)\n", i, proxyGroup1Removed.Proxies[i], i, proxyGroup1BeforeRemove.Proxies[i])
				}
			}
		})
		t.Run("RemoveProxyGroup removes the ProxyGroup from the proxyGroupStore", func(t *testing.T) {
			_, err := stores.GetProxyGroup(proxyGroup1Removed.GroupID)
			if err == nil {
				t.Fatal("stores.GetProxyGroup did not return an error on proxyGroup1Removed's GroupID\n")
			}
			if _, ok := err.(*stores.ProxyGroupNotFoundError); !ok {
				t.Fatalf("stores.GetProxyGroup did not return a ProxyGroupNotFoundError (actual error: %v)\n", err)
			}
		})
		t.Run("RemoveProxyGroup removes the ProxyGroup from the database", func(t *testing.T) {
			proxyGroup1RemovedDatabase, err := database.GetProxyGroup(proxyGroup1Removed.GroupID)
			if err != nil {
				t.Fatalf("database.GetProxyGroup returned an error: %v\n", err)
			}
			if proxyGroup1RemovedDatabase.GroupID != "" {
				t.Fatalf("database.GetProxyGroup returned a populated ProxyGroup with GroupID %s\n", proxyGroup1RemovedDatabase.GroupID)
			}
		})
		t.Run("RemoveProxyGroup removes the ProxyGroup's Proxies from the database", func(t *testing.T) {
			proxyGroupWithDatabaseProxies, err := database.GetProxies(entities.ProxyGroup{GroupID: proxyGroup1Removed.GroupID})
			if err != nil {
				t.Fatalf("database.GetProxyGroup returned an error: %v\n", err)
			}
			if len(proxyGroupWithDatabaseProxies.Proxies) != 0 {
				t.Fatalf("database.GetProxyGroup returned a populated ProxyGroup with %d Proxies\n", len(proxyGroupWithDatabaseProxies.Proxies))
			}
		})
		t.Run("RemoveProxyGroup removes the Proxy and ProxyGroup from every Task that uses the deleted ProxyGroup", func(t *testing.T) {
			// TODO
		})
		t.Run("RemoveProxyGroup removes the Proxy and ProxyGroup from every Monitor that uses the deleted ProxyGroup", func(t *testing.T) {
			// TODO
		})
	})

	var proxyGroup2ClonedPtr *entities.ProxyGroup
	t.Run("CloneProxyGroup", func(t *testing.T) {
		t.Run("CloneProxyGroup returns ProxyGroupNotFoundError for invalid proxyGroupID", func(t *testing.T) {
			_, err := stores.CloneProxyGroup("INVALID_GROUP_ID")
			if err == nil {
				t.Fatal("CloneProxyGroup did not return an error on an invalid proxyGroupID\n")
			}
			if _, ok := err.(*stores.ProxyGroupNotFoundError); !ok {
				t.Fatalf("CloneProxyGroup did not return a ProxyGroupNotFoundError (actual error: %v)\n", err)
			}
		})
		t.Run("CloneProxyGroup returns correct ProxyGroup", func(t *testing.T) {
			time.Sleep(1 * time.Second)
			proxyGroup2Ptr.Proxies[0].AddCount()
			proxyGroup2ClonedPtr, err = stores.CloneProxyGroup(proxyGroup2Ptr.GroupID)
			if err != nil {
				t.Fatalf("CloneProxyGroup returned an error: %v\n", err)
			}
			if proxyGroup2ClonedPtr.GroupID == proxyGroup2Ptr.GroupID {
				t.Fatalf("ProxyGroup returned by CloneProxyGroup and stored proxyGroup2Ptr have same GroupID (%s)\n", proxyGroup2ClonedPtr.GroupID)
			}
			if proxyGroup2ClonedPtr.Name == proxyGroup2Ptr.Name {
				t.Fatalf("ProxyGroup returned by CloneProxyGroup and stored proxyGroup2Ptr have same Name (%s)\n", proxyGroup2ClonedPtr.Name)
			}
			if proxyGroup2ClonedPtr.CreationDate == proxyGroup2Ptr.CreationDate {
				t.Fatalf("ProxyGroup returned by CloneProxyGroup and stored proxyGroup2Ptr have same CreationDate (%d)\n", proxyGroup2ClonedPtr.CreationDate)
			}
			if len(proxyGroup2ClonedPtr.Proxies) != len(proxyGroup2Ptr.Proxies) {
				t.Fatalf("ProxyGroup returned by CloneProxyGroup does not have same number of proxies (%d) as stored proxyGroup2Ptr (%d)\n", len(proxyGroup2ClonedPtr.Proxies), len(proxyGroup2Ptr.Proxies))
			}
			for i := 0; i < len(proxyGroup2ClonedPtr.Proxies); i++ {
				if entities.ProxyCleaner(*proxyGroup2ClonedPtr.Proxies[i]) != entities.ProxyCleaner(*proxyGroup2Ptr.Proxies[i]) {
					t.Fatalf("Proxy #%d in the ProxyGroup returned by CloneProxyGroup is not the same value (%s) as Proxy #%d in stored proxyGroup2Ptr (%s)\n", i, entities.ProxyCleaner(*proxyGroup2ClonedPtr.Proxies[i]), i, entities.ProxyCleaner(*proxyGroup2Ptr.Proxies[i]))
				}
				if proxyGroup2ClonedPtr.Proxies[i].ID == proxyGroup2Ptr.Proxies[i].ID {
					t.Fatalf("Proxy #%d in the ProxyGroup returned by CloneProxyGroup has the same ID (%s) as Proxy #%d in stored proxyGroup2Ptr (%s)\n", i, proxyGroup2ClonedPtr.Proxies[i].ID, i, proxyGroup2Ptr.Proxies[i].ID)
				}
				if proxyGroup2ClonedPtr.Proxies[i].CreationDate == proxyGroup2Ptr.Proxies[i].CreationDate {
					t.Fatalf("Proxy #%d in the ProxyGroup returned by CloneProxyGroup has the same CreationDate (%d) as Proxy #%d in stored proxyGroup2Ptr (%d)\n", i, proxyGroup2ClonedPtr.Proxies[i].CreationDate, i, proxyGroup2Ptr.Proxies[i].CreationDate)
				}
				if proxyGroup2ClonedPtr.Proxies[i].ProxyGroupID != proxyGroup2ClonedPtr.GroupID {
					t.Fatalf("Proxy #%d in the ProxyGroup returned by CloneProxyGroup has a different ProxyGroupID (%s) than the cloned ProxyGroup's GroupID (%s)\n", i, proxyGroup2ClonedPtr.Proxies[i].ProxyGroupID, proxyGroup2ClonedPtr.GroupID)
				}
				if proxyGroup2ClonedPtr.Proxies[i].Count != 0 {
					t.Fatalf("Proxy #%d in the ProxyGroup returned by CloneProxyGroup has a non-zero count (%d)\n", i, proxyGroup2ClonedPtr.Proxies[i].Count)
				}
			}
		})
		t.Run("CloneProxyGroup adds the new ProxyGroup to the proxyGroupStore", func(t *testing.T) {
			proxyGroup2ClonedStore, err := stores.GetProxyGroup(proxyGroup2ClonedPtr.GroupID)
			if err != nil {
				t.Fatalf("stores.GetProxyGroup failed: %v\n", err)
			}
			if proxyGroup2ClonedStore != proxyGroup2ClonedPtr {
				t.Fatalf("ProxyGroup returned by GetProxyGroup does not have same pointer (%p) as stored proxyGroup2ClonedPtr (%p)\n", proxyGroup2ClonedStore, proxyGroup2ClonedPtr)
			}
		})
		t.Run("CloneProxyGroup adds the new ProxyGroup to the database", func(t *testing.T) {
			proxyGroup2ClonedDatabase, err := database.GetProxyGroup(proxyGroup3Ptr.GroupID)
			if err != nil {
				t.Fatalf("database.GetProxyGroup failed: %v\n", err)
			}
			if proxyGroup2ClonedDatabase.GroupID != proxyGroup3Ptr.GroupID {
				t.Fatalf("ProxyGroup returned by database.GetProxyGroup does not have same GroupID (%s) as stored proxyGroup2ClonedPtr (%s)\n", proxyGroup2ClonedDatabase.GroupID, proxyGroup2ClonedPtr.GroupID)
			}
			if proxyGroup2ClonedDatabase.Name != proxyGroup3Ptr.Name {
				t.Fatalf("ProxyGroup returned by database.GetProxyGroup does not have same Name (%s) as stored proxyGroup2ClonedPtr (%s)\n", proxyGroup2ClonedDatabase.Name, proxyGroup2ClonedPtr.Name)
			}
			if proxyGroup2ClonedDatabase.CreationDate != proxyGroup3Ptr.CreationDate {
				t.Fatalf("ProxyGroup returned by database.GetProxyGroup does not have same CreationDate (%d) as stored proxyGroup2ClonedPtr (%d)\n", proxyGroup2ClonedDatabase.CreationDate, proxyGroup2ClonedPtr.CreationDate)
			}
		})
		t.Run("CloneProxyGroup returns a different pointer than the ProxyGroup it is cloning", func(t *testing.T) {
			if proxyGroup2ClonedPtr == proxyGroup2Ptr {
				t.Fatalf("ProxyGroup returned by CloneProxyGroup and stored proxyGroup2Ptr have same pointer (%p)\n", proxyGroup2ClonedPtr)
			}
		})
		t.Run("CloneProxyGroup returns a different pointer for each Proxy in the new ProxyGroup than for the corresponding Proxy in the ProxyGroup it is cloning", func(t *testing.T) {
			for i := 0; i < len(proxyGroup2ClonedPtr.Proxies); i++ {
				if proxyGroup2ClonedPtr.Proxies[i] == proxyGroup2Ptr.Proxies[i] {
					t.Fatalf("Proxy #%d in the ProxyGroup returned by CloneProxyGroup has the same pointer (%p) as Proxy #%d in stored proxyGroup2Ptr (%p)\n", i, proxyGroup2ClonedPtr.Proxies[i], i, proxyGroup2Ptr.Proxies[i])
				}
			}
		})
	})
}
