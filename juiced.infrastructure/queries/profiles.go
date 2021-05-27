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

// GetAllProfileGroups returns all ProfileGroup objects from the database
func GetAllProfileGroups() ([]entities.ProfileGroup, error) {
	client, err := mongo.NewClient(options.Client().ApplyURI("mongodb://localhost:27017"))
	profileGroups := make([]entities.ProfileGroup, 0)
	if err != nil {
		return profileGroups, err
	}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	err = client.Connect(ctx)
	defer client.Disconnect(ctx)
	if err != nil {
		return profileGroups, err
	}
	collection := client.Database("juiced").Collection("profile_groups")
	cursor, err := collection.Find(ctx, bson.M{})
	if err != nil {
		return profileGroups, err
	}
	defer cursor.Close(ctx)
	for cursor.Next(ctx) {
		var profileGroup entities.ProfileGroup
		cursor.Decode(&profileGroup)
		profileGroups = append(profileGroups, profileGroup)
	}
	err = cursor.Err()
	return profileGroups, err
}

// GetProfileGroup returns the ProfileGroup object from the database with the given groupID (if it exists)
func GetProfileGroup(groupID primitive.ObjectID) (entities.ProfileGroup, error) {
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
	err = collection.FindOne(ctx, filter).Decode(&profileGroup)
	return profileGroup, err
}

// GetAllProfiles returns all Profile objects from the database
func GetAllProfiles() ([]entities.Profile, error) {
	client, err := mongo.NewClient(options.Client().ApplyURI("mongodb://localhost:27017"))
	profiles := make([]entities.Profile, 0)
	if err != nil {
		return profiles, err
	}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	err = client.Connect(ctx)
	defer client.Disconnect(ctx)
	if err != nil {
		return profiles, err
	}
	collection := client.Database("juiced").Collection("profiles")
	cursor, err := collection.Find(ctx, bson.M{})
	if err != nil {
		return profiles, err
	}
	defer cursor.Close(ctx)
	for cursor.Next(ctx) {
		var profile entities.Profile
		cursor.Decode(&profile)
		profiles = append(profiles, profile)
	}
	err = cursor.Err()
	return profiles, err
}

// GetProfile returns the Profile object from the database with the given ID (if it exists)
func GetProfile(ID primitive.ObjectID) (entities.Profile, error) {
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
	err = collection.FindOne(ctx, filter).Decode(&profile)
	return profile, err
}

// ConvertProfileIDsToProfiles returns a ProfileGroupWithProfiles object from a ProfileGroup object
func ConvertProfileIDsToProfiles(profileGroup *entities.ProfileGroup) (entities.ProfileGroupWithProfiles, error) {
	profileGroupWithProfiles := entities.ProfileGroupWithProfiles{GroupID: profileGroup.GroupID, Name: profileGroup.Name, Profiles: []entities.Profile{}}
	profiles := []entities.Profile{}
	for i := 0; i < len(profileGroup.ProfileIDs); i++ {
		profile, err := GetProfile(profileGroup.ProfileIDs[i])
		if err != nil {
			if err.Error() != "mongo: no documents in result" {
				return profileGroupWithProfiles, err
			}
		} else {
			profiles = append(profiles, profile)
		}
	}
	profileGroupWithProfiles.SetProfiles(profiles)
	return profileGroupWithProfiles, nil
}

// ConvertProfilesToProfileIDs returns a ProfileGroup object from a ProfileGroupWithProfiles object
func ConvertProfilesToProfileIDs(profileGroupWithProfiles *entities.ProfileGroupWithProfiles) (entities.ProfileGroup, error) {
	profileGroup := entities.ProfileGroup{GroupID: profileGroupWithProfiles.GroupID, Name: profileGroupWithProfiles.Name, ProfileIDs: []primitive.ObjectID{}}
	profileIDs := []primitive.ObjectID{}
	for i := 0; i < len(profileGroupWithProfiles.Profiles); i++ {
		profileID := profileGroupWithProfiles.Profiles[i].ID
		profileIDs = append(profileIDs, profileID)
	}
	profileGroup.SetProfileIDs(profileIDs)
	return profileGroup, nil
}
