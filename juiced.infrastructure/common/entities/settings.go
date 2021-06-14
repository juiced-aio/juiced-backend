package entities

import "encoding/json"

// Settings is a class that holds details about a user's settings
type Settings struct {
	ID                int    `json:"id" db:"id"`
	DiscordWebhook    string `json:"discordWebhook" db:"discordWebhook"`
	TwoCaptchaAPIKey  string `json:"twoCaptchaAPIKey" db:"twoCaptchaAPIKey"`
	AntiCaptchaAPIKey string `json:"antiCaptchaAPIKey" db:"antiCaptchaAPIKey"`
	CapMonsterAPIKey  string `json:"capMonsterAPIKey" db:"capMonsterAPIKey"`
}

// ParseSettings returns a Settings object parsed from a JSON bytes array
func ParseSettings(settings *Settings, data []byte) error {
	err := json.Unmarshal(data, &settings)
	return err
}
