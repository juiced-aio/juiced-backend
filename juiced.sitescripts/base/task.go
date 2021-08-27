package base

import (
	"errors"
	"fmt"
	"log"
	"strings"
	"time"

	"backend.juicedbot.io/juiced.infrastructure/common/entities"
	"backend.juicedbot.io/juiced.infrastructure/common/enums"
	"backend.juicedbot.io/juiced.infrastructure/common/events"

	"backend.juicedbot.io/juiced.sitescripts/pokemoncenter"
	"backend.juicedbot.io/juiced.sitescripts/util"
)

type Task struct {
	RetailTask
	Retailer enums.Retailer
}

type RetailTask interface {
	GetTaskInfo() *util.TaskInfo
	FillStockInfo(util.StockInfo)
	GetTaskFunctions() []util.TaskFunction
}

func (baseTask *Task) CreateRetailerTask(task *entities.Task, profile entities.Profile, proxyGroup *entities.ProxyGroup, eventBus *events.EventBus, data interface{}) error {
	var err error

	switch baseTask.Retailer {
	case enums.PokemonCenter:
		input, ok := data.(pokemoncenter.TaskInput)
		if !ok {
			return errors.New("bad input")
		}
		err = baseTask.CreatePokemonCenterTask(task, profile, proxyGroup, eventBus, input)
	}

	return err
}

func (task *Task) RunTask() {
	taskInfo := task.GetTaskInfo()

	defer func() {
		if r := recover(); r != nil {
			task.PublishEvent(fmt.Sprintf(enums.TaskFailed, r), enums.TaskFail)
		} else {
			if !strings.Contains(taskInfo.Task.TaskStatus, enums.TaskIdle) &&
				!strings.Contains(taskInfo.Task.TaskStatus, enums.CheckedOut) &&
				!strings.Contains(taskInfo.Task.TaskStatus, enums.CardDeclined) &&
				!strings.Contains(taskInfo.Task.TaskStatus, enums.CheckoutFailed) {
				task.PublishEvent(enums.TaskIdle, enums.TaskStop)
			}
		}
		taskInfo.StopFlag = true
	}()

	if taskInfo.Task.TaskDelay == 0 {
		taskInfo.Task.TaskDelay = 2000
	}
	if taskInfo.Task.TaskQty <= 0 {
		taskInfo.Task.TaskQty = 1
	}

	taskFunctions := task.GetTaskFunctions()

	var success bool
	var status enums.TaskStatus
	for _, function := range taskFunctions {
		needToStop := task.CheckForStop()
		if needToStop {
			return
		}

		if function.StatusBegin != "" {
			task.PublishEvent(function.StatusBegin, enums.TaskUpdate)
		}

		if function.InBackground {
			go function.Function()
		} else if function.RefreshFunction {
			if function.RefreshEvery == 0 {
				function.RefreshEvery = 60
			}
			go task.RefreshWrapper(function)
		} else if function.SpecialFunction {
			success, status = function.Function()
		} else {
			success, status = task.RunUntilSuccessful(function)
		}

		if function.WaitingForMonitor {
			taskInfo.StartTime = time.Now()
		}

		if function.Checkout {
			taskInfo.EndTime = time.Now()

			go util.ProcessCheckout(&util.ProcessCheckoutInfo{
				TaskInfo: taskInfo,
				Success:  success,
				Status:   status,
				Embeds:   CreateDiscordEmbed(task.Retailer, status, taskInfo),
				Content:  "",
				Retailer: task.Retailer,
			})

			log.Println("STARTED AT: " + taskInfo.StartTime.String())
			log.Println("  ENDED AT: " + taskInfo.EndTime.String())
			log.Println("TIME TO CHECK OUT: ", taskInfo.EndTime.Sub(taskInfo.StartTime).Milliseconds())
		}
	}
}
