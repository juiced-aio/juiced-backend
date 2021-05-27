package queries

import (
	"errors"

	"backend.juicedbot.io/m/v2/juiced.infrastructure/common"
	"backend.juicedbot.io/m/v2/juiced.infrastructure/common/entities"
)

// GetUserInfo returns the user's UserInfo from the SQL database
func GetUserInfo() (int, entities.UserInfo, error) {
	userInfo := entities.UserInfo{}
	numRows := 0

	database := common.GetDatabase()
	if database == nil {
		return numRows, userInfo, errors.New("database not initialized")
	}

	rows, err := database.Queryx("SELECT * FROM userInfo")
	if err != nil {
		return numRows, userInfo, err
	}
	defer rows.Close()
	for rows.Next() {
		numRows += 1
		err = rows.StructScan(&userInfo)
		if err != nil {
			return numRows, userInfo, err
		}
	}
	return numRows, userInfo, err
}
