package database

import (
	"backend.juicedbot.io/juiced.infrastructure/entities"
)

func GetUserInfo() (entities.UserInfo, error) {
	userInfo := entities.UserInfo{}
	numRows := 0

	if database == nil {
		return userInfo, &DatabaseNotInitializedError{}
	}

	rows, err := database.Queryx("SELECT * FROM userInfo")
	if err != nil {
		return userInfo, err
	}
	defer rows.Close()
	for rows.Next() {
		numRows += 1
		err = rows.StructScan(&userInfo)
		if err != nil {
			return userInfo, err
		}
	}

	return userInfo, err
}

// SetUserInfo set the user's UserInfo in the SQL Database
func SetUserInfo(userInfo entities.UserInfo) error {
	if database == nil {
		return &DatabaseNotInitializedError{}
	}

	_, err := database.Exec("DELETE FROM userInfo")
	if err != nil {
		return err
	}

	userInfoInsert := `
		INSERT INTO userInfo (
			ID, email, licenseKey, deviceName, userVer,
			discordID, discordUsername, discordAvatarURL, 
			activationToken, refreshToken, expiresAt
		) VALUES (0, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`

	_, err = database.Exec(userInfoInsert,
		userInfo.Email, userInfo.LicenseKey, userInfo.DeviceName, userInfo.UserVer,
		userInfo.DiscordID, userInfo.DiscordUsername, userInfo.DiscordAvatarURL,
		userInfo.ActivationToken, userInfo.RefreshToken, userInfo.ExpiresAt)
	return err
}
