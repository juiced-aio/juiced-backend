package stores

import (
	e "errors"
	"strings"

	"backend.juicedbot.io/juiced.infrastructure/common"
	"backend.juicedbot.io/juiced.infrastructure/common/entities"
	"backend.juicedbot.io/juiced.infrastructure/common/enums"
	"backend.juicedbot.io/juiced.infrastructure/common/errors"
	"backend.juicedbot.io/juiced.infrastructure/common/events"
	"backend.juicedbot.io/juiced.infrastructure/queries"

	"backend.juicedbot.io/juiced.sitescripts/base"
	"backend.juicedbot.io/juiced.sitescripts/pokemoncenter"
	// Future sitescripts will be imported here
)

// TODO @silent: Handle TaskUpdatedEvent by updating the task's info in here
// TODO @silent: Handle TaskRemovedEvent by removing the task's info in here
// TODO @silent: Handle TaskGroupRemovedEvent by removing all of the task group's tasks in here

// TaskStore stores information about running Tasks
type TaskStore struct {
	Tasks    map[string]*base.Task
	EventBus *events.EventBus
}

// AddTaskToStore adds the Task to the TaskStore and returns true if successful
func (taskStore *TaskStore) AddTaskToStore(task *entities.Task) error {
	var queryError error
	// Get Profile, ProxyGroup for task
	profile, err := queries.GetProfile(task.TaskProfileID)
	if err != nil {
		queryError = err
	}
	var proxyGroup *entities.ProxyGroup
	if task.TaskProxyGroupID != "" {
		var ok bool
		proxyGroup, ok = proxyStore.ProxyGroups[task.TaskProxyGroupID]
		if !ok {
			queryError = e.New("proxy group failure")
		}
	}

	var taskInput interface{}

	switch task.TaskRetailer {
	// Let's see if we can abstract this
	//		Maybe make TaskInput an interface with a map[interface{}]interface{} and a CheckValidInput function?
	// Future sitescripts will have a case here
	// case enums.Amazon:
	// 	// Check if task exists in store already
	// 	if _, ok := taskStore.AmazonTasks[task.ID]; ok && !task.UpdateTask {
	// 		return nil
	// 	}
	// 	// Only return false on a query error if the task doesn't exist in the store already
	// 	if queryError != nil {
	// 		return queryError
	// 	}
	// 	// Make sure necessary fields exist
	// 	emptyString := ""
	// 	if task.AmazonTaskInfo.Email == emptyString || task.AmazonTaskInfo.Password == emptyString {
	// 		return e.New(errors.MissingTaskFieldsError)
	// 	}
	// 	if task.AmazonTaskInfo.LoginType == emptyString {
	// 		task.AmazonTaskInfo.LoginType = enums.LoginTypeBROWSER
	// 	}

	// 	// Create task
	// 	amazonTask, err := amazon.CreateAmazonTask(task, profile, proxyGroup, taskStore.EventBus, task.AmazonTaskInfo.LoginType, task.AmazonTaskInfo.Email, task.AmazonTaskInfo.Password)
	// 	if err != nil {
	// 		return e.New(errors.CreateBotTaskError + err.Error())
	// 	}
	// 	// Add task to store
	// 	taskStore.AmazonTasks[task.ID] = &amazonTask

	// case enums.BestBuy:
	// 	// Check if task exists in store already
	// 	if _, ok := taskStore.BestbuyTasks[task.ID]; ok && !task.UpdateTask {
	// 		return nil
	// 	}
	// 	// Only return false on a query error if the task doesn't exist in the store already
	// 	if queryError != nil {
	// 		return queryError
	// 	}
	// 	// Make sure necessary fields exist
	// 	emptyString := ""
	// 	if task.BestbuyTaskInfo.TaskType == emptyString || (task.BestbuyTaskInfo.TaskType == enums.TaskTypeAccount && (task.BestbuyTaskInfo.Email == emptyString || task.BestbuyTaskInfo.Password == emptyString)) {
	// 		return e.New(errors.MissingTaskFieldsError)
	// 	}
	// 	// Create task
	// 	bestbuyTask, err := bestbuy.CreateBestbuyTask(task, profile, proxyGroup, taskStore.EventBus, task.BestbuyTaskInfo.TaskType, task.BestbuyTaskInfo.LocationID, task.BestbuyTaskInfo.Email, task.BestbuyTaskInfo.Password)
	// 	if err != nil {
	// 		return e.New(errors.CreateBotTaskError + err.Error())
	// 	}
	// 	// Add task to store
	// 	taskStore.BestbuyTasks[task.ID] = &bestbuyTask

	// case enums.BoxLunch:
	// 	// Check if task exists in store already
	// 	if _, ok := taskStore.BoxlunchTasks[task.ID]; ok && !task.UpdateTask {
	// 		return nil
	// 	}
	// 	// Only return false on a query error if the task doesn't exist in the store already
	// 	if queryError != nil {
	// 		return queryError
	// 	}
	// 	// Create task
	// 	boxlunchTask, err := boxlunch.CreateBoxlunchTask(task, profile, proxyGroup, taskStore.EventBus)
	// 	if err != nil {
	// 		return e.New(errors.CreateBotTaskError + err.Error())
	// 	}
	// 	// Add task to store
	// 	taskStore.BoxlunchTasks[task.ID] = &boxlunchTask

	// case enums.GameStop:
	// 	// Check if task exists in store already
	// 	if _, ok := taskStore.GamestopTasks[task.ID]; ok && !task.UpdateTask {
	// 		return nil
	// 	}
	// 	// Only return false on a query error if the task doesn't exist in the store already
	// 	if queryError != nil {
	// 		return queryError
	// 	}

	// 	// Make sure necessary fields exist
	// 	emptyString := ""
	// 	if task.GamestopTaskInfo.TaskType == emptyString || (task.GamestopTaskInfo.TaskType == enums.TaskTypeAccount && (task.GamestopTaskInfo.Email == emptyString || task.GamestopTaskInfo.Password == emptyString)) {
	// 		return e.New(errors.MissingTaskFieldsError)
	// 	}

	// 	// Create task
	// 	gamestopTask, err := gamestop.CreateGamestopTask(task, profile, proxyGroup, taskStore.EventBus, task.GamestopTaskInfo.TaskType, task.GamestopTaskInfo.Email, task.GamestopTaskInfo.Password)
	// 	if err != nil {
	// 		return e.New(errors.CreateBotTaskError + err.Error())
	// 	}
	// 	// Add task to store
	// 	taskStore.GamestopTasks[task.ID] = &gamestopTask

	// case enums.Disney:
	// 	// Check if task exists in store already
	// 	if _, ok := taskStore.DisneyTasks[task.ID]; ok && !task.UpdateTask {
	// 		return nil
	// 	}
	// 	// Only return false on a query error if the task doesn't exist in the store already
	// 	if queryError != nil {
	// 		return queryError
	// 	}
	// 	// Make sure necessary fields exist
	// 	emptyString := ""
	// 	if task.DisneyTaskInfo.TaskType == emptyString || (task.DisneyTaskInfo.TaskType == enums.TaskTypeAccount && (task.DisneyTaskInfo.Email == emptyString || task.DisneyTaskInfo.Password == emptyString)) {
	// 		return e.New(errors.MissingTaskFieldsError)
	// 	}
	// 	// Create task
	// 	disneyTask, err := disney.CreateDisneyTask(task, profile, proxyGroup, taskStore.EventBus, task.DisneyTaskInfo.TaskType, task.DisneyTaskInfo.Email, task.DisneyTaskInfo.Password)
	// 	if err != nil {
	// 		return e.New(errors.CreateBotTaskError + err.Error())
	// 	}
	// 	// Add task to store
	// 	taskStore.DisneyTasks[task.ID] = &disneyTask

	// case enums.HotTopic:
	// 	// Check if task exists in store already
	// 	if _, ok := taskStore.HottopicTasks[task.ID]; ok && !task.UpdateTask {
	// 		return nil
	// 	}
	// 	// Only return false on a query error if the task doesn't exist in the store already
	// 	if queryError != nil {
	// 		return queryError
	// 	}
	// 	// Create task
	// 	hottopicTask, err := hottopic.CreateHottopicTask(task, profile, proxyGroup, taskStore.EventBus)
	// 	if err != nil {
	// 		return e.New(errors.CreateBotTaskError + err.Error())
	// 	}
	// 	// Add task to store
	// 	taskStore.HottopicTasks[task.ID] = &hottopicTask

	// case enums.Newegg:
	// 	// Check if task exists in store already
	// 	if _, ok := taskStore.NeweggTasks[task.ID]; ok && !task.UpdateTask {
	// 		return nil
	// 	}
	// 	// Only return false on a query error if the task doesn't exist in the store already
	// 	if queryError != nil {
	// 		return queryError
	// 	}
	// 	// Create task
	// 	neweggTask, err := newegg.CreateNeweggTask(task, profile, proxyGroup, taskStore.EventBus)
	// 	if err != nil {
	// 		return e.New(errors.CreateBotTaskError + err.Error())
	// 	}
	// 	// Add task to store
	// 	taskStore.NeweggTasks[task.ID] = &neweggTask

	// case enums.Shopify:
	// 	// Check if task exists in store already
	// 	if _, ok := taskStore.ShopifyTasks[task.ID]; ok && !task.UpdateTask {
	// 		return nil
	// 	}
	// 	// Only return false on a query error if the task doesn't exist in the store already
	// 	if queryError != nil {
	// 		return queryError
	// 	}
	// 	// Make sure necessary fields exist
	// 	emptyString := ""
	// 	if task.ShopifyTaskInfo.SiteURL == emptyString || task.ShopifyTaskInfo.ShopifyRetailer == emptyString {
	// 		return e.New(errors.MissingTaskFieldsError)
	// 	}

	// 	// Shopify Site specifics
	// 	site := task.ShopifyTaskInfo.ShopifyRetailer
	// 	switch site {
	// 	case enums.HotWheels:
	// 		if task.ShopifyTaskInfo.HotWheelsTaskInfo.Email != "" && task.ShopifyTaskInfo.HotWheelsTaskInfo.Password != "" {
	// 			return e.New(errors.MissingTaskFieldsError)
	// 		}
	// 	}
	// 	// Create task
	// 	shopifyTask, err := shopify.CreateShopifyTask(task, profile, proxyGroup, taskStore.EventBus, task.ShopifyTaskInfo.CouponCode, task.ShopifyTaskInfo.SiteURL, task.ShopifyTaskInfo.SitePassword, task.ShopifyTaskInfo.HotWheelsTaskInfo.Email, task.ShopifyTaskInfo.HotWheelsTaskInfo.Password)
	// 	if err != nil {
	// 		return e.New(errors.CreateBotTaskError + err.Error())
	// 	}
	// 	// Add task to store
	// 	taskStore.ShopifyTasks[task.ID] = &shopifyTask

	// case enums.Target:
	// 	// Check if task exists in store already
	// 	if _, ok := taskStore.TargetTasks[task.ID]; ok && !task.UpdateTask {
	// 		return nil
	// 	}
	// 	// Only return false on a query error if the task doesn't exist in the store already
	// 	if queryError != nil {
	// 		return queryError
	// 	}
	// 	// Make sure necessary fields exist
	// 	emptyString := ""
	// 	if task.TargetTaskInfo.Email == emptyString || task.TargetTaskInfo.Password == emptyString || task.TargetTaskInfo.PaymentType == emptyString {
	// 		return e.New(errors.MissingTaskFieldsError)
	// 	}
	// 	// Create task
	// 	targetTask, err := target.CreateTargetTask(task, profile, proxyGroup, taskStore.EventBus, task.TargetTaskInfo.Email, task.TargetTaskInfo.Password, task.TargetTaskInfo.PaymentType)
	// 	if err != nil {
	// 		return e.New(errors.CreateBotTaskError + err.Error())
	// 	}
	// 	// Add task to store
	// 	taskStore.TargetTasks[task.ID] = &targetTask

	// case enums.Topps:
	// 	// Check if task exists in store already
	// 	if _, ok := taskStore.ToppsTasks[task.ID]; ok && !task.UpdateTask {
	// 		return nil
	// 	}
	// 	// Only return false on a query error if the task doesn't exist in the store already
	// 	if queryError != nil {
	// 		return queryError
	// 	}

	// 	// Make sure necessary fields exist
	// 	emptyString := ""
	// 	if task.ToppsTaskInfo.TaskType == emptyString || (task.ToppsTaskInfo.TaskType == enums.TaskTypeAccount && (task.ToppsTaskInfo.Email == emptyString || task.ToppsTaskInfo.Password == emptyString)) {
	// 		return e.New(errors.MissingTaskFieldsError)
	// 	}

	// 	// Create task
	// 	toppsTask, err := topps.CreateToppsTask(task, profile, proxyGroup, taskStore.EventBus, task.ToppsTaskInfo.TaskType, task.ToppsTaskInfo.Email, task.ToppsTaskInfo.Password)
	// 	if err != nil {
	// 		return e.New(errors.CreateBotTaskError + err.Error())
	// 	}
	// 	// Add task to store
	// 	taskStore.ToppsTasks[task.ID] = &toppsTask

	// case enums.Walmart:
	// 	// Check if task exists in store already
	// 	if _, ok := taskStore.WalmartTasks[task.ID]; ok && !task.UpdateTask {
	// 		return nil
	// 	}
	// 	// Only return false on a query error if the task doesn't exist in the store already
	// 	if queryError != nil {
	// 		return queryError
	// 	}
	// 	// Create task
	// 	walmartTask, err := walmart.CreateWalmartTask(task, profile, proxyGroup, taskStore.EventBus)
	// 	if err != nil {
	// 		return e.New(errors.CreateBotTaskError + err.Error())
	// 	}
	// 	// Add task to store
	// 	taskStore.WalmartTasks[task.ID] = &walmartTask

	case enums.PokemonCenter:
		// Check if task exists in store already
		if _, ok := taskStore.Tasks[task.ID]; ok {
			return nil
		}
		// Only return false on a query error if the task doesn't exist in the store already
		if queryError != nil {
			return queryError
		}
		// Create task
		taskInput = pokemoncenter.TaskInput{
			Email:    task.PokemonCenterTaskInfo.Email,
			Password: task.PokemonCenterTaskInfo.Password,
			TaskType: task.PokemonCenterTaskInfo.TaskType,
		}
	}

	baseTask := &base.Task{Retailer: task.TaskRetailer}
	err = baseTask.CreateRetailerTask(task, profile, proxyGroup, taskStore.EventBus, taskInput)
	if err != nil {
		return e.New(errors.CreateBotTaskError + err.Error())
	}
	// Add task to store
	taskStore.Tasks[task.ID] = baseTask
	task.UpdateTask = false
	return nil
}

// StartTaskGroup runs the given TaskGroup's RunMonitor() function and the RunTask() function for each Task in the group and returns true if successful
func (taskStore *TaskStore) StartTaskGroup(taskGroup *entities.TaskGroup) ([]string, error) {
	// Start the task's TaskGroup (if it's already running, this will return true)
	var warnings []string
	err := monitorStore.StartMonitor(taskGroup)
	if err != nil {
		return warnings, err
	}

	for _, taskID := range taskGroup.TaskIDs {
		// Get the task
		task, err := queries.GetTask(taskID)
		if err == nil {
			profile, err := queries.GetProfile(task.TaskProfileID)
			if err == nil {
				if common.ValidCardType([]byte(profile.CreditCard.CardNumber), task.TaskRetailer) {
					// Add task to store (if it already exists, this will return true)
					err = taskStore.AddTaskToStore(&task)
					if err == nil {
						// Setting the stop flag to false before running the task
						taskStore.SetStopFlag(task.TaskRetailer, taskID, false)

						// If the Task is already running, then we're all set already
						if strings.Contains(task.TaskStatus, enums.TaskIdle) ||
							strings.Contains(task.TaskStatus, enums.TaskFailed) ||
							strings.Contains(task.TaskStatus, enums.CheckedOut) ||
							strings.Contains(task.TaskStatus, enums.CheckoutFailed) {
							// Otherwise, start the Task
							taskStore.RunTask(task.TaskRetailer, task.ID)
						}
					} else {
						warnings = append(warnings, err.Error())
					}
				} else {
					warnings = append(warnings, errors.StartTaskInvalidCardError+task.TaskRetailer)
				}
			} else {
				warnings = append(warnings, err.Error())
			}
		} else {
			warnings = append(warnings, err.Error())
		}
	}

	if len(taskGroup.TaskIDs) == len(warnings) {
		err = e.New(errors.StartMonitorInvalidCardError + taskGroup.MonitorRetailer)
		monitorStore.StopMonitor(taskGroup)
	}

	return warnings, err
}

// StopTaskGroup sets the stop field for the given TaskGroup's Monitor and each Task in the group and returns true if successful
func (taskStore *TaskStore) StopTaskGroup(taskGroup *entities.TaskGroup) error {
	// Stop the task's TaskGroup
	err := monitorStore.StopMonitor(taskGroup)
	if err != nil {
		return err
	}

	// Set the tasks StopFlags to true
	for _, taskID := range taskGroup.TaskIDs {
		taskStore.SetStopFlag(taskGroup.MonitorRetailer, taskID, true)
	}

	return nil
}

// StartTask runs the RunTask() function for the given Task and returns true if successful
func (taskStore *TaskStore) StartTask(task *entities.Task) error {
	profile, err := queries.GetProfile(task.TaskProfileID)
	if err != nil {
		return err
	}

	if !common.ValidCardType([]byte(profile.CreditCard.CardNumber), task.TaskRetailer) {
		return e.New(errors.StartTaskInvalidCardError + task.TaskRetailer)
	}

	taskGroup, err := queries.GetTaskGroup(task.TaskGroupID)
	if err != nil {
		return err
	}
	// Start the task's TaskGroup (if it's already running, this will return true)
	err = monitorStore.StartMonitor(&taskGroup)
	if err != nil {
		return err
	}

	// Add task to store (if it already exists, this will return true)
	err = taskStore.AddTaskToStore(task)
	if err != nil {
		return err
	}

	// If the Task is already running, then we're all set already
	if !strings.Contains(task.TaskStatus, enums.TaskIdle) &&
		!strings.Contains(task.TaskStatus, enums.TaskFailed) &&
		!strings.Contains(task.TaskStatus, enums.CheckedOut) &&
		!strings.Contains(task.TaskStatus, enums.CheckoutFailed) {
		return nil
	}

	// Set the task's StopFlag to true before running the task
	taskStore.SetStopFlag(task.TaskRetailer, task.ID, false)

	// Otherwise, start the Task
	taskStore.RunTask(task.TaskRetailer, task.ID)
	return nil
}

// StopTask sets the stop field for the given Task and returns true if successful
func (taskStore *TaskStore) StopTask(task *entities.Task) (bool, error) {
	if !taskStore.TasksRunning([]string{task.ID}, task.TaskRetailer) {
		return false, nil
	}
	return true, taskStore.SetStopFlag(task.TaskRetailer, task.ID, true)
}

// TasksRunning checks to see if any tasks in the taskGroup are running, if so it returns true
func (taskStore *TaskStore) TasksRunning(taskIDs []string, retailer enums.Retailer) bool {
	for _, taskID := range taskIDs {
		if baseTask, ok := taskStore.Tasks[taskID]; ok {
			taskInfo := baseTask.GetTaskInfo()
			if taskInfo != nil && !taskInfo.StopFlag {
				return true
			}
		}
	}

	return false
}

func (taskStore *TaskStore) UpdateTaskProxy(task *entities.Task, proxy *entities.Proxy) bool {
	if baseTask, ok := taskStore.Tasks[task.ID]; ok {
		taskInfo := baseTask.GetTaskInfo()
		if taskInfo == nil {
			return false
		}
		taskInfo.Proxy = proxy
		return true
	}

	return false
}

// func (taskStore *TaskStore) SetWalmartCardDetails(taskID string, cardInfo walmart.CardInfo) {
// 	if walmartTask, ok := taskStore.WalmartTasks[taskID]; ok {
// 		walmartTask.CardInfo = cardInfo
// 	}
// }

func (taskStore *TaskStore) RunTask(retailer enums.Retailer, taskID string) {
	if baseTask, ok := taskStore.Tasks[taskID]; ok {
		go baseTask.RunTask()
	}
}

var taskStore *TaskStore

// InitTaskStore initializes the singleton instance of the TaskStore
func InitTaskStore(eventBus *events.EventBus) {
	taskStore = &TaskStore{
		Tasks:    make(map[string]*base.Task),
		EventBus: eventBus,
	}
}

// GetTaskStatuses returns a list of tasks with the most up to date status
func GetTaskStatuses() map[string]string {
	taskStatuses := make(map[string]string)

	for taskID, baseTask := range taskStore.Tasks {
		taskInfo := baseTask.GetTaskInfo()
		taskStatuses[taskID] = taskInfo.Task.TaskStatus
	}

	return taskStatuses
}

// GetTaskStore returns the singleton instance of the EventBus
func GetTaskStore() *TaskStore {
	return taskStore
}
