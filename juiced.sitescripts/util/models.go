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

// // Discord webhook details
// type DiscordWebhook struct {
// 	Content interface{} `json:"content"`
// 	Embeds  []Embed     `json:"embeds"`
// }

// type HookInfo struct {
// 	Success bool
// 	Content string
// 	Embeds  []Embed
// }

// type Embed struct {
// 	Title     string    `json:"title"`
// 	Color     int       `json:"color"`
// 	Fields    []Field   `json:"fields"`
// 	Footer    Footer    `json:"footer"`
// 	Timestamp time.Time `json:"timestamp"`
// 	Thumbnail Thumbnail `json:"thumbnail"`
// }

// type Field struct {
// 	Name   string `json:"name"`
// 	Value  string `json:"value"`
// 	Inline bool   `json:"inline,omitempty"`
// }

// type Footer struct {
// 	Text    string `json:"text"`
// 	IconURL string `json:"icon_url"`
// }

// type Thumbnail struct {
// 	URL string `json:"url"`
// }

// type SensorRequest struct {
// 	SensorData string `json:"sensor_data"`
// }

// type SensorResponse struct {
// 	Success bool `json:"success"`
// }

// // All info needed for ProcessCheckout
// type ProcessCheckoutInfo struct {
// 	TaskInfo *entities.TaskInfo
// 	Success  bool
// 	Status   enums.OrderStatus
// 	Content  string
// 	Embeds   []entities.DiscordEmbed
// 	Retailer string
// }

// type PXValues struct {
// 	RefreshAt int64
// 	SetID     string
// 	VID       string
// 	UUID      string
// }
