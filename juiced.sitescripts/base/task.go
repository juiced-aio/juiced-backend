package base

import (
	"backend.juicedbot.io/juiced.client/http"
	"backend.juicedbot.io/juiced.infrastructure/common/entities"
	"backend.juicedbot.io/juiced.infrastructure/common/events"
)

type Task struct {
	Task       *entities.Task `json:"task"`
	Profile    entities.Profile
	Proxy      entities.Proxy `json:"proxy"`
	EventBus   *events.EventBus
	Client     http.Client
	StopFlag   bool
	ErrorField string
}
