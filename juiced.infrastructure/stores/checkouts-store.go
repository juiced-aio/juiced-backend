package stores

import (
	"backend.juicedbot.io/juiced.infrastructure/database"
	"backend.juicedbot.io/juiced.infrastructure/entities"
)

type CheckoutStore struct {
	Checkouts []entities.Checkout
}

var checkoutsStore CheckoutStore

func (store *CheckoutStore) Init() error {
	var err error
	store.Checkouts, err = database.GetCheckouts("", -1)
	return err
}

func GetCheckouts() []entities.Checkout {
	return checkoutsStore.Checkouts
}

func CreateCheckout(checkout entities.Checkout) error {
	err := database.CreateCheckout(checkout)
	if err != nil {
		return err
	}
	checkoutsStore.Checkouts = append(checkoutsStore.Checkouts, checkout)
	return nil
}
