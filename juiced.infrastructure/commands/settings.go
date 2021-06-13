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
func UpdateSettings(newSettings entities.Settings) (entities.Settings, error) {
	settings := entities.Settings{}
	database := common.GetDatabase()
	if database == nil {
		return settings, errors.New("database not initialized")
	}

	_, err := database.Exec("DELETE FROM tasks")
	if err != nil {
		return settings, err
	}

	statement, err := database.Preparex(`INSERT INTO settings (id, discordWebhook, twoCaptchaAPIKey, antiCaptchaAPIKey, capMonsterAPIKey) VALUES (?, ?, ?, ?, ?)`)
	if err != nil {
		return settings, err
	}
	_, err = statement.Exec(0, settings.DiscordWebhook, settings.TwoCaptchaAPIKey, settings.AntiCaptchaAPIKey, settings.CapMonsterAPIKey)
	if err != nil {
		return settings, err
	}

	return queries.GetSettings()
}
