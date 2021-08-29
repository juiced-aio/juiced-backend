package errors

const (
	MissingParameterError = "request URL missing a parameter: "
	IOUtilReadAllError    = "ioutil.ReadAll() function returned an error: "
	UnmarshalError        = "Parsing the request returned an error: "
	BadRequestError       = "The request was invalid: "
)
