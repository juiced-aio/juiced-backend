package entities

type UserInfo struct {
	ID               string `db:"ID"`
	Email            string `json:"cABQyJDDb6" db:"email"`
	LicenseKey       string `json:"q5izDieCbb" db:"licenseKey"`
	DeviceName       string `json:"deviceName" db:"deviceName"`
	UserVer          string `json:"userVer" db:"userVer"`
	DiscordID        string `json:"sQzlETuNin" db:"discordID"`
	DiscordUsername  string `json:"1x16bw24sz" db:"discordUsername"`
	DiscordAvatarURL string `json:"9WMNTAhnxb" db:"discordAvatarURL"`
	ActivationToken  string `json:"lUiFoCyvqa" db:"activationToken"`
	RefreshToken     string `json:"qM3DKGrRJE" db:"refreshToken"`
	ExpiresAt        int64  `json:"uYIvVc1ojh" db:"expiresAt"`
}
