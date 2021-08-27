package database

import (
	"backend.juicedbot.io/juiced.infrastructure/entities"
)

// GetCheckouts returns all the checkouts in the given timeframe and retailer
func GetCheckouts() ([]entities.Checkout, error) {

	checkouts := []entities.Checkout{}
	if database == nil {
		return checkouts, &DatabaseNotInitializedError{}
	}

	rows, err := database.Queryx("SELECT * FROM checkouts")
	if err != nil {
		return checkouts, err
	}

	defer rows.Close()
	for rows.Next() {
		tempCheckout := entities.Checkout{}
		err = rows.StructScan(&tempCheckout)
		if err != nil {
			return checkouts, err
		}

	}
	return checkouts, err
}

// CreateCheckout adds the Checkout object to the database
func CreateCheckout(checkout entities.Checkout) error {
	if database == nil {
		return &DatabaseNotInitializedError{}
	}

	statement, err := database.Preparex(`INSERT INTO checkouts (itemName, imageURL, sku, price, quantity, retailer, profileName, msToCheckout, time) VALUES (?, ?, ?, ?, ?, ?, ?, ?)`)
	if err != nil {
		return err
	}

	_, err = statement.Exec(checkout.ItemName, checkout.ImageURL, checkout.SKU, checkout.Price, checkout.Quantity, checkout.Retailer, checkout.ProfileName, checkout.MsToCheckout, checkout.Time)
	if err != nil {
		return err
	}

	return err
}
