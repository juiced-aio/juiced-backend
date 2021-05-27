package stores

import (
	"math/rand"

	"go.mongodb.org/mongo-driver/bson/primitive"

	"backend.juicedbot.io/m/v2/juiced.infrastructure/common/entities"
	"backend.juicedbot.io/m/v2/juiced.infrastructure/common/enums"
	"backend.juicedbot.io/m/v2/juiced.infrastructure/common/events"
	"backend.juicedbot.io/m/v2/juiced.infrastructure/queries"

	"backend.juicedbot.io/m/v2/juiced.sitescripts/target"
	"backend.juicedbot.io/m/v2/juiced.sitescripts/walmart"
	// Future sitescripts will be imported here
)

// MonitorStore stores information about running Monitors
type MonitorStore struct {
	TargetMonitors  map[primitive.ObjectID]*target.Monitor
	WalmartMonitors map[primitive.ObjectID]*walmart.Monitor
	EventBus        *events.EventBus
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
		if monitor.TargetMonitorInfo.MonitorType == "" || len(monitor.TargetMonitorInfo.TCINs) == 0 {
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
		if monitor.WalmartMonitorInfo.MonitorType == "" || len(monitor.WalmartMonitorInfo.SKUs) == 0 {
			return false
		}
		// Create monitor
		walmartMonitor, err := walmart.CreateWalmartMonitor(monitor, proxy, monitorStore.EventBus, monitor.WalmartMonitorInfo.MonitorType, monitor.WalmartMonitorInfo.SKUs)
		if err != nil {
			return false
		}
		// Add task to store
		monitorStore.WalmartMonitors[monitor.GroupID] = &walmartMonitor
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
		defer func() {
			recover()
			// TODO @silent: Let the UI know that a monitor failed
		}()
		// May panic (if it runs into a runtime error)
		go monitorStore.TargetMonitors[monitor.GroupID].RunMonitor()
	case enums.Walmart:
		go monitorStore.WalmartMonitors[monitor.GroupID].RunMonitor()
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
			return true
		}
		// Return true if the task doesn't exist
		return true
	case enums.Walmart:
		if walmartMonitor, ok := monitorStore.WalmartMonitors[monitor.GroupID]; ok {
			walmartMonitor.Monitor.StopFlag = true
			return true
		}
		// Return true if the task doesn't exist
		return true
	}
	return false
}

var monitorStore *MonitorStore

// InitMonitorStore initializes the singleton instance of the Store
func InitMonitorStore(eventBus *events.EventBus) {
	monitorStore = &MonitorStore{
		TargetMonitors: make(map[primitive.ObjectID]*target.Monitor),
		EventBus:       eventBus,
	}
}

// GetMonitorStore returns the singleton instance of the EventBus
func GetMonitorStore() *MonitorStore {
	return monitorStore
}
