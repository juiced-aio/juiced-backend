package entities

import (
	"fmt"
	"time"

	"backend.juicedbot.io/juiced.antibot/cloudflare"
	"backend.juicedbot.io/juiced.client/http"
	"backend.juicedbot.io/juiced.infrastructure/enums"
	"backend.juicedbot.io/juiced.infrastructure/events"
	"backend.juicedbot.io/juiced.infrastructure/util"
)

type Task struct {
	ID             string    `json:"ID" db:"ID"`
	TaskGroupID    string    `json:"taskGroupID" db:"taskGroupID"`
	Retailer       string    `json:"retailer" db:"retailer"`
	TaskSerialized string    `json:"-" db:"taskSerialized"`
	Task           *BaseTask `json:"task"`
	CreationDate   int64     `json:"creationDate" db:"creationDate"`
}

type BaseTask struct {
	RetailerTask *RetailerTask `json:"-"`

	// Task inputs, included in DB serialization and JSON
	TaskInput TaskInput `json:"taskInput"`

	// In-memory values, omitted in DB serialization but included in JSON
	Status         enums.TaskStatus `json:"status"`
	Running        bool             `json:"running"`
	ProductInfo    ProductInfo      `json:"productInfo"`
	ActualQuantity int              `json:"actualQuantity"`

	// In-memory values, omitted in DB serialization and JSON
	Task       *Task               `json:"-"`
	TaskGroup  *TaskGroup          `json:"-"`
	Profile    *Profile            `json:"-"`
	ProxyGroup *ProxyGroup         `json:"-"`
	Proxy      *Proxy              `json:"-"`
	Client     *http.Client        `json:"-"`
	Scraper    *cloudflare.Scraper `json:"-"`
	StopFlag   bool                `json:"-"`
}

type TaskInput struct {
	ProxyGroupID     string                 `json:"proxyGroupID"`
	ProfileID        string                 `json:"profileID"`
	Quantity         int                    `json:"quantity"`
	DelayMS          int                    `json:"delayMS"`
	SiteSpecificInfo map[string]interface{} `json:"siteSpecificInfo"`
}

type RetailerTask interface {
	GetSetupFunctions() []TaskFunction
	GetMainFunctions() []TaskFunction
}

type TaskFunction struct {
	Function         func() (bool, string)
	StatusBegin      enums.TaskStatus
	MaxRetries       int
	MsBetweenRetries int

	CheckoutFunction bool

	SpecialFunction bool
	InBackground    bool

	RefreshFunction bool
	RefreshAt       int64
	RefreshEvery    int
}

type ProductInfo struct {
	InStock          bool                   `json:"inStock"`
	InPriceRange     bool                   `json:"inPriceRange"`
	SKU              string                 `json:"sku"`
	Price            float64                `json:"price"`
	ItemName         string                 `json:"itemName"`
	ItemURL          string                 `json:"itemURL"`
	ImageURL         string                 `json:"imageURL"`
	SiteSpecificInfo map[string]interface{} `json:"siteSpecificInfo"`
}

func (task *BaseTask) PublishEvent(status enums.TaskStatus, eventType enums.TaskEventType) {
	task.Status = status
	events.GetEventBus().PublishTaskEvent(status, eventType, nil, task.Task.ID)
}

func (task *BaseTask) CheckForStop() bool {
	if task.StopFlag {
		task.PublishEvent(enums.TaskIdle, enums.TaskStop)
		return true
	}
	return false
}

func (task *BaseTask) WaitForMonitor() bool {
	for {
		needToStop := task.CheckForStop()
		if needToStop {
			return false
		}
		monitors := task.TaskGroup.Monitors
		for _, monitor := range monitors {
			if monitor.ProductInfo.InStock {
				task.ProductInfo = monitor.ProductInfo
				return true
			}
		}

		time.Sleep(time.Millisecond * util.MS_TO_WAIT)
	}
}

func (task *BaseTask) RunFunctions(functions []TaskFunction) bool {
	for _, function := range functions {
		needToStop := task.CheckForStop()
		if needToStop {
			return false
		}

		success := task.RunUntilSuccessful(function)
		if !success {
			return false
		}
	}
	return true
}

// RunUntilSuccessful runs a single function until (a) it succeeds, (b) the task needs to stop, or (c) it fails too many times.
// 		Passing in 0 for maxRetries will retry the function indefinitely.
//		Returns true if the function was successful, false if the function failed (and the task should stop)
func (task *BaseTask) RunUntilSuccessful(function TaskFunction) bool {
	var success bool
	if function.RefreshFunction {
		go task.RefreshWrapper(function)
		success = true
	} else if function.SpecialFunction {
		if function.InBackground {
			go function.Function()
			success = true
		} else {
			var status string
			success, status = function.Function()
			if success && status != "" {
				task.PublishEvent(status, enums.TaskUpdate)
			}
		}
	} else {
		attempt := 1
		if function.MaxRetries == 0 {
			attempt = -1
		}

		for success = task.RunUntilSuccessfulHelper(function.Function, attempt); !success; {
			needToStop := task.CheckForStop()
			if needToStop || attempt > function.MaxRetries {
				task.StopFlag = true
				return false
			}
			if attempt >= 0 {
				attempt++
			}

			if function.MsBetweenRetries == 0 {
				time.Sleep(time.Duration(task.TaskInput.DelayMS) * time.Millisecond)
			} else {
				time.Sleep(time.Duration(function.MsBetweenRetries) * time.Millisecond)
			}
		}
	}

	return success
}

func (task *BaseTask) RunUntilSuccessfulHelper(fn func() (bool, string), attempt int) bool {
	success, status := fn()

	if !success {
		if attempt > 0 {
			if status != "" {
				task.PublishEvent(fmt.Sprint(fmt.Sprintf("(Attempt #%d) ", attempt), status), enums.TaskUpdate)
			}
		} else {
			if status != "" {
				task.PublishEvent(fmt.Sprint("(Retrying) ", status), enums.TaskUpdate)
			}
		}
		return false
	}

	if status != "" {
		task.PublishEvent(status, enums.TaskUpdate)
	}
	return true
}

func (task *BaseTask) RefreshWrapper(function TaskFunction) {
	defer func() {
		if r := recover(); r != nil {
			task.RefreshWrapper(function)
		}
	}()

	for {
		if function.RefreshAt == 0 || time.Now().Unix() > function.RefreshAt {
			if success := task.RunUntilSuccessful(function); !success {
				return
			}
			function.RefreshAt = time.Now().Unix() + 1800
		}
		time.Sleep(util.WAIT_TIME)
	}
}
