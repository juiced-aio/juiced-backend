package hottopic

import (
	"strings"

	"backend.juicedbot.io/juiced.client/http"
	"backend.juicedbot.io/juiced.sitescripts/util"
)

// AddHottopicHeaders adds Hottopic-specific headers to the request
func AddHottopicHeaders(request *http.Request, referer ...string) {
	util.AddBaseHeaders(request)
	request.Header.Set("Accept", "*/*")
	request.Header.Set("Content-Type", "application/x-www-form-urlencoded; charset=UTF-8")
	request.Header.Set("Origin", "https://www.hottopic.com")
	// omitcsrfjwt: true
	// omitcorrelationid: true
	// credentials: include
	// TODO: Header order
	if len(referer) != 0 {
		request.Header.Set("Referer", referer[0])
	}
}

func getDwCont(resp string) string {
	return getStringInBetweenTwoString(resp, "cart?dwcont=", "\" method")
}

func getSecureKey(resp string) string {
	return getStringInBetweenTwoString(resp, "_securekey\" value=\"", "\"/>")
}

func getStringInBetweenTwoString(str string, startS string, endS string) (result string) {
	s := strings.Index(str, startS)
	if s == -1 {
		return result
	}
	newS := str[s+len(startS):]
	e := strings.Index(newS, endS)
	if e == -1 {
		return result
	}
	result = newS[:e]
	return result
}
