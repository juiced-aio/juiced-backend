package responses

var CreateAccountParseErrorResponse = Response{StatusCode: 400, ErrorCode: 2000, Message: "Could not parse account: "}
var CreateAccountCreateErrorResponse = Response{StatusCode: 400, ErrorCode: 2001, Message: "Could not create account: "}

var UpdateAccountEmptyIDErrorResponse = Response{StatusCode: 404, ErrorCode: 2100, Message: "ID parameter was empty"}
var UpdateAccountParseErrorResponse = Response{StatusCode: 400, ErrorCode: 2101, Message: "Could not parse account: "}
var UpdateAccountNotFoundErrorResponse = Response{StatusCode: 404, ErrorCode: 2102, Message: "Could not update account: "}
var UpdateAccountUpdateErrorResponse = Response{StatusCode: 500, ErrorCode: 2103, Message: "Could not update account: "}

var DeleteAccountsEmptyInputErrorResponse = Response{StatusCode: 400, ErrorCode: 2200, Message: "AccountIDs field was empty"}
var DeleteAccountsParseErrorResponse = Response{StatusCode: 400, ErrorCode: 2201, Message: "Could not parse account IDs: "}
var DeleteAccountsDeleteErrorResponse = Response{StatusCode: 500, ErrorCode: 2202, Message: "Could not delete account(s): "}

type AccountsSuccessResponse struct {
	SuccessAccountIDs []string `json:"successAccountIDs"`
	FailureAccountIDs []string `json:"failureAccountIDs"`
}
