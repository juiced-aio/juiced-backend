package queries

import (
	"errors"

	"backend.juicedbot.io/juiced.infrastructure/common"
	"backend.juicedbot.io/juiced.infrastructure/common/entities"
	"backend.juicedbot.io/juiced.infrastructure/common/enums"

	"time"
)

// GetCheckouts returns all the checkouts in the given timeframe and retailer
func GetCheckouts(retailer enums.Retailer, daysBack int) ([]entities.Checkout, error) {
	then := time.Now().Add(time.Duration(-daysBack) * (24 * time.Hour)).Unix()

	checkouts := []entities.Checkout{}
	database := common.GetDatabase()
	if database == nil {
		return checkouts, errors.New("database not initialized")
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
