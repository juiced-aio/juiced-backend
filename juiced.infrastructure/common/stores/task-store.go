package stores

import (
	"math/rand"

	"backend.juicedbot.io/juiced.infrastructure/common/entities"
	"backend.juicedbot.io/juiced.infrastructure/common/enums"
	"backend.juicedbot.io/juiced.infrastructure/common/events"
	"backend.juicedbot.io/juiced.infrastructure/queries"

	"backend.juicedbot.io/juiced.sitescripts/amazon"
	"backend.juicedbot.io/juiced.sitescripts/bestbuy"
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
		targetTask, err := target.CreateTargetTask(task, profile, proxy, taskStore.EventBus, task.TargetTaskInfo.CheckoutType, task.TargetTaskInfo.Email, task.TargetTaskInfo.Password, task.TargetTaskInfo.PaymentType)
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
		if task.GamestopTaskInfo.TaskType == "" || task.GamestopTaskInfo.Email == "" || task.GamestopTaskInfo.Password == "" {
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
	if task.TaskStatus != enums.TaskIdle {
		return true
	}

	// Otherwise, start the Task
	switch task.TaskRetailer {
	// Future sitescripts will have a case here
	case enums.Target:
		go taskStore.TargetTasks[task.ID].RunTask()

	case enums.Walmart:
		go taskStore.WalmartTasks[task.ID].RunTask()

	case enums.Amazon:
		go taskStore.AmazonTasks[task.ID].RunTask()

	case enums.BestBuy:
		go taskStore.BestbuyTasks[task.ID].RunTask()

	case enums.HotTopic:
		go taskStore.HottopicTasks[task.ID].RunTask()

	case enums.GameStop:
		go taskStore.GamestopTasks[task.ID].RunTask()
	}
	return true
}

// StopTask sets the stop field for the given Task and returns true if successful
func (taskStore *TaskStore) StopTask(task *entities.Task) bool {
	switch task.TaskRetailer {
	// Future sitescripts will have a case here
	case enums.Target:
		if targetTask, ok := taskStore.TargetTasks[task.ID]; ok {
			targetTask.Task.StopFlag = true
		}
		return true

	case enums.Walmart:
		if walmartTask, ok := taskStore.WalmartTasks[task.ID]; ok {
			walmartTask.Task.StopFlag = true
		}
		return true

	case enums.Amazon:
		if amazonTask, ok := taskStore.AmazonTasks[task.ID]; ok {
			amazonTask.Task.StopFlag = true
		}
		return true

	case enums.BestBuy:
		if bestbuyTask, ok := taskStore.BestbuyTasks[task.ID]; ok {
			bestbuyTask.Task.StopFlag = true
		}
		return true

	case enums.HotTopic:
		if hottopicTask, ok := taskStore.HottopicTasks[task.ID]; ok {
			hottopicTask.Task.StopFlag = true
		}
		return true

	case enums.GameStop:
		if gamestopTask, ok := taskStore.GamestopTasks[task.ID]; ok {
			gamestopTask.Task.StopFlag = true
		}
		return true
	}
	return false
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
	channel := make(chan events.Event)
	eventBus.Subscribe(channel)
	for {
		event := <-channel
		if event.EventType == events.ProductEventType {
			switch event.ProductEvent.Retailer {
			case enums.Target:
				inStockForPickup := event.ProductEvent.TargetData.InStockForPickup
				inStockForShip := event.ProductEvent.TargetData.InStockForShip
				for _, targetTask := range taskStore.TargetTasks {
					if targetTask.Task.Task.TaskGroupID == event.ProductEvent.MonitorID {
						if targetTask.CheckoutType == enums.CheckoutTypePICKUP && len(inStockForPickup) > 0 {
							targetTask.TCIN = inStockForPickup[rand.Intn(len(inStockForPickup))]
						} else {
							targetTask.TCIN = inStockForShip[rand.Intn(len(inStockForShip))]
						}
						targetTask.Task.DiscordWebhook = event.ProductEvent.DiscordWebhook
					}
				}

			case enums.Walmart:
				inStockForShip := event.ProductEvent.WalmartData.InStockForShip
				for _, walmartTask := range taskStore.WalmartTasks {
					if walmartTask.Task.Task.TaskGroupID == event.ProductEvent.MonitorID {
						walmartTask.Sku = inStockForShip[rand.Intn(len(inStockForShip))].Sku
						walmartTask.Task.DiscordWebhook = event.ProductEvent.DiscordWebhook
					}
				}

			case enums.Amazon:
				inStock := event.ProductEvent.AmazonData.InStock
				for _, amazonTask := range taskStore.AmazonTasks {
					if amazonTask.Task.Task.TaskGroupID == event.ProductEvent.MonitorID {
						amazonTask.TaskInfo.ASIN = inStock[rand.Intn(len(inStock))].ASIN
						amazonTask.TaskInfo.OfferID = inStock[rand.Intn(len(inStock))].OfferID
						amazonTask.TaskInfo.ItemName = inStock[rand.Intn(len(inStock))].ItemName
						amazonTask.CheckoutInfo.Price = inStock[rand.Intn(len(inStock))].Price
						amazonTask.CheckoutInfo.AntiCsrf = inStock[rand.Intn(len(inStock))].AntiCsrf
						amazonTask.CheckoutInfo.PID = inStock[rand.Intn(len(inStock))].PID
						amazonTask.CheckoutInfo.RID = inStock[rand.Intn(len(inStock))].RID
						amazonTask.CheckoutInfo.ImageURL = inStock[rand.Intn(len(inStock))].ImageURL
						amazonTask.CheckoutInfo.UA = inStock[rand.Intn(len(inStock))].UA
						amazonTask.CheckoutInfo.MonitorType = enums.MonitorType(inStock[rand.Intn(len(inStock))].MonitorType)
					}
				}

			case enums.BestBuy:
				inStock := event.ProductEvent.BestbuyData.InStock
				for _, bestbuyTask := range taskStore.BestbuyTasks {
					if bestbuyTask.Task.Task.TaskGroupID == event.ProductEvent.MonitorID {
						bestbuyTask.CheckoutInfo.SKUInStock = inStock[rand.Intn(len(inStock))].SKU
						bestbuyTask.CheckoutInfo.Price = inStock[rand.Intn(len(inStock))].Price
						bestbuyTask.Task.DiscordWebhook = event.ProductEvent.DiscordWebhook
					}
				}

			case enums.HotTopic:
				inStock := event.ProductEvent.HottopicData.InStock
				for _, hotTopicTask := range taskStore.HottopicTasks {
					if hotTopicTask.Task.Task.TaskGroupID == event.ProductEvent.MonitorID {
						hotTopicTask.Pid = inStock[rand.Intn(len(inStock))].PID
					}
				}

			case enums.GameStop:
				inStock := event.ProductEvent.GamestopData.InStock
				for _, gamestopTask := range taskStore.GamestopTasks {
					if gamestopTask.Task.Task.TaskGroupID == event.ProductEvent.MonitorID {
						gamestopTask.CheckoutInfo.SKUInStock = inStock[rand.Intn(len(inStock))].SKU
						gamestopTask.CheckoutInfo.Price = inStock[rand.Intn(len(inStock))].Price
						gamestopTask.CheckoutInfo.ItemName = inStock[rand.Intn(len(inStock))].ItemName
						gamestopTask.CheckoutInfo.PID = inStock[rand.Intn(len(inStock))].PID
						gamestopTask.CheckoutInfo.ImageURL = inStock[rand.Intn(len(inStock))].ImageURL
						gamestopTask.CheckoutInfo.ProductURL = inStock[rand.Intn(len(inStock))].ProductURL
						gamestopTask.Task.DiscordWebhook = event.ProductEvent.DiscordWebhook
					}
				}
			}
		}
	}
}

// GetTaskStore returns the singleton instance of the EventBus
func GetTaskStore() *TaskStore {
	return taskStore
}
