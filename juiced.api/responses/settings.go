package responses

var UpdateSettingsParseErrorResponse = Response{StatusCode: 400, ErrorCode: 1000, Message: "Could not parse settings: "}
var UpdateSettingsUpdateErrorResponse = Response{StatusCode: 500, ErrorCode: 1001, Message: "Could not update settings: "}
var UpdateSettingsStartAYCDWarningResponse = Response{StatusCode: 200, ErrorCode: 1002, Message: "Could not restart AYCD, restarting the bot is recommended: "}
