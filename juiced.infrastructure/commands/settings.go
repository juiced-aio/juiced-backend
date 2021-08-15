package commands

import (
	"errors"

	"backend.juicedbot.io/juiced.infrastructure/common"
	"backend.juicedbot.io/juiced.infrastructure/common/entities"
	"backend.juicedbot.io/juiced.infrastructure/common/enums"
	"backend.juicedbot.io/juiced.infrastructure/queries"
	_ "github.com/jmoiron/sqlx"
	_ "github.com/mattn/go-sqlite3"
)

// UpdateSettings updates the Settings object in the database
func UpdateSettings(settings entities.Settings) (entities.Settings, error) {
	database := common.GetDatabase()
	if database == nil {
		return settings, errors.New("database not initialized")
	}

	_, err := database.Exec("DELETE FROM settings")
	if err != nil {
		return settings, err
	}

	statement, err := database.Preparex(`INSERT INTO settings (id, successDiscordWebhook, failureDiscordWebhook, twoCaptchaAPIKey, antiCaptchaAPIKey, capMonsterAPIKey, aycdAccessToken, aycdAPIKey, darkMode, useAnimations) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`)
	if err != nil {
		return settings, err
	}
	_, err = statement.Exec(0, settings.SuccessDiscordWebhook, settings.FailureDiscordWebhook, settings.TwoCaptchaAPIKey, settings.AntiCaptchaAPIKey, settings.CapMonsterAPIKey, settings.AYCDAccessToken, settings.AYCDAPIKey, settings.DarkMode, settings.UseAnimations)
	if err != nil {
		return settings, err
	}

	return queries.GetSettings()
}

// AddAccount adds an Account object to the database
func AddAccount(account entities.Account) error {
	database := common.GetDatabase()
	if database == nil {
		return errors.New("database not initialized")
	}

	encryptedEmail, err := common.Aes256Encrypt(account.Email, enums.UserKey)
	if err != nil {
		return err
	}

	encryptedPassword, err := common.Aes256Encrypt(account.Password, enums.UserKey)
	if err != nil {
		return err
	}

	statement, err := database.Preparex(`INSERT INTO accounts (ID, retailer, email, password, creationDate) VALUES (?, ?, ?, ?, ?)`)
	if err != nil {
		return err
	}

	_, err = statement.Exec(account.ID, account.Retailer, encryptedEmail, encryptedPassword, account.CreationDate)

	return err
}

// UpdateAccount updates an Account object in the database
func UpdateAccount(ID string, newAccount entities.Account) (entities.Account, error) {
	account := entities.Account{}
	_, err := RemoveAccount(ID)
	if err != nil {
		return account, err
	}

	err = AddAccount(newAccount)

	return account, err
}

// RemoveAccount removes an Account object from the database
func RemoveAccount(ID string) (entities.Account, error) {
	account := entities.Account{}
	database := common.GetDatabase()
	if database == nil {
		return account, errors.New("database not initialized")
	}
	account, err := queries.GetAccount(ID)
	if err != nil {
		return account, err
	}

	statement, err := database.Preparex(`DELETE FROM accounts WHERE ID = @p1`)
	if err != nil {
		return account, err
	}
	_, err = statement.Exec(ID)

	return account, err
}
