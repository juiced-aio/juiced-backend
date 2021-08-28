package endpoints

import (
	"time"

	"backend.juicedbot.io/juiced.api/errors"
	"backend.juicedbot.io/juiced.api/responses"
	"backend.juicedbot.io/juiced.infrastructure/common"
	"backend.juicedbot.io/juiced.infrastructure/entities"
	"backend.juicedbot.io/juiced.infrastructure/stores"

	"encoding/json"
	"io/ioutil"
	"net/http"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
)

// GetProxyGroupEndpoint handles the GET request at /api/proxy/group/{groupID}
func GetProxyGroupEndpoint(response http.ResponseWriter, request *http.Request) {
	response.Header().Set("content-type", "application/json")
	response.Header().Set("Access-Control-Allow-Origin", "http://localhost:3000")
	var proxyGroup entities.ProxyGroup
	errorsList := make([]string, 0)

	params := mux.Vars(request)
	groupID, ok := params["GroupID"]
	if ok {
		proxyGroupPtr, err := stores.GetProxyGroup(groupID)
		if err == nil {
			proxyGroup = *proxyGroupPtr
		} else {
			errorsList = append(errorsList, errors.GetProxyGroupError+err.Error())
		}
	} else {
		errorsList = append(errorsList, errors.MissingParameterError)
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
	proxyGroups := stores.GetAllProxyGroups()

	data := []entities.ProxyGroup{}
	for i := 0; i < len(proxyGroups); i++ {
		data = append(data, *proxyGroups[i])
	}
	result := &responses.ProxyGroupResponse{Success: true, Data: data, Errors: make([]string, 0)}
	json.NewEncoder(response).Encode(result)
}

// CreateProxyGroupEndpoint handles the POST request at /api/proxy/group
func CreateProxyGroupEndpoint(response http.ResponseWriter, request *http.Request) {
	response.Header().Set("content-type", "application/json")
	response.Header().Set("Access-Control-Allow-Origin", "http://localhost:3000")
	proxyGroup := entities.ProxyGroup{GroupID: uuid.New().String()}
	errorsList := make([]string, 0)

	body, err := ioutil.ReadAll(request.Body)
	if err == nil {
		err = json.Unmarshal(body, &proxyGroup)
		if err == nil {
			for _, proxy := range proxyGroup.Proxies {
				proxy.ProxyGroupID = proxyGroup.GroupID
			}
			proxyGroup.CreationDate = time.Now().Unix()

			proxyGroupPtr, err := stores.CreateProxyGroup(proxyGroup)
			if err == nil {
				proxyGroup = *proxyGroupPtr
			} else {
				errorsList = append(errorsList, errors.CreateProxyGroupError+err.Error())
			}
		} else {
			errorsList = append(errorsList, errors.ParseProxyGroupError+err.Error())
		}
	} else {
		errorsList = append(errorsList, errors.IOUtilReadAllError+err.Error())
	}

	result := &responses.ProxyGroupResponse{Success: true, Data: []entities.ProxyGroup{proxyGroup}, Errors: make([]string, 0)}
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
	var proxyGroup entities.ProxyGroup
	errorsList := make([]string, 0)

	params := mux.Vars(request)
	groupID, ok := params["GroupID"]
	if ok {
		taskGroups, err := stores.GetTaskGroupsByProxyGroupID(groupID)
		if err == nil {
			next := true
			monitorStore := stores.GetMonitorStore()
			for _, taskGroup := range taskGroups {
				updated := monitorStore.UpdateMonitorProxy(&taskGroup, &entities.Proxy{})
				if !updated {
					next = false
				}
			}
			if next {
				tasks, err := stores.GetTasksByProxyGroupID(groupID)
				if err == nil {
					next := true
					taskStore := stores.GetTaskStore()
					for _, task := range tasks {
						updated := taskStore.UpdateTaskProxy(&task, &entities.Proxy{})
						if !updated {
							next = false
							break
						}
					}
					if next {
						proxyGroup, err = stores.RemoveProxyGroup(groupID)
						if err == nil {
							stores.GetProxyStore().RemoveProxyGroup(groupID)
						} else {
							errorsList = append(errorsList, errors.RemoveProxyGroupError+err.Error())
						}
					} else {
						errorsList = append(errorsList, errors.RemoveProxyGroupError+"error while updating a tasks proxy")
					}
				} else {
					errorsList = append(errorsList, errors.GetTaskError+err.Error())
				}
			} else {
				errorsList = append(errorsList, errors.RemoveProxyGroupError+"error while updating a taskgroups proxy")
			}
		} else {
			errorsList = append(errorsList, errors.RemoveProxyGroupError+err.Error())
		}
	} else {
		errorsList = append(errorsList, errors.MissingParameterError)
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
	var newProxyGroup entities.ProxyGroup
	errorsList := make([]string, 0)

	params := mux.Vars(request)
	groupID, ok := params["GroupID"]
	if ok {
		newProxyGroup = entities.ProxyGroup{GroupID: groupID}
		body, err := ioutil.ReadAll(request.Body)
		if err == nil {
			err = json.Unmarshal(body, &newProxyGroup)
			if err == nil {
				newProxyGroupPtr, err := stores.UpdateProxyGroup(groupID, newProxyGroup)
				if err == nil {
					newProxyGroup = *newProxyGroupPtr
				} else {
					errorsList = append(errorsList, errors.UpdateProxyGroupError+err.Error())
				}
			} else {
				errorsList = append(errorsList, errors.ParseProxyGroupError+err.Error())
			}
		} else {
			errorsList = append(errorsList, errors.IOUtilReadAllError+err.Error())
		}
	} else {
		errorsList = append(errorsList, errors.MissingParameterError)
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
	var newProxyGroup entities.ProxyGroup
	errorsList := make([]string, 0)

	params := mux.Vars(request)
	groupID, ok := params["GroupID"]
	if ok {
		oldProxyGroupPtr, err := stores.GetProxyGroup(groupID)
		if err == nil {
			newProxyGroup = *oldProxyGroupPtr
			newProxyGroup.GroupID = uuid.New().String()
			newProxyGroup.Name = newProxyGroup.Name + " (Copy " + common.RandID(3) + ")"
			newProxyGroup.CreationDate = time.Now().Unix()
			for _, proxy := range newProxyGroup.Proxies {
				proxy.ID = uuid.New().String()
				proxy.ProxyGroupID = newProxyGroup.GroupID
			}
			newProxyGroupPtr, err := stores.CreateProxyGroup(newProxyGroup)
			if err == nil {
				newProxyGroup = *newProxyGroupPtr
			} else {
				errorsList = append(errorsList, errors.CreateProfileGroupError+err.Error())
			}
		} else {
			errorsList = append(errorsList, errors.GetProxyGroupError+err.Error())
		}
	} else {
		errorsList = append(errorsList, errors.MissingParameterError)
	}
	result := &responses.ProxyGroupResponse{Success: true, Data: []entities.ProxyGroup{newProxyGroup}, Errors: make([]string, 0)}
	if len(errorsList) > 0 {
		response.WriteHeader(http.StatusBadRequest)
		result = &responses.ProxyGroupResponse{Success: false, Data: make([]entities.ProxyGroup, 0), Errors: errorsList}
	}
	json.NewEncoder(response).Encode(result)
}
