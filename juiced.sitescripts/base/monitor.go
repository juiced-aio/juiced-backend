package base

import (
	"backend.juicedbot.io/juiced.client/http"
	"backend.juicedbot.io/juiced.infrastructure/common/entities"
	"backend.juicedbot.io/juiced.infrastructure/common/events"
	"backend.juicedbot.io/juiced.sitescripts/hawk-go"
)

type Monitor struct {
	TaskGroup  *entities.TaskGroup
	Proxy      entities.Proxy
	Proxies    []entities.Proxy
	EventBus   *events.EventBus
	Client     http.Client
	Scraper    hawk.Scraper
	StopFlag   bool
	ErrorField string
}
