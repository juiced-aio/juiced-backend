package events

import (
	"backend.juicedbot.io/m/v2/juiced.infrastructure/common/enums"

	"go.mongodb.org/mongo-driver/bson/primitive"
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

type HottopicData struct {
	Pid string
}

type HottopicStockData struct {
	InStock []HotTopicSingleStockData
}

type HotTopicSingleStockData struct {
	PID         string
	MonitorType enums.MonitorType
}

// ProductEvent is fired whenever a Monitor finds a product
type ProductEvent struct {
	DiscordWebhook string             `json:"discordWebhook"`
	Retailer       enums.Retailer     `json:"retailer"`
	MonitorID      primitive.ObjectID `json:"monitorID"`
	TargetData     TargetStockData    `json:"targetStockData"`
	WalmartData    WalmartStockData   `json:"walmartStockData"`
	AmazonData     AmazonStockData    `json:"amazonStockData"`
	BestbuyData    BestbuyStockData   `json:"bestbuyStockData"`
	HottopicData   HottopicStockData  `json:"HottopicStockData"`
}
