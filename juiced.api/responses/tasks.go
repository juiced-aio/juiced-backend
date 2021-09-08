package responses

var CreateTaskParseErrorResponse = Response{StatusCode: 400, ErrorCode: 6000, Message: "Could not parse task: "}
var CreateTaskEmptyInputErrorResponse = Response{StatusCode: 400, ErrorCode: 6001, Message: "One or more required fields were empty"}
var CreateTaskInvalidProfileGroupErrorResponse = Response{StatusCode: 400, ErrorCode: 6002, Message: "Could not find profile group: "}
var CreateTaskCreateErrorResponse = Response{StatusCode: 500, ErrorCode: 6003, Message: "Could not create task: "}

var UpdateTaskEmptyIDErrorResponse = Response{StatusCode: 404, ErrorCode: 6100, Message: "ID parameter was empty"}
var UpdateTaskParseErrorResponse = Response{StatusCode: 400, ErrorCode: 6101, Message: "Could not parse task: "}
var UpdateTaskNotFoundErrorResponse = Response{StatusCode: 404, ErrorCode: 6102, Message: "Could not update task: "}
var UpdateTaskUpdateErrorResponse = Response{StatusCode: 500, ErrorCode: 6103, Message: "Could not update task: "}

var DeleteTasksEmptyInputErrorResponse = Response{StatusCode: 400, ErrorCode: 6200, Message: "TaskIDs field was empty"}
var DeleteTasksParseErrorResponse = Response{StatusCode: 400, ErrorCode: 6201, Message: "Could not parse task IDs: "}
var DeleteTasksDeleteErrorResponse = Response{StatusCode: 500, ErrorCode: 6202, Message: "Could not delete task(s): "}

var CloneTasksEmptyInputErrorResponse = Response{StatusCode: 400, ErrorCode: 6300, Message: "TaskIDs field was empty"}
var CloneTasksParseErrorResponse = Response{StatusCode: 400, ErrorCode: 6301, Message: "Could not parse task IDs: "}
var CloneTasksCloneErrorResponse = Response{StatusCode: 500, ErrorCode: 6302, Message: "Could not create new task(s): "}

type TasksSuccessResponse struct {
	SuccessTaskIDs []string `json:"successTaskIDs"`
	FailureTaskIDs []string `json:"failureTaskIDs"`
}
