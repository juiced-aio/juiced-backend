package entities

type Channel = string

const (
	ALPHA   Channel = "Alpha"
	BETA    Channel = "Beta"
	Release Channel = "Release"
)

type UserInfo struct {
	Email            string `json:"email" db:"email"`
	UserID           string `json:"userID" db:"userID"`
	LicenseKey       string `json:"licenseKey" db:"licenseKey"`
	DeviceName       string `json:"deviceName" db:"deviceName"`
	DiscordID        string `json:"discordID" db:"discordID"`
	DiscordUsername  string `json:"discordUsername" db:"discordUsername"`
	DiscordAvatarURL string `json:"discordAvatarURL" db:"discordAvatarURL"`
	ActivationToken  string `json:"activationToken" db:"activationToken"`
	RefreshToken     string `json:"refreshToken" db:"refreshToken"`
	ExpiresAt        int64  `json:"expiresAt" db:"expiresAt"`

	Channel         Channel `json:"channel" db:"channel"`
	BackendVersion  string  `json:"backendVersion" db:"backendVersion"`
	EncryptionKey   string  `json:"encryptionKey"`
	FrontendVersion string  `json:"frontendVersion"`
}
