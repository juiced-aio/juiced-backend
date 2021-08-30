package cloudflare

import (
	"bytes"
	"encoding/json"
	"io"
	"io/ioutil"
	"log"
	"regexp"
	"strings"
	"time"

	"backend.juicedbot.io/juiced.client/http"

	"github.com/anaskhan96/soup"
)

func ReadAndCopyBody(response *http.Response) ([]byte, error) {
	var body []byte
	var err error
	var b bytes.Buffer
	t := io.TeeReader(response.Body, &b)
	body, err = ioutil.ReadAll(t)
	if err != nil {
		return body, err
	}
	response.Body = ioutil.NopCloser(bytes.NewBuffer(body))
	return body, err
}

func ReadAndUnmarshalBody(respBody io.ReadCloser, x interface{}) error {
	body, err := ioutil.ReadAll(respBody)
	if err != nil {
		return err
	}
	err = json.Unmarshal(body, &x)
	return err
}

func CheckForCaptcha(body string) bool {
	doc := soup.HTMLParse(body)

	element := doc.Find("input", "name", "cf_captcha_kind")
	if element.Error != nil {
		return false
	}
	if val, ok := element.Attrs()["value"]; ok && val == "h" {
		return true
	}

	return false
}

func IsNewIUAMChallenge(response *http.Response) bool {
	body, err := ReadAndCopyBody(response)
	if err != nil {
		return false
	}
	firstReg, err := regexp.MatchString(`cpo.src\s*=\s*"/cdn-cgi/challenge-platform/?\w?/?\w?/orchestrate/jsch/v1`, string(body))
	if err != nil {
		return false
	}
	secondReg, err := regexp.MatchString(`window._cf_chl_opt`, string(body))
	if err != nil {
		return false
	}
	return strings.Contains(response.Header.Get("Server"), "cloudflare") &&
		(response.StatusCode == 429 || response.StatusCode == 503) &&
		firstReg && secondReg

}

func IsFingerprintChallenge(response *http.Response) bool {
	if response.StatusCode == 429 {
		body, err := ReadAndCopyBody(response)
		if err != nil {
			return false
		}
		if strings.Contains(string(body), "/fingerprint/script/") {
			return true
		}

	}
	return false
}

func IsNewCaptchaChallenge(response *http.Response) bool {
	body, err := ReadAndCopyBody(response)
	if err != nil {
		return false
	}
	firstReg, err := regexp.MatchString(`cpo.src\s*=\s*"/cdn-cgi/challenge-platform/?\w?/?\w?/orchestrate/.*/v1`, string(body))
	if err != nil {
		return false
	}
	secondReg, err := regexp.MatchString(`window._cf_chl_opt`, string(body))
	if err != nil {
		return false
	}
	return strings.Contains(response.Header.Get("Server"), "cloudflare") &&
		(response.StatusCode == 403) &&
		firstReg && secondReg
}

func (scraper *Scraper) HandleLoopError(errFormat string, err error) {
	if scraper.Debug {
		log.Printf(errFormat, err.Error())
	}
	time.Sleep(time.Duration(scraper.ErrorDelay) * time.Second)
}

func CreateInitHeaders(ua string) http.RawHeader {
	return http.RawHeader{
		{"sec-ch-ua", `" Not;A Brand";v="99", "Google Chrome";v="91", "Chromium";v="91"`},
		{"sec-ch-ua-mobile", "?0"},
		{"upgrade-insecure-requests", "1"},
		{"user-agent", ua},
		{"accept", "text/html,application/xhtml+xml,application/xml;q=0.9,image/avif,image/webp,image/apng,*/*;q=0.8,application/signed-exchange;v=b3;q=0.9"},
		{"sec-fetch-site", "none"},
		{"sec-fetch-mode", "navigate"},
		{"sec-fetch-user", "?1"},
		{"sec-fetch-dest", "document"},
		{"accept-encoding", "gzip, deflate"},
		{"accept-language", "en-US,en;q=0.9"},
	}
}

func CreateChallengeHeaders(cl, ua, challenge, origin, referer string) http.RawHeader {
	return http.RawHeader{
		{"content-length", cl},
		{"sec-ch-ua", `"Chromium";v="92", " Not A;Brand";v="99", "Google Chrome";v="92"`},
		{"sec-ch-ua-mobile", `?0`},
		{"user-agent", ua},
		{"cf-challenge", challenge},
		{"content-type", `application/x-www-form-urlencoded`},
		{"accept", `*/*`},
		{"origin", origin},
		{"sec-fetch-site", `same-origin`},
		{"sec-fetch-mode", `cors`},
		{"sec-fetch-dest", `empty`},
		{"referer", referer},
		{"accept-encoding", `gzip, deflate`},
		{"accept-language", `en-US,en;q=0.9`},
	}
}

func CreateSubmitHeaders(cl, origin, ua, referer string) http.RawHeader {
	return http.RawHeader{
		{"content-length", cl},
		{"cache-control", `max-age=0`},
		{"sec-ch-ua", `"Chromium";v="92", " Not A;Brand";v="99", "Google Chrome";v="92"`},
		{"sec-ch-ua-mobile", `?0`},
		{"upgrade-insecure-requests", `1`},
		{"origin", origin},
		{"content-type", `application/x-www-form-urlencoded`},
		{"user-agent", ua},
		{"accept", `text/html,application/xhtml+xml,application/xml;q=0.9,image/avif,image/webp,image/apng,*/*;q=0.8,application/signed-exchange;v=b3;q=0.9`},
		{"sec-fetch-site", `same-origin`},
		{"sec-fetch-mode", `navigate`},
		{"sec-fetch-dest", `document`},
		{"referer", referer},
		{"accept-encoding", `gzip, deflate`},
		{"accept-language", `en-US,en;q=0.9`},
	}
}
