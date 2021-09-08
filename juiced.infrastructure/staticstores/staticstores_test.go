package staticstores_test

import (
	"log"
	"os"
	"testing"

	"backend.juicedbot.io/juiced.infrastructure/database"
	"backend.juicedbot.io/juiced.infrastructure/staticstores"
)

func TestMain(m *testing.M) {
	err := database.InitTestDatabase()
	if err == nil {
		err = staticstores.InitStores()
		if err == nil {
			os.Exit(m.Run())
		} else {
			log.Println("Error initializing static stores: " + err.Error())
		}
	} else {
		log.Println("Error initializing test database: " + err.Error())
	}
	os.Exit(1)
}

func TestCheckoutsStore(t *testing.T) {
	t.Run("GetCheckouts", func(t *testing.T) {

	})

	t.Run("CreateCheckouts", func(t *testing.T) {

	})
}

func TestSettingsStore(t *testing.T) {
	t.Run("GetSettings", func(t *testing.T) {

	})

	t.Run("UpdateSettings", func(t *testing.T) {

	})
}

func TestUserInfoStore(t *testing.T) {
	t.Run("GetUserInfo", func(t *testing.T) {

	})

	t.Run("SetUserInfo", func(t *testing.T) {

	})
}
