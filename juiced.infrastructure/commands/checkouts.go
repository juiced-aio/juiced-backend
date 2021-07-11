package commands

import (
	"errors"

	"backend.juicedbot.io/juiced.infrastructure/common"
	"backend.juicedbot.io/juiced.infrastructure/common/entities"
	_ "github.com/jmoiron/sqlx"
	_ "github.com/mattn/go-sqlite3"
)

// CreateCheckout adds the Checkout object to the database
func CreateCheckout(checkout entities.Checkout) error {
	database := common.GetDatabase()
	if database == nil {
		return errors.New("database not initialized")
	}

	statement, err := database.Preparex(`INSERT INTO checkouts (itemName, sku, price, quantity, retailer, profileName, msToCheckout, time) VALUES (?, ?, ?, ?, ?, ?, ?, ?)`)
	if err != nil {
		return err
	}

	_, err = statement.Exec(checkout.ItemName, checkout.SKU, checkout.Price, checkout.Quantity, checkout.Retailer, checkout.ProfileName, checkout.MsToCheckout, checkout.Time)
	if err != nil {
		return err
	}

	return err
}
