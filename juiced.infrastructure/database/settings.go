package database

import (
	"backend.juicedbot.io/juiced.infrastructure/entities"
)

func GetSettings() (entities.Settings, error) {
	settings := entities.Settings{}
	if database == nil {
		return settings, &DatabaseNotInitializedError{}
	}

	rows, err := database.Queryx("SELECT * FROM settings")
	if err != nil {
		return settings, err
	}

	defer rows.Close()
	for rows.Next() {
		err = rows.StructScan(&settings)
		if err != nil {
			return settings, err
		}
	}
	return settings, err
}

func UpdateSettings(settings entities.Settings) error {
	if database == nil {
		return &DatabaseNotInitializedError{}
	}

	_, err := database.Exec("DELETE FROM settings")
	if err != nil {
		return err
	}

	statement, err := database.Preparex(`INSERT INTO settings (id, successDiscordWebhook, failureDiscordWebhook, twoCaptchaAPIKey, antiCaptchaAPIKey, capMonsterAPIKey, aycdAccessToken, aycdAPIKey, darkMode, useAnimations) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`)
	if err != nil {
		return err
	}
	_, err = statement.Exec(0, settings.SuccessDiscordWebhook, settings.FailureDiscordWebhook, settings.TwoCaptchaAPIKey, settings.AntiCaptchaAPIKey, settings.CapMonsterAPIKey, settings.AYCDAccessToken, settings.AYCDAPIKey, settings.DarkMode, settings.UseAnimations)

	return err
}
