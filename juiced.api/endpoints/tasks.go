package endpoints

// // UpdateTasksEndpoint handles the PUT request at /api/task/group/{groupID}/updateTasks
// func UpdateTasksEndpoint(response http.ResponseWriter, request *http.Request) {
// 	response.Header().Set("content-type", "application/json")
// 	response.Header().Set("Access-Control-Allow-Origin", "http://localhost:3000")
// 	var taskGroup entities.TaskGroup
// 	var err error
// 	errorsList := make([]string, 0)

// 	// Let's see if we can abstract this
// 	type UpdateTasksRequest struct {
// 		TaskIDs               []string                       `json:"taskIDs"`
// 		ProfileID             string                         `json:"profileID"`
// 		ProxyGroupID          string                         `json:"proxyGroupID"`
// 		Quantity              int                            `json:"quantity"`
// 		AmazonTaskInfo        entities.AmazonTaskInfo        `json:"amazonTaskInfo"`
// 		BestbuyTaskInfo       entities.BestbuyTaskInfo       `json:"bestbuyTaskInfo"`
// 		BoxlunchTaskInfo      entities.BoxlunchTaskInfo      `json:"boxlunchTaskInfo"`
// 		DisneyTaskInfo        entities.DisneyTaskInfo        `json:"disneyTaskInfo"`
// 		GamestopTaskInfo      entities.GamestopTaskInfo      `json:"gamestopTaskInfo"`
// 		HottopicTaskInfo      entities.HottopicTaskInfo      `json:"hottopicTaskInfo"`
// 		NeweggTaskInfo        entities.NeweggTaskInfo        `json:"neweggTaskInfo"`
// 		PokemonCenterTaskInfo entities.PokemonCenterTaskInfo `json:"pokemoncenterTaskInfo"`
// 		ShopifyTaskInfo       entities.ShopifyTaskInfo       `json:"shopifyTaskInfo"`
// 		TargetTaskInfo        entities.TargetTaskInfo        `json:"targetTaskInfo"`
// 		ToppsTaskInfo         entities.ToppsTaskInfo         `json:"toppsTaskInfo"`
// 		WalmartTaskInfo       entities.WalmartTaskInfo       `json:"walmartTaskInfo"`
// 	}

// 	params := mux.Vars(request)
// 	groupID, ok := params["GroupID"]
// 	if ok {
// 		taskGroup, err = stores.GetTaskGroup(groupID)
// 		if err == nil {
// 			body, err := ioutil.ReadAll(request.Body)
// 			if err == nil {
// 				updateTasksRequestInfo := UpdateTasksRequest{}
// 				err = json.Unmarshal(body, &updateTasksRequestInfo)
// 				if err == nil {
// 					singleTask := len(updateTasksRequestInfo.TaskIDs) == 1
// 					for _, taskID := range updateTasksRequestInfo.TaskIDs {
// 						task, err := stores.GetTask(taskID)
// 						if err == nil {
// 							taskStore := stores.GetTaskStore()
// 							var wasRunning bool
// 							wasRunning, err = taskStore.StopTask(&task)
// 							if err == nil {
// 								if updateTasksRequestInfo.ProfileID != "DO_NOT_UPDATE" {
// 									task.TaskProfileID = updateTasksRequestInfo.ProfileID
// 								}
// 								if updateTasksRequestInfo.ProxyGroupID != "DO_NOT_UPDATE" {
// 									task.TaskProxyGroupID = updateTasksRequestInfo.ProxyGroupID
// 								}
// 								if updateTasksRequestInfo.Quantity != -1 && updateTasksRequestInfo.Quantity > 0 {
// 									task.TaskQty = updateTasksRequestInfo.Quantity
// 								}
// 								// Let's see if we can abstract this
// 								switch taskGroup.MonitorRetailer {
// 								case enums.Amazon:
// 									if singleTask || updateTasksRequestInfo.AmazonTaskInfo.Email != "" {
// 										task.AmazonTaskInfo.Email = updateTasksRequestInfo.AmazonTaskInfo.Email
// 									}
// 									if singleTask || updateTasksRequestInfo.AmazonTaskInfo.Password != "" {
// 										task.AmazonTaskInfo.Password = updateTasksRequestInfo.AmazonTaskInfo.Password
// 									}

// 								case enums.BestBuy:
// 									if updateTasksRequestInfo.BestbuyTaskInfo.TaskType != "DO_NOT_UPDATE" {
// 										task.BestbuyTaskInfo.TaskType = updateTasksRequestInfo.BestbuyTaskInfo.TaskType
// 									}
// 									if singleTask || updateTasksRequestInfo.BestbuyTaskInfo.Email != "" {
// 										task.BestbuyTaskInfo.Email = updateTasksRequestInfo.BestbuyTaskInfo.Email
// 									}
// 									if singleTask || updateTasksRequestInfo.BestbuyTaskInfo.Password != "" {
// 										task.BestbuyTaskInfo.Password = updateTasksRequestInfo.BestbuyTaskInfo.Password
// 									}
// 								case enums.BoxLunch:

// 								case enums.Disney:
// 									if updateTasksRequestInfo.DisneyTaskInfo.TaskType != "DO_NOT_UPDATE" {
// 										task.DisneyTaskInfo.TaskType = updateTasksRequestInfo.DisneyTaskInfo.TaskType
// 									}
// 									if singleTask || updateTasksRequestInfo.BestbuyTaskInfo.Email != "" {
// 										task.DisneyTaskInfo.Email = updateTasksRequestInfo.DisneyTaskInfo.Email
// 									}
// 									if singleTask || updateTasksRequestInfo.BestbuyTaskInfo.Password != "" {
// 										task.DisneyTaskInfo.Password = updateTasksRequestInfo.DisneyTaskInfo.Password
// 									}

// 								case enums.GameStop:
// 									if updateTasksRequestInfo.GamestopTaskInfo.TaskType != "DO_NOT_UPDATE" {
// 										task.GamestopTaskInfo.TaskType = updateTasksRequestInfo.GamestopTaskInfo.TaskType
// 									}
// 									if singleTask || updateTasksRequestInfo.GamestopTaskInfo.Email != "" {
// 										task.GamestopTaskInfo.Email = updateTasksRequestInfo.GamestopTaskInfo.Email
// 									}
// 									if singleTask || updateTasksRequestInfo.GamestopTaskInfo.Password != "" {
// 										task.GamestopTaskInfo.Password = updateTasksRequestInfo.GamestopTaskInfo.Password
// 									}

// 								case enums.HotTopic:
// 									// TODO @silent

// 								case enums.Newegg:

// 								case enums.PokemonCenter:
// 									if updateTasksRequestInfo.PokemonCenterTaskInfo.TaskType != "DO_NOT_UPDATE" {
// 										task.PokemonCenterTaskInfo.TaskType = updateTasksRequestInfo.PokemonCenterTaskInfo.TaskType
// 									}
// 									if singleTask || updateTasksRequestInfo.PokemonCenterTaskInfo.Email != "" {
// 										task.PokemonCenterTaskInfo.Email = updateTasksRequestInfo.PokemonCenterTaskInfo.Email
// 									}
// 									if singleTask || updateTasksRequestInfo.PokemonCenterTaskInfo.Password != "" {
// 										task.PokemonCenterTaskInfo.Password = updateTasksRequestInfo.PokemonCenterTaskInfo.Password
// 									}

// 								case enums.Shopify:
// 									if updateTasksRequestInfo.ShopifyTaskInfo.CouponCode != "DO_NOT_UPDATE" {
// 										task.ShopifyTaskInfo.CouponCode = updateTasksRequestInfo.ShopifyTaskInfo.CouponCode
// 									}
// 									if singleTask || updateTasksRequestInfo.ShopifyTaskInfo.HotWheelsTaskInfo.Email != "" {
// 										task.ShopifyTaskInfo.HotWheelsTaskInfo.Email = updateTasksRequestInfo.ShopifyTaskInfo.HotWheelsTaskInfo.Email
// 									}
// 									if singleTask || updateTasksRequestInfo.ShopifyTaskInfo.HotWheelsTaskInfo.Password != "" {
// 										task.ShopifyTaskInfo.HotWheelsTaskInfo.Password = updateTasksRequestInfo.ShopifyTaskInfo.HotWheelsTaskInfo.Password
// 									}

// 								case enums.Target:
// 									if updateTasksRequestInfo.TargetTaskInfo.CheckoutType != "DO_NOT_UPDATE" {
// 										task.TargetTaskInfo.CheckoutType = updateTasksRequestInfo.TargetTaskInfo.CheckoutType
// 									}
// 									if updateTasksRequestInfo.TargetTaskInfo.PaymentType != "DO_NOT_UPDATE" {
// 										task.TargetTaskInfo.PaymentType = updateTasksRequestInfo.TargetTaskInfo.PaymentType
// 									}
// 									if singleTask || updateTasksRequestInfo.TargetTaskInfo.Email != "" {
// 										task.TargetTaskInfo.Email = updateTasksRequestInfo.TargetTaskInfo.Email
// 									}
// 									if singleTask || updateTasksRequestInfo.TargetTaskInfo.Password != "" {
// 										task.TargetTaskInfo.Password = updateTasksRequestInfo.TargetTaskInfo.Password
// 									}

// 								case enums.Topps:
// 									if updateTasksRequestInfo.ToppsTaskInfo.TaskType != "DO_NOT_UPDATE" {
// 										task.ToppsTaskInfo.TaskType = updateTasksRequestInfo.ToppsTaskInfo.TaskType
// 									}
// 									if singleTask || updateTasksRequestInfo.ToppsTaskInfo.Email != "" {
// 										task.ToppsTaskInfo.Email = updateTasksRequestInfo.ToppsTaskInfo.Email
// 									}
// 									if singleTask || updateTasksRequestInfo.ToppsTaskInfo.Password != "" {
// 										task.ToppsTaskInfo.Password = updateTasksRequestInfo.ToppsTaskInfo.Password
// 									}

// 								case enums.Walmart:

// 								}
// 								_, err = stores.UpdateTask(taskID, task)
// 								if err == nil {
// 									task.UpdateTask = true
// 									if wasRunning {
// 										err = taskStore.StartTask(&task)
// 										if err != nil {
// 											errorsList = append(errorsList, errors.StartTaskError+err.Error())
// 										}
// 									}
// 								} else {
// 									errorsList = append(errorsList, errors.UpdateTaskError+err.Error())
// 								}
// 							} else {
// 								errorsList = append(errorsList, errors.StopTaskError+err.Error())
// 							}
// 						} else {
// 							errorsList = append(errorsList, errors.GetTaskError+err.Error())
// 						}
// 					}
// 				} else {
// 					errorsList = append(errorsList, errors.ParseUpdateTasksRequestError+err.Error())
// 				}
// 			} else {
// 				errorsList = append(errorsList, errors.IOUtilReadAllError+err.Error())
// 			}
// 		} else {
// 			errorsList = append(errorsList, errors.GetTaskGroupError+err.Error())
// 		}
// 	} else {
// 		errorsList = append(errorsList, errors.MissingParameterError)
// 	}

// 	taskGroup, err := stores.ConvertTaskIDsToTasks(&taskGroup)
// 	if err != nil {
// 		errorsList = append(errorsList, errors.GetTaskError+err.Error())
// 	}
// 	data := []entities.TaskGroup{UpdateStatuses(taskGroup)}

// 	result := &responses.TaskGroupResponse{Success: true, Data: data, Errors: make([]string, 0)}
// 	if len(errorsList) > 0 {
// 		response.WriteHeader(http.StatusBadRequest)
// 		result = &responses.TaskGroupResponse{Success: false, Data: data, Errors: errorsList}
// 	}
// 	json.NewEncoder(response).Encode(result)
// }
