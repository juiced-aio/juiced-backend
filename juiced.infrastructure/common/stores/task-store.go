package stores

import (
	"math/rand"

	"go.mongodb.org/mongo-driver/bson/primitive"

	"backend.juicedbot.io/m/v2/juiced.infrastructure/common/entities"
	"backend.juicedbot.io/m/v2/juiced.infrastructure/common/enums"
	"backend.juicedbot.io/m/v2/juiced.infrastructure/common/events"
	"backend.juicedbot.io/m/v2/juiced.infrastructure/queries"

	"backend.juicedbot.io/m/v2/juiced.sitescripts/amazon"
	"backend.juicedbot.io/m/v2/juiced.sitescripts/target"
	"backend.juicedbot.io/m/v2/juiced.sitescripts/walmart"
	// Future sitescripts will be imported here
)

// TODO @silent: Handle TaskUpdatedEvent by updating the task's info in here
// TODO @silent: Handle TaskRemovedEvent by removing the task's info in here
// TODO @silent: Handle TaskGroupRemovedEvent by removing all of the task group's tasks in here

// TaskStore stores information about running Tasks
type TaskStore struct {
	TargetTasks  map[primitive.ObjectID]*target.Task
	WalmartTasks map[primitive.ObjectID]*walmart.Task
	AmazonTasks  map[primitive.ObjectID]*amazon.Task
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
	if !task.TaskProxyGroupID.IsZero() {
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
		if task.TargetTaskInfo.CheckoutType == "" || task.TargetTaskInfo.Email == "" || task.TargetTaskInfo.Password == "" || task.TargetTaskInfo.PaymentType == "" {
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
		if _, ok := taskStore.TargetTasks[task.ID]; ok {
			return true
		}
		// Only return false on a query error if the task doesn't exist in the store already
		if queryError {
			return false
		}
		// Make sure necessary fields exist
		if task.AmazonTaskInfo.LoginType == "" || task.AmazonTaskInfo.Email == "" || task.AmazonTaskInfo.Password == "" {
			return false
		}
		// Create task
		amazonTask, err := amazon.CreateAmazonTask(task, profile, proxy, taskStore.EventBus, task.AmazonTaskInfo.LoginType, task.AmazonTaskInfo.Email, task.AmazonTaskInfo.Password)
		if err != nil {
			return false
		}
		// Add task to store
		taskStore.AmazonTasks[task.ID] = &amazonTask

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
	defer func() {
		recover()
		// TODO @silent: Let the UI know that a task failed
	}()
	// May panic (if it runs into a runtime error)
	switch task.TaskRetailer {
	// Future sitescripts will have a case here
	case enums.Target:
		go taskStore.TargetTasks[task.ID].RunTask()
	case enums.Walmart:
		go taskStore.WalmartTasks[task.ID].RunTask()
	case enums.Amazon:
		go taskStore.AmazonTasks[task.ID].RunTask()
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
	}
	return false
}

var taskStore *TaskStore

// InitTaskStore initializes the singleton instance of the TaskStore
func InitTaskStore(eventBus *events.EventBus) {
	taskStore = &TaskStore{
		TargetTasks:  make(map[primitive.ObjectID]*target.Task),
		WalmartTasks: make(map[primitive.ObjectID]*walmart.Task),
		AmazonTasks:  make(map[primitive.ObjectID]*amazon.Task),
		EventBus:     eventBus,
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
					}
					walmartTask.Task.DiscordWebhook = event.ProductEvent.DiscordWebhook
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
			}
		}
	}
}

// GetTaskStore returns the singleton instance of the EventBus
func GetTaskStore() *TaskStore {
	return taskStore
}
