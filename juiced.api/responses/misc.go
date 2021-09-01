package responses

var SetVersionParseErrorResponse = Response{StatusCode: 400, ErrorCode: 9000, Message: "Could not parse request: "}
var SetVersionEmptyInputErrorResponse = Response{StatusCode: 400, ErrorCode: 9001, Message: "Version field was empty"}
var SetVersionStartRPCWarningResponse = Response{StatusCode: 200, ErrorCode: 9002, Message: "Could not launch Discord Rich Presence service"}

var TestDiscordParseErrorResponse = Response{StatusCode: 400, ErrorCode: 9100, Message: "Could not parse request: "}
var TestDiscordEmptyInputErrorResponse = Response{StatusCode: 400, ErrorCode: 9101, Message: "Both SuccessWebhook and FailureWebhook fields were empty"}
var TestDiscordTestErrorResponse = Response{StatusCode: 500, ErrorCode: 9102, Message: "Sending the webhook failed: "}

type TestDiscordResponse struct {
	SuccessMessage string `json:"successMessage"`
	FailureMessage string `json:"failureMessage"`
}
