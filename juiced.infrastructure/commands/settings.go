package commands

import (
	"errors"

	"backend.juicedbot.io/juiced.infrastructure/common"
	"backend.juicedbot.io/juiced.infrastructure/common/entities"
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

	statement, err := database.Preparex(`INSERT INTO settings (id, successDiscordWebhook, failureDiscordWebhook, twoCaptchaAPIKey, antiCaptchaAPIKey, capMonsterAPIKey, aycdAccessToken, aycdAPIKey, darkMode) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`)
	if err != nil {
		return settings, err
	}
	_, err = statement.Exec(0, settings.SuccessDiscordWebhook, settings.FailureDiscordWebhook, settings.TwoCaptchaAPIKey, settings.AntiCaptchaAPIKey, settings.CapMonsterAPIKey, settings.AYCDAccessToken, settings.AYCDAPIKey, settings.DarkMode)
	if err != nil {
		return settings, err
	}

	return queries.GetSettings()
}
