package stores

import (
	"fmt"
	"strings"
	"time"

	"backend.juicedbot.io/juiced.infrastructure/database"
	"backend.juicedbot.io/juiced.infrastructure/discord"
	"backend.juicedbot.io/juiced.infrastructure/entities"
	"backend.juicedbot.io/juiced.infrastructure/enums"
	"backend.juicedbot.io/juiced.security/auth/util"
)

type TaskStore struct {
	Tasks map[string]*entities.Task
}

var taskStore TaskStore

func (store *TaskStore) Init() error {
	tasks, err := database.GetAllTasks()
	if err != nil {
		return err
	}

	for _, task := range tasks {
		store.Tasks[task.ID] = &task
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

	return tasks
}

func GetTasks(taskIDs []string) []*entities.Task {
	tasks := []*entities.Task{}
	for _, taskID := range taskIDs {
		if task, ok := taskStore.Tasks[taskID]; ok {
			tasks = append(tasks, task)
		}
	}

	return tasks
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

	ranSetupFunctions := task.RunFunctions(task.RetailerTask.GetSetupFunctions())
	if !ranSetupFunctions {
		return
	}

	gotProductInfo := task.WaitForMonitor()
	if !gotProductInfo {
		return
	}

	startTime := time.Now().Unix()

	ranMainFunctions := task.RunFunctions(task.RetailerTask.GetMainFunctions())
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

	util.ProcessCheckout(task, util.ProcessCheckoutInfo{
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
