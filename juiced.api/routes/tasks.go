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

	// swagger:operation POST /api/task/group/remove TaskGroup DeleteTaskGroupEndpoint
	//
	// Deletes and returns the TaskGroups with given GroupIDs.
	//
	// ---
	// parameters:
	// - name: GroupIDs
	//   in: body
	//   description: IDs of TaskGroups to remove
	//   type: string
	//   required: true
	// responses:
	//   '200':
	//     description: TaskGroups response
	//     schema:
	//       "$ref": "#/responses/TaskGroupResponseSwagger"
	router.HandleFunc("/api/task/group/remove", endpoints.RemoveTaskGroupsEndpoint).Methods("POST")

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

	// swagger:operation POST /api/task/group/clone TaskGroup CloneTaskGroupEndpoint
	//
	// Clones the TaskGroup with the given GroupIDs and returns the clones.
	//
	// ---
	// parameters:
	// - name: GroupID
	//   in: body
	//   description: IDs of TaskGroups to clone
	//   required: true
	// responses:
	//   '200':
	//     description: TaskGroups response
	//     schema:
	//       "$ref": "#/responses/TaskGroupResponseSwagger"
	router.HandleFunc("/api/task/group/clone", endpoints.CloneTaskGroupsEndpoint).Methods("POST")

	// swagger:operation POST /api/task/group/start TaskGroup StartTaskGroupEndpoint
	//
	// Starts given TaskGroups Monitors and all of their Tasks
	//
	// ---
	// parameters:
	// - name: GroupIDs
	//   in: body
	//   description: GroupIDs of TaskGroups to start
	//   required: false
	// responses:
	//   '200':
	//     description: TaskGroups response
	//     schema:
	//       "$ref": "#/responses/TaskGroupResponseSwagger"
	router.HandleFunc("/api/task/group/start", endpoints.StartTaskGroupsEndpoint).Methods("POST")

	// swagger:operation POST /api/task/group/stop TaskGroup StopTaskGroupsEndpoint
	//
	// Stops given TaskGroups Monitors and all of their Tasks
	//
	// ---
	// parameters:
	// - name: GroupIDs
	//   in: body
	//   description: GroupIDs of TaskGroups to stop
	//   required: false
	// responses:
	//   '200':
	//     description: TaskGroups response
	//     schema:
	//       "$ref": "#/responses/TaskGroupResponseSwagger"
	router.HandleFunc("/api/task/group/stop", endpoints.StopTaskGroupsEndpoint).Methods("POST")

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

	// swagger:operation PUT /api/task/group/{GroupID}/updateTasks Task UpdateTasksEndpoint
	//
	// Updates and returns the Tasks in group with ID {GroupID}.
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
	router.HandleFunc("/api/task/group/{GroupID}/updateTasks", endpoints.UpdateTasksEndpoint).Methods("PUT")

	// swagger:operation POST /api/task/clone Task CloneTasksEndpoint
	//
	// Clones the Tasks with the given IDs and returns the clones.
	//
	// ---
	// parameters:
	// - name: IDs
	//   in: body
	//   description: IDs of Tasks to clone
	//   required: false
	// responses:
	//   '200':
	//     description: Tasks response
	//     schema:
	//       "$ref": "#/responses/TaskResponseSwagger"
	router.HandleFunc("/api/task/clone", endpoints.CloneTasksEndpoint).Methods("POST")

	// swagger:operation POST /api/task/start Task StartTasksEndpoint
	//
	// Starts a Task
	//
	// ---
	// parameters:
	// - name: IDs
	//   in: body
	//   description: IDs of Tasks to start
	//   required: false
	// responses:
	//   '200':
	//     description: Tasks response
	//     schema:
	//       "$ref": "#/responses/TaskResponseSwagger"
	router.HandleFunc("/api/task/start", endpoints.StartTasksEndpoint).Methods("POST")

	// swagger:operation POST /api/task/stop Task StopTasksEndpoint
	//
	// Stops a Task
	//
	// ---
	// parameters:
	// - name: IDs
	//   in: body
	//   description: IDs of Tasks to stop
	//   required: false
	// responses:
	//   '200':
	//     description: Tasks response
	//     schema:
	//       "$ref": "#/responses/TaskResponseSwagger"
	router.HandleFunc("/api/task/stop", endpoints.StopTasksEndpoint).Methods("POST")

	// endpoints for each retailer for create task and create task group
}
