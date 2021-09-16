package hottopic

import (
	"backend.juicedbot.io/juiced.client/http"
	u "backend.juicedbot.io/juiced.infrastructure/util"
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

func getDwCont(body string) (string, error) {
	return u.FindInString2(body, "cart?dwcont=", `"`)
}

func getSecureKey(body string) (string, error) {
	return u.FindInString2(body, `_securekey" value="`, `"`)
}
