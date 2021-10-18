package events

import "backend.juicedbot.io/juiced.infrastructure/enums"

type AccountEvent struct {
	Status    enums.AccountStatus    `json:"status"`
	EventType enums.AccountEventType `json:"eventType"`
	Data      interface{}            `json:"data"`
	AccountID string                 `json:"accountID"`
}
