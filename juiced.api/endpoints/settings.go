package endpoints

import (
	"backend.juicedbot.io/juiced.api/responses"
	"backend.juicedbot.io/juiced.infrastructure/commands"
	"backend.juicedbot.io/juiced.infrastructure/common/entities"
	"backend.juicedbot.io/juiced.infrastructure/common/errors"
	"backend.juicedbot.io/juiced.infrastructure/queries"

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
			newSettings, err = commands.UpdateSettings(newSettings)
			if err != nil {
				errorsList = append(errorsList, errors.UpdateSettingsError+err.Error())
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
