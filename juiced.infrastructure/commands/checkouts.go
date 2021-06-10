package commands

import (
	"backend.juicedbot.io/juiced.infrastructure/common/entities"

	"context"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// CreateCheckout adds the Checkout object to the database
func CreateCheckout(checkout entities.Checkout) error {
	client, err := mongo.NewClient(options.Client().ApplyURI("mongodb://localhost:27017"))
	if err != nil {
		return err
	}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	err = client.Connect(ctx)
	defer client.Disconnect(ctx)
	if err != nil {
		return err
	}
	collection := client.Database("juiced").Collection("checkouts")
	_, err = collection.InsertOne(ctx, checkout)
	return err
}
