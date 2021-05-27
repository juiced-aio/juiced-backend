package main

import (
	"time"

	api "backend.juicedbot.io/m/v2/juiced.api"
	"backend.juicedbot.io/m/v2/juiced.infrastructure/common/entities"
	"backend.juicedbot.io/m/v2/juiced.infrastructure/common/events"
	"backend.juicedbot.io/m/v2/juiced.infrastructure/common/stores"
	"backend.juicedbot.io/m/v2/juiced.infrastructure/queries"
	"backend.juicedbot.io/m/v2/juiced.security/auth/util"
	ws "backend.juicedbot.io/m/v2/juiced.ws"
)

func main() {
	events.InitEventBus()
	eventBus := events.GetEventBus()
	userInfo, err := queries.GetUserInfo()
	if err != nil {
		eventBus.PublishCloseEvent()
	} else {
		go Heartbeat(eventBus, userInfo)
		go stores.InitTaskStore(eventBus)
		stores.InitMonitorStore(eventBus)
		stores.InitCaptchaStore(eventBus)
		go api.StartServer()
		go ws.StartWebsocketServer(eventBus)
	}
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
