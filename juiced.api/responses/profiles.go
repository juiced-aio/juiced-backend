package responses

import (
	"backend.juicedbot.io/juiced.infrastructure/entities"
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

// ImportProfileResponse is the response that the /api/profile/import request receives
type ImportProfileResponse struct {
	Success         bool                                `json:"success"`
	NewProfiles     []entities.Profile                  `json:"newProfiles"`
	SkippedProfiles int                                 `json:"skippedProfiles"`
	SkippedGroups   int                                 `json:"skippedGroups"`
	Data            []entities.ProfileGroupWithProfiles `json:"data"`
	Errors          []string                            `json:"errors"`
}
