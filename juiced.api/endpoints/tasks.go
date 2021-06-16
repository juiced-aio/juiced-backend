package endpoints

import (
	"strings"

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

	"github.com/google/uuid"
	"github.com/gorilla/mux"
)

// GetTaskGroupEndpoint handles the GET request at /api/task/group/{groupID}
func GetTaskGroupEndpoint(response http.ResponseWriter, request *http.Request) {
	response.Header().Set("content-type", "application/json")
	response.Header().Set("Access-Control-Allow-Origin", "http://localhost:3000")
	var taskGroup entities.TaskGroup
	var err error
	errorsList := make([]string, 0)

	params := mux.Vars(request)
	groupID, ok := params["GroupID"]
	if ok {
		taskGroup, err = queries.GetTaskGroup(groupID)
		if err != nil {
			errorsList = append(errorsList, errors.GetTaskGroupError+err.Error())
		}
	} else {
		errorsList = append(errorsList, errors.MissingParameterError)
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
	groupID := uuid.New().String()
	taskGroup := &entities.TaskGroup{GroupID: groupID, TaskIDs: []string{}, MonitorDelay: 2000}
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
	var taskGroup entities.TaskGroup
	var err error
	errorsList := make([]string, 0)

	params := mux.Vars(request)
	groupID, ok := params["GroupID"]
	if ok {
		taskGroup, err = commands.RemoveTaskGroup(groupID)
		if err != nil {
			errorsList = append(errorsList, errors.RemoveTaskGroupError+err.Error())
		}
	} else {
		errorsList = append(errorsList, errors.MissingParameterError)
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
	var newTaskGroup entities.TaskGroup
	errorsList := make([]string, 0)

	type UpdateTaskGroupRequest struct {
		Name                string `json:"name"`
		MonitorInput        string `json:"input"`
		MonitorDelay        int    `json:"delay"`
		MonitorProxyGroupID string `json:"proxyGroupId"`
	}

	params := mux.Vars(request)
	groupID, ok := params["GroupID"]
	if ok {
		taskGroup, err := queries.GetTaskGroup(groupID)
		if err == nil {
			body, err := ioutil.ReadAll(request.Body)
			if err == nil {
				updateTaskGroupRequestInfo := UpdateTaskGroupRequest{}
				err = json.Unmarshal(body, &updateTaskGroupRequestInfo)
				if err == nil {
					taskGroup.Name = updateTaskGroupRequestInfo.Name
					taskGroup.MonitorDelay = updateTaskGroupRequestInfo.MonitorDelay
					taskGroup.MonitorProxyGroupID = updateTaskGroupRequestInfo.MonitorProxyGroupID
					switch taskGroup.MonitorRetailer {
					case enums.BestBuy:
						maxPrice := -1
						if len(taskGroup.BestbuyMonitorInfo.Monitors) > 0 {
							maxPrice = taskGroup.BestbuyMonitorInfo.Monitors[0].MaxPrice
						}

						newMonitors := make([]entities.BestbuySingleMonitorInfo, 0)
						if updateTaskGroupRequestInfo.MonitorInput != "" {
							skus := strings.Split(updateTaskGroupRequestInfo.MonitorInput, ",")
							for _, sku := range skus {
								monitor := entities.BestbuySingleMonitorInfo{
									MonitorID:   uuid.New().String(),
									TaskGroupID: taskGroup.GroupID,
									SKU:         sku,
									MaxPrice:    maxPrice,
								}
								newMonitors = append(newMonitors, monitor)
							}
						}
						taskGroup.BestbuyMonitorInfo.Monitors = newMonitors

					case enums.GameStop:
						maxPrice := -1
						if len(taskGroup.GamestopMonitorInfo.Monitors) > 0 {
							maxPrice = taskGroup.GamestopMonitorInfo.Monitors[0].MaxPrice
						}

						newMonitors := make([]entities.GamestopSingleMonitorInfo, 0)
						if updateTaskGroupRequestInfo.MonitorInput != "" {
							skus := strings.Split(updateTaskGroupRequestInfo.MonitorInput, ",")
							for _, sku := range skus {
								monitor := entities.GamestopSingleMonitorInfo{
									MonitorID:   uuid.New().String(),
									TaskGroupID: taskGroup.GroupID,
									SKU:         sku,
									MaxPrice:    maxPrice,
								}
								newMonitors = append(newMonitors, monitor)
							}
						}
						taskGroup.GamestopMonitorInfo.Monitors = newMonitors

					case enums.Target:
						maxPrice := -1
						if len(taskGroup.TargetMonitorInfo.Monitors) > 0 {
							maxPrice = taskGroup.TargetMonitorInfo.Monitors[0].MaxPrice
						}

						newMonitors := make([]entities.TargetSingleMonitorInfo, 0)
						if updateTaskGroupRequestInfo.MonitorInput != "" {
							tcins := strings.Split(updateTaskGroupRequestInfo.MonitorInput, ",")
							for _, tcin := range tcins {
								monitor := entities.TargetSingleMonitorInfo{
									MonitorID:   uuid.New().String(),
									TaskGroupID: taskGroup.GroupID,
									TCIN:        tcin,
									MaxPrice:    maxPrice,
								}
								newMonitors = append(newMonitors, monitor)
							}
						}
						taskGroup.TargetMonitorInfo.Monitors = newMonitors

					case enums.Walmart:
						taskGroup.WalmartMonitorInfo.SKUsJoined = updateTaskGroupRequestInfo.MonitorInput
						skus := make([]string, 0)
						if updateTaskGroupRequestInfo.MonitorInput != "" {
							skus = strings.Split(updateTaskGroupRequestInfo.MonitorInput, ",")
						}
						taskGroup.WalmartMonitorInfo.SKUs = skus
					}
					newTaskGroup, err = commands.UpdateTaskGroup(groupID, taskGroup)
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
		errorsList = append(errorsList, errors.MissingParameterError)
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
	var newTaskGroup entities.TaskGroup
	var err error
	errorsList := make([]string, 0)

	params := mux.Vars(request)
	groupID, ok := params["GroupID"]

	if ok {
		newTaskGroup, err = queries.GetTaskGroup(groupID)
		if err == nil {
			newTaskGroup.SetGroupID(uuid.New().String())
			newTaskGroup.SetName(newTaskGroup.Name + " (Copy " + common.RandID(4) + ")")
			newTaskIDs := make([]string, 0)
			for _, taskID := range newTaskGroup.TaskIDs {
				var task entities.Task
				task, err = queries.GetTask(taskID)
				if err != nil {
					break
				}
				task.ID = uuid.New().String()
				err = commands.CreateTask(task)
				if err != nil {
					break
				}
				newTaskIDs = append(newTaskIDs, task.ID)
			}
			if err == nil {
				newTaskGroup.TaskIDs = newTaskIDs
				err = commands.CreateTaskGroup(newTaskGroup)
				if err != nil {
					errorsList = append(errorsList, errors.CreateTaskGroupError+err.Error())
				}
			} else {
				errorsList = append(errorsList, errors.CreateTaskGroupError+err.Error())
			}
		} else {
			errorsList = append(errorsList, errors.GetTaskGroupError+err.Error())
		}
	} else {
		errorsList = append(errorsList, errors.MissingParameterError)
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
	var newTaskGroup entities.TaskGroup
	errorsList := make([]string, 0)

	type DeleteTasksRequest struct {
		TaskIDs []string `json:"taskIDs"`
	}

	params := mux.Vars(request)
	groupID, ok := params["GroupID"]
	if ok {
		body, err := ioutil.ReadAll(request.Body)
		if err == nil {
			newTaskGroup, err = queries.GetTaskGroup(groupID)
			if err == nil {
				deleteTasksRequestInfo := DeleteTasksRequest{}
				err = json.Unmarshal(body, &deleteTasksRequestInfo)
				if err == nil {
					newTaskIDs := make([]string, 0)
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
		errorsList = append(errorsList, errors.MissingParameterError)
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
	var task entities.Task
	var err error
	errorsList := make([]string, 0)

	params := mux.Vars(request)
	ID, ok := params["ID"]
	if ok {
		task, err = queries.GetTask(ID)
		if err != nil {
			errorsList = append(errorsList, errors.GetTaskError+err.Error())
		}
	} else {
		errorsList = append(errorsList, errors.MissingParameterError)
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
	task := &entities.Task{ID: uuid.New().String(), TaskSize: make([]string, 0), TaskQty: 1, TaskStatus: enums.TaskIdle}
	var newTaskGroup entities.TaskGroup
	errorsList := make([]string, 0)

	type CreateTaskRequest struct {
		NumTasksPerProfile int                       `json:"numTasksPerProfile"`
		ProfileIDs         []string                  `json:"profileIDs"`
		ProfileGroupID     string                    `json:"profileGroupID"`
		ProxyGroupID       string                    `json:"proxyGroupID"`
		Retailer           string                    `json:"retailer"`
		Sizes              []string                  `json:"sizes"`
		Quantity           int                       `json:"quantity"`
		Delay              int                       `json:"delay"`
		TargetTaskInfo     entities.TargetTaskInfo   `json:"targetTaskInfo"`
		WalmartTaskInfo    entities.WalmartTaskInfo  `json:"walmartTaskInfo"`
		AmazonTaskInfo     entities.AmazonTaskInfo   `json:"amazonTaskInfo"`
		BestbuyTaskInfo    entities.BestbuyTaskInfo  `json:"bestbuyTaskInfo"`
		GamestopTaskInfo   entities.GamestopTaskInfo `json:"gamestopTaskInfo"`
		HottopicTaskInfo   entities.HottopicTaskInfo `json:"hottopicTaskInfo"`
	}

	params := mux.Vars(request)
	groupID, ok := params["GroupID"]
	if ok {
		task.TaskGroupID = groupID
		body, err := ioutil.ReadAll(request.Body)
		if err == nil {
			createTaskRequestInfo := CreateTaskRequest{}
			err = json.Unmarshal(body, &createTaskRequestInfo)
			if err == nil {
				task.TaskProxyGroupID = createTaskRequestInfo.ProxyGroupID
				task.TaskRetailer = createTaskRequestInfo.Retailer
				task.TaskSize = createTaskRequestInfo.Sizes
				task.TaskSizeJoined = strings.Join(createTaskRequestInfo.Sizes, ",")
				if createTaskRequestInfo.Quantity > 0 {
					task.TaskQty = createTaskRequestInfo.Quantity
				}
				if createTaskRequestInfo.Delay > 0 {
					task.TaskDelay = createTaskRequestInfo.Delay
				}
				switch createTaskRequestInfo.Retailer {
				case enums.Amazon:
					task.AmazonTaskInfo = createTaskRequestInfo.AmazonTaskInfo

				case enums.BestBuy:
					task.BestbuyTaskInfo = createTaskRequestInfo.BestbuyTaskInfo

				case enums.GameStop:
					task.GamestopTaskInfo = createTaskRequestInfo.GamestopTaskInfo

				case enums.HotTopic:
					task.HottopicTaskInfo = createTaskRequestInfo.HottopicTaskInfo

				case enums.Target:
					task.TargetTaskInfo = createTaskRequestInfo.TargetTaskInfo

				case enums.Walmart:
					task.WalmartTaskInfo = createTaskRequestInfo.WalmartTaskInfo

				}

				profileIDs := createTaskRequestInfo.ProfileIDs
				if createTaskRequestInfo.ProfileGroupID != "" {
					var profileGroup entities.ProfileGroup
					profileGroup, err = queries.GetProfileGroup(createTaskRequestInfo.ProfileGroupID)
					profileIDs = profileGroup.ProfileIDs
				}

				if err == nil {
					oldTaskGroup, err := queries.GetTaskGroup(groupID)
					if err == nil {
						for i := 0; i < len(profileIDs); i++ {
							task.SetTaskProfileID(profileIDs[i])
							var createTaskError error
							for j := 0; j < createTaskRequestInfo.NumTasksPerProfile; j++ {
								task.SetID(uuid.New().String())
								err = commands.CreateTask(*task)
								if err != nil {
									createTaskError = err
									break
								}
								oldTaskGroup.SetTaskIDs(append(oldTaskGroup.TaskIDs, task.ID))
							}
							if createTaskError != nil {
								break
							}
						}

						if err == nil {
							newTaskGroup, err = commands.UpdateTaskGroup(groupID, oldTaskGroup)
							if err != nil {
								errorsList = append(errorsList, errors.AddTaskToGroupError+err.Error())
							}
						} else {
							errorsList = append(errorsList, errors.CreateTaskError+err.Error())
						}
					} else {
						errorsList = append(errorsList, errors.GetTaskGroupError+err.Error())
					}
				} else {
					errorsList = append(errorsList, errors.GetProfileGroupError+err.Error())
				}
			} else {
				errorsList = append(errorsList, errors.ParseTaskError+err.Error())
			}
		} else {
			errorsList = append(errorsList, errors.IOUtilReadAllError+err.Error())
		}
	} else {
		errorsList = append(errorsList, errors.MissingParameterError)
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
	var task entities.Task
	errorsList := make([]string, 0)

	params := mux.Vars(request)
	ID, ok := params["ID"]
	if ok {
		newTask := &entities.Task{ID: ID}

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
		errorsList = append(errorsList, errors.MissingParameterError)
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
	var task entities.Task
	var err error
	errorsList := make([]string, 0)

	params := mux.Vars(request)
	ID, ok := params["ID"]
	if ok {
		task, err = queries.GetTask(ID)
		if err == nil {
			task.SetID(uuid.New().String())
			err = commands.CreateTask(task)
			if err == nil {
				var taskGroup entities.TaskGroup
				taskGroup, err = queries.GetTaskGroup(task.TaskGroupID)
				taskGroup.TaskIDs = append(taskGroup.TaskIDs, task.ID)
				if err == nil {
					taskGroup, err = commands.UpdateTaskGroup(taskGroup.GroupID, taskGroup)
					if err != nil {
						errorsList = append(errorsList, errors.CreateTaskError+err.Error())
					}
				} else {
					errorsList = append(errorsList, errors.CreateTaskError+err.Error())
				}
			} else {
				errorsList = append(errorsList, errors.CreateTaskError+err.Error())
			}
		} else {
			errorsList = append(errorsList, errors.GetTaskError+err.Error())
		}
	} else {
		errorsList = append(errorsList, errors.MissingParameterError)
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
	var taskToStart entities.Task
	var err error
	errorsList := make([]string, 0)

	params := mux.Vars(request)
	ID, ok := params["ID"]
	if ok {
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
		errorsList = append(errorsList, errors.MissingParameterError)
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
	var taskToStop entities.Task
	var err error
	errorsList := make([]string, 0)

	params := mux.Vars(request)
	ID, ok := params["ID"]
	if ok {
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
		errorsList = append(errorsList, errors.MissingParameterError)
	}

	result := &responses.TaskResponse{Success: true, Data: []entities.Task{taskToStop}, Errors: make([]string, 0)}
	if len(errorsList) > 0 {
		response.WriteHeader(http.StatusBadRequest)
		result = &responses.TaskResponse{Success: false, Data: make([]entities.Task, 0), Errors: errorsList}
	}
	json.NewEncoder(response).Encode(result)
}
