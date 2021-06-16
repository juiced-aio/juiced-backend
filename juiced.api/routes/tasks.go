package routes

import (
	"backend.juicedbot.io/juiced.api/endpoints"

	"github.com/gorilla/mux"
)

// RouteTasksEndpoints routes endpoints that handle tasks and task groups
func RouteTasksEndpoints(router *mux.Router) {
	// swagger:operation GET /api/task/group TaskGroup GetAllTaskGroupsEndpoint
	//
	// Returns a list of all TaskGroups
	//
	// ---
	// responses:
	//   '200':
	//     description: TaskGroups response
	//     schema:
	//       "$ref": "#/responses/TaskGroupsResponseSwagger"
	router.HandleFunc("/api/task/group", endpoints.GetAllTaskGroupsEndpoint).Methods("GET")

	// swagger:operation GET /api/task/group/{GroupID} TaskGroup GetTaskGroupEndpoint
	//
	// Returns the TaskGroups with GroupID {GroupID}.
	//
	// ---
	// parameters:
	// - name: GroupID
	//   in: path
	//   description: ID of TaskGroups to retrieve
	//   type: string
	//   required: true
	// responses:
	//   '200':
	//     description: TaskGroups response
	//     schema:
	//       "$ref": "#/responses/TaskGroupsResponseSwagger"
	router.HandleFunc("/api/task/group/{GroupID}", endpoints.GetTaskGroupEndpoint).Methods("GET")

	// swagger:operation POST /api/task/group TaskGroup CreateTaskGroupEndpoint
	//
	// Creates a TaskGroup in the database
	//
	// ---
	// parameters:
	// - name: TaskGroupDetails
	//   in: body
	//   description: TaskGroup details
	//   required: true
	//   schema:
	//     "$ref": "#/models/CreateOrUpdateTaskGroupRequest"
	// responses:
	//   '200':
	//     description: TaskGroup response
	//     schema:
	//       "$ref": "#/responses/TaskGroupResponseSwagger"
	router.HandleFunc("/api/task/group", endpoints.CreateTaskGroupEndpoint).Methods("POST")

	// swagger:operation DELETE /api/task/group/{GroupID} TaskGroup DeleteTaskGroupEndpoint
	//
	// Deletes and returns the TaskGroup with GroupID {GroupID}.
	//
	// ---
	// parameters:
	// - name: GroupID
	//   in: path
	//   description: ID of TaskGroup to retrieve
	//   type: string
	//   required: true
	// responses:
	//   '200':
	//     description: TaskGroup response
	//     schema:
	//       "$ref": "#/responses/TaskGroupResponseSwagger"
	router.HandleFunc("/api/task/group/{GroupID}", endpoints.RemoveTaskGroupEndpoint).Methods("DELETE")

	// swagger:operation PUT /api/task/group/{GroupID} TaskGroup UpdateTaskGroupEndpoint
	//
	// Updates and returns the TaskGroup with GroupID {GroupID}.
	//
	// ---
	// parameters:
	// - name: GroupID
	//   in: path
	//   description: ID of TaskGroup to update
	//   type: string
	//   required: false
	// - name: TaskGroupDetails
	//   in: body
	//   description: Details to update TaskGroup with
	//   required: true
	//   schema:
	//     "$ref": "#/models/CreateOrUpdateTaskGroupRequest"
	// responses:
	//   '200':
	//     description: TaskGroup response
	//     schema:
	//       "$ref": "#/responses/TaskGroupResponseSwagger"
	router.HandleFunc("/api/task/group/{GroupID}", endpoints.UpdateTaskGroupEndpoint).Methods("PUT")

	// swagger:operation POST /api/task/group/{GroupID}/clone TaskGroup CloneTaskGroupEndpoint
	//
	// Clones the TaskGroup with GroupID {GroupID} and returns the clone.
	//
	// ---
	// parameters:
	// - name: GroupID
	//   in: path
	//   description: ID of TaskGroup to clone
	//   type: string
	//   required: false
	// responses:
	//   '200':
	//     description: TaskGroup response
	//     schema:
	//       "$ref": "#/responses/TaskGroupResponseSwagger"
	router.HandleFunc("/api/task/group/{GroupID}/clone", endpoints.CloneTaskGroupEndpoint).Methods("POST")

	// swagger:operation POST /api/task/group/{GroupID}/start TaskGroup StartTaskGroupEndpoint
	//
	// Starts a TaskGroup's Monitor and all of its Tasks
	//
	// ---
	// parameters:
	// - name: GroupID
	//   in: path
	//   description: GroupID of TaskGroup to start
	//   type: string
	//   required: false
	// responses:
	//   '200':
	//     description: TaskGroup response
	//     schema:
	//       "$ref": "#/responses/TaskGroupResponseSwagger"
	router.HandleFunc("/api/task/group/{GroupID}/start", endpoints.StartTaskGroupEndpoint).Methods("POST")

	// swagger:operation POST /api/task/group/{GroupID}/stop TaskGroup StopTaskGroupEndpoint
	//
	// Stops a TaskGroup's Monitor and all of its Tasks
	//
	// ---
	// parameters:
	// - name: GroupID
	//   in: path
	//   description: GroupID of TaskGroup to stop
	//   type: string
	//   required: false
	// responses:
	//   '200':
	//     description: TaskGroup response
	//     schema:
	//       "$ref": "#/responses/TaskGroupResponseSwagger"
	router.HandleFunc("/api/task/group/{GroupID}/stop", endpoints.StopTaskGroupEndpoint).Methods("POST")

	// swagger:operation POST /api/task/group/{GroupID}/removeTasks TaskGroup RemoveTasksEndpoint
	//
	// Deletes Tasks from the group with GroupID {GroupID}.
	//
	// ---
	// parameters:
	// - name: GroupID
	//   in: path
	//   description: GroupID of Tasks to remove
	//   type: string
	//   required: true
	// - name: TaskIDs
	//   in: body
	//   description: IDs of Tasks to remove from Group
	//   required: true
	//   schema:
	//     "$ref": "#/models/CreateOrUpdateTaskRequest"
	// responses:
	//   '200':
	//     description: Task response
	//     schema:
	//       "$ref": "#/responses/TaskResponseSwagger"
	router.HandleFunc("/api/task/group/{GroupID}/removeTasks", endpoints.RemoveTasksEndpoint).Methods("POST")

	// swagger:operation GET /api/task Task GetAllTasksEndpoint
	//
	// Returns a list of all Tasks
	//
	// ---
	// responses:
	//   '200':
	//     description: Task response
	//     schema:
	//       "$ref": "#/responses/TaskResponseSwagger"
	router.HandleFunc("/api/task", endpoints.GetAllTasksEndpoint).Methods("GET")

	// swagger:operation GET /api/task/{ID} Task GetTaskEndpoint
	//
	// Returns the Task with ID {ID}.
	//
	// ---
	// parameters:
	// - name: ID
	//   in: path
	//   description: ID of Task to retrieve
	//   type: string
	//   required: true
	// responses:
	//   '200':
	//     description: Task response
	//     schema:
	//       "$ref": "#/responses/TaskResponseSwagger"
	router.HandleFunc("/api/task/{ID}", endpoints.GetTaskEndpoint).Methods("GET")

	// swagger:operation POST /api/task/{GroupID} Task CreateTaskEndpoint
	//
	// Creates a Task in the database
	//
	// ---
	// parameters:
	// - name: GroupID
	//   in: path
	//   description: ID of TaskGroup to add task to
	//   type: string
	//   required: false
	// - name: TaskDetails
	//   in: body
	//   description: Task details
	//   required: true
	//   schema:
	//     "$ref": "#/models/CreateOrUpdateTaskRequest"
	// responses:
	//   '200':
	//     description: Task response
	//     schema:
	//       "$ref": "#/responses/TaskResponseSwagger"
	router.HandleFunc("/api/task/{GroupID}", endpoints.CreateTaskEndpoint).Methods("POST")

	// swagger:operation PUT /api/task/{ID} Task UpdateTaskEndpoint
	//
	// Updates and returns the Task with ID {ID}.
	//
	// ---
	// parameters:
	// - name: ID
	//   in: path
	//   description: ID of Task to update
	//   type: string
	//   required: false
	// - name: TaskDetails
	//   in: body
	//   description: Details to update Task with
	//   required: true
	//   schema:
	//     "$ref": "#/models/CreateOrUpdateTaskRequest"
	// responses:
	//   '200':
	//     description: Task response
	//     schema:
	//       "$ref": "#/responses/TaskResponseSwagger"
	router.HandleFunc("/api/task/{ID}", endpoints.UpdateTaskEndpoint).Methods("PUT")

	// swagger:operation POST /api/task/{ID}/clone Task CloneTaskEndpoint
	//
	// Clones the Task with ID {ID} and returns the clone.
	//
	// ---
	// parameters:
	// - name: ID
	//   in: path
	//   description: ID of Task to clone
	//   type: string
	//   required: false
	// responses:
	//   '200':
	//     description: Task response
	//     schema:
	//       "$ref": "#/responses/TaskResponseSwagger"
	router.HandleFunc("/api/task/{ID}/clone", endpoints.CloneTaskEndpoint).Methods("POST")

	// swagger:operation POST /api/task/{ID}/start Task StartTaskEndpoint
	//
	// Starts a Task
	//
	// ---
	// parameters:
	// - name: ID
	//   in: path
	//   description: ID of Task to start
	//   type: string
	//   required: false
	// responses:
	//   '200':
	//     description: Task response
	//     schema:
	//       "$ref": "#/responses/TaskResponseSwagger"
	router.HandleFunc("/api/task/{ID}/start", endpoints.StartTaskEndpoint).Methods("POST")

	// swagger:operation POST /api/task/{ID}/stop Task StopTaskEndpoint
	//
	// Stops a Task
	//
	// ---
	// parameters:
	// - name: ID
	//   in: path
	//   description: ID of Task to stop
	//   type: string
	//   required: false
	// responses:
	//   '200':
	//     description: Task response
	//     schema:
	//       "$ref": "#/responses/TaskResponseSwagger"
	router.HandleFunc("/api/task/{ID}/stop", endpoints.StopTaskEndpoint).Methods("POST")

	// endpoints for each retailer for create task and create task group
}
