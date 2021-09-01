package responses

var CreateProfileParseErrorResponse = Response{StatusCode: 400, ErrorCode: 4000, Message: "Could not parse profile: "}
var CreateProfileCreateErrorResponse = Response{StatusCode: 400, ErrorCode: 4001, Message: "Could not create profile: "}

var UpdateProfileEmptyIDErrorResponse = Response{StatusCode: 404, ErrorCode: 4100, Message: "ID parameter was empty"}
var UpdateProfileParseErrorResponse = Response{StatusCode: 400, ErrorCode: 4101, Message: "Could not parse profile: "}
var UpdateProfileNotFoundErrorResponse = Response{StatusCode: 404, ErrorCode: 4102, Message: "Could not update profile: "}
var UpdateProfileUpdateErrorResponse = Response{StatusCode: 500, ErrorCode: 4103, Message: "Could not update profile: "}

var DeleteProfilesEmptyInputErrorResponse = Response{StatusCode: 400, ErrorCode: 4200, Message: "ProfileIDs field was empty"}
var DeleteProfilesParseErrorResponse = Response{StatusCode: 400, ErrorCode: 4201, Message: "Could not parse profile IDs: "}
var DeleteProfilesDeleteErrorResponse = Response{StatusCode: 500, ErrorCode: 4202, Message: "Could not delete profile(s): "}

var CloneProfilesEmptyInputErrorResponse = Response{StatusCode: 400, ErrorCode: 4300, Message: "ProfileIDs field was empty"}
var CloneProfilesParseErrorResponse = Response{StatusCode: 400, ErrorCode: 4301, Message: "Could not parse profile IDs: "}
var CloneProfilesCloneErrorResponse = Response{StatusCode: 500, ErrorCode: 4302, Message: "Could not create new profile(s): "}

type ProfilesSuccessResponse struct {
	SuccessProfileIDs []string `json:"successProfileIDs"`
	FailureProfileIDs []string `json:"failureProfileIDs"`
}
