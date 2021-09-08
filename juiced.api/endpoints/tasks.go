package endpoints

// // UpdateTaskGroupEndpoint handles the PUT request at /api/task/group/{GroupID}
// func UpdateTaskGroupEndpoint(response http.ResponseWriter, request *http.Request) {
// 	response.Header().Set("content-type", "application/json")
// 	response.Header().Set("Access-Control-Allow-Origin", "http://localhost:3000")
// 	var newTaskGroup entities.TaskGroup
// 	errorsList := make([]string, 0)

// 	// Let's see if we can abstract this
// 	type AmazonUpdateInfo struct {
// 		MonitorType enums.MonitorType `json:"monitorType"`
// 	}
// 	type BestBuyUpdateInfo struct{}
// 	type BoxlunchUpdateInfo struct {
// 		Sizes  string `json:"sizes"`
// 		Colors string `json:"colors"`
// 	}
// 	type DisneyUpdateInfo struct {
// 		Sizes  string `json:"sizes"`
// 		Colors string `json:"colors"`
// 	}
// 	type GamestopUpdateInfo struct{}
// 	type NeweggUpdateInfo struct{}
// 	type PokemonCenterUpdateInfo struct{}
// 	type HottopicUpdateInfo struct {
// 		Sizes  string `json:"sizes"`
// 		Colors string `json:"colors"`
// 	}
// 	type ShopifyUpdateInfo struct{}

// 	type ToppsUpdateInfo struct{}
// 	type TargetUpdateInfo struct {
// 		CheckoutType enums.CheckoutType `json:"checkoutType"`
// 		StoreID      string             `json:"storeID"`
// 	}
// 	type WalmartUpdateInfo struct {
// 		SoldByWalmart bool `json:"soldByWalmart"`
// 	}

// 	// Let's see if we can abstract this
// 	type UpdateTaskGroupRequest struct {
// 		Name                    string                  `json:"name"`
// 		MonitorInput            string                  `json:"input"`
// 		MonitorDelay            int                     `json:"delay"`
// 		MonitorProxyGroupID     string                  `json:"proxyGroupId"`
// 		MaxPrice                int                     `json:"maxPrice"`
// 		AmazonUpdateInfo        AmazonUpdateInfo        `json:"amazonUpdateInfo"`
// 		BestbuyUpdateInfo       BestBuyUpdateInfo       `json:"bestbuyUpdateInfo"`
// 		BoxlunchUpdateInfo      BoxlunchUpdateInfo      `json:"boxlunchUpdateInfo"`
// 		DisneyUpdateInfo        DisneyUpdateInfo        `json:"disneyUpdateInfo"`
// 		GamestopUpdateInfo      GamestopUpdateInfo      `json:"gamestopUpdateInfo"`
// 		HottopicUpdateInfo      HottopicUpdateInfo      `json:"hottopicUpdateInfo"`
// 		NeweggUpdateInfo        NeweggUpdateInfo        `json:"neweggUpdateInfo"`
// 		PokemonCenterUpdateInfo PokemonCenterUpdateInfo `json:"pokemoncenterUpdateInfo"`
// 		ShopifyUpdateInfo       ShopifyUpdateInfo       `json:"shopifyUpdateInfo"`
// 		TargetUpdateInfo        TargetUpdateInfo        `json:"targetUpdateInfo"`
// 		ToppsUpdateInfo         ToppsUpdateInfo         `json:"toppsUpdateInfo"`
// 		WalmartUpdateInfo       WalmartUpdateInfo       `json:"walmartUpdateInfo"`
// 	}

// 	params := mux.Vars(request)
// 	groupID, ok := params["GroupID"]
// 	if ok {
// 		taskGroup, err := stores.GetTaskGroup(groupID)
// 		if err == nil {
// 			monitorStore := stores.GetMonitorStore()
// 			err = monitorStore.StopMonitor(&taskGroup)
// 			if err == nil {
// 				body, err := ioutil.ReadAll(request.Body)
// 				if err == nil {
// 					updateTaskGroupRequestInfo := UpdateTaskGroupRequest{}
// 					err = json.Unmarshal(body, &updateTaskGroupRequestInfo)
// 					if err == nil {
// 						taskGroup.Name = updateTaskGroupRequestInfo.Name
// 						taskGroup.MonitorDelay = updateTaskGroupRequestInfo.MonitorDelay
// 						taskGroup.MonitorProxyGroupID = updateTaskGroupRequestInfo.MonitorProxyGroupID
// 						maxPrice := updateTaskGroupRequestInfo.MaxPrice
// 						// Let's see if we can abstract this
// 						switch taskGroup.MonitorRetailer {
// 						case enums.Amazon:
// 							newMonitors := make([]entities.AmazonSingleMonitorInfo, 0)
// 							if updateTaskGroupRequestInfo.MonitorInput != "" {
// 								skus := strings.Split(updateTaskGroupRequestInfo.MonitorInput, ",")
// 								for _, sku := range skus {
// 									monitor := entities.AmazonSingleMonitorInfo{
// 										MonitorID:   uuid.New().String(),
// 										TaskGroupID: taskGroup.GroupID,
// 										MaxPrice:    maxPrice,
// 									}
// 									switch updateTaskGroupRequestInfo.AmazonUpdateInfo.MonitorType {
// 									case enums.SlowSKUMonitor:
// 										monitor.ASIN = sku
// 									case enums.FastSKUMonitor:
// 										monitor.OFID = sku
// 									}
// 									newMonitors = append(newMonitors, monitor)
// 								}
// 								taskGroup.AmazonMonitorInfo.Monitors = newMonitors
// 							}

// 						case enums.BestBuy:
// 							newMonitors := make([]entities.BestbuySingleMonitorInfo, 0)
// 							if updateTaskGroupRequestInfo.MonitorInput != "" {
// 								skus := strings.Split(updateTaskGroupRequestInfo.MonitorInput, ",")
// 								for _, sku := range skus {
// 									monitor := entities.BestbuySingleMonitorInfo{
// 										MonitorID:   uuid.New().String(),
// 										TaskGroupID: taskGroup.GroupID,
// 										SKU:         sku,
// 										MaxPrice:    maxPrice,
// 									}
// 									newMonitors = append(newMonitors, monitor)
// 								}
// 								taskGroup.BestbuyMonitorInfo.Monitors = newMonitors
// 							}

// 						case enums.BoxLunch:
// 							newMonitors := make([]entities.BoxlunchSingleMonitorInfo, 0)
// 							if updateTaskGroupRequestInfo.MonitorInput != "" {
// 								pids := strings.Split(updateTaskGroupRequestInfo.MonitorInput, ",")
// 								monitorType := enums.SKUMonitor
// 								if len(taskGroup.BoxlunchMonitorInfo.Monitors) > 0 {
// 									monitorType = taskGroup.BoxlunchMonitorInfo.Monitors[0].MonitorType
// 								}
// 								for _, pid := range pids {
// 									monitor := entities.BoxlunchSingleMonitorInfo{
// 										MonitorID:   uuid.New().String(),
// 										TaskGroupID: taskGroup.GroupID,
// 										Pid:         pid,
// 										Size:        updateTaskGroupRequestInfo.BoxlunchUpdateInfo.Sizes,
// 										Color:       updateTaskGroupRequestInfo.BoxlunchUpdateInfo.Colors,
// 										MaxPrice:    maxPrice,
// 										MonitorType: monitorType,
// 									}
// 									newMonitors = append(newMonitors, monitor)
// 								}
// 								taskGroup.BoxlunchMonitorInfo.Monitors = newMonitors
// 							}

// 						case enums.Disney:
// 							newMonitors := make([]entities.DisneySingleMonitorInfo, 0)
// 							if updateTaskGroupRequestInfo.MonitorInput != "" {
// 								pids := strings.Split(updateTaskGroupRequestInfo.MonitorInput, ",")
// 								for _, pid := range pids {
// 									monitor := entities.DisneySingleMonitorInfo{
// 										MonitorID:   uuid.New().String(),
// 										TaskGroupID: taskGroup.GroupID,
// 										PID:         pid,
// 										Size:        updateTaskGroupRequestInfo.DisneyUpdateInfo.Sizes,
// 										Color:       updateTaskGroupRequestInfo.DisneyUpdateInfo.Colors,
// 										MaxPrice:    maxPrice,
// 									}
// 									newMonitors = append(newMonitors, monitor)
// 								}
// 								taskGroup.DisneyMonitorInfo.Monitors = newMonitors
// 							}

// 						case enums.GameStop:
// 							newMonitors := make([]entities.GamestopSingleMonitorInfo, 0)
// 							if updateTaskGroupRequestInfo.MonitorInput != "" {
// 								skus := strings.Split(updateTaskGroupRequestInfo.MonitorInput, ",")
// 								for _, sku := range skus {
// 									monitor := entities.GamestopSingleMonitorInfo{
// 										MonitorID:   uuid.New().String(),
// 										TaskGroupID: taskGroup.GroupID,
// 										SKU:         sku,
// 										MaxPrice:    maxPrice,
// 									}
// 									newMonitors = append(newMonitors, monitor)
// 								}
// 								taskGroup.GamestopMonitorInfo.Monitors = newMonitors
// 							}

// 						case enums.HotTopic:
// 							newMonitors := make([]entities.HottopicSingleMonitorInfo, 0)
// 							if updateTaskGroupRequestInfo.MonitorInput != "" {
// 								pids := strings.Split(updateTaskGroupRequestInfo.MonitorInput, ",")
// 								monitorType := enums.SKUMonitor
// 								if len(taskGroup.HottopicMonitorInfo.Monitors) > 0 {
// 									monitorType = taskGroup.HottopicMonitorInfo.Monitors[0].MonitorType
// 								}
// 								for _, pid := range pids {
// 									monitor := entities.HottopicSingleMonitorInfo{
// 										MonitorID:   uuid.New().String(),
// 										TaskGroupID: taskGroup.GroupID,
// 										Pid:         pid,
// 										Size:        updateTaskGroupRequestInfo.HottopicUpdateInfo.Sizes,
// 										Color:       updateTaskGroupRequestInfo.HottopicUpdateInfo.Colors,
// 										MaxPrice:    maxPrice,
// 										MonitorType: monitorType,
// 									}
// 									newMonitors = append(newMonitors, monitor)
// 								}
// 								taskGroup.HottopicMonitorInfo.Monitors = newMonitors
// 							}

// 						case enums.Newegg:
// 							newMonitors := make([]entities.NeweggSingleMonitorInfo, 0)
// 							if updateTaskGroupRequestInfo.MonitorInput != "" {
// 								skus := strings.Split(updateTaskGroupRequestInfo.MonitorInput, ",")
// 								for _, sku := range skus {
// 									monitor := entities.NeweggSingleMonitorInfo{
// 										MonitorID:   uuid.New().String(),
// 										TaskGroupID: taskGroup.GroupID,
// 										SKU:         sku,
// 										MaxPrice:    maxPrice,
// 									}
// 									newMonitors = append(newMonitors, monitor)
// 								}
// 								taskGroup.NeweggMonitorInfo.Monitors = newMonitors
// 							}

// 						case enums.PokemonCenter:
// 							maxPrice := -1
// 							if len(taskGroup.PokemonCenterMonitorInfo.Monitors) > 0 {
// 								maxPrice = taskGroup.PokemonCenterMonitorInfo.Monitors[0].MaxPrice
// 							}

// 							newMonitors := make([]entities.PokemonCenterSingleMonitorInfo, 0)
// 							if updateTaskGroupRequestInfo.MonitorInput != "" {
// 								skus := strings.Split(updateTaskGroupRequestInfo.MonitorInput, ",")
// 								for _, sku := range skus {
// 									monitor := entities.PokemonCenterSingleMonitorInfo{
// 										MonitorID:   uuid.New().String(),
// 										TaskGroupID: taskGroup.GroupID,
// 										SKU:         sku,
// 										MaxPrice:    maxPrice,
// 									}
// 									newMonitors = append(newMonitors, monitor)
// 								}
// 							}
// 							taskGroup.PokemonCenterMonitorInfo.Monitors = newMonitors

// 						case enums.Shopify:
// 							newMonitors := make([]entities.ShopifySingleMonitorInfo, 0)
// 							if updateTaskGroupRequestInfo.MonitorInput != "" {
// 								vids := strings.Split(updateTaskGroupRequestInfo.MonitorInput, ",")
// 								for _, vid := range vids {
// 									monitor := entities.ShopifySingleMonitorInfo{
// 										MonitorID:   uuid.New().String(),
// 										TaskGroupID: taskGroup.GroupID,
// 										VariantID:   vid,
// 										MaxPrice:    maxPrice,
// 									}
// 									newMonitors = append(newMonitors, monitor)
// 								}
// 								taskGroup.ShopifyMonitorInfo.Monitors = newMonitors
// 							}

// 						case enums.Target:
// 							newMonitors := make([]entities.TargetSingleMonitorInfo, 0)
// 							if updateTaskGroupRequestInfo.MonitorInput != "" {
// 								tcins := strings.Split(updateTaskGroupRequestInfo.MonitorInput, ",")
// 								for _, tcin := range tcins {
// 									monitor := entities.TargetSingleMonitorInfo{
// 										MonitorID:    uuid.New().String(),
// 										TaskGroupID:  taskGroup.GroupID,
// 										TCIN:         tcin,
// 										MaxPrice:     maxPrice,
// 										CheckoutType: updateTaskGroupRequestInfo.TargetUpdateInfo.CheckoutType,
// 									}
// 									newMonitors = append(newMonitors, monitor)
// 								}
// 								taskGroup.TargetMonitorInfo.StoreID = updateTaskGroupRequestInfo.TargetUpdateInfo.StoreID
// 								taskGroup.TargetMonitorInfo.Monitors = newMonitors
// 							}

// 						case enums.Topps:
// 							newMonitors := make([]entities.ToppsSingleMonitorInfo, 0)
// 							if updateTaskGroupRequestInfo.MonitorInput != "" {
// 								items := strings.Split(updateTaskGroupRequestInfo.MonitorInput, ",")
// 								for _, item := range items {
// 									monitor := entities.ToppsSingleMonitorInfo{
// 										MonitorID:   uuid.New().String(),
// 										TaskGroupID: taskGroup.GroupID,
// 										Item:        item,
// 										MaxPrice:    maxPrice,
// 									}
// 									newMonitors = append(newMonitors, monitor)
// 								}
// 								taskGroup.ToppsMonitorInfo.Monitors = newMonitors
// 							}

// 						case enums.Walmart:
// 							newMonitors := make([]entities.WalmartSingleMonitorInfo, 0)
// 							if updateTaskGroupRequestInfo.MonitorInput != "" {
// 								ids := strings.Split(updateTaskGroupRequestInfo.MonitorInput, ",")
// 								monitorType := enums.SKUMonitor
// 								if len(taskGroup.WalmartMonitorInfo.Monitors) > 0 {
// 									monitorType = taskGroup.WalmartMonitorInfo.Monitors[0].MonitorType
// 								}
// 								for _, id := range ids {
// 									monitor := entities.WalmartSingleMonitorInfo{
// 										MonitorID:     uuid.New().String(),
// 										TaskGroupID:   taskGroup.GroupID,
// 										ID:            id,
// 										MaxPrice:      maxPrice,
// 										SoldByWalmart: updateTaskGroupRequestInfo.WalmartUpdateInfo.SoldByWalmart,
// 										MonitorType:   monitorType,
// 									}
// 									newMonitors = append(newMonitors, monitor)
// 								}
// 								taskGroup.WalmartMonitorInfo.Monitors = newMonitors
// 							}

// 						}

// 						newTaskGroup, err = stores.UpdateTaskGroup(groupID, taskGroup)
// 						if err == nil {
// 							newTaskGroup.UpdateMonitor = true
// 							err = monitorStore.AddMonitorToStore(&newTaskGroup)
// 							if err != nil {
// 								errorsList = append(errorsList, errors.UpdateTaskGroupError+err.Error())
// 							}
// 						} else {
// 							errorsList = append(errorsList, errors.UpdateTaskGroupError+err.Error())
// 						}
// 					} else {
// 						errorsList = append(errorsList, errors.ParseUpdateTaskGroupRequestError+err.Error())
// 					}
// 				} else {
// 					errorsList = append(errorsList, errors.IOUtilReadAllError+err.Error())
// 				}
// 			} else {
// 				errorsList = append(errorsList, errors.StopMonitorError+err.Error())
// 			}
// 		} else {
// 			errorsList = append(errorsList, errors.GetTaskGroupError+err.Error())
// 		}
// 	} else {
// 		errorsList = append(errorsList, errors.MissingParameterError)
// 	}
// 	newTaskGroup, err := stores.ConvertTaskIDsToTasks(&newTaskGroup)
// 	if err != nil {
// 		errorsList = append(errorsList, errors.GetTaskError+err.Error())
// 	}
// 	data := []entities.TaskGroup{UpdateStatuses(newTaskGroup)}
// 	result := &responses.TaskGroupResponse{Success: true, Data: data, Errors: make([]string, 0)}
// 	if len(errorsList) > 0 {
// 		response.WriteHeader(http.StatusBadRequest)
// 		result = &responses.TaskGroupResponse{Success: false, Data: data, Errors: errorsList}
// 	}
// 	json.NewEncoder(response).Encode(result)
// }

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
