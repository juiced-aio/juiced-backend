package errors

// IOUtilReadAllError is the error encountered when the ioutil.ReadAll() function returns an error
const IOUtilReadAllError = "ioutil.ReadAll() function returned an error: "

// ParseObjectIDError is the error encountered when parsing a bson ObjectID from a string returns an error
const ParseObjectIDError = "Parsing the ObjectID returned an error: "

// ParseQueryParamsError is the error encountered when parsing a query parameter (e.g. string to int) returns an error
const ParseQueryParamsError = "Parsing the query returned an error: "
