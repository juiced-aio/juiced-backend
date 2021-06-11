package api

import (
	"backend.juicedbot.io/juiced.api/routes"

	"net/http"

	"github.com/gorilla/mux"
	"github.com/rs/cors"
)

// StartServer launches the local server that hosts the API for communication between the app and the backend
func StartServer() {
	router := mux.NewRouter()
	sh := http.StripPrefix("/swaggerui/", http.FileServer(http.Dir("../swaggerui/")))
	router.PathPrefix("/swaggerui/").Handler(sh)
	routes.RouteProxiesEndpoints(router)
	routes.RouteProfilesEndpoints(router)
	routes.RouteTasksEndpoints(router)
	routes.RouteCheckoutsEndpoints(router)
	c := cors.New(cors.Options{
		AllowedOrigins: []string{"http://localhost:3000"},
		AllowedMethods: []string{"GET", "POST", "PUT", "PATCH", "DELETE"},
	})
	handler := c.Handler(router)
	http.ListenAndServe(":10000", handler)
}
