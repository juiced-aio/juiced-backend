package responses

import "backend.juicedbot.io/juiced.infrastructure/entities"

var CreateProfileParseErrorResponse = Response{StatusCode: 400, ErrorCode: 4000, Message: "Could not parse profile: "}
var CreateProfileUnsupportedPaymentMethodErrorResponse = Response{StatusCode: 400, ErrorCode: 4001, Message: "Unsupported payment method"}
var CreateProfileCreateErrorResponse = Response{StatusCode: 400, ErrorCode: 4002, Message: "Could not create profile: "}

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

var ImportProfilesEmptyInputErrorResponse = Response{StatusCode: 400, ErrorCode: 4400, Message: "FilePath field was empty"}
var ImportProfilesEmptyProfilesErrorResponse = Response{StatusCode: 400, ErrorCode: 4401, Message: "No valid profiles detected"}
var ImportProfilesParseErrorResponse = Response{StatusCode: 400, ErrorCode: 4402, Message: "Could not parse input: "}
var ImportProfilesImportErrorResponse = Response{StatusCode: 500, ErrorCode: 4403, Message: "Could not import profile(s): "}

var ExportProfilesEmptyInputErrorResponse = Response{StatusCode: 400, ErrorCode: 4500, Message: "FilePath or ProfileIDs field was empty"}
var ExportProfilesEmptyProfilesErrorResponse = Response{StatusCode: 400, ErrorCode: 4501, Message: "No valid profiles detected"}
var ExportProfilesFilePathNotJSONErrorResponse = Response{StatusCode: 400, ErrorCode: 4502, Message: "FilePath must end in .json"}
var ExportProfilesParseErrorResponse = Response{StatusCode: 400, ErrorCode: 4503, Message: "Could not parse input: "}
var ExportProfilesFileAlreadyExists = Response{StatusCode: 400, ErrorCode: 4504, Message: "File already exists"}
var ExportProfilesExportErrorResponse = Response{StatusCode: 500, ErrorCode: 4505, Message: "Could not export profile(s): "}

type ProfilesSuccessResponse struct {
	SuccessProfileIDs []string `json:"successProfileIDs"`
	FailureProfileIDs []string `json:"failureProfileIDs"`
}

type ImportProfilesSuccessResponse struct {
	NewProfiles      []entities.Profile      `json:"newProfiles"`
	SkippedProfiles  int                     `json:"skippedProfiles"`
	NewProfileGroups []entities.ProfileGroup `json:"newProfileGroups"`
	SkippedGroups    int                     `json:"skippedGroups"`
}
