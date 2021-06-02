package hottopic

import (
	"net/http"
	"strings"

	"backend.juicedbot.io/m/v2/juiced.sitescripts/util"
)

// AddWalmartHeaders adds Walmart-specific headers to the request
func AddHottopicHeaders(request *http.Request, referer ...string) {
	util.AddBaseHeaders(request)
	request.Header.Set("Accept", "application/json")
	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("Origin", "https://www.walmart.com")
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
