package newegg

import (
	"encoding/base64"
	"fmt"
	"math/rand"
	"strings"
	"time"

	"backend.juicedbot.io/juiced.client/http"
	"backend.juicedbot.io/juiced.sitescripts/util"
)

func BecomeGuest(client http.Client) bool {
	resp, _, err := util.MakeRequest(&util.Request{
		Client: client,
		Method: "GET",
		URL:    BaseEndpoint,
		RawHeaders: http.RawHeader{
			{"sec-ch-ua", `"Chromium";v="92", " Not A;Brand";v="99", "Google Chrome";v="92"`},
			{"sec-ch-ua-mobile", `?0`},
			{"upgrade-insecure-requests", `1`},
			{"user-agent", `Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/92.0.4515.131 Safari/537.36`},
			{"accept", `text/html,application/xhtml+xml,application/xml;q=0.9,image/avif,image/webp,image/apng,*/*;q=0.8,application/signed-exchange;v=b3;q=0.9`},
			{"sec-fetch-site", `none`},
			{"sec-fetch-mode", `navigate`},
			{"sec-fetch-user", `?1`},
			{"sec-fetch-dest", `document`},
			{"accept-encoding", `gzip, deflate, br`},
			{"accept-language", `en-US,en;q=0.9`},
		},
	})
	if resp.StatusCode != 200 || err != nil {
		return false
	}

	return true
}

func CreateExtras() (string, string) {
	nonceBytes := make([]byte, 16)
	_, err := rand.Read(nonceBytes)
	if err != nil {
		return "", ""
	}

	newSign := make([]byte, 16)
	_, err = rand.Read(newSign)
	if err != nil {
		return "", ""
	}

	params := util.CreateParams(map[string]string{
		"timestamp": fmt.Sprint(time.Now().Unix()),
		"nonce":     strings.ToLower(fmt.Sprintf("%X", nonceBytes)),
		"appId":     "107630",
	})

	return params, base64.StdEncoding.EncodeToString([]byte(strings.ToLower(fmt.Sprintf("%X", newSign))))

}
