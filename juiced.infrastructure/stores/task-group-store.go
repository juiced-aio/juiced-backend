package stores

import (
	"fmt"
	"reflect"
	"sort"
	"strings"
	"time"

	"backend.juicedbot.io/juiced.antibot/cloudflare"
	"backend.juicedbot.io/juiced.infrastructure/database"
	"backend.juicedbot.io/juiced.infrastructure/entities"
	"backend.juicedbot.io/juiced.infrastructure/enums"
	"backend.juicedbot.io/juiced.infrastructure/events"
	u "backend.juicedbot.io/juiced.infrastructure/util"
	"backend.juicedbot.io/juiced.sitescripts/pokemoncenter"
	"backend.juicedbot.io/juiced.sitescripts/util"
	"github.com/google/uuid"
)

type TaskGroupStore struct {
	TaskGroups map[string]*entities.TaskGroup
}

var taskGroupStore TaskGroupStore

func InitTaskGroupStore() error {
	taskGroupStore = TaskGroupStore{
		TaskGroups: make(map[string]*entities.TaskGroup),
	}

	taskGroups, err := database.GetAllTaskGroups()
	if err != nil {
		return err
	}

	for _, taskGroup := range taskGroups {
		taskGroup := taskGroup
		taskGroupPtr := &taskGroup
		skip := false
		for _, monitor := range taskGroupPtr.Monitors {
			monitor.Status = enums.MonitorIdle
			monitor.TaskGroup = taskGroupPtr

			if monitor.MonitorInput.ProxyGroupID != "" {
				proxyGroup, err := GetProxyGroup(monitor.MonitorInput.ProxyGroupID)
				if err == nil {
					monitor.ProxyGroup = proxyGroup
				}
			}

			var retailerMonitor entities.Monitor
			switch taskGroup.Retailer {
			case enums.PokemonCenter:
				retailerMonitor, err = pokemoncenter.CreateMonitor(monitor.MonitorInput, monitor)

			}
			if err != nil {
				skip = true
			}
			monitor.Monitor = &retailerMonitor
		}

		if !skip {
			taskGroupPtr.Tasks = GetTasks(taskGroupPtr.TaskIDs)
			for _, task := range taskGroupPtr.Tasks {
				task.Task.TaskGroup = taskGroupPtr
			}
			taskGroupStore.TaskGroups[taskGroupPtr.GroupID] = taskGroupPtr
		}
	}

	return nil
}

type TaskGroupNotFoundError struct {
	ID string
}

func (e *TaskGroupNotFoundError) Error() string {
	return fmt.Sprintf("TaskGroup with ID %s not found", e.ID)
}

func GetAllTaskGroups() []*entities.TaskGroup {
	taskGroups := []*entities.TaskGroup{}

	for _, taskGroup := range taskGroupStore.TaskGroups {
		taskGroups = append(taskGroups, taskGroup)
	}

	sort.SliceStable(taskGroups, func(i, j int) bool {
		return taskGroups[i].CreationDate < taskGroups[j].CreationDate
	})

	return taskGroups
}

func GetTaskGroup(groupID string) (*entities.TaskGroup, error) {
	taskGroup, ok := taskGroupStore.TaskGroups[groupID]
	if !ok {
		return nil, &TaskGroupNotFoundError{groupID}
	}

	return taskGroup, nil
}

func CreateTaskGroup(taskGroup entities.TaskGroup) (*entities.TaskGroup, error) {
	taskGroupPtr := &taskGroup
	var err error

	if taskGroupPtr.GroupID == "" {
		taskGroup.GroupID = uuid.New().String()
	}
	if taskGroupPtr.CreationDate == 0 {
		taskGroup.CreationDate = time.Now().Unix()
	}

	for _, monitor := range taskGroupPtr.Monitors {
		if monitor.MonitorInput.DelayMS <= 0 {
			monitor.MonitorInput.DelayMS = 2000
		}
		if monitor.MonitorInput.MaxPrice <= 0 {
			monitor.MonitorInput.MaxPrice = -1
		}

		monitor.MonitorID = uuid.New().String()
		monitor.Status = enums.MonitorIdle
		monitor.TaskGroup = taskGroupPtr

		var proxyGroupPtr *entities.ProxyGroup
		if monitor.MonitorInput.ProxyGroupID != "" {
			proxyGroupPtr, err = GetProxyGroup(monitor.MonitorInput.ProxyGroupID)
			if err != nil {
				return nil, err
			}
		}

		monitor.ProxyGroup = proxyGroupPtr

		var retailerMonitor entities.Monitor
		switch taskGroup.Retailer {
		case enums.PokemonCenter:
			retailerMonitor, err = pokemoncenter.CreateMonitor(monitor.MonitorInput, monitor)

		}
		if err != nil {
			return nil, err
		}
		monitor.Monitor = &retailerMonitor
	}

	err = database.CreateTaskGroup(*taskGroupPtr)
	if err != nil {
		return nil, err
	}

	taskGroupStore.TaskGroups[taskGroup.GroupID] = taskGroupPtr
	return taskGroupPtr, nil
}

func UpdateTaskGroup(groupID string, newTaskGroup entities.TaskGroup) (*entities.TaskGroup, error) {
	taskGroup, err := GetTaskGroup(groupID)
	if err != nil {
		return nil, err
	}

	taskGroup.Name = newTaskGroup.Name

	// TODO

	return taskGroup, database.UpdateTaskGroup(groupID, *taskGroup)
}

func RemoveTaskGroup(groupID string) (entities.TaskGroup, error) {
	taskGroup, err := GetTaskGroup(groupID)
	if err != nil {
		return entities.TaskGroup{}, err
	}

	err = StopTaskGroup(groupID, true)
	if err != nil {
		return entities.TaskGroup{}, err
	}

	delete(taskGroupStore.TaskGroups, groupID)

	return *taskGroup, database.RemoveTaskGroup(groupID, true)
}

func CloneTaskGroup(groupID string) (*entities.TaskGroup, error) {
	taskGroupPtr, err := GetTaskGroup(groupID)
	if err != nil {
		return nil, err
	}

	newTaskGroup := *taskGroupPtr
	newTaskGroup.Name += " " + u.RandString(3)
	newTaskGroup.GroupID = ""
	newTaskGroup.CreationDate = 0

	newMonitors := []*entities.BaseMonitor{}
	var err_ error
	for _, monitor := range newTaskGroup.Monitors {
		newMonitor := *monitor
		newMonitor.MonitorID = uuid.New().String()
		newMonitor.Status = enums.MonitorIdle
		newMonitor.Running = false
		newMonitor.ProductInfo = entities.ProductInfo{}

		newMonitor.TaskGroup = &newTaskGroup
		newMonitor.Proxy = nil
		newMonitor.Client = nil
		newMonitor.Scraper = nil
		newMonitor.StopFlag = false

		var retailerMonitor entities.Monitor
		switch newTaskGroup.Retailer {
		case enums.PokemonCenter:
			retailerMonitor, err = pokemoncenter.CreateMonitor(newMonitor.MonitorInput, &newMonitor)

		}
		if err == nil {
			newMonitor.Monitor = &retailerMonitor
			newMonitors = append(newMonitors, &newMonitor)
		} else {
			if err_ == nil {
				err_ = err
			}
		}
	}

	if len(newMonitors) == 0 {
		return nil, err_
	}

	newTaskGroup.Monitors = newMonitors

	newTaskGroupPtr, err := CreateTaskGroup(newTaskGroup)
	if err != nil {
		return nil, err
	}

	newTasks := []*entities.Task{}
	newTaskIDs := []string{}
	for _, taskID := range taskGroupPtr.TaskIDs {
		var newTask *entities.Task
		newTask, err = CloneTask(taskID, newTaskGroupPtr.GroupID)
		if err == nil {
			newTasks = append(newTasks, newTask)
			newTaskIDs = append(newTaskIDs, newTask.ID)
		}
	}

	newTaskGroupPtr.TaskIDs = newTaskIDs
	newTaskGroupPtr.Tasks = newTasks

	return UpdateTaskGroup(newTaskGroupPtr.GroupID, *newTaskGroupPtr)
}

func AddTasksToGroup(groupID string, tasks []*entities.Task) (*entities.TaskGroup, error) {
	taskGroupPtr, err := GetTaskGroup(groupID)
	if err != nil {
		return nil, err
	}

	for _, task := range tasks {
		taskGroupPtr.TaskIDs = append(taskGroupPtr.TaskIDs, task.ID)
		taskGroupPtr.Tasks = append(taskGroupPtr.Tasks, task)
	}

	return UpdateTaskGroup(groupID, *taskGroupPtr)
}

func RemoveTasksFromGroup(groupID string, taskIDs []string) (*entities.TaskGroup, error) {
	taskGroupPtr, err := GetTaskGroup(groupID)
	if err != nil {
		return nil, err
	}

	newTasks := []*entities.Task{}
	newTaskIDs := []string{}
	for _, task := range taskGroupPtr.Tasks {
		if !u.InSlice(taskIDs, task.ID) {
			newTasks = append(newTasks, task)
			newTaskIDs = append(newTaskIDs, task.ID)
		}
	}
	taskGroupPtr.TaskIDs = newTaskIDs
	taskGroupPtr.TaskIDsSerialized = strings.Join(newTaskIDs, ",")
	taskGroupPtr.Tasks = newTasks

	return UpdateTaskGroup(groupID, *taskGroupPtr)
}

func StartTaskGroup(groupID string, startTasks bool) error {
	taskGroupPtr, err := GetTaskGroup(groupID)
	if err != nil {
		return err
	}

	for _, monitor := range taskGroupPtr.Monitors {
		go RunMonitor(monitor)
	}

	if startTasks {
		for _, taskID := range taskGroupPtr.TaskIDs {
			err_ := StartTask(taskID)
			if err == nil {
				err = err_
			}
		}
	}

	return err
}

func CheckForTaskGroupStart(groupID string) {
	taskGroupPtr, err := GetTaskGroup(groupID)
	if err == nil {
		runningTasks := TasksRunning(taskGroupPtr)

		if runningTasks {
			StartTaskGroup(groupID, false)
		}
	}
}

func StopTaskGroup(groupID string, stopTasks bool) error {
	taskGroupPtr, err := GetTaskGroup(groupID)
	if err != nil {
		return err
	}

	for _, monitor := range taskGroupPtr.Monitors {
		monitor.StopFlag = true
		monitor.Status = enums.MonitorIdle
		monitor.Running = false
		monitor.ProductInfo = entities.ProductInfo{}

		monitor.Proxy = nil
		monitor.Client = nil
		monitor.Scraper = nil
	}

	for _, taskID := range taskGroupPtr.TaskIDs {
		err_ := StopTask(taskID)
		if err == nil {
			err = err_
		}
	}

	return err
}

func CheckForTaskGroupStop(groupID string) {
	taskGroupPtr, err := GetTaskGroup(groupID)
	runningTasks := false
	if err == nil {
		for _, task := range taskGroupPtr.Tasks {
			if task.Task.Running && !task.Task.StopFlag {
				runningTasks = true
			}
		}
	}

	if !runningTasks {
		StopTaskGroup(groupID, false)
	}
}

func TasksRunning(taskGroup *entities.TaskGroup) bool {
	runningTasks := false
	for _, task := range taskGroup.Tasks {
		if task.Task.Running && !task.Task.StopFlag {
			runningTasks = true
		}
	}
	return runningTasks
}

func RunMonitor(monitor *entities.BaseMonitor) {
	defer func() {
		if r := recover(); r != nil {
			monitor.StopFlag = true
			monitor.PublishEvent(enums.MonitorFail, fmt.Sprintf(enums.MonitorFailed, r), nil)
		}
		monitor.Running = false
	}()

	if monitor.Running {
		return
	}

	monitor.StopFlag = false
	monitor.Running = true
	monitor.PublishEvent(enums.MonitorStart, enums.Searching, nil)

	monitorClient, err := util.CreateClient()
	if err != nil {
		panic("could not create HTTP client")
	}
	monitor.Client = &monitorClient
	monitorScraper := cloudflare.Init(monitorClient, u.HAWK_KEY, false)
	monitor.Scraper = &monitorScraper

	for {
		needToStop := monitor.CheckForStop()
		if needToStop {
			return
		}

		if monitor.ProxyGroup != nil {
			proxy := monitor.ProxyGroup.GetRandomLeastUsedProxy()
			if proxy != nil {
				monitor.UpdateProxy(proxy)
			}
		}

		retailerMonitor := *monitor.Monitor
		productInfo, err := retailerMonitor.GetProductInfo()
		if err == nil && productInfo.SKU != "" && !reflect.DeepEqual(monitor.ProductInfo, productInfo) {
			events.GetEventBus().PublishMonitorEvent(enums.SendingProductInfoToTasks, enums.MonitorUpdate, productInfo, monitor.TaskGroup.GroupID, monitor.MonitorID)
			monitor.ProductInfo = productInfo
		}

		monitor.Sleep()
	}
}
