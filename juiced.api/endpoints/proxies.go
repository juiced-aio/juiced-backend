package endpoints

import (
	"juiced/juiced.api/responses"
	"juiced/juiced.infrastructure/commands"
	"juiced/juiced.infrastructure/common"
	"juiced/juiced.infrastructure/common/entities"
	"juiced/juiced.infrastructure/common/errors"
	"juiced/juiced.infrastructure/queries"

	"encoding/json"
	"io/ioutil"
	"net/http"

	"github.com/gorilla/mux"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// GetProxyGroupEndpoint handles the GET request at /api/proxy/group/{groupID}
func GetProxyGroupEndpoint(response http.ResponseWriter, request *http.Request) {
	response.Header().Set("content-type", "application/json")
	response.Header().Set("Access-Control-Allow-Origin", "http://localhost:3000")
	proxyGroup := entities.ProxyGroup{}
	errorsList := make([]string, 0)

	params := mux.Vars(request)
	groupID, err := primitive.ObjectIDFromHex(params["GroupID"])
	if err == nil {
		proxyGroup, err = queries.GetProxyGroup(groupID)
		if err != nil {
			errorsList = append(errorsList, errors.GetProxyGroupError+err.Error())
		}
	} else {
		errorsList = append(errorsList, errors.ParseObjectIDError+err.Error())
	}
	result := &responses.ProxyGroupResponse{Success: true, Data: []entities.ProxyGroup{proxyGroup}, Errors: make([]string, 0)}
	if len(errorsList) > 0 {
		response.WriteHeader(http.StatusBadRequest)
		result = &responses.ProxyGroupResponse{Success: false, Data: make([]entities.ProxyGroup, 0), Errors: errorsList}
	}
	json.NewEncoder(response).Encode(result)
}

// GetAllProxyGroupsEndpoint handles the GET request at /api/proxy/group
func GetAllProxyGroupsEndpoint(response http.ResponseWriter, request *http.Request) {
	response.Header().Set("content-type", "application/json")
	response.Header().Set("Access-Control-Allow-Origin", "http://localhost:3000")
	errorsList := make([]string, 0)
	proxyGroups, err := queries.GetAllProxyGroups()
	if err != nil {
		errorsList = append(errorsList, errors.GetAllProxyGroupsError+err.Error())
	}
	result := &responses.ProxyGroupResponse{Success: true, Data: proxyGroups, Errors: make([]string, 0)}
	if len(errorsList) > 0 {
		response.WriteHeader(http.StatusBadRequest)
		result = &responses.ProxyGroupResponse{Success: false, Data: make([]entities.ProxyGroup, 0), Errors: errorsList}
	}
	json.NewEncoder(response).Encode(result)
}

// CreateProxyGroupEndpoint handles the POST request at /api/proxy/group
func CreateProxyGroupEndpoint(response http.ResponseWriter, request *http.Request) {
	response.Header().Set("content-type", "application/json")
	response.Header().Set("Access-Control-Allow-Origin", "http://localhost:3000")
	proxyGroup := &entities.ProxyGroup{GroupID: primitive.NewObjectID()}
	errorsList := make([]string, 0)

	body, err := ioutil.ReadAll(request.Body)
	if err == nil {
		err = entities.ParseProxyGroup(proxyGroup, body)
		if err == nil {
			err = commands.CreateProxyGroup(*proxyGroup)
			if err != nil {
				errorsList = append(errorsList, errors.CreateProxyGroupError+err.Error())
			}
		} else {
			errorsList = append(errorsList, errors.ParseProxyGroupError+err.Error())
		}
	} else {
		errorsList = append(errorsList, errors.IOUtilReadAllError+err.Error())
	}
	result := &responses.ProxyGroupResponse{Success: true, Data: []entities.ProxyGroup{*proxyGroup}, Errors: make([]string, 0)}
	if len(errorsList) > 0 {
		response.WriteHeader(http.StatusBadRequest)
		result = &responses.ProxyGroupResponse{Success: false, Data: make([]entities.ProxyGroup, 0), Errors: errorsList}
	}
	json.NewEncoder(response).Encode(result)
}

// RemoveProxyGroupEndpoint handles the DELETE request at /api/proxy/group/{GroupID}
func RemoveProxyGroupEndpoint(response http.ResponseWriter, request *http.Request) {
	response.Header().Set("content-type", "application/json")
	response.Header().Set("Access-Control-Allow-Origin", "http://localhost:3000")
	proxyGroup := entities.ProxyGroup{}
	errorsList := make([]string, 0)

	params := mux.Vars(request)
	groupID, err := primitive.ObjectIDFromHex(params["GroupID"])
	if err == nil {
		proxyGroup, err = commands.RemoveProxyGroup(groupID)
		if err != nil {
			errorsList = append(errorsList, errors.RemoveProxyGroupError+err.Error())
		}
	} else {
		errorsList = append(errorsList, errors.ParseObjectIDError+err.Error())
	}
	result := &responses.ProxyGroupResponse{Success: true, Data: []entities.ProxyGroup{proxyGroup}, Errors: make([]string, 0)}
	if len(errorsList) > 0 {
		response.WriteHeader(http.StatusBadRequest)
		result = &responses.ProxyGroupResponse{Success: false, Data: make([]entities.ProxyGroup, 0), Errors: errorsList}
	}
	json.NewEncoder(response).Encode(result)
}

// UpdateProxyGroupEndpoint handles the PUT request at /api/proxy/group/{GroupID}
func UpdateProxyGroupEndpoint(response http.ResponseWriter, request *http.Request) {
	response.Header().Set("content-type", "application/json")
	response.Header().Set("Access-Control-Allow-Origin", "http://localhost:3000")
	errorsList := make([]string, 0)

	params := mux.Vars(request)
	groupID, err := primitive.ObjectIDFromHex(params["GroupID"])
	newProxyGroup := entities.ProxyGroup{GroupID: groupID}
	if err == nil {
		body, err := ioutil.ReadAll(request.Body)
		if err == nil {
			err = entities.ParseProxyGroup(&newProxyGroup, body)
			if err == nil {
				newProxyGroup, err = commands.UpdateProxyGroup(groupID, newProxyGroup)
				if err != nil {
					errorsList = append(errorsList, errors.UpdateProxyGroupError+err.Error())
				}
			} else {
				errorsList = append(errorsList, errors.ParseProxyGroupError+err.Error())
			}
		} else {
			errorsList = append(errorsList, errors.IOUtilReadAllError+err.Error())
		}
	} else {
		errorsList = append(errorsList, errors.ParseObjectIDError+err.Error())
	}
	result := &responses.ProxyGroupResponse{Success: true, Data: []entities.ProxyGroup{newProxyGroup}, Errors: make([]string, 0)}
	if len(errorsList) > 0 {
		response.WriteHeader(http.StatusBadRequest)
		result = &responses.ProxyGroupResponse{Success: false, Data: make([]entities.ProxyGroup, 0), Errors: errorsList}
	}
	json.NewEncoder(response).Encode(result)
}

// CloneProxyGroupEndpoint handles the POST request at /api/proxy/group/{GroupID}/clone
func CloneProxyGroupEndpoint(response http.ResponseWriter, request *http.Request) {
	response.Header().Set("content-type", "application/json")
	response.Header().Set("Access-Control-Allow-Origin", "http://localhost:3000")
	proxyGroup := entities.ProxyGroup{}
	errorsList := make([]string, 0)

	params := mux.Vars(request)
	groupID, err := primitive.ObjectIDFromHex(params["GroupID"])
	if err == nil {
		proxyGroup, err = queries.GetProxyGroup(groupID)
		if err == nil {
			proxyGroup.SetGroupID(primitive.NewObjectID())
			proxyGroup.SetName(proxyGroup.Name + " (Copy " + common.RandID(4) + ")")
			for i := 0; i < len(proxyGroup.Proxies); i++ {
				proxy := &proxyGroup.Proxies[i]
				proxy.SetID(primitive.NewObjectID())
			}
			err = commands.CreateProxyGroup(proxyGroup)
			if err != nil {
				errorsList = append(errorsList, errors.CreateProxyGroupError+err.Error())
			}
		} else {
			errorsList = append(errorsList, errors.GetProxyGroupError+err.Error())
		}
	} else {
		errorsList = append(errorsList, errors.ParseObjectIDError+err.Error())
	}
	result := &responses.ProxyGroupResponse{Success: true, Data: []entities.ProxyGroup{proxyGroup}, Errors: make([]string, 0)}
	if len(errorsList) > 0 {
		response.WriteHeader(http.StatusBadRequest)
		result = &responses.ProxyGroupResponse{Success: false, Data: make([]entities.ProxyGroup, 0), Errors: errorsList}
	}
	json.NewEncoder(response).Encode(result)
}
