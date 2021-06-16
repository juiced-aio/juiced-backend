package entities

import (
	"encoding/json"

	"backend.juicedbot.io/juiced.client/http"
	"backend.juicedbot.io/juiced.infrastructure/common/enums"
)

// Task is a class that holds details about a single bot task
type Task struct {
	ID               string           `json:"ID" db:"ID"`
	TaskGroupID      string           `json:"taskGroupID" db:"taskGroupID"`
	TaskProfileID    string           `json:"profileID" db:"profileID"`
	TaskProxyGroupID string           `json:"proxyGroupID" db:"proxyGroupID"`
	TaskRetailer     enums.Retailer   `json:"retailer" db:"retailer"`
	TaskSize         []string         `json:"size" db:"size"`
	TaskSizeJoined   string           `json:"sizeJoined" db:"sizeJoined"`
	TaskQty          int              `json:"qty" db:"qty"`
	TaskStatus       enums.TaskStatus `json:"status" db:"status"`
	TaskDelay        int              `json:"taskDelay" db:"taskDelay"`
	CreationDate     int64            `json:"creationDate" db:"creationDate"`
	TargetTaskInfo   TargetTaskInfo   `json:"targetTaskInfo"`
	WalmartTaskInfo  WalmartTaskInfo  `json:"walmartTaskInfo"`
	AmazonTaskInfo   AmazonTaskInfo   `json:"amazonTaskInfo"`
	BestbuyTaskInfo  BestbuyTaskInfo  `json:"bestbuyTaskInfo"`
	GamestopTaskInfo GamestopTaskInfo `json:"gamestopTaskInfo"`
	HottopicTaskInfo HottopicTaskInfo `json:"hottopicTaskInfo"`
	// Future sitescripts will have a field here
}

type HottopicTaskInfo struct {
	TaskID      string   `json:"taskID" db:"taskID"`
	TaskGroupID string   `json:"taskGroupID" db:"taskGroupID"`
	Pids        []string `json:"pids" db:"pids"`
}

type TargetTaskInfo struct {
	TaskID       string             `json:"taskID" db:"taskID"`
	TaskGroupID  string             `json:"taskGroupID" db:"taskGroupID"`
	CheckoutType enums.CheckoutType `json:"checkoutType" db:"checkoutType"`
	Email        string             `json:"email" db:"email"`
	Password     string             `json:"password" db:"password"`
	PaymentType  enums.PaymentType  `json:"paymentType" db:"paymentType"`
}

type WalmartTaskInfo struct {
	TaskID      string `json:"taskID" db:"taskID"`
	TaskGroupID string `json:"taskGroupID" db:"taskGroupID"`
}

type AmazonTaskInfo struct {
	TaskID      string          `json:"taskID" db:"taskID"`
	TaskGroupID string          `json:"taskGroupID" db:"taskGroupID"`
	Email       string          `json:"email" db:"email"`
	Password    string          `json:"password" db:"password"`
	LoginType   enums.LoginType `json:"loginType" db:"loginType"`
}

type BestbuyTaskInfo struct {
	TaskID      string         `json:"taskID" db:"taskID"`
	TaskGroupID string         `json:"taskGroupID" db:"taskGroupID"`
	Email       string         `json:"email" db:"email"`
	Password    string         `json:"password" db:"password"`
	TaskType    enums.TaskType `json:"taskType" db:"taskType"`
}

type GamestopTaskInfo struct {
	TaskID      string         `json:"taskID" db:"taskID"`
	TaskGroupID string         `json:"taskGroupID" db:"taskGroupID"`
	Email       string         `json:"email" db:"email"`
	Password    string         `json:"password" db:"password"`
	TaskType    enums.TaskType `json:"taskType" db:"taskType"`
}

// SetID updates the Task's ID
func (task *Task) SetID(ID string) {
	task.ID = ID
}

// SetTaskProfileID updates the Task's TaskProfileID
func (task *Task) SetTaskProfileID(TaskProfileID string) {
	task.TaskProfileID = TaskProfileID
}

// SetTaskStatus updates the Tasks's TaskStatus
func (task *Task) SetTaskStatus(TaskStatus enums.TaskStatus) {
	task.TaskStatus = TaskStatus
}

// ParseTask returns a Task object parsed from a JSON bytes array
func ParseTask(task *Task, data []byte) error {
	err := json.Unmarshal(data, &task)
	return err
}

// TaskGroupWithTasks is a class that holds a list of Tasks and a Monitor
type TaskGroupWithTasks struct {
	GroupID             string              `json:"groupID" db:"groupID"`
	Name                string              `json:"name" db:"name"`
	MonitorProxyGroupID string              `json:"proxyGroupID" db:"proxyGroupID"`
	MonitorRetailer     enums.Retailer      `json:"retailer" db:"retailer"`
	MonitorDelay        int                 `json:"delay" db:"delay"`
	MonitorStatus       enums.MonitorStatus `json:"status" db:"status"`
	Tasks               []Task              `json:"tasks"`
	TargetMonitorInfo   TargetMonitorInfo   `json:"targetMonitorInfo"`
	WalmartMonitorInfo  WalmartMonitorInfo  `json:"walmartMonitorInfo"`
	AmazonMonitorInfo   AmazonMonitorInfo   `json:"amazonMonitorInfo"`
	BestbuyMonitorInfo  BestbuyMonitorInfo  `json:"bestbuyMonitorInfo"`
	GamestopMonitorInfo GamestopMonitorInfo `json:"gamestopMonitorInfo"`
	HottopicMonitorInfo HottopicMonitorInfo `json:"hottopicMonitorInfo"`
	// Future sitescripts will have a field here
}

// SetTasks updates the TaskGroupWithTasks's TaskIDs
func (taskGroup *TaskGroupWithTasks) SetTasks(tasks []Task) {
	taskGroup.Tasks = tasks
}

// TaskGroup is a class that holds a list of TaskIDs and a Monitor
type TaskGroup struct {
	GroupID             string              `json:"groupID" db:"groupID"`
	Name                string              `json:"name" db:"name"`
	MonitorProxyGroupID string              `json:"proxyGroupID" db:"proxyGroupID"`
	MonitorRetailer     enums.Retailer      `json:"retailer" db:"retailer"`
	MonitorInput        string              `json:"input" db:"input"`
	MonitorDelay        int                 `json:"delay" db:"delay"`
	MonitorStatus       enums.MonitorStatus `json:"status" db:"status"`
	TaskIDs             []string            `json:"taskIDs" db:"taskIDs"`
	TaskIDsJoined       string              `json:"taskIDsJoined" db:"taskIDsJoined"`
	CreationDate        int64               `json:"creationDate" db:"creationDate"`
	TargetMonitorInfo   TargetMonitorInfo   `json:"targetMonitorInfo"`
	WalmartMonitorInfo  WalmartMonitorInfo  `json:"walmartMonitorInfo"`
	AmazonMonitorInfo   AmazonMonitorInfo   `json:"amazonMonitorInfo"`
	BestbuyMonitorInfo  BestbuyMonitorInfo  `json:"bestbuyMonitorInfo"`
	GamestopMonitorInfo GamestopMonitorInfo `json:"gamestopMonitorInfo"`
	HottopicMonitorInfo HottopicMonitorInfo `json:"hottopicMonitorInfo"`

	// Future sitescripts will have a field here
}

type TargetSingleMonitorInfo struct {
	MonitorID    string             `json:"monitorID" db:"monitorID"`
	TaskGroupID  string             `json:"taskGroupID" db:"taskGroupID"`
	TCIN         string             `json:"tcin" db:"tcin"`
	MaxPrice     int                `json:"maxPrice" db:"maxPrice"`
	CheckoutType enums.CheckoutType `json:"checkoutType" db:"checkoutType"`
}

// TargetMonitorInfo is a class that holds Target-specific details for a single monitor
type TargetMonitorInfo struct {
	ID          string                    `json:"ID" db:"ID"`
	TaskGroupID string                    `json:"taskGroupID" db:"taskGroupID"`
	Monitors    []TargetSingleMonitorInfo `json:"monitors"`
	StoreID     string                    `json:"storeID" db:"storeID"`
}

type WalmartMonitorInfo struct {
	ID          string            `json:"ID" db:"ID"`
	TaskGroupID string            `json:"taskGroupID" db:"taskGroupID"`
	MonitorType enums.MonitorType `json:"monitorType" db:"monitorType"`
	SKUs        []string          `json:"skus"`
	SKUsJoined  string            `json:"skusJoined" db:"skusJoined"`
	MaxPrice    int               `json:"maxPrice" db:"maxPrice"`
}

type AmazonSingleMonitorInfo struct {
	MonitorID   string            `json:"monitorID" db:"monitorID"`
	TaskGroupID string            `json:"taskGroupID" db:"taskGroupID"`
	MonitorType enums.MonitorType `json:"monitorType" db:"monitorType"`
	ASIN        string            `json:"asin" db:"asin"`
	OFID        string            `json:"ofid" db:"ofid"`
	MaxPrice    int               `json:"maxPrice" db:"maxPrice"`
	Client      http.Client
}

type AmazonMonitorInfo struct {
	ID          string                    `json:"ID" db:"ID"`
	TaskGroupID string                    `json:"taskGroupID" db:"taskGroupID"`
	Monitors    []AmazonSingleMonitorInfo `json:"monitors"`
}

type BestbuySingleMonitorInfo struct {
	MonitorID   string `json:"monitorID" db:"monitorID"`
	TaskGroupID string `json:"taskGroupID" db:"taskGroupID"`
	SKU         string `json:"sku" db:"sku"`
	MaxPrice    int    `json:"maxPrice" db:"maxPrice"`
}

type BestbuyMonitorInfo struct {
	ID          string                     `json:"ID" db:"ID"`
	TaskGroupID string                     `json:"taskGroupID" db:"taskGroupID"`
	Monitors    []BestbuySingleMonitorInfo `json:"monitors"`
}

type HottopicSingleMonitorInfo struct {
	MonitorID   string            `json:"monitorID" db:"monitorID"`
	TaskGroupID string            `json:"taskGroupID" db:"taskGroupID"`
	Pid         string            `json:"pids" db:"pids"`
	Size        string            `json:"size" db:"size"`
	Color       string            `json:"color" db:"color"`
	MaxPrice    int               `json:"maxPrice" db:"maxPrice"`
	MonitorType enums.MonitorType `json:"monitorType" db:"monitorType"`
}
type HottopicMonitorInfo struct {
	ID          string                      `json:"ID" db:"ID"`
	TaskGroupID string                      `json:"taskGroupID" db:"taskGroupID"`
	Monitors    []HottopicSingleMonitorInfo `json:"monitors"`
}

type GamestopSingleMonitorInfo struct {
	MonitorID   string `json:"monitorID" db:"monitorID"`
	TaskGroupID string `json:"taskGroupID" db:"taskGroupID"`
	SKU         string `json:"sku" db:"sku"`
	MaxPrice    int    `json:"maxPrice" db:"maxPrice"`
}

type GamestopMonitorInfo struct {
	ID          string                      `json:"ID" db:"ID"`
	TaskGroupID string                      `json:"taskGroupID" db:"taskGroupID"`
	Monitors    []GamestopSingleMonitorInfo `json:"monitors"`
}

// AddTasksToGroup adds the given Tasks to the TaskGroup
func (taskGroup *TaskGroup) AddTasksToGroup(tasksToAdd []string) {
	tasks := taskGroup.TaskIDs
	for i := 0; i < len(tasksToAdd); i++ {
		tasks = append(tasks, tasksToAdd[i])
	}
	taskGroup.TaskIDs = tasks
}

// SetTaskIDs updates the TaskGroup's TaskIDs
func (taskGroup *TaskGroup) SetTaskIDs(taskIDs []string) {
	taskGroup.TaskIDs = taskIDs
}

// SetName updates the TaskGroup's Name
func (taskGroup *TaskGroup) SetName(name string) {
	taskGroup.Name = name
}

// SetGroupID updates the TaskGroup's GroupID
func (taskGroup *TaskGroup) SetGroupID(GroupID string) {
	taskGroup.GroupID = GroupID
}

// SetMonitorRetailer updates the TaskGroup's MonitorRetailer
func (taskGroup *TaskGroup) SetMonitorRetailer(MonitorRetailer enums.Retailer) {
	taskGroup.MonitorRetailer = MonitorRetailer
}

// SetMonitorStatus updates the TaskGroup's MonitorStatus
func (taskGroup *TaskGroup) SetMonitorStatus(MonitorStatus enums.MonitorStatus) {
	taskGroup.MonitorStatus = MonitorStatus
}

// ParseTaskGroup returns a TaskGroup object parsed from a JSON bytes array
func ParseTaskGroup(taskGroup *TaskGroup, data []byte) error {
	err := json.Unmarshal(data, &taskGroup)
	return err
}
