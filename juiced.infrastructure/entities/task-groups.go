package entities

import (
	"backend.juicedbot.io/juiced.client/http"
	"backend.juicedbot.io/juiced.infrastructure/enums"
	"backend.juicedbot.io/juiced.infrastructure/events"
	"backend.juicedbot.io/juiced.sitescripts/hawk-go"
)

type TaskGroup struct {
	GroupID       string       `json:"groupID" db:"groupID"`
	Name          string       `json:"name"`
	TaskIDs       []string     `json:"-"`
	TaskIDsJoined string       `json:"-" db:"taskIDsJoined"`
	ProxyGroupID  string       `json:"proxyGroupID" db:"proxyGroupID"`
	ProxyGroup    *ProxyGroup  `json:"-"`
	MonitorInfo   string       `json:"monitorInfo" db:"monitorInfo"` // MonitorInfo is Monitor serialized
	Monitor       *BaseMonitor `json:"-"`
	CreationDate  int64        `json:"creationDate" db:"creationDate"`
}

type BaseMonitor struct {
	RetailerMonitor
	Retailer string              `json:"retailer"`
	Quantity int                 `json:"quantity"`
	Status   enums.MonitorStatus `json:"status"`
	DelayMS  int                 `json:"delayMS"`
}

type RetailerMonitor interface {
	RunMonitor()
	GetMonitorInfo() *MonitorInfo
}

type MonitorInfo struct {
	TaskGroup  *TaskGroup
	Proxy      *Proxy
	ProxyGroup *ProxyGroup
	EventBus   *events.EventBus
	Client     http.Client
	Scraper    hawk.Scraper
	StopFlag   bool
	ErrorField string

	InStock []StockInfo
}

func (taskGroup *TaskGroup) PublishEvent(status enums.MonitorStatus, eventType enums.MonitorEventType) {
	monitorInfo := taskGroup.Monitor.GetMonitorInfo()
	if monitorInfo == nil {
		return
	}
	monitorInfo.TaskGroup.Monitor.Status = status
	monitorInfo.EventBus.PublishMonitorEvent(status, eventType, nil, monitorInfo.TaskGroup.GroupID)
}

func (taskGroup *TaskGroup) CheckForStop() bool {
	monitorInfo := taskGroup.Monitor.GetMonitorInfo()
	if monitorInfo == nil {
		return true
	}
	if monitorInfo.StopFlag {
		taskGroup.PublishEvent(enums.MonitorIdle, enums.MonitorStop)
		return true
	}
	return false
}
