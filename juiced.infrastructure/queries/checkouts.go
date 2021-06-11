package queries

import (
	"backend.juicedbot.io/juiced.infrastructure/common/entities"
	"backend.juicedbot.io/juiced.infrastructure/common/enums"

	"context"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// GetCheckouts returns all the checkouts in the given timeframe and retailer
func GetCheckouts(retailer enums.Retailer, daysBack int) ([]entities.Checkout, error) {
	then := time.Now().Add(time.Duration(-daysBack) * (24 * time.Hour))
	now := time.Now()

	client, err := mongo.NewClient(options.Client().ApplyURI("mongodb://localhost:27017"))
	checkouts := make([]entities.Checkout, 0)
	if err != nil {
		return checkouts, err
	}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	err = client.Connect(ctx)
	defer client.Disconnect(ctx)
	if err != nil {
		return checkouts, err
	}
	collection := client.Database("juiced").Collection("checkouts")

	var cursor *mongo.Cursor
	// No filters
	if retailer == "" && daysBack == -1 {
		cursor, err = collection.Find(ctx, bson.M{})
	}
	// Only time filter
	if retailer == "" && daysBack != -1 {
		cursor, err = collection.Find(ctx, bson.M{
			"time": bson.M{
				"$gt": then,
				"$lt": now,
			},
		})
	}
	// Only retailer filter
	if retailer != "" && daysBack == -1 {
		cursor, err = collection.Find(ctx, bson.M{
			"retailer": retailer,
		})
	}
	// Both retailer and time filter
	if retailer != "" && daysBack != -1 {
		cursor, err = collection.Find(ctx, bson.M{
			"time": bson.M{
				"$gt": then,
				"$lt": now,
			},
			"retailer": retailer,
		})
	}
	if err != nil {
		return checkouts, err
	}
	defer cursor.Close(ctx)
	for cursor.Next(ctx) {
		var checkout entities.Checkout
		cursor.Decode(&checkout)
		checkouts = append(checkouts, checkout)
	}
	err = cursor.Err()
	return checkouts, err
}
