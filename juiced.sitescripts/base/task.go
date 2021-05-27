package base

import (
	"net/http"

	"backend.juicedbot.io/m/v2/juiced.infrastructure/common/entities"
	"backend.juicedbot.io/m/v2/juiced.infrastructure/common/events"
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
