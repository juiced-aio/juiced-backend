package stores

import (
	"backend.juicedbot.io/juiced.infrastructure/common/enums"
)

func (taskStore *TaskStore) SetStopFlag(retailer enums.Retailer, ID string, flag bool) bool {
	switch retailer {
	// Future sitescripts will have a case here
	case enums.Amazon:
		if amazonTask, ok := taskStore.AmazonTasks[ID]; ok {
			amazonTask.Task.StopFlag = flag
		}
		return true

	case enums.BestBuy:
		if bestbuyTask, ok := taskStore.BestbuyTasks[ID]; ok {
			bestbuyTask.Task.StopFlag = flag
		}
		return true

	case enums.GameStop:
		if gamestopTask, ok := taskStore.GamestopTasks[ID]; ok {
			gamestopTask.Task.StopFlag = flag
		}
		return true

	case enums.BoxLunch:
		if boxlunchTask, ok := taskStore.BoxlunchTasks[ID]; ok {
			boxlunchTask.Task.StopFlag = flag
		}
		return true

	case enums.HotTopic:
		if hottopicTask, ok := taskStore.HottopicTasks[ID]; ok {
			hottopicTask.Task.StopFlag = flag
		}
		return true

	case enums.Shopify:
		if shopifyTask, ok := taskStore.ShopifyTasks[ID]; ok {
			shopifyTask.Task.StopFlag = flag
		}
		return true

	case enums.Target:
		if targetTask, ok := taskStore.TargetTasks[ID]; ok {
			targetTask.Task.StopFlag = flag
		}
		return true

	case enums.Walmart:
		if walmartTask, ok := taskStore.WalmartTasks[ID]; ok {
			walmartTask.Task.StopFlag = flag
		}
		return true

	}

	return true
}
