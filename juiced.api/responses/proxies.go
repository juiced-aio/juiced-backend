package responses

var CreateProxyGroupParseErrorResponse = Response{StatusCode: 400, ErrorCode: 3000, Message: "Could not parse proxy group: "}
var CreateProxyGroupCreateErrorResponse = Response{StatusCode: 400, ErrorCode: 3001, Message: "Could not create proxy group: "}

var UpdateProxyGroupEmptyIDErrorResponse = Response{StatusCode: 404, ErrorCode: 3100, Message: "ID parameter was empty"}
var UpdateProxyGroupParseErrorResponse = Response{StatusCode: 400, ErrorCode: 3101, Message: "Could not parse proxy group: "}
var UpdateProxyGroupNotFoundErrorResponse = Response{StatusCode: 404, ErrorCode: 3102, Message: "Could not update proxy group: "}
var UpdateProxyGroupUpdateErrorResponse = Response{StatusCode: 500, ErrorCode: 3103, Message: "Could not update proxy group: "}

var DeleteProxyGroupsEmptyInputErrorResponse = Response{StatusCode: 400, ErrorCode: 3200, Message: "ProxyGroupIDs field was empty"}
var DeleteProxyGroupsParseErrorResponse = Response{StatusCode: 400, ErrorCode: 3201, Message: "Could not parse proxy group IDs: "}
var DeleteProxyGroupsDeleteErrorResponse = Response{StatusCode: 500, ErrorCode: 3202, Message: "Could not delete proxy group(s): "}

var CloneProxyGroupsEmptyInputErrorResponse = Response{StatusCode: 400, ErrorCode: 3300, Message: "ProxyGroupIDs field was empty"}
var CloneProxyGroupsParseErrorResponse = Response{StatusCode: 400, ErrorCode: 3301, Message: "Could not parse proxy group IDs: "}
var CloneProxyGroupsCloneErrorResponse = Response{StatusCode: 500, ErrorCode: 3302, Message: "Could not create new proxy group(s): "}

type ProxyGroupsSuccessResponse struct {
	SuccessProxyGroupIDs []string `json:"successProxyGroupIDs"`
	FailureProxyGroupIDs []string `json:"failureProxyGroupIDs"`
}
