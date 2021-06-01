package responses

import (
	"backend.juicedbot.io/juiced.infrastructure/common/entities"
)

// ProfileGroupResponse is the response that any /api/profile/group request receives
type ProfileGroupResponse struct {
	Success bool                                `json:"success"`
	Data    []entities.ProfileGroupWithProfiles `json:"data"`
	Errors  []string                            `json:"errors"`
}

// ProfileResponse is the response that any /api/profile request receives
type ProfileResponse struct {
	Success bool               `json:"success"`
	Data    []entities.Profile `json:"data"`
	Errors  []string           `json:"errors"`
}
