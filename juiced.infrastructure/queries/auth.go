package queries

import (
	"backend.juicedbot.io/juiced.infrastructure/common"
	"backend.juicedbot.io/juiced.infrastructure/common/entities"
)

// GetUserInfo returns the user's UserInfo from the SQL database
func GetUserInfo() (entities.UserInfo, error) {
	userInfo := entities.UserInfo{}

	database, err := common.ConnectToDatabase()
	if err != nil {
		return userInfo, err
	}

	rows, err := database.Queryx("SELECT * FROM userInfo")
	if err != nil {
		return userInfo, err
	}
	defer rows.Close()
	for rows.Next() {
		err = rows.StructScan(&userInfo)
		if err != nil {
			return userInfo, err
		}
	}
	return userInfo, err
}
