package stores

import (
	"math/rand"

	"go.mongodb.org/mongo-driver/bson/primitive"

	"backend.juicedbot.io/juiced.infrastructure/common/entities"
	"backend.juicedbot.io/juiced.infrastructure/common/enums"
	"backend.juicedbot.io/juiced.infrastructure/common/events"
	"backend.juicedbot.io/juiced.infrastructure/queries"

	"backend.juicedbot.io/juiced.sitescripts/amazon"
	"backend.juicedbot.io/juiced.sitescripts/bestbuy"
	"backend.juicedbot.io/juiced.sitescripts/gamestop"
	"backend.juicedbot.io/juiced.sitescripts/target"
	"backend.juicedbot.io/juiced.sitescripts/walmart"
	// Future sitescripts will be imported here
)

// MonitorStore stores information about running Monitors
type MonitorStore struct {
	TargetMonitors   map[primitive.ObjectID]*target.Monitor
	WalmartMonitors  map[primitive.ObjectID]*walmart.Monitor
	AmazonMonitors   map[primitive.ObjectID]*amazon.Monitor
	BestbuyMonitors  map[primitive.ObjectID]*bestbuy.Monitor
	GamestopMonitors map[primitive.ObjectID]*gamestop.Monitor
	EventBus         *events.EventBus
}

// AddMonitorToStore adds the Monitor to the Store and returns true if successful
func (monitorStore *MonitorStore) AddMonitorToStore(monitor *entities.TaskGroup) bool {
	queryError := false
	// Get Proxy for monitor
	proxy := entities.Proxy{}
	if !monitor.MonitorProxyGroupID.IsZero() {
		proxyGroup, err := queries.GetProxyGroup(monitor.MonitorProxyGroupID)
		if err != nil {
			queryError = true
		}
		proxy = proxyGroup.Proxies[rand.Intn(len(proxyGroup.Proxies))]
	}
	switch monitor.MonitorRetailer {
	// Future sitescripts will have a case here
	case enums.Target:
		// Check if monitor exists in store already
		if _, ok := monitorStore.TargetMonitors[monitor.GroupID]; ok {
			return true
		}
		// Only return false on a query error if the monitor doesn't exist in the store already
		if queryError {
			return false
		}
		// Make sure necessary fields exist
		emptyString := ""
		if monitor.TargetMonitorInfo.MonitorType == emptyString || len(monitor.TargetMonitorInfo.TCINs) == 0 {
			return false
		}
		// Create monitor
		targetMonitor, err := target.CreateTargetMonitor(monitor, proxy, monitorStore.EventBus, monitor.TargetMonitorInfo.MonitorType, monitor.TargetMonitorInfo.TCINs, monitor.TargetMonitorInfo.StoreID)
		if err != nil {
			return false
		}
		// Add task to store
		monitorStore.TargetMonitors[monitor.GroupID] = &targetMonitor

	case enums.Walmart:
		// Check if monitor exists in store already
		if _, ok := monitorStore.WalmartMonitors[monitor.GroupID]; ok {
			return true
		}
		// Only return false on a query error if the monitor doesn't exist in the store already
		if queryError {
			return false
		}
		// Make sure necessary fields exist
		emptyString := ""
		if monitor.WalmartMonitorInfo.MonitorType == emptyString || len(monitor.WalmartMonitorInfo.Products) == 0 {
			return false
		}
		// Create monitor
		walmartMonitor, err := walmart.CreateWalmartMonitor(monitor, proxy, monitorStore.EventBus, monitor.WalmartMonitorInfo.MonitorType, monitor.WalmartMonitorInfo.Products)
		if err != nil {
			return false
		}
		// Add task to store
		monitorStore.WalmartMonitors[monitor.GroupID] = &walmartMonitor

	case enums.Amazon:
		if _, ok := monitorStore.AmazonMonitors[monitor.GroupID]; ok {
			return true
		}

		if queryError {
			return false
		}

		if len(monitor.AmazonMonitorInfo.Monitors) == 0 {
			return false
		}

		amazonMonitor, err := amazon.CreateAmazonMonitor(monitor, proxy, monitorStore.EventBus, monitor.AmazonMonitorInfo.Monitors)
		if err != nil {
			return false
		}

		monitorStore.AmazonMonitors[monitor.GroupID] = &amazonMonitor

	case enums.BestBuy:
		if _, ok := monitorStore.BestbuyMonitors[monitor.GroupID]; ok {
			return true
		}

		if queryError {
			return false
		}

		if len(monitor.BestbuyMonitorInfo.Monitors) == 0 {
			return false
		}

		bestbuyMonitor, err := bestbuy.CreateBestbuyMonitor(monitor, proxy, monitorStore.EventBus, monitor.BestbuyMonitorInfo.Monitors)
		if err != nil {
			return false
		}

		monitorStore.BestbuyMonitors[monitor.GroupID] = &bestbuyMonitor

	case enums.GameStop:
		if _, ok := monitorStore.GamestopMonitors[monitor.GroupID]; ok {
			return true
		}

		if queryError {
			return false
		}

		if len(monitor.GamestopMonitorInfo.Monitors) == 0 {
			return false
		}

		gamestopMonitor, err := gamestop.CreateGamestopMonitor(monitor, proxy, monitorStore.EventBus, monitor.GamestopMonitorInfo.Monitors)
		if err != nil {
			return false
		}

		monitorStore.GamestopMonitors[monitor.GroupID] = &gamestopMonitor

	}
	return true
}

// StartMonitor runs the Run() function for the given Monitor and returns true if successful
func (monitorStore *MonitorStore) StartMonitor(monitor *entities.TaskGroup) bool {
	// Add monitor to store (if it already exists, this will return true)
	added := monitorStore.AddMonitorToStore(monitor)
	if !added {
		return false
	}

	// If the Monitor is already running, then we're all set already
	if monitor.MonitorStatus != enums.MonitorIdle {
		return true
	}

	// Otherwise, start the Monitor
	switch monitor.MonitorRetailer {
	// Future sitescripts will have a case here
	case enums.Target:
		go monitorStore.TargetMonitors[monitor.GroupID].RunMonitor()
	case enums.Walmart:
		go monitorStore.WalmartMonitors[monitor.GroupID].RunMonitor()
	case enums.Amazon:
		go monitorStore.AmazonMonitors[monitor.GroupID].RunMonitor()
	case enums.BestBuy:
		go monitorStore.BestbuyMonitors[monitor.GroupID].RunMonitor()
	case enums.GameStop:
		go monitorStore.GamestopMonitors[monitor.GroupID].RunMonitor()
	}

	return true
}

// StopMonitor sets the stop field for the given Monitor and returns true if successful
func (monitorStore *MonitorStore) StopMonitor(monitor *entities.TaskGroup) bool {
	switch monitor.MonitorRetailer {
	// Future sitescripts will have a case here
	case enums.Target:
		if targetMonitor, ok := monitorStore.TargetMonitors[monitor.GroupID]; ok {
			targetMonitor.Monitor.StopFlag = true
		}
		// Return true if the task doesn't exist
		return true
	case enums.Walmart:
		if walmartMonitor, ok := monitorStore.WalmartMonitors[monitor.GroupID]; ok {
			walmartMonitor.Monitor.StopFlag = true
		}
		// Return true if the task doesn't exist
		return true
	case enums.Amazon:
		if amazonMonitor, ok := monitorStore.AmazonMonitors[monitor.GroupID]; ok {
			amazonMonitor.Monitor.StopFlag = true
			return true
		}
		return true
	case enums.BestBuy:
		if bestbuyMonitor, ok := monitorStore.BestbuyMonitors[monitor.GroupID]; ok {
			bestbuyMonitor.Monitor.StopFlag = true
			return true
		}
		return true
	case enums.GameStop:
		if gamestopMonitor, ok := monitorStore.GamestopMonitors[monitor.GroupID]; ok {
			gamestopMonitor.Monitor.StopFlag = true
			return true
		}
		return true
	}
	return false
}

var monitorStore *MonitorStore

// InitMonitorStore initializes the singleton instance of the Store
func InitMonitorStore(eventBus *events.EventBus) {
	monitorStore = &MonitorStore{
		TargetMonitors:   make(map[primitive.ObjectID]*target.Monitor),
		WalmartMonitors:  make(map[primitive.ObjectID]*walmart.Monitor),
		AmazonMonitors:   make(map[primitive.ObjectID]*amazon.Monitor),
		BestbuyMonitors:  make(map[primitive.ObjectID]*bestbuy.Monitor),
		GamestopMonitors: make(map[primitive.ObjectID]*gamestop.Monitor),
		EventBus:         eventBus,
	}
}

// GetMonitorStore returns the singleton instance of the EventBus
func GetMonitorStore() *MonitorStore {
	return monitorStore
}
