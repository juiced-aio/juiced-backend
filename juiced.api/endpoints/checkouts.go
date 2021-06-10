package endpoints

import (
	"backend.juicedbot.io/juiced.api/responses"
	"backend.juicedbot.io/juiced.infrastructure/common/entities"
	"backend.juicedbot.io/juiced.infrastructure/common/errors"
	"backend.juicedbot.io/juiced.infrastructure/queries"

	"encoding/json"
	"net/http"

	"github.com/gorilla/mux"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// GetAllCheckoutsEndpoint handles the GET request at /api/checkouts
func GetAllCheckoutsEndpoint(response http.ResponseWriter, request *http.Request) {
	response.Header().Set("content-type", "application/json")
	response.Header().Set("Access-Control-Allow-Origin", "http://localhost:3000")
	checkout := entities.Checkout{}
	errorsList := make([]string, 0)

	params := mux.Vars(request)
	groupID, err := primitive.ObjectIDFromHex(params["GroupID"])
	if err == nil {
		taskGroup, err = queries.GetTaskGroup(groupID)
		if err != nil {
			errorsList = append(errorsList, errors.GetTaskGroupError+err.Error())
		}
	} else {
		errorsList = append(errorsList, errors.ParseObjectIDError+err.Error())
	}
	newTaskGroupWithTasks, err := queries.ConvertTaskIDsToTasks(&taskGroup)
	if err != nil {
		errorsList = append(errorsList, errors.GetTaskError+err.Error())
	}
	data := []entities.TaskGroupWithTasks{newTaskGroupWithTasks}
	result := &responses.TaskGroupResponse{Success: true, Data: data, Errors: make([]string, 0)}
	if len(errorsList) > 0 {
		response.WriteHeader(http.StatusBadRequest)
		result = &responses.TaskGroupResponse{Success: false, Data: data, Errors: errorsList}
	}
	json.NewEncoder(response).Encode(result)
}
