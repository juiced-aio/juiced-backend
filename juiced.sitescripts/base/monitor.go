package base

import (
	"backend.juicedbot.io/juiced.client/http"
	"backend.juicedbot.io/juiced.infrastructure/common/entities"
	"backend.juicedbot.io/juiced.infrastructure/common/events"
)

type Monitor struct {
	TaskGroup  *entities.TaskGroup
	Proxy      *entities.Proxy
	ProxyGroup *entities.ProxyGroup
	EventBus   *events.EventBus
	Client     http.Client
	StopFlag   bool
	ErrorField string
}
