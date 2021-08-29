package entities

import (
	"time"

	"backend.juicedbot.io/juiced.client/client"
	"backend.juicedbot.io/juiced.client/http"
	"backend.juicedbot.io/juiced.infrastructure/enums"
	"backend.juicedbot.io/juiced.infrastructure/events"
	"backend.juicedbot.io/juiced.sitescripts/hawk-go"
)

type TaskGroup struct {
	GroupID            string         `json:"groupID" db:"groupID"`
	Name               string         `json:"name"`
	Retailer           string         `json:"retailer" db:"retailer"`
	TaskIDs            []string       `json:"-"`
	TaskIDsSerialized  string         `json:"-" db:"taskIDsSerialized"`
	MonitorsSerialized string         `json:"-" db:"monitorsSerialized"`
	Monitors           []*BaseMonitor `json:"monitors"`
	CreationDate       int64          `json:"creationDate" db:"creationDate"`
	Tasks              []*Task        `json:"tasks"`
}

type BaseMonitor struct {
	Monitor *Monitor

	// Monitor inputs, included in DB serialization
	MonitorInput MonitorInput `json:"monitorInput"`

	// In-memory values, omitted in DB serialization but included in JSON
	Status      enums.MonitorStatus `json:"status"`
	Running     bool                `json:"running"`
	ProductInfo ProductInfo         `json:"productInfo"`

	// In-memory values, omitted in DB serialization and JSON
	TaskGroup  *TaskGroup   `json:"-"`
	ProxyGroup *ProxyGroup  `json:"-"`
	Proxy      *Proxy       `json:"-"`
	Client     *http.Client `json:"-"`
	Scraper    hawk.Scraper `json:"-"`
	StopFlag   bool         `json:"-"`
}

type MonitorInput struct {
	ProxyGroupID     string                 `json:"proxyGroupID"`
	DelayMS          int                    `json:"delayMS"`
	MaxPrice         int                    `json:"maxPrice"`
	MonitorType      enums.MonitorType      `json:"monitorType"`
	Input            string                 `json:"input"`
	SiteSpecificInfo map[string]interface{} `json:"siteSpecificInfo"`
}

type Monitor interface {
	GetProductInfo() (ProductInfo, error)
}

func (monitor *BaseMonitor) PublishEvent(status enums.MonitorStatus, eventType enums.MonitorEventType) {
	monitor.Status = status
	events.GetEventBus().PublishMonitorEvent(status, eventType, nil, monitor.TaskGroup.GroupID)
}

func (monitor *BaseMonitor) CheckForStop() bool {
	if monitor.StopFlag {
		monitor.PublishEvent(enums.MonitorIdle, enums.MonitorStop)
		return true
	}
	return false
}

func (monitor *BaseMonitor) Sleep() {
	time.Sleep(time.Duration(monitor.MonitorInput.DelayMS) * time.Millisecond)
}

func (monitor *BaseMonitor) UpdateProxy(proxy *Proxy) error {
	monitor.Proxy.RemoveCount()
	if proxy == nil {
		err := client.UpdateProxy(monitor.Client, "")
		if err != nil {
			return err
		}
	} else {
		err := client.UpdateProxy(monitor.Client, ProxyCleaner(*proxy))
		if err != nil {
			return err
		}
		monitor.Proxy.AddCount()
	}

	monitor.Proxy = proxy
	return nil
}
