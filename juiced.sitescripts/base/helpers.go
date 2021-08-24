package base

import (
	"errors"

	"backend.juicedbot.io/juiced.infrastructure/common/entities"
	"backend.juicedbot.io/juiced.infrastructure/common/enums"
	"backend.juicedbot.io/juiced.infrastructure/common/events"
	"backend.juicedbot.io/juiced.sitescripts/pokemoncenter"
	"backend.juicedbot.io/juiced.sitescripts/util"
)

func CreateRetailerTask(retailer enums.Retailer, task *entities.Task, profile entities.Profile, proxyGroup *entities.ProxyGroup, eventBus *events.EventBus, data interface{}) (Task, error) {
	task_ := Task{
		Retailer: retailer,
	}
	var err error

	switch retailer {
	case enums.PokemonCenter:
		input, ok := data.(pokemoncenter.TaskInput)
		if !ok {
			return task_, errors.New("bad input")
		}
		task_.PokemonCenterTask, err = CreatePokemonCenterTask(task, profile, proxyGroup, eventBus, input)
	}

	return task_, err
}

func (task *Task) GetTaskInfo() *util.TaskInfo {
	switch task.Retailer {
	case enums.PokemonCenter:
		if task.PokemonCenterTask == nil {
			return nil
		}
		return task.PokemonCenterTask.TaskInfo
	}

	return nil
}

func (task *Task) RunTask() {
	switch task.Retailer {
	case enums.PokemonCenter:
		task.PokemonCenterTask.RunTask()
	}
}

func CreateRetailerMonitor(retailer enums.Retailer, taskGroup *entities.TaskGroup, proxyGroup *entities.ProxyGroup, eventBus *events.EventBus, data interface{}) (Monitor, error) {
	monitor := Monitor{
		Retailer: retailer,
	}
	var err error

	switch retailer {
	case enums.PokemonCenter:
		monitors, ok := data.([]entities.PokemonCenterSingleMonitorInfo)
		if !ok {
			return monitor, errors.New("bad input")
		}
		monitor.PokemonCenterMonitor, err = CreatePokemonCenterMonitor(taskGroup, proxyGroup, eventBus, monitors)
	}

	return monitor, err
}

func (monitor *Monitor) GetMonitorInfo() *util.MonitorInfo {
	switch monitor.Retailer {
	case enums.PokemonCenter:
		if monitor.PokemonCenterMonitor == nil {
			return nil
		}
		return monitor.PokemonCenterMonitor.MonitorInfo
	}

	return nil
}

func (monitor *Monitor) RunMonitor() {
	switch monitor.Retailer {
	case enums.PokemonCenter:
		monitor.PokemonCenterMonitor.RunMonitor()
	}
}
