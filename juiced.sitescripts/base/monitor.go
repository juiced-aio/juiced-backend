package base

import (
	"net/http"

	"juiced/juiced.infrastructure/common/entities"
	"juiced/juiced.infrastructure/common/events"
)

type Monitor struct {
	TaskGroup  *entities.TaskGroup
	Proxy      entities.Proxy
	EventBus   *events.EventBus
	Client     http.Client
	StopFlag   bool
	ErrorField string
}
