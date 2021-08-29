package util

import (
	"backend.juicedbot.io/juiced.client/http"
	"backend.juicedbot.io/juiced.sitescripts/hawk-go"
)

type ErrorType = string

const (
	EncryptionParsingError    ErrorType = "ENCRYPTION_PARSING_ERROR"
	EncryptionEncryptingError ErrorType = "ENCRYPTION_ENCRYPTING_ERROR"
	RequestMarshalBodyError   ErrorType = "REQUEST_MARSHAL_BODY_ERROR"
	RequestCreateError        ErrorType = "REQUEST_CREATE_ERROR"
	RequestDoError            ErrorType = "REQUEST_DO_ERROR"
	RequestReadBodyError      ErrorType = "REQUEST_READ_BODY_ERROR"
	RequestUnmarshalBodyError ErrorType = "REQUEST_UNMARSHAL_BODY_ERROR"
	ShippingTypeError         ErrorType = "INVALID_SHIPPING_TYPE"
	LoginDetailsError         ErrorType = "LOGIN_DETAILS_ERROR"
)

type AddHeadersFunction func(*http.Request, ...string)

type Request struct {
	Client             *http.Client
	Scraper            hawk.Scraper
	Method             string
	URL                string
	Headers            http.Header
	RawHeaders         [][2]string
	AddHeadersFunction AddHeadersFunction
	Referer            string
	Data               []byte
	RequestBodyStruct  interface{}
	ResponseBodyStruct interface{}
	RandOpt            string
}

// type SensorRequest struct {
// 	SensorData string `json:"sensor_data"`
// }

// type SensorResponse struct {
// 	Success bool `json:"success"`
// }

// type PXValues struct {
// 	RefreshAt int64
// 	SetID     string
// 	VID       string
// 	UUID      string
// }
