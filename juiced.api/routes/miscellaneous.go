package routes

import (
	"backend.juicedbot.io/juiced.api/endpoints"
	"github.com/gorilla/mux"
)

// RouteCheckoutsEndpoints routes endpoints that handle miscellaneous things
func RouteMiscellaneousEndpoints(router *mux.Router) {
	router.HandleFunc("/api/settings/testWebhooks", endpoints.TestWebhooksEndpoint).Methods("POST")
	router.HandleFunc("/api/setVersion", endpoints.SetVersion).Methods("POST")
}
