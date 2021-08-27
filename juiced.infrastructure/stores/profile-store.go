package stores

import (
	"fmt"

	"backend.juicedbot.io/juiced.infrastructure/database"
	"backend.juicedbot.io/juiced.infrastructure/entities"
	"github.com/google/uuid"
)

type ProfileStore struct {
	Profiles map[string]*entities.Profile
}

var profileStore ProfileStore

func (store *ProfileStore) Init() error {
	profiles, err := database.GetAllProfiles()
	if err != nil {
		return err
	}

	for _, profile := range profiles {
		store.Profiles[profile.ID] = &profile
	}

	return nil
}

type ProfileNotFoundError struct {
	ID string
}

func (e *ProfileNotFoundError) Error() string {
	return fmt.Sprintf("Profile with ID %s not found", e.ID)
}

func GetAllProfiles() []*entities.Profile {
	profiles := []*entities.Profile{}
	for _, profile := range profileStore.Profiles {
		profiles = append(profiles, profile)
	}

	return profiles
}

func GetProfiles(profileIDs []string) []*entities.Profile {
	profiles := []*entities.Profile{}
	for _, profileID := range profileIDs {
		if profile, ok := profileStore.Profiles[profileID]; ok {
			profiles = append(profiles, profile)
		}
	}

	return profiles
}

func GetProfile(profileID string) (*entities.Profile, error) {
	profile, ok := profileStore.Profiles[profileID]
	if !ok {
		return nil, &ProfileNotFoundError{profileID}
	}

	return profile, nil
}

func CreateProfile(profile entities.Profile) (*entities.Profile, error) {
	profile.ID = uuid.New().String()

	err := database.CreateProfile(profile)

	profilePtr := &profile
	profileStore.Profiles[profile.ID] = profilePtr

	return profilePtr, err
}
