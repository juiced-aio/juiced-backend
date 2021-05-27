package commands

import (
	"errors"

	"backend.juicedbot.io/m/v2/juiced.infrastructure/common"
	"backend.juicedbot.io/m/v2/juiced.infrastructure/common/entities"
	"backend.juicedbot.io/m/v2/juiced.infrastructure/queries"

	_ "github.com/jmoiron/sqlx"
	_ "github.com/mattn/go-sqlite3"
)

// SetUserInfo set the user's UserInfo in the SQL Database
func SetUserInfo(userInfo entities.UserInfo) error {
	database := common.GetDatabase()
	if database == nil {
		return errors.New("database not initialized")
	}

	numRows, existingUserInfo, err := queries.GetUserInfo()
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
		WHERE ID = 0
		`

	if numRows > 0 && (existingUserInfo.Email == "" || err != nil) {
		_, err := database.Exec(userInfoInsert,
			userInfo.Email, userInfo.LicenseKey, userInfo.DeviceName,
			userInfo.DiscordID, userInfo.DiscordUsername, userInfo.DiscordAvatarURL,
			userInfo.ActivationToken, userInfo.RefreshToken, userInfo.ExpiresAt)
		if err == nil {
			return nil
		}
	}

	userInfoInsert = `
		INSERT INTO userInfo (
			ID, email, licenseKey, deviceName,
			discordID, discordUsername, discordAvatarURL, 
			activationToken, refreshToken, expiresAt
		) VALUES (0, ?, ?, ?, ?, ?, ?, ?, ?, ?)`

	_, err = database.Exec(userInfoInsert,
		userInfo.Email, userInfo.LicenseKey, userInfo.DeviceName,
		userInfo.DiscordID, userInfo.DiscordUsername, userInfo.DiscordAvatarURL,
		userInfo.ActivationToken, userInfo.RefreshToken, userInfo.ExpiresAt)

	return err
}
