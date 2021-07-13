package stores

import (
	"math/rand"

	"backend.juicedbot.io/juiced.infrastructure/common/entities"
	"backend.juicedbot.io/juiced.infrastructure/common/enums"
	"backend.juicedbot.io/juiced.infrastructure/common/events"
	"backend.juicedbot.io/juiced.infrastructure/queries"

	"backend.juicedbot.io/juiced.sitescripts/amazon"
	"backend.juicedbot.io/juiced.sitescripts/bestbuy"
	"backend.juicedbot.io/juiced.sitescripts/boxlunch"
	"backend.juicedbot.io/juiced.sitescripts/gamestop"
	"backend.juicedbot.io/juiced.sitescripts/hottopic"
	"backend.juicedbot.io/juiced.sitescripts/target"
	"backend.juicedbot.io/juiced.sitescripts/walmart"
	// Future sitescripts will be imported here
)

// TODO @silent: Handle TaskUpdatedEvent by updating the task's info in here
// TODO @silent: Handle TaskRemovedEvent by removing the task's info in here
// TODO @silent: Handle TaskGroupRemovedEvent by removing all of the task group's tasks in here

// TaskStore stores information about running Tasks
type TaskStore struct {
	TargetTasks   map[string]*target.Task
	WalmartTasks  map[string]*walmart.Task
	AmazonTasks   map[string]*amazon.Task
	BestbuyTasks  map[string]*bestbuy.Task
	BoxlunchTasks map[string]*boxlunch.Task
	HottopicTasks map[string]*hottopic.Task
	GamestopTasks map[string]*gamestop.Task
	// Future sitescripts will have a field here
	EventBus *events.EventBus
}

// AddTaskToStore adds the Task to the TaskStore and returns true if successful
func (taskStore *TaskStore) AddTaskToStore(task *entities.Task) bool {
	queryError := false
	// Get Profile, Proxy for task
	profile, err := queries.GetProfile(task.TaskProfileID)
	if err != nil {
		queryError = true
	}
	proxy := entities.Proxy{}
	if task.TaskProxyGroupID != "" {
		proxyGroup, err := queries.GetProxyGroup(task.TaskProxyGroupID)
		if err != nil {
			queryError = true
		}
		proxy = proxyGroup.Proxies[rand.Intn(len(proxyGroup.Proxies))]
	}
	switch task.TaskRetailer {
	// Future sitescripts will have a case here
	case enums.Target:
		// Check if task exists in store already
		if _, ok := taskStore.TargetTasks[task.ID]; ok {
			return true
		}
		// Only return false on a query error if the task doesn't exist in the store already
		if queryError {
			return false
		}
		// Make sure necessary fields exist
		emptyString := ""
		if task.TargetTaskInfo.CheckoutType == emptyString || task.TargetTaskInfo.Email == emptyString || task.TargetTaskInfo.Password == emptyString || task.TargetTaskInfo.PaymentType == emptyString {
			return false
		}
		// Create task
		targetTask, err := target.CreateTargetTask(task, profile, proxy, taskStore.EventBus, task.TargetTaskInfo.CheckoutType, task.TargetTaskInfo.Email, task.TargetTaskInfo.Password, task.TargetTaskInfo.StoreID, task.TargetTaskInfo.PaymentType)
		if err != nil {
			return false
		}
		// Add task to store
		taskStore.TargetTasks[task.ID] = &targetTask

	case enums.Walmart:
		// Check if task exists in store already
		if _, ok := taskStore.WalmartTasks[task.ID]; ok {
			return true
		}
		// Only return false on a query error if the task doesn't exist in the store already
		if queryError {
			return false
		}
		// Create task
		walmartTask, err := walmart.CreateWalmartTask(task, profile, proxy, taskStore.EventBus)
		if err != nil {
			return false
		}
		// Add task to store
		taskStore.WalmartTasks[task.ID] = &walmartTask

	case enums.Amazon:
		// Check if task exists in store already
		if _, ok := taskStore.AmazonTasks[task.ID]; ok {
			return true
		}
		// Only return false on a query error if the task doesn't exist in the store already
		if queryError {
			return false
		}
		// Make sure necessary fields exist
		emptyString := ""
		if task.AmazonTaskInfo.LoginType == emptyString || task.AmazonTaskInfo.Email == emptyString || task.AmazonTaskInfo.Password == emptyString {
			return false
		}
		// Create task
		amazonTask, err := amazon.CreateAmazonTask(task, profile, proxy, taskStore.EventBus, task.AmazonTaskInfo.LoginType, task.AmazonTaskInfo.Email, task.AmazonTaskInfo.Password)
		if err != nil {
			return false
		}
		// Add task to store
		taskStore.AmazonTasks[task.ID] = &amazonTask

	case enums.BestBuy:
		// Check if task exists in store already
		if _, ok := taskStore.BestbuyTasks[task.ID]; ok {
			return true
		}
		// Only return false on a query error if the task doesn't exist in the store already
		if queryError {
			return false
		}
		// Make sure necessary fields exist
		emptyString := ""
		if task.BestbuyTaskInfo.TaskType == emptyString || (task.BestbuyTaskInfo.TaskType == enums.TaskTypeAccount && (task.BestbuyTaskInfo.Email == emptyString || task.BestbuyTaskInfo.Password == emptyString)) {
			return false
		}
		// Create task
		bestbuyTask, err := bestbuy.CreateBestbuyTask(task, profile, proxy, taskStore.EventBus, task.BestbuyTaskInfo.TaskType, task.BestbuyTaskInfo.Email, task.BestbuyTaskInfo.Password)
		if err != nil {
			return false
		}
		// Add task to store
		taskStore.BestbuyTasks[task.ID] = &bestbuyTask

	case enums.HotTopic:
		// Check if task exists in store already
		if _, ok := taskStore.HottopicTasks[task.ID]; ok {
			return true
		}
		// Only return false on a query error if the task doesn't exist in the store already
		if queryError {
			return false
		}
		// Make sure necessary fields exist
		if len(task.HottopicTaskInfo.Pids) == 0 {
			return false
		}
		// Create task
		hottopicTask, err := hottopic.CreateHottopicTask(task, profile, proxy, taskStore.EventBus)
		if err != nil {
			return false
		}
		// Add task to store
		taskStore.HottopicTasks[task.ID] = &hottopicTask

	case enums.BoxLunch:
		// Check if task exists in store already
		if _, ok := taskStore.BoxlunchTasks[task.ID]; ok {
			return true
		}
		// Only return false on a query error if the task doesn't exist in the store already
		if queryError {
			return false
		}
		// Make sure necessary fields exist
		if len(task.BoxlunchTaskInfo.Pids) == 0 {
			return false
		}
		// Create task
		boxlunchTask, err := boxlunch.CreateBoxlunchTask(task, profile, proxy, taskStore.EventBus)
		if err != nil {
			return false
		}
		// Add task to store
		taskStore.BoxlunchTasks[task.ID] = &boxlunchTask

	case enums.GameStop:
		// Check if task exists in store already
		if _, ok := taskStore.GamestopTasks[task.ID]; ok {
			return true
		}
		// Only return false on a query error if the task doesn't exist in the store already
		if queryError {
			return false
		}

		// Make sure necessary fields exist
		emptyString := ""
		if task.GamestopTaskInfo.TaskType == emptyString || (task.GamestopTaskInfo.TaskType == enums.TaskTypeAccount && (task.GamestopTaskInfo.Email == emptyString || task.GamestopTaskInfo.Password == emptyString)) {
			return false
		}

		// Create task
		gamestopTask, err := gamestop.CreateGamestopTask(task, profile, proxy, taskStore.EventBus, task.GamestopTaskInfo.TaskType, task.GamestopTaskInfo.Email, task.GamestopTaskInfo.Password)
		if err != nil {
			return false
		}
		// Add task to store
		taskStore.GamestopTasks[task.ID] = &gamestopTask

	}
	return true
}

// StartTaskGroup runs the given TaskGroup's RunMonitor() function and the RunTask() function for each Task in the group and returns true if successful
func (taskStore *TaskStore) StartTaskGroup(taskGroup *entities.TaskGroup) bool {
	// Start the task's TaskGroup (if it's already running, this will return true)
	started := monitorStore.StartMonitor(taskGroup)
	if !started {
		return false
	}

	for _, taskID := range taskGroup.TaskIDs {
		// Get the task
		task, err := queries.GetTask(taskID)
		if err != nil {
			return false
		}

		// Add task to store (if it already exists, this will return true)
		added := taskStore.AddTaskToStore(&task)
		if !added {
			return false
		}

		// Setting the stop flag to false before running the task
		taskStore.SetStopFlag(task.TaskRetailer, taskID, false)

		// If the Task is already running, then we're all set already
		if task.TaskStatus == enums.TaskIdle ||
			task.TaskStatus == enums.CheckedOut ||
			task.TaskStatus == enums.CheckoutFailed {
			// Otherwise, start the Task
			taskStore.RunTask(task.TaskRetailer, task.ID)
		}
	}

	return true
}

// StopTaskGroup sets the stop field for the given TaskGroup's Monitor and each Task in the group and returns true if successful
func (taskStore *TaskStore) StopTaskGroup(taskGroup *entities.TaskGroup) bool {
	// Stop the task's TaskGroup
	stopped := monitorStore.StopMonitor(taskGroup)
	if !stopped {
		return false
	}

	// Set the tasks StopFlags to true
	for _, taskID := range taskGroup.TaskIDs {
		taskStore.SetStopFlag(taskGroup.MonitorRetailer, taskID, true)

	}

	return true
}

// StartTask runs the RunTask() function for the given Task and returns true if successful
func (taskStore *TaskStore) StartTask(task *entities.Task) bool {
	taskGroup, err := queries.GetTaskGroup(task.TaskGroupID)
	if err != nil {
		return false
	}
	// Start the task's TaskGroup (if it's already running, this will return true)
	started := monitorStore.StartMonitor(&taskGroup)
	if !started {
		return false
	}

	// Add task to store (if it already exists, this will return true)
	added := taskStore.AddTaskToStore(task)
	if !added {
		return false
	}

	// If the Task is already running, then we're all set already
	if task.TaskStatus != enums.TaskIdle &&
		task.TaskStatus != enums.CheckedOut &&
		task.TaskStatus != enums.CheckoutFailed {
		return true
	}

	// Set the task's StopFlag to true before running the task
	taskStore.SetStopFlag(task.TaskRetailer, task.ID, false)

	// Otherwise, start the Task
	taskStore.RunTask(task.TaskRetailer, task.ID)
	return true
}

// StopTask sets the stop field for the given Task and returns true if successful
func (taskStore *TaskStore) StopTask(task *entities.Task) bool {
	return taskStore.SetStopFlag(task.TaskRetailer, task.ID, true)
}

// TasksRunning checks to see if any tasks in the taskGroup are running, if so it returns true
func (taskStore *TaskStore) TasksRunning(taskGroup *entities.TaskGroup) bool {
	for _, taskID := range taskGroup.TaskIDs {
		switch taskGroup.MonitorRetailer {
		// Future sitescripts will have a case here
		case enums.Target:
			if targetTask, ok := taskStore.TargetTasks[taskID]; ok {
				if !targetTask.Task.StopFlag {
					return true
				}
			}

		case enums.Walmart:
			if walmartTask, ok := taskStore.WalmartTasks[taskID]; ok {
				if !walmartTask.Task.StopFlag {
					return true
				}
			}

		case enums.Amazon:
			if amazonTask, ok := taskStore.AmazonTasks[taskID]; ok {
				if !amazonTask.Task.StopFlag {
					return true
				}
			}

		case enums.BestBuy:
			if bestbuyTask, ok := taskStore.BestbuyTasks[taskID]; ok {
				if !bestbuyTask.Task.StopFlag {
					return true
				}
			}

		case enums.HotTopic:
			if hottopicTask, ok := taskStore.HottopicTasks[taskID]; ok {
				if !hottopicTask.Task.StopFlag {
					return true
				}
			}

		case enums.GameStop:
			if gamestopTask, ok := taskStore.GamestopTasks[taskID]; ok {
				if !gamestopTask.Task.StopFlag {
					return true
				}
			}
		}
	}

	return false
}

func (taskStore *TaskStore) UpdateTaskProxy(task *entities.Task, proxy entities.Proxy) bool {
	switch task.TaskRetailer {
	case enums.Target:
		if targetTask, ok := taskStore.TargetTasks[task.ID]; ok {
			targetTask.Task.Proxy = proxy
		}
		return true

	case enums.Walmart:
		if walmartTask, ok := taskStore.WalmartTasks[task.ID]; ok {
			walmartTask.Task.Proxy = proxy
		}
		return true

	case enums.Amazon:
		if amazonTask, ok := taskStore.AmazonTasks[task.ID]; ok {
			amazonTask.Task.Proxy = proxy
		}
		return true

	case enums.BestBuy:
		if bestbuyTask, ok := taskStore.BestbuyTasks[task.ID]; ok {
			bestbuyTask.Task.Proxy = proxy
		}
		return true

	case enums.HotTopic:
		if hottopicTask, ok := taskStore.HottopicTasks[task.ID]; ok {
			hottopicTask.Task.Proxy = proxy
		}
		return true

	case enums.GameStop:
		if gamestopTask, ok := taskStore.GamestopTasks[task.ID]; ok {
			gamestopTask.Task.Proxy = proxy
		}
		return true
	}
	return false
}

func (taskStore *TaskStore) SetWalmartCardDetails(taskID string, cardInfo walmart.CardInfo) {
	if walmartTask, ok := taskStore.WalmartTasks[taskID]; ok {
		walmartTask.CardInfo = cardInfo
	}
}

func (taskStore *TaskStore) RunTask(retailer enums.Retailer, taskID string) {
	switch retailer {
	// Future sitescripts will have a case here
	case enums.Target:
		go taskStore.TargetTasks[taskID].RunTask()

	case enums.Walmart:
		go taskStore.WalmartTasks[taskID].RunTask()

	case enums.Amazon:
		go taskStore.AmazonTasks[taskID].RunTask()

	case enums.BestBuy:
		go taskStore.BestbuyTasks[taskID].RunTask()

	case enums.HotTopic:
		go taskStore.HottopicTasks[taskID].RunTask()

	case enums.GameStop:
		go taskStore.GamestopTasks[taskID].RunTask()
	}
}

var taskStore *TaskStore

// InitTaskStore initializes the singleton instance of the TaskStore
func InitTaskStore(eventBus *events.EventBus) {
	taskStore = &TaskStore{
		TargetTasks:   make(map[string]*target.Task),
		WalmartTasks:  make(map[string]*walmart.Task),
		AmazonTasks:   make(map[string]*amazon.Task),
		BestbuyTasks:  make(map[string]*bestbuy.Task),
		HottopicTasks: make(map[string]*hottopic.Task),
		GamestopTasks: make(map[string]*gamestop.Task),
		EventBus:      eventBus,
	}
}

// GetTaskStore returns the singleton instance of the EventBus
func GetTaskStore() *TaskStore {
	return taskStore
}
