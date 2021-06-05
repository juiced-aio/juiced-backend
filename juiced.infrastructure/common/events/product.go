package events

import (
	"backend.juicedbot.io/juiced.infrastructure/common/entities"
	"backend.juicedbot.io/juiced.infrastructure/common/enums"

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
	Product entities.WalmartProduct `json:"product"`
	OfferID string                  `json:"offerID"`
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
	DiscordWebhook string             `json:"discordWebhook"`
	Retailer       enums.Retailer     `json:"retailer"`
	MonitorID      primitive.ObjectID `json:"monitorID"`
	TargetData     TargetStockData    `json:"targetStockData"`
	WalmartData    WalmartStockData   `json:"walmartStockData"`
	AmazonData     AmazonStockData    `json:"amazonStockData"`
	BestbuyData    BestbuyStockData   `json:"bestbuyStockData"`
	GamestopData   GamestopStockData  `json:"gamestopStockData"`
}
