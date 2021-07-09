package queries

import (
	"math/rand"
	"reflect"
	"testing"

	"backend.juicedbot.io/juiced.infrastructure/common"
	"backend.juicedbot.io/juiced.infrastructure/common/entities"
)

var allTaskGroups []entities.TaskGroup
var allTasks []entities.Task

func TestMain(m *testing.M) {
	common.InitDatabase()
	var err error
	allTaskGroups, err = GetAllTaskGroups()
	if err != nil {
		panic(err)
	}
	allTasks, err = GetAllTasks()
	if err != nil {
		panic(err)
	}

	m.Run()
}

func TestGetAllTaskGroups(t *testing.T) {
	tests := []struct {
		name    string
		wantErr bool
	}{
		{name: "Success", wantErr: false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := GetAllTaskGroups()
			if (err != nil) != tt.wantErr {
				t.Errorf("GetAllTaskGroups() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
		})
	}
}

func TestGetTaskGroup(t *testing.T) {
	currentGroup := allTaskGroups[rand.Intn(len(allTaskGroups))]
	type args struct {
		groupID string
	}
	tests := []struct {
		name    string
		args    args
		want    entities.TaskGroup
		wantErr bool
	}{
		{name: "Found", args: args{groupID: currentGroup.GroupID}, want: currentGroup, wantErr: false},
		{name: "No GroupID", args: args{groupID: ""}, want: entities.TaskGroup{}, wantErr: false},
		{name: "Doesn't Exist", args: args{groupID: "NOT REAL"}, want: entities.TaskGroup{}, wantErr: false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := GetTaskGroup(tt.args.groupID)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetTaskGroup() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetTaskGroup() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGetTaskGroupsByProxyGroupID(t *testing.T) {
	currentGroup := allTaskGroups[rand.Intn(len(allTaskGroups))]
	type args struct {
		proxyGroupID string
	}

	//MonitorProxyGroupID can be empty so this is only testing for errors
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{name: "Success", wantErr: false, args: args{proxyGroupID: currentGroup.MonitorProxyGroupID}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := GetTaskGroupsByProxyGroupID(tt.args.proxyGroupID)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetTaskGroupsByProxyGroupID() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
		})
	}
}

func TestGetAllTasks(t *testing.T) {
	tests := []struct {
		name    string
		wantErr bool
	}{
		{name: "Success", wantErr: false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := GetAllTasks()
			if (err != nil) != tt.wantErr {
				t.Errorf("GetAllTasks() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
		})
	}
}

func TestGetTasksByProfileID(t *testing.T) {
	currentTask := allTasks[rand.Intn(len(allTasks))]
	type args struct {
		profileID string
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{name: "Success", wantErr: false, args: args{profileID: currentTask.TaskProfileID}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := GetTasksByProfileID(tt.args.profileID)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetTasksByProfileID() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
		})
	}
}

func TestGetTasksByProxyGroupID(t *testing.T) {
	currentTask := allTasks[rand.Intn(len(allTasks))]
	type args struct {
		proxyGroupID string
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{name: "Success", wantErr: false, args: args{proxyGroupID: currentTask.TaskProxyGroupID}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := GetTasksByProxyGroupID(tt.args.proxyGroupID)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetTasksByProxyGroupID() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
		})
	}
}

func TestGetTask(t *testing.T) {
	currentTask := allTasks[rand.Intn(len(allTasks))]
	type args struct {
		ID string
	}
	tests := []struct {
		name    string
		args    args
		want    entities.Task
		wantErr bool
	}{
		{name: "Success", wantErr: false, args: args{ID: currentTask.ID}, want: currentTask},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := GetTask(tt.args.ID)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetTask() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetTask() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGetSettings(t *testing.T) {
	tests := []struct {
		name    string
		wantErr bool
	}{
		{name: "Success", wantErr: false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := GetSettings()
			if (err != nil) != tt.wantErr {
				t.Errorf("GetSettings() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
		})
	}
}
