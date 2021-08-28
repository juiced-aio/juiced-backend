package endpoints

import (
	"os"
	"strings"
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

// GetProfileGroupEndpoint handles the GET request at /api/profile/group/{groupID}
func GetProfileGroupEndpoint(response http.ResponseWriter, request *http.Request) {
	response.Header().Set("content-type", "application/json")
	response.Header().Set("Access-Control-Allow-Origin", "http://localhost:3000")
	var profileGroup entities.ProfileGroup
	errorsList := make([]string, 0)

	params := mux.Vars(request)
	groupID, ok := params["GroupID"]
	if ok {
		profileGroupPtr, err := stores.GetProfileGroup(groupID)
		if err == nil {
			profileGroup = *profileGroupPtr
		} else {
			errorsList = append(errorsList, errors.GetProfileGroupError+err.Error())
		}
	} else {
		errorsList = append(errorsList, errors.MissingParameterError)
	}

	result := &responses.ProfileGroupResponse{Success: true, Data: []entities.ProfileGroup{profileGroup}, Errors: make([]string, 0)}
	if len(errorsList) > 0 {
		response.WriteHeader(http.StatusBadRequest)
		result = &responses.ProfileGroupResponse{Success: false, Data: []entities.ProfileGroup{}, Errors: errorsList}
	}
	json.NewEncoder(response).Encode(result)
}

// GetAllProfileGroupsEndpoint handles the GET request at /api/profile/group
func GetAllProfileGroupsEndpoint(response http.ResponseWriter, request *http.Request) {
	response.Header().Set("content-type", "application/json")
	response.Header().Set("Access-Control-Allow-Origin", "http://localhost:3000")
	profileGroups := stores.GetAllProfileGroups()

	data := []entities.ProfileGroup{}
	for i := 0; i < len(profileGroups); i++ {
		data = append(data, *profileGroups[i])
	}
	result := &responses.ProfileGroupResponse{Success: true, Data: data, Errors: make([]string, 0)}
	json.NewEncoder(response).Encode(result)
}

// CreateProfileGroupEndpoint handles the POST request at /api/profile/group
func CreateProfileGroupEndpoint(response http.ResponseWriter, request *http.Request) {
	response.Header().Set("content-type", "application/json")
	response.Header().Set("Access-Control-Allow-Origin", "http://localhost:3000")
	profileGroup := entities.ProfileGroup{GroupID: uuid.New().String(), ProfileIDs: []string{}}
	errorsList := make([]string, 0)

	body, err := ioutil.ReadAll(request.Body)
	if err == nil {
		err = json.Unmarshal(body, &profileGroup)
		if err == nil {
			profileGroupPtr, err := stores.CreateProfileGroup(profileGroup)
			if err == nil {
				profileGroup = *profileGroupPtr
			} else {
				errorsList = append(errorsList, errors.CreateProfileGroupError+err.Error())
			}
		} else {
			errorsList = append(errorsList, errors.ParseProfileGroupError+err.Error())
		}
	} else {
		errorsList = append(errorsList, errors.IOUtilReadAllError+err.Error())
	}
	if err != nil {
		errorsList = append(errorsList, errors.GetProfileError+err.Error())
	}

	result := &responses.ProfileGroupResponse{Success: true, Data: []entities.ProfileGroup{profileGroup}, Errors: make([]string, 0)}
	if len(errorsList) > 0 {
		response.WriteHeader(http.StatusBadRequest)
		result = &responses.ProfileGroupResponse{Success: false, Data: make([]entities.ProfileGroup, 0), Errors: errorsList}
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
		profileGroup, err = stores.RemoveProfileGroup(groupID)
		if err != nil {
			errorsList = append(errorsList, errors.RemoveProfileGroupError+err.Error())
		}
	} else {
		errorsList = append(errorsList, errors.MissingParameterError)
	}

	result := &responses.ProfileGroupResponse{Success: true, Data: []entities.ProfileGroup{profileGroup}, Errors: make([]string, 0)}
	if len(errorsList) > 0 {
		response.WriteHeader(http.StatusBadRequest)
		result = &responses.ProfileGroupResponse{Success: false, Data: make([]entities.ProfileGroup, 0), Errors: errorsList}
	}
	json.NewEncoder(response).Encode(result)
}

// UpdateProfileGroupEndpoint handles the PUT request at /api/profile/group/{GroupID}
func UpdateProfileGroupEndpoint(response http.ResponseWriter, request *http.Request) {
	response.Header().Set("content-type", "application/json")
	response.Header().Set("Access-Control-Allow-Origin", "http://localhost:3000")
	var profileGroup entities.ProfileGroup
	errorsList := make([]string, 0)

	params := mux.Vars(request)
	groupID, ok := params["GroupID"]
	if ok {
		profileGroupPtr, err := stores.GetProfileGroup(groupID)
		if err == nil {
			body, err := ioutil.ReadAll(request.Body)
			if err == nil {
				type Name struct {
					Name string `json:"name"`
				}
				name := Name{}
				err := json.Unmarshal(body, &name)
				if err == nil {
					profileGroupPtr.Name = name.Name
					profileGroupPtr, err = stores.UpdateProfileGroup(groupID, *profileGroupPtr)
					if err == nil {
						profileGroup = *profileGroupPtr
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
	} else {
		errorsList = append(errorsList, errors.MissingParameterError)
	}

	result := &responses.ProfileGroupResponse{Success: true, Data: []entities.ProfileGroup{profileGroup}, Errors: make([]string, 0)}
	if len(errorsList) > 0 {
		response.WriteHeader(http.StatusBadRequest)
		result = &responses.ProfileGroupResponse{Success: false, Data: make([]entities.ProfileGroup, 0), Errors: errorsList}
	}
	json.NewEncoder(response).Encode(result)
}

// CloneProfileGroupEndpoint handles the POST request at /api/profile/group/{GroupID}/clone
func CloneProfileGroupEndpoint(response http.ResponseWriter, request *http.Request) {
	response.Header().Set("content-type", "application/json")
	response.Header().Set("Access-Control-Allow-Origin", "http://localhost:3000")
	var newProfileGroup entities.ProfileGroup
	errorsList := make([]string, 0)

	params := mux.Vars(request)
	groupID, ok := params["GroupID"]
	if ok {
		oldProfileGroupPtr, err := stores.GetProfileGroup(groupID)
		if err == nil {
			newProfileGroup = *oldProfileGroupPtr
			newProfileGroup.GroupID = uuid.New().String()
			newProfileGroup.Name = newProfileGroup.Name + " (Copy " + common.RandID(3) + ")"
			newProfileGroup.CreationDate = time.Now().Unix()
			newProfileGroupPtr, err := stores.CreateProfileGroup(newProfileGroup)
			if err == nil {
				newProfileGroup = *newProfileGroupPtr
			} else {
				errorsList = append(errorsList, errors.CreateProfileGroupError+err.Error())
			}
		} else {
			errorsList = append(errorsList, errors.GetProfileGroupError+err.Error())
		}
	} else {
		errorsList = append(errorsList, errors.MissingParameterError)
	}

	result := &responses.ProfileGroupResponse{Success: true, Data: []entities.ProfileGroup{newProfileGroup}, Errors: make([]string, 0)}
	if len(errorsList) > 0 {
		response.WriteHeader(http.StatusBadRequest)
		result = &responses.ProfileGroupResponse{Success: false, Data: make([]entities.ProfileGroup, 0), Errors: errorsList}
	}
	json.NewEncoder(response).Encode(result)
}

// AddProfilesToGroupEndpoint handles the POST request at /api/profile/group/{GroupID}/add
func AddProfilesToGroupEndpoint(response http.ResponseWriter, request *http.Request) {
	response.Header().Set("content-type", "application/json")
	response.Header().Set("Access-Control-Allow-Origin", "http://localhost:3000")
	var profileGroup entities.ProfileGroup
	errorsList := make([]string, 0)

	params := mux.Vars(request)
	groupID, ok := params["GroupID"]
	if ok {
		profileGroupPtr, err := stores.GetProfileGroup(groupID)
		if err == nil {
			body, err := ioutil.ReadAll(request.Body)
			if err == nil {
				type ProfileIDs struct {
					ProfileIDs []string `json:"profileIDs"`
				}
				profileIDs := ProfileIDs{}
				err := json.Unmarshal(body, &profileIDs)
				if err == nil {
					profileGroupPtr.AddProfileIDsToGroup(profileIDs.ProfileIDs)
					profileGroupPtr, err = stores.UpdateProfileGroup(groupID, *profileGroupPtr)
					if err == nil {
						profileGroup = *profileGroupPtr
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
	} else {
		errorsList = append(errorsList, errors.MissingParameterError)
	}

	result := &responses.ProfileGroupResponse{Success: true, Data: []entities.ProfileGroup{profileGroup}, Errors: make([]string, 0)}
	if len(errorsList) > 0 {
		response.WriteHeader(http.StatusBadRequest)
		result = &responses.ProfileGroupResponse{Success: false, Data: make([]entities.ProfileGroup, 0), Errors: errorsList}
	}
	json.NewEncoder(response).Encode(result)
}

// RemoveProfilesFromGroupEndpoint handles the POST request at /api/profile/group/{GroupID}/remove
func RemoveProfilesFromGroupEndpoint(response http.ResponseWriter, request *http.Request) {
	response.Header().Set("content-type", "application/json")
	response.Header().Set("Access-Control-Allow-Origin", "http://localhost:3000")
	var profileGroup entities.ProfileGroup
	errorsList := make([]string, 0)

	params := mux.Vars(request)
	groupID, ok := params["GroupID"]
	if ok {
		profileGroupPtr, err := stores.GetProfileGroup(groupID)
		if err == nil {
			body, err := ioutil.ReadAll(request.Body)
			if err == nil {
				type ProfileIDs struct {
					ProfileIDs []string `json:"profileIDs"`
				}
				profileIDs := ProfileIDs{}
				err := json.Unmarshal(body, &profileIDs)
				if err == nil {
					profileGroupPtr.RemoveProfileIDsFromGroup(profileIDs.ProfileIDs)
					profileGroupPtr, err = stores.UpdateProfileGroup(groupID, *profileGroupPtr)
					if err == nil {
						profileGroup = *profileGroupPtr
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
	} else {
		errorsList = append(errorsList, errors.MissingParameterError)
	}

	data := []entities.ProfileGroup{profileGroup}
	result := &responses.ProfileGroupResponse{Success: true, Data: data, Errors: make([]string, 0)}
	if len(errorsList) > 0 {
		response.WriteHeader(http.StatusBadRequest)
		result = &responses.ProfileGroupResponse{Success: false, Data: make([]entities.ProfileGroup, 0), Errors: errorsList}
	}
	json.NewEncoder(response).Encode(result)
}

// GetAllProfilesEndpoint handles the GET request at /api/profile/all
func GetAllProfilesEndpoint(response http.ResponseWriter, request *http.Request) {
	response.Header().Set("content-type", "application/json")
	response.Header().Set("Access-Control-Allow-Origin", "http://localhost:3000")
	profiles := stores.GetAllProfiles()

	data := []entities.Profile{}
	for i := 0; i < len(profiles); i++ {
		data = append(data, *profiles[i])
	}
	result := &responses.ProfileResponse{Success: true, Data: data, Errors: make([]string, 0)}
	json.NewEncoder(response).Encode(result)
}

// GetProfileEndpoint handles the GET request at /api/profile/{ID}
func GetProfileEndpoint(response http.ResponseWriter, request *http.Request) {
	response.Header().Set("content-type", "application/json")
	response.Header().Set("Access-Control-Allow-Origin", "http://localhost:3000")
	var profile entities.Profile
	errorsList := make([]string, 0)

	params := mux.Vars(request)
	ID, ok := params["ID"]
	if ok {
		profilePtr, err := stores.GetProfile(ID)
		if err == nil {
			profile = *profilePtr
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

// CreateProfileEndpoint handles the POST request at /api/profile
func CreateProfileEndpoint(response http.ResponseWriter, request *http.Request) {
	response.Header().Set("content-type", "application/json")
	response.Header().Set("Access-Control-Allow-Origin", "http://localhost:3000")
	profile := entities.Profile{ID: uuid.New().String()}
	errorsList := make([]string, 0)

	body, err := ioutil.ReadAll(request.Body)
	if err == nil {
		err = json.Unmarshal(body, &profile)
		if err == nil {
			profile.CreationDate = time.Now().Unix()
			profilePtr, err := stores.CreateProfile(profile)
			if err == nil {
				profile = *profilePtr
			} else {
				errorsList = append(errorsList, errors.CreateProfileError+err.Error())
			}
		} else {
			errorsList = append(errorsList, errors.ParseProfileError+err.Error())
		}
	} else {
		errorsList = append(errorsList, errors.IOUtilReadAllError+err.Error())
	}

	result := &responses.ProfileResponse{Success: true, Data: []entities.Profile{profile}, Errors: make([]string, 0)}
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
	errorsList := make([]string, 0)

	params := mux.Vars(request)
	ID, ok := params["ID"]
	if ok {
		tasks, err := stores.GetTasksByProfileID(ID)
		if err == nil {
			next := true
			taskStore := stores.GetTaskStore()
			for _, task := range tasks {
				_, err = taskStore.StopTask(&task)
				if err != nil {
					next = false
					break
				}
			}
			if next {
				err = stores.RemoveTasksByProfileID(ID)
				if err == nil {
					profile, err = stores.RemoveProfile(ID)
					if err != nil {
						errorsList = append(errorsList, errors.RemoveProfileError+err.Error())
					}
				} else {
					errorsList = append(errorsList, errors.RemoveProfileError+err.Error())
				}
			} else {
				errorsList = append(errorsList, errors.StopTaskError+err.Error())
			}
		} else {
			errorsList = append(errorsList, errors.GetTaskError+err.Error())
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
	var newProfile entities.Profile
	errorsList := make([]string, 0)

	params := mux.Vars(request)
	ID, ok := params["ID"]

	if ok {
		newProfile = entities.Profile{ID: ID}
		body, err := ioutil.ReadAll(request.Body)
		if err == nil {
			err = json.Unmarshal(body, &newProfile)
			if err == nil {
				profilePtr, err := stores.UpdateProfile(ID, newProfile)
				if err == nil {
					newProfile = *profilePtr
				} else {
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

	result := &responses.ProfileResponse{Success: true, Data: []entities.Profile{newProfile}, Errors: make([]string, 0)}
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
	var newProfile entities.Profile
	errorsList := make([]string, 0)

	params := mux.Vars(request)
	ID, ok := params["ID"]
	if ok {
		profile, err := stores.GetProfile(ID)
		if err == nil {
			newProfile := *profile
			newProfile.ID = uuid.New().String()
			newProfile.Name = profile.Name + " (Copy " + common.RandID(3) + ")"
			newProfile.CreationDate = time.Now().Unix()
			newProfile.ShippingAddress.ID = uuid.New().String()
			newProfile.BillingAddress.ID = uuid.New().String()
			newProfile.CreditCard.ID = uuid.New().String()
			newProfilePtr, err := stores.CreateProfile(newProfile)
			if err == nil {
				newProfile = *newProfilePtr
			} else {
				errorsList = append(errorsList, errors.CreateProfileError+err.Error())
			}
		} else {
			errorsList = append(errorsList, errors.GetProfileError+err.Error())
		}
	} else {
		errorsList = append(errorsList, errors.MissingParameterError)
	}
	result := &responses.ProfileResponse{Success: true, Data: []entities.Profile{newProfile}, Errors: make([]string, 0)}
	if len(errorsList) > 0 {
		response.WriteHeader(http.StatusBadRequest)
		result = &responses.ProfileResponse{Success: false, Data: make([]entities.Profile, 0), Errors: errorsList}
	}
	json.NewEncoder(response).Encode(result)
}

// ImportProfilesEndpoint handles the POST request at /api/profile/import
func ImportProfilesEndpoint(response http.ResponseWriter, request *http.Request) {
	response.Header().Set("content-type", "application/json")
	response.Header().Set("Access-Control-Allow-Origin", "http://localhost:3000")
	newProfiles := []entities.Profile{}
	skippedProfiles := 0
	skippedGroups := 0
	var err error
	errorsList := make([]string, 0)

	type ImportProfilesRequest struct {
		FilePath string   `json:"filePath"`
		GroupIDs []string `json:"groupIDs"`
	}

	type ImportProfilesFileFormat struct {
		Profiles []entities.Profile `json:"profiles"`
	}

	body, err := ioutil.ReadAll(request.Body)
	if err == nil {
		importProfilesRequestInfo := ImportProfilesRequest{}
		err = json.Unmarshal(body, &importProfilesRequestInfo)
		if err == nil {
			validGroupIDs := []string{}
			validGroups := []*entities.ProfileGroup{}
			for _, groupID := range importProfilesRequestInfo.GroupIDs {
				group, err := stores.GetProfileGroup(groupID)
				if err == nil && group.GroupID != "" {
					validGroups = append(validGroups, group)
					validGroupIDs = append(validGroupIDs, groupID)
				} else {
					skippedGroups++
				}
			}

			file, err := os.Open(importProfilesRequestInfo.FilePath)
			if err == nil {
				defer file.Close()
				byteValue, err := ioutil.ReadAll(file)
				if err == nil {
					profiles := ImportProfilesFileFormat{}
					err = json.Unmarshal(byteValue, &profiles)
					if err == nil && len(profiles.Profiles) > 0 {
						newProfileIDs := []string{}
						for _, newProfile := range profiles.Profiles {
							// TODO @silent: Validate all fields
							existingProfile, err := stores.GetProfileByName(newProfile.Name)
							if err != nil || existingProfile.ID == "" || existingProfile.Name == "" {
								newProfile.ID = uuid.New().String()
								newProfile.ProfileGroupIDs = validGroupIDs
								if len(validGroupIDs) > 0 {
									newProfile.ProfileGroupIDsJoined = strings.Join(validGroupIDs, ",")
								}
								newProfile.CreationDate = time.Now().Unix()

								newProfile.ShippingAddress.ID = uuid.New().String()
								newProfile.BillingAddress.ID = uuid.New().String()
								newProfile.CreditCard.ID = uuid.New().String()
								cardType := common.DetectCardType([]byte(newProfile.CreditCard.CardNumber))
								if cardType != "" {
									newProfile.CreditCard.CardType = cardType

									newProfilePtr, err := stores.CreateProfile(newProfile)
									if err == nil {
										newProfile = *newProfilePtr
										newProfiles = append(newProfiles, newProfile)
										newProfileIDs = append(newProfileIDs, newProfile.ID)
									} else {
										skippedProfiles++
									}
								} else {
									skippedProfiles++
								}
							} else {
								skippedProfiles++
							}
						}

						for _, group := range validGroups {
							group.AddProfileIDsToGroup(newProfileIDs)
							_, err = stores.UpdateProfileGroup(group.GroupID, *group)
							if err != nil {
								skippedGroups++
							}
						}
					} else {
						if err != nil {
							errorsList = append(errorsList, errors.ParseImportProfilesFileError+err.Error())
						} else {
							errorsList = append(errorsList, errors.ParseImportProfilesFileError+"No profiles detected.")
						}
					}
				} else {
					errorsList = append(errorsList, errors.ReadFileError+err.Error())
				}
			} else {
				errorsList = append(errorsList, errors.OpenFileError+err.Error())
			}
		} else {
			errorsList = append(errorsList, errors.ParseImportProfilesRequestError+err.Error())
		}
	} else {
		errorsList = append(errorsList, errors.IOUtilReadAllError+err.Error())
	}

	profileGroups := stores.GetAllProfileGroups()
	if err != nil {
		errorsList = append(errorsList, errors.GetAllProfileGroupsError+err.Error())
	}

	data := []entities.ProfileGroup{}
	for i := 0; i < len(profileGroups); i++ {
		data = append(data, *profileGroups[i])
	}
	result := &responses.ImportProfileResponse{Success: true, NewProfiles: newProfiles, SkippedProfiles: skippedProfiles, SkippedGroups: skippedGroups, Data: data, Errors: make([]string, 0)}
	if len(errorsList) > 0 {
		response.WriteHeader(http.StatusBadRequest)
		result = &responses.ImportProfileResponse{Success: false, NewProfiles: []entities.Profile{}, SkippedProfiles: 0, SkippedGroups: 0, Data: make([]entities.ProfileGroup, 0), Errors: errorsList}
	}
	json.NewEncoder(response).Encode(result)
}
