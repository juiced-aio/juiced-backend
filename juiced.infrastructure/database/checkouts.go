package database

import (
	"backend.juicedbot.io/juiced.infrastructure/entities"
	"backend.juicedbot.io/juiced.infrastructure/enums"

	"time"
)

// GetCheckouts returns all the checkouts in the given timeframe and retailer
func GetCheckouts(retailer enums.Retailer, daysBack int) ([]entities.Checkout, error) {
	then := time.Now().Add(time.Duration(-daysBack) * (24 * time.Hour)).Unix()

	checkouts := []entities.Checkout{}
	if database == nil {
		return checkouts, &DatabaseNotInitializedError{}
	}

	rows, err := database.Queryx("SELECT * FROM checkouts")
	if err != nil {
		return checkouts, err
	}

	emptyString := ""
	defer rows.Close()
	for rows.Next() {
		tempCheckout := entities.Checkout{}
		err = rows.StructScan(&tempCheckout)
		if err != nil {
			return checkouts, err
		}
		// No filters
		if retailer == emptyString && daysBack == -1 {
			checkouts = append(checkouts, tempCheckout)
		}
		// Only time filter
		if retailer == emptyString && daysBack != -1 {
			if then > tempCheckout.Time {
				checkouts = append(checkouts, tempCheckout)
			}
		}
		// Only retailer filter
		if retailer != emptyString && daysBack == -1 {
			if tempCheckout.Retailer == retailer {
				checkouts = append(checkouts, tempCheckout)
			}
		}
		// Both retailer and time filter
		if retailer != emptyString && daysBack != -1 {
			if tempCheckout.Retailer == retailer && then > tempCheckout.Time {
				checkouts = append(checkouts, tempCheckout)
			}
		}
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
