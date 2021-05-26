package util

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

type ActivateResponse struct {
	Success              bool                      `json:"BRdW40qYEd"`
	LicensesToDeactivate []LicenseToDeactivateInfo `json:"izz9E9YoY4"`
	Email                string                    `json:"CdpO887IuH"`
	LicenseKey           string                    `json:"q5izDieCbb"`
	DiscordID            string                    `json:"sQzlETuNin"`
	DiscordUsername      string                    `json:"1x16bw24sz"`
	DiscordAvatarURL     string                    `json:"9WMNTAhnxb"`
	ActivationToken      string                    `json:"lUiFoCyvqa"`
	RefreshToken         string                    `json:"qM3DKGrRJE"`
	ExpiresAt            int64                     `json:"uYIvVc1ojh"`
	ErrorMessage         string                    `json:"4gwjptUGmw"`
}

type EncryptedActivateResponse struct {
	Success              string                    `json:"BRdW40qYEd"`
	LicensesToDeactivate []LicenseToDeactivateInfo `json:"izz9E9YoY4"`
	Email                string                    `json:"CdpO887IuH"`
	LicenseKey           string                    `json:"q5izDieCbb"`
	DiscordID            string                    `json:"sQzlETuNin"`
	DiscordUsername      string                    `json:"1x16bw24sz"`
	DiscordAvatarURL     string                    `json:"9WMNTAhnxb"`
	ActivationToken      string                    `json:"lUiFoCyvqa"`
	RefreshToken         string                    `json:"qM3DKGrRJE"`
	ExpiresAt            string                    `json:"uYIvVc1ojh"`
	ErrorMessage         string                    `json:"4gwjptUGmw"`
}

type LicenseToDeactivateInfo struct {
	LicenseKey  string `json:"6IBEpoL8u6"`
	LicenseType string `json:"mtAeKfQ8J8"`
	DeviceName  string `json:"5YtIwrXOHR"`
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

type DownloadBackendRequest struct {
	DiscordID       string `json:"DzNq85M9Sp"`
	HWID            string `json:"1p669XAC1A"`
	DeviceName      string `json:"nR4YugfUUW"`
	ActivationToken string `json:"xh70lJzPxO"`
}

type AuthenticationResult = int

const (
	ERROR_AUTHENTICATE_HWID AuthenticationResult = iota
	ERROR_AUTHENTICATE_CREATE_IV
	ERROR_AUTHENTICATE_ENCRYPT_TIMESTAMP
	ERROR_AUTHENTICATE_ENCRYPT_ACTIVATION_TOKEN
	ERROR_AUTHENTICATE_ENCRYPT_HWID
	ERROR_AUTHENTICATE_ENCRYPT_DEVICE_NAME
	ERROR_AUTHENTICATE_ENCRYPT_HEADER_A
	ERROR_AUTHENTICATE_ENCRYPT_HEADER_B
	ERROR_AUTHENTICATE_ENCRYPT_HEADER_C
	ERROR_AUTHENTICATE_ENCRYPT_HEADER_D
	ERROR_AUTHENTICATE_ENCRYPT_HEADER_E
	ERROR_AUTHENTICATE_REQUEST
	ERROR_AUTHENTICATE_READ_BODY
	ERROR_AUTHENTICATE_DECRYPT_RESPONSE
	ERROR_AUTHENTICATE_TOKEN_EXPIRED
	ERROR_AUTHENTICATE_FAILED
	SUCCESS_AUTHENTICATE
)

type ActivationResult = int

const (
	ERROR_ACTIVATE_HWID ActivationResult = iota
	ERROR_ACTIVATE_CREATE_IV
	ERROR_ACTIVATE_ENCRYPT_TIMESTAMP
	ERROR_ACTIVATE_ENCRYPT_EMAIL
	ERROR_ACTIVATE_ENCRYPT_PASSWORD
	ERROR_ACTIVATE_ENCRYPT_HWID
	ERROR_ACTIVATE_ENCRYPT_DEVICE_NAME
	ERROR_ACTIVATE_REQUEST
	ERROR_ACTIVATE_READ_BODY
	ERROR_ACTIVATE_DECRYPT_RESPONSE
	ERROR_ACTIVATE_FAILED
	SUCCESS_ACTIVATE_ERROR_SET_USER_INFO
	SUCCESS_ACTIVATE
)

type RefreshResult = int

const (
	ERROR_REFRESH_HWID RefreshResult = iota
	ERROR_REFRESH_CREATE_IV
	ERROR_REFRESH_ENCRYPT_TIMESTAMP
	ERROR_REFRESH_ENCRYPT_ACTIVATION_TOKEN
	ERROR_REFRESH_ENCRYPT_REFRESH_TOKEN
	ERROR_REFRESH_ENCRYPT_HWID
	ERROR_REFRESH_ENCRYPT_DEVICE_NAME
	ERROR_REFRESH_ENCRYPT_HEADER_A
	ERROR_REFRESH_ENCRYPT_HEADER_B
	ERROR_REFRESH_ENCRYPT_HEADER_C
	ERROR_REFRESH_ENCRYPT_HEADER_D
	ERROR_REFRESH_ENCRYPT_HEADER_E
	ERROR_REFRESH_REQUEST
	ERROR_REFRESH_READ_BODY
	ERROR_REFRESH_DECRYPT_RESPONSE
	ERROR_REFRESH_FAILED
	SUCCESS_REFRESH_ERROR_SET_USER_INFO
	SUCCESS_REFRESH
)

type AuthErrorCode = int

const (
	SUCCESS AuthErrorCode = iota
	NO_STORED_INFO
	ERROR_CONNECTING_TO_DATABASE
	ERROR_AUTHENTICATING_EXISTING_INFO
	ERROR_READING_REQUEST_BODY
	ERROR_BEFORE_REQUEST
	ERROR_DURING_REQUEST
	ERROR_AFTER_REQUEST
)
