package stores

import (
	"time"

	"backend.juicedbot.io/juiced.infrastructure/database"
	"backend.juicedbot.io/juiced.infrastructure/entities"
	"backend.juicedbot.io/juiced.infrastructure/enums"
)

type CheckoutStore struct {
	Checkouts []entities.Checkout
}

var checkoutsStore CheckoutStore

func (store *CheckoutStore) Init() error {
	var err error
	store.Checkouts, err = database.GetCheckouts()
	return err
}

func GetCheckouts(retailer enums.Retailer, daysBack int) []entities.Checkout {
	checkouts := []entities.Checkout{}

	then := time.Now().Add(time.Duration(-daysBack) * (24 * time.Hour)).Unix()
	for _, checkout := range checkoutsStore.Checkouts {
		// No filters
		if retailer == "" && daysBack == -1 {
			checkouts = append(checkouts, checkout)
		}
		// Only time filter
		if retailer == "" && daysBack != -1 {
			if then > checkout.Time {
				checkouts = append(checkouts, checkout)
			}
		}
		// Only retailer filter
		if retailer != "" && daysBack == -1 {
			if checkout.Retailer == retailer {
				checkouts = append(checkouts, checkout)
			}
		}
		// Both retailer and time filter
		if retailer != "" && daysBack != -1 {
			if checkout.Retailer == retailer && then > checkout.Time {
				checkouts = append(checkouts, checkout)
			}
		}
	}

	return checkouts
}

func CreateCheckout(checkout entities.Checkout) error {
	err := database.CreateCheckout(checkout)
	if err != nil {
		return err
	}
	checkoutsStore.Checkouts = append(checkoutsStore.Checkouts, checkout)
	return nil
}
