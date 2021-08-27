package stores

import (
	"backend.juicedbot.io/juiced.infrastructure/database"
	"backend.juicedbot.io/juiced.infrastructure/entities"
)

type UserInfoStore struct {
	UserInfo entities.UserInfo
}

var userInfoStore UserInfoStore

func (store *UserInfoStore) Init() error {
	var err error
	store.UserInfo, err = database.GetUserInfo()
	return err
}

func GetUserInfo() entities.UserInfo {
	return userInfoStore.UserInfo
}

func SetUserInfo(userInfo entities.UserInfo) error {
	err := database.SetUserInfo(userInfo)
	if err != nil {
		return err
	}
	userInfoStore.UserInfo = userInfo
	return nil
}
