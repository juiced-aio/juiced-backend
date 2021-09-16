package hottopic

import (
	"time"

	"backend.juicedbot.io/juiced.infrastructure/entities"
	"backend.juicedbot.io/juiced.infrastructure/enums"
)

const MAX_RETRIES = 5

func CreateTask(input entities.TaskInput, baseTask *entities.BaseTask) (entities.RetailerTask, error) {
	// err := ValidateTaskInput(input.SiteSpecificInfo)
	// if err != nil {
	// 	return nil, err
	// }
	return &Task{
		Input:    input,
		BaseTask: baseTask,
	}, nil
}

func (task *Task) GetSetupFunctions() []entities.TaskFunction {
	setupTaskFunctions := []entities.TaskFunction{
		{
			Function:    task.Setup,
			StatusBegin: enums.SettingUp,
			MaxRetries:  MAX_RETRIES,
		},
	}
	return setupTaskFunctions
}

func (task *Task) GetMainFunctions() []entities.TaskFunction {
	mainTaskFunctions := []entities.TaskFunction{}
	return mainTaskFunctions
}

func (task *Task) Setup() (bool, string) {
	for {
		time.Sleep(1 * time.Second)
	}

	return false, ""
}
