package logging

import (
	"backend.juicedbot.io/juiced.client/http"
)

func compareHeaders(h1, h2 http.Header) http.Header {
	checker := make(map[string]string)
	h3 := http.Header{}
	for key, value := range h1 {
		checker[key] = value[0]
	}
	for key, value := range h2 {
		header, ok := checker[key]
		if ok {
			if value[0] != header {
				h3[key] = value
			}
		} else {
			h3[key] = value
		}
	}

	return h3
}

func compareRawHeaders(h1, h2 http.RawHeader) http.RawHeader {
	checker := make(map[string]string)
	h3 := http.RawHeader{}
	for _, value1 := range h1 {
		checker[value1[0]] = value1[1]
	}
	for _, value2 := range h2 {
		header, ok := checker[value2[0]]
		if ok {
			if value2[1] != header {
				h3 = append(h3, value2)
			}
		} else {
			h3 = append(h3, value2)
		}
	}

	return h3
}
