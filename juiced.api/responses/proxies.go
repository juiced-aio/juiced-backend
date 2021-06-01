package responses

import (
	"backend.juicedbot.io/juiced.infrastructure/common/entities"
)

// ProxyGroupResponse is the response that any /api/proxy/group request receives
type ProxyGroupResponse struct {
	Success bool                  `json:"success"`
	Data    []entities.ProxyGroup `json:"data"`
	Errors  []string              `json:"errors"`
}
