package base

import (
	"backend.juicedbot.io/juiced.infrastructure/common/entities"
	"backend.juicedbot.io/juiced.infrastructure/common/events"
	"backend.juicedbot.io/juiced.sitescripts/pokemoncenter"
	"backend.juicedbot.io/juiced.sitescripts/util"
)

func (baseTask *Task) CreatePokemonCenterTask(task *entities.Task, profile entities.Profile, proxyGroup *entities.ProxyGroup, eventBus *events.EventBus, input pokemoncenter.TaskInput) error {
	baseTask.RetailTask = &pokemoncenter.Task{
		TaskInfo: &util.TaskInfo{
			Task:       task,
			Profile:    profile,
			ProxyGroup: proxyGroup,
			EventBus:   eventBus,
		},
		Input: input,
	}

	return nil
}

func (baseMonitor *Monitor) CreatePokemonCenterMonitor(taskGroup *entities.TaskGroup, proxyGroup *entities.ProxyGroup, eventBus *events.EventBus, singleMonitors []entities.PokemonCenterSingleMonitorInfo) error {
	storedPokemonCenterMonitors := make(map[string]entities.PokemonCenterSingleMonitorInfo)
	skus := []string{}

	for _, monitor := range singleMonitors {
		storedPokemonCenterMonitors[monitor.SKU] = monitor
		skus = append(skus, monitor.SKU)
	}

	baseMonitor.RetailMonitor = &pokemoncenter.Monitor{
		MonitorInfo: &util.MonitorInfo{
			TaskGroup:  taskGroup,
			ProxyGroup: proxyGroup,
			EventBus:   eventBus,
		},
		SKUs:        skus,
		SKUWithInfo: storedPokemonCenterMonitors,
	}

	return nil
}
