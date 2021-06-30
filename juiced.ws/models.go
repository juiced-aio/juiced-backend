package ws

import "backend.juicedbot.io/juiced.sitescripts/walmart"

type IncomingMessage struct {
	EventType   string           `json:"eventType"`
	TaskID      string           `json:"taskID"`
	CardDetails walmart.CardInfo `json:"cardDetails"`
}
