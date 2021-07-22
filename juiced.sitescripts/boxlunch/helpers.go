package boxlunch

import (
	"backend.juicedbot.io/juiced.client/http"
	"backend.juicedbot.io/juiced.infrastructure/common"
	"backend.juicedbot.io/juiced.sitescripts/util"
)

// AddBoxlunch adds Boxlunch-specific headers to the request
func AddBoxlunchHeaders(request *http.Request, referer ...string) {
	util.AddBaseHeaders(request)
	request.Header.Set("Accept", "*/*")
	request.Header.Set("Content-Type", "application/x-www-form-urlencoded; charset=UTF-8")
	request.Header.Set("Origin", "https://www.boxlunch.com")
	// omitcsrfjwt: true
	// omitcorrelationid: true
	// credentials: include
	// TODO: Header order
	if len(referer) != 0 {
		request.Header.Set("Referer", referer[0])
	}
}

func getDwCont(body string) (string, error) {
	return common.FindInString(body, "cart?dwcont=", "\"")
}

func getSecureKey(body string) (string, error) {
	return common.FindInString(body, "_securekey\" value=\"", "\"/>")
}
