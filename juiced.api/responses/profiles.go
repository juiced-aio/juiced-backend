package responses

var CreateProfileParseErrorResponse = Response{StatusCode: 400, ErrorCode: 4000, Message: "Could not parse profile: "}
var CreateProfileCreateErrorResponse = Response{StatusCode: 400, ErrorCode: 4001, Message: "Could not create profile: "}
