package commands

import (
	"backend.juicedbot.io/m/v2/juiced.infrastructure/common/entities"

	"context"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// CreateProfileGroup adds the ProfileGroup object to the database
func CreateProfileGroup(profileGroup entities.ProfileGroup) error {
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
	collection := client.Database("juiced").Collection("profile_groups")
	_, err = collection.InsertOne(ctx, profileGroup)
	return err
}

// RemoveProfileGroup removes the ProfileGroup from the database with the given groupID and returns it (if it exists)
func RemoveProfileGroup(groupID primitive.ObjectID) (entities.ProfileGroup, error) {
	client, err := mongo.NewClient(options.Client().ApplyURI("mongodb://localhost:27017"))
	profileGroup := entities.ProfileGroup{}
	if err != nil {
		return profileGroup, err
	}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	err = client.Connect(ctx)
	defer client.Disconnect(ctx)
	if err != nil {
		return profileGroup, err
	}
	collection := client.Database("juiced").Collection("profile_groups")
	filter := bson.D{primitive.E{Key: "groupid", Value: groupID}}
	err = collection.FindOneAndDelete(ctx, filter).Decode(&profileGroup)
	return profileGroup, err
}

// UpdateProfileGroup updates the ProfileGroup from the database with the given groupID and returns it (if it exists)
func UpdateProfileGroup(groupID primitive.ObjectID, newProfileGroup entities.ProfileGroup) (entities.ProfileGroup, error) {
	client, err := mongo.NewClient(options.Client().ApplyURI("mongodb://localhost:27017"))
	profileGroup := entities.ProfileGroup{}
	if err != nil {
		return profileGroup, err
	}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	err = client.Connect(ctx)
	defer client.Disconnect(ctx)
	if err != nil {
		return profileGroup, err
	}
	collection := client.Database("juiced").Collection("profile_groups")
	filter := bson.D{primitive.E{Key: "groupid", Value: groupID}}
	opts := options.FindOneAndReplace().SetReturnDocument(options.After)
	err = collection.FindOneAndReplace(ctx, filter, newProfileGroup, opts).Decode(&profileGroup)

	return profileGroup, err
}

// CreateProfile adds the Profile object to the database
func CreateProfile(profile entities.Profile) error {
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
	collection := client.Database("juiced").Collection("profiles")
	_, err = collection.InsertOne(ctx, profile)
	return err
}

// RemoveProfile removes the Profile from the database with the given ID and returns it (if it exists)
func RemoveProfile(ID primitive.ObjectID) (entities.Profile, error) {
	client, err := mongo.NewClient(options.Client().ApplyURI("mongodb://localhost:27017"))
	profile := entities.Profile{}
	if err != nil {
		return profile, err
	}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	err = client.Connect(ctx)
	defer client.Disconnect(ctx)
	if err != nil {
		return profile, err
	}
	collection := client.Database("juiced").Collection("profiles")
	filter := bson.D{primitive.E{Key: "id", Value: ID}}
	err = collection.FindOneAndDelete(ctx, filter).Decode(&profile)
	return profile, err
}

// UpdateProfile updates the Profile from the database with the given ID and returns it (if it exists)
func UpdateProfile(ID primitive.ObjectID, newProfile entities.Profile) (entities.Profile, error) {
	client, err := mongo.NewClient(options.Client().ApplyURI("mongodb://localhost:27017"))
	profile := entities.Profile{}
	if err != nil {
		return profile, err
	}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	err = client.Connect(ctx)
	defer client.Disconnect(ctx)
	if err != nil {
		return profile, err
	}
	collection := client.Database("juiced").Collection("profiles")
	filter := bson.D{primitive.E{Key: "id", Value: ID}}
	opts := options.FindOneAndReplace().SetReturnDocument(options.After)
	err = collection.FindOneAndReplace(ctx, filter, newProfile, opts).Decode(&profile)

	return profile, err
}
