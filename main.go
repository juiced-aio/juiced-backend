package main

import (
	"log"
	"time"

	api "backend.juicedbot.io/juiced.api"
	"backend.juicedbot.io/juiced.infrastructure/common"
	"backend.juicedbot.io/juiced.infrastructure/common/entities"
	"backend.juicedbot.io/juiced.infrastructure/common/events"
	"backend.juicedbot.io/juiced.infrastructure/common/stores"
	"backend.juicedbot.io/juiced.infrastructure/queries"
	"backend.juicedbot.io/juiced.security/auth/util"
	ws "backend.juicedbot.io/juiced.ws"
)

func main() {
	// Initalize the event bus
	events.InitEventBus()
	eventBus := events.GetEventBus()

	// Start the websocket server
	go ws.StartWebsocketServer(eventBus)

	go func() {
		// Wait for the client to connect to the websocket server
		channel := make(chan events.Event)
		eventBus.Subscribe(channel)
		for {
			event := <-channel
			if event.EventType == events.ConnectEventType {
				break
			}
		}

		// Initalize the database
		err := common.InitDatabase()
		if err != nil {
			log.Println(err.Error())
			eventBus.PublishCloseEvent()
		}

		// Get the user's info
		_, userInfo, err := queries.GetUserInfo()
		if err != nil {
			log.Println(err.Error())
			eventBus.PublishCloseEvent()
		} else {
			go Heartbeat(eventBus, userInfo)
			go stores.InitTaskStore(eventBus)
			stores.InitMonitorStore(eventBus)
			stores.InitCaptchaStore(eventBus)
			go api.StartServer()
		}
	}()
	for {
	}
}

func Heartbeat(eventBus *events.EventBus, userInfo entities.UserInfo) {
	lastChecked := time.Now()
	var err error
	for {
		if time.Since(lastChecked).Seconds() > 60 {
			userInfo, err = util.Heartbeat(userInfo, 0)
			if err != nil {
				eventBus.PublishCloseEvent()
			}
			lastChecked = time.Now()
		}
	}
}
