package base

import (
	"backend.juicedbot.io/juiced.client/http"
	"backend.juicedbot.io/juiced.infrastructure/common/entities"
	"backend.juicedbot.io/juiced.infrastructure/common/events"
	"backend.juicedbot.io/juiced.sitescripts/hawk-go"
)

type Task struct {
	Task       *entities.Task
	Profile    entities.Profile
	Proxy      entities.Proxy
	EventBus   *events.EventBus
	Client     http.Client
	Scraper    hawk.Scraper
	StopFlag   bool
	ErrorField string
}
