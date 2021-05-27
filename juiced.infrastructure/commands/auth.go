package commands

import (
	"backend.juicedbot.io/m/v2/juiced.infrastructure/common"
	"backend.juicedbot.io/m/v2/juiced.infrastructure/common/entities"
	"backend.juicedbot.io/m/v2/juiced.infrastructure/queries"

	_ "github.com/jmoiron/sqlx"
	_ "github.com/mattn/go-sqlite3"
)

// SetUserInfo set the user's UserInfo in the SQL Database
func SetUserInfo(userInfo entities.UserInfo) error {
	database, err := common.ConnectToDatabase()
	if err != nil {
		return err
	}

	existingUserInfo, err := queries.GetUserInfo()
	userInfoInsert := `
		UPDATE userInfo
		SET email = ?,
			licenseKey = ?,
			deviceName = ?,
			discordID = ?,
			discordUsername = ?,
			discordAvatarURL = ?, 
			activationToken = ?,
			refreshToken = ?,
			expiresAt = ?
		WHERE email = ?
		`

	if existingUserInfo.Email == "" || err != nil {
		_, err = database.Exec(userInfoInsert,
			userInfo.Email, userInfo.LicenseKey, userInfo.DeviceName,
			userInfo.DiscordID, userInfo.DiscordUsername, userInfo.DiscordAvatarURL,
			userInfo.ActivationToken, userInfo.RefreshToken, userInfo.ExpiresAt,
			existingUserInfo.Email)
		if err == nil {
			return nil
		}
	}

	userInfoInsert = `
		INSERT INTO userInfo (
			email, licenseKey, deviceName,
			discordID, discordUsername, discordAvatarURL, 
			activationToken, refreshToken, expiresAt
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`

	_, err = database.Exec(userInfoInsert,
		userInfo.Email, userInfo.LicenseKey, userInfo.DeviceName,
		userInfo.DiscordID, userInfo.DiscordUsername, userInfo.DiscordAvatarURL,
		userInfo.ActivationToken, userInfo.RefreshToken, userInfo.ExpiresAt)

	return err
}
