package stores

import "backend.juicedbot.io/juiced.infrastructure/entities"

type ProfileStore struct {
	Profiles map[string]*entities.Profile
}

var profileStore ProfileStore

func (store *ProfileStore) Init() error {
	// TODO
	return nil
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
