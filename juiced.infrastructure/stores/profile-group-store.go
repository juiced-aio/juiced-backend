package stores

import (
	"fmt"
	"sort"
	"strings"
	"time"

	"backend.juicedbot.io/juiced.infrastructure/database"
	"backend.juicedbot.io/juiced.infrastructure/entities"
	"backend.juicedbot.io/juiced.infrastructure/util"
	"github.com/google/uuid"
)

type ProfileGroupStore struct {
	ProfileGroups map[string]*entities.ProfileGroup
}

var profileGroupStore ProfileGroupStore

func InitProfileGroupStore() error {
	profileGroupStore = ProfileGroupStore{
		ProfileGroups: make(map[string]*entities.ProfileGroup),
	}

	profileGroups, err := database.GetAllProfileGroups()
	if err != nil {
		return err
	}

	for _, profileGroup := range profileGroups {
		profileGroup := profileGroup
		profileGroup.Profiles = GetProfiles(profileGroup.ProfileIDs)
		profileGroupStore.ProfileGroups[profileGroup.GroupID] = &profileGroup
	}

	return nil
}

type ProfileGroupNotFoundError struct {
	ID string
}

func (e *ProfileGroupNotFoundError) Error() string {
	return fmt.Sprintf("ProfileGroup with ID %s not found", e.ID)
}

func GetAllProfileGroups() []*entities.ProfileGroup {
	profileGroups := []*entities.ProfileGroup{}

	for _, profileGroup := range profileGroupStore.ProfileGroups {
		profileGroups = append(profileGroups, profileGroup)
	}

	sort.SliceStable(profileGroups, func(i, j int) bool {
		return profileGroups[i].CreationDate < profileGroups[j].CreationDate
	})

	return profileGroups
}

func GetProfileGroup(groupID string) (*entities.ProfileGroup, error) {
	profileGroup, ok := profileGroupStore.ProfileGroups[groupID]
	if !ok {
		return nil, &ProfileGroupNotFoundError{groupID}
	}

	return profileGroup, nil
}

func CreateProfileGroup(profileGroup entities.ProfileGroup) (*entities.ProfileGroup, error) {
	if profileGroup.GroupID == "" {
		profileGroup.GroupID = uuid.New().String()
	}
	if profileGroup.CreationDate == 0 {
		profileGroup.CreationDate = time.Now().Unix()
	}
	profileGroup.Profiles = GetProfiles(profileGroup.ProfileIDs)

	err := database.CreateProfileGroup(profileGroup)
	if err != nil {
		return nil, err
	}

	profileGroupPtr := &profileGroup
	profileGroupStore.ProfileGroups[profileGroup.GroupID] = profileGroupPtr

	return profileGroupPtr, nil
}

func UpdateProfileGroup(groupID string, newProfileGroup entities.ProfileGroup) (*entities.ProfileGroup, error) {
	profileGroup, err := GetProfileGroup(groupID)
	if err != nil {
		return nil, err
	}

	profileGroup.Name = newProfileGroup.Name
	profileGroup.ProfileIDs = newProfileGroup.ProfileIDs
	profileGroup.ProfileIDsJoined = newProfileGroup.ProfileIDsJoined
	profileGroup.Profiles = GetProfiles(profileGroup.ProfileIDs)

	return profileGroup, database.UpdateProfileGroup(groupID, *profileGroup)
}

func RemoveProfileGroup(groupID string) (entities.ProfileGroup, error) {
	profileGroup, err := GetProfileGroup(groupID)
	if err != nil {
		return entities.ProfileGroup{}, err
	}

	for _, profileID := range profileGroup.ProfileIDs {
		RemoveGroupIDFromProfile(profileID, groupID)
	}

	delete(profileGroupStore.ProfileGroups, groupID)

	return *profileGroup, database.RemoveProfileGroup(groupID)
}

func CloneProfileGroup(groupID string) (*entities.ProfileGroup, error) {
	profileGroupPtr, err := GetProfileGroup(groupID)
	if err != nil {
		return nil, err
	}

	newProfileGroup := *profileGroupPtr
	newProfileGroup.Name += " " + util.RandString(3)
	newProfileGroup.GroupID = ""
	newProfileGroup.CreationDate = 0

	newProfileGroupPtr, err := CreateProfileGroup(newProfileGroup)
	if err != nil {
		return nil, err
	}

	for _, profileID := range newProfileGroupPtr.ProfileIDs {
		AddGroupIDToProfile(profileID, newProfileGroupPtr.GroupID)
	}

	return newProfileGroupPtr, nil
}

func AddProfilesToGroup(groupID string, profileIDs []string) (*entities.ProfileGroup, error) {
	profileGroupPtr, err := GetProfileGroup(groupID)
	if err != nil {
		return nil, err
	}

	for _, profileID := range profileIDs {
		newProfile, err := AddGroupIDToProfile(profileID, groupID)
		if err == nil {
			profileGroupPtr.ProfileIDs = append(profileGroupPtr.ProfileIDs, profileID)
			profileGroupPtr.Profiles = append(profileGroupPtr.Profiles, newProfile)
		}
	}
	profileGroupPtr.ProfileIDsJoined = strings.Join(profileGroupPtr.ProfileIDs, ",")

	return UpdateProfileGroup(groupID, *profileGroupPtr)
}

func RemoveProfilesFromGroup(groupID string, profileIDs []string) (*entities.ProfileGroup, error) {
	profileGroupPtr, err := GetProfileGroup(groupID)
	if err != nil {
		return nil, err
	}

	newProfiles := []*entities.Profile{}
	newProfileIDs := []string{}
	for _, profile := range profileGroupPtr.Profiles {
		if util.InSlice(profileIDs, profile.ID) {
			RemoveGroupIDFromProfile(profile.ID, groupID)
		} else {
			newProfiles = append(newProfiles, profile)
			newProfileIDs = append(newProfileIDs, profile.ID)
		}
	}
	profileGroupPtr.ProfileIDs = newProfileIDs
	profileGroupPtr.ProfileIDsJoined = strings.Join(newProfileIDs, ",")
	profileGroupPtr.Profiles = newProfiles

	return UpdateProfileGroup(groupID, *profileGroupPtr)
}
