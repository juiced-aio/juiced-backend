package requests

type CreateTasksRequest struct {
	NumTasksPerProfile int                    `json:"numTasksPerProfile"`
	TaskGroupID        string                 `json:"taskGroupID"`
	Retailer           string                 `json:"retailer"`
	ProfileIDs         []string               `json:"profileIDs"`
	ProfileGroupID     string                 `json:"profileGroupID"`
	ProxyGroupID       string                 `json:"proxyGroupID"`
	ProfileID          string                 `json:"profileID"`
	Quantity           int                    `json:"quantity"`
	DelayMS            int                    `json:"delayMS"`
	SiteSpecificInfo   map[string]interface{} `json:"siteSpecificInfo"`
}
