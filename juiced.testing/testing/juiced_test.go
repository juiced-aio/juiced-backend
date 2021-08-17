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
	stores.InitProxyStore()
	stores.GetProxyStore().AddProxyGroup(MainProxyGroup)
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

func TestAmazon(t *testing.T) {
	MainTask.AmazonTaskInfo = &entities.AmazonTaskInfo{
		TaskID:      MainTaskID,
		TaskGroupID: MainTaskGroupID,
		Email:       "example@gmail.com",
		Password:    "examplepass",
		LoginType:   enums.LoginTypeBROWSER,
	}

	MainTaskGroup.AmazonMonitorInfo = &entities.AmazonMonitorInfo{
		ID:          MainMonitorID,
		TaskGroupID: MainTaskGroupID,
		Monitors: []entities.AmazonSingleMonitorInfo{{
			MonitorID:   MainMonitorID,
			TaskGroupID: MainTaskGroupID,
			MonitorType: enums.FastSKUMonitor,
			ASIN:        "B07R11LGDQ",
			OFID:        "vrQLfZnyPNErArVdtORZv59PCLQnuVKvA6Ab5lDv5qpP2RH04BS%2FfoGwiZC%2Fi3933yvoIJsw%2F%2BAXGacjPAVPQiyS1Pe3hda3UagfQxm51u%2B4iIfQ5zx3p7zg%2BAsf4MpT%2BB3ZH0yGBhq8S4R0a1LCtA%3D%3D",
			MaxPrice:    -1,
		}},
	}
	MainTaskGroup.MonitorRetailer = enums.Amazon
	MainTask.TaskRetailer = enums.Amazon
	MainTaskGroup.MonitorStatus = enums.MonitorIdle
	TestDriver(MainTask, *MainProfile, *MainTaskGroup)
	select {}
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

	MainTaskGroup.BestbuyMonitorInfo = &entities.BestbuyMonitorInfo{
		ID:          MainMonitorID,
		TaskGroupID: MainTaskGroupID,
		Monitors: []entities.BestbuySingleMonitorInfo{{
			MonitorID:   MainMonitorID,
			TaskGroupID: MainTaskGroupID,
			SKU:         "5901353",
			MaxPrice:    -1,
		}},
	}
	MainTaskGroup.MonitorRetailer = enums.BestBuy
	MainTask.TaskRetailer = enums.BestBuy
	MainTaskGroup.MonitorStatus = enums.MonitorIdle
	TestDriver(MainTask, *MainProfile, *MainTaskGroup)
	select {}
}

func TestWalmart(t *testing.T) {
	MainTask.WalmartTaskInfo = &entities.WalmartTaskInfo{
		TaskID:      MainTaskID,
		TaskGroupID: MainTaskGroupID,
	}

	MainTaskGroup.WalmartMonitorInfo = &entities.WalmartMonitorInfo{
		ID:          MainMonitorID,
		TaskGroupID: MainTaskGroupID,
		Monitors: []entities.WalmartSingleMonitorInfo{{
			MonitorID:     MainMonitorID,
			TaskGroupID:   MainTaskGroupID,
			ID:            "855737650",
			MaxPrice:      -1,
			SoldByWalmart: true,
			MonitorType:   enums.SKUMonitor,
		}},
	}
	MainTaskGroup.MonitorRetailer = enums.Walmart
	MainTask.TaskRetailer = enums.Walmart
	MainTaskGroup.MonitorStatus = enums.MonitorIdle
	TestDriver(MainTask, *MainProfile, *MainTaskGroup)
	select {}
}
