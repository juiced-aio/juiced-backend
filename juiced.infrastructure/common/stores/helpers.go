package stores

import (
	e "errors"

	"backend.juicedbot.io/juiced.infrastructure/common/entities"
	"backend.juicedbot.io/juiced.infrastructure/common/enums"
	"backend.juicedbot.io/juiced.infrastructure/common/errors"
)

func (taskStore *TaskStore) SetStopFlag(retailer enums.Retailer, ID string, flag bool) error {
	switch retailer {
	// Future sitescripts will have a case here
	case enums.Amazon:
		if amazonTask, ok := taskStore.AmazonTasks[ID]; ok {
			amazonTask.Task.StopFlag = flag
		}

	case enums.BestBuy:
		if bestbuyTask, ok := taskStore.BestbuyTasks[ID]; ok {
			bestbuyTask.Task.StopFlag = flag
		}

	case enums.BoxLunch:
		if boxlunchTask, ok := taskStore.BoxlunchTasks[ID]; ok {
			boxlunchTask.Task.StopFlag = flag
		}

	case enums.Disney:
		if disneyTask, ok := taskStore.DisneyTasks[ID]; ok {
			disneyTask.Task.StopFlag = flag
		}

	case enums.GameStop:
		if gamestopTask, ok := taskStore.GamestopTasks[ID]; ok {
			gamestopTask.Task.StopFlag = flag
		}

	case enums.HotTopic:
		if hottopicTask, ok := taskStore.HottopicTasks[ID]; ok {
			hottopicTask.Task.StopFlag = flag
		}

	case enums.Newegg:
		if neweggTask, ok := taskStore.NeweggTasks[ID]; ok {
			neweggTask.Task.StopFlag = flag
		}

	case enums.PokemonCenter:
		if pokemonCenterTask, ok := taskStore.PokemonCenterTasks[ID]; ok {
			pokemonCenterTask.Task.StopFlag = flag
		}

	case enums.Shopify:
		if shopifyTask, ok := taskStore.ShopifyTasks[ID]; ok {
			shopifyTask.Task.StopFlag = flag
		}

	case enums.Target:
		if targetTask, ok := taskStore.TargetTasks[ID]; ok {
			targetTask.Task.StopFlag = flag
		}

	case enums.Topps:
		if toppsTask, ok := taskStore.ToppsTasks[ID]; ok {
			toppsTask.Task.StopFlag = flag
		}

	case enums.Walmart:
		if walmartTask, ok := taskStore.WalmartTasks[ID]; ok {
			walmartTask.Task.StopFlag = flag
		}

	default:
		return e.New(errors.InvalidTaskRetailerError)

	}

	return nil
}

func (taskStore *TaskStore) GetTask(retailer enums.Retailer, ID string) *entities.Task {
	switch retailer {
	// Future sitescripts will have a case here
	case enums.Amazon:
		if amazonTask, ok := taskStore.AmazonTasks[ID]; ok {
			return amazonTask.Task.Task
		}

	case enums.BestBuy:
		if bestbuyTask, ok := taskStore.BestbuyTasks[ID]; ok {
			return bestbuyTask.Task.Task
		}

	case enums.BoxLunch:
		if boxlunchTask, ok := taskStore.BoxlunchTasks[ID]; ok {
			return boxlunchTask.Task.Task
		}

	case enums.Disney:
		if disneyTask, ok := taskStore.DisneyTasks[ID]; ok {
			return disneyTask.Task.Task
		}

	case enums.GameStop:
		if gamestopTask, ok := taskStore.GamestopTasks[ID]; ok {
			return gamestopTask.Task.Task
		}

	case enums.HotTopic:
		if hottopicTask, ok := taskStore.HottopicTasks[ID]; ok {
			return hottopicTask.Task.Task
		}

	case enums.Newegg:
		if neweggTask, ok := taskStore.NeweggTasks[ID]; ok {
			return neweggTask.Task.Task
		}

	case enums.PokemonCenter:
		if pokemonCenterTask, ok := taskStore.PokemonCenterTasks[ID]; ok {
			return pokemonCenterTask.Task.Task
		}

	case enums.Shopify:
		if shopifyTask, ok := taskStore.ShopifyTasks[ID]; ok {
			return shopifyTask.Task.Task
		}

	case enums.Target:
		if targetTask, ok := taskStore.TargetTasks[ID]; ok {
			return targetTask.Task.Task
		}

	case enums.Topps:
		if toppsTask, ok := taskStore.ToppsTasks[ID]; ok {
			return toppsTask.Task.Task
		}

	case enums.Walmart:
		if walmartTask, ok := taskStore.WalmartTasks[ID]; ok {
			return walmartTask.Task.Task
		}
	}

	return nil
}

func (monitorStore *MonitorStore) GetMonitor(retailer enums.Retailer, ID string) *entities.TaskGroup {
	switch retailer {
	// Future sitescripts will have a case here
	case enums.Amazon:
		if amazonMonitor, ok := monitorStore.AmazonMonitors[ID]; ok {
			return amazonMonitor.Monitor.TaskGroup
		}

	case enums.BestBuy:
		if bestbuyMonitor, ok := monitorStore.BestbuyMonitors[ID]; ok {
			return bestbuyMonitor.Monitor.TaskGroup
		}

	case enums.BoxLunch:
		if boxlunchMonitor, ok := monitorStore.BoxlunchMonitors[ID]; ok {
			return boxlunchMonitor.Monitor.TaskGroup
		}

	case enums.Disney:
		if disneyMonitor, ok := monitorStore.DisneyMonitors[ID]; ok {
			return disneyMonitor.Monitor.TaskGroup
		}

	case enums.GameStop:
		if gamestopMonitor, ok := monitorStore.GamestopMonitors[ID]; ok {
			return gamestopMonitor.Monitor.TaskGroup
		}

	case enums.HotTopic:
		if hottopicMonitor, ok := monitorStore.HottopicMonitors[ID]; ok {
			return hottopicMonitor.Monitor.TaskGroup
		}

	case enums.Newegg:
		if neweggMonitor, ok := monitorStore.NeweggMonitors[ID]; ok {
			return neweggMonitor.Monitor.TaskGroup
		}

	case enums.PokemonCenter:
		if pokemoncenterMonitor, ok := monitorStore.PokemonCenterMonitors[ID]; ok {
			return pokemoncenterMonitor.Monitor.TaskGroup
		}

	case enums.Shopify:
		if shopifyMonitor, ok := monitorStore.ShopifyMonitors[ID]; ok {
			return shopifyMonitor.Monitor.TaskGroup
		}

	case enums.Target:
		if targetMonitor, ok := monitorStore.TargetMonitors[ID]; ok {
			return targetMonitor.Monitor.TaskGroup
		}

	case enums.Topps:
		if toppsMonitor, ok := monitorStore.ToppsMonitors[ID]; ok {
			return toppsMonitor.Monitor.TaskGroup
		}

	case enums.Walmart:
		if walmartMonitor, ok := monitorStore.WalmartMonitors[ID]; ok {
			return walmartMonitor.Monitor.TaskGroup
		}
	}

	return nil
}
