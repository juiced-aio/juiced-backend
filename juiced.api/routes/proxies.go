package routes

import (
	"backend.juicedbot.io/m/v2/juiced.api/endpoints"

	"github.com/gorilla/mux"
)

// RouteProxiesEndpoints routes endpoints that handle proxies and proxy groups
func RouteProxiesEndpoints(router *mux.Router) {
	// swagger:operation GET /api/proxy/group ProxyGroup GetAllProxyGroupsEndpoint
	//
	// Returns a list of all ProxyGroups
	//
	// ---
	// responses:
	//   '200':
	//     description: ProxyGroup response
	//     schema:
	//       "$ref": "#/responses/ProxyGroupResponseSwagger"
	router.HandleFunc("/api/proxy/group", endpoints.GetAllProxyGroupsEndpoint).Methods("GET")

	// swagger:operation GET /api/proxy/group/{GroupID} ProxyGroup GetProxyGroupEndpoint
	//
	// Returns the ProxyGroup with GroupID {GroupID}.
	//
	// ---
	// parameters:
	// - name: GroupID
	//   in: path
	//   description: ID of ProxyGroup to retrieve
	//   type: string
	//   required: true
	// responses:
	//   '200':
	//     description: ProxyGroup response
	//     schema:
	//       "$ref": "#/responses/ProxyGroupResponseSwagger"
	router.HandleFunc("/api/proxy/group/{GroupID}", endpoints.GetProxyGroupEndpoint).Methods("GET")

	// swagger:operation POST /api/proxy/group ProxyGroup CreateProxyGroupEndpoint
	//
	// Creates a ProxyGroup in the database
	//
	// ---
	// parameters:
	// - name: ProxyGroupDetails
	//   in: body
	//   description: ProxyGroup details
	//   required: true
	//   schema:
	//     "$ref": "#/models/CreateOrUpdateProxyGroupRequest"
	// responses:
	//   '200':
	//     description: ProxyGroup response
	//     schema:
	//       "$ref": "#/responses/ProxyGroupResponseSwagger"
	router.HandleFunc("/api/proxy/group", endpoints.CreateProxyGroupEndpoint).Methods("POST")

	// swagger:operation DELETE /api/proxy/group/{GroupID} ProxyGroup DeleteProxyGroupEndpoint
	//
	// Deletes and returns the ProxyGroup with GroupID {GroupID}.
	//
	// ---
	// parameters:
	// - name: GroupID
	//   in: path
	//   description: ID of ProxyGroup to retrieve
	//   type: string
	//   required: true
	// responses:
	//   '200':
	//     description: ProxyGroup response
	//     schema:
	//       "$ref": "#/responses/ProxyGroupResponseSwagger"
	router.HandleFunc("/api/proxy/group/{GroupID}", endpoints.RemoveProxyGroupEndpoint).Methods("DELETE")

	// swagger:operation PUT /api/proxy/group/{GroupID} ProxyGroup UpdateProxyGroupEndpoint
	//
	// Updates and returns the ProxyGroup with GroupID {GroupID}.
	//
	// ---
	// parameters:
	// - name: GroupID
	//   in: path
	//   description: ID of ProxyGroup to update
	//   type: string
	//   required: false
	// - name: ProxyGroupDetails
	//   in: body
	//   description: Details to update ProxyGroup with
	//   required: true
	//   schema:
	//     "$ref": "#/models/CreateOrUpdateProxyGroupRequest"
	// responses:
	//   '200':
	//     description: ProxyGroup response
	//     schema:
	//       "$ref": "#/responses/ProxyGroupResponseSwagger"
	router.HandleFunc("/api/proxy/group/{GroupID}", endpoints.UpdateProxyGroupEndpoint).Methods("PUT")

	// swagger:operation POST /api/proxy/group/{GroupID}/clone ProxyGroup CloneProxyGroupEndpoint
	//
	// Clones the ProxyGroup with GroupID {GroupID} and returns the clone.
	//
	// ---
	// parameters:
	// - name: GroupID
	//   in: path
	//   description: ID of ProxyGroup to clone
	//   type: string
	//   required: false
	// responses:
	//   '200':
	//     description: ProxyGroup response
	//     schema:
	//       "$ref": "#/responses/ProxyGroupResponseSwagger"
	router.HandleFunc("/api/proxy/group/{GroupID}/clone", endpoints.CloneProxyGroupEndpoint).Methods("POST")

	// proxy test
}
