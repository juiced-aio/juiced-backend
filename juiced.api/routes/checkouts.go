package routes

import (
	"backend.juicedbot.io/juiced.api/endpoints"

	"github.com/gorilla/mux"
)

// RouteCheckoutsEndpoints routes endpoints that handle checkouts
func RouteCheckoutsEndpoints(router *mux.Router) {
	// swagger:operation GET /api/checkouts Checkout GetAllCheckoutsEndpoint
	//
	// Returns a list of all Checkouts
	//
	// ---
	// responses:
	//   '200':
	//     description: Checkouts response
	//     schema:
	//       "$ref": "#/responses/CheckoutsResponseSwagger"
	router.HandleFunc("/api/checkouts", endpoints.GetAllCheckoutsEndpoint).Methods("GET")
}
