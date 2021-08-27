package stores

import "backend.juicedbot.io/juiced.infrastructure/entities"

type ProfileStore struct {
	Profiles []entities.Profile
}

var profileStore ProfileStore

func (store *ProfileStore) Init() error {
	// TODO
	return nil
}
