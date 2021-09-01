package responses

var CreateProfileGroupParseErrorResponse = Response{StatusCode: 400, ErrorCode: 5000, Message: "Could not parse profile group: "}
var CreateProfileGroupCreateErrorResponse = Response{StatusCode: 400, ErrorCode: 5001, Message: "Could not create profile group: "}

var UpdateProfileGroupEmptyIDErrorResponse = Response{StatusCode: 404, ErrorCode: 5100, Message: "ID parameter was empty"}
var UpdateProfileGroupParseErrorResponse = Response{StatusCode: 400, ErrorCode: 5101, Message: "Could not parse profile group: "}
var UpdateProfileGroupNotFoundErrorResponse = Response{StatusCode: 404, ErrorCode: 5102, Message: "Could not update profile group: "}
var UpdateProfileGroupUpdateErrorResponse = Response{StatusCode: 500, ErrorCode: 5103, Message: "Could not update profile group: "}

var DeleteProfileGroupsEmptyInputErrorResponse = Response{StatusCode: 400, ErrorCode: 5200, Message: "ProfileGroupIDs field was empty"}
var DeleteProfileGroupsParseErrorResponse = Response{StatusCode: 400, ErrorCode: 5201, Message: "Could not parse profile group IDs: "}
var DeleteProfileGroupsDeleteErrorResponse = Response{StatusCode: 500, ErrorCode: 5202, Message: "Could not delete profile group(s): "}

var CloneProfileGroupsEmptyInputErrorResponse = Response{StatusCode: 400, ErrorCode: 3300, Message: "ProfileGroupIDs field was empty"}
var CloneProfileGroupsParseErrorResponse = Response{StatusCode: 400, ErrorCode: 3301, Message: "Could not parse profile group IDs: "}
var CloneProfileGroupsCloneErrorResponse = Response{StatusCode: 500, ErrorCode: 3302, Message: "Could not create new profile group(s): "}

type ProfileGroupsSuccessResponse struct {
	SuccessProfileGroupIDs []string `json:"successProfileGroupIDs"`
	FailureProfileGroupIDs []string `json:"failureProfileGroupIDs"`
}
