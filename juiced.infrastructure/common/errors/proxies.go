package errors

// ParseProxyGroupError is the error encountered when parsing JSON into a ProxyGroup returns an error
const ParseProxyGroupError = "Parsing the JSON into a ProxyGroup returned an error: "

// CreateProxyGroupError is the error encountered when inserting a ProxyGroup into the DB returns an error
const CreateProxyGroupError = "Inserting the ProxyGroup into the DB returned an error: "

// GetProxyGroupError is the error encountered when retrieving a ProxyGroup from the DB returns an error
const GetProxyGroupError = "Retrieving the ProxyGroup with the given ID returned an error: "

// GetAllProxyGroupsError is the error encountered when retrieving all ProxyGroups from the DB returns an error
const GetAllProxyGroupsError = "Retrieving all ProxyGroups returned an error: "

// RemoveProxyGroupError is the error encountered when removing a ProxyGroup from the DB returns an error
const RemoveProxyGroupError = "Removing the ProxyGroup with the given ID returned an error: "

// UpdateProxyGroupError is the error encountered when updating a ProxyGroup from the DB returns an error
const UpdateProxyGroupError = "Updating the ProxyGroup with the given ID returned an error: "

// ParseRemoveProxyGroupsRequestError is the error encountered when parsing JSON into a RemoveProxyGroupsRequest object returns an error
const ParseRemoveProxyGroupsRequestError = "Parsing the JSON into a RemoveProxyGroupsRequest returned an error: "

// ParseCloneProxyGroupsRequestError is the error encountered when parsing JSON into a CloneProxyGroupsRequest object returns an error
const ParseCloneProxyGroupsRequestError = "Parsing the JSON into a CloneProxyGroupsRequest returned an error: "
