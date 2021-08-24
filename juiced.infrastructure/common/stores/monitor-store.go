package stores

import (
	e "errors"
	"math/rand"
	"time"

	"backend.juicedbot.io/juiced.infrastructure/common"
	"backend.juicedbot.io/juiced.infrastructure/common/entities"
	"backend.juicedbot.io/juiced.infrastructure/common/enums"
	"backend.juicedbot.io/juiced.infrastructure/common/errors"
	"backend.juicedbot.io/juiced.infrastructure/common/events"

	"backend.juicedbot.io/juiced.sitescripts/amazon"
	"backend.juicedbot.io/juiced.sitescripts/bestbuy"
	"backend.juicedbot.io/juiced.sitescripts/bigcartel"
	"backend.juicedbot.io/juiced.sitescripts/boxlunch"
	"backend.juicedbot.io/juiced.sitescripts/disney"
	"backend.juicedbot.io/juiced.sitescripts/gamestop"
	"backend.juicedbot.io/juiced.sitescripts/hottopic"
	"backend.juicedbot.io/juiced.sitescripts/newegg"
	"backend.juicedbot.io/juiced.sitescripts/shopify"
	"backend.juicedbot.io/juiced.sitescripts/target"
	"backend.juicedbot.io/juiced.sitescripts/topps"
	"backend.juicedbot.io/juiced.sitescripts/walmart"
	// Future sitescripts will be imported here
)

// MonitorStore stores information about running Monitors
type MonitorStore struct {
	AmazonMonitors    map[string]*amazon.Monitor
	BestbuyMonitors   map[string]*bestbuy.Monitor
	BigCartelMonitors map[string]*bigcartel.Monitor
	BoxlunchMonitors  map[string]*boxlunch.Monitor
	DisneyMonitors    map[string]*disney.Monitor
	GamestopMonitors  map[string]*gamestop.Monitor
	HottopicMonitors  map[string]*hottopic.Monitor
	NeweggMonitors    map[string]*newegg.Monitor
	ShopifyMonitors   map[string]*shopify.Monitor
	TargetMonitors    map[string]*target.Monitor
	ToppsMonitors     map[string]*topps.Monitor
	WalmartMonitors   map[string]*walmart.Monitor
	EventBus          *events.EventBus
}

// AddMonitorToStore adds the Monitor to the Store and returns true if successful
func (monitorStore *MonitorStore) AddMonitorToStore(monitor *entities.TaskGroup) error {
	var queryError error
	// Get ProxyGroup for monitor
	var proxyGroup *entities.ProxyGroup
	if monitor.MonitorProxyGroupID != "" {
		var ok bool

		proxyGroup, ok = proxyStore.ProxyGroups[monitor.MonitorProxyGroupID]
		if !ok {
			queryError = e.New("proxy group failure")
		}
	}
	switch monitor.MonitorRetailer {
	// Future sitescripts will have a case here
	case enums.Amazon:
		if _, ok := monitorStore.AmazonMonitors[monitor.GroupID]; ok && !monitor.UpdateMonitor {
			return nil
		}

		if queryError != nil {
			return queryError
		}

		if len(monitor.AmazonMonitorInfo.Monitors) == 0 {
			return e.New(errors.NoMonitorsError)
		}

		for _, monitor := range monitor.AmazonMonitorInfo.Monitors {
			if monitor.MonitorType == enums.FastSKUMonitor {
				if monitor.OFID == "" {
					return e.New(errors.MissingMonitorFieldsError)
				}
			}
		}

		amazonMonitor, err := amazon.CreateAmazonMonitor(monitor, proxyGroup, monitorStore.EventBus, monitor.AmazonMonitorInfo.Monitors)
		if err != nil {
			return e.New(errors.CreateMonitorError + err.Error())
		}

		monitorStore.AmazonMonitors[monitor.GroupID] = &amazonMonitor

	case enums.BestBuy:
		if _, ok := monitorStore.BestbuyMonitors[monitor.GroupID]; ok && !monitor.UpdateMonitor {
			return nil
		}

		if queryError != nil {
			return queryError
		}

		if len(monitor.BestbuyMonitorInfo.Monitors) == 0 {
			return e.New(errors.NoMonitorsError)
		}

		bestbuyMonitor, err := bestbuy.CreateBestbuyMonitor(monitor, proxyGroup, monitorStore.EventBus, monitor.BestbuyMonitorInfo.Monitors)
		if err != nil {
			return e.New(errors.CreateMonitorError + err.Error())
		}

		monitorStore.BestbuyMonitors[monitor.GroupID] = &bestbuyMonitor

	case enums.Disney:
		if _, ok := monitorStore.DisneyMonitors[monitor.GroupID]; ok && !monitor.UpdateMonitor {
			return nil
		}

		if queryError != nil {
			return queryError
		}

		if len(monitor.DisneyMonitorInfo.Monitors) == 0 {
			return e.New(errors.NoMonitorsError)
		}

		disneyMonitor, err := disney.CreateDisneyMonitor(monitor, proxyGroup, monitorStore.EventBus, monitor.DisneyMonitorInfo.Monitors)
		if err != nil {
			return e.New(errors.CreateMonitorError + err.Error())
		}

		monitorStore.DisneyMonitors[monitor.GroupID] = &disneyMonitor

	case enums.BoxLunch:
		if _, ok := monitorStore.BoxlunchMonitors[monitor.GroupID]; ok && !monitor.UpdateMonitor {
			return nil
		}

		if queryError != nil {
			return queryError
		}

		if len(monitor.BoxlunchMonitorInfo.Monitors) == 0 {
			return e.New(errors.NoMonitorsError)
		}

		boxlunchMonitor, err := boxlunch.CreateBoxlunchMonitor(monitor, proxyGroup, monitorStore.EventBus, monitor.BoxlunchMonitorInfo.Monitors)
		if err != nil {
			return e.New(errors.CreateMonitorError + err.Error())
		}

		monitorStore.BoxlunchMonitors[monitor.GroupID] = &boxlunchMonitor

	case enums.GameStop:
		if _, ok := monitorStore.GamestopMonitors[monitor.GroupID]; ok && !monitor.UpdateMonitor {
			return nil
		}

		if queryError != nil {
			return queryError
		}

		if len(monitor.GamestopMonitorInfo.Monitors) == 0 {
			return e.New(errors.NoMonitorsError)
		}

		gamestopMonitor, err := gamestop.CreateGamestopMonitor(monitor, proxyGroup, monitorStore.EventBus, monitor.GamestopMonitorInfo.Monitors)
		if err != nil {
			return e.New(errors.CreateMonitorError + err.Error())
		}
		monitorStore.GamestopMonitors[monitor.GroupID] = &gamestopMonitor

	case enums.HotTopic:
		if _, ok := monitorStore.HottopicMonitors[monitor.GroupID]; ok && !monitor.UpdateMonitor {
			return nil
		}

		if queryError != nil {
			return queryError
		}

		if len(monitor.HottopicMonitorInfo.Monitors) == 0 {
			return e.New(errors.NoMonitorsError)
		}

		hottopicMonitor, err := hottopic.CreateHottopicMonitor(monitor, proxyGroup, monitorStore.EventBus, monitor.HottopicMonitorInfo.Monitors)
		if err != nil {
			return e.New(errors.CreateMonitorError + err.Error())
		}

		monitorStore.HottopicMonitors[monitor.GroupID] = &hottopicMonitor

	case enums.Newegg:
		if _, ok := monitorStore.NeweggMonitors[monitor.GroupID]; ok && !monitor.UpdateMonitor {
			return nil
		}

		if queryError != nil {
			return queryError
		}

		if len(monitor.NeweggMonitorInfo.Monitors) == 0 {
			return e.New(errors.NoMonitorsError)
		}

		neweggMonitor, err := newegg.CreateNeweggMonitor(monitor, proxyGroup, monitorStore.EventBus, monitor.NeweggMonitorInfo.Monitors)
		if err != nil {
			return e.New(errors.CreateMonitorError + err.Error())
		}
		monitorStore.NeweggMonitors[monitor.GroupID] = &neweggMonitor

	case enums.Shopify:
		if _, ok := monitorStore.ShopifyMonitors[monitor.GroupID]; ok && !monitor.UpdateMonitor {
			return nil
		}

		if queryError != nil {
			return queryError
		}

		if len(monitor.ShopifyMonitorInfo.Monitors) == 0 {
			return e.New(errors.NoMonitorsError)
		}

		shopifyMonitor, err := shopify.CreateShopifyMonitor(monitor, proxyGroup, monitorStore.EventBus, monitor.ShopifyMonitorInfo.SiteURL, monitor.ShopifyMonitorInfo.SitePassword, monitor.ShopifyMonitorInfo.Monitors)
		if err != nil {
			return e.New(errors.CreateMonitorError + err.Error())
		}
		monitorStore.ShopifyMonitors[monitor.GroupID] = &shopifyMonitor

	case enums.BigCartel:
		if _, ok := monitorStore.BigCartelMonitors[monitor.GroupID]; ok && !monitor.UpdateMonitor {
			return nil
		}

		if queryError != nil {
			return queryError
		}

		if len(monitor.BigCartelMonitorInfo.Monitors) == 0 {
			return e.New(errors.NoMonitorsError)
		}

		bigCatelMonitors, err := bigcartel.CreateBigCartelMonitor(monitor, proxies, monitorStore.EventBus, monitor.BigCartelMonitorInfo.SiteURL, monitor.BigCartelMonitorInfo.Monitors)
		if err != nil {
			return e.New(errors.CreateMonitorError + err.Error())
		}
		monitorStore.BigCartelMonitors[monitor.GroupID] = &bigCatelMonitors

	case enums.Target:
		// Check if monitor exists in store already
		if _, ok := monitorStore.TargetMonitors[monitor.GroupID]; ok && !monitor.UpdateMonitor {
			return nil
		}
		// Only return false on a query error if the monitor doesn't exist in the store already
		if queryError != nil {
			return queryError
		}
		// Make sure necessary fields exist
		if len(monitor.TargetMonitorInfo.Monitors) == 0 {
			return e.New(errors.NoMonitorsError)

		}
		// Create monitor
		targetMonitor, err := target.CreateTargetMonitor(monitor, proxyGroup, monitorStore.EventBus, monitor.TargetMonitorInfo)
		if err != nil {
			return e.New(errors.CreateMonitorError + err.Error())
		}
		// Add task to store
		monitorStore.TargetMonitors[monitor.GroupID] = &targetMonitor

	case enums.Topps:
		if _, ok := monitorStore.ToppsMonitors[monitor.GroupID]; ok && !monitor.UpdateMonitor {
			return nil
		}

		if queryError != nil {
			return queryError
		}

		if len(monitor.ToppsMonitorInfo.Monitors) == 0 {
			return e.New(errors.NoMonitorsError)
		}

		toppsMonitor, err := topps.CreateToppsMonitor(monitor, proxyGroup, monitorStore.EventBus, monitor.ToppsMonitorInfo.Monitors)
		if err != nil {
			return e.New(errors.CreateMonitorError + err.Error())
		}
		monitorStore.ToppsMonitors[monitor.GroupID] = &toppsMonitor

	case enums.Walmart:
		// Check if monitor exists in store already
		if _, ok := monitorStore.WalmartMonitors[monitor.GroupID]; ok && !monitor.UpdateMonitor {
			return nil
		}
		// Only return false on a query error if the monitor doesn't exist in the store already
		if queryError != nil {
			return queryError
		}
		// Make sure necessary fields exist
		if len(monitor.WalmartMonitorInfo.Monitors) == 0 {
			return e.New(errors.NoMonitorsError)
		}
		// Create monitor
		walmartMonitor, err := walmart.CreateWalmartMonitor(monitor, proxyGroup, monitorStore.EventBus, monitor.WalmartMonitorInfo.Monitors)
		if err != nil {
			return e.New(errors.CreateMonitorError + err.Error())
		}
		// Add task to store
		monitorStore.WalmartMonitors[monitor.GroupID] = &walmartMonitor

	}
	monitor.UpdateMonitor = false
	return nil
}

// StartMonitor runs the Run() function for the given Monitor and returns true if successful
func (monitorStore *MonitorStore) StartMonitor(monitor *entities.TaskGroup) error {
	// Add monitor to store (if it already exists, this will return true)
	err := monitorStore.AddMonitorToStore(monitor)
	if err != nil {
		return err
	}

	// If the Monitor is already running, then we're all set already
	if monitor.MonitorStatus != enums.MonitorIdle {
		return nil
	}

	// Otherwise, start the Monitor
	switch monitor.MonitorRetailer {
	// Future sitescripts will have a case here
	case enums.Amazon:
		if amazonMonitor, ok := monitorStore.AmazonMonitors[monitor.GroupID]; ok {
			amazonMonitor.Monitor.StopFlag = false
		}
		go monitorStore.AmazonMonitors[monitor.GroupID].RunMonitor()

	case enums.BestBuy:
		if bestbuyMonitor, ok := monitorStore.BestbuyMonitors[monitor.GroupID]; ok {
			bestbuyMonitor.Monitor.StopFlag = false
		}
		go monitorStore.BestbuyMonitors[monitor.GroupID].RunMonitor()

	case enums.BoxLunch:
		if boxlunchMonitor, ok := monitorStore.BoxlunchMonitors[monitor.GroupID]; ok {
			boxlunchMonitor.Monitor.StopFlag = false
		}
		go monitorStore.BoxlunchMonitors[monitor.GroupID].RunMonitor()

	case enums.Disney:
		if disneyMonitor, ok := monitorStore.DisneyMonitors[monitor.GroupID]; ok {
			disneyMonitor.Monitor.StopFlag = false
		}
		go monitorStore.DisneyMonitors[monitor.GroupID].RunMonitor()

	case enums.GameStop:
		if gamestopMonitor, ok := monitorStore.GamestopMonitors[monitor.GroupID]; ok {
			gamestopMonitor.Monitor.StopFlag = false
		}
		go monitorStore.GamestopMonitors[monitor.GroupID].RunMonitor()

	case enums.HotTopic:
		if hottopicMonitor, ok := monitorStore.HottopicMonitors[monitor.GroupID]; ok {
			hottopicMonitor.Monitor.StopFlag = false
		}
		go monitorStore.HottopicMonitors[monitor.GroupID].RunMonitor()

	case enums.Newegg:
		if neweggMonitor, ok := monitorStore.NeweggMonitors[monitor.GroupID]; ok {
			neweggMonitor.Monitor.StopFlag = false
		}
		go monitorStore.NeweggMonitors[monitor.GroupID].RunMonitor()

	case enums.Shopify:
		if shopifyMonitor, ok := monitorStore.ShopifyMonitors[monitor.GroupID]; ok {
			shopifyMonitor.Monitor.StopFlag = false
		}
		go monitorStore.ShopifyMonitors[monitor.GroupID].RunMonitor()

	case enums.BigCartel:
		if bigcartelMonitor, ok := monitorStore.BigCartelMonitors[monitor.GroupID]; ok {
			bigcartelMonitor.Monitor.StopFlag = false
		}
		go monitorStore.BigCartelMonitors[monitor.GroupID].RunMonitor()

	case enums.Target:
		if targetMonitor, ok := monitorStore.TargetMonitors[monitor.GroupID]; ok {
			targetMonitor.Monitor.StopFlag = false
		}
		go monitorStore.TargetMonitors[monitor.GroupID].RunMonitor()

	case enums.Topps:
		if toppsMonitor, ok := monitorStore.ToppsMonitors[monitor.GroupID]; ok {
			toppsMonitor.Monitor.StopFlag = false
		}
		go monitorStore.ToppsMonitors[monitor.GroupID].RunMonitor()

	case enums.Walmart:
		if walmartMonitor, ok := monitorStore.WalmartMonitors[monitor.GroupID]; ok {
			walmartMonitor.Monitor.StopFlag = false
		}
		go monitorStore.WalmartMonitors[monitor.GroupID].RunMonitor()
	}

	return nil
}

// StopMonitor sets the stop field for the given Monitor and returns true if successful
func (monitorStore *MonitorStore) StopMonitor(monitor *entities.TaskGroup) error {
	switch monitor.MonitorRetailer {
	// Future sitescripts will have a case here
	case enums.Amazon:
		if amazonMonitor, ok := monitorStore.AmazonMonitors[monitor.GroupID]; ok {
			amazonMonitor.Monitor.StopFlag = true
		}

	case enums.BestBuy:
		if bestbuyMonitor, ok := monitorStore.BestbuyMonitors[monitor.GroupID]; ok {
			bestbuyMonitor.Monitor.StopFlag = true
		}

	case enums.BoxLunch:
		if boxlunchMonitor, ok := monitorStore.BoxlunchMonitors[monitor.GroupID]; ok {
			boxlunchMonitor.Monitor.StopFlag = true
		}

	case enums.Disney:
		if disneyMonitor, ok := monitorStore.DisneyMonitors[monitor.GroupID]; ok {
			disneyMonitor.Monitor.StopFlag = true
		}

	case enums.GameStop:
		if gamestopMonitor, ok := monitorStore.GamestopMonitors[monitor.GroupID]; ok {
			gamestopMonitor.Monitor.StopFlag = true
		}

	case enums.HotTopic:
		if hottopicMonitor, ok := monitorStore.HottopicMonitors[monitor.GroupID]; ok {
			hottopicMonitor.Monitor.StopFlag = true
		}

	case enums.Newegg:
		if neweggMonitor, ok := monitorStore.NeweggMonitors[monitor.GroupID]; ok {
			neweggMonitor.Monitor.StopFlag = true
		}

	case enums.Shopify:
		if shopifyMonitor, ok := monitorStore.ShopifyMonitors[monitor.GroupID]; ok {
			shopifyMonitor.Monitor.StopFlag = true
		}

	case enums.BigCartel:
		if bigCartelMonitor, ok := monitorStore.BigCartelMonitors[monitor.GroupID]; ok {
			bigCartelMonitor.Monitor.StopFlag = true
		}

	case enums.Target:
		if targetMonitor, ok := monitorStore.TargetMonitors[monitor.GroupID]; ok {
			targetMonitor.Monitor.StopFlag = true
		}

	case enums.Topps:
		if toppsMonitor, ok := monitorStore.ToppsMonitors[monitor.GroupID]; ok {
			toppsMonitor.Monitor.StopFlag = true
		}

	case enums.Walmart:
		if walmartMonitor, ok := monitorStore.WalmartMonitors[monitor.GroupID]; ok {
			walmartMonitor.Monitor.StopFlag = true
		}

	default:
		return e.New(errors.InvalidMonitorRetailerError)

	}
	return nil
}

// UpdateMonitorProxy will update the given monitor with the given proxy and return true if successful
func (monitorStore *MonitorStore) UpdateMonitorProxy(monitor *entities.TaskGroup, proxy *entities.Proxy) bool {
	switch monitor.MonitorRetailer {
	// Future sitescripts will have a case here
	case enums.Amazon:
		if amazonMonitor, ok := monitorStore.AmazonMonitors[monitor.GroupID]; ok {
			amazonMonitor.Monitor.Proxy = proxy
		}
		return true

	case enums.BestBuy:
		if bestbuyMonitor, ok := monitorStore.BestbuyMonitors[monitor.GroupID]; ok {
			bestbuyMonitor.Monitor.Proxy = proxy
		}
		return true

	case enums.BoxLunch:
		if boxlunchMonitor, ok := monitorStore.BoxlunchMonitors[monitor.GroupID]; ok {
			boxlunchMonitor.Monitor.StopFlag = true
		}
		return true

	case enums.Disney:
		if disneyMonitor, ok := monitorStore.DisneyMonitors[monitor.GroupID]; ok {
			disneyMonitor.Monitor.Proxy = proxy
		}
		return true

	case enums.GameStop:
		if gamestopMonitor, ok := monitorStore.GamestopMonitors[monitor.GroupID]; ok {
			gamestopMonitor.Monitor.Proxy = proxy
		}
		return true

	case enums.HotTopic:
		if hottopicMonitor, ok := monitorStore.HottopicMonitors[monitor.GroupID]; ok {
			hottopicMonitor.Monitor.Proxy = proxy
		}
		return true

	case enums.Newegg:
		if neweggMonitor, ok := monitorStore.NeweggMonitors[monitor.GroupID]; ok {
			neweggMonitor.Monitor.Proxy = proxy
		}
		return true

	case enums.Shopify:
		if shopifyMonitor, ok := monitorStore.ShopifyMonitors[monitor.GroupID]; ok {
			shopifyMonitor.Monitor.Proxy = proxy
		}
		return true
	case enums.BigCartel:
		if bigCartelMonitor, ok := monitorStore.BigCartelMonitors[monitor.GroupID]; ok {
			bigCartelMonitor.Monitor.Proxy = proxy
		}
		return true
	case enums.Target:
		if targetMonitor, ok := monitorStore.TargetMonitors[monitor.GroupID]; ok {
			targetMonitor.Monitor.Proxy = proxy
		}
		return true

	case enums.Topps:
		if toppsMonitor, ok := monitorStore.ToppsMonitors[monitor.GroupID]; ok {
			toppsMonitor.Monitor.Proxy = proxy
		}
		return true

	case enums.Walmart:
		if walmartMonitor, ok := monitorStore.WalmartMonitors[monitor.GroupID]; ok {
			walmartMonitor.Monitor.Proxy = proxy
		}
		return true

	}
	return false
}

// TODO: Test the efficiency of these functions.
// It's technically O(n^2), but most users won't have more than 5-10 task groups running at once, tops.
func (monitorStore *MonitorStore) CheckAmazonMonitorStock() {
	for {
		for monitorID, amazonMonitor := range monitorStore.AmazonMonitors {
			if len(amazonMonitor.InStock) > 0 {
				taskGroup := amazonMonitor.Monitor.TaskGroup
				for _, taskID := range taskGroup.TaskIDs {
					if amazonTask, ok := taskStore.AmazonTasks[taskID]; ok {
						if ok && amazonTask.Task.Task.TaskGroupID == monitorID {
							randomNumber := rand.Intn(len(amazonMonitor.InStock))
							amazonTask.StockData = amazonMonitor.InStock[randomNumber]
						}
					}
				}
			}
		}
		time.Sleep(common.MS_TO_WAIT)
	}
}

func (monitorStore *MonitorStore) CheckBestBuyMonitorStock() {
	for {
		for monitorID, bestbuyMonitor := range monitorStore.BestbuyMonitors {
			if len(bestbuyMonitor.InStock) > 0 {
				taskGroup := bestbuyMonitor.Monitor.TaskGroup
				for _, taskID := range taskGroup.TaskIDs {
					if bestbuyTask, ok := taskStore.BestbuyTasks[taskID]; ok {
						if ok && bestbuyTask.Task.Task.TaskGroupID == monitorID {
							randomNumber := rand.Intn(len(bestbuyMonitor.InStock))
							bestbuyTask.CheckoutInfo.SKUInStock = bestbuyMonitor.InStock[randomNumber].SKU
							bestbuyTask.CheckoutInfo.Price = bestbuyMonitor.InStock[randomNumber].Price
						}
					}
				}
			}
		}
		time.Sleep(common.MS_TO_WAIT)
	}
}

func (monitorStore *MonitorStore) CheckDisneyMonitorStock() {
	for {
		for monitorID, disneyMonitor := range monitorStore.DisneyMonitors {
			if len(disneyMonitor.InStock) > 0 {
				taskGroup := disneyMonitor.Monitor.TaskGroup
				for _, taskID := range taskGroup.TaskIDs {
					if disneyTask, ok := taskStore.DisneyTasks[taskID]; ok {
						if ok && disneyTask.Task.Task.TaskGroupID == monitorID {
							randomNumber := rand.Intn(len(disneyMonitor.InStock))
							disneyTask.StockData = disneyMonitor.InStock[randomNumber]
						}
					}
				}
			}
		}
		time.Sleep(common.MS_TO_WAIT)
	}
}

func (monitorStore *MonitorStore) CheckBoxlunchMonitorStock() {
	for {
		for monitorID, boxlunchMonitor := range monitorStore.BoxlunchMonitors {
			if len(boxlunchMonitor.InStock) > 0 {
				taskGroup := boxlunchMonitor.Monitor.TaskGroup
				for _, taskID := range taskGroup.TaskIDs {
					if boxlunchTask, ok := taskStore.BoxlunchTasks[taskID]; ok {
						if ok && boxlunchTask.Task.Task.TaskGroupID == monitorID {
							boxlunchTask.StockData = boxlunchMonitor.InStock[rand.Intn(len(boxlunchMonitor.InStock))]
						}
					}
				}
			}
		}
		time.Sleep(common.MS_TO_WAIT)
	}
}

func (monitorStore *MonitorStore) CheckGameStopMonitorStock() {
	for {
		for monitorID, gamestopMonitor := range monitorStore.GamestopMonitors {
			if len(gamestopMonitor.InStock) > 0 {
				taskGroup := gamestopMonitor.Monitor.TaskGroup
				for _, taskID := range taskGroup.TaskIDs {
					if gamestopTask, ok := taskStore.GamestopTasks[taskID]; ok {
						if ok && gamestopTask.Task.Task.TaskGroupID == monitorID {
							randomNumber := rand.Intn(len(gamestopMonitor.InStock))
							gamestopTask.StockData = gamestopMonitor.InStock[randomNumber]
						}
					}
				}
			}
		}
		time.Sleep(common.MS_TO_WAIT)
	}
}

func (monitorStore *MonitorStore) CheckHotTopicMonitorStock() {
	for {
		for monitorID, hottopicMonitor := range monitorStore.HottopicMonitors {
			if len(hottopicMonitor.InStock) > 0 {
				taskGroup := hottopicMonitor.Monitor.TaskGroup
				for _, taskID := range taskGroup.TaskIDs {
					if hottopicTask, ok := taskStore.HottopicTasks[taskID]; ok {
						if ok && hottopicTask.Task.Task.TaskGroupID == monitorID {
							hottopicTask.StockData = hottopicMonitor.InStock[rand.Intn(len(hottopicMonitor.InStock))]
						}
					}
				}
			}
		}
		time.Sleep(common.MS_TO_WAIT)
	}
}

func (monitorStore *MonitorStore) CheckNeweggMonitorStock() {
	for {
		for monitorID, neweggMonitor := range monitorStore.NeweggMonitors {
			if len(neweggMonitor.InStock) > 0 {
				taskGroup := neweggMonitor.Monitor.TaskGroup
				for _, taskID := range taskGroup.TaskIDs {
					if neweggTask, ok := taskStore.NeweggTasks[taskID]; ok {
						if ok && neweggTask.Task.Task.TaskGroupID == monitorID {
							neweggTask.StockData = neweggMonitor.InStock[rand.Intn(len(neweggMonitor.InStock))]
						}
					}
				}
			}
		}
		time.Sleep(1 * time.Second / 100)
	}
}

func (monitorStore *MonitorStore) CheckShopifyMonitorStock() {
	for {
		for monitorID, shopifyMonitor := range monitorStore.ShopifyMonitors {
			if len(shopifyMonitor.InStock) > 0 {
				taskGroup := shopifyMonitor.Monitor.TaskGroup
				for _, taskID := range taskGroup.TaskIDs {
					if shopifyTask, ok := taskStore.ShopifyTasks[taskID]; ok {
						if ok && shopifyTask.Task.Task.TaskGroupID == monitorID {
							shopifyTask.InStockData = shopifyMonitor.InStock[rand.Intn(len(shopifyMonitor.InStock))]
						}
					}
				}
			}
		}
		time.Sleep(common.MS_TO_WAIT)
	}
}

func (monitorStore *MonitorStore) CheckBigCartelMonitorStock() {
	for {
		for monitorID, bigCartelMonitor := range monitorStore.BigCartelMonitors {
			if len(bigCartelMonitor.InStock) > 0 {
				taskGroup := bigCartelMonitor.Monitor.TaskGroup
				for _, taskID := range taskGroup.TaskIDs {
					if bigCartelTask, ok := taskStore.BigCartelTasks[taskID]; ok {
						if ok && bigCartelTask.Task.Task.TaskGroupID == monitorID {
							bigCartelTask.InStockData = bigCartelMonitor.InStock[rand.Intn(len(bigCartelMonitor.InStock))]
						}
					}
				}
			}
		}
		time.Sleep(1 * time.Second / 100)
	}
}

func (monitorStore *MonitorStore) CheckTargetMonitorStock() {
	for {
		for monitorID, targetMonitor := range monitorStore.TargetMonitors {
			if len(targetMonitor.InStockForPickup) > 0 || len(targetMonitor.InStockForShip) > 0 {
				taskGroup := targetMonitor.Monitor.TaskGroup
				for _, taskID := range taskGroup.TaskIDs {
					if targetTask, ok := taskStore.TargetTasks[taskID]; ok {
						if ok && targetTask.Task.Task.TaskGroupID == monitorID {
							var inStockForShip []target.SingleStockData
							var inStockForPickup []target.SingleStockData

							for _, value := range targetMonitor.InStockForShip.Items() {
								inStockForShip = append(inStockForShip, value.(target.SingleStockData))
							}
							for _, value := range targetMonitor.InStockForPickup.Items() {
								inStockForPickup = append(inStockForPickup, value.(target.SingleStockData))
							}

							if len(inStockForPickup) > 0 {
								targetTask.InStockData = inStockForPickup[rand.Intn(len(inStockForPickup))]
								targetTask.AccountInfo.StoreID = targetMonitor.StoreID
								targetTask.CheckoutType = enums.CheckoutTypePICKUP
							} else if len(inStockForShip) > 0 {
								targetTask.InStockData = inStockForShip[rand.Intn(len(inStockForShip))]
								targetTask.CheckoutType = enums.CheckoutTypeSHIP
							}
						}
					}
				}
			}
		}
		time.Sleep(common.MS_TO_WAIT)
	}
}

func (monitorStore *MonitorStore) CheckToppsMonitorStock() {
	for {
		for monitorID, toppsMonitor := range monitorStore.ToppsMonitors {
			if len(toppsMonitor.InStock) > 0 {
				taskGroup := toppsMonitor.Monitor.TaskGroup
				for _, taskID := range taskGroup.TaskIDs {
					if toppsTask, ok := taskStore.ToppsTasks[taskID]; ok {
						if ok && toppsTask.Task.Task.TaskGroupID == monitorID {
							toppsTask.StockData = toppsMonitor.InStock[rand.Intn(len(toppsMonitor.InStock))]
						}
					}
				}
			}
		}
		time.Sleep(common.MS_TO_WAIT)
	}
}

func (monitorStore *MonitorStore) CheckWalmartMonitorStock() {
	for {
		for monitorID, walmartMonitor := range monitorStore.WalmartMonitors {
			if len(walmartMonitor.InStockForShip) > 0 {
				taskGroup := walmartMonitor.Monitor.TaskGroup
				for _, taskID := range taskGroup.TaskIDs {
					if walmartTask, ok := taskStore.WalmartTasks[taskID]; ok {
						if ok && walmartTask.Task.Task.TaskGroupID == monitorID {
							randomNumber := rand.Intn(len(walmartMonitor.InStockForShip))
							walmartTask.StockData = walmartMonitor.InStockForShip[randomNumber]
						}
					}
				}
			}
		}
		time.Sleep(common.MS_TO_WAIT)
	}
}

var monitorStore *MonitorStore

// InitMonitorStore initializes the singleton instance of the Store
func InitMonitorStore(eventBus *events.EventBus) {
	monitorStore = &MonitorStore{
		AmazonMonitors:    make(map[string]*amazon.Monitor),
		BestbuyMonitors:   make(map[string]*bestbuy.Monitor),
		BigCartelMonitors: make(map[string]*bigcartel.Monitor),
		BoxlunchMonitors:  make(map[string]*boxlunch.Monitor),
		DisneyMonitors:    make(map[string]*disney.Monitor),
		GamestopMonitors:  make(map[string]*gamestop.Monitor),
		HottopicMonitors:  make(map[string]*hottopic.Monitor),
		ShopifyMonitors:   make(map[string]*shopify.Monitor),
		NeweggMonitors:    make(map[string]*newegg.Monitor),
		TargetMonitors:    make(map[string]*target.Monitor),
		ToppsMonitors:     make(map[string]*topps.Monitor),
		WalmartMonitors:   make(map[string]*walmart.Monitor),

		EventBus: eventBus,
	}

	go monitorStore.CheckAmazonMonitorStock()
	go monitorStore.CheckBestBuyMonitorStock()
	go monitorStore.CheckBoxlunchMonitorStock()
	go monitorStore.CheckDisneyMonitorStock()
	go monitorStore.CheckGameStopMonitorStock()
	go monitorStore.CheckHotTopicMonitorStock()
	go monitorStore.CheckNeweggMonitorStock()
	go monitorStore.CheckShopifyMonitorStock()
	go monitorStore.CheckBigCartelMonitorStock()
	go monitorStore.CheckTargetMonitorStock()
	go monitorStore.CheckToppsMonitorStock()
	go monitorStore.CheckWalmartMonitorStock()
}

// GetMonitorStatus returns the status of the given TaskGroup's monitor
func GetMonitorStatus(groupID string) string {
	if monitor, ok := monitorStore.AmazonMonitors[groupID]; ok {
		return monitor.Monitor.TaskGroup.MonitorStatus
	}
	if monitor, ok := monitorStore.BestbuyMonitors[groupID]; ok {
		return monitor.Monitor.TaskGroup.MonitorStatus
	}
	if monitor, ok := monitorStore.DisneyMonitors[groupID]; ok {
		return monitor.Monitor.TaskGroup.MonitorStatus
	}
	if monitor, ok := monitorStore.GamestopMonitors[groupID]; ok {
		return monitor.Monitor.TaskGroup.MonitorStatus
	}
	if monitor, ok := monitorStore.HottopicMonitors[groupID]; ok {
		return monitor.Monitor.TaskGroup.MonitorStatus
	}
	if monitor, ok := monitorStore.NeweggMonitors[groupID]; ok {
		return monitor.Monitor.TaskGroup.MonitorStatus
	}
	if monitor, ok := monitorStore.ShopifyMonitors[groupID]; ok {
		return monitor.Monitor.TaskGroup.MonitorStatus
	}
	if monitor, ok := monitorStore.TargetMonitors[groupID]; ok {
		return monitor.Monitor.TaskGroup.MonitorStatus
	}
	if monitor, ok := monitorStore.ToppsMonitors[groupID]; ok {
		return monitor.Monitor.TaskGroup.MonitorStatus
	}
	if monitor, ok := monitorStore.WalmartMonitors[groupID]; ok {
		return monitor.Monitor.TaskGroup.MonitorStatus
	}

	return ""
}

// GetMonitorStore returns the singleton instance of the EventBus
func GetMonitorStore() *MonitorStore {
	return monitorStore
}
