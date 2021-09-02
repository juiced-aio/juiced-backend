package responses

var CreateTaskParseErrorResponse = Response{StatusCode: 400, ErrorCode: 6000, Message: "Could not parse task: "}
var CreateTaskEmptyInputErrorResponse = Response{StatusCode: 400, ErrorCode: 6001, Message: "One or more required fields were empty"}
var CreateTaskInvalidProfileGroupErrorResponse = Response{StatusCode: 400, ErrorCode: 6002, Message: "Could not find profile group: "}
var CreateTaskCreateErrorResponse = Response{StatusCode: 500, ErrorCode: 6003, Message: "Could not create task: "}
