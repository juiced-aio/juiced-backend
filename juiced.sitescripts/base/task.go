package base

import (
	"net/http"

	"backend.juicedbot.io/juiced.infrastructure/common/entities"
	"backend.juicedbot.io/juiced.infrastructure/common/events"
)

type Task struct {
	Task           *entities.Task
	Profile        entities.Profile
	Proxy          entities.Proxy
	DiscordWebhook string
	EventBus       *events.EventBus
	Client         http.Client
	StopFlag       bool
	ErrorField     string
}
