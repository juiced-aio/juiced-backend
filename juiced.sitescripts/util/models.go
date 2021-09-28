package util

import (
	"backend.juicedbot.io/juiced.antibot/cloudflare"
	"backend.juicedbot.io/juiced.client/http"
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
	Scraper            cloudflare.Scraper
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

var DefaultRawHeaders = [][2]string{
	{"sec-ch-ua", `" Not A;Brand";v="99", "Chromium";v="90", "Google Chrome";v="90"`},
	{"sec-ch-ua-mobile", "?0"},
	{"upgrade-insecure-requests", "1"},
	{"user-agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/90.0.4430.212 Safari/537.36"},
	{"accept", "text/html,application/xhtml+xml,application/xml;q=0.9,image/avif,image/webp,image/apng,*/*;q=0.8,application/signed-exchange;v=b3;q=0.9"},
	{"sec-fetch-site", "none"},
	{"sec-fetch-mode", "navigate"},
	{"sec-fetch-user", "?1"},
	{"sec-fetch-dest", "document"},
	{"accept-encoding", "gzip, deflate, br"},
	{"accept-language", "en-US,en;q=0.9"},
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
