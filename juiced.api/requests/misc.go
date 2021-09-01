package requests

type SetVersionRequest struct {
	Version string `json:"version"`
}

type TestDiscordRequest struct {
	SuccessWebhook string `json:"successWebhook"`
	FailureWebhook string `json:"failureWebhook"`
}
