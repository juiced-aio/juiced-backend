package requests

type SetVersionRequest struct {
	Version string `json:"version"`
	Channel string `json:"channel"`
}

type TestDiscordRequest struct {
	SuccessWebhook string `json:"successWebhook"`
	FailureWebhook string `json:"failureWebhook"`
}
