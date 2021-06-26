package enums

type CaptchaType = string

const (
	ReCaptchaV2    CaptchaType = "ReCaptchaV2"
	ReCaptchaV3    CaptchaType = "ReCaptchaV3"
	HCaptcha       CaptchaType = "HCaptcha"
	GeeTestCaptcha CaptchaType = "GeeTestCaptcha"
)

type CaptchaAPI = string

const (
	TwoCaptcha  CaptchaAPI = "2Captcha"
	AntiCaptcha CaptchaAPI = "AntiCaptcha"
	CapMonster  CaptchaAPI = "CapMonster"
)

type ReCaptchaSitekey = string

const (
	WalmartSitekey   ReCaptchaSitekey = "6Lc8-RIaAAAAAPWSm2FVTyBg-Zkz2UjsWWfrkgYN"
	HotWheelsSitekey ReCaptchaSitekey = "6LeXJ7oUAAAAAHIpfRvgjs3lcJiO_zMC1LAZWlSz"
)

var ReCaptchaSitekeys = map[Retailer]ReCaptchaSitekey{
	Walmart:   WalmartSitekey,
	HotWheels: HotWheelsSitekey,
}

type HCaptchaSitekey = string

var HCaptchaSitekeys = map[Retailer]HCaptchaSitekey{}

type GeeTestCaptchaSitekey = string
