package responses

import "backend.juicedbot.io/juiced.infrastructure/entities"

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

var StartTasksEmptyInputErrorResponse = Response{StatusCode: 400, ErrorCode: 6400, Message: "TaskIDs field was empty"}
var StartTasksParseErrorResponse = Response{StatusCode: 400, ErrorCode: 6401, Message: "Could not parse task IDs: "}
var StartTasksStartErrorResponse = Response{StatusCode: 500, ErrorCode: 6402, Message: "Could not start task(s): "}

var StopTasksEmptyInputErrorResponse = Response{StatusCode: 400, ErrorCode: 6500, Message: "TaskIDs field was empty"}
var StopTasksParseErrorResponse = Response{StatusCode: 400, ErrorCode: 6501, Message: "Could not parse task IDs: "}
var StopTasksStopErrorResponse = Response{StatusCode: 500, ErrorCode: 6502, Message: "Could not stop task(s): "}

var UpdateTasksEmptyInputErrorResponse = Response{StatusCode: 400, ErrorCode: 6600, Message: "TaskIDs field was empty"}
var UpdateTasksParseErrorResponse = Response{StatusCode: 400, ErrorCode: 6601, Message: "Could not parse task IDs: "}
var UpdateTasksUpdateErrorResponse = Response{StatusCode: 500, ErrorCode: 6602, Message: "Could not update task(s): "}

type TasksSuccessResponse struct {
	SuccessTaskIDs []string `json:"successTaskIDs"`
	FailureTaskIDs []string `json:"failureTaskIDs"`
}

type TasksResponse struct {
	Data []entities.Task `json:"data"`
}
