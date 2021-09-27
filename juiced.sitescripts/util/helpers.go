package util

// // TernaryOperator is a make-shift ternary operator since Golang doesn't have one out of the box
// func TernaryOperator(condition bool, trueOutcome interface{}, falseOutcome interface{}) interface{} {
// 	if condition {
// 		return trueOutcome
// 	}
// 	return falseOutcome
// }

// func RandomNumberInt(min int, max int) int64 {
// 	rand.Seed(time.Now().UnixNano())
// 	a := int64(rand.Intn(max-min) + min)
// 	return a
// }

// func Randomizer(s string) string {
// 	nums := RandomNumberInt(0, 100)
// 	if nums < 50 {
// 		return s
// 	}
// 	return ""

// }

// // Function to generate valid abck cookies using an API
// func NewAbck(abckClient *http.Client, location string, BaseEndpoint, AkamaiEndpoint string) error {
// 	var ParsedBase, _ = url.Parse(BaseEndpoint)

// 	userInfo := stores.GetUserInfo()

// 	var abckCookie string
// 	var genResponse sec.ExperimentalAkamaiAPIResponse
// 	for _, cookie := range abckClient.Jar.Cookies(ParsedBase) {
// 		if cookie.Name == "_abck" {
// 			abckCookie = cookie.Value
// 		}
// 	}

// 	if abckCookie == "" {
// 		_, _, err := MakeRequest(&Request{
// 			Client: *abckClient,
// 			Method: "GET",
// 			URL:    AkamaiEndpoint,
// 			RawHeaders: [][2]string{
// 				{"sec-ch-ua", "\" Not A;Brand\";v=\"99\", \"Chromium\";v=\"90\", \"Google Chrome\";v=\"90\""},
// 				{"sec-ch-ua-mobile", "?0"},
// 				{"user-agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/90.0.4430.212 Safari/537.36"},
// 				{"content-type", "text/plain;charset=UTF-8"},
// 				{"accept", "*/*"},
// 				{"origin", BaseEndpoint},
// 				{"sec-fetch-site", "same-origin"},
// 				{"sec-fetch-mode", "cors"},
// 				{"sec-fetch-dest", "empty"},
// 				{"referer", location},
// 				{"accept-encoding", "gzip, deflate, br"},
// 				{"accept-language", "en-US,en;q=0.9"},
// 			},
// 		})
// 		if err != nil {
// 			return err
// 		}
// 	}

// 	genResponse, _, err := sec.ExperimentalAkamai(location, "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/92.0.4515.107 Safari/537.36", abckCookie, 0, 0, 0, 0, userInfo)
// 	if err != nil {
// 		return err
// 	}

// 	sensorRequest := SensorRequest{
// 		SensorData: genResponse.SensorData,
// 	}

// 	data, err := json.Marshal(sensorRequest)
// 	if err != nil {
// 		return err
// 	}

// 	sensorResponse := SensorResponse{}
// 	resp, _, err := MakeRequest(&Request{
// 		Client: *abckClient,
// 		Method: "POST",
// 		URL:    AkamaiEndpoint,
// 		RawHeaders: [][2]string{
// 			{"content-length", fmt.Sprint(len(data))},
// 			{"sec-ch-ua", "\" Not A;Brand\";v=\"99\", \"Chromium\";v=\"90\", \"Google Chrome\";v=\"90\""},
// 			{"sec-ch-ua-mobile", "?0"},
// 			{"user-agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/92.0.4515.107 Safari/537.36"},
// 			{"content-type", "text/plain;charset=UTF-8"},
// 			{"accept", "*/*"},
// 			{"origin", BaseEndpoint},
// 			{"sec-fetch-site", "same-origin"},
// 			{"sec-fetch-mode", "cors"},
// 			{"sec-fetch-dest", "empty"},
// 			{"referer", location},
// 			{"accept-encoding", "gzip, deflate, br"},
// 			{"accept-language", "en-US,en;q=0.9"},
// 		},
// 		Data:               data,
// 		ResponseBodyStruct: &sensorResponse,
// 	})
// 	if err != nil {
// 		return err
// 	}
// 	if !sensorResponse.Success || resp.StatusCode != 201 {
// 		return errors.New("bad sensor")
// 	}

// 	for _, cookie := range abckClient.Jar.Cookies(ParsedBase) {
// 		if cookie.Name == "_abck" {
// 			abckCookie = cookie.Value
// 		}
// 	}

// 	genResponse, _, err = sec.ExperimentalAkamai(location, "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/92.0.4515.107 Safari/537.36", abckCookie, 1, genResponse.SavedD3, genResponse.SavedStartTS, genResponse.DeviceNum, userInfo)
// 	if err != nil {
// 		return err
// 	}

// 	sensorRequest = SensorRequest{
// 		SensorData: genResponse.SensorData,
// 	}
// 	data, _ = json.Marshal(sensorRequest)

// 	sensorResponse = SensorResponse{}
// 	resp, _, err = MakeRequest(&Request{
// 		Client: *abckClient,
// 		Method: "POST",
// 		URL:    AkamaiEndpoint,
// 		RawHeaders: [][2]string{
// 			{"content-length", fmt.Sprint(len(data))},
// 			{"sec-ch-ua", "\" Not A;Brand\";v=\"99\", \"Chromium\";v=\"90\", \"Google Chrome\";v=\"90\""},
// 			{"sec-ch-ua-mobile", "?0"},
// 			{"user-agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/90.0.4430.212 Safari/537.36"},
// 			{"content-type", "text/plain;charset=UTF-8"},
// 			{"accept", "*/*"},
// 			{"origin", BaseEndpoint},
// 			{"sec-fetch-site", "same-origin"},
// 			{"sec-fetch-mode", "cors"},
// 			{"sec-fetch-dest", "empty"},
// 			{"referer", location},
// 			{"accept-encoding", "gzip, deflate, br"},
// 			{"accept-language", "en-US,en;q=0.9"},
// 		},
// 		Data:               data,
// 		ResponseBodyStruct: &sensorResponse,
// 	})
// 	if err != nil {
// 		return err
// 	}
// 	if !sensorResponse.Success {
// 		return errors.New("bad sensor")
// 	}

// 	switch resp.StatusCode {
// 	case 201:
// 		if ParsedBase.Host == "www.gamestop.com" {
// 			if len(abckCookie) > 488 {
// 				return nil
// 			}
// 		} else {
// 			for _, cookie := range abckClient.Jar.Cookies(ParsedBase) {
// 				if cookie.Name == "_abck" {
// 					validator, _ := FindInString(cookie.Value, "~", "~")
// 					if validator == "-1" {
// 						NewAbck(abckClient, location, BaseEndpoint, AkamaiEndpoint)
// 					}

// 				}
// 			}
// 		}

// 		return nil
// 	}
// 	return errors.New(resp.Status)
// }

// func GetPXCookie(site string, proxy *entities.Proxy, cancellationToken *CancellationToken) (string, PXValues, bool, error) {
// 	var pxValues PXValues

// 	userInfo := stores.GetUserInfo()

// 	pxResponse, _, err := sec.PX(site, ProxyCleaner(proxy), userInfo)
// 	if err != nil {
// 		return "", pxValues, false, err
// 	}

// 	if pxResponse.PX3 == "" || pxResponse.SetID == "" || pxResponse.UUID == "" || pxResponse.VID == "" {
// 		if cancellationToken.Cancel {
// 			return "", pxValues, true, err
// 		}
// 		return "", pxValues, false, errors.New("retry")
// 	}

// 	return pxResponse.PX3, PXValues{
// 		SetID: pxResponse.SetID,
// 		UUID:  pxResponse.UUID,
// 		VID:   pxResponse.VID,
// 	}, false, nil
// }

// func GetPXCapCookie(site, setID, vid, uuid, token string, proxy *entities.Proxy, cancellationToken *CancellationToken) (string, bool, error) {
// 	userInfo := stores.GetUserInfo()
// 	px3, _, err := sec.PXCap(site, ProxyCleaner(proxy), setID, vid, uuid, token, userInfo)
// 	if err != nil {
// 		return "", false, err
// 	}
// 	if px3 == "" {
// 		if cancellationToken.Cancel {
// 			return "", true, err
// 		}
// 		return "", false, errors.New("retry")
// 	}
// 	return px3, false, nil
// }

// // Returns the value of a cookie with the given cookieName and url
// func GetCookie(client http.Client, uri string, cookieName string) (string, error) {
// 	u, err := url.Parse(uri)
// 	if err != nil {
// 		return "", err
// 	}
// 	for _, cookie := range client.Jar.Cookies(u) {
// 		if cookie.Name == cookieName {
// 			return cookie.Value, nil
// 		}
// 	}
// 	return "", errors.New("no cookie with name: " + cookieName)
// }
