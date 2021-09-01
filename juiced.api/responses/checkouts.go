package responses

var GetCheckoutsParseErrorResponse = Response{StatusCode: 400, ErrorCode: 8000, Message: "Could not parse days field: "}
var GetCheckoutsNegativeDaysErrorResponse = Response{StatusCode: 400, ErrorCode: 8001, Message: "Days field was negative"}
