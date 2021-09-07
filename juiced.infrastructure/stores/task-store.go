package stores

import (
	"fmt"
	"sort"
	"strings"
	"time"

	"backend.juicedbot.io/juiced.infrastructure/database"
	"backend.juicedbot.io/juiced.infrastructure/discord"
	"backend.juicedbot.io/juiced.infrastructure/entities"
	"backend.juicedbot.io/juiced.infrastructure/enums"
	sec "backend.juicedbot.io/juiced.security/auth/util"
	"backend.juicedbot.io/juiced.sitescripts/pokemoncenter"
	"github.com/google/uuid"
)

type TaskStore struct {
	Tasks map[string]*entities.Task
}

var taskStore TaskStore

func InitTaskStore() error {
	taskStore = TaskStore{
		Tasks: make(map[string]*entities.Task),
	}

	tasks, err := database.GetAllTasks()
	if err != nil {
		return err
	}

	for _, task := range tasks {
		task := task
		taskStore.Tasks[task.ID] = &task
	}

	return nil
}

type TaskNotFoundError struct {
	ID string
}

func (e *TaskNotFoundError) Error() string {
	return fmt.Sprintf("Task with ID %s not found", e.ID)
}

func GetAllTasks() []*entities.Task {
	tasks := []*entities.Task{}
	for _, task := range taskStore.Tasks {
		tasks = append(tasks, task)
	}

	sort.SliceStable(tasks, func(i, j int) bool {
		return tasks[i].CreationDate < tasks[j].CreationDate
	})

	return tasks
}

func GetTasks(taskIDs []string) []*entities.Task {
	tasks := []*entities.Task{}
	for _, taskID := range taskIDs {
		if task, ok := taskStore.Tasks[taskID]; ok {
			tasks = append(tasks, task)
		}
	}

	sort.SliceStable(tasks, func(i, j int) bool {
		return tasks[i].CreationDate < tasks[j].CreationDate
	})

	return tasks
}

func CreateTask(task entities.Task) (*entities.Task, error) {
	taskPtr := &task

	if taskPtr.ID == "" {
		taskPtr.ID = uuid.New().String()
	}
	if taskPtr.CreationDate == 0 {
		taskPtr.CreationDate = time.Now().Unix()
	}

	taskGroupPtr, err := GetTaskGroup(task.TaskGroupID)
	if err != nil {
		return nil, err
	}

	profilePtr, err := GetProfile(task.Task.TaskInput.ProfileID)
	if err != nil {
		return nil, err
	}

	var proxyGroupPtr *entities.ProxyGroup
	var proxyPtr *entities.Proxy
	if task.Task.TaskInput.ProxyGroupID != "" {
		proxyGroupPtr, err = GetProxyGroup(task.Task.TaskInput.ProxyGroupID)
		if err != nil {
			return nil, err
		} else {
			proxyPtr = proxyGroupPtr.GetRandomLeastUsedProxy()
		}
	}

	task.Task.Status = enums.TaskIdle
	task.Task.Task = taskPtr
	task.Task.TaskGroup = taskGroupPtr
	task.Task.Profile = profilePtr
	task.Task.ProxyGroup = proxyGroupPtr
	task.Task.Proxy = proxyPtr

	if task.Task.TaskInput.DelayMS <= 0 {
		task.Task.TaskInput.DelayMS = 2000
	}
	if task.Task.TaskInput.Quantity <= 0 {
		task.Task.TaskInput.Quantity = 1
	}

	var retailerTask entities.RetailerTask
	switch task.Retailer {
	case enums.PokemonCenter:
		retailerTask, err = pokemoncenter.CreateTask(task.Task.TaskInput, task.Task)

	}
	if err != nil {
		return nil, err
	}
	task.Task.RetailerTask = &retailerTask

	err = database.CreateTask(*taskPtr)
	if err != nil {
		return nil, err
	}

	taskStore.Tasks[task.ID] = taskPtr

	return taskPtr, nil
}

func GetTask(taskID string) (*entities.Task, error) {
	task, ok := taskStore.Tasks[taskID]
	if !ok {
		return nil, &TaskNotFoundError{taskID}
	}

	return task, nil
}

func RunTask(taskID string) error {
	// TODO

	return nil
}

func RunRetailerTask(task *entities.BaseTask) {
	defer func() {
		if r := recover(); r != nil {
			task.StopFlag = true
			task.PublishEvent(enums.TaskFail, fmt.Sprintf(enums.TaskFailed, r))
		}
		task.Running = false
	}()

	if task.Running {
		return
	}

	task.Running = true
	task.PublishEvent(enums.TaskStart, enums.TaskStarted)

	retailerTask := *task.RetailerTask
	ranSetupFunctions := task.RunFunctions(retailerTask.GetSetupFunctions())
	if !ranSetupFunctions {
		return
	}

	gotProductInfo := task.WaitForMonitor()
	if !gotProductInfo {
		return
	}

	startTime := time.Now().Unix()

	ranMainFunctions := task.RunFunctions(retailerTask.GetMainFunctions())
	if !ranMainFunctions {
		return
	}

	endTime := time.Now().Unix()

	proxy := ""
	if task.Proxy != nil {
		proxy = entities.ProxyCleaner(*task.Proxy)
	}

	profile := ""
	if task.Profile != nil {
		profile = task.Profile.Name
	}

	sec.ProcessCheckout(task, sec.ProcessCheckoutInfo{
		ProductInfo:  task.ProductInfo,
		Quantity:     task.ActualQuantity,
		MsToCheckout: endTime - startTime,
		Success:      strings.Contains(task.Status, enums.CheckedOut),
		Status:       task.Status,
		Content:      "",
		Embeds:       discord.CreateDiscordEmbed(task.Task.Retailer, proxy, profile, task.Status, task.ProductInfo),
		Retailer:     task.Task.Retailer,
	})
}
