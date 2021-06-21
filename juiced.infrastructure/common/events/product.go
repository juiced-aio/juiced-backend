package events

import (
	"backend.juicedbot.io/juiced.infrastructure/common/enums"
)

type TargetStockData struct {
	InStockForShip   []string
	InStockForPickup []string
}

type WalmartStockData struct {
	InStockForShip []WalmartSingleStockData
}

type WalmartSingleStockData struct {
	Sku     string `json:"sku"`
	OfferID string `json:"offerID"`
}

type AmazonSingleStockData struct {
	ASIN        string
	OfferID     string
	Price       int
	ItemName    string
	AntiCsrf    string
	PID         string
	RID         string
	ImageURL    string
	UA          string
	MonitorType enums.MonitorType
}

type AmazonStockData struct {
	InStock []AmazonSingleStockData
}

type BestbuySingleStockData struct {
	SKU   string
	Price int
}

type BestbuyStockData struct {
	InStock []BestbuySingleStockData
}

type HotTopicSingleStockData struct {
	PID string
}

type HottopicStockData struct {
	InStock []HotTopicSingleStockData
}

type GamestopSingleStockData struct {
	SKU        string
	Price      int
	ItemName   string
	PID        string
	ImageURL   string
	ProductURL string
}

type GamestopStockData struct {
	InStock []GamestopSingleStockData
}

// ProductEvent is fired whenever a Monitor finds a product
type ProductEvent struct {
	Retailer     enums.Retailer    `json:"retailer"`
	MonitorID    string            `json:"monitorID"`
	TargetData   TargetStockData   `json:"targetStockData"`
	WalmartData  WalmartStockData  `json:"walmartStockData"`
	AmazonData   AmazonStockData   `json:"amazonStockData"`
	BestbuyData  BestbuyStockData  `json:"bestbuyStockData"`
	GamestopData GamestopStockData `json:"gamestopStockData"`
	HottopicData HottopicStockData `json:"hottopicStockData"`
}
