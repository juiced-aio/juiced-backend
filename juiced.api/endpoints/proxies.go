package endpoints

import (
	"time"

	"backend.juicedbot.io/juiced.api/responses"
	"backend.juicedbot.io/juiced.infrastructure/commands"
	"backend.juicedbot.io/juiced.infrastructure/common"
	"backend.juicedbot.io/juiced.infrastructure/common/entities"
	"backend.juicedbot.io/juiced.infrastructure/common/errors"
	"backend.juicedbot.io/juiced.infrastructure/common/stores"
	"backend.juicedbot.io/juiced.infrastructure/queries"

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
	var err error
	errorsList := make([]string, 0)

	params := mux.Vars(request)
	groupID, ok := params["GroupID"]
	if ok {
		proxyGroup, err = queries.GetProxyGroup(groupID)
		if err != nil {
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
	proxyGroup := &entities.ProxyGroup{GroupID: uuid.New().String()}
	errorsList := make([]string, 0)

	body, err := ioutil.ReadAll(request.Body)
	if err == nil {
		err = entities.ParseProxyGroup(proxyGroup, body)
		if err == nil {
			// This should probably be done on the frontend
			proxiesWithGroupID := []entities.Proxy{}
			for _, proxy := range proxyGroup.Proxies {
				proxy.ProxyGroupID = proxyGroup.GroupID
				proxiesWithGroupID = append(proxiesWithGroupID, proxy)
			}
			proxyGroup.CreationDate = time.Now().Unix()
			proxyGroup.Proxies = proxiesWithGroupID
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

// RemoveProxyGroupsEndpoint handles the DELETE request at /api/proxy/group/remove
func RemoveProxyGroupsEndpoint(response http.ResponseWriter, request *http.Request) {
	response.Header().Set("content-type", "application/json")
	response.Header().Set("Access-Control-Allow-Origin", "http://localhost:3000")
	var proxyGroups []entities.ProxyGroup
	errorsList := make([]string, 0)

	type RemoveProxyGroupsRequest struct {
		GroupIDs []string `json:"groupIDs"`
	}

	body, err := ioutil.ReadAll(request.Body)
	if err == nil {
		removeProxyGroupsRequest := RemoveProxyGroupsRequest{}
		err = json.Unmarshal(body, &removeProxyGroupsRequest)
		if err == nil {
			for _, groupID := range removeProxyGroupsRequest.GroupIDs {
				taskGroups, err := queries.GetTaskGroupsByProxyGroupID(groupID)
				if err == nil {
					next := true
					monitorStore := stores.GetMonitorStore()
					for _, taskGroup := range taskGroups {
						updated := monitorStore.UpdateMonitorProxy(&taskGroup, entities.Proxy{})
						if !updated {
							next = false
						}
					}
					if next {
						tasks, err := queries.GetTasksByProxyGroupID(groupID)
						if err == nil {
							next := true
							taskStore := stores.GetTaskStore()
							for _, task := range tasks {
								updated := taskStore.UpdateTaskProxy(&task, entities.Proxy{})
								if !updated {
									next = false
									break
								}
							}
							if next {
								proxyGroup, err := commands.RemoveProxyGroup(groupID)
								if err == nil {
									proxyGroups = append(proxyGroups, proxyGroup)
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
			}
		} else {
			errorsList = append(errorsList, errors.ParseRemoveProxyGroupsRequestError+err.Error())
		}
	} else {
		errorsList = append(errorsList, errors.IOUtilReadAllError+err.Error())
	}

	data := []entities.ProxyGroup{}
	data = append(data, proxyGroups...)

	result := &responses.ProxyGroupResponse{Success: true, Data: data, Errors: make([]string, 0)}
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
		errorsList = append(errorsList, errors.MissingParameterError)
	}
	result := &responses.ProxyGroupResponse{Success: true, Data: []entities.ProxyGroup{newProxyGroup}, Errors: make([]string, 0)}
	if len(errorsList) > 0 {
		response.WriteHeader(http.StatusBadRequest)
		result = &responses.ProxyGroupResponse{Success: false, Data: make([]entities.ProxyGroup, 0), Errors: errorsList}
	}
	json.NewEncoder(response).Encode(result)
}

// CloneProxyGroupsEndpoint handles the POST request at /api/proxy/group/clone
func CloneProxyGroupsEndpoint(response http.ResponseWriter, request *http.Request) {
	response.Header().Set("content-type", "application/json")
	response.Header().Set("Access-Control-Allow-Origin", "http://localhost:3000")
	var proxyGroups []entities.ProxyGroup
	var err error
	errorsList := make([]string, 0)

	type CloneProxyGroupsRequest struct {
		GroupIDs []string `json:"groupIDs"`
	}
	body, err := ioutil.ReadAll(request.Body)
	if err == nil {
		cloneProxyGroupsRequest := CloneProxyGroupsRequest{}
		err = json.Unmarshal(body, &cloneProxyGroupsRequest)
		if err == nil {
			for _, groupID := range cloneProxyGroupsRequest.GroupIDs {
				proxyGroup, err := queries.GetProxyGroup(groupID)
				if err == nil {
					newGroupID := uuid.New().String()
					proxyGroup.SetGroupID(newGroupID)
					proxyGroup.SetName(proxyGroup.Name + " (Copy " + common.RandID(4) + ")")
					proxyGroup.CreationDate = time.Now().Unix()
					for i := 0; i < len(proxyGroup.Proxies); i++ {
						proxy := &proxyGroup.Proxies[i]
						proxy.SetID(uuid.New().String())
						proxy.ProxyGroupID = newGroupID
					}
					err = commands.CreateProxyGroup(proxyGroup)
					if err == nil {
						proxyGroups = append(proxyGroups, proxyGroup)
					} else {
						errorsList = append(errorsList, errors.CreateProxyGroupError+err.Error())
					}
				} else {
					errorsList = append(errorsList, errors.GetProxyGroupError+err.Error())
				}
			}
		} else {
			errorsList = append(errorsList, errors.ParseRemoveProxyGroupsRequestError+err.Error())
		}
	} else {
		errorsList = append(errorsList, errors.IOUtilReadAllError+err.Error())
	}

	data := []entities.ProxyGroup{}
	data = append(data, proxyGroups...)

	result := &responses.ProxyGroupResponse{Success: true, Data: data, Errors: make([]string, 0)}
	if len(errorsList) > 0 {
		response.WriteHeader(http.StatusBadRequest)
		result = &responses.ProxyGroupResponse{Success: false, Data: make([]entities.ProxyGroup, 0), Errors: errorsList}
	}
	json.NewEncoder(response).Encode(result)
}
