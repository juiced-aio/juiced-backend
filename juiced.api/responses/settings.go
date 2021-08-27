package responses

import (
	"backend.juicedbot.io/juiced.infrastructure/entities"
)

// SettingsResponse is the response that any /api/settings request receives
type SettingsResponse struct {
	Success bool              `json:"success"`
	Data    entities.Settings `json:"data"`
	Errors  []string          `json:"errors"`
}
