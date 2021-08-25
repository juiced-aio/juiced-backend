package stores

import (
	"testing"
	"time"

	"backend.juicedbot.io/juiced.infrastructure/common/entities"
	"backend.juicedbot.io/juiced.infrastructure/common/enums"
	"backend.juicedbot.io/juiced.infrastructure/common/events"
	"backend.juicedbot.io/juiced.sitescripts/amazon"
	"backend.juicedbot.io/juiced.sitescripts/base"
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
)

var amazonMonitorAsset = map[string]*amazon.Monitor{"amazon_test_monitor": {Monitor: monitorAsset, ASINs: []string{"B08V7GT6F3"}}}

var amazonMonitorInfoAsset = &entities.AmazonMonitorInfo{
	Monitors: []entities.AmazonSingleMonitorInfo{{
		MonitorType: enums.SlowSKUMonitor,
		ASIN:        "B08V7GT6F3",
		MaxPrice:    -1,
	}},
}

var bestbuyMonitorAsset = map[string]*bestbuy.Monitor{"bestbuy_test_monitor": {Monitor: monitorAsset, SKUs: []string{"6457447"}}}

var bestbuyMonitorInfoAsset = &entities.BestbuyMonitorInfo{
	Monitors: []entities.BestbuySingleMonitorInfo{
		{
			SKU:      "6457447",
			MaxPrice: -1,
		},
	},
}

var boxlunchMonitorAsset = map[string]*boxlunch.Monitor{"boxlunch_test_monitor": {Monitor: monitorAsset, Pids: []string{""}}}

var boxlunchMonitorInfoAsset = &entities.BoxlunchMonitorInfo{
	Monitors: []entities.BoxlunchSingleMonitorInfo{
		{
			Pid:      "",
			MaxPrice: -1,
		},
	},
}

var disneyMonitorAsset = map[string]*disney.Monitor{"disney_test_monitor": {Monitor: monitorAsset, Pids: []string{"5813057814019M"}}}

var disneyMonitorInfoAsset = &entities.DisneyMonitorInfo{
	Monitors: []entities.DisneySingleMonitorInfo{
		{
			PID:  "5813057814019M",
			Size: "2",
		},
	},
}

var gamestopMonitorAsset = map[string]*gamestop.Monitor{"gamestop_test_monitor": {Monitor: monitorAsset, SKUs: []string{"11105919"}}}

var gamestopMonitorInfoAsset = &entities.GamestopMonitorInfo{
	Monitors: []entities.GamestopSingleMonitorInfo{
		{
			SKU:      "11105919",
			MaxPrice: -1,
		},
	},
}

var hottopicMonitorAsset = map[string]*hottopic.Monitor{"hottopic_test_monitor": {Monitor: monitorAsset, Pids: []string{"16078565"}}}

var hottopicMonitorInfoAsset = &entities.HottopicMonitorInfo{
	Monitors: []entities.HottopicSingleMonitorInfo{
		{
			Pid:         "11105919",
			Size:        "SM",
			Color:       "RED",
			MaxPrice:    -1,
			MonitorType: enums.SKUMonitor,
		},
	},
}

var neweggMonitorAsset = map[string]*newegg.Monitor{"newegg_test_monitor": {Monitor: monitorAsset, SKUs: []string{"N82E16820147790"}}}

var neweggMonitorInfoAsset = &entities.NeweggMonitorInfo{
	Monitors: []entities.NeweggSingleMonitorInfo{
		{
			SKU:      "N82E16820147790",
			MaxPrice: -1,
		},
	},
}

var shopifyMonitorAsset = map[string]*shopify.Monitor{"shopify_test_monitor": {Monitor: monitorAsset, VIDs: []string{""}}}

var shopifyMonitorInfoAsset = &entities.ShopifyMonitorInfo{
	Monitors: []entities.ShopifySingleMonitorInfo{
		{
			VariantID: "",
			MaxPrice:  -1,
		},
	},
}

var targetMonitorsAsset = map[string]*target.Monitor{"target_test_monitor": {Monitor: monitorAsset, MonitorType: "SKU_MONITOR", TCINs: []string{"81622440"}, StoreID: "1120"}}

var targetMonitorInfoAsset = &entities.TargetMonitorInfo{
	Monitors: []entities.TargetSingleMonitorInfo{
		{
			TCIN:         "81622440",
			MaxPrice:     -1,
			CheckoutType: enums.CheckoutTypeSHIP,
		},
	},
	StoreID:     "1120",
	MonitorType: enums.SKUMonitor,
}

var toppsMonitorAsset = map[string]*topps.Monitor{"topps_test_monitor": {Monitor: monitorAsset, Items: []string{"on-card-auto-to-25-greg-maddux-2021-mlb-topps-now-reg-turn-back-the-clock-card-134b"}}}

var toppsMonitorInfoAsset = &entities.ToppsMonitorInfo{
	Monitors: []entities.ToppsSingleMonitorInfo{
		{
			Item:     "on-card-auto-to-25-greg-maddux-2021-mlb-topps-now-reg-turn-back-the-clock-card-134b",
			MaxPrice: -1,
		},
	},
}

var walmartMonitorAsset = map[string]*walmart.Monitor{"walmart_test_monitor": {Monitor: monitorAsset, IDs: []string{"544900177"}}}

var walmartMonitorInfoAsset = &entities.WalmartMonitorInfo{
	Monitors: []entities.WalmartSingleMonitorInfo{
		{
			ID:          "134448590",
			MaxPrice:    -1,
			MonitorType: enums.SKUMonitor,
		},
	},
}

var pokemonCenterMonitorAsset = map[string]*pokemoncenter.Monitor{"pokemoncenter_test_monitor": {Monitor: monitorAsset, SKUs: []string{""}}}

var pokemonCenterMonitorInfoAsset = &entities.PokemonCenterMonitorInfo{
	Monitors: []entities.PokemonCenterSingleMonitorInfo{
		{
			SKU:      "",
			MaxPrice: -1,
		},
	},
}

var taskgroupAsset = entities.TaskGroup{
	GroupID:             "",
	Name:                "test_taskgroup",
	MonitorProxyGroupID: "",
	MonitorInput:        "",
	MonitorDelay:        2000,
	MonitorStatus:       enums.MonitorIdle,
	CreationDate:        time.Now().Unix(),
}

var monitorAsset = base.Monitor{
	TaskGroup: &taskgroupAsset,
	Proxy: &entities.Proxy{
		Host: "localhost",
		Port: "3000",
	},
}

func TestMain(m *testing.M) {
	events.InitEventBus()
	eventBus := events.GetEventBus()
	monitorAsset.EventBus = eventBus
	m.Run()
}

//	There is some type of memory problem here and Gamestops BecomeGuest helper function seems to be the biggest cause but the error still happens most of the time
func TestStartMonitor(t *testing.T) {
	type args struct {
		monitor *entities.TaskGroup
	}
	tests := []struct {
		name         string
		retailer     enums.Retailer
		monitorStore *MonitorStore
		args         args
		want         error
	}{
		//{"amazon_test", enums.Amazon, &MonitorStore{}, args{&taskgroupAsset}, nil},
		{"bestbuy_test", enums.BestBuy, &MonitorStore{}, args{&taskgroupAsset}, nil},
		{"disney_test", enums.Disney, &MonitorStore{}, args{&taskgroupAsset}, nil},
		{"gamestop_test", enums.GameStop, &MonitorStore{}, args{&taskgroupAsset}, nil},
		{"hottopic_test", enums.HotTopic, &MonitorStore{}, args{&taskgroupAsset}, nil},
		{"newegg_test", enums.Newegg, &MonitorStore{}, args{&taskgroupAsset}, nil},
		{"shopify_test", enums.Shopify, &MonitorStore{}, args{&taskgroupAsset}, nil},
		{"target_test", enums.Target, &MonitorStore{}, args{&taskgroupAsset}, nil},
		{"topps_test", enums.Topps, &MonitorStore{}, args{&taskgroupAsset}, nil},
		{"walmart_test", enums.Walmart, &MonitorStore{}, args{&taskgroupAsset}, nil},
		{"pokemoncenter_test", enums.PokemonCenter, &MonitorStore{}, args{&taskgroupAsset}, nil},
	}

	for _, tt := range tests {
		tt.args.monitor.MonitorRetailer = tt.retailer
		switch tt.retailer {
		case enums.Amazon:
			tt.monitorStore.AmazonMonitors = amazonMonitorAsset
			tt.args.monitor.AmazonMonitorInfo = amazonMonitorInfoAsset
		case enums.BestBuy:
			tt.monitorStore.BestbuyMonitors = bestbuyMonitorAsset
			tt.args.monitor.BestbuyMonitorInfo = bestbuyMonitorInfoAsset
		case enums.BoxLunch:
			tt.monitorStore.BoxlunchMonitors = boxlunchMonitorAsset
			tt.args.monitor.BoxlunchMonitorInfo = boxlunchMonitorInfoAsset
		case enums.Disney:
			tt.monitorStore.DisneyMonitors = disneyMonitorAsset
			tt.args.monitor.DisneyMonitorInfo = disneyMonitorInfoAsset
		case enums.GameStop:
			tt.monitorStore.GamestopMonitors = gamestopMonitorAsset
			tt.args.monitor.GamestopMonitorInfo = gamestopMonitorInfoAsset
		case enums.HotTopic:
			tt.monitorStore.HottopicMonitors = hottopicMonitorAsset
			tt.args.monitor.HottopicMonitorInfo = hottopicMonitorInfoAsset
		case enums.Newegg:
			tt.monitorStore.NeweggMonitors = neweggMonitorAsset
			tt.args.monitor.NeweggMonitorInfo = neweggMonitorInfoAsset
		case enums.Shopify:
			tt.monitorStore.ShopifyMonitors = shopifyMonitorAsset
			tt.args.monitor.ShopifyMonitorInfo = shopifyMonitorInfoAsset
		case enums.Target:
			tt.monitorStore.TargetMonitors = targetMonitorsAsset
			tt.args.monitor.TargetMonitorInfo = targetMonitorInfoAsset
		case enums.Topps:
			tt.monitorStore.ToppsMonitors = toppsMonitorAsset
			tt.args.monitor.ToppsMonitorInfo = toppsMonitorInfoAsset
		case enums.Walmart:
			tt.monitorStore.WalmartMonitors = walmartMonitorAsset
			tt.args.monitor.WalmartMonitorInfo = walmartMonitorInfoAsset
		case enums.PokemonCenter:
			tt.monitorStore.PokemonCenterMonitors = pokemonCenterMonitorAsset
			tt.args.monitor.PokemonCenterMonitorInfo = pokemonCenterMonitorInfoAsset
		}
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.monitorStore.StartMonitor(tt.args.monitor); got != tt.want {
				t.Errorf("MonitorStore.StartMonitor() = %v, want %v", got, tt.want)
			}
		})
	}
}
