package stores

import (
	"fmt"

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

func GetProfileGroup(groupID string) (*entities.ProfileGroup, error) {
	profileGroup, ok := profileGroupStore.ProfileGroups[groupID]
	if !ok {
		return nil, &ProfileGroupNotFoundError{groupID}
	}

	return profileGroup, nil
}

func GetAllProfileGroups() []*entities.ProfileGroup {
	profileGroups := []*entities.ProfileGroup{}

	for _, profileGroup := range profileGroupStore.ProfileGroups {
		profileGroups = append(profileGroups, profileGroup)
	}

	return profileGroups
}

func CreateProfileGroup(profileGroup entities.ProfileGroup) (*entities.ProfileGroup, error) {
	profileGroup.GroupID = uuid.New().String()
	profileGroup.Profiles = GetProfiles(profileGroup.ProfileIDs)

	err := database.CreateProfileGroup(profileGroup)

	profileGroupPtr := &profileGroup
	profileGroupStore.ProfileGroups[profileGroup.GroupID] = profileGroupPtr

	return profileGroupPtr, err
}

func RemoveProfileGroup(groupID string) (entities.ProfileGroup, error) {
	profileGroup, err := GetProfileGroup(groupID)
	if err != nil {
		return entities.ProfileGroup{}, err
	}

	delete(profileGroupStore.ProfileGroups, groupID)
	return *profileGroup, nil
}
