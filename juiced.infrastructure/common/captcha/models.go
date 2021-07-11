package captcha

type KeyError = string

const (
	CaptchaSuccess          KeyError = "CAPTCHA_SUCCESS"
	BadTwoCapKeyError       KeyError = "BAD_2CAP_KEY"
	TwoCapZeroBalanceError  KeyError = "2CAP_ZERO_BAL"
	BadAntiCapKeyError      KeyError = "BAD_ANTICAP_KEY"
	AntiCapZeroBalanceError KeyError = "ANTICAP_ZERO_BAL"
	BadCapMonKeyError       KeyError = "BAD_CAPMON_KEY"
	CapMonZeroBalanceError  KeyError = "CAPMON_ZERO_BAL"
)

type AntiCaptchaTask struct {
	Clientkey string              `json:"clientKey"`
	Task      AntiCaptchaTaskInfo `json:"task"`
}
type AntiCaptchaTaskInfo struct {
	Type          string  `json:"type"`
	Websiteurl    string  `json:"websiteURL"`
	Gt            string  `json:"gt"`
	Challenge     string  `json:"challenge"`
	Websitekey    string  `json:"websiteKey"`
	MinScore      float64 `json:"minScore"`
	PageAction    string  `json:"pageAction"`
	IsEnterprise  bool    `json:"isEnterprise"`
	Proxytype     string  `json:"proxyType"`
	Proxyaddress  string  `json:"proxyAddress"`
	Proxyport     int     `json:"proxyPort"`
	Proxylogin    string  `json:"proxyLogin"`
	Proxypassword string  `json:"proxyPassword"`
	Useragent     string  `json:"userAgent"`
	Cookie        string  `json:"cookie"`
}

type AntiCaptchaStart struct {
	TaskID           int    `json:"taskId"`
	Errorid          int    `json:"errorId"`
	ErrorCode        string `json:"errorCode"`
	Errordescription string `json:"errorDescription"`
}

type AntiCaptchaRequest struct {
	Clientkey string `json:"clientKey"`
	Taskid    int    `json:"taskId"`
}

type AntiCaptchaResponse struct {
	Errorid    int      `json:"errorId"`
	ErrorCode  string   `json:"errorCode"`
	Status     string   `json:"status"`
	Solution   Solution `json:"solution"`
	Cost       string   `json:"cost"`
	IP         string   `json:"ip"`
	Createtime int      `json:"createTime"`
	Endtime    int      `json:"endTime"`
	Solvecount int      `json:"solveCount"`
}
type Solution struct {
	GRecaptchaResponse string `json:"gRecaptchaResponse"`
	Challenge          string `json:"challenge"`
	Validate           string `json:"validate"`
	Seccode            string `json:"seccode"`
}

type CapMonsterTask struct {
	Clientkey string             `json:"clientKey"`
	Task      CapMonsterTaskInfo `json:"task"`
}
type CapMonsterTaskInfo struct {
	Type          string  `json:"type"`
	Websiteurl    string  `json:"websiteURL"`
	Websitekey    string  `json:"websiteKey"`
	MinScore      float64 `json:"minScore"`
	PageAction    string  `json:"pageAction"`
	Proxytype     string  `json:"proxyType"`
	Proxyaddress  string  `json:"proxyAddress"`
	Proxyport     int     `json:"proxyPort"`
	Proxylogin    string  `json:"proxyLogin"`
	Proxypassword string  `json:"proxyPassword"`
	Useragent     string  `json:"userAgent"`
}

type CapMonsterStart struct {
	TaskID           int    `json:"taskId"`
	Errorid          int    `json:"errorId"`
	Errorcode        string `json:"errorCode"`
	Errordescription string `json:"errorDescription"`
}

type CapMonsterRequest struct {
	Clientkey string `json:"clientKey"`
	Taskid    int    `json:"taskId"`
}

type CapMonsterResponse struct {
	Errorid   int      `json:"errorId"`
	ErrorCode string   `json:"errorCode"`
	Status    string   `json:"status"`
	Solution  Solution `json:"solution"`
}

type TwoCaptchaGeeTestResponse struct {
	Challenge string `json:"challenge"`
	Validate  string `json:"validate"`
	Seccode   string `json:"seccode"`
}
