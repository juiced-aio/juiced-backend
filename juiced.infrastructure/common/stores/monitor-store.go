package stores

import (
	"log"
	"math/rand"
	"time"

	"backend.juicedbot.io/juiced.infrastructure/common/entities"
	"backend.juicedbot.io/juiced.infrastructure/common/enums"
	"backend.juicedbot.io/juiced.infrastructure/common/events"
	"backend.juicedbot.io/juiced.infrastructure/queries"

	"backend.juicedbot.io/juiced.sitescripts/amazon"
	"backend.juicedbot.io/juiced.sitescripts/bestbuy"
	"backend.juicedbot.io/juiced.sitescripts/gamestop"
	"backend.juicedbot.io/juiced.sitescripts/hottopic"
	"backend.juicedbot.io/juiced.sitescripts/target"
	"backend.juicedbot.io/juiced.sitescripts/walmart"
	// Future sitescripts will be imported here
)

// MonitorStore stores information about running Monitors
type MonitorStore struct {
	AmazonMonitors   map[string]*amazon.Monitor
	BestbuyMonitors  map[string]*bestbuy.Monitor
	GamestopMonitors map[string]*gamestop.Monitor
	HottopicMonitors map[string]*hottopic.Monitor
	TargetMonitors   map[string]*target.Monitor
	WalmartMonitors  map[string]*walmart.Monitor
	EventBus         *events.EventBus
}

// AddMonitorToStore adds the Monitor to the Store and returns true if successful
func (monitorStore *MonitorStore) AddMonitorToStore(monitor *entities.TaskGroup) bool {
	queryError := false
	// Get Proxy for monitor
	proxies := []entities.Proxy{}
	if monitor.MonitorProxyGroupID != "" {
		proxyGroup, err := queries.GetProxyGroup(monitor.MonitorProxyGroupID)
		if err != nil {
			queryError = true
		}
		proxies = proxyGroup.Proxies
	}
	switch monitor.MonitorRetailer {
	// Future sitescripts will have a case here
	case enums.Amazon:
		if _, ok := monitorStore.AmazonMonitors[monitor.GroupID]; ok && !monitor.UpdateMonitor {
			return true
		}

		if queryError {
			return false
		}

		if len(monitor.AmazonMonitorInfo.Monitors) == 0 {
			return false
		}

		amazonMonitor, err := amazon.CreateAmazonMonitor(monitor, proxies, monitorStore.EventBus, monitor.AmazonMonitorInfo.Monitors)
		if err != nil {
			log.Println(8)
			log.Println(err.Error())
			return false
		}

		monitorStore.AmazonMonitors[monitor.GroupID] = &amazonMonitor

	case enums.BestBuy:
		if _, ok := monitorStore.BestbuyMonitors[monitor.GroupID]; ok && !monitor.UpdateMonitor {
			return true
		}

		if queryError {
			return false
		}

		if len(monitor.BestbuyMonitorInfo.Monitors) == 0 {
			return false
		}

		bestbuyMonitor, err := bestbuy.CreateBestbuyMonitor(monitor, proxies, monitorStore.EventBus, monitor.BestbuyMonitorInfo.Monitors)
		if err != nil {
			return false
		}

		monitorStore.BestbuyMonitors[monitor.GroupID] = &bestbuyMonitor

	case enums.GameStop:
		if _, ok := monitorStore.GamestopMonitors[monitor.GroupID]; ok && !monitor.UpdateMonitor {
			return true
		}

		if queryError {
			return false
		}

		if len(monitor.GamestopMonitorInfo.Monitors) == 0 {
			return false
		}

		gamestopMonitor, err := gamestop.CreateGamestopMonitor(monitor, proxies, monitorStore.EventBus, monitor.GamestopMonitorInfo.Monitors)
		if err != nil {
			return false
		}
		monitorStore.GamestopMonitors[monitor.GroupID] = &gamestopMonitor

	case enums.HotTopic:
		if _, ok := monitorStore.HottopicMonitors[monitor.GroupID]; ok && !monitor.UpdateMonitor {
			return true
		}

		if queryError {
			return false
		}

		if len(monitor.HottopicMonitorInfo.Monitors) == 0 {
			return false
		}

		hottopicMonitor, err := hottopic.CreateHottopicMonitor(monitor, proxies, monitorStore.EventBus, monitor.HottopicMonitorInfo.Monitors)
		if err != nil {
			return false
		}

		monitorStore.HottopicMonitors[monitor.GroupID] = &hottopicMonitor

	case enums.Target:
		// Check if monitor exists in store already
		if _, ok := monitorStore.TargetMonitors[monitor.GroupID]; ok && !monitor.UpdateMonitor {
			return true
		}
		// Only return false on a query error if the monitor doesn't exist in the store already
		if queryError {
			return false
		}
		// Make sure necessary fields exist
		if len(monitor.TargetMonitorInfo.Monitors) == 0 {
			return false

		}
		// Create monitor
		targetMonitor, err := target.CreateTargetMonitor(monitor, proxies, monitorStore.EventBus, monitor.TargetMonitorInfo)
		if err != nil {
			return false
		}
		// Add task to store
		monitorStore.TargetMonitors[monitor.GroupID] = &targetMonitor

	case enums.Walmart:
		// Check if monitor exists in store already
		if _, ok := monitorStore.WalmartMonitors[monitor.GroupID]; ok && !monitor.UpdateMonitor {
			return true
		}
		// Only return false on a query error if the monitor doesn't exist in the store already
		if queryError {
			return false
		}
		// Make sure necessary fields exist
		emptyString := ""
		if monitor.WalmartMonitorInfo.MonitorType == emptyString || len(monitor.WalmartMonitorInfo.SKUs) == 0 {
			return false
		}
		// Create monitor
		walmartMonitor, err := walmart.CreateWalmartMonitor(monitor, proxies, monitorStore.EventBus, monitor.WalmartMonitorInfo.MonitorType, monitor.WalmartMonitorInfo.SKUs)
		if err != nil {
			return false
		}
		// Add task to store
		monitorStore.WalmartMonitors[monitor.GroupID] = &walmartMonitor

	}
	monitor.UpdateMonitor = false
	return true
}

// StartMonitor runs the Run() function for the given Monitor and returns true if successful
func (monitorStore *MonitorStore) StartMonitor(monitor *entities.TaskGroup) bool {
	// Add monitor to store (if it already exists, this will return true)
	added := monitorStore.AddMonitorToStore(monitor)
	if !added {
		log.Println(4)
		return false
	}

	// If the Monitor is already running, then we're all set already
	if monitor.MonitorStatus != enums.MonitorIdle {
		return true
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

	case enums.Target:
		if targetMonitor, ok := monitorStore.TargetMonitors[monitor.GroupID]; ok {
			targetMonitor.Monitor.StopFlag = false
		}
		go monitorStore.TargetMonitors[monitor.GroupID].RunMonitor()

	case enums.Walmart:
		if walmartMonitor, ok := monitorStore.WalmartMonitors[monitor.GroupID]; ok {
			walmartMonitor.Monitor.StopFlag = false
		}
		go monitorStore.WalmartMonitors[monitor.GroupID].RunMonitor()
	}

	return true
}

// StopMonitor sets the stop field for the given Monitor and returns true if successful
func (monitorStore *MonitorStore) StopMonitor(monitor *entities.TaskGroup) bool {
	switch monitor.MonitorRetailer {
	// Future sitescripts will have a case here
	case enums.Amazon:
		if amazonMonitor, ok := monitorStore.AmazonMonitors[monitor.GroupID]; ok {
			amazonMonitor.Monitor.StopFlag = true
		}
		return true

	case enums.BestBuy:
		if bestbuyMonitor, ok := monitorStore.BestbuyMonitors[monitor.GroupID]; ok {
			bestbuyMonitor.Monitor.StopFlag = true
		}
		return true

	case enums.GameStop:
		if gamestopMonitor, ok := monitorStore.GamestopMonitors[monitor.GroupID]; ok {
			gamestopMonitor.Monitor.StopFlag = true
		}
		return true

	case enums.HotTopic:
		if hottopicMonitor, ok := monitorStore.HottopicMonitors[monitor.GroupID]; ok {
			hottopicMonitor.Monitor.StopFlag = true
		}
		return true

	case enums.Target:
		if targetMonitor, ok := monitorStore.TargetMonitors[monitor.GroupID]; ok {
			targetMonitor.Monitor.StopFlag = true
		}
		return true

	case enums.Walmart:
		if walmartMonitor, ok := monitorStore.WalmartMonitors[monitor.GroupID]; ok {
			walmartMonitor.Monitor.StopFlag = true
		}
		return true

	}
	return false
}

// UpdateMonitorProxy will update the given monitor with the given proxy and return true if successful
func (monitorStore *MonitorStore) UpdateMonitorProxy(monitor *entities.TaskGroup, proxy entities.Proxy) bool {
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

	case enums.Target:
		if targetMonitor, ok := monitorStore.TargetMonitors[monitor.GroupID]; ok {
			targetMonitor.Monitor.Proxy = proxy
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
							amazonTask.TaskInfo.ASIN = amazonMonitor.InStock[randomNumber].ASIN
							amazonTask.TaskInfo.OfferID = amazonMonitor.InStock[randomNumber].OfferID
							amazonTask.TaskInfo.ItemName = amazonMonitor.InStock[randomNumber].ItemName
							amazonTask.CheckoutInfo.Price = amazonMonitor.InStock[randomNumber].Price
							amazonTask.CheckoutInfo.AntiCsrf = amazonMonitor.InStock[randomNumber].AntiCsrf
							amazonTask.CheckoutInfo.PID = amazonMonitor.InStock[randomNumber].PID
							amazonTask.CheckoutInfo.RID = amazonMonitor.InStock[randomNumber].RID
							amazonTask.CheckoutInfo.ImageURL = amazonMonitor.InStock[randomNumber].ImageURL
							amazonTask.CheckoutInfo.UA = amazonMonitor.InStock[randomNumber].UA
							amazonTask.CheckoutInfo.MonitorType = enums.MonitorType(amazonMonitor.InStock[randomNumber].MonitorType)
						}
					}
				}
			}
		}
		time.Sleep(1 * time.Second / 100)
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
		time.Sleep(1 * time.Second / 100)
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
							gamestopTask.CheckoutInfo.SKUInStock = gamestopMonitor.InStock[randomNumber].SKU
							gamestopTask.CheckoutInfo.Price = gamestopMonitor.InStock[randomNumber].Price
							gamestopTask.CheckoutInfo.ItemName = gamestopMonitor.InStock[randomNumber].ItemName
							gamestopTask.CheckoutInfo.PID = gamestopMonitor.InStock[randomNumber].PID
							gamestopTask.CheckoutInfo.ImageURL = gamestopMonitor.InStock[randomNumber].ImageURL
							gamestopTask.CheckoutInfo.ProductURL = gamestopMonitor.InStock[randomNumber].ProductURL
						}
					}
				}
			}
		}
		time.Sleep(1 * time.Second / 100)
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
							hottopicTask.Pid = hottopicMonitor.InStock[rand.Intn(len(hottopicMonitor.InStock))].PID
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

							if targetTask.CheckoutType == enums.CheckoutTypePICKUP && len(inStockForPickup) > 0 {
								targetTask.InStockData = inStockForPickup[rand.Intn(len(inStockForPickup))]
								targetTask.AccountInfo.StoreID = targetMonitor.StoreID
							} else if targetTask.CheckoutType == enums.CheckoutTypeSHIP && len(inStockForShip) > 0 {
								targetTask.InStockData = inStockForShip[rand.Intn(len(inStockForShip))]
							} else {
								if len(inStockForShip) > 0 {
									targetTask.InStockData = inStockForShip[rand.Intn(len(inStockForShip))]
								} else if len(inStockForPickup) > 0 {
									targetTask.InStockData = inStockForPickup[rand.Intn(len(inStockForPickup))]
									targetTask.AccountInfo.StoreID = targetMonitor.StoreID
								}
							}

						}
					}
				}
			}
		}
		time.Sleep(1 * time.Second / 100)
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
							walmartTask.Sku = walmartMonitor.InStockForShip[randomNumber].Sku
							walmartTask.OfferID = walmartMonitor.InStockForShip[randomNumber].OfferID
						}
					}
				}
			}
		}
		time.Sleep(1 * time.Second / 100)
	}
}

var monitorStore *MonitorStore

// InitMonitorStore initializes the singleton instance of the Store
func InitMonitorStore(eventBus *events.EventBus) {
	monitorStore = &MonitorStore{
		AmazonMonitors:   make(map[string]*amazon.Monitor),
		BestbuyMonitors:  make(map[string]*bestbuy.Monitor),
		GamestopMonitors: make(map[string]*gamestop.Monitor),
		HottopicMonitors: make(map[string]*hottopic.Monitor),
		TargetMonitors:   make(map[string]*target.Monitor),
		WalmartMonitors:  make(map[string]*walmart.Monitor),

		EventBus: eventBus,
	}

	go monitorStore.CheckAmazonMonitorStock()
	go monitorStore.CheckBestBuyMonitorStock()
	go monitorStore.CheckGameStopMonitorStock()
	go monitorStore.CheckHotTopicMonitorStock()
	go monitorStore.CheckTargetMonitorStock()
	go monitorStore.CheckWalmartMonitorStock()
}

// GetMonitorStore returns the singleton instance of the EventBus
func GetMonitorStore() *MonitorStore {
	return monitorStore
}
