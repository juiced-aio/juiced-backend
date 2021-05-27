package main

import (
	"time"

	api "backend.juicedbot.io/m/v2/juiced.api"
	"backend.juicedbot.io/m/v2/juiced.infrastructure/common"
	"backend.juicedbot.io/m/v2/juiced.infrastructure/common/entities"
	"backend.juicedbot.io/m/v2/juiced.infrastructure/common/events"
	"backend.juicedbot.io/m/v2/juiced.infrastructure/common/stores"
	"backend.juicedbot.io/m/v2/juiced.infrastructure/queries"
	"backend.juicedbot.io/m/v2/juiced.security/auth/util"
	ws "backend.juicedbot.io/m/v2/juiced.ws"
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
			eventBus.PublishCloseEvent()
		}

		// Get the user's info
		_, userInfo, err := queries.GetUserInfo()
		if err != nil {
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
