package disney

import (
	"math/rand"

	"backend.juicedbot.io/juiced.client/http"
	u "backend.juicedbot.io/juiced.infrastructure/util"
)

func RandomString(selection string, length int) (r string) {
	for i := 0; i < length; i++ {
		r += string(selection[rand.Intn(len(selection))])
	}
	return
}

func IsValidAkamaiCookie(cookieJar http.CookieJar) bool {
	for _, cookie := range cookieJar.Cookies(BaseURL) {
		if cookie.Name == "_abck" {
			validator, _ := u.FindInString(cookie.Value, "~", "~")
			if validator == "-1" {
				return false
			} else {
				return true
			}
		}
	}
	return false
}
