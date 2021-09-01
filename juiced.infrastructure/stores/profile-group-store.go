package stores

import (
	"fmt"
	"strings"
	"time"

	"backend.juicedbot.io/juiced.infrastructure/database"
	"backend.juicedbot.io/juiced.infrastructure/entities"
	"github.com/google/uuid"
)

type ProfileGroupStore struct {
	ProfileGroups map[string]*entities.ProfileGroup
}

var profileGroupStore ProfileGroupStore

func (store *ProfileGroupStore) Init() error {
	profileGroups, err := database.GetAllProfileGroups()
	if err != nil {
		return err
	}

	for _, profileGroup := range profileGroups {
		profileGroup.Profiles = GetProfiles(profileGroup.ProfileIDs)
		store.ProfileGroups[profileGroup.GroupID] = &profileGroup
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

	delete(profileGroupStore.ProfileGroups, groupID)
	return *profileGroup, database.RemoveProfileGroup(groupID)
}

func CloneProfileGroup(groupID string) (*entities.ProfileGroup, error) {
	profileGroupPtr, err := GetProfileGroup(groupID)
	if err != nil {
		return nil, err
	}

	newProfileGroup := *profileGroupPtr
	newProfileGroup.GroupID = ""
	newProfileGroup.CreationDate = 0

	newProfileGroupPtr, err := CreateProfileGroup(newProfileGroup)
	if err != nil {
		return nil, err
	}

	for _, profileID := range newProfileGroupPtr.ProfileIDs {
		profile, err := GetProfile(profileID)
		if err == nil {
			profile.ProfileGroupIDs = append(profile.ProfileGroupIDs, newProfileGroupPtr.GroupID)
			profile.ProfileGroupIDsJoined = strings.Join(profile.ProfileGroupIDs, ",")
		}
	}
	return newProfileGroupPtr, nil
}
