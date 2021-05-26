package main

import (
	api "juiced/juiced.api"
	"juiced/juiced.infrastructure/common/entities"
	"juiced/juiced.infrastructure/common/events"
	"juiced/juiced.infrastructure/common/stores"
	"juiced/juiced.infrastructure/queries"
	"juiced/juiced.security/auth/util"
	ws "juiced/juiced.ws"
	"time"
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
