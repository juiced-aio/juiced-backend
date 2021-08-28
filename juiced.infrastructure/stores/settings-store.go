package stores

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

func AddAccount(account entities.Account) error {
	err := database.AddAccount(account)
	if err != nil {
		return err
	}
	settingsStore.Settings.Accounts = append(settingsStore.Settings.Accounts, &account)
	return nil
}

func UpdateAccount(accountID string, newAccount entities.Account) error {
	err := database.UpdateAccount(accountID, newAccount)
	if err != nil {
		return err
	}
	for _, account := range settingsStore.Settings.Accounts {
		if account.ID == accountID {
			account.Email = newAccount.Email
			account.Password = newAccount.Password
			account.Retailer = newAccount.Retailer
		} else {
			settingsStore.Settings.Accounts = append(settingsStore.Settings.Accounts, account)
		}
	}
	return nil
}

func RemoveAccount(accountID string) error {
	err := database.RemoveAccount(accountID)
	if err != nil {
		return err
	}
	accounts := []*entities.Account{}
	for _, account := range settingsStore.Settings.Accounts {
		if account.ID != accountID {
			accounts = append(accounts, account)
		}
	}
	settingsStore.Settings.Accounts = accounts
	return nil
}
