package responses

var CreateTaskGroupParseErrorResponse = Response{StatusCode: 400, ErrorCode: 5000, Message: "Could not parse task group: "}
var CreateTaskGroupCreateErrorResponse = Response{StatusCode: 400, ErrorCode: 5001, Message: "Could not create task group: "}
