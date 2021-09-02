package requests

import "backend.juicedbot.io/juiced.infrastructure/entities"

type ProfilesRequest struct {
	ProfileIDs []string `json:"profileIDs"`
}

type ImportProfilesRequest struct {
	GroupIDs []string `json:"groupIDs"`
	FilePath string   `json:"filepath"`
}

type ExportProfilesRequest struct {
	ProfileIDs []string `json:"profileIDs"`
	FilePath   string   `json:"filepath"`
}

type ProfilesFileFormat struct {
	Profiles []entities.Profile `json:"profiles"`
}
