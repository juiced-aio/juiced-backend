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
	AYCD        CaptchaAPI = "AYCD"
)

type ReCaptchaSitekey = string

const (
	DisneySiteKey    ReCaptchaSitekey = "6Le2CasZAAAAAIVarP3wVo8isBezMJODg68gegRg"
	HotWheelsSitekey ReCaptchaSitekey = "6LeXJ7oUAAAAAHIpfRvgjs3lcJiO_zMC1LAZWlSz"
	ToppsSiteKey     ReCaptchaSitekey = "6LeBF1oaAAAAAOE7aQAZOLBjA1AVAYjVc9ulo4xh"
	WalmartSitekey   ReCaptchaSitekey = "6Lc8-RIaAAAAAPWSm2FVTyBg-Zkz2UjsWWfrkgYN"
)

var ReCaptchaSitekeys = map[Retailer]ReCaptchaSitekey{
	Disney:    DisneySiteKey,
	HotWheels: HotWheelsSitekey,
	Topps:     ToppsSiteKey,
	Walmart:   WalmartSitekey,
}

type HCaptchaSitekey = string

var HCaptchaSitekeys = map[Retailer]HCaptchaSitekey{}

type GeeTestCaptchaSitekey = string
