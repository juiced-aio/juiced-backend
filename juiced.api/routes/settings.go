package routes

import (
	"backend.juicedbot.io/juiced.api/endpoints"

	"github.com/gorilla/mux"
)

// RouteSettingsEndpoints routes endpoints that handle settings
func RouteSettingsEndpoints(router *mux.Router) {
	// swagger:operation GET /api/settings Settings GetSettingsEndpoint
	//
	// Returns the user's settings.
	//
	// ---
	// responses:
	//   '200':
	//     description: Settings response
	//     schema:
	//       "$ref": "#/responses/SettingsResponseSwagger"
	router.HandleFunc("/api/settings", endpoints.GetSettingsEndpoint).Methods("GET")

	// swagger:operation PUT /api/settings Settings UpdateSettingsEndpoint
	//
	// Updates and returns the user's settings.
	//
	// ---
	// parameters:
	// - name: SettingsDetails
	//   in: body
	//   description: Details to update settings with
	//   required: true
	//   schema:
	//     "$ref": "#/models/UpdateSettingsRequest"
	// responses:
	//   '200':
	//     description: Settings response
	//     schema:
	//       "$ref": "#/responses/SettingsResponseSwagger"
	router.HandleFunc("/api/settings", endpoints.UpdateSettingsEndpoint).Methods("PUT")
}
