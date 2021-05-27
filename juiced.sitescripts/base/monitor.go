package base

import (
	"net/http"

	"backend.juicedbot.io/juiced.infrastructure/common/entities"
	"backend.juicedbot.io/juiced.infrastructure/common/events"
)

type Monitor struct {
	TaskGroup  *entities.TaskGroup
	Proxy      entities.Proxy
	EventBus   *events.EventBus
	Client     http.Client
	StopFlag   bool
	ErrorField string
}
