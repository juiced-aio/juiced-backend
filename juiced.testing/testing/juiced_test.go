package testing

import (
	"log"
	"math/rand"
	"os"
	"testing"
	"time"

	api "backend.juicedbot.io/juiced.api"
	"backend.juicedbot.io/juiced.infrastructure/common"
	"backend.juicedbot.io/juiced.infrastructure/common/captcha"
	"backend.juicedbot.io/juiced.infrastructure/common/entities"
	"backend.juicedbot.io/juiced.infrastructure/common/enums"
	"backend.juicedbot.io/juiced.infrastructure/common/events"
	"backend.juicedbot.io/juiced.infrastructure/common/stores"
	"backend.juicedbot.io/juiced.infrastructure/queries"
	"backend.juicedbot.io/juiced.sitescripts/util"
)

func TestMain(m *testing.M) {
	events.InitEventBus()
	eventBus := events.GetEventBus()
	err := common.InitDatabase()
	if err != nil {
		log.Println(err)
		return
	}
	rand.Seed(time.Now().UnixNano())
	go stores.InitTaskStore(eventBus)
	stores.InitMonitorStore(eventBus)
	captcha.InitCaptchaStore(eventBus)
	err = captcha.InitAycd()
	if err == nil {
		log.Println("Initialized AYCD.")
		settings, err := queries.GetSettings()
		if err == nil {
			if settings.AYCDAccessToken != "" && settings.AYCDAPIKey != "" {
				err = captcha.ConnectToAycd(settings.AYCDAccessToken, settings.AYCDAPIKey)
				if err != nil {
					log.Println("Error connecting to AYCD: " + err.Error())
					// TODO @silent: Handle
				} else {
					log.Println("Connected to AYCD.")
				}
			}
		}
	} else {
		log.Println("Error initializing AYCD: " + err.Error())
		// TODO @silent: Handle
	}
	taskStore = stores.GetTaskStore()
	monitorStore = stores.GetMonitorStore()
	go util.DiscordWebhookQueue()
	go api.StartServer()
	os.Exit(m.Run())
}

func TestBestbuy(t *testing.T) {
	MainTask.BestbuyTaskInfo = &entities.BestbuyTaskInfo{
		TaskID:      MainTaskID,
		TaskGroupID: MainTaskGroupID,
		Email:       "",
		Password:    "",
		LocationID:  "",
		TaskType:    enums.TaskTypeGuest,
	}
	monitorID := "cc4883da-e7d4-4a82-8b5a-83f78358fecf"
	MainTaskGroup.BestbuyMonitorInfo = &entities.BestbuyMonitorInfo{
		ID:          monitorID,
		TaskGroupID: MainTaskGroupID,
		Monitors: []entities.BestbuySingleMonitorInfo{{
			MonitorID:   monitorID,
			TaskGroupID: MainTaskGroupID,
			SKU:         "5901353",
			MaxPrice:    -1,
		}},
	}
	MainTaskGroup.MonitorRetailer = enums.BestBuy
	MainTask.TaskRetailer = enums.BestBuy
	MainTaskGroup.MonitorStatus = enums.MonitorIdle
	TestDriver(MainTask, *MainProfile, *MainTaskGroup, *MainProxyGroup)
	select {}
}
