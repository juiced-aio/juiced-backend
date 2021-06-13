package endpoints

import (
	"backend.juicedbot.io/juiced.api/responses"
	"backend.juicedbot.io/juiced.infrastructure/commands"
	"backend.juicedbot.io/juiced.infrastructure/common"
	"backend.juicedbot.io/juiced.infrastructure/common/entities"
	"backend.juicedbot.io/juiced.infrastructure/common/errors"
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

// RemoveProfileGroupEndpoint handles the DELETE request at /api/profile/group/{GroupID}
func RemoveProfileGroupEndpoint(response http.ResponseWriter, request *http.Request) {
	response.Header().Set("content-type", "application/json")
	response.Header().Set("Access-Control-Allow-Origin", "http://localhost:3000")
	var profileGroup entities.ProfileGroup
	var err error
	errorsList := make([]string, 0)

	params := mux.Vars(request)
	groupID, ok := params["GroupID"]
	if ok {
		profileGroup, err = commands.RemoveProfileGroup(groupID)
		if err != nil {
			errorsList = append(errorsList, errors.RemoveProfileGroupError+err.Error())
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

// CloneProfileGroupEndpoint handles the POST request at /api/profile/group/{GroupID}/clone
func CloneProfileGroupEndpoint(response http.ResponseWriter, request *http.Request) {
	response.Header().Set("content-type", "application/json")
	response.Header().Set("Access-Control-Allow-Origin", "http://localhost:3000")
	var newProfileGroup entities.ProfileGroup
	var err error
	errorsList := make([]string, 0)

	params := mux.Vars(request)
	groupID, ok := params["GroupID"]
	if ok {
		newProfileGroup, err = queries.GetProfileGroup(groupID)
		if err == nil {
			newProfileGroup.SetGroupID(uuid.New().String())
			newProfileGroup.SetName(newProfileGroup.Name + " (Copy " + common.RandID(4) + ")")
			err = commands.CreateProfileGroup(newProfileGroup)
			if err != nil {
				errorsList = append(errorsList, errors.CreateProfileGroupError+err.Error())
			}
		} else {
			errorsList = append(errorsList, errors.GetProfileGroupError+err.Error())
		}
	} else {
		errorsList = append(errorsList, errors.MissingParameterError)
	}
	newProfileGroupWithProfiles, err := queries.ConvertProfileIDsToProfiles(&newProfileGroup)
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

// AddProfilesToGroupEndpoint handles the POST request at /api/profile/group/{GroupID}/add
func AddProfilesToGroupEndpoint(response http.ResponseWriter, request *http.Request) {
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
				type ProfileIDs struct {
					ProfileIDs []string `json:"profileIDs"`
				}
				profileIDs := ProfileIDs{}
				err := json.Unmarshal(body, &profileIDs)
				if err == nil {
					profileGroup.AddProfileIDsToGroup(profileIDs.ProfileIDs)
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

// RemoveProfilesFromGroupEndpoint handles the POST request at /api/profile/group/{GroupID}/remove
func RemoveProfilesFromGroupEndpoint(response http.ResponseWriter, request *http.Request) {
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
				type ProfileIDs struct {
					ProfileIDs []string `json:"profileIDs"`
				}
				profileIDs := ProfileIDs{}
				err := json.Unmarshal(body, &profileIDs)
				if err == nil {
					profileGroup.RemoveProfileIDsFromGroup(profileIDs.ProfileIDs)
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

// RemoveProfileEndpoint handles the DELETE request at /api/profile/{ID}
func RemoveProfileEndpoint(response http.ResponseWriter, request *http.Request) {
	response.Header().Set("content-type", "application/json")
	response.Header().Set("Access-Control-Allow-Origin", "http://localhost:3000")
	var profile entities.Profile
	var err error
	errorsList := make([]string, 0)

	params := mux.Vars(request)
	ID, ok := params["ID"]
	if ok {
		profile, err = commands.RemoveProfile(ID)
		if err != nil {
			errorsList = append(errorsList, errors.RemoveProfileError+err.Error())
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

// CloneProfileEndpoint handles the POST request at /api/profile/{ID}/clone
func CloneProfileEndpoint(response http.ResponseWriter, request *http.Request) {
	response.Header().Set("content-type", "application/json")
	response.Header().Set("Access-Control-Allow-Origin", "http://localhost:3000")
	var profile entities.Profile
	var err error
	errorsList := make([]string, 0)

	params := mux.Vars(request)
	ID, ok := params["ID"]
	if ok {
		profile, err = queries.GetProfile(ID)
		if err == nil {
			newProfileID := uuid.New().String()
			profile.SetID(newProfileID)
			profile.SetName(profile.Name + " (Copy " + common.RandID(4) + ")")
			shippingAddress := &profile.ShippingAddress
			shippingAddress.SetID(uuid.New().String())
			shippingAddress.ProfileID = newProfileID
			shippingAddress.ProfileGroupID = profile.ProfileGroupID
			billingAddress := &profile.BillingAddress
			billingAddress.SetID(uuid.New().String())
			billingAddress.ProfileID = newProfileID
			billingAddress.ProfileGroupID = profile.ProfileGroupID
			creditCard := &profile.CreditCard
			creditCard.SetID(uuid.New().String())
			creditCard.ProfileID = newProfileID
			creditCard.ProfileGroupID = profile.ProfileGroupID
			err = commands.CreateProfile(profile)
			if err != nil {
				errorsList = append(errorsList, errors.CreateProfileError+err.Error())
			}
		} else {
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
