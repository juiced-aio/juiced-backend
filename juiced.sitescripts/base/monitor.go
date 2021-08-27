package base

import (
	"errors"

	"backend.juicedbot.io/juiced.infrastructure/common/entities"
	"backend.juicedbot.io/juiced.infrastructure/common/enums"
	"backend.juicedbot.io/juiced.infrastructure/common/events"
	"backend.juicedbot.io/juiced.sitescripts/util"
)

type Monitor struct {
	RetailMonitor
	Retailer enums.Retailer
}

type RetailMonitor interface {
	GetMonitorInfo() *util.MonitorInfo
	RunMonitor()
}

func CreateRetailerMonitor(retailer enums.Retailer, taskGroup *entities.TaskGroup, proxyGroup *entities.ProxyGroup, eventBus *events.EventBus, data interface{}) (Monitor, error) {
	baseMonitor := Monitor{
		Retailer: retailer,
	}
	var err error

	switch retailer {
	case enums.PokemonCenter:
		monitors, ok := data.([]entities.PokemonCenterSingleMonitorInfo)
		if !ok {
			return baseMonitor, errors.New("bad input")
		}
		err = baseMonitor.CreatePokemonCenterMonitor(taskGroup, proxyGroup, eventBus, monitors)
	}

	return baseMonitor, err
}

func (monitor *Monitor) RunMonitor() {
	monitor.RetailMonitor.RunMonitor()
}
