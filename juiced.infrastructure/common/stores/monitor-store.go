package stores

import (
	e "errors"
	"math/rand"
	"strings"
	"time"

	"backend.juicedbot.io/juiced.infrastructure/common"
	"backend.juicedbot.io/juiced.infrastructure/common/entities"
	"backend.juicedbot.io/juiced.infrastructure/common/enums"
	"backend.juicedbot.io/juiced.infrastructure/common/errors"
	"backend.juicedbot.io/juiced.infrastructure/common/events"

	"backend.juicedbot.io/juiced.sitescripts/amazon"
	"backend.juicedbot.io/juiced.sitescripts/bestbuy"
	"backend.juicedbot.io/juiced.sitescripts/boxlunch"
	"backend.juicedbot.io/juiced.sitescripts/disney"
	"backend.juicedbot.io/juiced.sitescripts/gamestop"
	"backend.juicedbot.io/juiced.sitescripts/hottopic"
	"backend.juicedbot.io/juiced.sitescripts/newegg"
	"backend.juicedbot.io/juiced.sitescripts/pokemoncenter"
	"backend.juicedbot.io/juiced.sitescripts/shopify"
	"backend.juicedbot.io/juiced.sitescripts/target"
	"backend.juicedbot.io/juiced.sitescripts/topps"
	"backend.juicedbot.io/juiced.sitescripts/walmart"
	// Future sitescripts will be imported here
)

// MonitorStore stores information about running Monitors
type MonitorStore struct {
	AmazonMonitors        map[string]*amazon.Monitor
	BestbuyMonitors       map[string]*bestbuy.Monitor
	BoxlunchMonitors      map[string]*boxlunch.Monitor
	DisneyMonitors        map[string]*disney.Monitor
	GamestopMonitors      map[string]*gamestop.Monitor
	HottopicMonitors      map[string]*hottopic.Monitor
	NeweggMonitors        map[string]*newegg.Monitor
	PokemonCenterMonitors map[string]*pokemoncenter.Monitor
	ShopifyMonitors       map[string]*shopify.Monitor
	TargetMonitors        map[string]*target.Monitor
	ToppsMonitors         map[string]*topps.Monitor
	WalmartMonitors       map[string]*walmart.Monitor
	EventBus              *events.EventBus
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

	case enums.PokemonCenter:
		if _, ok := monitorStore.PokemonCenterMonitors[monitor.GroupID]; ok && !monitor.UpdateMonitor {
			return nil
		}

		if queryError != nil {
			return queryError
		}

		if len(monitor.PokemonCenterMonitorInfo.Monitors) == 0 {
			return e.New(errors.NoMonitorsError)
		}

		pokemonCenterMonitor, err := pokemoncenter.CreatePokemonCenterMonitor(monitor, proxyGroup, monitorStore.EventBus, monitor.PokemonCenterMonitorInfo.Monitors)
		if err != nil {
			return e.New(errors.CreateMonitorError + err.Error())
		}
		monitorStore.PokemonCenterMonitors[monitor.GroupID] = &pokemonCenterMonitor

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

// AddTestMonitorToStore adds the Test Monitor to the Store and returns true if successful
func (monitorStore *MonitorStore) AddTestMonitorToStore(monitor *entities.TaskGroup) error {
	var queryError error
	// Get Proxy for monitor
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

func (monitorStore *MonitorStore) UpdateMonitor(newMonitor *entities.TaskGroup) error {
	monitor := monitorStore.GetMonitor(newMonitor.MonitorRetailer, newMonitor.GroupID)

	if monitor == nil {
		return e.New("task group not found")
	}

	return monitorStore.AddMonitorToStore(newMonitor)
}

// StartMonitor runs the Run() function for the given Monitor and returns true if successful
func (monitorStore *MonitorStore) StartMonitor(monitor *entities.TaskGroup) error {
	// Add monitor to store (if it already exists, this will return true)
	err := monitorStore.AddMonitorToStore(monitor)
	if err != nil {
		return err
	}

	monitor = monitorStore.GetMonitor(monitor.MonitorRetailer, monitor.GroupID)

	// If the Monitor is already running, then we're all set already
	if !strings.Contains(monitor.MonitorStatus, strings.ReplaceAll(enums.MonitorIdle, " %s", "")) &&
		!strings.Contains(monitor.MonitorStatus, strings.ReplaceAll(enums.MonitorFailed, " %s", "")) {
		return nil
	}

	// Otherwise, start the Monitor
	switch monitor.MonitorRetailer {
	// Future sitescripts will have a case here
	case enums.Amazon:
		amazonMonitor, ok := monitorStore.AmazonMonitors[monitor.GroupID]
		if ok {
			amazonMonitor.InStock = amazonMonitor.InStock[:0]
			amazonMonitor.Monitor.StopFlag = false
			go monitorStore.AmazonMonitors[monitor.GroupID].RunMonitor()
		}

	case enums.BestBuy:
		bestbuyMonitor, ok := monitorStore.BestbuyMonitors[monitor.GroupID]
		if ok {
			bestbuyMonitor.InStock = bestbuyMonitor.InStock[:0]
			bestbuyMonitor.Monitor.StopFlag = false
			go monitorStore.BestbuyMonitors[monitor.GroupID].RunMonitor()
		}

	case enums.BoxLunch:
		boxlunchMonitor, ok := monitorStore.BoxlunchMonitors[monitor.GroupID]
		if ok {
			boxlunchMonitor.InStock = boxlunchMonitor.InStock[:0]
			boxlunchMonitor.Monitor.StopFlag = false
			go monitorStore.BoxlunchMonitors[monitor.GroupID].RunMonitor()
		}

	case enums.Disney:
		disneyMonitor, ok := monitorStore.DisneyMonitors[monitor.GroupID]
		if ok {
			disneyMonitor.InStock = disneyMonitor.InStock[:0]
			disneyMonitor.Monitor.StopFlag = false
			go monitorStore.DisneyMonitors[monitor.GroupID].RunMonitor()
		}

	case enums.GameStop:
		gamestopMonitor, ok := monitorStore.GamestopMonitors[monitor.GroupID]
		if ok {
			gamestopMonitor.InStock = gamestopMonitor.InStock[:0]
			gamestopMonitor.Monitor.StopFlag = false
			go monitorStore.GamestopMonitors[monitor.GroupID].RunMonitor()
		}

	case enums.HotTopic:
		hottopicMonitor, ok := monitorStore.HottopicMonitors[monitor.GroupID]
		if ok {
			hottopicMonitor.InStock = hottopicMonitor.InStock[:0]
			hottopicMonitor.Monitor.StopFlag = false
			go monitorStore.HottopicMonitors[monitor.GroupID].RunMonitor()
		}

	case enums.Newegg:
		neweggMonitor, ok := monitorStore.NeweggMonitors[monitor.GroupID]
		if ok {
			neweggMonitor.InStock = neweggMonitor.InStock[:0]
			neweggMonitor.Monitor.StopFlag = false
			go monitorStore.NeweggMonitors[monitor.GroupID].RunMonitor()
		}

	case enums.PokemonCenter:
		pokemonCenterMonitor, ok := monitorStore.PokemonCenterMonitors[monitor.GroupID]
		if ok {
			pokemonCenterMonitor.InStock = pokemonCenterMonitor.InStock[:0]
			pokemonCenterMonitor.Monitor.StopFlag = false
			go monitorStore.PokemonCenterMonitors[monitor.GroupID].RunMonitor()
		}

	case enums.Shopify:
		shopifyMonitor, ok := monitorStore.ShopifyMonitors[monitor.GroupID]
		if ok {
			shopifyMonitor.InStock = shopifyMonitor.InStock[:0]
			shopifyMonitor.Monitor.StopFlag = false
			go monitorStore.ShopifyMonitors[monitor.GroupID].RunMonitor()
		}

	case enums.Target:
		targetMonitor, ok := monitorStore.TargetMonitors[monitor.GroupID]
		if ok {
			targetMonitor.InStockForShip = targetMonitor.InStockForShip[:0]
			targetMonitor.InStockForPickup = targetMonitor.InStockForPickup[:0]
			targetMonitor.Monitor.StopFlag = false
			go monitorStore.TargetMonitors[monitor.GroupID].RunMonitor()
		}

	case enums.Topps:
		toppsMonitor, ok := monitorStore.ToppsMonitors[monitor.GroupID]
		if ok {
			toppsMonitor.InStock = toppsMonitor.InStock[:0]
			toppsMonitor.Monitor.StopFlag = false
			go monitorStore.ToppsMonitors[monitor.GroupID].RunMonitor()
		}

	case enums.Walmart:
		walmartMonitor, ok := monitorStore.WalmartMonitors[monitor.GroupID]
		if ok {
			walmartMonitor.InStockForShip = walmartMonitor.InStockForShip[:0]
			walmartMonitor.Monitor.StopFlag = false
			go monitorStore.WalmartMonitors[monitor.GroupID].RunMonitor()
		}

	}

	return nil
}

// StartTestMonitor runs the Run() function for the given Monitor and returns true if successful
func (monitorStore *MonitorStore) StartTestMonitor(monitor *entities.TaskGroup) error {
	// Add monitor to store (if it already exists, this will return true)
	err := monitorStore.AddTestMonitorToStore(monitor)
	if err != nil {
		return err
	}

	// If the Monitor is already running, then we're all set already
	if monitor.MonitorStatus != enums.MonitorIdle {
		return nil
	}

	// If the Monitor is already running, then we're all set already
	if !strings.Contains(monitor.MonitorStatus, strings.ReplaceAll(enums.MonitorIdle, " %s", "")) &&
		!strings.Contains(monitor.MonitorStatus, strings.ReplaceAll(enums.MonitorFailed, " %s", "")) {
		return nil
	}

	// Otherwise, start the Monitor
	switch monitor.MonitorRetailer {
	// Future sitescripts will have a case here
	case enums.Amazon:
		amazonMonitor, ok := monitorStore.AmazonMonitors[monitor.GroupID]
		if ok {
			amazonMonitor.InStock = amazonMonitor.InStock[:0]
			amazonMonitor.Monitor.StopFlag = false
			go monitorStore.AmazonMonitors[monitor.GroupID].RunMonitor()
		}

	case enums.BestBuy:
		bestbuyMonitor, ok := monitorStore.BestbuyMonitors[monitor.GroupID]
		if ok {
			bestbuyMonitor.InStock = bestbuyMonitor.InStock[:0]
			bestbuyMonitor.Monitor.StopFlag = false
			go monitorStore.BestbuyMonitors[monitor.GroupID].RunMonitor()
		}

	case enums.BoxLunch:
		boxlunchMonitor, ok := monitorStore.BoxlunchMonitors[monitor.GroupID]
		if ok {
			boxlunchMonitor.InStock = boxlunchMonitor.InStock[:0]
			boxlunchMonitor.Monitor.StopFlag = false
			go monitorStore.BoxlunchMonitors[monitor.GroupID].RunMonitor()
		}

	case enums.Disney:
		disneyMonitor, ok := monitorStore.DisneyMonitors[monitor.GroupID]
		if ok {
			disneyMonitor.InStock = disneyMonitor.InStock[:0]
			disneyMonitor.Monitor.StopFlag = false
			go monitorStore.DisneyMonitors[monitor.GroupID].RunMonitor()
		}

	case enums.GameStop:
		gamestopMonitor, ok := monitorStore.GamestopMonitors[monitor.GroupID]
		if ok {
			gamestopMonitor.InStock = gamestopMonitor.InStock[:0]
			gamestopMonitor.Monitor.StopFlag = false
			go monitorStore.GamestopMonitors[monitor.GroupID].RunMonitor()
		}

	case enums.HotTopic:
		hottopicMonitor, ok := monitorStore.HottopicMonitors[monitor.GroupID]
		if ok {
			hottopicMonitor.InStock = hottopicMonitor.InStock[:0]
			hottopicMonitor.Monitor.StopFlag = false
			go monitorStore.HottopicMonitors[monitor.GroupID].RunMonitor()
		}

	case enums.Newegg:
		neweggMonitor, ok := monitorStore.NeweggMonitors[monitor.GroupID]
		if ok {
			neweggMonitor.InStock = neweggMonitor.InStock[:0]
			neweggMonitor.Monitor.StopFlag = false
			go monitorStore.NeweggMonitors[monitor.GroupID].RunMonitor()
		}

	case enums.PokemonCenter:
		pokemonCenterMonitor, ok := monitorStore.PokemonCenterMonitors[monitor.GroupID]
		if ok {
			pokemonCenterMonitor.InStock = pokemonCenterMonitor.InStock[:0]
			pokemonCenterMonitor.Monitor.StopFlag = false
			go monitorStore.PokemonCenterMonitors[monitor.GroupID].RunMonitor()
		}

	case enums.Shopify:
		shopifyMonitor, ok := monitorStore.ShopifyMonitors[monitor.GroupID]
		if ok {
			shopifyMonitor.InStock = shopifyMonitor.InStock[:0]
			shopifyMonitor.Monitor.StopFlag = false
			go monitorStore.ShopifyMonitors[monitor.GroupID].RunMonitor()
		}

	case enums.Target:
		targetMonitor, ok := monitorStore.TargetMonitors[monitor.GroupID]
		if ok {
			targetMonitor.InStockForShip = targetMonitor.InStockForShip[:0]
			targetMonitor.InStockForPickup = targetMonitor.InStockForPickup[:0]
			targetMonitor.Monitor.StopFlag = false
			go monitorStore.TargetMonitors[monitor.GroupID].RunMonitor()
		}

	case enums.Topps:
		toppsMonitor, ok := monitorStore.ToppsMonitors[monitor.GroupID]
		if ok {
			toppsMonitor.InStock = toppsMonitor.InStock[:0]
			toppsMonitor.Monitor.StopFlag = false
			go monitorStore.ToppsMonitors[monitor.GroupID].RunMonitor()
		}

	case enums.Walmart:
		walmartMonitor, ok := monitorStore.WalmartMonitors[monitor.GroupID]
		if ok {
			walmartMonitor.InStockForShip = walmartMonitor.InStockForShip[:0]
			walmartMonitor.Monitor.StopFlag = false
			go monitorStore.WalmartMonitors[monitor.GroupID].RunMonitor()
		}

	}

	return nil
}

// StopMonitor sets the stop field for the given Monitor and returns true if successful
func (monitorStore *MonitorStore) StopMonitor(monitor *entities.TaskGroup) (bool, error) {
	wasRunning := false
	switch monitor.MonitorRetailer {
	// Future sitescripts will have a case here
	case enums.Amazon:
		if amazonMonitor, ok := monitorStore.AmazonMonitors[monitor.GroupID]; ok {
			if !amazonMonitor.Monitor.StopFlag {
				wasRunning = true
			}
			amazonMonitor.Monitor.StopFlag = true
		}

	case enums.BestBuy:
		if bestbuyMonitor, ok := monitorStore.BestbuyMonitors[monitor.GroupID]; ok {
			if !bestbuyMonitor.Monitor.StopFlag {
				wasRunning = true
			}
			bestbuyMonitor.Monitor.StopFlag = true
		}

	case enums.BoxLunch:
		if boxlunchMonitor, ok := monitorStore.BoxlunchMonitors[monitor.GroupID]; ok {
			if !boxlunchMonitor.Monitor.StopFlag {
				wasRunning = true
			}
			boxlunchMonitor.Monitor.StopFlag = true
		}

	case enums.Disney:
		if disneyMonitor, ok := monitorStore.DisneyMonitors[monitor.GroupID]; ok {
			if !disneyMonitor.Monitor.StopFlag {
				wasRunning = true
			}
			disneyMonitor.Monitor.StopFlag = true
		}

	case enums.GameStop:
		if gamestopMonitor, ok := monitorStore.GamestopMonitors[monitor.GroupID]; ok {
			if !gamestopMonitor.Monitor.StopFlag {
				wasRunning = true
			}
			gamestopMonitor.Monitor.StopFlag = true
		}

	case enums.HotTopic:
		if hottopicMonitor, ok := monitorStore.HottopicMonitors[monitor.GroupID]; ok {
			if !hottopicMonitor.Monitor.StopFlag {
				wasRunning = true
			}
			hottopicMonitor.Monitor.StopFlag = true
		}

	case enums.Newegg:
		if neweggMonitor, ok := monitorStore.NeweggMonitors[monitor.GroupID]; ok {
			if !neweggMonitor.Monitor.StopFlag {
				wasRunning = true
			}
			neweggMonitor.Monitor.StopFlag = true
		}

	case enums.PokemonCenter:
		if pokemonCenterMonitor, ok := monitorStore.PokemonCenterMonitors[monitor.GroupID]; ok {
			if !pokemonCenterMonitor.Monitor.StopFlag {
				wasRunning = true
			}
			pokemonCenterMonitor.Monitor.StopFlag = true
		}

	case enums.Shopify:
		if shopifyMonitor, ok := monitorStore.ShopifyMonitors[monitor.GroupID]; ok {
			if !shopifyMonitor.Monitor.StopFlag {
				wasRunning = true
			}
			shopifyMonitor.Monitor.StopFlag = true
		}

	case enums.Target:
		if targetMonitor, ok := monitorStore.TargetMonitors[monitor.GroupID]; ok {
			if !targetMonitor.Monitor.StopFlag {
				wasRunning = true
			}
			targetMonitor.Monitor.StopFlag = true
		}

	case enums.Topps:
		if toppsMonitor, ok := monitorStore.ToppsMonitors[monitor.GroupID]; ok {
			if !toppsMonitor.Monitor.StopFlag {
				wasRunning = true
			}
			toppsMonitor.Monitor.StopFlag = true
		}

	case enums.Walmart:
		if walmartMonitor, ok := monitorStore.WalmartMonitors[monitor.GroupID]; ok {
			if !walmartMonitor.Monitor.StopFlag {
				wasRunning = true
			}
			walmartMonitor.Monitor.StopFlag = true
		}

	default:
		return false, e.New(errors.InvalidMonitorRetailerError)
	}
	return wasRunning, nil
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
			boxlunchMonitor.Monitor.Proxy = proxy
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

	case enums.PokemonCenter:
		if pokemonCenterMonitor, ok := monitorStore.PokemonCenterMonitors[monitor.GroupID]; ok {
			pokemonCenterMonitor.Monitor.Proxy = proxy
		}
		return true

	case enums.Shopify:
		if shopifyMonitor, ok := monitorStore.ShopifyMonitors[monitor.GroupID]; ok {
			shopifyMonitor.Monitor.Proxy = proxy
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

func (monitorStore *MonitorStore) CheckAmazonMonitorStock() {
	for {
		for monitorID, amazonMonitor := range monitorStore.AmazonMonitors {
			if !amazonMonitor.Monitor.StopFlag && len(amazonMonitor.InStock) > 0 {
				taskGroup := amazonMonitor.Monitor.TaskGroup
				for _, taskID := range taskGroup.TaskIDs {
					if amazonTask, ok := taskStore.AmazonTasks[taskID]; ok && !amazonTask.Task.HasStockData {
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
			if !bestbuyMonitor.Monitor.StopFlag && len(bestbuyMonitor.InStock) > 0 {
				taskGroup := bestbuyMonitor.Monitor.TaskGroup
				for _, taskID := range taskGroup.TaskIDs {
					if bestbuyTask, ok := taskStore.BestbuyTasks[taskID]; ok && !bestbuyTask.Task.HasStockData {
						if ok && bestbuyTask.Task.Task.TaskGroupID == monitorID {
							randomNumber := rand.Intn(len(bestbuyMonitor.InStock))
							bestbuyTask.StockData = bestbuyMonitor.InStock[randomNumber]
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
			if !boxlunchMonitor.Monitor.StopFlag && len(boxlunchMonitor.InStock) > 0 {
				taskGroup := boxlunchMonitor.Monitor.TaskGroup
				for _, taskID := range taskGroup.TaskIDs {
					if boxlunchTask, ok := taskStore.BoxlunchTasks[taskID]; ok && !boxlunchTask.Task.HasStockData {
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

func (monitorStore *MonitorStore) CheckDisneyMonitorStock() {
	for {
		for monitorID, disneyMonitor := range monitorStore.DisneyMonitors {
			if !disneyMonitor.Monitor.StopFlag && len(disneyMonitor.InStock) > 0 {
				taskGroup := disneyMonitor.Monitor.TaskGroup
				for _, taskID := range taskGroup.TaskIDs {
					if disneyTask, ok := taskStore.DisneyTasks[taskID]; ok && !disneyTask.Task.HasStockData {
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

func (monitorStore *MonitorStore) CheckGameStopMonitorStock() {
	for {
		for monitorID, gamestopMonitor := range monitorStore.GamestopMonitors {
			if !gamestopMonitor.Monitor.StopFlag && len(gamestopMonitor.InStock) > 0 {
				taskGroup := gamestopMonitor.Monitor.TaskGroup
				for _, taskID := range taskGroup.TaskIDs {
					if gamestopTask, ok := taskStore.GamestopTasks[taskID]; ok && !gamestopTask.Task.HasStockData {
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
			if !hottopicMonitor.Monitor.StopFlag && len(hottopicMonitor.InStock) > 0 {
				taskGroup := hottopicMonitor.Monitor.TaskGroup
				for _, taskID := range taskGroup.TaskIDs {
					if hottopicTask, ok := taskStore.HottopicTasks[taskID]; ok && !hottopicTask.Task.HasStockData {
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
			if !neweggMonitor.Monitor.StopFlag && len(neweggMonitor.InStock) > 0 {
				taskGroup := neweggMonitor.Monitor.TaskGroup
				for _, taskID := range taskGroup.TaskIDs {
					if neweggTask, ok := taskStore.NeweggTasks[taskID]; ok && !neweggTask.Task.HasStockData {
						if ok && neweggTask.Task.Task.TaskGroupID == monitorID {
							neweggTask.StockData = neweggMonitor.InStock[rand.Intn(len(neweggMonitor.InStock))]
						}
					}
				}
			}
		}
		time.Sleep(common.MS_TO_WAIT)
	}
}

func (monitorStore *MonitorStore) CheckPokemonCenterMonitorStock() {
	for {
		for monitorID, pokemonCenterMonitor := range monitorStore.PokemonCenterMonitors {
			if !pokemonCenterMonitor.Monitor.StopFlag && len(pokemonCenterMonitor.InStock) > 0 {
				taskGroup := pokemonCenterMonitor.Monitor.TaskGroup
				for _, taskID := range taskGroup.TaskIDs {
					if pokemonCenterTask, ok := taskStore.PokemonCenterTasks[taskID]; ok && !pokemonCenterTask.Task.HasStockData {
						if ok && pokemonCenterTask.Task.Task.TaskGroupID == monitorID {
							randomNumber := rand.Intn(len(pokemonCenterMonitor.InStock))
							pokemonCenterTask.StockData = pokemonCenterMonitor.InStock[randomNumber]
						}
					}
				}
			}
		}
		time.Sleep(common.MS_TO_WAIT)
	}
}

func (monitorStore *MonitorStore) CheckShopifyMonitorStock() {
	for {
		for monitorID, shopifyMonitor := range monitorStore.ShopifyMonitors {
			if !shopifyMonitor.Monitor.StopFlag && len(shopifyMonitor.InStock) > 0 {
				taskGroup := shopifyMonitor.Monitor.TaskGroup
				for _, taskID := range taskGroup.TaskIDs {
					if shopifyTask, ok := taskStore.ShopifyTasks[taskID]; ok && !shopifyTask.Task.HasStockData {
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

func (monitorStore *MonitorStore) CheckTargetMonitorStock() {
	for {
		for monitorID, targetMonitor := range monitorStore.TargetMonitors {
			if !targetMonitor.Monitor.StopFlag && len(targetMonitor.InStockForPickup) > 0 || len(targetMonitor.InStockForShip) > 0 {
				taskGroup := targetMonitor.Monitor.TaskGroup
				for _, taskID := range taskGroup.TaskIDs {
					if targetTask, ok := taskStore.TargetTasks[taskID]; ok && !targetTask.Task.HasStockData {
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
			if !toppsMonitor.Monitor.StopFlag && len(toppsMonitor.InStock) > 0 {
				taskGroup := toppsMonitor.Monitor.TaskGroup
				for _, taskID := range taskGroup.TaskIDs {
					if toppsTask, ok := taskStore.ToppsTasks[taskID]; ok && !toppsTask.Task.HasStockData {
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
			if !walmartMonitor.Monitor.StopFlag && len(walmartMonitor.InStockForShip) > 0 {
				taskGroup := walmartMonitor.Monitor.TaskGroup
				for _, taskID := range taskGroup.TaskIDs {
					if walmartTask, ok := taskStore.WalmartTasks[taskID]; ok && !walmartTask.Task.HasStockData {
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
		AmazonMonitors:        make(map[string]*amazon.Monitor),
		BestbuyMonitors:       make(map[string]*bestbuy.Monitor),
		BoxlunchMonitors:      make(map[string]*boxlunch.Monitor),
		DisneyMonitors:        make(map[string]*disney.Monitor),
		GamestopMonitors:      make(map[string]*gamestop.Monitor),
		HottopicMonitors:      make(map[string]*hottopic.Monitor),
		ShopifyMonitors:       make(map[string]*shopify.Monitor),
		NeweggMonitors:        make(map[string]*newegg.Monitor),
		PokemonCenterMonitors: make(map[string]*pokemoncenter.Monitor),
		TargetMonitors:        make(map[string]*target.Monitor),
		ToppsMonitors:         make(map[string]*topps.Monitor),
		WalmartMonitors:       make(map[string]*walmart.Monitor),

		EventBus: eventBus,
	}

	go monitorStore.CheckAmazonMonitorStock()
	go monitorStore.CheckBestBuyMonitorStock()
	go monitorStore.CheckBoxlunchMonitorStock()
	go monitorStore.CheckDisneyMonitorStock()
	go monitorStore.CheckGameStopMonitorStock()
	go monitorStore.CheckHotTopicMonitorStock()
	go monitorStore.CheckNeweggMonitorStock()
	go monitorStore.CheckPokemonCenterMonitorStock()
	go monitorStore.CheckShopifyMonitorStock()
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
	if monitor, ok := monitorStore.PokemonCenterMonitors[groupID]; ok {
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

func (monitorStore *MonitorStore) CheckMonitorTasksRunning() {
	for _, amazonMonitor := range monitorStore.AmazonMonitors {
		if !taskStore.TasksRunning(amazonMonitor.Monitor.TaskGroup.TaskIDs, amazonMonitor.Monitor.TaskGroup.MonitorRetailer) {
			monitorStore.StopMonitor(amazonMonitor.Monitor.TaskGroup)
		}
	}

	for _, bestbuyMonitor := range monitorStore.BestbuyMonitors {
		if !taskStore.TasksRunning(bestbuyMonitor.Monitor.TaskGroup.TaskIDs, bestbuyMonitor.Monitor.TaskGroup.MonitorRetailer) {
			monitorStore.StopMonitor(bestbuyMonitor.Monitor.TaskGroup)
		}
	}

	for _, boxlunchMonitor := range monitorStore.BoxlunchMonitors {
		if !taskStore.TasksRunning(boxlunchMonitor.Monitor.TaskGroup.TaskIDs, boxlunchMonitor.Monitor.TaskGroup.MonitorRetailer) {
			monitorStore.StopMonitor(boxlunchMonitor.Monitor.TaskGroup)
		}
	}

	for _, disneyMonitor := range monitorStore.DisneyMonitors {
		if !taskStore.TasksRunning(disneyMonitor.Monitor.TaskGroup.TaskIDs, disneyMonitor.Monitor.TaskGroup.MonitorRetailer) {
			monitorStore.StopMonitor(disneyMonitor.Monitor.TaskGroup)
		}
	}

	for _, gamestopMonitor := range monitorStore.GamestopMonitors {
		if !taskStore.TasksRunning(gamestopMonitor.Monitor.TaskGroup.TaskIDs, gamestopMonitor.Monitor.TaskGroup.MonitorRetailer) {
			monitorStore.StopMonitor(gamestopMonitor.Monitor.TaskGroup)
		}
	}

	for _, hottopicMonitor := range monitorStore.HottopicMonitors {
		if !taskStore.TasksRunning(hottopicMonitor.Monitor.TaskGroup.TaskIDs, hottopicMonitor.Monitor.TaskGroup.MonitorRetailer) {
			monitorStore.StopMonitor(hottopicMonitor.Monitor.TaskGroup)
		}
	}

	for _, neweggMonitor := range monitorStore.NeweggMonitors {
		if !taskStore.TasksRunning(neweggMonitor.Monitor.TaskGroup.TaskIDs, neweggMonitor.Monitor.TaskGroup.MonitorRetailer) {
			monitorStore.StopMonitor(neweggMonitor.Monitor.TaskGroup)
		}
	}

	for _, pokemonCenterMonitor := range monitorStore.PokemonCenterMonitors {
		if !taskStore.TasksRunning(pokemonCenterMonitor.Monitor.TaskGroup.TaskIDs, enums.PokemonCenter) {
			monitorStore.StopMonitor(pokemonCenterMonitor.Monitor.TaskGroup)
		}
	}

	for _, shopifyMonitor := range monitorStore.ShopifyMonitors {
		if !taskStore.TasksRunning(shopifyMonitor.Monitor.TaskGroup.TaskIDs, shopifyMonitor.Monitor.TaskGroup.MonitorRetailer) {
			monitorStore.StopMonitor(shopifyMonitor.Monitor.TaskGroup)
		}
	}

	for _, targetMonitor := range monitorStore.TargetMonitors {
		if !taskStore.TasksRunning(targetMonitor.Monitor.TaskGroup.TaskIDs, targetMonitor.Monitor.TaskGroup.MonitorRetailer) {
			monitorStore.StopMonitor(targetMonitor.Monitor.TaskGroup)
		}
	}

	for _, toppsMonitor := range monitorStore.ToppsMonitors {
		if !taskStore.TasksRunning(toppsMonitor.Monitor.TaskGroup.TaskIDs, toppsMonitor.Monitor.TaskGroup.MonitorRetailer) {
			monitorStore.StopMonitor(toppsMonitor.Monitor.TaskGroup)
		}
	}

	for _, walmartMonitor := range monitorStore.WalmartMonitors {
		if !taskStore.TasksRunning(walmartMonitor.Monitor.TaskGroup.TaskIDs, walmartMonitor.Monitor.TaskGroup.MonitorRetailer) {
			monitorStore.StopMonitor(walmartMonitor.Monitor.TaskGroup)
		}
	}
}

// GetMonitorStore returns the singleton instance of the EventBus
func GetMonitorStore() *MonitorStore {
	return monitorStore
}
