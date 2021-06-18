package errors

// ParseTaskGroupError is the error encountered when parsing JSON into a TaskGroup returns an error
const ParseTaskGroupError = "Parsing the JSON into a TaskGroup returned an error: "

// CreateTaskGroupError is the error encountered when inserting a TaskGroup into the DB returns an error
const CreateTaskGroupError = "Inserting the TaskGroup into the DB returned an error: "

// GetTaskGroupError is the error encountered when retrieving a TaskGroup from the DB returns an error
const GetTaskGroupError = "Retrieving the TaskGroup with the given ID returned an error: "

// GetAllTaskGroupsError is the error encountered when retrieving all TaskGroups from the DB returns an error
const GetAllTaskGroupsError = "Retrieving all TaskGroups returned an error: "

// RemoveTaskGroupError is the error encountered when removing a TaskGroup from the DB returns an error
const RemoveTaskGroupError = "Removing the TaskGroup with the given ID returned an error: "

// UpdateTaskGroupError is the error encountered when updating a TaskGroup from the DB returns an error
const UpdateTaskGroupError = "Updating the TaskGroup with the given ID returned an error: "

// StartTaskGroupError is the error encountered when starting a TaskGroup returns an error
const StartTaskGroupError = "Starting the TaskGroup encountered an error."

// StopTaskGroupError is the error encountered when stopping a TaskGroup returns an error
const StopTaskGroupError = "Stopping the TaskGroup encountered an error."

// ParseTaskError is the error encountered when parsing JSON into a Task returns an error
const ParseTaskError = "Parsing the JSON into a Task returned an error: "

// ParseDeleteTasksRequestError is the error encountere when parsing JSON into a DeleteTasksRequest object returns an error
const ParseDeleteTasksRequestError = "Parsing the JSON into a DeleteTasksRequest returned an error: "

// ParseUpdateTasksRequestError is the error encountere when parsing JSON into a UpdateTasksRequest object returns an error
const ParseUpdateTasksRequestError = "Parsing the JSON into a UpdateTasksRequest returned an error: "

// CreateTaskError is the error encountered when inserting a Task into the DB returns an error
const CreateTaskError = "Inserting the Task into the DB returned an error: "

// GetTaskError is the error encountered when retrieving a Task from the DB returns an error
const GetTaskError = "Retrieving the Task with the given ID returned an error: "

// GetAllTasksError is the error encountered when retrieving all Tasks from the DB returns an error
const GetAllTasksError = "Retrieving all Tasks returned an error: "

// RemoveTaskError is the error encountered when removing a Task from the DB returns an error
const RemoveTaskError = "Removing the Task with the given ID returned an error: "

// UpdateTaskError is the error encountered when updating a Task from the DB returns an error
const UpdateTaskError = "Updating the Task with the given ID returned an error: "

// AddTaskToGroupError is the error encountered when adding a Task to a TaskGroup in the DB returns an error
const AddTaskToGroupError = "Adding the Task to the TaskGroup with the given GroupID returned an error: "

// RemoveTasksFromGroupError is the error encountered when removing tasks from a TaskGroup in the DB returns an error
const RemoveTasksFromGroupError = "Removing the Tasks from the TaskGroup with the given GroupID returned an error: "

// StartTaskError is the error encountered when starting a Task returns an error
const StartTaskError = "Starting the Task encountered an error."

// StartTaskError is the error encountered when stopping a Task returns an error
const StopTaskError = "Stopping the Task encountered an error."
