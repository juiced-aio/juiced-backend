package routes

import (
	"github.com/gorilla/mux"
)

// RouteTasksEndpoints routes endpoints that handle tasks and task groups
func RouteTasksEndpoints(router *mux.Router) {
	// router.HandleFunc("/api/task/group", endpoints.GetAllTaskGroupsEndpoint).Methods("GET")
	// router.HandleFunc("/api/task/group/{GroupID}", endpoints.GetTaskGroupEndpoint).Methods("GET")
	// router.HandleFunc("/api/task/group", endpoints.CreateTaskGroupEndpoint).Methods("POST")
	// router.HandleFunc("/api/task/group/{GroupID}", endpoints.RemoveTaskGroupEndpoint).Methods("DELETE")
	// router.HandleFunc("/api/task/group/{GroupID}", endpoints.UpdateTaskGroupEndpoint).Methods("PUT")
	// router.HandleFunc("/api/task/group/{GroupID}/clone", endpoints.CloneTaskGroupEndpoint).Methods("POST")
	// router.HandleFunc("/api/task/group/{GroupID}/start", endpoints.StartTaskGroupEndpoint).Methods("POST")
	// router.HandleFunc("/api/task/group/{GroupID}/stop", endpoints.StopTaskGroupEndpoint).Methods("POST")
	// router.HandleFunc("/api/task/group/{GroupID}/removeTasks", endpoints.RemoveTasksEndpoint).Methods("POST")

	// router.HandleFunc("/api/task", endpoints.GetAllTasksEndpoint).Methods("GET")
	// router.HandleFunc("/api/task/{ID}", endpoints.GetTaskEndpoint).Methods("GET")
	// router.HandleFunc("/api/task/{GroupID}", endpoints.CreateTaskEndpoint).Methods("POST")
	// router.HandleFunc("/api/task/group/{GroupID}/updateTasks", endpoints.UpdateTasksEndpoint).Methods("PUT")
	// router.HandleFunc("/api/task/{ID}/clone", endpoints.CloneTaskEndpoint).Methods("POST")
	// router.HandleFunc("/api/task/{ID}/start", endpoints.StartTaskEndpoint).Methods("POST")
	// router.HandleFunc("/api/task/{ID}/stop", endpoints.StopTaskEndpoint).Methods("POST")
}
