package endpoints

import (
	"log"
	"time"

	"backend.juicedbot.io/juiced.api/responses"
	"backend.juicedbot.io/juiced.infrastructure/commands"
	"backend.juicedbot.io/juiced.infrastructure/common"
	"backend.juicedbot.io/juiced.infrastructure/common/captcha"
	"backend.juicedbot.io/juiced.infrastructure/common/entities"
	"backend.juicedbot.io/juiced.infrastructure/common/errors"
	"backend.juicedbot.io/juiced.infrastructure/queries"
	"github.com/google/uuid"
	"github.com/gorilla/mux"

	"encoding/json"
	"io/ioutil"
	"net/http"
)

// GetSettingsEndpoint handles the GET request at /api/settings
func GetSettingsEndpoint(response http.ResponseWriter, request *http.Request) {
	response.Header().Set("content-type", "application/json")
	response.Header().Set("Access-Control-Allow-Origin", "http://localhost:3000")
	var settings entities.Settings
	errorsList := make([]string, 0)

	settings, err := queries.GetSettings()
	if err != nil {
		errorsList = append(errorsList, errors.GetSettingsError+err.Error())
	}
	result := &responses.SettingsResponse{Success: true, Data: settings, Errors: make([]string, 0)}
	if len(errorsList) > 0 {
		response.WriteHeader(http.StatusBadRequest)
		result = &responses.SettingsResponse{Success: false, Data: entities.Settings{}, Errors: errorsList}
	}
	json.NewEncoder(response).Encode(result)
}

// UpdateSettingsEndpoint handles the PUT request at /api/settings
func UpdateSettingsEndpoint(response http.ResponseWriter, request *http.Request) {
	response.Header().Set("content-type", "application/json")
	response.Header().Set("Access-Control-Allow-Origin", "http://localhost:3000")
	var newSettings entities.Settings
	errorsList := make([]string, 0)

	body, err := ioutil.ReadAll(request.Body)
	if err == nil {
		err = entities.ParseSettings(&newSettings, body)
		if err == nil {
			currentSettings, err := queries.GetSettings()
			if err == nil {
				if newSettings.SuccessDiscordWebhook == "-1" {
					newSettings.SuccessDiscordWebhook = currentSettings.SuccessDiscordWebhook
				}
				if newSettings.FailureDiscordWebhook == "-1" {
					newSettings.FailureDiscordWebhook = currentSettings.FailureDiscordWebhook
				}
				if newSettings.TwoCaptchaAPIKey == "-1" {
					newSettings.TwoCaptchaAPIKey = currentSettings.TwoCaptchaAPIKey
				}
				if newSettings.AntiCaptchaAPIKey == "-1" {
					newSettings.AntiCaptchaAPIKey = currentSettings.AntiCaptchaAPIKey
				}
				if newSettings.CapMonsterAPIKey == "-1" {
					newSettings.CapMonsterAPIKey = currentSettings.CapMonsterAPIKey
				}
				aycdChanged := false
				if newSettings.AYCDAccessToken == "-1" {
					newSettings.AYCDAccessToken = currentSettings.AYCDAccessToken
				} else {
					aycdChanged = true
				}
				if newSettings.AYCDAPIKey == "-1" {
					newSettings.AYCDAPIKey = currentSettings.AYCDAPIKey
				} else {
					aycdChanged = true
				}
				if !newSettings.DarkModeUpdate {
					newSettings.DarkMode = currentSettings.DarkMode
				}
				if !newSettings.UseAnimationsUpdate {
					newSettings.UseAnimations = currentSettings.UseAnimations
				}
				newSettings, err = commands.UpdateSettings(newSettings)
				if err != nil {
					errorsList = append(errorsList, errors.UpdateSettingsError+err.Error())
				} else {
					if aycdChanged && newSettings.AYCDAccessToken != "" && newSettings.AYCDAPIKey != "" {
						err = captcha.ConnectToAycd(newSettings.AYCDAccessToken, newSettings.AYCDAPIKey)
						if err != nil {
							// TODO @silent: Handle
							log.Println("Error connecting to AYCD with new credentials: " + err.Error())
						} else {
							log.Println("Connected to AYCD with new credentials.")
						}
					}
				}
			} else {
				errorsList = append(errorsList, errors.GetSettingsError+err.Error())
			}
		} else {
			errorsList = append(errorsList, errors.ParseSettingsError+err.Error())
		}
	} else {
		errorsList = append(errorsList, errors.IOUtilReadAllError+err.Error())
	}
	result := &responses.SettingsResponse{Success: true, Data: newSettings, Errors: make([]string, 0)}
	if len(errorsList) > 0 {
		response.WriteHeader(http.StatusBadRequest)
		result = &responses.SettingsResponse{Success: false, Data: entities.Settings{}, Errors: errorsList}
	}
	json.NewEncoder(response).Encode(result)
}

// AddAccountEndpoint handles the POST request at /api/settings/accounts
func AddAccountEndpoint(response http.ResponseWriter, request *http.Request) {
	response.Header().Set("content-type", "application/json")
	response.Header().Set("Access-Control-Allow-Origin", "http://localhost:3000")
	var settings entities.Settings
	var newAccount entities.Account
	errorsList := make([]string, 0)

	body, err := ioutil.ReadAll(request.Body)
	if err == nil {
		settings, err = queries.GetSettings()
		if err == nil {
			err = json.Unmarshal(body, &newAccount)
			if err == nil {
				newAccount.ID = uuid.New().String()
				newAccount.CreationDate = time.Now().Unix()
				err = commands.AddAccount(newAccount)
				if err == nil {
					settings.Accounts = append(settings.Accounts, newAccount)
				} else {
					errorsList = append(errorsList, errors.AddAccountError+err.Error())
				}
			} else {
				errorsList = append(errorsList, errors.ParseAccountError+err.Error())
			}
		} else {
			errorsList = append(errorsList, errors.GetSettingsError+err.Error())
		}
	} else {
		errorsList = append(errorsList, errors.IOUtilReadAllError+err.Error())
	}
	result := &responses.SettingsResponse{Success: true, Data: settings, Errors: make([]string, 0)}
	if len(errorsList) > 0 {
		response.WriteHeader(http.StatusBadRequest)
		result = &responses.SettingsResponse{Success: false, Data: entities.Settings{}, Errors: errorsList}
	}
	json.NewEncoder(response).Encode(result)
}

// UpdateAccountEndpoint handles the PUT request at /api/settings/accounts/{ID}
func UpdateAccountEndpoint(response http.ResponseWriter, request *http.Request) {
	response.Header().Set("content-type", "application/json")
	response.Header().Set("Access-Control-Allow-Origin", "http://localhost:3000")
	var settings entities.Settings
	var newAccount entities.Account
	errorsList := make([]string, 0)

	params := mux.Vars(request)
	ID, ok := params["ID"]
	if ok {
		body, err := ioutil.ReadAll(request.Body)
		if err == nil {
			settings, err = queries.GetSettings()
			if err == nil {
				err = json.Unmarshal(body, &newAccount)
				if err == nil {
					newAccounts := []entities.Account{}
					for _, account := range settings.Accounts {
						if account.ID == ID {
							_, err = commands.UpdateAccount(ID, newAccount)
							if err != nil {
								errorsList = append(errorsList, errors.UpdateAccountError+err.Error())
							}
							newAccounts = append(newAccounts, newAccount)
						} else {
							newAccounts = append(newAccounts, account)
						}
					}
					settings.Accounts = newAccounts
				} else {
					errorsList = append(errorsList, errors.ParseAccountError+err.Error())
				}
			} else {
				errorsList = append(errorsList, errors.GetSettingsError+err.Error())
			}
		} else {
			errorsList = append(errorsList, errors.IOUtilReadAllError+err.Error())
		}
	} else {
		errorsList = append(errorsList, errors.MissingParameterError)
	}
	result := &responses.SettingsResponse{Success: true, Data: settings, Errors: make([]string, 0)}
	if len(errorsList) > 0 {
		response.WriteHeader(http.StatusBadRequest)
		result = &responses.SettingsResponse{Success: false, Data: entities.Settings{}, Errors: errorsList}
	}
	json.NewEncoder(response).Encode(result)
}

// RemoveAccountsEndpoint handles the POST request at /api/settings/accounts/remove
func RemoveAccountsEndpoint(response http.ResponseWriter, request *http.Request) {
	response.Header().Set("content-type", "application/json")
	response.Header().Set("Access-Control-Allow-Origin", "http://localhost:3000")
	var settings entities.Settings
	errorsList := make([]string, 0)

	type DeleteAccountsRequest struct {
		AccountIDs []string `json:"accountIDs"`
	}

	body, err := ioutil.ReadAll(request.Body)
	if err == nil {
		settings, err = queries.GetSettings()
		if err == nil {
			deleteAccountsRequestInfo := DeleteAccountsRequest{}
			err = json.Unmarshal(body, &deleteAccountsRequestInfo)
			if err == nil {
				newAccounts := []entities.Account{}
				for _, account := range settings.Accounts {
					if common.InSlice(deleteAccountsRequestInfo.AccountIDs, account.ID) {
						_, err = commands.RemoveAccount(account.ID)
						if err != nil {
							errorsList = append(errorsList, errors.RemoveAccountError+err.Error())
						}
					} else {
						newAccounts = append(newAccounts, account)
					}
				}
				settings.Accounts = newAccounts
			} else {
				errorsList = append(errorsList, errors.ParseRemoveAccountsRequestError+err.Error())
			}
		} else {
			errorsList = append(errorsList, errors.GetSettingsError+err.Error())
		}
	} else {
		errorsList = append(errorsList, errors.IOUtilReadAllError+err.Error())
	}
	result := &responses.SettingsResponse{Success: true, Data: settings, Errors: make([]string, 0)}
	if len(errorsList) > 0 {
		response.WriteHeader(http.StatusBadRequest)
		result = &responses.SettingsResponse{Success: false, Data: entities.Settings{}, Errors: errorsList}
	}
	json.NewEncoder(response).Encode(result)
}
