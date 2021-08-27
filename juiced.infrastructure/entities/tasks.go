package entities

import (
	"net/http"
	"time"

	"backend.juicedbot.io/juiced.infrastructure/common/enums"
	"backend.juicedbot.io/juiced.infrastructure/common/events"
	"backend.juicedbot.io/juiced.sitescripts/hawk-go"
)

type Task struct {
	ID           string `json:"ID" db:"ID"`
	TaskGroupID  string `json:"taskGroupID" db:"taskGroupID"`
	TaskGroup    *TaskGroup
	ProfileID    string `json:"profileID" db:"profileID"`
	Profile      *Profile
	ProxyGroupID string `json:"proxyGroupID" db:"proxyGroupID"`
	ProxyGroup   *ProxyGroup
	Task         BaseTask `json:"taskInfo" db:"taskInfo"`
	CreationDate int64    `json:"creationDate" db:"creationDate"`
}

type CreateTaskRequest struct {
	TaskGroupID  string `json:"taskGroupID" db:"taskGroupID"`
	ProfileID    string `json:"profileID" db:"profileID"`
	ProxyGroupID string `json:"proxyGroupID" db:"proxyGroupID"`
}

func CreateTask() (*Task, error) {
	// TODO
	task := &Task{}

	return task, nil
}

type BaseTask struct {
	RetailerTask
	Retailer string           `json:"retailer"`
	Quantity int              `json:"quantity"`
	Status   enums.TaskStatus `json:"status"`
	DelayMS  int              `json:"delayMS"`
}

type RetailerTask interface {
	GetTaskInfo() *TaskInfo
	FillStockInfo(StockInfo)
	GetTaskFunctions() []TaskFunction
}

type TaskFunction struct {
	Function          func() (bool, string)
	StatusBegin       enums.TaskStatus
	InBackground      bool
	SpecialFunction   bool
	RefreshFunction   bool
	RefreshAt         int64
	RefreshEvery      int
	MaxRetries        int
	MsBetweenRetries  int
	WaitingForMonitor bool
	Checkout          bool
}

type SiteSpecificInfo map[string]interface{}

type StockInfo struct {
	SKU          string
	Price        float64
	ItemName     string
	ImageURL     string
	SiteSpecific SiteSpecificInfo
}

type TaskInfo struct {
	Task       *Task
	Profile    *Profile
	Proxy      *Proxy
	ProxyGroup *ProxyGroup
	EventBus   *events.EventBus
	StopFlag   bool
	StartTime  time.Time
	EndTime    time.Time
	Client     http.Client
	Scraper    hawk.Scraper
	ErrorField string

	StockInfo StockInfo
}
