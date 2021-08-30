package cloudflare

import (
	"time"

	"backend.juicedbot.io/juiced.client/http"
)

type Scraper struct {
	Client                         http.Client
	CaptchaFunction                func(originalURL string, siteKey string) (string, error)
	FingerprintChallenge           bool
	Script                         string
	InitScript                     *http.Response
	ChallengePayload               *http.Response
	MainPayloadResponse            *http.Response
	InitURL                        string
	RequestURL                     string
	CaptchaScript                  string
	ApiDomain                      string
	TimeOut                        int
	ErrorDelay                     int
	InitHeaders                    http.RawHeader
	ChallengeHeaders               http.RawHeader
	SubmitHeaders                  http.RawHeader
	OriginalRequest                *http.Response
	Domain                         string
	Debug                          bool
	Key                            string
	AuthParams                     map[string]string
	Md                             string
	Captcha                        bool
	StartTime                      time.Time
	SolveRetries                   int
	SolveMaxRetries                int
	Result                         string
	Name                           string
	BaseObj                        string
	RequestPass                    string
	RequestR                       string
	TS                             int
	TargetURL                      string
	InitPayloadRetries             int
	InitPayloadMaxRetries          int
	KeyStrUriSafe                  string
	InitChallengeRetries           int
	InitChallengeMaxRetries        int
	FetchingChallengeRetries       int
	FetchingChallengeMaxRetries    int
	SubmitChallengeRetries         int
	SubmitChallengeMaxRetries      int
	ChallengeResultRetries         int
	ChallengeResultMaxRetries      int
	FinalApi                       apiResponse
	SubmitFinalChallengeRetries    int
	SubmitFinalChallengeMaxRetries int
	RerunRetries                   int
	RerunMaxRetries                int
	CaptchaRetries                 int
	CaptchaMaxRetries              int
	FirstCaptchaResult             apiResponse
	CaptchaResponseAPI             apiResponse
	SubmitCaptchaRetries           int
	SubmitCaptchaMaxRetries        int
}

type apiResponse struct {
	URL          string `json:"url"`
	ResultURL    string `json:"result_url"`
	Result       string `json:"result"`
	Name         string `json:"name"`
	BaseObj      string `json:"baseobj"`
	Pass         string `json:"pass"`
	R            string `json:"r"`
	TS           int    `json:"ts"`
	Md           string `json:"md"`
	Status       string `json:"status"`
	Captcha      bool   `json:"captcha"`
	JschlVc      string `json:"jschl_vc"`
	JschlAnswer  string `json:"jschl_answer"`
	CfChCpReturn string `json:"cf_ch_cp_return"`
	SiteKey      string `json:"sitekey"`
	Click        bool   `json:"click"`
	Valid        bool   `json:"valid"`
}
