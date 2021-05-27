package commands

import (
	"backend.juicedbot.io/juiced.infrastructure/common/entities"

	"context"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// CreateProxyGroup adds the ProxyGroup object to the database
func CreateProxyGroup(proxyGroup entities.ProxyGroup) error {
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
	collection := client.Database("juiced").Collection("proxy_groups")
	_, err = collection.InsertOne(ctx, proxyGroup)
	return err
}

// RemoveProxyGroup removes the ProxyGroup from the database with the given groupID and returns it (if it exists)
func RemoveProxyGroup(groupID primitive.ObjectID) (entities.ProxyGroup, error) {
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
	err = collection.FindOneAndDelete(ctx, filter).Decode(&proxyGroup)
	return proxyGroup, err
}

// UpdateProxyGroup updates the ProxyGroup from the database with the given groupID and returns it (if it exists)
func UpdateProxyGroup(groupID primitive.ObjectID, newProxyGroup entities.ProxyGroup) (entities.ProxyGroup, error) {
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
	opts := options.FindOneAndReplace().SetReturnDocument(options.After)
	err = collection.FindOneAndReplace(ctx, filter, newProxyGroup, opts).Decode(&proxyGroup)

	return proxyGroup, err
}
