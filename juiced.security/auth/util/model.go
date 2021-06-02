package util

import "time"

type AuthenticateRequest struct {
	ActivationToken string `json:"908tqy5VI2"`
	HWID            string `json:"BN9GSbnV6z"`
	DeviceName      string `json:"8Cgy8rh1Zp"`
}

type AuthenticateResponse struct {
	Success      bool   `json:"Gzdgq0zpma"`
	ErrorMessage string `json:"KARYl4Xg6o"`
}

type EncryptedAuthenticateResponse struct {
	Success      string `json:"Gzdgq0zpma"`
	ErrorMessage string `json:"KARYl4Xg6o"`
}

type ActivateRequest struct {
	Email      string `json:"BAD3Clg0xO"`
	Password   string `json:"JqedLzhouX"`
	HWID       string `json:"SYDvnWydIl"`
	DeviceName string `json:"6C5yNgnr37"`
}

type RefreshTokenRequest struct {
	ActivationToken string `json:"ah93iyhisE"`
	HWID            string `json:"5Q5M79ijzq"`
	DeviceName      string `json:"5U5uoGXnrH"`
	RefreshToken    string `json:"uYqnf3KUfh"`
}

type RefreshTokenResponse struct {
	Success         bool   `json:"4RvV3IUgR3"`
	ActivationToken string `json:"Q3G2BfRTS9"`
	RefreshToken    string `json:"n8pdbUXnXc"`
	ExpiresAt       int64  `json:"lf4dzxSwTs"`
	ErrorMessage    string `json:"bSqF23Iosy"`
}

type EncryptedRefreshTokenResponse struct {
	Success         string `json:"4RvV3IUgR3"`
	ActivationToken string `json:"Q3G2BfRTS9"`
	RefreshToken    string `json:"n8pdbUXnXc"`
	ExpiresAt       string `json:"lf4dzxSwTs"`
	ErrorMessage    string `json:"bSqF23Iosy"`
}

// Had to add this due to import cycling
type DiscordWebhookRequest struct {
	ActivationToken    string             `json:"xBwmRTJarL"`
	HWID               string             `json:"2wpOd42ZuD"`
	DeviceName         string             `json:"R7v60V37JA"`
	Success            string             `json:"j9TH4ZtGPH"`
	DiscordInformation DiscordInformation `json:"w2ohPULKJl"`
}

type DiscordInformation struct {
	Content string  `json:"d3ZrADZzcP"`
	Embeds  []Embed `json:"igBjO6n2bK"`
}
type DiscordEmbed struct {
	Title     string           `json:"title"`
	Color     int              `json:"color"`
	Fields    []DiscordField   `json:"fields"`
	Footer    DiscordFooter    `json:"footer"`
	Timestamp time.Time        `json:"timestamp"`
	Thumbnail DiscordThumbnail `json:"thumbnail"`
}

type DiscordField struct {
	Name   string `json:"name"`
	Value  string `json:"value"`
	Inline bool   `json:"inline,omitempty"`
}
type DiscordFooter struct {
	Text    string `json:"text"`
	IconURL string `json:"icon_url"`
}

type DiscordThumbnail struct {
	URL string `json:"url"`
}

type Embed struct {
	Title  string  `json:"UX3jllPu6d"`
	Fields []Field `json:"U3QuXzM0Q5"`
}

type Field struct {
	Name   string `json:"lJ5cmXdF3L"`
	Value  string `json:"LsD8aRXXXl"`
	Inline string `json:"H4dLDoqOuW"`
}

type DiscordWebhookResponse struct {
	Success      bool   `json:"6sn25iGFzS"`
	ErrorMessage string `json:"sMAsFnt7zJ"`
}

type EncryptedDiscordWebhookResponse struct {
	Success      string `json:"6sn25iGFzS"`
	ErrorMessage string `json:"sMAsFnt7zJ"`
}

type PXRequest struct {
	ActivationToken string `json:"zWI5WK8e5p"`
	HWID            string `json:"ExX0R7udUk"`
	DeviceName      string `json:"CUXOK794sk"`
	Proxy           string `json:"NvcUkcMcxD"`
	Site            string `json:"4jRwi6R1q1"`
}

type PXResponse struct {
	Success       bool          `json:"XOjKVhpMBN"`
	PXAPIResponse PXAPIResponse `json:"lJf19YVKtS"`
	ErrorMessage  string        `json:"PhPbyDK8L5"`
}

type EncryptedPXResponse struct {
	Success       string        `json:"XOjKVhpMBN"`
	PXAPIResponse PXAPIResponse `json:"lJf19YVKtS"`
	ErrorMessage  string        `json:"PhPbyDK8L5"`
}

type PXAPIResponse struct {
	SetID     string `json:"8ba6ptzbmw"`
	UUID      string `json:"S0MYYSz60P"`
	VID       string `json:"KjuYL7Ry1z"`
	UserAgent string `json:"c7AN5ynLe0"`
	PX3       string `json:"MUYKWy9bg6"`
}

type PXCapRequest struct {
	ActivationToken string `json:"wpo2IdNvaw"`
	HWID            string `json:"G4joAgkUe4"`
	DeviceName      string `json:"onYGxGEsj6"`
	Proxy           string `json:"uwwJXgoemC"`
	Site            string `json:"CgrnU15JBS"`
	SetID           string `json:"VSkzzFL5Nl"`
	UUID            string `json:"lWi3ipz90O"`
	VID             string `json:"3jEMiMn7XI"`
	Token           string `json:"tUaaAXr2qB"`
}

type PXCapResponse struct {
	Success      bool   `json:"GI4pD8JN8x"`
	PX3          string `json:"xaESu1lcdG"`
	ErrorMessage string `json:"0CpDnUbWEK"`
}

type EncryptedPXCapResponse struct {
	Success      string `json:"GI4pD8JN8x"`
	PX3          string `json:"xaESu1lcdG"`
	ErrorMessage string `json:"0CpDnUbWEK"`
}

type AkamaiRequest struct {
	ActivationToken string `json:"d4FmMviLeH"`
	HWID            string `json:"UIzN2pmSS4"`
	DeviceName      string `json:"QSyf9TPhgl"`
	PageURL         string `json:"BU6AK4vBSN"`
	SkipKact        string `json:"wqUSQ1OIb7"`
	SkipMact        string `json:"FCrYYORzgf"`
	OnBlur          string `json:"FIwDYyYcMF"`
	OnFocus         string `json:"xuvNcrzBpb"`
	Abck            string `json:"VihhvZ1c3w"`
	SensorDataLink  string `json:"kSJHn8HCYF"`
	Ver             string `json:"Di9bfAwAhv"`
	FirstPost       string `json:"FhIOsSwtOZ"`
	PixelID         string `json:"hC4JNGBKb4"`
	PixelG          string `json:"Lq7pGfPGm7"`
	JSON            string `json:"M9xdoAUOeE"`
}

type AkamaiResponse struct {
	Success           bool              `json:"krZupQuo9w"`
	AkamaiAPIResponse AkamaiAPIResponse `json:"xaESu1lcdG"`
	ErrorMessage      string            `json:"p1Lq0L1U7f"`
}

type EncryptedAkamaiResponse struct {
	Success           string            `json:"krZupQuo9w"`
	AkamaiAPIResponse AkamaiAPIResponse `json:"xaESu1lcdG"`
	ErrorMessage      string            `json:"p1Lq0L1U7f"`
}

type AkamaiAPIResponse struct {
	SensorData string `json:"9QuV8IpkJQ"`
	Pixel      string `json:"iqB8qihe20"`
}
