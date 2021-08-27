package entities

// ReCaptchaToken stores information about a single ReCaptcha token
type ReCaptchaToken struct {
	URL   string
	Proxy Proxy
	Token string
}

// HCaptchaToken stores information about a single HCaptcha token
type HCaptchaToken struct {
	URL   string
	Proxy Proxy
	Token string
}

// GeeTestCaptchaToken stores information about a single GeeTestCaptcha token
type GeeTestCaptchaToken struct {
	URL   string
	Proxy Proxy
	Token GeeTestCaptchaTokenValues
}

type GeeTestCaptchaTokenValues struct {
	Challenge string
	Vaildate  string
	SecCode   string
}
