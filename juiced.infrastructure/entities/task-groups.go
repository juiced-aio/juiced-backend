package entities

import (
	"backend.juicedbot.io/juiced.client/http"
	"backend.juicedbot.io/juiced.infrastructure/events"
	"backend.juicedbot.io/juiced.sitescripts/hawk-go"
)

type TaskGroup struct {
}

type MonitorInfo struct {
	TaskGroup  *TaskGroup
	Proxy      *Proxy
	ProxyGroup *ProxyGroup
	EventBus   *events.EventBus
	Client     http.Client
	Scraper    hawk.Scraper
	StopFlag   bool
	ErrorField string

	InStock []StockInfo
}
