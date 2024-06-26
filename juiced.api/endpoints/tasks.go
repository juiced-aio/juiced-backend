package endpoints

import (
	"log"
	"runtime/debug"
	"strings"
	"time"

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

// UpdateTaskStatuses updates the TaskGroupWithTask's Tasks with the correct statuses from the task store
func UpdateStatuses(taskGroupWithTasks entities.TaskGroupWithTasks) entities.TaskGroupWithTasks {
	taskStatuses := stores.GetTaskStatuses()
	newTasks := []entities.Task{}
	for _, task := range taskGroupWithTasks.Tasks {
		status := taskStatuses[task.ID]
		if status != "" {
			task.SetTaskStatus(status)
		}
		newTasks = append(newTasks, task)
	}
	taskGroupWithTasks.SetTasks(newTasks)

	monitorStatus := stores.GetMonitorStatus(taskGroupWithTasks.GroupID)
	if monitorStatus != "" {
		taskGroupWithTasks.MonitorStatus = monitorStatus
	}

	return taskGroupWithTasks
}

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
	data := []entities.TaskGroupWithTasks{UpdateStatuses(newTaskGroupWithTasks)}
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
		data = append(data, UpdateStatuses(newTaskGroupWithTasks))
	}
	result := &responses.TaskGroupResponse{Success: true, Data: data, Errors: make([]string, 0)}
	if len(errorsList) > 0 {
		response.WriteHeader(http.StatusBadRequest)
		result = &responses.TaskGroupResponse{Success: false, Data: data, Errors: errorsList}
	}
	err = json.NewEncoder(response).Encode(result)
	if err != nil {
		log.Println(err.Error())
	}
}

// CreateTaskGroupEndpoint handles the POST request at /api/task/group
func CreateTaskGroupEndpoint(response http.ResponseWriter, request *http.Request) {
	defer func() {
		if r := recover(); r != nil {
			log.Println(r)
			log.Println(string(debug.Stack()))
		}
	}()
	response.Header().Set("content-type", "application/json")
	response.Header().Set("Access-Control-Allow-Origin", "http://localhost:3000")
	groupID := uuid.New().String()
	taskGroup := &entities.TaskGroup{GroupID: groupID, TaskIDs: []string{}, MonitorDelay: 2000}
	errorsList := make([]string, 0)

	body, err := ioutil.ReadAll(request.Body)
	if err == nil {
		err = entities.ParseTaskGroup(taskGroup, body)
		if err == nil {
			taskGroup.CreationDate = time.Now().Unix()
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
	err = json.NewEncoder(response).Encode(result)
	if err != nil {
		log.Println(err.Error())
	}
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
		taskGroup, err = queries.GetTaskGroup(groupID)
		if err == nil {
			monitorStore := stores.GetMonitorStore()
			_, err = monitorStore.StopMonitor(&taskGroup)
			if err == nil {
				next := true
				for _, taskID := range taskGroup.TaskIDs {
					taskToStop, err := queries.GetTask(taskID)
					if err == nil {
						taskStore := stores.GetTaskStore()
						_, err = taskStore.StopTask(&taskToStop)
						if err != nil {
							next = false
							errorsList = append(errorsList, errors.StopTaskError+err.Error())
							break
						}
					} else {
						errorsList = append(errorsList, errors.GetTaskError+err.Error())
						break
					}
				}
				if next {
					taskGroup, err = commands.RemoveTaskGroup(groupID, true)
					if err != nil {
						errorsList = append(errorsList, errors.RemoveTaskGroupError+err.Error())
					}
				}
			} else {
				errorsList = append(errorsList, errors.StopMonitorError+err.Error())
			}
		} else {
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

	type AmazonUpdateInfo struct {
		MonitorType enums.MonitorType `json:"monitorType"`
	}
	type BestBuyUpdateInfo struct{}
	type BoxlunchUpdateInfo struct {
		Sizes  string `json:"sizes"`
		Colors string `json:"colors"`
	}
	type DisneyUpdateInfo struct {
		Sizes  string `json:"sizes"`
		Colors string `json:"colors"`
	}
	type GamestopUpdateInfo struct{}
	type NeweggUpdateInfo struct{}
	type PokemonCenterUpdateInfo struct{}
	type HottopicUpdateInfo struct {
		Sizes  string `json:"sizes"`
		Colors string `json:"colors"`
	}
	type ShopifyUpdateInfo struct{}

	type ToppsUpdateInfo struct{}
	type TargetUpdateInfo struct {
		CheckoutType enums.CheckoutType `json:"checkoutType"`
		StoreID      string             `json:"storeID"`
	}
	type WalmartUpdateInfo struct {
		SoldByWalmart bool `json:"soldByWalmart"`
	}

	type UpdateTaskGroupRequest struct {
		Name                    string                  `json:"name"`
		MonitorInput            string                  `json:"input"`
		MonitorDelay            int                     `json:"delay"`
		MonitorProxyGroupID     string                  `json:"proxyGroupId"`
		MaxPrice                int                     `json:"maxPrice"`
		AmazonUpdateInfo        AmazonUpdateInfo        `json:"amazonUpdateInfo"`
		BestbuyUpdateInfo       BestBuyUpdateInfo       `json:"bestbuyUpdateInfo"`
		BoxlunchUpdateInfo      BoxlunchUpdateInfo      `json:"boxlunchUpdateInfo"`
		DisneyUpdateInfo        DisneyUpdateInfo        `json:"disneyUpdateInfo"`
		GamestopUpdateInfo      GamestopUpdateInfo      `json:"gamestopUpdateInfo"`
		HottopicUpdateInfo      HottopicUpdateInfo      `json:"hottopicUpdateInfo"`
		NeweggUpdateInfo        NeweggUpdateInfo        `json:"neweggUpdateInfo"`
		PokemonCenterUpdateInfo PokemonCenterUpdateInfo `json:"pokemoncenterUpdateInfo"`
		ShopifyUpdateInfo       ShopifyUpdateInfo       `json:"shopifyUpdateInfo"`
		TargetUpdateInfo        TargetUpdateInfo        `json:"targetUpdateInfo"`
		ToppsUpdateInfo         ToppsUpdateInfo         `json:"toppsUpdateInfo"`
		WalmartUpdateInfo       WalmartUpdateInfo       `json:"walmartUpdateInfo"`
	}

	params := mux.Vars(request)
	groupID, ok := params["GroupID"]
	if ok {
		taskGroup, err := queries.GetTaskGroup(groupID)
		if err == nil {
			monitorStore := stores.GetMonitorStore()
			wasRunning, err := monitorStore.StopMonitor(&taskGroup)
			if err == nil {
				body, err := ioutil.ReadAll(request.Body)
				if err == nil {
					updateTaskGroupRequestInfo := UpdateTaskGroupRequest{}
					err = json.Unmarshal(body, &updateTaskGroupRequestInfo)
					if err == nil {
						taskGroup.Name = updateTaskGroupRequestInfo.Name
						taskGroup.MonitorDelay = updateTaskGroupRequestInfo.MonitorDelay
						taskGroup.MonitorProxyGroupID = updateTaskGroupRequestInfo.MonitorProxyGroupID
						maxPrice := updateTaskGroupRequestInfo.MaxPrice
						switch taskGroup.MonitorRetailer {
						case enums.Amazon:
							newMonitors := make([]entities.AmazonSingleMonitorInfo, 0)
							if updateTaskGroupRequestInfo.MonitorInput != "" {
								skus := strings.Split(updateTaskGroupRequestInfo.MonitorInput, ",")
								for _, sku := range skus {
									monitor := entities.AmazonSingleMonitorInfo{
										MonitorID:   uuid.New().String(),
										TaskGroupID: taskGroup.GroupID,
										MaxPrice:    maxPrice,
										MonitorType: updateTaskGroupRequestInfo.AmazonUpdateInfo.MonitorType,
									}
									switch updateTaskGroupRequestInfo.AmazonUpdateInfo.MonitorType {
									case enums.SlowSKUMonitor:
										monitor.ASIN = sku
									case enums.FastSKUMonitor:
										if !strings.Contains(sku, ":") {
											continue
										}
										monitor.ASIN = strings.Split(sku, ":")[0]
										monitor.OFID = strings.Split(sku, ":")[1]
									}
									newMonitors = append(newMonitors, monitor)
								}
								taskGroup.AmazonMonitorInfo.Monitors = newMonitors
							}

						case enums.BestBuy:
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
								taskGroup.BestbuyMonitorInfo.Monitors = newMonitors
							}

						case enums.BoxLunch:
							newMonitors := make([]entities.BoxlunchSingleMonitorInfo, 0)
							if updateTaskGroupRequestInfo.MonitorInput != "" {
								pids := strings.Split(updateTaskGroupRequestInfo.MonitorInput, ",")
								monitorType := enums.SKUMonitor
								if len(taskGroup.BoxlunchMonitorInfo.Monitors) > 0 {
									monitorType = taskGroup.BoxlunchMonitorInfo.Monitors[0].MonitorType
								}
								for _, pid := range pids {
									monitor := entities.BoxlunchSingleMonitorInfo{
										MonitorID:   uuid.New().String(),
										TaskGroupID: taskGroup.GroupID,
										Pid:         pid,
										Size:        updateTaskGroupRequestInfo.BoxlunchUpdateInfo.Sizes,
										Color:       updateTaskGroupRequestInfo.BoxlunchUpdateInfo.Colors,
										MaxPrice:    maxPrice,
										MonitorType: monitorType,
									}
									newMonitors = append(newMonitors, monitor)
								}
								taskGroup.BoxlunchMonitorInfo.Monitors = newMonitors
							}

						case enums.Disney:
							newMonitors := make([]entities.DisneySingleMonitorInfo, 0)
							if updateTaskGroupRequestInfo.MonitorInput != "" {
								pids := strings.Split(updateTaskGroupRequestInfo.MonitorInput, ",")
								for _, pid := range pids {
									monitor := entities.DisneySingleMonitorInfo{
										MonitorID:   uuid.New().String(),
										TaskGroupID: taskGroup.GroupID,
										PID:         pid,
										Size:        updateTaskGroupRequestInfo.DisneyUpdateInfo.Sizes,
										Color:       updateTaskGroupRequestInfo.DisneyUpdateInfo.Colors,
										MaxPrice:    maxPrice,
									}
									newMonitors = append(newMonitors, monitor)
								}
								taskGroup.DisneyMonitorInfo.Monitors = newMonitors
							}

						case enums.GameStop:
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
								taskGroup.GamestopMonitorInfo.Monitors = newMonitors
							}

						case enums.HotTopic:
							newMonitors := make([]entities.HottopicSingleMonitorInfo, 0)
							if updateTaskGroupRequestInfo.MonitorInput != "" {
								pids := strings.Split(updateTaskGroupRequestInfo.MonitorInput, ",")
								monitorType := enums.SKUMonitor
								if len(taskGroup.HottopicMonitorInfo.Monitors) > 0 {
									monitorType = taskGroup.HottopicMonitorInfo.Monitors[0].MonitorType
								}
								for _, pid := range pids {
									monitor := entities.HottopicSingleMonitorInfo{
										MonitorID:   uuid.New().String(),
										TaskGroupID: taskGroup.GroupID,
										Pid:         pid,
										Size:        updateTaskGroupRequestInfo.HottopicUpdateInfo.Sizes,
										Color:       updateTaskGroupRequestInfo.HottopicUpdateInfo.Colors,
										MaxPrice:    maxPrice,
										MonitorType: monitorType,
									}
									newMonitors = append(newMonitors, monitor)
								}
								taskGroup.HottopicMonitorInfo.Monitors = newMonitors
							}

						case enums.Newegg:
							newMonitors := make([]entities.NeweggSingleMonitorInfo, 0)
							if updateTaskGroupRequestInfo.MonitorInput != "" {
								skus := strings.Split(updateTaskGroupRequestInfo.MonitorInput, ",")
								for _, sku := range skus {
									monitor := entities.NeweggSingleMonitorInfo{
										MonitorID:   uuid.New().String(),
										TaskGroupID: taskGroup.GroupID,
										SKU:         sku,
										MaxPrice:    maxPrice,
									}
									newMonitors = append(newMonitors, monitor)
								}
								taskGroup.NeweggMonitorInfo.Monitors = newMonitors
							}

						case enums.PokemonCenter:
							newMonitors := make([]entities.PokemonCenterSingleMonitorInfo, 0)
							if updateTaskGroupRequestInfo.MonitorInput != "" {
								skus := strings.Split(updateTaskGroupRequestInfo.MonitorInput, ",")
								for _, sku := range skus {
									monitor := entities.PokemonCenterSingleMonitorInfo{
										MonitorID:   uuid.New().String(),
										TaskGroupID: taskGroup.GroupID,
										SKU:         sku,
										MaxPrice:    maxPrice,
									}
									newMonitors = append(newMonitors, monitor)
								}
							}
							taskGroup.PokemonCenterMonitorInfo.Monitors = newMonitors

						case enums.Shopify:
							newMonitors := make([]entities.ShopifySingleMonitorInfo, 0)
							if updateTaskGroupRequestInfo.MonitorInput != "" {
								vids := strings.Split(updateTaskGroupRequestInfo.MonitorInput, ",")
								for _, vid := range vids {
									monitor := entities.ShopifySingleMonitorInfo{
										MonitorID:   uuid.New().String(),
										TaskGroupID: taskGroup.GroupID,
										VariantID:   vid,
										MaxPrice:    maxPrice,
									}
									newMonitors = append(newMonitors, monitor)
								}
								taskGroup.ShopifyMonitorInfo.Monitors = newMonitors
							}

						case enums.Target:
							newMonitors := make([]entities.TargetSingleMonitorInfo, 0)
							if updateTaskGroupRequestInfo.MonitorInput != "" {
								tcins := strings.Split(updateTaskGroupRequestInfo.MonitorInput, ",")
								for _, tcin := range tcins {
									monitor := entities.TargetSingleMonitorInfo{
										MonitorID:    uuid.New().String(),
										TaskGroupID:  taskGroup.GroupID,
										TCIN:         tcin,
										MaxPrice:     maxPrice,
										CheckoutType: updateTaskGroupRequestInfo.TargetUpdateInfo.CheckoutType,
									}
									newMonitors = append(newMonitors, monitor)
								}
								taskGroup.TargetMonitorInfo.StoreID = updateTaskGroupRequestInfo.TargetUpdateInfo.StoreID
								taskGroup.TargetMonitorInfo.Monitors = newMonitors
							}

						case enums.Topps:
							newMonitors := make([]entities.ToppsSingleMonitorInfo, 0)
							if updateTaskGroupRequestInfo.MonitorInput != "" {
								items := strings.Split(updateTaskGroupRequestInfo.MonitorInput, ",")
								for _, item := range items {
									monitor := entities.ToppsSingleMonitorInfo{
										MonitorID:   uuid.New().String(),
										TaskGroupID: taskGroup.GroupID,
										Item:        item,
										MaxPrice:    maxPrice,
									}
									newMonitors = append(newMonitors, monitor)
								}
								taskGroup.ToppsMonitorInfo.Monitors = newMonitors
							}

						case enums.Walmart:
							newMonitors := make([]entities.WalmartSingleMonitorInfo, 0)
							if updateTaskGroupRequestInfo.MonitorInput != "" {
								ids := strings.Split(updateTaskGroupRequestInfo.MonitorInput, ",")
								monitorType := enums.SKUMonitor
								if len(taskGroup.WalmartMonitorInfo.Monitors) > 0 {
									monitorType = taskGroup.WalmartMonitorInfo.Monitors[0].MonitorType
								}
								for _, id := range ids {
									monitor := entities.WalmartSingleMonitorInfo{
										MonitorID:     uuid.New().String(),
										TaskGroupID:   taskGroup.GroupID,
										ID:            id,
										MaxPrice:      maxPrice,
										SoldByWalmart: updateTaskGroupRequestInfo.WalmartUpdateInfo.SoldByWalmart,
										MonitorType:   monitorType,
									}
									newMonitors = append(newMonitors, monitor)
								}
								taskGroup.WalmartMonitorInfo.Monitors = newMonitors
							}

						}

						newTaskGroup, err = commands.UpdateTaskGroup(groupID, taskGroup)
						if err == nil {
							newTaskGroup.UpdateMonitor = true
							if wasRunning {
								err = monitorStore.StartMonitor(&newTaskGroup)
								if err != nil {
									errorsList = append(errorsList, errors.StartTaskGroupError+err.Error())
								}
							} else {
								if monitorStore.GetMonitor(newTaskGroup.MonitorRetailer, newTaskGroup.GroupID) != nil {
									err = monitorStore.UpdateMonitor(&newTaskGroup)
									if err != nil {
										errorsList = append(errorsList, errors.UpdateTaskGroupError+err.Error())
									}
								}
							}
						} else {
							errorsList = append(errorsList, errors.UpdateTaskGroupError+err.Error())
						}
					} else {
						errorsList = append(errorsList, errors.ParseUpdateTaskGroupRequestError+err.Error())
					}
				} else {
					errorsList = append(errorsList, errors.IOUtilReadAllError+err.Error())
				}
			} else {
				errorsList = append(errorsList, errors.StopMonitorError+err.Error())
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
	data := []entities.TaskGroupWithTasks{UpdateStatuses(newTaskGroupWithTasks)}
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
			newTaskGroup.CreationDate = time.Now().Unix()
			newTaskIDs := make([]string, 0)
			for _, taskID := range newTaskGroup.TaskIDs {
				var task entities.Task
				task, err = queries.GetTask(taskID)
				if err != nil {
					break
				}
				task.ID = uuid.New().String()
				task.TaskGroupID = newTaskGroup.GroupID
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

// StartTaskGroupEndpoint handles the POST request at /api/task/group/{GroupID}/start
func StartTaskGroupEndpoint(response http.ResponseWriter, request *http.Request) {
	response.Header().Set("content-type", "application/json")
	response.Header().Set("Access-Control-Allow-Origin", "http://localhost:3000")
	var taskGroupToStart entities.TaskGroup
	var err error
	errorsList := make([]string, 0)
	warningsList := make([]string, 0)

	params := mux.Vars(request)
	groupID, ok := params["GroupID"]
	if ok {
		taskGroupToStart, err = queries.GetTaskGroup(groupID)
		if err == nil {
			taskStore := stores.GetTaskStore()
			warningsList, err = taskStore.StartTaskGroup(&taskGroupToStart)
			if err != nil {
				errorsList = append(errorsList, errors.StartTaskGroupError+err.Error())
			}
		} else {
			errorsList = append(errorsList, errors.GetTaskError+err.Error())
		}
	} else {
		errorsList = append(errorsList, errors.MissingParameterError)
	}

	taskGroupToStartWithTasks, err := queries.ConvertTaskIDsToTasks(&taskGroupToStart)
	if err != nil {
		errorsList = append(errorsList, errors.GetTaskError+err.Error())
	}

	result := &responses.TaskGroupResponse{Success: true, Data: []entities.TaskGroupWithTasks{UpdateStatuses(taskGroupToStartWithTasks)}, Errors: make([]string, 0), Warnings: warningsList}
	if len(errorsList) > 0 {
		response.WriteHeader(http.StatusBadRequest)
		result = &responses.TaskGroupResponse{Success: false, Data: make([]entities.TaskGroupWithTasks, 0), Errors: errorsList}
	}
	json.NewEncoder(response).Encode(result)
}

// StopTaskGroupEndpoint handles the POST request at /api/task/group/{GroupID}/stop
func StopTaskGroupEndpoint(response http.ResponseWriter, request *http.Request) {
	response.Header().Set("content-type", "application/json")
	response.Header().Set("Access-Control-Allow-Origin", "http://localhost:3000")
	var taskGroupToStop entities.TaskGroup
	var err error
	errorsList := make([]string, 0)

	params := mux.Vars(request)
	groupID, ok := params["GroupID"]
	if ok {
		taskGroupToStop, err = queries.GetTaskGroup(groupID)
		if err == nil {
			taskStore := stores.GetTaskStore()
			err = taskStore.StopTaskGroup(&taskGroupToStop)
			if err != nil {
				errorsList = append(errorsList, errors.StopTaskError+err.Error())
			}
		} else {
			errorsList = append(errorsList, errors.GetTaskError+err.Error())
		}
	} else {
		errorsList = append(errorsList, errors.MissingParameterError)
	}

	taskGroupToStopWithTasks, err := queries.ConvertTaskIDsToTasks(&taskGroupToStop)
	if err != nil {
		errorsList = append(errorsList, errors.GetTaskError+err.Error())
	}

	result := &responses.TaskGroupResponse{Success: true, Data: []entities.TaskGroupWithTasks{UpdateStatuses(taskGroupToStopWithTasks)}, Errors: make([]string, 0)}
	if len(errorsList) > 0 {
		response.WriteHeader(http.StatusBadRequest)
		result = &responses.TaskGroupResponse{Success: false, Data: make([]entities.TaskGroupWithTasks, 0), Errors: errorsList}
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
						taskStore := stores.GetTaskStore()
						for i := 0; i < len(deleteTasksRequestInfo.TaskIDs); i++ {
							task, err := queries.GetTask(deleteTasksRequestInfo.TaskIDs[i])
							if err == nil {
								taskStore.StopTask(&task)
							} else {
								errorsList = append(errorsList, errors.RemoveTaskError+err.Error())
							}
						}
						if !taskStore.TasksRunning(newTaskGroup.TaskIDs, newTaskGroup.MonitorRetailer) {
							monitorStore := stores.GetMonitorStore()
							_, err = monitorStore.StopMonitor(&newTaskGroup)
							if err != nil {
								errorsList = append(errorsList, errors.StopMonitorError+err.Error())
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
	data := []entities.TaskGroupWithTasks{UpdateStatuses(newTaskGroupWithTasks)}
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
		NumTasksPerProfile    int                             `json:"numTasksPerProfile"`
		ProfileIDs            []string                        `json:"profileIDs"`
		ProfileGroupID        string                          `json:"profileGroupID"`
		ProxyGroupID          string                          `json:"proxyGroupID"`
		Retailer              string                          `json:"retailer"`
		Sizes                 []string                        `json:"sizes"`
		Quantity              int                             `json:"quantity"`
		Delay                 int                             `json:"delay"`
		AmazonTaskInfo        *entities.AmazonTaskInfo        `json:"amazonTaskInfo"`
		BestbuyTaskInfo       *entities.BestbuyTaskInfo       `json:"bestbuyTaskInfo"`
		BoxlunchTaskInfo      *entities.BoxlunchTaskInfo      `json:"boxlunchTaskInfo"`
		DisneyTaskInfo        *entities.DisneyTaskInfo        `json:"disneyTaskInfo"`
		GamestopTaskInfo      *entities.GamestopTaskInfo      `json:"gamestopTaskInfo"`
		HottopicTaskInfo      *entities.HottopicTaskInfo      `json:"hottopicTaskInfo"`
		NeweggTaskInfo        *entities.NeweggTaskInfo        `json:"neweggTaskInfo"`
		PokemonCenterTaskInfo *entities.PokemonCenterTaskInfo `json:"pokemoncenterTaskInfo"`
		ShopifyTaskInfo       *entities.ShopifyTaskInfo       `json:"shopifyTaskInfo"`
		TargetTaskInfo        *entities.TargetTaskInfo        `json:"targetTaskInfo"`
		ToppsTaskInfo         *entities.ToppsTaskInfo         `json:"toppsTaskInfo"`
		WalmartTaskInfo       *entities.WalmartTaskInfo       `json:"walmartTaskInfo"`
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

				case enums.BoxLunch:
					task.BoxlunchTaskInfo = createTaskRequestInfo.BoxlunchTaskInfo

				case enums.Disney:
					task.DisneyTaskInfo = createTaskRequestInfo.DisneyTaskInfo

				case enums.GameStop:
					task.GamestopTaskInfo = createTaskRequestInfo.GamestopTaskInfo

				case enums.HotTopic:
					task.HottopicTaskInfo = createTaskRequestInfo.HottopicTaskInfo

				case enums.Newegg:
					task.NeweggTaskInfo = createTaskRequestInfo.NeweggTaskInfo

				case enums.Shopify:
					task.ShopifyTaskInfo = createTaskRequestInfo.ShopifyTaskInfo
					taskGroup, err := queries.GetTaskGroup(task.TaskGroupID)
					if err == nil {
						task.ShopifyTaskInfo.SitePassword = taskGroup.ShopifyMonitorInfo.SitePassword
						task.ShopifyTaskInfo.SiteURL = taskGroup.ShopifyMonitorInfo.SiteURL
					} else {
						errorsList = append(errorsList, errors.GetTaskGroupError+err.Error())
					}

				case enums.Target:
					task.TargetTaskInfo = createTaskRequestInfo.TargetTaskInfo

				case enums.Topps:
					task.ToppsTaskInfo = createTaskRequestInfo.ToppsTaskInfo

				case enums.Walmart:
					task.WalmartTaskInfo = createTaskRequestInfo.WalmartTaskInfo

				case enums.PokemonCenter:
					task.PokemonCenterTaskInfo = createTaskRequestInfo.PokemonCenterTaskInfo
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
						if createTaskRequestInfo.Retailer == enums.Amazon {
							for j := 0; j < createTaskRequestInfo.NumTasksPerProfile; j++ {
								task.SetID(uuid.New().String())
								task.CreationDate = time.Now().Unix()
								err = commands.CreateTask(*task)
								oldTaskGroup.SetTaskIDs(append(oldTaskGroup.TaskIDs, task.ID))
							}
						} else {
							for i := 0; i < len(profileIDs); i++ {
								profile, err := queries.GetProfile(profileIDs[i])
								if profile.ID != "" && err == nil {
									task.SetTaskProfileID(profileIDs[i])
									var createTaskError error
									for j := 0; j < createTaskRequestInfo.NumTasksPerProfile; j++ {
										task.SetID(uuid.New().String())
										task.CreationDate = time.Now().Unix()
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
	data := []entities.TaskGroupWithTasks{UpdateStatuses(newTaskGroupWithTasks)}
	result := &responses.TaskGroupResponse{Success: true, Data: data, Errors: make([]string, 0)}
	if len(errorsList) > 0 {
		response.WriteHeader(http.StatusBadRequest)
		result = &responses.TaskGroupResponse{Success: false, Data: data, Errors: errorsList}
	}
	json.NewEncoder(response).Encode(result)
}

// UpdateTasksEndpoint handles the PUT request at /api/task/group/{groupID}/updateTasks
func UpdateTasksEndpoint(response http.ResponseWriter, request *http.Request) {
	response.Header().Set("content-type", "application/json")
	response.Header().Set("Access-Control-Allow-Origin", "http://localhost:3000")
	var taskGroup entities.TaskGroup
	var err error
	errorsList := make([]string, 0)

	type UpdateTasksRequest struct {
		TaskIDs               []string                       `json:"taskIDs"`
		ProfileID             string                         `json:"profileID"`
		ProxyGroupID          string                         `json:"proxyGroupID"`
		Quantity              int                            `json:"quantity"`
		AmazonTaskInfo        entities.AmazonTaskInfo        `json:"amazonTaskInfo"`
		BestbuyTaskInfo       entities.BestbuyTaskInfo       `json:"bestbuyTaskInfo"`
		BoxlunchTaskInfo      entities.BoxlunchTaskInfo      `json:"boxlunchTaskInfo"`
		DisneyTaskInfo        entities.DisneyTaskInfo        `json:"disneyTaskInfo"`
		GamestopTaskInfo      entities.GamestopTaskInfo      `json:"gamestopTaskInfo"`
		HottopicTaskInfo      entities.HottopicTaskInfo      `json:"hottopicTaskInfo"`
		NeweggTaskInfo        entities.NeweggTaskInfo        `json:"neweggTaskInfo"`
		PokemonCenterTaskInfo entities.PokemonCenterTaskInfo `json:"pokemoncenterTaskInfo"`
		ShopifyTaskInfo       entities.ShopifyTaskInfo       `json:"shopifyTaskInfo"`
		TargetTaskInfo        entities.TargetTaskInfo        `json:"targetTaskInfo"`
		ToppsTaskInfo         entities.ToppsTaskInfo         `json:"toppsTaskInfo"`
		WalmartTaskInfo       entities.WalmartTaskInfo       `json:"walmartTaskInfo"`
	}

	params := mux.Vars(request)
	groupID, ok := params["GroupID"]
	if ok {
		taskGroup, err = queries.GetTaskGroup(groupID)
		if err == nil {
			body, err := ioutil.ReadAll(request.Body)
			if err == nil {
				updateTasksRequestInfo := UpdateTasksRequest{}
				err = json.Unmarshal(body, &updateTasksRequestInfo)
				if err == nil {
					singleTask := len(updateTasksRequestInfo.TaskIDs) == 1
					for _, taskID := range updateTasksRequestInfo.TaskIDs {
						task, err := queries.GetTask(taskID)
						if err == nil {
							taskStore := stores.GetTaskStore()
							var wasRunning bool
							wasRunning, err = taskStore.StopTask(&task)
							if err == nil {
								err = taskStore.SetDontPublishEvents(task.TaskRetailer, task.ID, true)
								if err == nil {
									if updateTasksRequestInfo.ProfileID != "DO_NOT_UPDATE" {
										task.TaskProfileID = updateTasksRequestInfo.ProfileID
									}
									if updateTasksRequestInfo.ProxyGroupID != "DO_NOT_UPDATE" {
										task.TaskProxyGroupID = updateTasksRequestInfo.ProxyGroupID
									}
									if updateTasksRequestInfo.Quantity != -1 && updateTasksRequestInfo.Quantity > 0 {
										task.TaskQty = updateTasksRequestInfo.Quantity
									}
									switch taskGroup.MonitorRetailer {
									case enums.Amazon:
										if singleTask || updateTasksRequestInfo.AmazonTaskInfo.Email != "" {
											task.AmazonTaskInfo.Email = updateTasksRequestInfo.AmazonTaskInfo.Email
										}
										if singleTask || updateTasksRequestInfo.AmazonTaskInfo.Password != "" {
											task.AmazonTaskInfo.Password = updateTasksRequestInfo.AmazonTaskInfo.Password
										}

									case enums.BestBuy:
										if updateTasksRequestInfo.BestbuyTaskInfo.TaskType != "DO_NOT_UPDATE" {
											task.BestbuyTaskInfo.TaskType = updateTasksRequestInfo.BestbuyTaskInfo.TaskType
										}
										if singleTask || updateTasksRequestInfo.BestbuyTaskInfo.Email != "" {
											task.BestbuyTaskInfo.Email = updateTasksRequestInfo.BestbuyTaskInfo.Email
										}
										if singleTask || updateTasksRequestInfo.BestbuyTaskInfo.Password != "" {
											task.BestbuyTaskInfo.Password = updateTasksRequestInfo.BestbuyTaskInfo.Password
										}
									case enums.BoxLunch:

									case enums.Disney:
										if updateTasksRequestInfo.DisneyTaskInfo.TaskType != "DO_NOT_UPDATE" {
											task.DisneyTaskInfo.TaskType = updateTasksRequestInfo.DisneyTaskInfo.TaskType
										}
										if singleTask || updateTasksRequestInfo.BestbuyTaskInfo.Email != "" {
											task.DisneyTaskInfo.Email = updateTasksRequestInfo.DisneyTaskInfo.Email
										}
										if singleTask || updateTasksRequestInfo.BestbuyTaskInfo.Password != "" {
											task.DisneyTaskInfo.Password = updateTasksRequestInfo.DisneyTaskInfo.Password
										}

									case enums.GameStop:
										if updateTasksRequestInfo.GamestopTaskInfo.TaskType != "DO_NOT_UPDATE" {
											task.GamestopTaskInfo.TaskType = updateTasksRequestInfo.GamestopTaskInfo.TaskType
										}
										if singleTask || updateTasksRequestInfo.GamestopTaskInfo.Email != "" {
											task.GamestopTaskInfo.Email = updateTasksRequestInfo.GamestopTaskInfo.Email
										}
										if singleTask || updateTasksRequestInfo.GamestopTaskInfo.Password != "" {
											task.GamestopTaskInfo.Password = updateTasksRequestInfo.GamestopTaskInfo.Password
										}

									case enums.HotTopic:

									case enums.Newegg:

									case enums.PokemonCenter:
										if updateTasksRequestInfo.PokemonCenterTaskInfo.TaskType != "DO_NOT_UPDATE" {
											task.PokemonCenterTaskInfo.TaskType = updateTasksRequestInfo.PokemonCenterTaskInfo.TaskType
										}
										if updateTasksRequestInfo.PokemonCenterTaskInfo.AddressType != "DO_NOT_UPDATE" {
											task.PokemonCenterTaskInfo.AddressType = updateTasksRequestInfo.PokemonCenterTaskInfo.AddressType
										}
										if singleTask || updateTasksRequestInfo.PokemonCenterTaskInfo.Email != "" {
											task.PokemonCenterTaskInfo.Email = updateTasksRequestInfo.PokemonCenterTaskInfo.Email
										}
										if singleTask || updateTasksRequestInfo.PokemonCenterTaskInfo.Password != "" {
											task.PokemonCenterTaskInfo.Password = updateTasksRequestInfo.PokemonCenterTaskInfo.Password
										}

									case enums.Shopify:
										if updateTasksRequestInfo.ShopifyTaskInfo.CouponCode != "DO_NOT_UPDATE" {
											task.ShopifyTaskInfo.CouponCode = updateTasksRequestInfo.ShopifyTaskInfo.CouponCode
										}
										if singleTask || updateTasksRequestInfo.ShopifyTaskInfo.HotWheelsTaskInfo.Email != "" {
											task.ShopifyTaskInfo.HotWheelsTaskInfo.Email = updateTasksRequestInfo.ShopifyTaskInfo.HotWheelsTaskInfo.Email
										}
										if singleTask || updateTasksRequestInfo.ShopifyTaskInfo.HotWheelsTaskInfo.Password != "" {
											task.ShopifyTaskInfo.HotWheelsTaskInfo.Password = updateTasksRequestInfo.ShopifyTaskInfo.HotWheelsTaskInfo.Password
										}

									case enums.Target:
										if updateTasksRequestInfo.TargetTaskInfo.CheckoutType != "DO_NOT_UPDATE" {
											task.TargetTaskInfo.CheckoutType = updateTasksRequestInfo.TargetTaskInfo.CheckoutType
										}
										if updateTasksRequestInfo.TargetTaskInfo.PaymentType != "DO_NOT_UPDATE" {
											task.TargetTaskInfo.PaymentType = updateTasksRequestInfo.TargetTaskInfo.PaymentType
										}
										if singleTask || updateTasksRequestInfo.TargetTaskInfo.Email != "" {
											task.TargetTaskInfo.Email = updateTasksRequestInfo.TargetTaskInfo.Email
										}
										if singleTask || updateTasksRequestInfo.TargetTaskInfo.Password != "" {
											task.TargetTaskInfo.Password = updateTasksRequestInfo.TargetTaskInfo.Password
										}

									case enums.Topps:
										if updateTasksRequestInfo.ToppsTaskInfo.TaskType != "DO_NOT_UPDATE" {
											task.ToppsTaskInfo.TaskType = updateTasksRequestInfo.ToppsTaskInfo.TaskType
										}
										if singleTask || updateTasksRequestInfo.ToppsTaskInfo.Email != "" {
											task.ToppsTaskInfo.Email = updateTasksRequestInfo.ToppsTaskInfo.Email
										}
										if singleTask || updateTasksRequestInfo.ToppsTaskInfo.Password != "" {
											task.ToppsTaskInfo.Password = updateTasksRequestInfo.ToppsTaskInfo.Password
										}

									case enums.Walmart:

									}
									_, err := commands.UpdateTask(taskID, task)
									if err == nil {
										task.UpdateTask = true
										if wasRunning {
											err = taskStore.StartTask(&task)
											if err != nil {
												errorsList = append(errorsList, errors.StartTaskError+err.Error())
											}
										} else {
											if taskStore.GetTask(task.TaskRetailer, task.ID) != nil {
												err = taskStore.UpdateTask(&task)
												if err != nil {
													errorsList = append(errorsList, errors.UpdateTaskError+err.Error())
												}
											}
										}
									} else {
										errorsList = append(errorsList, errors.UpdateTaskError+err.Error())
									}
								} else {
									errorsList = append(errorsList, errors.StopTaskError+err.Error())
								}
							} else {
								errorsList = append(errorsList, errors.StopTaskError+err.Error())
							}
						} else {
							errorsList = append(errorsList, errors.GetTaskError+err.Error())
						}
					}
				} else {
					errorsList = append(errorsList, errors.ParseUpdateTasksRequestError+err.Error())
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

	taskGroupWithTasks, err := queries.ConvertTaskIDsToTasks(&taskGroup)
	if err != nil {
		errorsList = append(errorsList, errors.GetTaskError+err.Error())
	}
	data := []entities.TaskGroupWithTasks{UpdateStatuses(taskGroupWithTasks)}

	result := &responses.TaskGroupResponse{Success: true, Data: data, Errors: make([]string, 0)}
	if len(errorsList) > 0 {
		response.WriteHeader(http.StatusBadRequest)
		result = &responses.TaskGroupResponse{Success: false, Data: data, Errors: errorsList}
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
			task.CreationDate = time.Now().Unix()
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
			err = taskStore.StartTask(&taskToStart)
			if err != nil {
				errorsList = append(errorsList, errors.StartTaskError+err.Error())
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
	var taskGroup entities.TaskGroup
	var err error
	errorsList := make([]string, 0)

	params := mux.Vars(request)
	ID, ok := params["ID"]
	if ok {
		taskToStop, err = queries.GetTask(ID)
		if err == nil {
			taskStore := stores.GetTaskStore()
			_, err = taskStore.StopTask(&taskToStop)
			if err == nil {
				taskGroup, err = queries.GetTaskGroup(taskToStop.TaskGroupID)
				if err == nil {
					if !taskStore.TasksRunning(taskGroup.TaskIDs, taskGroup.MonitorRetailer) {
						monitorStore := stores.GetMonitorStore()
						_, err = monitorStore.StopMonitor(&taskGroup)
						if err != nil {
							errorsList = append(errorsList, errors.StopMonitorError+err.Error())
						}
					}
				} else {
					errorsList = append(errorsList, errors.GetTaskGroupError+err.Error())
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

	result := &responses.TaskResponse{Success: true, Data: []entities.Task{taskToStop}, Errors: make([]string, 0)}
	if len(errorsList) > 0 {
		response.WriteHeader(http.StatusBadRequest)
		result = &responses.TaskResponse{Success: false, Data: make([]entities.Task, 0), Errors: errorsList}
	}
	json.NewEncoder(response).Encode(result)
}
