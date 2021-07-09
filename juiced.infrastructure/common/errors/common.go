package errors

// IOUtilReadAllError is the error encountered when the ioutil.ReadAll() function returns an error
const IOUtilReadAllError = "ioutil.ReadAll() function returned an error: "

// MissingParameterError is the error when a URL Parameter can not be found
const MissingParameterError = "URL is missing a parameter"

// MissingRequiredBody is the error when a Requests body doesn't have needed values
const MissingRequiredBody = "Request Body is missing required values"
