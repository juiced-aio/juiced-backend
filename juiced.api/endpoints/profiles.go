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

// GetProfileGroupEndpoint handles the GET request at /api/profile/group/{groupID}
func GetProfileGroupEndpoint(response http.ResponseWriter, request *http.Request) {
	response.Header().Set("content-type", "application/json")
	response.Header().Set("Access-Control-Allow-Origin", "http://localhost:3000")
	var profileGroup entities.ProfileGroup
	var err error
	errorsList := make([]string, 0)

	params := mux.Vars(request)
	groupID, ok := params["GroupID"]
	if ok {
		profileGroup, err = queries.GetProfileGroup(groupID)
		if err != nil {
			errorsList = append(errorsList, errors.GetProfileGroupError+err.Error())
		}
	} else {
		errorsList = append(errorsList, errors.MissingParameterError)
	}
	newProfileGroupWithProfiles, err := queries.ConvertProfileIDsToProfiles(&profileGroup)
	if err != nil {
		errorsList = append(errorsList, errors.GetProfileError+err.Error())
	}
	data := []entities.ProfileGroupWithProfiles{newProfileGroupWithProfiles}
	result := &responses.ProfileGroupResponse{Success: true, Data: data, Errors: make([]string, 0)}
	if len(errorsList) > 0 {
		response.WriteHeader(http.StatusBadRequest)
		result = &responses.ProfileGroupResponse{Success: false, Data: data, Errors: errorsList}
	}
	json.NewEncoder(response).Encode(result)
}

// GetAllProfileGroupsEndpoint handles the GET request at /api/profile/group
func GetAllProfileGroupsEndpoint(response http.ResponseWriter, request *http.Request) {
	response.Header().Set("content-type", "application/json")
	response.Header().Set("Access-Control-Allow-Origin", "http://localhost:3000")
	errorsList := make([]string, 0)
	profileGroups, err := queries.GetAllProfileGroups()
	if err != nil {
		errorsList = append(errorsList, errors.GetAllProfileGroupsError+err.Error())
	}
	data := []entities.ProfileGroupWithProfiles{}
	for i := 0; i < len(profileGroups); i++ {
		newProfileGroupWithProfiles, err := queries.ConvertProfileIDsToProfiles(&profileGroups[i])
		if err != nil {
			errorsList = append(errorsList, errors.GetProfileError+err.Error())
		}
		data = append(data, newProfileGroupWithProfiles)
	}
	result := &responses.ProfileGroupResponse{Success: true, Data: data, Errors: make([]string, 0)}
	if len(errorsList) > 0 {
		response.WriteHeader(http.StatusBadRequest)
		result = &responses.ProfileGroupResponse{Success: false, Data: data, Errors: errorsList}
	}
	json.NewEncoder(response).Encode(result)
}

// CreateProfileGroupEndpoint handles the POST request at /api/profile/group
func CreateProfileGroupEndpoint(response http.ResponseWriter, request *http.Request) {
	response.Header().Set("content-type", "application/json")
	response.Header().Set("Access-Control-Allow-Origin", "http://localhost:3000")
	profileGroup := &entities.ProfileGroup{GroupID: uuid.New().String(), ProfileIDs: []string{}}
	errorsList := make([]string, 0)

	body, err := ioutil.ReadAll(request.Body)
	if err == nil {
		err = entities.ParseProfileGroup(profileGroup, body)
		if err == nil {
			profileGroup.CreationDate = time.Now().Unix()
			err = commands.CreateProfileGroup(*profileGroup)
			if err != nil {
				errorsList = append(errorsList, errors.CreateProfileGroupError+err.Error())
			}
		} else {
			errorsList = append(errorsList, errors.ParseProfileGroupError+err.Error())
		}
	} else {
		errorsList = append(errorsList, errors.IOUtilReadAllError+err.Error())
	}
	newProfileGroupWithProfiles, err := queries.ConvertProfileIDsToProfiles(profileGroup)
	if err != nil {
		errorsList = append(errorsList, errors.GetProfileError+err.Error())
	}
	data := []entities.ProfileGroupWithProfiles{newProfileGroupWithProfiles}
	result := &responses.ProfileGroupResponse{Success: true, Data: data, Errors: make([]string, 0)}
	if len(errorsList) > 0 {
		response.WriteHeader(http.StatusBadRequest)
		result = &responses.ProfileGroupResponse{Success: false, Data: data, Errors: errorsList}
	}
	json.NewEncoder(response).Encode(result)
}

// RemoveProfileGroupsEndpoint handles the POST request at /api/profile/group/remove
func RemoveProfileGroupsEndpoint(response http.ResponseWriter, request *http.Request) {
	response.Header().Set("content-type", "application/json")
	response.Header().Set("Access-Control-Allow-Origin", "http://localhost:3000")
	var profileGroups []entities.ProfileGroup
	var err error
	errorsList := make([]string, 0)

	type RemoveProfileGroupsRequest struct {
		GroupIDs []string `json:"groupIDs"`
	}

	body, err := ioutil.ReadAll(request.Body)
	if err == nil {
		removeProfileGroupsRequest := RemoveProfileGroupsRequest{}
		err = json.Unmarshal(body, &removeProfileGroupsRequest)
		if err == nil {
			for _, groupID := range removeProfileGroupsRequest.GroupIDs {
				profileGroup, err := commands.RemoveProfileGroup(groupID)
				if err == nil {
					profileGroups = append(profileGroups, profileGroup)
				} else {
					errorsList = append(errorsList, errors.RemoveProfileGroupError+err.Error())
				}
			}
		} else {
			errorsList = append(errorsList, errors.ParseRemoveProfileGroupsRequestError+err.Error())
		}
	} else {
		errorsList = append(errorsList, errors.IOUtilReadAllError+err.Error())
	}

	data := []entities.ProfileGroupWithProfiles{}

	for _, profileGroup := range profileGroups {
		newProfileGroupWithProfiles, err := queries.ConvertProfileIDsToProfiles(&profileGroup)
		if err != nil {
			errorsList = append(errorsList, errors.GetProfileError+err.Error())
		}
		data = append(data, newProfileGroupWithProfiles)
	}

	result := &responses.ProfileGroupResponse{Success: true, Data: data, Errors: make([]string, 0)}
	if len(errorsList) > 0 {
		response.WriteHeader(http.StatusBadRequest)
		result = &responses.ProfileGroupResponse{Success: false, Data: data, Errors: errorsList}
	}
	json.NewEncoder(response).Encode(result)
}

// UpdateProfileGroupEndpoint handles the PUT request at /api/profile/group/{GroupID}
func UpdateProfileGroupEndpoint(response http.ResponseWriter, request *http.Request) {
	response.Header().Set("content-type", "application/json")
	response.Header().Set("Access-Control-Allow-Origin", "http://localhost:3000")
	var profileGroup entities.ProfileGroup
	var err error
	errorsList := make([]string, 0)

	params := mux.Vars(request)
	groupID, ok := params["GroupID"]
	if ok {
		profileGroup, err = queries.GetProfileGroup(groupID)
		if err == nil {
			body, err := ioutil.ReadAll(request.Body)
			if err == nil {
				type Name struct {
					Name string `json:"name"`
				}
				name := Name{}
				err := json.Unmarshal(body, &name)
				if err == nil {
					profileGroup.SetName(name.Name)
					profileGroup, err = commands.UpdateProfileGroup(groupID, profileGroup)
					if err != nil {
						errorsList = append(errorsList, errors.UpdateProfileGroupError+err.Error())
					}
				} else {
					errorsList = append(errorsList, errors.ParseProfileGroupError+err.Error())
				}
			} else {
				errorsList = append(errorsList, errors.IOUtilReadAllError+err.Error())
			}
		} else {
			errorsList = append(errorsList, errors.GetProfileGroupError+err.Error())
		}
	} else {
		errorsList = append(errorsList, errors.MissingParameterError)
	}
	profileGroupWithProfiles, err := queries.ConvertProfileIDsToProfiles(&profileGroup)
	if err != nil {
		errorsList = append(errorsList, errors.GetProfileError+err.Error())
	}
	data := []entities.ProfileGroupWithProfiles{profileGroupWithProfiles}
	result := &responses.ProfileGroupResponse{Success: true, Data: data, Errors: make([]string, 0)}
	if len(errorsList) > 0 {
		response.WriteHeader(http.StatusBadRequest)
		result = &responses.ProfileGroupResponse{Success: false, Data: data, Errors: errorsList}
	}
	json.NewEncoder(response).Encode(result)
}

// CloneProfileGroupsEndpoint handles the POST request at /api/profile/group/clone
func CloneProfileGroupsEndpoint(response http.ResponseWriter, request *http.Request) {
	response.Header().Set("content-type", "application/json")
	response.Header().Set("Access-Control-Allow-Origin", "http://localhost:3000")
	var newProfileGroups []entities.ProfileGroup
	var err error
	errorsList := make([]string, 0)

	type CloneProfileGroupsRequest struct {
		GroupIDs []string `json:"groupIDs"`
	}

	body, err := ioutil.ReadAll(request.Body)
	if err == nil {
		cloneProfileGroupsRequest := CloneProfileGroupsRequest{}
		err = json.Unmarshal(body, &cloneProfileGroupsRequest)
		if err == nil {
			for _, groupID := range cloneProfileGroupsRequest.GroupIDs {
				newProfileGroup, err := queries.GetProfileGroup(groupID)
				if err == nil {
					newProfileGroup.SetGroupID(uuid.New().String())
					newProfileGroup.SetName(newProfileGroup.Name + " (Copy " + common.RandID(4) + ")")
					newProfileGroup.CreationDate = time.Now().Unix()
					err = commands.CreateProfileGroup(newProfileGroup)
					if err == nil {
						newProfileGroups = append(newProfileGroups, newProfileGroup)
					} else {
						errorsList = append(errorsList, errors.CreateProfileGroupError+err.Error())
					}
				} else {
					errorsList = append(errorsList, errors.GetProfileGroupError+err.Error())
				}
			}
		} else {
			errorsList = append(errorsList, errors.ParseCloneProfileGroupsRequestError+err.Error())
		}
	} else {
		errorsList = append(errorsList, errors.IOUtilReadAllError+err.Error())
	}

	data := []entities.ProfileGroupWithProfiles{}
	for _, newProfileGroup := range newProfileGroups {
		newProfileGroupWithProfiles, err := queries.ConvertProfileIDsToProfiles(&newProfileGroup)
		if err != nil {
			errorsList = append(errorsList, errors.GetProfileError+err.Error())
		}
		data = append(data, newProfileGroupWithProfiles)
	}

	result := &responses.ProfileGroupResponse{Success: true, Data: data, Errors: make([]string, 0)}
	if len(errorsList) > 0 {
		response.WriteHeader(http.StatusBadRequest)
		result = &responses.ProfileGroupResponse{Success: false, Data: data, Errors: errorsList}
	}
	json.NewEncoder(response).Encode(result)
}

// AddProfilesToGroupsEndpoint handles the POST request at /api/profile/group/add
func AddProfilesToGroupsEndpoint(response http.ResponseWriter, request *http.Request) {
	response.Header().Set("content-type", "application/json")
	response.Header().Set("Access-Control-Allow-Origin", "http://localhost:3000")
	var profileGroups []entities.ProfileGroup
	var err error
	errorsList := make([]string, 0)

	type AddProfilesToGroupsRequest struct {
		GroupIDs []string `json:"groupIDs"`
	}

	body, err := ioutil.ReadAll(request.Body)
	if err == nil {
		addProfilesToGroupsRequest := AddProfilesToGroupsRequest{}
		err = json.Unmarshal(body, &addProfilesToGroupsRequest)
		if err == nil {
			for _, groupID := range addProfilesToGroupsRequest.GroupIDs {
				profileGroup, err := queries.GetProfileGroup(groupID)
				if err == nil {
					body, err := ioutil.ReadAll(request.Body)
					if err == nil {
						type ProfileIDs struct {
							ProfileIDs []string `json:"profileIDs"`
						}
						profileIDs := ProfileIDs{}
						err := json.Unmarshal(body, &profileIDs)
						if err == nil {
							profileGroup.AddProfileIDsToGroup(profileIDs.ProfileIDs)
							profileGroup, err = commands.UpdateProfileGroup(groupID, profileGroup)
							if err == nil {
								profileGroups = append(profileGroups, profileGroup)
							} else {
								errorsList = append(errorsList, errors.UpdateProfileGroupError+err.Error())
							}
						} else {
							errorsList = append(errorsList, errors.ParseProfileGroupError+err.Error())
						}
					} else {
						errorsList = append(errorsList, errors.IOUtilReadAllError+err.Error())
					}
				} else {
					errorsList = append(errorsList, errors.GetProfileGroupError+err.Error())
				}
			}
		} else {
			errorsList = append(errorsList, errors.ParseAddProfilesToGroupsRequestError+err.Error())
		}
	} else {
		errorsList = append(errorsList, errors.IOUtilReadAllError+err.Error())
	}

	data := []entities.ProfileGroupWithProfiles{}
	for _, profileGroup := range profileGroups {
		profileGroupWithProfiles, err := queries.ConvertProfileIDsToProfiles(&profileGroup)
		if err != nil {
			errorsList = append(errorsList, errors.GetProfileError+err.Error())
		}
		data = append(data, profileGroupWithProfiles)
	}

	result := &responses.ProfileGroupResponse{Success: true, Data: data, Errors: make([]string, 0)}
	if len(errorsList) > 0 {
		response.WriteHeader(http.StatusBadRequest)
		result = &responses.ProfileGroupResponse{Success: false, Data: data, Errors: errorsList}
	}
	json.NewEncoder(response).Encode(result)
}

// RemoveProfilesFromGroupsEndpoint handles the POST request at /api/profile/group/remove
func RemoveProfilesFromGroupsEndpoint(response http.ResponseWriter, request *http.Request) {
	response.Header().Set("content-type", "application/json")
	response.Header().Set("Access-Control-Allow-Origin", "http://localhost:3000")
	var profileGroups []entities.ProfileGroup
	var err error
	errorsList := make([]string, 0)

	type RemoveProfilesFromGroupsRequest struct {
		GroupIDs []string `json:"groupIDs"`
	}

	body, err := ioutil.ReadAll(request.Body)
	if err == nil {
		removeProfilesFromGroupsRequest := RemoveProfilesFromGroupsRequest{}
		err = json.Unmarshal(body, &removeProfilesFromGroupsRequest)
		if err == nil {
			for _, groupID := range removeProfilesFromGroupsRequest.GroupIDs {
				profileGroup, err := queries.GetProfileGroup(groupID)
				if err == nil {
					body, err := ioutil.ReadAll(request.Body)
					if err == nil {
						type ProfileIDs struct {
							ProfileIDs []string `json:"profileIDs"`
						}
						profileIDs := ProfileIDs{}
						err := json.Unmarshal(body, &profileIDs)
						if err == nil {
							profileGroup.RemoveProfileIDsFromGroup(profileIDs.ProfileIDs)
							profileGroup, err = commands.UpdateProfileGroup(groupID, profileGroup)
							if err == nil {
								profileGroups = append(profileGroups, profileGroup)
							} else {
								errorsList = append(errorsList, errors.UpdateProfileGroupError+err.Error())
							}
						} else {
							errorsList = append(errorsList, errors.ParseProfileGroupError+err.Error())
						}
					} else {
						errorsList = append(errorsList, errors.IOUtilReadAllError+err.Error())
					}
				} else {
					errorsList = append(errorsList, errors.GetProfileGroupError+err.Error())
				}
			}
		} else {
			errorsList = append(errorsList, errors.ParseRemoveProfilesFromGroupsRequestError+err.Error())
		}
	} else {
		errorsList = append(errorsList, errors.IOUtilReadAllError+err.Error())
	}

	data := []entities.ProfileGroupWithProfiles{}
	for _, profileGroup := range profileGroups {
		profileGroupWithProfiles, err := queries.ConvertProfileIDsToProfiles(&profileGroup)
		if err != nil {
			errorsList = append(errorsList, errors.GetProfileError+err.Error())
		}
		data = append(data, profileGroupWithProfiles)
	}

	result := &responses.ProfileGroupResponse{Success: true, Data: data, Errors: make([]string, 0)}
	if len(errorsList) > 0 {
		response.WriteHeader(http.StatusBadRequest)
		result = &responses.ProfileGroupResponse{Success: false, Data: data, Errors: errorsList}
	}
	json.NewEncoder(response).Encode(result)
}

// GetAllProfilesEndpoint handles the GET request at /api/profile/all
func GetAllProfilesEndpoint(response http.ResponseWriter, request *http.Request) {
	response.Header().Set("content-type", "application/json")
	response.Header().Set("Access-Control-Allow-Origin", "http://localhost:3000")
	errorsList := make([]string, 0)
	profiles, err := queries.GetAllProfiles()
	if err != nil {
		errorsList = append(errorsList, errors.GetAllProfilesError+err.Error())
	}
	result := &responses.ProfileResponse{Success: true, Data: profiles, Errors: make([]string, 0)}
	if len(errorsList) > 0 {
		response.WriteHeader(http.StatusBadRequest)
		result = &responses.ProfileResponse{Success: false, Data: make([]entities.Profile, 0), Errors: errorsList}
	}
	json.NewEncoder(response).Encode(result)
}

// GetProfileEndpoint handles the GET request at /api/profile/{ID}
func GetProfileEndpoint(response http.ResponseWriter, request *http.Request) {
	response.Header().Set("content-type", "application/json")
	response.Header().Set("Access-Control-Allow-Origin", "http://localhost:3000")
	var profile entities.Profile
	var err error
	errorsList := make([]string, 0)

	params := mux.Vars(request)
	ID, ok := params["ID"]
	if ok {
		profile, err = queries.GetProfile(ID)
		if err != nil {
			errorsList = append(errorsList, errors.GetProfileError+err.Error())
		}
	} else {
		errorsList = append(errorsList, errors.MissingParameterError)
	}
	result := &responses.ProfileResponse{Success: true, Data: []entities.Profile{profile}, Errors: make([]string, 0)}
	if len(errorsList) > 0 {
		response.WriteHeader(http.StatusBadRequest)
		result = &responses.ProfileResponse{Success: false, Data: make([]entities.Profile, 0), Errors: errorsList}
	}
	json.NewEncoder(response).Encode(result)
}

// CreateProfileEndpoint handles the POST request at /api/profile
func CreateProfileEndpoint(response http.ResponseWriter, request *http.Request) {
	response.Header().Set("content-type", "application/json")
	response.Header().Set("Access-Control-Allow-Origin", "http://localhost:3000")
	profile := &entities.Profile{ID: uuid.New().String()}
	errorsList := make([]string, 0)

	body, err := ioutil.ReadAll(request.Body)
	if err == nil {
		err = entities.ParseProfile(profile, body)
		if err == nil {
			profile.CreationDate = time.Now().Unix()
			err = commands.CreateProfile(*profile)
			if err != nil {
				errorsList = append(errorsList, errors.CreateProfileError+err.Error())
			}
		} else {
			errorsList = append(errorsList, errors.ParseProfileError+err.Error())
		}
	} else {
		errorsList = append(errorsList, errors.IOUtilReadAllError+err.Error())
	}
	result := &responses.ProfileResponse{Success: true, Data: []entities.Profile{*profile}, Errors: make([]string, 0)}
	if len(errorsList) > 0 {
		response.WriteHeader(http.StatusBadRequest)
		result = &responses.ProfileResponse{Success: false, Data: make([]entities.Profile, 0), Errors: errorsList}
	}
	json.NewEncoder(response).Encode(result)
}

// RemoveProfilesEndpoint handles the POST request at /api/profile/remove
func RemoveProfilesEndpoint(response http.ResponseWriter, request *http.Request) {
	response.Header().Set("content-type", "application/json")
	response.Header().Set("Access-Control-Allow-Origin", "http://localhost:3000")
	var profiles []entities.Profile
	errorsList := make([]string, 0)

	type RemoveProfilesRequest struct {
		IDs []string `json:"ids"`
	}

	body, err := ioutil.ReadAll(request.Body)
	if err == nil {
		removeProfilesRequest := RemoveProfilesRequest{}
		err = json.Unmarshal(body, &removeProfilesRequest)
		if err == nil {
			for _, id := range removeProfilesRequest.IDs {
				tasks, err := queries.GetTasksByProfileID(id)
				if err == nil {
					next := true
					taskStore := stores.GetTaskStore()
					for _, task := range tasks {
						stopped := taskStore.StopTask(&task)
						if !stopped {
							next = false
							break
						}
					}
					if next {
						err = commands.RemoveTasksByProfileID(id)
						if err == nil {
							profile, err := commands.RemoveProfile(id)
							if err == nil {
								profiles = append(profiles, profile)
							} else {
								errorsList = append(errorsList, errors.RemoveProfileError+err.Error())
							}
						} else {
							errorsList = append(errorsList, errors.RemoveProfileError+err.Error())
						}
					} else {
						errorsList = append(errorsList, errors.StopTaskError)
					}
				} else {
					errorsList = append(errorsList, errors.GetTaskError)
				}
			}
		} else {
			errorsList = append(errorsList, errors.ParseAddProfilesToGroupsRequestError+err.Error())
		}
	} else {
		errorsList = append(errorsList, errors.IOUtilReadAllError+err.Error())
	}

	data := []entities.Profile{}
	data = append(data, profiles...)

	result := &responses.ProfileResponse{Success: true, Data: data, Errors: make([]string, 0)}
	if len(errorsList) > 0 {
		response.WriteHeader(http.StatusBadRequest)
		result = &responses.ProfileResponse{Success: false, Data: make([]entities.Profile, 0), Errors: errorsList}
	}
	json.NewEncoder(response).Encode(result)
}

// UpdateProfileEndpoint handles the PUT request at /api/profile/{ID}
func UpdateProfileEndpoint(response http.ResponseWriter, request *http.Request) {
	response.Header().Set("content-type", "application/json")
	response.Header().Set("Access-Control-Allow-Origin", "http://localhost:3000")
	var profile entities.Profile
	errorsList := make([]string, 0)

	params := mux.Vars(request)
	ID, ok := params["ID"]

	if ok {
		newProfile := &entities.Profile{ID: ID}
		body, err := ioutil.ReadAll(request.Body)
		if err == nil {
			err = entities.ParseProfile(newProfile, body)
			if err == nil {
				profile, err = commands.UpdateProfile(ID, *newProfile)
				if err != nil {
					errorsList = append(errorsList, errors.UpdateProfileError+err.Error())
				}
			} else {
				errorsList = append(errorsList, errors.ParseProfileError+err.Error())
			}
		} else {
			errorsList = append(errorsList, errors.IOUtilReadAllError+err.Error())
		}
	} else {
		errorsList = append(errorsList, errors.MissingParameterError)
	}
	result := &responses.ProfileResponse{Success: true, Data: []entities.Profile{profile}, Errors: make([]string, 0)}
	if len(errorsList) > 0 {
		response.WriteHeader(http.StatusBadRequest)
		result = &responses.ProfileResponse{Success: false, Data: make([]entities.Profile, 0), Errors: errorsList}
	}
	json.NewEncoder(response).Encode(result)
}

// CloneProfilesEndpoint handles the POST request at /api/profile/clone
func CloneProfilesEndpoint(response http.ResponseWriter, request *http.Request) {
	response.Header().Set("content-type", "application/json")
	response.Header().Set("Access-Control-Allow-Origin", "http://localhost:3000")
	var profiles []entities.Profile
	var err error
	errorsList := make([]string, 0)

	type CloneProfilesRequest struct {
		IDs []string `json:"ids"`
	}
	body, err := ioutil.ReadAll(request.Body)
	if err == nil {
		cloneProfilesRequest := CloneProfilesRequest{}
		err = json.Unmarshal(body, &cloneProfilesRequest)
		if err == nil {
			for _, id := range cloneProfilesRequest.IDs {
				profile, err := queries.GetProfile(id)
				if err == nil {
					newProfileID := uuid.New().String()
					profile.SetID(newProfileID)
					profile.SetName(profile.Name + " (Copy " + common.RandID(4) + ")")
					profile.CreationDate = time.Now().Unix()
					shippingAddress := &profile.ShippingAddress
					shippingAddress.SetID(uuid.New().String())
					shippingAddress.ProfileID = newProfileID
					billingAddress := &profile.BillingAddress
					billingAddress.SetID(uuid.New().String())
					billingAddress.ProfileID = newProfileID
					creditCard := &profile.CreditCard
					creditCard.SetID(uuid.New().String())
					creditCard.ProfileID = newProfileID
					err = commands.CreateProfile(profile)
					if err == nil {
						profiles = append(profiles, profile)
					} else {
						errorsList = append(errorsList, errors.CreateProfileError+err.Error())
					}
				} else {
					errorsList = append(errorsList, errors.GetProfileError+err.Error())
				}
			}
		} else {
			errorsList = append(errorsList, errors.ParseCloneProfilesRequestError+err.Error())
		}
	} else {
		errorsList = append(errorsList, errors.IOUtilReadAllError+err.Error())
	}

	data := []entities.Profile{}
	data = append(data, profiles...)

	result := &responses.ProfileResponse{Success: true, Data: data, Errors: make([]string, 0)}
	if len(errorsList) > 0 {
		response.WriteHeader(http.StatusBadRequest)
		result = &responses.ProfileResponse{Success: false, Data: make([]entities.Profile, 0), Errors: errorsList}
	}
	json.NewEncoder(response).Encode(result)
}
