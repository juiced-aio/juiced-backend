package bigcartel

import (
	"fmt"
	"time"

	"backend.juicedbot.io/juiced.infrastructure/common/enums"
)

//Improves readability on RunTask
func (task *Task) RunUntilSuccessful(runTaskResult bool, status string) (bool, bool) {
	needToStop := task.CheckForStop()
	// -1 retry = unlimited amount of retries.
	if needToStop || task.Retry > task.MaxRetry {
		task.Task.StopFlag = true //if retry is over the limit we want to set our stop flag.
		return true, true
	}
	if !runTaskResult { //We have failed the task
		if status != "" { //Check if we need to publish event
			task.PublishEvent(fmt.Sprint(status, " Retry: ", task.Retry), enums.TaskUpdate) //if failure then also send back retry number
		}
		if task.Retry >= 0 {
			task.Retry++ //increment our retry.
		}
		time.Sleep(time.Duration(task.Task.Task.TaskDelay) * time.Millisecond)
		return false, false
	} else { //We have completed the task succesfully
		if status != "" { //Check if we need to publish event
			task.PublishEvent(status, enums.TaskUpdate) //If success then just publish the status
		}
		//If we want we can reset retry amount here, depending on if we want the retries global or per task.
		task.Retry = 0
	}

	return true, false //Finally if we have reached this point then task was a success.
}
