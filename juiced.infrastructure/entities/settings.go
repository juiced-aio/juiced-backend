package entities

// Settings is a class that holds details about a user's settings
type Settings struct {
	ID                    int    `json:"ID" db:"ID"`
	SuccessDiscordWebhook string `json:"successDiscordWebhook" db:"successDiscordWebhook"`
	FailureDiscordWebhook string `json:"failureDiscordWebhook" db:"failureDiscordWebhook"`
	TwoCaptchaAPIKey      string `json:"twoCaptchaAPIKey" db:"twoCaptchaAPIKey"`
	AntiCaptchaAPIKey     string `json:"antiCaptchaAPIKey" db:"antiCaptchaAPIKey"`
	CapMonsterAPIKey      string `json:"capMonsterAPIKey" db:"capMonsterAPIKey"`
	AYCDAccessToken       string `json:"aycdAccessToken" db:"aycdAccessToken"`
	AYCDAPIKey            string `json:"aycdAPIKey" db:"aycdAPIKey"`
	DarkModeUpdate        bool   `json:"darkModeUpdate"`
	DarkMode              bool   `json:"darkMode" db:"darkMode"`
	UseAnimationsUpdate   bool   `json:"useAnimationsUpdate"`
	UseAnimations         bool   `json:"useAnimations" db:"useAnimations"`
}
