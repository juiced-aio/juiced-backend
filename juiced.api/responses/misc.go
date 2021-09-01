package responses

var SetVersionParseErrorResponse = Response{StatusCode: 400, ErrorCode: 9000, Message: "Could not parse request: "}
var SetVersionEmptyInputErrorResponse = Response{StatusCode: 400, ErrorCode: 9001, Message: "Version field was empty"}
var SetVersionStartRPCWarningResponse = Response{StatusCode: 200, ErrorCode: 9002, Message: "Could not launch Discord Rich Presence service"}
