package responses

var CreateTaskGroupParseErrorResponse = Response{StatusCode: 400, ErrorCode: 7000, Message: "Could not parse task group: "}
var CreateTaskGroupCreateErrorResponse = Response{StatusCode: 500, ErrorCode: 7001, Message: "Could not create task group: "}

var UpdateTaskGroupEmptyIDErrorResponse = Response{StatusCode: 404, ErrorCode: 7100, Message: "ID parameter was empty"}
var UpdateTaskGroupParseErrorResponse = Response{StatusCode: 400, ErrorCode: 7101, Message: "Could not parse task group: "}
var UpdateTaskGroupNotFoundErrorResponse = Response{StatusCode: 404, ErrorCode: 7102, Message: "Could not update task group: "}
var UpdateTaskGroupUpdateErrorResponse = Response{StatusCode: 500, ErrorCode: 7103, Message: "Could not update task group: "}

var DeleteTaskGroupsEmptyInputErrorResponse = Response{StatusCode: 400, ErrorCode: 7200, Message: "TaskGroupIDs field was empty"}
var DeleteTaskGroupsParseErrorResponse = Response{StatusCode: 400, ErrorCode: 7201, Message: "Could not parse task group IDs: "}
var DeleteTaskGroupsDeleteErrorResponse = Response{StatusCode: 500, ErrorCode: 7202, Message: "Could not delete task group(s): "}

var CloneTaskGroupsEmptyInputErrorResponse = Response{StatusCode: 400, ErrorCode: 7300, Message: "TaskGroupIDs field was empty"}
var CloneTaskGroupsParseErrorResponse = Response{StatusCode: 400, ErrorCode: 7301, Message: "Could not parse task group IDs: "}
var CloneTaskGroupsCloneErrorResponse = Response{StatusCode: 500, ErrorCode: 7302, Message: "Could not create new task group(s): "}

type TaskGroupsSuccessResponse struct {
	SuccessTaskGroupIDs []string `json:"successTaskGroupIDs"`
	FailureTaskGroupIDs []string `json:"failureTaskGroupIDs"`
}