package stores

import "backend.juicedbot.io/juiced.infrastructure/entities"

type ProfileGroupStore struct {
	ProfileGroups []entities.ProfileGroup
}

var profileGroupStore ProfileGroupStore

func (store *ProfileGroupStore) Init() error {
	var err error
	store.ProfileGroups, err = database.GetProfileGroups()
	return err
}
