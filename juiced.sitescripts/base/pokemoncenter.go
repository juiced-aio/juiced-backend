package base

import (
	"backend.juicedbot.io/juiced.infrastructure/common/entities"
	"backend.juicedbot.io/juiced.infrastructure/common/events"
	"backend.juicedbot.io/juiced.sitescripts/pokemoncenter"
	"backend.juicedbot.io/juiced.sitescripts/util"
)

func CreatePokemonCenterTask(task *entities.Task, profile entities.Profile, proxyGroup *entities.ProxyGroup, eventBus *events.EventBus, input pokemoncenter.TaskInput) (*pokemoncenter.Task, error) {
	return &pokemoncenter.Task{
		TaskInfo: &util.TaskInfo{
			Task:       task,
			Profile:    profile,
			ProxyGroup: proxyGroup,
			EventBus:   eventBus,
		},
		Input: input,
	}, nil
}

func CreatePokemonCenterMonitor(taskGroup *entities.TaskGroup, proxyGroup *entities.ProxyGroup, eventBus *events.EventBus, singleMonitors []entities.PokemonCenterSingleMonitorInfo) (*pokemoncenter.Monitor, error) {
	storedPokemonCenterMonitors := make(map[string]entities.PokemonCenterSingleMonitorInfo)
	skus := []string{}

	for _, monitor := range singleMonitors {
		storedPokemonCenterMonitors[monitor.SKU] = monitor
		skus = append(skus, monitor.SKU)
	}

	return &pokemoncenter.Monitor{
		MonitorInfo: &util.MonitorInfo{
			TaskGroup:  taskGroup,
			ProxyGroup: proxyGroup,
			EventBus:   eventBus,
		},
		SKUs:        skus,
		SKUWithInfo: storedPokemonCenterMonitors,
	}, nil
}
