package endpoints

import (
	"backend.juicedbot.io/juiced.api/responses"
	"backend.juicedbot.io/juiced.infrastructure/commands"
	"backend.juicedbot.io/juiced.infrastructure/common"
	"backend.juicedbot.io/juiced.infrastructure/common/entities"
	"backend.juicedbot.io/juiced.infrastructure/common/enums"
	"backend.juicedbot.io/juiced.infrastructure/common/errors"
	"backend.juicedbot.io/juiced.infrastructure/common/stores"
	"backend.juicedbot.io/juiced.infrastructure/queries"

	"encoding/json"
	"io/ioutil"
	"net/http"

	"github.com/gorilla/mux"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// GetTaskGroupEndpoint handles the GET request at /api/task/group/{groupID}
func GetTaskGroupEndpoint(response http.ResponseWriter, request *http.Request) {
	response.Header().Set("content-type", "application/json")
	response.Header().Set("Access-Control-Allow-Origin", "http://localhost:3000")
	taskGroup := entities.TaskGroup{}
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

// GetAllTaskGroupsEndpoint handles the GET request at /api/task/group
func GetAllTaskGroupsEndpoint(response http.ResponseWriter, request *http.Request) {
	response.Header().Set("content-type", "application/json")
	response.Header().Set("Access-Control-Allow-Origin", "http://localhost:3000")
	errorsList := make([]string, 0)
	taskGroups, err := queries.GetAllTaskGroups()
	if err != nil {
		errorsList = append(errorsList, errors.GetAllTaskGroupsError+err.Error())
	}
	data := []entities.TaskGroupWithTasks{}
	for i := 0; i < len(taskGroups); i++ {
		newTaskGroupWithTasks, err := queries.ConvertTaskIDsToTasks(&taskGroups[i])
		if err != nil {
			errorsList = append(errorsList, errors.GetTaskError+err.Error())
		}
		data = append(data, newTaskGroupWithTasks)
	}
	result := &responses.TaskGroupResponse{Success: true, Data: data, Errors: make([]string, 0)}
	if len(errorsList) > 0 {
		response.WriteHeader(http.StatusBadRequest)
		result = &responses.TaskGroupResponse{Success: false, Data: data, Errors: errorsList}
	}
	json.NewEncoder(response).Encode(result)
}

// CreateTaskGroupEndpoint handles the POST request at /api/task/group
func CreateTaskGroupEndpoint(response http.ResponseWriter, request *http.Request) {
	response.Header().Set("content-type", "application/json")
	response.Header().Set("Access-Control-Allow-Origin", "http://localhost:3000")
	taskGroup := &entities.TaskGroup{GroupID: primitive.NewObjectID(), TaskIDs: []primitive.ObjectID{}, MonitorDelay: 2000}
	errorsList := make([]string, 0)

	body, err := ioutil.ReadAll(request.Body)
	if err == nil {
		err = entities.ParseTaskGroup(taskGroup, body)
		if err == nil {
			err = commands.CreateTaskGroup(*taskGroup)
			if err != nil {
				errorsList = append(errorsList, errors.CreateTaskGroupError+err.Error())
			}
		} else {
			errorsList = append(errorsList, errors.ParseTaskGroupError+err.Error())
		}
	} else {
		errorsList = append(errorsList, errors.IOUtilReadAllError+err.Error())
	}
	newTaskGroupWithTasks, err := queries.ConvertTaskIDsToTasks(taskGroup)
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

// RemoveTaskGroupEndpoint handles the DELETE request at /api/task/group/{GroupID}
func RemoveTaskGroupEndpoint(response http.ResponseWriter, request *http.Request) {
	response.Header().Set("content-type", "application/json")
	response.Header().Set("Access-Control-Allow-Origin", "http://localhost:3000")
	taskGroup := entities.TaskGroup{}
	errorsList := make([]string, 0)

	params := mux.Vars(request)
	groupID, err := primitive.ObjectIDFromHex(params["GroupID"])
	if err == nil {
		taskGroup, err = commands.RemoveTaskGroup(groupID)
		if err != nil {
			errorsList = append(errorsList, errors.RemoveTaskGroupError+err.Error())
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

// UpdateTaskGroupEndpoint handles the PUT request at /api/task/group/{GroupID}
func UpdateTaskGroupEndpoint(response http.ResponseWriter, request *http.Request) {
	response.Header().Set("content-type", "application/json")
	response.Header().Set("Access-Control-Allow-Origin", "http://localhost:3000")
	errorsList := make([]string, 0)

	params := mux.Vars(request)
	groupID, err := primitive.ObjectIDFromHex(params["GroupID"])
	newTaskGroup := entities.TaskGroup{GroupID: groupID}
	if err == nil {
		oldTaskGroup, err := queries.GetTaskGroup(groupID)
		if err == nil {
			body, err := ioutil.ReadAll(request.Body)
			if err == nil {
				err = entities.ParseTaskGroup(&newTaskGroup, body)
				newTaskGroup.SetGroupID(oldTaskGroup.GroupID)
				newTaskGroup.SetMonitorRetailer(oldTaskGroup.MonitorRetailer)
				newTaskGroup.SetMonitorStatus(oldTaskGroup.MonitorStatus)
				newTaskGroup.AddTasksToGroup(oldTaskGroup.TaskIDs)
				if err == nil {
					newTaskGroup, err = commands.UpdateTaskGroup(groupID, newTaskGroup)
					if err != nil {
						errorsList = append(errorsList, errors.UpdateTaskGroupError+err.Error())
					}
				} else {
					errorsList = append(errorsList, errors.ParseTaskGroupError+err.Error())
				}
			} else {
				errorsList = append(errorsList, errors.IOUtilReadAllError+err.Error())
			}
		} else {
			errorsList = append(errorsList, errors.GetTaskGroupError+err.Error())
		}
	} else {
		errorsList = append(errorsList, errors.ParseObjectIDError+err.Error())
	}
	newTaskGroupWithTasks, err := queries.ConvertTaskIDsToTasks(&newTaskGroup)
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

// CloneTaskGroupEndpoint handles the POST request at /api/task/group/{GroupID}/clone
func CloneTaskGroupEndpoint(response http.ResponseWriter, request *http.Request) {
	response.Header().Set("content-type", "application/json")
	response.Header().Set("Access-Control-Allow-Origin", "http://localhost:3000")
	errorsList := make([]string, 0)

	params := mux.Vars(request)
	groupID, err := primitive.ObjectIDFromHex(params["GroupID"])
	newTaskGroup := entities.TaskGroup{}
	if err == nil {
		newTaskGroup, err = queries.GetTaskGroup(groupID)
		if err == nil {
			newTaskGroup.SetGroupID(primitive.NewObjectID())
			newTaskGroup.SetName(newTaskGroup.Name + " (Copy " + common.RandID(4) + ")")
			err = commands.CreateTaskGroup(newTaskGroup)
			if err != nil {
				errorsList = append(errorsList, errors.CreateTaskGroupError+err.Error())
			}
		} else {
			errorsList = append(errorsList, errors.GetTaskGroupError+err.Error())
		}
	} else {
		errorsList = append(errorsList, errors.ParseObjectIDError+err.Error())
	}
	newTaskGroupWithTasks, err := queries.ConvertTaskIDsToTasks(&newTaskGroup)
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

// RemoveTasksEndpoint handles the POST request at api/task/group/{GroupID}/removeTasks
func RemoveTasksEndpoint(response http.ResponseWriter, request *http.Request) {
	response.Header().Set("content-type", "application/json")
	response.Header().Set("Access-Control-Allow-Origin", "http://localhost:3000")
	errorsList := make([]string, 0)

	type DeleteTasksRequest struct {
		TaskIDs []primitive.ObjectID `json:"taskIDs"`
	}

	params := mux.Vars(request)
	groupID, err := primitive.ObjectIDFromHex(params["GroupID"])
	newTaskGroup := entities.TaskGroup{}
	if err == nil {
		body, err := ioutil.ReadAll(request.Body)
		if err == nil {
			newTaskGroup, err = queries.GetTaskGroup(groupID)
			if err == nil {
				deleteTasksRequestInfo := DeleteTasksRequest{}
				err = json.Unmarshal(body, &deleteTasksRequestInfo)
				if err == nil {
					newTaskIDs := make([]primitive.ObjectID, 0)
					for i := 0; i < len(newTaskGroup.TaskIDs); i++ {
						found := false
						for j := 0; j < len(deleteTasksRequestInfo.TaskIDs); j++ {
							if newTaskGroup.TaskIDs[i] == deleteTasksRequestInfo.TaskIDs[j] {
								found = true
							}
						}
						if !found {
							newTaskIDs = append(newTaskIDs, newTaskGroup.TaskIDs[i])
						}
					}
					newTaskGroup.SetTaskIDs(newTaskIDs)
					newTaskGroup, err = commands.UpdateTaskGroup(groupID, newTaskGroup)
					if err == nil {
						for i := 0; i < len(deleteTasksRequestInfo.TaskIDs); i++ {
							_, err = commands.RemoveTask(deleteTasksRequestInfo.TaskIDs[i])
							if err != nil {
								errorsList = append(errorsList, errors.RemoveTaskError+err.Error())
							}
						}
					} else {
						errorsList = append(errorsList, errors.RemoveTasksFromGroupError+err.Error())
					}
				} else {
					errorsList = append(errorsList, errors.ParseDeleteTasksRequestError+err.Error())
				}
			} else {
				errorsList = append(errorsList, errors.GetTaskGroupError+err.Error())
			}
		} else {
			errorsList = append(errorsList, errors.IOUtilReadAllError+err.Error())
		}
	} else {
		errorsList = append(errorsList, errors.ParseObjectIDError+err.Error())
	}
	newTaskGroupWithTasks, err := queries.ConvertTaskIDsToTasks(&newTaskGroup)
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

// GetAllTasksEndpoint handles the GET request at /api/task/all
func GetAllTasksEndpoint(response http.ResponseWriter, request *http.Request) {
	response.Header().Set("content-type", "application/json")
	response.Header().Set("Access-Control-Allow-Origin", "http://localhost:3000")
	errorsList := make([]string, 0)
	tasks, err := queries.GetAllTasks()
	if err != nil {
		errorsList = append(errorsList, errors.GetAllTasksError+err.Error())
	}
	result := &responses.TaskResponse{Success: true, Data: tasks, Errors: make([]string, 0)}
	if len(errorsList) > 0 {
		response.WriteHeader(http.StatusBadRequest)
		result = &responses.TaskResponse{Success: false, Data: make([]entities.Task, 0), Errors: errorsList}
	}
	json.NewEncoder(response).Encode(result)
}

// GetTaskEndpoint handles the GET request at /api/task/{ID}
func GetTaskEndpoint(response http.ResponseWriter, request *http.Request) {
	response.Header().Set("content-type", "application/json")
	response.Header().Set("Access-Control-Allow-Origin", "http://localhost:3000")
	task := entities.Task{}
	errorsList := make([]string, 0)

	params := mux.Vars(request)
	ID, err := primitive.ObjectIDFromHex(params["ID"])
	if err == nil {
		task, err = queries.GetTask(ID)
		if err != nil {
			errorsList = append(errorsList, errors.GetTaskError+err.Error())
		}
	} else {
		errorsList = append(errorsList, errors.ParseObjectIDError+err.Error())
	}
	result := &responses.TaskResponse{Success: true, Data: []entities.Task{task}, Errors: make([]string, 0)}
	if len(errorsList) > 0 {
		response.WriteHeader(http.StatusBadRequest)
		result = &responses.TaskResponse{Success: false, Data: make([]entities.Task, 0), Errors: errorsList}
	}
	json.NewEncoder(response).Encode(result)
}

// CreateTaskEndpoint handles the POST request at /api/task/{groupID}
func CreateTaskEndpoint(response http.ResponseWriter, request *http.Request) {
	response.Header().Set("content-type", "application/json")
	response.Header().Set("Access-Control-Allow-Origin", "http://localhost:3000")
	task := &entities.Task{ID: primitive.NewObjectID(), TaskSize: make([]string, 0), TaskQty: 1, TaskStatus: enums.TaskIdle}
	newTaskGroup := entities.TaskGroup{}
	errorsList := make([]string, 0)

	type CreateTaskRequest struct {
		NumTasksPerProfile int                  `json:"numTasksPerProfile"`
		ProfileIDs         []primitive.ObjectID `json:"profileIDs"`
		ProfileGroupID     primitive.ObjectID   `json:"profileGroupID"`
	}

	params := mux.Vars(request)
	groupID, err := primitive.ObjectIDFromHex(params["GroupID"])
	if err == nil {
		task.TaskGroupID = groupID
		body, err := ioutil.ReadAll(request.Body)
		if err == nil {
			err = entities.ParseTask(task, body)
			if err == nil {
				createTaskRequestInfo := CreateTaskRequest{}
				err = json.Unmarshal(body, &createTaskRequestInfo)
				if err == nil {
					var profileIDs []primitive.ObjectID
					if primitive.ObjectID.IsZero(createTaskRequestInfo.ProfileGroupID) {
						profileIDs = createTaskRequestInfo.ProfileIDs
					} else {
						profileGroup, err := queries.GetProfileGroup(createTaskRequestInfo.ProfileGroupID)
						if err != nil {
							errorsList = append(errorsList, errors.GetProfileGroupError+err.Error())
						}
						profileIDs = profileGroup.ProfileIDs
					}
					if err == nil {
						for i := 0; i < len(profileIDs); i++ {
							task.SetTaskProfileID(profileIDs[i])
							for j := 0; j < createTaskRequestInfo.NumTasksPerProfile; j++ {
								task.SetID(primitive.NewObjectID())
								err = commands.CreateTask(*task)
								if err == nil {
									oldTaskGroup, err := queries.GetTaskGroup(groupID)
									if err == nil {
										oldTaskGroup.SetTaskIDs(append(oldTaskGroup.TaskIDs, task.ID))
										newTaskGroup, err = commands.UpdateTaskGroup(groupID, oldTaskGroup)
										if err != nil {
											errorsList = append(errorsList, errors.AddTaskToGroupError+err.Error())
										}
									} else {
										errorsList = append(errorsList, errors.AddTaskToGroupError+err.Error())
									}
								} else {
									errorsList = append(errorsList, errors.CreateTaskError+err.Error())
								}
							}
						}
					}
				} else {
					errorsList = append(errorsList, errors.ParseTaskError+err.Error())
				}
			} else {
				errorsList = append(errorsList, errors.ParseTaskError+err.Error())
			}
		} else {
			errorsList = append(errorsList, errors.IOUtilReadAllError+err.Error())
		}
	} else {
		errorsList = append(errorsList, errors.ParseObjectIDError+err.Error())
	}

	newTaskGroupWithTasks, err := queries.ConvertTaskIDsToTasks(&newTaskGroup)
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

// UpdateTaskEndpoint handles the PUT request at /api/task/{ID}
func UpdateTaskEndpoint(response http.ResponseWriter, request *http.Request) {
	response.Header().Set("content-type", "application/json")
	response.Header().Set("Access-Control-Allow-Origin", "http://localhost:3000")
	task := entities.Task{}
	errorsList := make([]string, 0)

	params := mux.Vars(request)
	ID, err := primitive.ObjectIDFromHex(params["ID"])
	newTask := &entities.Task{ID: ID}
	if err == nil {
		body, err := ioutil.ReadAll(request.Body)
		if err == nil {
			err = entities.ParseTask(newTask, body)
			if err == nil {
				task, err = commands.UpdateTask(ID, *newTask)
				if err != nil {
					errorsList = append(errorsList, errors.UpdateTaskError+err.Error())
				}
			} else {
				errorsList = append(errorsList, errors.ParseTaskError+err.Error())
			}
		} else {
			errorsList = append(errorsList, errors.IOUtilReadAllError+err.Error())
		}
	} else {
		errorsList = append(errorsList, errors.ParseObjectIDError+err.Error())
	}
	result := &responses.TaskResponse{Success: true, Data: []entities.Task{task}, Errors: make([]string, 0)}
	if len(errorsList) > 0 {
		response.WriteHeader(http.StatusBadRequest)
		result = &responses.TaskResponse{Success: false, Data: make([]entities.Task, 0), Errors: errorsList}
	}
	json.NewEncoder(response).Encode(result)
}

// CloneTaskEndpoint handles the POST request at /api/task/{ID}/clone
func CloneTaskEndpoint(response http.ResponseWriter, request *http.Request) {
	response.Header().Set("content-type", "application/json")
	response.Header().Set("Access-Control-Allow-Origin", "http://localhost:3000")
	task := entities.Task{}
	errorsList := make([]string, 0)

	params := mux.Vars(request)
	ID, err := primitive.ObjectIDFromHex(params["ID"])
	if err == nil {
		task, err = queries.GetTask(ID)
		if err == nil {
			task.SetID(primitive.NewObjectID())
			err = commands.CreateTask(task)
			if err != nil {
				errorsList = append(errorsList, errors.CreateTaskError+err.Error())
			}
		} else {
			errorsList = append(errorsList, errors.GetTaskError+err.Error())
		}
	} else {
		errorsList = append(errorsList, errors.ParseObjectIDError+err.Error())
	}
	result := &responses.TaskResponse{Success: true, Data: []entities.Task{task}, Errors: make([]string, 0)}
	if len(errorsList) > 0 {
		response.WriteHeader(http.StatusBadRequest)
		result = &responses.TaskResponse{Success: false, Data: make([]entities.Task, 0), Errors: errorsList}
	}
	json.NewEncoder(response).Encode(result)
}

// StartTaskEndpoint handles the POST request at /api/task/{ID}/start
func StartTaskEndpoint(response http.ResponseWriter, request *http.Request) {
	response.Header().Set("content-type", "application/json")
	response.Header().Set("Access-Control-Allow-Origin", "http://localhost:3000")
	taskToStart := entities.Task{}
	errorsList := make([]string, 0)

	params := mux.Vars(request)
	ID, err := primitive.ObjectIDFromHex(params["ID"])
	if err == nil {
		taskToStart, err = queries.GetTask(ID)
		if err == nil {
			taskStore := stores.GetTaskStore()
			started := taskStore.StartTask(&taskToStart)
			if !started {
				errorsList = append(errorsList, errors.StartTaskError)
			}
		} else {
			errorsList = append(errorsList, errors.GetTaskError+err.Error())
		}
	} else {
		errorsList = append(errorsList, errors.ParseObjectIDError+err.Error())
	}
	result := &responses.TaskResponse{Success: true, Data: []entities.Task{taskToStart}, Errors: make([]string, 0)}
	if len(errorsList) > 0 {
		response.WriteHeader(http.StatusBadRequest)
		result = &responses.TaskResponse{Success: false, Data: make([]entities.Task, 0), Errors: errorsList}
	}
	json.NewEncoder(response).Encode(result)
}

// StopTaskEndpoint handles the POST request at /api/task/{ID}/stop
func StopTaskEndpoint(response http.ResponseWriter, request *http.Request) {
	response.Header().Set("content-type", "application/json")
	response.Header().Set("Access-Control-Allow-Origin", "http://localhost:3000")
	taskToStop := entities.Task{}
	errorsList := make([]string, 0)

	params := mux.Vars(request)
	ID, err := primitive.ObjectIDFromHex(params["ID"])
	if err == nil {
		taskToStop, err = queries.GetTask(ID)
		if err == nil {
			taskStore := stores.GetTaskStore()
			started := taskStore.StopTask(&taskToStop)
			if !started {
				errorsList = append(errorsList, errors.StopTaskError)
			}
		} else {
			errorsList = append(errorsList, errors.GetTaskError+err.Error())
		}
	} else {
		errorsList = append(errorsList, errors.ParseObjectIDError+err.Error())
	}

	result := &responses.TaskResponse{Success: true, Data: []entities.Task{taskToStop}, Errors: make([]string, 0)}
	if len(errorsList) > 0 {
		response.WriteHeader(http.StatusBadRequest)
		result = &responses.TaskResponse{Success: false, Data: make([]entities.Task, 0), Errors: errorsList}
	}
	json.NewEncoder(response).Encode(result)
}

// StartTestTaskEndpoint handles the POST request at /api/testtask/{ID}/start
func StartTestTaskEndpoint(response http.ResponseWriter, request *http.Request) {
	response.Header().Set("content-type", "application/json")
	response.Header().Set("Access-Control-Allow-Origin", "http://localhost:3000")
	taskToStart := entities.Task{}
	errorsList := make([]string, 0)

	params := mux.Vars(request)
	ID, err := primitive.ObjectIDFromHex(params["ID"])
	if err == nil {
		taskToStart, err = queries.GetTestTask(ID)
		if err == nil {
			taskStore := stores.GetTaskStore()
			started := taskStore.StartTask(&taskToStart)
			if !started {
				errorsList = append(errorsList, errors.StartTaskError)
			}
		} else {
			errorsList = append(errorsList, errors.GetTaskError+err.Error())
		}
	} else {
		errorsList = append(errorsList, errors.ParseObjectIDError+err.Error())
	}
	result := &responses.TaskResponse{Success: true, Data: []entities.Task{taskToStart}, Errors: make([]string, 0)}
	if len(errorsList) > 0 {
		response.WriteHeader(http.StatusBadRequest)
		result = &responses.TaskResponse{Success: false, Data: make([]entities.Task, 0), Errors: errorsList}
	}
	json.NewEncoder(response).Encode(result)
}

// StopTestTaskEndpoint handles the POST request at /api/testtask/{ID}/stop
func StopTestTaskEndpoint(response http.ResponseWriter, request *http.Request) {
	response.Header().Set("content-type", "application/json")
	response.Header().Set("Access-Control-Allow-Origin", "http://localhost:3000")
	taskToStop := entities.Task{}
	errorsList := make([]string, 0)

	params := mux.Vars(request)
	ID, err := primitive.ObjectIDFromHex(params["ID"])
	if err == nil {
		taskToStop, err = queries.GetTestTask(ID)
		if err == nil {
			taskStore := stores.GetTaskStore()
			started := taskStore.StopTask(&taskToStop)
			if !started {
				errorsList = append(errorsList, errors.StopTaskError)
			}
		} else {
			errorsList = append(errorsList, errors.GetTaskError+err.Error())
		}
	} else {
		errorsList = append(errorsList, errors.ParseObjectIDError+err.Error())
	}

	result := &responses.TaskResponse{Success: true, Data: []entities.Task{taskToStop}, Errors: make([]string, 0)}
	if len(errorsList) > 0 {
		response.WriteHeader(http.StatusBadRequest)
		result = &responses.TaskResponse{Success: false, Data: make([]entities.Task, 0), Errors: errorsList}
	}
	json.NewEncoder(response).Encode(result)
}
