package entities

import (
	"backend.juicedbot.io/juiced.infrastructure/common/enums"
)

// Checkout contains all the info that is stored in the local db
type Checkout struct {
	ItemName    string         `json:"itemName" db:"itemName"`
	SKU         string         `json:"sku" db:"sku"`
	Price       int            `json:"price" db:"price"`
	Quantity    int            `json:"quantity" db:"quantity"`
	Retailer    enums.Retailer `json:"retailer" db:"retailer"`
	ProfileName string         `json:"profileName" db:"profileName"`
	Time        int64          `json:"time" db:"time"`
}
