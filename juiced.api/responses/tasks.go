package responses

var CreateTaskParseErrorResponse = Response{StatusCode: 400, ErrorCode: 6000, Message: "Could not parse task: "}
var CreateTaskCreateErrorResponse = Response{StatusCode: 400, ErrorCode: 6001, Message: "Could not create task: "}
