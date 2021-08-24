package ws

type IncomingMessage struct {
	EventType string `json:"eventType"`
	TaskID    string `json:"taskID"`
	// CardDetails walmart.CardInfo `json:"cardDetails"`
}
