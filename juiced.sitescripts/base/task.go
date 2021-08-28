package base

import (
	"time"

	"backend.juicedbot.io/juiced.client/http"
	"backend.juicedbot.io/juiced.infrastructure/common/entities"
	"backend.juicedbot.io/juiced.infrastructure/common/events"
	"backend.juicedbot.io/juiced.sitescripts/hawk-go"
)

type Task struct {
	Task         *entities.Task
	Profile      entities.Profile
	Proxy        *entities.Proxy
	ProxyGroup   *entities.ProxyGroup
	EventBus     *events.EventBus
	Client       http.Client
	Scraper      hawk.Scraper
	StartTime    time.Time
	EndTime      time.Time
	HasStockData bool
	StopFlag     bool
	ErrorField   string
}
