package stores_test

import (
	"log"
	"os"
	"testing"

	"backend.juicedbot.io/juiced.infrastructure/database"
	"backend.juicedbot.io/juiced.infrastructure/staticstores"
	"backend.juicedbot.io/juiced.infrastructure/stores"
)

func TestMain(m *testing.M) {
	err := database.InitTestDatabase()
	if err == nil {
		err = staticstores.InitStores()
		if err == nil {
			err = stores.InitStores()
			if err == nil {
				os.Exit(m.Run())
			} else {
				log.Println("Error initializing stores: " + err.Error())
			}
		} else {
			log.Println("Error initializing static stores: " + err.Error())
		}
	} else {
		log.Println("Error initializing test database: " + err.Error())
	}
	os.Exit(1)
}