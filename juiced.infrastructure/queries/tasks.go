package queries

import (
	"backend.juicedbot.io/juiced.infrastructure/common/entities"

	"context"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// GetAllTaskGroups returns all TaskGroup objects from the database
func GetAllTaskGroups() ([]entities.TaskGroup, error) {
	client, err := mongo.NewClient(options.Client().ApplyURI("mongodb://localhost:27017"))
	taskGroups := make([]entities.TaskGroup, 0)
	if err != nil {
		return taskGroups, err
	}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	err = client.Connect(ctx)
	defer client.Disconnect(ctx)
	if err != nil {
		return taskGroups, err
	}
	collection := client.Database("juiced").Collection("task_groups")
	cursor, err := collection.Find(ctx, bson.M{})
	if err != nil {
		return taskGroups, err
	}
	defer cursor.Close(ctx)
	for cursor.Next(ctx) {
		var taskGroup entities.TaskGroup
		cursor.Decode(&taskGroup)
		taskGroups = append(taskGroups, taskGroup)
	}
	err = cursor.Err()
	return taskGroups, err
}

// GetTaskGroup returns the TaskGroup object from the database with the given groupID (if it exists)
func GetTaskGroup(groupID primitive.ObjectID) (entities.TaskGroup, error) {
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
	err = collection.FindOne(ctx, filter).Decode(&taskGroup)
	return taskGroup, err
}

// GetAllTasks returns all Task objects from the database
func GetAllTasks() ([]entities.Task, error) {
	client, err := mongo.NewClient(options.Client().ApplyURI("mongodb://localhost:27017"))
	tasks := make([]entities.Task, 0)
	if err != nil {
		return tasks, err
	}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	err = client.Connect(ctx)
	defer client.Disconnect(ctx)
	if err != nil {
		return tasks, err
	}
	collection := client.Database("juiced").Collection("tasks")
	cursor, err := collection.Find(ctx, bson.M{})
	if err != nil {
		return tasks, err
	}
	defer cursor.Close(ctx)
	for cursor.Next(ctx) {
		var task entities.Task
		cursor.Decode(&task)
		tasks = append(tasks, task)
	}
	err = cursor.Err()
	return tasks, err
}

// GetTask returns the Task object from the database with the given ID (if it exists)
func GetTask(ID primitive.ObjectID) (entities.Task, error) {
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
	err = collection.FindOne(ctx, filter).Decode(&task)
	return task, err
}

// ConvertTaskIDsToTasks returns a TaskGroupWithTasks object from a TaskGroup object
func ConvertTaskIDsToTasks(taskGroup *entities.TaskGroup) (entities.TaskGroupWithTasks, error) {
	taskGroupWithTasks := entities.TaskGroupWithTasks{
		GroupID: taskGroup.GroupID, Name: taskGroup.Name,
		MonitorProxyGroupID: taskGroup.MonitorProxyGroupID,
		MonitorRetailer:     taskGroup.MonitorRetailer,
		MonitorDelay:        taskGroup.MonitorDelay,
		MonitorStatus:       taskGroup.MonitorStatus,
		TargetMonitorInfo:   taskGroup.TargetMonitorInfo,
		WalmartMonitorInfo:  taskGroup.WalmartMonitorInfo,
		AmazonMonitorInfo:   taskGroup.AmazonMonitorInfo,
		BestbuyMonitorInfo:  taskGroup.BestbuyMonitorInfo,
		Tasks:               []entities.Task{},
	}
	tasks := []entities.Task{}
	for i := 0; i < len(taskGroup.TaskIDs); i++ {
		task, err := GetTask(taskGroup.TaskIDs[i])
		if err != nil {
			return taskGroupWithTasks, err
		}
		tasks = append(tasks, task)
	}
	taskGroupWithTasks.SetTasks(tasks)
	return taskGroupWithTasks, nil
}
