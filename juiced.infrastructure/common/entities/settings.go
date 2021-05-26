package entities

import "encoding/json"

// Settings is a class that holds details about a user's settings
type Settings struct {
	ID                int    `json:"id"`
	DiscordWebhook    string `json:"discordWebhook"`
	TwoCaptchaAPIKey  string `json:"twoCaptchaAPIKey"`
	AntiCaptchaAPIKey string `json:"antiCaptchaAPIKey"`
	CapMonsterAPIKey  string `json:"capMonsterAPIKey"`
}

// ParseSettings returns a Settings object parsed from a JSON bytes array
func ParseSettings(settings *Settings, data []byte) error {
	err := json.Unmarshal(data, &settings)
	return err
}
