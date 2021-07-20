package boxlunch

import (
	"backend.juicedbot.io/juiced.client/http"
	"backend.juicedbot.io/juiced.infrastructure/common"
	"backend.juicedbot.io/juiced.sitescripts/util"
)

// AddBoxLunch adds BoxLunch-specific headers to the request
func AddBoxLunchHeaders(request *http.Request, referer ...string) {
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

func getDwCont(resp string) (string, error) {
	return common.FindInString(resp, "cart?dwcont=", "\" method")
}

func getSecureKey(resp string) (string, error) {
	return common.FindInString(resp, "_securekey\" value=\"", "\"/>")
}
