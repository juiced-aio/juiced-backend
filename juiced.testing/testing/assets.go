package testing

import "backend.juicedbot.io/juiced.infrastructure/common/entities"

var MainTaskGroupID = "b629f72d-1930-42e6-9314-4c429599414b"
var MainTaskID = "a2c7d260-8603-405b-a228-b120d076f2d3"
var MainProfileGroupID = "aa868707-16d0-4ce9-b601-b5a73c10666b"
var MainProfileID = "ab9df680-31ae-4920-b030-692be4edc467"
var MainAddressID = "33dc12a7-9845-4ac9-8e13-e9226af6f5d3"
var MainCardID = "b427a75f-eb56-4264-b0c4-7209b0626e89"
var MainProxyGroupID = "dadf58d8-e5f7-4b4a-b86e-72e072d30222"

var MainTaskGroup = &entities.TaskGroup{
	GroupID:             MainTaskGroupID,
	Name:                "MainTaskGroup",
	MonitorProxyGroupID: MainProxyGroupID,
	MonitorRetailer:     "",
	MonitorInput:        "",
	MonitorDelay:        0,
	MonitorStatus:       "",
	TaskIDs:             []string{MainTaskID},
	TaskIDsJoined:       "",
	UpdateMonitor:       false,
	CreationDate:        0,
	AmazonMonitorInfo:   &entities.AmazonMonitorInfo{},
	BestbuyMonitorInfo:  &entities.BestbuyMonitorInfo{},
	BoxlunchMonitorInfo: &entities.BoxlunchMonitorInfo{},
	DisneyMonitorInfo:   &entities.DisneyMonitorInfo{},
	GamestopMonitorInfo: &entities.GamestopMonitorInfo{},
	HottopicMonitorInfo: &entities.HottopicMonitorInfo{},
	ShopifyMonitorInfo:  &entities.ShopifyMonitorInfo{},
	TargetMonitorInfo:   &entities.TargetMonitorInfo{},
	WalmartMonitorInfo:  &entities.WalmartMonitorInfo{},
}

var MainTask = &entities.Task{
	ID:               MainTaskID,
	TaskGroupID:      MainTaskGroupID,
	TaskProfileID:    "",
	TaskProxyGroupID: MainProxyGroupID,
	TaskRetailer:     "",
	TaskSize:         []string{},
	TaskSizeJoined:   "",
	TaskQty:          0,
	TaskStatus:       "",
	TaskDelay:        0,
	CreationDate:     0,
	AmazonTaskInfo:   &entities.AmazonTaskInfo{},
	BestbuyTaskInfo:  &entities.BestbuyTaskInfo{},
	BoxlunchTaskInfo: &entities.BoxlunchTaskInfo{},
	DisneyTaskInfo:   &entities.DisneyTaskInfo{},
	GamestopTaskInfo: &entities.GamestopTaskInfo{},
	HottopicTaskInfo: &entities.HottopicTaskInfo{},
	ShopifyTaskInfo:  &entities.ShopifyTaskInfo{},
	TargetTaskInfo:   &entities.TargetTaskInfo{},
	WalmartTaskInfo:  &entities.WalmartTaskInfo{},
}

var MainProfileGroup = &entities.ProfileGroup{
	GroupID:          MainProfileGroupID,
	Name:             "MainProfileGroup",
	ProfileIDs:       []string{MainProfileID},
	ProfileIDsJoined: "",
	CreationDate:     0,
}

var MainProfile = &entities.Profile{
	ID:                    MainProfileID,
	ProfileGroupIDs:       []string{},
	ProfileGroupIDsJoined: "",
	Name:                  "MainProfile",
	Email:                 "test@gmail.com",
	PhoneNumber:           "8059991001",
	ShippingAddress:       MainAddress,
	BillingAddress:        MainAddress,
	CreditCard:            MainCard,
	CreationDate:          0,
}

var MainAddress = entities.Address{
	ID:          MainAddressID,
	ProfileID:   MainProfileID,
	FirstName:   "Juiced",
	LastName:    "AIO",
	Address1:    "3500 Data Dr",
	Address2:    "",
	City:        "Rancho Cordova",
	ZipCode:     "95670",
	StateCode:   "CA",
	CountryCode: "US",
}

var MainCard = entities.Card{
	ID:             MainCardID,
	ProfileID:      MainProfileID,
	CardholderName: "Juiced AIO",
	CardNumber:     "4767718212263745",
	ExpMonth:       "02",
	ExpYear:        "26",
	CVV:            "260",
	CardType:       "Visa",
}

var MainProxyGroup = &entities.ProxyGroup{
	GroupID:      MainProxyGroupID,
	Name:         "MainProxyGroup",
	Proxies:      []entities.Proxy{{Host: "localhost", Port: "8888"}},
	CreationDate: 0,
}
