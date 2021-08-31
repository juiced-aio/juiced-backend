package staticstores

import (
	"backend.juicedbot.io/juiced.infrastructure/database"
	"backend.juicedbot.io/juiced.infrastructure/entities"
)

type SettingsStore struct {
	Settings entities.Settings
}

var settingsStore SettingsStore

func (store *SettingsStore) Init() error {
	var err error
	store.Settings, err = database.GetSettings()
	return err
}

func GetSettings() entities.Settings {
	return settingsStore.Settings
}

func UpdateSettings(settings entities.Settings) error {
	err := database.UpdateSettings(settings)
	if err != nil {
		return err
	}
	settingsStore.Settings = settings
	return nil
}
