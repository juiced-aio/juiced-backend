package hottopic

import (
	"backend.juicedbot.io/m/v2/juiced.infrastructure/common/entities"
	"backend.juicedbot.io/m/v2/juiced.infrastructure/common/enums"
	"backend.juicedbot.io/m/v2/juiced.infrastructure/common/events"
	"backend.juicedbot.io/m/v2/juiced.sitescripts/base"
	"backend.juicedbot.io/m/v2/juiced.sitescripts/util"
)

// CreateHottopicMonitor takes a TaskGroup entity and turns it into a Hottopic Monitor
func CreateHottopicMonitor(taskGroup *entities.TaskGroup, proxy entities.Proxy, eventBus *events.EventBus, singleMonitors []entities.HottopicSingleMonitorInfo) (Monitor, error) {
	storedHottopicMonitors := make(map[string]entities.HottopicSingleMonitorInfo)
	hottopicMonitor := Monitor{}

	pids := []string{}
	for _, monitor := range singleMonitors {
		client, err := util.CreateClient(proxy)
		if err != nil {
			return hottopicMonitor, err
		}
		storedHottopicMonitors[monitor.Pid] = entities.HottopicSingleMonitorInfo{
			Pid:    monitor.Pid,
			Client: client,
		}
		pids = append(pids, monitor.Pid)
		for created := false; !created; {
			hottopicMonitor = Monitor{
				Monitor: base.Monitor{
					TaskGroup: taskGroup,
					Proxy:     proxy,
					EventBus:  eventBus,
				},
				Pids: pids,
			}
			created = true
		}
	}
	return hottopicMonitor, nil
}

// PublishEvent wraps the EventBus's PublishMonitorEvent function
func (monitor *Monitor) PublishEvent(status enums.MonitorStatus, eventType enums.MonitorEventType) {
	monitor.Monitor.TaskGroup.SetMonitorStatus(status)
	monitor.Monitor.EventBus.PublishMonitorEvent(status, eventType, nil, monitor.Monitor.TaskGroup.GroupID)
}

func (monitor *Monitor) CheckForStop() bool {
	if monitor.Monitor.StopFlag {
		monitor.PublishEvent(enums.MonitorIdle, enums.MonitorStop)
		return true
	}
	return false
}

func (monitor *Monitor) RunMonitor() {
	if monitor.Monitor.TaskGroup.MonitorStatus == enums.MonitorIdle {
		monitor.PublishEvent(enums.WaitingForProductData, enums.MonitorStart)
	}
	needToStop := monitor.CheckForStop()
	if needToStop {
		return
	}
	for _, pid := range monitor.Pids {
		somethingInStock := false
		switch monitor.PidWithInfo[pid].MonitorType {
		case enums.SKUMonitor:
			somethingInStock = monitor.isInStock(pid)
		}
		if somethingInStock {
			needToStop := monitor.CheckForStop()
			if needToStop {
				return
			}
			monitor.RunningMonitors = util.RemoveFromSlice(monitor.RunningMonitors, pid)
			monitor.PublishEvent(enums.SendingProductInfoToTasks, enums.MonitorUpdate)
			monitor.SendToTasks()
		}

	}
}

func (monitor *Monitor) isInStock(pid string) bool {
	return true
}

func (monitor *Monitor) SendToTasks() {
	data := events.HottopicStockData{
		InStock: []events.HotTopicSingleStockData{monitor.EventInfo},
	}
	monitor.Monitor.EventBus.PublishProductEvent(enums.Hottopic, data, monitor.Monitor.TaskGroup.GroupID)
}
