package entities

import (
	"time"

	"backend.juicedbot.io/juiced.infrastructure/common/enums"
)

// Checkout contains all the info that is stored in the local db
type Checkout struct {
	ItemName    string         `json:"itemName" bson:"itemName"`
	SKU         string         `json:"sku" bson:"sku"`
	Price       int            `json:"price" bson:"price"`
	Quantity    int            `json:"quantity" bson:"quantity"`
	Retailer    enums.Retailer `json:"retailer" bson:"retailer"`
	ProfileName string         `json:"profileName" bson:"profileName"`
	Time        time.Time      `json:"time" bson:"time"`
}
