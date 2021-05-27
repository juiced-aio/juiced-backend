package queries

import (
	"backend.juicedbot.io/m/v2/juiced.infrastructure/common/entities"

	"context"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// GetAllProxyGroups returns all ProxyGroup objects from the database
func GetAllProxyGroups() ([]entities.ProxyGroup, error) {
	client, err := mongo.NewClient(options.Client().ApplyURI("mongodb://localhost:27017"))
	proxyGroups := make([]entities.ProxyGroup, 0)
	if err != nil {
		return proxyGroups, err
	}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	err = client.Connect(ctx)
	defer client.Disconnect(ctx)
	if err != nil {
		return proxyGroups, err
	}
	collection := client.Database("juiced").Collection("proxy_groups")
	cursor, err := collection.Find(ctx, bson.M{})
	if err != nil {
		return proxyGroups, err
	}
	defer cursor.Close(ctx)
	for cursor.Next(ctx) {
		var proxyGroup entities.ProxyGroup
		cursor.Decode(&proxyGroup)
		proxyGroups = append(proxyGroups, proxyGroup)
	}
	err = cursor.Err()
	return proxyGroups, err
}

// GetProxyGroup returns the ProxyGroup object from the database with the given groupID (if it exists)
func GetProxyGroup(groupID primitive.ObjectID) (entities.ProxyGroup, error) {
	client, err := mongo.NewClient(options.Client().ApplyURI("mongodb://localhost:27017"))
	proxyGroup := entities.ProxyGroup{}
	if err != nil {
		return proxyGroup, err
	}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	err = client.Connect(ctx)
	defer client.Disconnect(ctx)
	if err != nil {
		return proxyGroup, err
	}
	collection := client.Database("juiced").Collection("proxy_groups")
	filter := bson.D{primitive.E{Key: "groupid", Value: groupID}}
	err = collection.FindOne(ctx, filter).Decode(&proxyGroup)
	return proxyGroup, err
}
