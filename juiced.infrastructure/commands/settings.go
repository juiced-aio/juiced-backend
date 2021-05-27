package commands

import (
	"context"
	"time"

	"backend.juicedbot.io/m/v2/juiced.infrastructure/common/entities"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// UpdateSettings updates the Settings object in the database
func UpdateSettings(newSettings entities.Settings) (entities.Settings, error) {
	client, err := mongo.NewClient(options.Client().ApplyURI("mongodb://localhost:27017"))
	settings := entities.Settings{}
	if err != nil {
		return settings, err
	}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	err = client.Connect(ctx)
	defer client.Disconnect(ctx)
	if err != nil {
		return settings, err
	}
	collection := client.Database("juiced").Collection("settings")
	filter := bson.D{primitive.E{Key: "id", Value: 0}}
	opts := options.FindOneAndReplace().SetReturnDocument(options.After)
	err = collection.FindOneAndReplace(ctx, filter, newSettings, opts).Decode(&settings)

	return settings, err
}
