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
	TaskSize         []string         `json:"size"`
	TaskSizeJoined   string           `json:"sizeJoined" db:"sizeJoined"`
	TaskQty          int              `json:"qty" db:"qty"`
	TaskStatus       enums.TaskStatus `json:"status" db:"status"`
	TaskDelay        int              `json:"taskDelay" db:"taskDelay"`
	UpdateTask       bool
	CreationDate     int64             `json:"creationDate" db:"creationDate"`
	AmazonTaskInfo   *AmazonTaskInfo   `json:"amazonTaskInfo,omitempty"`
	BestbuyTaskInfo  *BestbuyTaskInfo  `json:"bestbuyTaskInfo,omitempty"`
	BoxlunchTaskInfo *BoxlunchTaskInfo `json:"boxlunchTaskInfo,omitempty"`
	DisneyTaskInfo   *DisneyTaskInfo   `json:"disneyTaskInfo,omitempty"`
	GamestopTaskInfo *GamestopTaskInfo `json:"gamestopTaskInfo,omitempty"`
	HottopicTaskInfo *HottopicTaskInfo `json:"hottopicTaskInfo,omitempty"`
	NeweggTaskInfo   *NeweggTaskInfo   `json:"neweggTaskInfo,omitempty"`
	ShopifyTaskInfo  *ShopifyTaskInfo  `json:"shopifyTaskInfo,omitempty"`
	TargetTaskInfo   *TargetTaskInfo   `json:"targetTaskInfo,omitempty"`
	ToppsTaskInfo    *ToppsTaskInfo    `json:"toppsTaskInfo,omitempty"`
	WalmartTaskInfo  *WalmartTaskInfo  `json:"walmartTaskInfo,omitempty"`
	// Future sitescripts will have a field here
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
	LocationID  string         `json:"locationID" db:"locationID"`
	TaskType    enums.TaskType `json:"taskType" db:"taskType"`
}

type BoxlunchTaskInfo struct {
	TaskID      string `json:"taskID" db:"taskID"`
	TaskGroupID string `json:"taskGroupID" db:"taskGroupID"`
}

type DisneyTaskInfo struct {
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

type HottopicTaskInfo struct {
	TaskID      string `json:"taskID" db:"taskID"`
	TaskGroupID string `json:"taskGroupID" db:"taskGroupID"`
}

type HotWheelsTaskInfo struct {
	TaskID      string `json:"taskID" db:"taskID"`
	TaskGroupID string `json:"taskGroupID" db:"taskGroupID"`
	Email       string `json:"email" db:"email"`
	Password    string `json:"password" db:"password"`
}

type NeweggTaskInfo struct {
	TaskID      string `json:"taskID" db:"taskID"`
	TaskGroupID string `json:"taskGroupID" db:"taskGroupID"`
}

type ShopifyTaskInfo struct {
	TaskID            string                `json:"taskID" db:"taskID"`
	TaskGroupID       string                `json:"taskGroupID" db:"taskGroupID"`
	CouponCode        string                `json:"couponCode" db:"couponCode"`
	SiteURL           string                `json:"siteURL" db:"siteURL"`
	SitePassword      string                `json:"sitePassword" db:"sitePassword"`
	ShopifyRetailer   enums.ShopifyRetailer `json:"shopifyRetailer" db:"shopifyRetailer"`
	HotWheelsTaskInfo *HotWheelsTaskInfo    `json:"hotWheelsTaskInfo,omitempty"`
}

type TargetTaskInfo struct {
	TaskID       string             `json:"taskID" db:"taskID"`
	TaskGroupID  string             `json:"taskGroupID" db:"taskGroupID"`
	CheckoutType enums.CheckoutType `json:"checkoutType" db:"checkoutType"`
	Email        string             `json:"email" db:"email"`
	Password     string             `json:"password" db:"password"`
	PaymentType  enums.PaymentType  `json:"paymentType" db:"paymentType"`
}

type ToppsTaskInfo struct {
	TaskID      string         `json:"taskID" db:"taskID"`
	TaskGroupID string         `json:"taskGroupID" db:"taskGroupID"`
	Email       string         `json:"email" db:"email"`
	Password    string         `json:"password" db:"password"`
	TaskType    enums.TaskType `json:"taskType" db:"taskType"`
}

type WalmartTaskInfo struct {
	TaskID      string `json:"taskID" db:"taskID"`
	TaskGroupID string `json:"taskGroupID" db:"taskGroupID"`
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
	GroupID             string               `json:"groupID" db:"groupID"`
	Name                string               `json:"name" db:"name"`
	MonitorProxyGroupID string               `json:"proxyGroupID" db:"proxyGroupID"`
	MonitorRetailer     enums.Retailer       `json:"retailer" db:"retailer"`
	MonitorDelay        int                  `json:"delay" db:"delay"`
	MonitorStatus       enums.MonitorStatus  `json:"status" db:"status"`
	Tasks               []Task               `json:"tasks"`
	AmazonMonitorInfo   *AmazonMonitorInfo   `json:"amazonMonitorInfo,omitempty"`
	BestbuyMonitorInfo  *BestbuyMonitorInfo  `json:"bestbuyMonitorInfo,omitempty"`
	BoxlunchMonitorInfo *BoxlunchMonitorInfo `json:"boxlunchMonitorInfo,omitempty"`
	DisneyMonitorInfo   *DisneyMonitorInfo   `json:"disneyMonitorInfo,omitempty"`
	GamestopMonitorInfo *GamestopMonitorInfo `json:"gamestopMonitorInfo,omitempty"`
	HottopicMonitorInfo *HottopicMonitorInfo `json:"hottopicMonitorInfo,omitempty"`
	NeweggMonitorInfo   *NeweggMonitorInfo   `json:"neweggMonitorInfo,omitempty"`
	ShopifyMonitorInfo  *ShopifyMonitorInfo  `json:"shopifyMonitorInfo,omitempty"`
	TargetMonitorInfo   *TargetMonitorInfo   `json:"targetMonitorInfo,omitempty"`
	ToppsMonitorInfo    *ToppsMonitorInfo    `json:"toppsMonitorInfo,omitempty"`
	WalmartMonitorInfo  *WalmartMonitorInfo  `json:"walmartMonitorInfo,omitempty"`

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
	UpdateMonitor       bool
	CreationDate        int64                `json:"creationDate" db:"creationDate"`
	AmazonMonitorInfo   *AmazonMonitorInfo   `json:"amazonMonitorInfo,omitempty"`
	BestbuyMonitorInfo  *BestbuyMonitorInfo  `json:"bestbuyMonitorInfo,omitempty"`
	BoxlunchMonitorInfo *BoxlunchMonitorInfo `json:"boxlunchMonitorInfo,omitempty"`
	DisneyMonitorInfo   *DisneyMonitorInfo   `json:"disneyMonitorInfo,omitempty"`
	GamestopMonitorInfo *GamestopMonitorInfo `json:"gamestopMonitorInfo,omitempty"`
	HottopicMonitorInfo *HottopicMonitorInfo `json:"hottopicMonitorInfo,omitempty"`
	NeweggMonitorInfo   *NeweggMonitorInfo   `json:"neweggMonitorInfo,omitempty"`
	ShopifyMonitorInfo  *ShopifyMonitorInfo  `json:"shopifyMonitorInfo,omitempty"`
	TargetMonitorInfo   *TargetMonitorInfo   `json:"targetMonitorInfo,omitempty"`
	ToppsMonitorInfo    *ToppsMonitorInfo    `json:"toppsMonitorInfo,omitempty"`
	WalmartMonitorInfo  *WalmartMonitorInfo  `json:"walmartMonitorInfo,omitempty"`

	// Future sitescripts will have a field here
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

type BoxlunchSingleMonitorInfo struct {
	MonitorID   string            `json:"monitorID" db:"monitorID"`
	TaskGroupID string            `json:"taskGroupID" db:"taskGroupID"`
	Pid         string            `json:"pid" db:"pid"`
	Size        string            `json:"size" db:"size"`
	Color       string            `json:"color" db:"color"`
	MaxPrice    int               `json:"maxPrice" db:"maxPrice"`
	MonitorType enums.MonitorType `json:"monitorType" db:"monitorType"`
}

type BoxlunchMonitorInfo struct {
	ID          string                      `json:"ID" db:"ID"`
	TaskGroupID string                      `json:"taskGroupID" db:"taskGroupID"`
	Monitors    []BoxlunchSingleMonitorInfo `json:"monitors"`
}

type DisneySingleMonitorInfo struct {
	MonitorID   string `json:"monitorID" db:"monitorID"`
	TaskGroupID string `json:"taskGroupID" db:"taskGroupID"`
	PID         string `json:"pid" db:"pid"`
	Size        string `json:"size" db:"size"`
	Color       string `json:"color" db:"color"`
	MaxPrice    int    `json:"maxPrice" db:"maxPrice"`
}
type DisneyMonitorInfo struct {
	ID          string                    `json:"ID" db:"ID"`
	TaskGroupID string                    `json:"taskGroupID" db:"taskGroupID"`
	Monitors    []DisneySingleMonitorInfo `json:"monitors"`
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

type HottopicSingleMonitorInfo struct {
	MonitorID   string            `json:"monitorID" db:"monitorID"`
	TaskGroupID string            `json:"taskGroupID" db:"taskGroupID"`
	Pid         string            `json:"pid" db:"pid"`
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

type NeweggSingleMonitorInfo struct {
	MonitorID   string `json:"monitorID" db:"monitorID"`
	TaskGroupID string `json:"taskGroupID" db:"taskGroupID"`
	SKU         string `json:"sku" db:"sku"`
	MaxPrice    int    `json:"maxPrice" db:"maxPrice"`
}

type NeweggMonitorInfo struct {
	ID          string                    `json:"ID" db:"ID"`
	TaskGroupID string                    `json:"taskGroupID" db:"taskGroupID"`
	Monitors    []NeweggSingleMonitorInfo `json:"monitors"`
}

type ShopifySingleMonitorInfo struct {
	MonitorID   string `json:"monitorID" db:"monitorID"`
	TaskGroupID string `json:"taskGroupID" db:"taskGroupID"`
	VariantID   string `json:"variantID" db:"variantID"`
	MaxPrice    int    `json:"maxPrice" db:"maxPrice"`
}

type ShopifyMonitorInfo struct {
	ID           string                     `json:"ID" db:"ID"`
	TaskGroupID  string                     `json:"taskGroupID" db:"taskGroupID"`
	SiteURL      string                     `json:"siteURL" db:"siteURL"`
	SitePassword string                     `json:"sitePassword" db:"sitePassword"`
	Monitors     []ShopifySingleMonitorInfo `json:"monitors"`
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
	MonitorType enums.MonitorType         `json:"monitorType" db:"monitorType"`
}

type ToppsSingleMonitorInfo struct {
	MonitorID   string `json:"monitorID" db:"monitorID"`
	TaskGroupID string `json:"taskGroupID" db:"taskGroupID"`
	Item        string `json:"item" db:"item"`
	MaxPrice    int    `json:"maxPrice" db:"maxPrice"`
}

type ToppsMonitorInfo struct {
	ID          string                   `json:"ID" db:"ID"`
	TaskGroupID string                   `json:"taskGroupID" db:"taskGroupID"`
	Monitors    []ToppsSingleMonitorInfo `json:"monitors"`
}

type WalmartMonitorInfo struct {
	ID          string                     `json:"ID" db:"ID"`
	TaskGroupID string                     `json:"taskGroupID" db:"taskGroupID"`
	Monitors    []WalmartSingleMonitorInfo `json:"monitors"`
}

type WalmartSingleMonitorInfo struct {
	MonitorID     string            `json:"monitorID" db:"monitorID"`
	TaskGroupID   string            `json:"taskGroupID" db:"taskGroupID"`
	ID            string            `json:"id" db:"id"`
	MaxPrice      int               `json:"maxPrice" db:"maxPrice"`
	SoldByWalmart bool              `json:"soldByWalmart" db:"soldByWalmart"`
	MonitorType   enums.MonitorType `json:"monitorType" db:"monitorType"`
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
