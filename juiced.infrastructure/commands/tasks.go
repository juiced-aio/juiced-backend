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

// CreateTaskGroup adds the TaskGroup object to the database
func CreateTaskGroup(taskGroup entities.TaskGroup) error {
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
	collection := client.Database("juiced").Collection("task_groups")
	_, err = collection.InsertOne(ctx, taskGroup)
	return err
}

// RemoveTaskGroup removes the TaskGroup from the database with the given groupID and returns it (if it exists)
func RemoveTaskGroup(groupID primitive.ObjectID) (entities.TaskGroup, error) {
	client, err := mongo.NewClient(options.Client().ApplyURI("mongodb://localhost:27017"))
	taskGroup := entities.TaskGroup{}
	if err != nil {
		return taskGroup, err
	}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	err = client.Connect(ctx)
	defer client.Disconnect(ctx)
	if err != nil {
		return taskGroup, err
	}
	collection := client.Database("juiced").Collection("task_groups")
	filter := bson.D{primitive.E{Key: "groupID", Value: groupID}}
	err = collection.FindOneAndDelete(ctx, filter).Decode(&taskGroup)
	return taskGroup, err
}

// UpdateTaskGroup updates the TaskGroup from the database with the given groupID and returns it (if it exists)
func UpdateTaskGroup(groupID primitive.ObjectID, newTaskGroup entities.TaskGroup) (entities.TaskGroup, error) {
	client, err := mongo.NewClient(options.Client().ApplyURI("mongodb://localhost:27017"))
	taskGroup := entities.TaskGroup{}
	if err != nil {
		return taskGroup, err
	}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	err = client.Connect(ctx)
	defer client.Disconnect(ctx)
	if err != nil {
		return taskGroup, err
	}
	collection := client.Database("juiced").Collection("task_groups")
	filter := bson.D{primitive.E{Key: "groupID", Value: groupID}}
	opts := options.FindOneAndReplace().SetReturnDocument(options.After)
	err = collection.FindOneAndReplace(ctx, filter, newTaskGroup, opts).Decode(&taskGroup)

	return taskGroup, err
}

// CreateTask adds the Task object to the database
func CreateTask(task entities.Task) error {
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
	collection := client.Database("juiced").Collection("tasks")
	_, err = collection.InsertOne(ctx, task)
	return err
}

// RemoveTask removes the Task from the database with the given ID and returns it (if it exists)
func RemoveTask(ID primitive.ObjectID) (entities.Task, error) {
	client, err := mongo.NewClient(options.Client().ApplyURI("mongodb://localhost:27017"))
	task := entities.Task{}
	if err != nil {
		return task, err
	}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	err = client.Connect(ctx)
	defer client.Disconnect(ctx)
	if err != nil {
		return task, err
	}
	collection := client.Database("juiced").Collection("tasks")
	filter := bson.D{primitive.E{Key: "id", Value: ID}}
	err = collection.FindOneAndDelete(ctx, filter).Decode(&task)
	return task, err
}

// UpdateTask updates the Task from the database with the given ID and returns it (if it exists)
func UpdateTask(ID primitive.ObjectID, newTask entities.Task) (entities.Task, error) {
	client, err := mongo.NewClient(options.Client().ApplyURI("mongodb://localhost:27017"))
	task := entities.Task{}
	if err != nil {
		return task, err
	}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	err = client.Connect(ctx)
	defer client.Disconnect(ctx)
	if err != nil {
		return task, err
	}
	collection := client.Database("juiced").Collection("tasks")
	filter := bson.D{primitive.E{Key: "id", Value: ID}}
	opts := options.FindOneAndReplace().SetReturnDocument(options.After)
	err = collection.FindOneAndReplace(ctx, filter, newTask, opts).Decode(&task)

	return task, err
}

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
