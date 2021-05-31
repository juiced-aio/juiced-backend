package entities

import (
	"encoding/json"
	"net/http"

	"go.mongodb.org/mongo-driver/bson/primitive"

	"backend.juicedbot.io/m/v2/juiced.infrastructure/common/enums"
)

// Task is a class that holds details about a single bot task
type Task struct {
	ID               primitive.ObjectID `json:"ID"`
	TaskGroupID      primitive.ObjectID `json:"taskGroupID"`
	TaskProfileID    primitive.ObjectID `json:"profileID"`
	TaskProxyGroupID primitive.ObjectID `json:"proxyGroupID"`
	TaskRetailer     enums.Retailer     `json:"retailer"`
	TaskSize         []string           `json:"size"`
	TaskQty          int                `json:"qty"`
	TaskStatus       enums.TaskStatus   `json:"status"`
	TaskDelay        int                `json:"taskDelay"`
	TargetTaskInfo   TargetTaskInfo     `json:"targetTaskInfo"`
	WalmartTaskInfo  WalmartTaskInfo    `json:"walmartTaskInfo"`
	AmazonTaskInfo   AmazonTaskInfo     `json:"amazonTaskInfo"`
	BestbuyTaskInfo  BestbuyTaskInfo    `json:"bestbuyTaskInfo"`
	// Future sitescripts will have a field here
}

type WalmartTaskInfo struct {
	MaxPrice int `json:"maxPrice"`
}

// TargetTaskInfo is a class that holds Target-specific details for a single bot task
type TargetTaskInfo struct {
	CheckoutType enums.CheckoutType `json:"checkoutType"`
	Email        string             `json:"email"`
	Password     string             `json:"password"`
	PaymentType  enums.PaymentType  `json:"paymentType"`
}
type AmazonTaskInfo struct {
	Email     string          `json:"email"`
	Password  string          `json:"password"`
	LoginType enums.LoginType `json:"loginType"`
}

type BestbuyTaskInfo struct {
	Email    string         `json:"email"`
	Password string         `json:"password"`
	TaskType enums.TaskType `json:"taskType"`
}

// SetID updates the Task's ID
func (task *Task) SetID(ID primitive.ObjectID) {
	task.ID = ID
}

// SetTaskProfileID updates the Task's TaskProfileID
func (task *Task) SetTaskProfileID(TaskProfileID primitive.ObjectID) {
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
	GroupID             primitive.ObjectID  `json:"groupID" bson:"groupID"`
	Name                string              `json:"name" bson:"name"`
	MonitorProxyGroupID primitive.ObjectID  `json:"proxyGroupID" bson:"proxyGroupID"`
	MonitorRetailer     enums.Retailer      `json:"retailer" bson:"retailer"`
	MonitorDelay        int                 `json:"delay" bson:"delay"`
	MonitorStatus       enums.MonitorStatus `json:"status" bson:"status"`
	Tasks               []Task              `json:"tasks" bson:"tasks"`
	TargetMonitorInfo   TargetMonitorInfo   `json:"targetMonitorInfo" bson:"targetMonitorInfo"`
	WalmartMonitorInfo  WalmartMonitorInfo  `json:"walmartMonitorInfo" bson:"walmartMonitorInfo"`
	AmazonMonitorInfo   AmazonMonitorInfo   `json:"amazonMonitorInfo" bson:"amazonMonitorInfo"`
	BestbuyMonitorInfo  BestbuyMonitorInfo  `json:"bestbuyMonitorInfo" bson:"bestbuyMonitorInfo"`
	// Future sitescripts will have a field here
}

// SetTasks updates the TaskGroupWithTasks's TaskIDs
func (taskGroup *TaskGroupWithTasks) SetTasks(tasks []Task) {
	taskGroup.Tasks = tasks
}

// TaskGroup is a class that holds a list of TaskIDs and a Monitor
type TaskGroup struct {
	GroupID             primitive.ObjectID   `json:"groupID" bson:"groupID"`
	Name                string               `json:"name" bson:"name"`
	MonitorProxyGroupID primitive.ObjectID   `json:"proxyGroupID" bson:"proxyGroupID"`
	MonitorRetailer     enums.Retailer       `json:"retailer" bson:"retailer"`
	MonitorInput        string               `json:"input" bson:"input"`
	MonitorDelay        int                  `json:"delay" bson:"delay"`
	MonitorStatus       enums.MonitorStatus  `json:"status" bson:"status"`
	TargetMonitorInfo   TargetMonitorInfo    `json:"targetMonitorInfo" bson:"targetMonitorInfo"`
	WalmartMonitorInfo  WalmartMonitorInfo   `json:"walmartMonitorInfo" bson:"walmartMonitorInfo"`
	AmazonMonitorInfo   AmazonMonitorInfo    `json:"amazonMonitorInfo" bson:"amazonMonitorInfo"`
	BestbuyMonitorInfo  BestbuyMonitorInfo   `json:"bestbuyMonitorInfo" bson:"bestbuyMonitorInfo"`
	TaskIDs             []primitive.ObjectID `json:"taskIDs" bson:"taskIDs"`
	// Future sitescripts will have a field here
}

// TargetMonitorInfo is a class that holds Target-specific details for a single monitor
type TargetMonitorInfo struct {
	MonitorType enums.MonitorType `json:"monitorType"`
	TCINs       []string          `json:"tcins"`
	StoreID     string            `json:"storeID"`
}

type WalmartMonitorInfo struct {
	MonitorType enums.MonitorType `json:"monitorType"`
	SKUs        []string          `json:"skus"`
}

type AmazonSingleMonitorInfo struct {
	ASIN        string            `json:"asin"`
	OFID        string            `json:"ofid"`
	MaxPrice    int               `json:"maxPrice"`
	MonitorType enums.MonitorType `json:"monitorType"`
	Client      http.Client
}

type AmazonMonitorInfo struct {
	Monitors []AmazonSingleMonitorInfo `json:"monitors"`
}

type BestbuySingleMonitorInfo struct {
	SKU      string `json:"skus"`
	MaxPrice int    `json:"maxPrice"`
}

type BestbuyMonitorInfo struct {
	Monitors []BestbuySingleMonitorInfo `json:"monitors"`
}

// AddTasksToGroup adds the given Tasks to the TaskGroup
func (taskGroup *TaskGroup) AddTasksToGroup(tasksToAdd []primitive.ObjectID) {
	tasks := taskGroup.TaskIDs
	for i := 0; i < len(tasksToAdd); i++ {
		tasks = append(tasks, tasksToAdd[i])
	}
	taskGroup.TaskIDs = tasks
}

// SetTaskIDs updates the TaskGroup's TaskIDs
func (taskGroup *TaskGroup) SetTaskIDs(taskIDs []primitive.ObjectID) {
	taskGroup.TaskIDs = taskIDs
}

// SetName updates the TaskGroup's Name
func (taskGroup *TaskGroup) SetName(name string) {
	taskGroup.Name = name
}

// SetGroupID updates the TaskGroup's GroupID
func (taskGroup *TaskGroup) SetGroupID(GroupID primitive.ObjectID) {
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
