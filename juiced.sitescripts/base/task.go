package base

import (
	"net/http"

	"juiced/juiced.infrastructure/common/entities"
	"juiced/juiced.infrastructure/common/events"
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
