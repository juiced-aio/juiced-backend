package endpoints

// import (
// 	"encoding/json"
// 	"io/ioutil"
// 	"net/http"
// 	"sync"
// 	"time"

// 	"backend.juicedbot.io/juiced.api/errors"
// 	"backend.juicedbot.io/juiced.api/responses"
// 	rpc "backend.juicedbot.io/juiced.rpc"
// 	"backend.juicedbot.io/juiced.sitescripts/util"
// )

// func TestWebhooksEndpoint(response http.ResponseWriter, request *http.Request) {
// 	response.Header().Set("content-type", "application/json")
// 	response.Header().Set("Access-Control-Allow-Origin", "http://localhost:3000")
// 	errorsList := make([]string, 0)

// 	type TestWebhooksRequest struct {
// 		SuccessWebhook string `json:"successDiscordWebhook"`
// 		FailureWebhook string `json:"failureDiscordWebhook"`
// 	}

// 	embed := util.Embed{
// 		Footer: util.Footer{
// 			Text:    "Juiced AIO",
// 			IconURL: "https://media.discordapp.net/attachments/849430464036077598/855979506204278804/Icon_1.png?width=128&height=128",
// 		},
// 		Timestamp: time.Now(),
// 	}

// 	body, err := ioutil.ReadAll(request.Body)
// 	if err == nil {
// 		testWebhooksRequest := TestWebhooksRequest{}
// 		err = json.Unmarshal(body, &testWebhooksRequest)
// 		if err == nil {
// 			wg := sync.WaitGroup{}
// 			wg.Add(2)
// 			go func() {
// 				if testWebhooksRequest.SuccessWebhook != "" {
// 					embed.Title = "Success Webhook"
// 					embed.Color = 16742912
// 					if !util.SendDiscordWebhook(testWebhooksRequest.SuccessWebhook, []util.Embed{embed}) {
// 						errorsList = append(errorsList, errors.TestSuccessWebhookError)
// 					}
// 				}
// 				wg.Done()
// 			}()
// 			go func() {
// 				if testWebhooksRequest.FailureWebhook != "" {
// 					embed.Title = "Failure Webhook"
// 					embed.Color = 14495044
// 					if !util.SendDiscordWebhook(testWebhooksRequest.FailureWebhook, []util.Embed{embed}) {
// 						errorsList = append(errorsList, errors.TestFailureWebhookError)
// 					}
// 				}
// 				wg.Done()
// 			}()
// 			wg.Wait()
// 		}
// 	} else {
// 		errorsList = append(errorsList, errors.IOUtilReadAllError+err.Error())
// 	}
// 	result := &responses.MiscellaneousResponse{Success: true, Errors: make([]string, 0)}
// 	if len(errorsList) > 0 {
// 		response.WriteHeader(http.StatusBadRequest)
// 		result = &responses.MiscellaneousResponse{Success: false, Errors: errorsList}
// 	}
// 	json.NewEncoder(response).Encode(result)
// }

// func SetVersion(response http.ResponseWriter, request *http.Request) {
// 	response.Header().Set("content-type", "application/json")
// 	response.Header().Set("Access-Control-Allow-Origin", "http://localhost:3000")
// 	errorsList := make([]string, 0)

// 	type SetVersionRequest struct {
// 		Version string `json:"version"`
// 	}

// 	body, err := ioutil.ReadAll(request.Body)
// 	if err == nil {
// 		setVersionRequest := SetVersionRequest{}
// 		err = json.Unmarshal(body, &setVersionRequest)
// 		if err == nil {
// 			rpc.SetActivity(setVersionRequest.Version)
// 		}
// 	} else {
// 		errorsList = append(errorsList, errors.IOUtilReadAllError+err.Error())
// 	}

// 	result := &responses.MiscellaneousResponse{Success: true, Errors: make([]string, 0)}
// 	if len(errorsList) > 0 {
// 		response.WriteHeader(http.StatusBadRequest)
// 		result = &responses.MiscellaneousResponse{Success: false, Errors: errorsList}
// 	}
// 	json.NewEncoder(response).Encode(result)
// }
