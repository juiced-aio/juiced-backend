package stores

import (
	e "errors"

	"backend.juicedbot.io/juiced.infrastructure/common/enums"
)

func (taskStore *TaskStore) SetStopFlag(retailer enums.Retailer, ID string, flag bool) error {
	if baseTask, ok := taskStore.Tasks[ID]; ok {
		taskInfo := baseTask.GetTaskInfo()
		if taskInfo == nil {
			return e.New("task info nil pointer")
		}
		taskInfo.StopFlag = true
	}
	return nil
}
