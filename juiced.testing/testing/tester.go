package testing

import (
	e "errors"

	"backend.juicedbot.io/juiced.infrastructure/common"
	"backend.juicedbot.io/juiced.infrastructure/common/entities"
	"backend.juicedbot.io/juiced.infrastructure/common/enums"
	"backend.juicedbot.io/juiced.infrastructure/common/errors"
	"backend.juicedbot.io/juiced.infrastructure/common/stores"
)

var taskStore *stores.TaskStore

var monitorStore *stores.MonitorStore

// TestDriver
func TestDriver(task *entities.Task, profile entities.Profile, taskGroup entities.TaskGroup) error {
	if !common.ValidCardType([]byte(profile.CreditCard.CardNumber), task.TaskRetailer) {
		return e.New(errors.StartTaskInvalidCardError + task.TaskRetailer)
	}

	// Start the task's TaskGroup (if it's already running, this will return true)
	err := monitorStore.StartTestMonitor(&taskGroup)
	if err != nil {
		return err
	}

	// Add task to store (if it already exists, this will return true)
	err = taskStore.AddTestTaskToStore(task, profile)
	if err != nil {
		return err
	}

	// If the Task is already running, then we're all set already
	if task.TaskStatus != enums.TaskIdle &&
		task.TaskStatus != enums.CheckedOut &&
		task.TaskStatus != enums.CheckoutFailed {
		return nil
	}

	// Set the task's StopFlag to true before running the task
	taskStore.SetStopFlag(task.TaskRetailer, task.ID, false)

	// Otherwise, start the Task
	taskStore.RunTask(task.TaskRetailer, task.ID)
	return nil
}
