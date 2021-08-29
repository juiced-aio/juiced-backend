package stores

import (
	"fmt"
	"reflect"
	"time"

	"backend.juicedbot.io/juiced.infrastructure/database"
	"backend.juicedbot.io/juiced.infrastructure/entities"
	"backend.juicedbot.io/juiced.infrastructure/enums"
	"backend.juicedbot.io/juiced.infrastructure/events"
	"backend.juicedbot.io/juiced.sitescripts/pokemoncenter"
	"backend.juicedbot.io/juiced.sitescripts/util"
	"github.com/google/uuid"
)

type TaskGroupStore struct {
	TaskGroups map[string]*entities.TaskGroup
}

var taskGroupStore TaskGroupStore

func (store *TaskGroupStore) Init() error {
	taskGroups, err := database.GetAllTaskGroups()
	if err != nil {
		return err
	}

	for _, taskGroup := range taskGroups {
		taskGroupPtr := &taskGroup
		for _, monitor := range taskGroup.Monitors {
			monitor.Status = enums.MonitorIdle
			monitor.TaskGroup = taskGroupPtr

			if monitor.MonitorInput.ProxyGroupID != "" {
				proxyGroup, err := GetProxyGroup(monitor.MonitorInput.ProxyGroupID)
				if err == nil {
					monitor.ProxyGroup = proxyGroup
				}
			}
		}
		taskGroup.Tasks = GetTasks(taskGroup.TaskIDs)
		store.TaskGroups[taskGroup.GroupID] = taskGroupPtr
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
	taskGroupStore.TaskGroups[taskGroup.GroupID] = taskGroupPtr

	if taskGroupPtr.GroupID == "" {
		taskGroup.GroupID = uuid.New().String()
	}
	if taskGroupPtr.CreationDate == 0 {
		taskGroup.CreationDate = time.Now().Unix()
	}

	for _, monitor := range taskGroupPtr.Monitors {
		monitor.Status = enums.MonitorIdle
		monitor.TaskGroup = taskGroupPtr

		proxyGroup, err := GetProxyGroup(monitor.MonitorInput.ProxyGroupID)
		if err != nil {
			return nil, err
		}

		monitor.ProxyGroup = proxyGroup

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

	err := database.CreateTaskGroup(*taskGroupPtr)

	return taskGroupPtr, err
}

func RunTaskGroup(taskGroupID string) error {
	// TODO

	return nil
}

func RunMonitor(monitor *entities.BaseMonitor) {
	defer func() {
		if r := recover(); r != nil {
			monitor.StopFlag = true
			monitor.PublishEvent(enums.MonitorFail, fmt.Sprintf(enums.MonitorFailed, r))
		}
		monitor.Running = false
	}()

	if monitor.Running {
		return
	}

	monitor.Running = true
	monitor.PublishEvent(enums.MonitorStart, enums.Searching)

	monitorClient, err := util.CreateClient()
	if err != nil {
		panic("could not create HTTP client")
	}
	monitor.Client = &monitorClient

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
			events.GetEventBus().PublishMonitorEvent(enums.SendingProductInfoToTasks, enums.MonitorUpdate, productInfo, monitor.TaskGroup.GroupID)
			monitor.ProductInfo = productInfo
		}

		monitor.Sleep()
	}
}
