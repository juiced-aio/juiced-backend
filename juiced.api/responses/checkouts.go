package responses

import (
	"backend.juicedbot.io/juiced.infrastructure/entities"
)

// CheckoutResponse is the response that any /api/checkout request receives
type CheckoutResponse struct {
	Success bool                `json:"success"`
	Data    []entities.Checkout `json:"data"`
	Errors  []string            `json:"errors"`
}
