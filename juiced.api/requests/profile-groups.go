package requests

type ProfileGroupsRequest struct {
	ProfileGroupIDs []string `json:"profileGroupIDs"`
	ProfileIDs      []string `json:"profileIDs"`
}
