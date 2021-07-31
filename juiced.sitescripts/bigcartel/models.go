package bigcartel

import (
	"backend.juicedbot.io/juiced.infrastructure/common/entities"
	"backend.juicedbot.io/juiced.sitescripts/base"
)

const (
	AddToCart          = "/cart"
	SubmitAddressEmail = "/store/%s/carts/%s"
	SubmitPayment      = "/store/%s/carts/%s"
	Checkout           = "/store/%s/carts/%s"
)

type Monitor struct {
	Monitor         base.Monitor
	RunningMonitors []string
	Pids            []string
	InStock         []BigCartelInStockData
	PidWithInfo     map[string]entities.BoxlunchSingleMonitorInfo
	SiteInfo        SiteInfo
	Skus            []string
}

type Task struct {
	Task      base.Task
	StockData BigCartelInStockData
	SiteInfo  SiteInfo
}

type BigCartelInStockData struct {
}

type SiteInfo struct {
	BaseUrl  string
	Email    string
	Password string
	Sku      string
}
