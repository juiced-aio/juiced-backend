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

// ProductEvent is fired whenever a Monitor finds a product
type ProductEvent struct {
	DiscordWebhook string             `json:"discordWebhook"`
	Retailer       enums.Retailer     `json:"retailer"`
	MonitorID      primitive.ObjectID `json:"monitorID"`
	TargetData     TargetStockData    `json:"targetStockData"`
	WalmartData    WalmartStockData   `json:"walmartStockData"`
}
