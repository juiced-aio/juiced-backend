package entities

import (
	"encoding/json"

	"backend.juicedbot.io/juiced.infrastructure/common/enums"
)

// Settings is a class that holds details about a user's settings
type Settings struct {
	ID                    int       `json:"id" db:"id"`
	SuccessDiscordWebhook string    `json:"successDiscordWebhook" db:"successDiscordWebhook"`
	FailureDiscordWebhook string    `json:"failureDiscordWebhook" db:"failureDiscordWebhook"`
	TwoCaptchaAPIKey      string    `json:"twoCaptchaAPIKey" db:"twoCaptchaAPIKey"`
	AntiCaptchaAPIKey     string    `json:"antiCaptchaAPIKey" db:"antiCaptchaAPIKey"`
	CapMonsterAPIKey      string    `json:"capMonsterAPIKey" db:"capMonsterAPIKey"`
	AYCDAccessToken       string    `json:"aycdAccessToken" db:"aycdAccessToken"`
	AYCDAPIKey            string    `json:"aycdAPIKey" db:"aycdAPIKey"`
	DarkModeUpdate        bool      `json:"darkModeUpdate"`
	DarkMode              bool      `json:"darkMode" db:"darkMode"`
	UseAnimationsUpdate   bool      `json:"useAnimationsUpdate"`
	UseAnimations         bool      `json:"useAnimations" db:"useAnimations"`
	Accounts              []Account `json:"accounts"`
}

type Account struct {
	ID           string         `json:"ID" db:"ID"`
	Retailer     enums.Retailer `json:"retailer" db:"retailer"`
	Email        string         `json:"email" db:"email"`
	Password     string         `json:"password" db:"password"`
	CreationDate int64          `json:"creationDate" db:"creationDate"`
}

// ParseSettings returns a Settings object parsed from a JSON bytes array
func ParseSettings(settings *Settings, data []byte) error {
	err := json.Unmarshal(data, &settings)
	return err
}
