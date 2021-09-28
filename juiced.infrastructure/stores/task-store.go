package stores

import (
	"fmt"
	"log"
	"runtime/debug"
	"sort"
	"strings"
	"time"

	"backend.juicedbot.io/juiced.antibot/cloudflare"
	"backend.juicedbot.io/juiced.infrastructure/database"
	"backend.juicedbot.io/juiced.infrastructure/discord"
	"backend.juicedbot.io/juiced.infrastructure/entities"
	"backend.juicedbot.io/juiced.infrastructure/enums"
	u "backend.juicedbot.io/juiced.infrastructure/util"
	sec "backend.juicedbot.io/juiced.security/auth/util"

	"backend.juicedbot.io/juiced.sitescripts/hottopic"
	"backend.juicedbot.io/juiced.sitescripts/pokemoncenter"
	"backend.juicedbot.io/juiced.sitescripts/util"

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

	var err_ error
	for _, task := range tasks {
		task := task
		taskPtr := &task

		taskPtr.Task.Status = enums.TaskIdle
		taskPtr.Task.StatusPercentage = 0
		taskPtr.Task.Task = taskPtr

		profilePtr, err := GetProfile(taskPtr.Task.TaskInput.ProfileID)
		if err == nil {
			taskPtr.Task.Profile = profilePtr

			if taskPtr.Task.TaskInput.ProxyGroupID != "" {
				proxyGroup, err := GetProxyGroup(taskPtr.Task.TaskInput.ProxyGroupID)
				if err == nil {
					taskPtr.Task.ProxyGroup = proxyGroup
				}
			}

			var retailerTask entities.RetailerTask
			switch task.Retailer {
			case enums.HotTopic:
				retailerTask, err = hottopic.CreateTask(task.Task.TaskInput, task.Task)
			case enums.PokemonCenter:
				retailerTask, err = pokemoncenter.CreateTask(task.Task.TaskInput, task.Task)

			}
			if err != nil {
				if err_ == nil {
					err_ = err
				}
			} else {
				task.Task.RetailerTask = &retailerTask
			}

			taskStore.Tasks[task.ID] = taskPtr
		}
	}

	return err_
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

func GetTask(taskID string) (*entities.Task, error) {
	task, ok := taskStore.Tasks[taskID]
	if !ok {
		return nil, &TaskNotFoundError{taskID}
	}

	return task, nil
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
	if task.Task.TaskInput.ProxyGroupID != "" {
		proxyGroupPtr, err = GetProxyGroup(task.Task.TaskInput.ProxyGroupID)
		if err != nil {
			return nil, err
		}
	}

	task.Task.Status = enums.TaskIdle
	task.Task.StatusPercentage = 0
	task.Task.Task = taskPtr
	task.Task.TaskGroup = taskGroupPtr
	task.Task.Profile = profilePtr
	task.Task.ProxyGroup = proxyGroupPtr

	if task.Task.TaskInput.DelayMS <= 0 {
		task.Task.TaskInput.DelayMS = 2000
	}
	if task.Task.TaskInput.Quantity <= 0 {
		task.Task.TaskInput.Quantity = 1
	}

	var retailerTask entities.RetailerTask
	switch task.Retailer {
	case enums.HotTopic:
		retailerTask, err = hottopic.CreateTask(task.Task.TaskInput, task.Task)
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

func UpdateTask(taskID string, newTask entities.Task) (*entities.Task, error) {
	task, err := GetTask(taskID)
	if err != nil {
		return nil, err
	}

	// TODO

	return task, database.UpdateTask(taskID, *task)
}

func RemoveTask(taskID string) (entities.Task, error) {
	task, err := GetTask(taskID)
	if err != nil {
		return entities.Task{}, err
	}

	err = StopTask(taskID)
	if err != nil {
		return entities.Task{}, err
	}

	err = database.RemoveTask(taskID)
	if err != nil {
		return entities.Task{}, err
	}

	delete(taskStore.Tasks, taskID)
	_, err = RemoveTasksFromGroup(task.TaskGroupID, []string{taskID})

	return *task, err
}

func CloneTask(taskID, taskGroupID string) (*entities.Task, error) {
	taskPtr, err := GetTask(taskID)
	if err != nil {
		return nil, err
	}

	newTask := *taskPtr
	newTask.ID = ""
	newTask.CreationDate = 0
	if taskGroupID != "" {
		newTask.TaskGroupID = taskGroupID
	}

	newBaseTask := *taskPtr.Task
	newBaseTask.Status = enums.TaskIdle
	newBaseTask.StatusPercentage = 0
	newBaseTask.Running = false
	newBaseTask.ProductInfo = entities.ProductInfo{}
	newBaseTask.ActualQuantity = 0

	taskGroupPtr, err := GetTaskGroup(newTask.TaskGroupID)
	if err != nil {
		return nil, err
	}

	profilePtr, err := GetProfile(newBaseTask.TaskInput.ProfileID)
	if err != nil {
		return nil, err
	}

	var proxyGroupPtr *entities.ProxyGroup
	if newBaseTask.TaskInput.ProxyGroupID != "" {
		proxyGroupPtr, err = GetProxyGroup(newBaseTask.TaskInput.ProxyGroupID)
		if err != nil {
			return nil, err
		}
	}

	newBaseTask.TaskGroup = taskGroupPtr
	newBaseTask.Profile = profilePtr
	newBaseTask.ProxyGroup = proxyGroupPtr

	newBaseTask.Client = nil
	newBaseTask.Scraper = nil
	newBaseTask.StopFlag = false

	newBaseTaskPtr := &newBaseTask

	var retailerTask entities.RetailerTask
	switch newTask.Retailer {
	case enums.HotTopic:
		retailerTask, err = hottopic.CreateTask(newBaseTask.TaskInput, newBaseTaskPtr)
	case enums.PokemonCenter:
		retailerTask, err = pokemoncenter.CreateTask(newBaseTask.TaskInput, newBaseTaskPtr)

	}
	if err != nil {
		return nil, err
	}
	newBaseTask.RetailerTask = &retailerTask

	newTask.Task = newBaseTaskPtr

	newTaskPtr, err := CreateTask(newTask)
	if err != nil {
		return nil, err
	}

	newTaskPtr.Task.Task = newTaskPtr

	if taskGroupID == "" {
		_, err = AddTasksToGroup(newTaskPtr.TaskGroupID, []*entities.Task{newTaskPtr})
		if err != nil {
			return nil, err
		}
	}

	return newTaskPtr, nil
}

func StartTask(taskID string) error {
	task, err := GetTask(taskID)
	if err != nil {
		return err
	}

	go RunRetailerTask(task.Task)

	return nil
}

func StopTask(taskID string) error {
	task, err := GetTask(taskID)
	if err != nil {
		return err
	}

	task.Task.StopFlag = true
	task.Task.Status = enums.TaskIdle
	task.Task.StatusPercentage = 0
	task.Task.Running = false
	task.Task.ProductInfo = entities.ProductInfo{}
	task.Task.ActualQuantity = 0

	task.Task.Proxy = nil
	task.Task.Client = nil
	task.Task.Scraper = nil

	return nil
}

func RunRetailerTask(task *entities.BaseTask) {
	defer func() {
		if r := recover(); r != nil {
			log.Println(r)
			log.Println(string(debug.Stack()))
			task.StopFlag = true
			task.PublishEvent(enums.TaskFail, 0, fmt.Sprintf(enums.TaskFailed, r))
		}
		task.Running = false
		CheckForTaskGroupStop(task.Task.TaskGroupID)
	}()

	if task.Running {
		return
	}

	if task.ProxyGroup != nil {
		task.Proxy = task.ProxyGroup.GetRandomLeastUsedProxy()
	}

	task.StopFlag = false
	task.Running = true

	CheckForTaskGroupStart(task.Task.TaskGroupID)

	taskClient, err := util.CreateClient(task.Proxy)
	if err != nil {
		panic("could not create HTTP client")
	}
	task.Client = &taskClient
	taskScraper := cloudflare.Init(taskClient, u.HAWK_KEY, false)
	task.Scraper = &taskScraper
	task.PublishEvent(enums.TaskStarted, 5, enums.TaskStart)

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
