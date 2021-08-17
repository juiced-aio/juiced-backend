package stores

import (
	e "errors"

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
