package main

import (
	"time"

	api "backend.juicedbot.io/juiced.api"
	"backend.juicedbot.io/juiced.infrastructure/common"
	"backend.juicedbot.io/juiced.infrastructure/common/entities"
	"backend.juicedbot.io/juiced.infrastructure/common/events"
	"backend.juicedbot.io/juiced.infrastructure/common/stores"
	"backend.juicedbot.io/juiced.infrastructure/queries"

	sec "backend.juicedbot.io/juiced.security/auth/util"
	"backend.juicedbot.io/juiced.sitescripts/util"

	ws "backend.juicedbot.io/juiced.ws"
	"github.com/hugolgst/rich-go/client"
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
				eventBus.Unsubscribe(channel)
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
			go util.DiscordWebhookQueue()
			go api.StartServer()

			err := client.Login("856936229223006248")
			// No need to close the app if Discord RPC doesn't work. It's not a necessary feature.
			// If it breaks for everyone at once for some reason, don't want to entirely break the app without a hotfix.
			if err == nil {
				start := time.Now()
				client.SetActivity(client.Activity{
					Details:    "Beta - " + userInfo.UserVer, // TODO @silent -- Show the application version, rather than the backend version
					LargeImage: "main-juiced",
					LargeText:  "Juiced",
					SmallImage: "",
					SmallText:  "",
					Timestamps: &client.Timestamps{
						Start: &start,
					},
					Buttons: []*client.Button{
						{
							Label: "Dashboard",
							Url:   "https://dash.juicedbot.io/",
						},
					},
				})
			}
		}
	}()
	select {}
}

func Heartbeat(eventBus *events.EventBus, userInfo entities.UserInfo) {
	lastChecked := time.Now()
	var err error
	for {
		if time.Since(lastChecked).Seconds() > 60 {
			userInfo, err = sec.Heartbeat(userInfo, 0)
			if err != nil {
				eventBus.PublishCloseEvent()
			}
			lastChecked = time.Now()
		}
	}
}
