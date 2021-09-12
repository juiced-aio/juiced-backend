package controller_test

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	api "backend.juicedbot.io/juiced.api"
	"backend.juicedbot.io/juiced.infrastructure/database"
	"backend.juicedbot.io/juiced.infrastructure/entities"
	"backend.juicedbot.io/juiced.infrastructure/staticstores"
	"backend.juicedbot.io/juiced.infrastructure/stores"
	"backend.juicedbot.io/juiced.infrastructure/util"
	"github.com/gofiber/fiber/v2"
)

func TestMain(m *testing.M) {
	err := database.InitTestDatabase()
	if err == nil {
		err = staticstores.InitStores()
		if err == nil {
			err = stores.InitStores()
			if err == nil {
				api.StartServer()
				for api.GetApp().Server() == nil {
					time.Sleep(util.WAIT_TIME)
				}
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

const BASE_ENDPOINT = "http://127.0.0.1:10000/api/v1"

//////////////
// SETTINGS //
//////////////

const SETTINGS_ENDPOINT = BASE_ENDPOINT + "/settings"

func TestSettingsEndpoints(t *testing.T) {
	t.Run("GetSettings endpoint", func(t *testing.T) {
		t.Run("GetSettings endpoint returns default settings", func(t *testing.T) {
			req := httptest.NewRequest("GET", SETTINGS_ENDPOINT, nil)
			resp, err := api.GetApp().Test(req)
			if err != nil {
				t.Errorf("app.Test failed: %v", err)
				t.FailNow()
			}

			if resp.StatusCode != fiber.StatusOK {
				t.Errorf("Request failed with status code %d", resp.StatusCode)
				t.FailNow()
			}

			defer resp.Body.Close()
			body, err := ioutil.ReadAll(resp.Body)
			if err != nil {
				t.Errorf("ioutil.Readall failed: %v", err)
				t.FailNow()
			}

			response := entities.Settings{}
			err = json.Unmarshal(body, &response)
			if err != nil {
				t.Errorf("json.Unmarshal failed: %v (body: %s)", err, string(body))
				t.FailNow()
			}

			defaultSettings := entities.Settings{}
			if response != defaultSettings {
				t.Errorf("response != defaultSettings (response: %s)", response)
				t.FailNow()
			}
		})

		t.Run("GetSettings endpoint returns updated settings", func(t *testing.T) {
			newSettings := entities.Settings{
				SuccessDiscordWebhook: "TEST_SuccessDiscordWebhook",
				FailureDiscordWebhook: "TEST_FailureDiscordWebhook",
				TwoCaptchaAPIKey:      "TEST_TwoCaptchaAPIKey",
				AntiCaptchaAPIKey:     "TEST_AntiCaptchaAPIKey",
				CapMonsterAPIKey:      "TEST_CapMonsterAPIKey",
				AYCDAccessToken:       "TEST_AYCDAccessToken",
				AYCDAPIKey:            "TEST_AYCDAPIKey",
				DarkMode:              true,
				UseAnimations:         true,
			}
			err := staticstores.UpdateSettings(newSettings)
			if err != nil {
				t.Errorf("Failed to call staticstores.UpdateSettings: %v", err)
				t.FailNow()
			}

			req := httptest.NewRequest("GET", SETTINGS_ENDPOINT, nil)
			resp, err := api.GetApp().Test(req)
			if err != nil {
				t.Errorf("app.Test failed: %v", err)
				t.FailNow()
			}

			if resp.StatusCode != fiber.StatusOK {
				t.Errorf("Request failed with status code %d", resp.StatusCode)
				t.FailNow()
			}

			defer resp.Body.Close()
			body, err := ioutil.ReadAll(resp.Body)
			if err != nil {
				t.Errorf("ioutil.Readall failed: %v", err)
				t.FailNow()
			}

			response := entities.Settings{}
			err = json.Unmarshal(body, &response)
			if err != nil {
				t.Errorf("json.Unmarshal failed: %v (body: %s)", err, string(body))
				t.FailNow()
			}

			if response != newSettings {
				t.Errorf("response != newSettings (response: %s)", response)
				t.FailNow()
			}
		})
	})
	t.Run("UpdateSettings endpoint", func(t *testing.T) {

	})
}

//////////////
// ACCOUNTS //
//////////////

func TestAccountEndpoints(t *testing.T) {

}

func TestProxyGroupEndpoints(t *testing.T) {

}

func TestProfileEndpoints(t *testing.T) {

}

func TestProfileGroupEndpoints(t *testing.T) {

}

func TestTaskEndpoints(t *testing.T) {

}

func TestTaskGroupEndpoints(t *testing.T) {

}

func TestCheckoutEndpoints(t *testing.T) {

}

func TestMiscEndpoints(t *testing.T) {

}
