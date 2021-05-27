package routes

import (
	"backend.juicedbot.io/m/v2/juiced.api/endpoints"

	"github.com/gorilla/mux"
)

// RouteProfilesEndpoints routes endpoints that handle profiless and profiles groups
func RouteProfilesEndpoints(router *mux.Router) {
	// swagger:operation GET /api/profile/group ProfileGroup GetAllProfileGroupsEndpoint
	//
	// Returns a list of all ProfileGroups
	//
	// ---
	// responses:
	//   '200':
	//     description: ProfileGroup response
	//     schema:
	//       "$ref": "#/responses/ProfileGroupResponseSwagger"
	router.HandleFunc("/api/profile/group", endpoints.GetAllProfileGroupsEndpoint).Methods("GET")

	// swagger:operation GET /api/profile/group/{GroupID} ProfileGroup GetProfileGroupEndpoint
	//
	// Returns the ProfileGroup with GroupID {GroupID}.
	//
	// ---
	// parameters:
	// - name: GroupID
	//   in: path
	//   description: ID of ProfileGroup to retrieve
	//   type: string
	//   required: true
	// responses:
	//   '200':
	//     description: ProfileGroup response
	//     schema:
	//       "$ref": "#/responses/ProfileGroupResponseSwagger"
	router.HandleFunc("/api/profile/group/{GroupID}", endpoints.GetProfileGroupEndpoint).Methods("GET")

	// swagger:operation POST /api/profile/group ProfileGroup CreateProfileGroupEndpoint
	//
	// Creates a ProfileGroup in the database
	//
	// ---
	// parameters:
	// - name: ProfileGroupDetails
	//   in: body
	//   description: ProfileGroup details
	//   required: true
	//   schema:
	//     "$ref": "#/models/CreateOrUpdateProfileGroupRequest"
	// responses:
	//   '200':
	//     description: ProfileGroup response
	//     schema:
	//       "$ref": "#/responses/ProfileGroupResponseSwagger"
	router.HandleFunc("/api/profile/group", endpoints.CreateProfileGroupEndpoint).Methods("POST")

	// swagger:operation DELETE /api/profile/group/{GroupID} ProfileGroup RemoveProfileGroupEndpoint
	//
	// Deletes and returns the ProfileGroup with GroupID {GroupID}.
	//
	// ---
	// parameters:
	// - name: GroupID
	//   in: path
	//   description: ID of ProfileGroup to retrieve
	//   type: string
	//   required: true
	// responses:
	//   '200':
	//     description: ProfileGroup response
	//     schema:
	//       "$ref": "#/responses/ProfileGroupResponseSwagger"
	router.HandleFunc("/api/profile/group/{GroupID}", endpoints.RemoveProfileGroupEndpoint).Methods("DELETE")

	// swagger:operation PUT /api/profile/group/{GroupID} ProfileGroup UpdateProfileGroupEndpoint
	//
	// Updates and returns the ProfileGroup with GroupID {GroupID}.
	//
	// ---
	// parameters:
	// - name: GroupID
	//   in: path
	//   description: ID of ProfileGroup to update
	//   type: string
	//   required: false
	// - name: Name
	//   in: body
	//   description: New name to update ProfileGroup with
	//   required: true
	//   schema:
	//     "$ref": "#/models/CreateOrUpdateProfileGroupRequest"
	// responses:
	//   '200':
	//     description: ProfileGroup response
	//     schema:
	//       "$ref": "#/responses/ProfileGroupResponseSwagger"
	router.HandleFunc("/api/profile/group/{GroupID}", endpoints.UpdateProfileGroupEndpoint).Methods("PUT")

	// swagger:operation POST /api/profile/group/{GroupID}/clone ProfileGroup CloneProfileGroupEndpoint
	//
	// Clones the ProfileGroup with GroupID {GroupID} and returns the clone.
	//
	// ---
	// parameters:
	// - name: GroupID
	//   in: path
	//   description: ID of ProfileGroup to clone
	//   type: string
	//   required: false
	// responses:
	//   '200':
	//     description: ProfileGroup response
	//     schema:
	//       "$ref": "#/responses/ProfileGroupResponseSwagger"
	router.HandleFunc("/api/profile/group/{GroupID}/clone", endpoints.CloneProfileGroupEndpoint).Methods("POST")

	// swagger:operation POST /api/profile/group/{GroupID}/add ProfileGroup AddProfilesToGroupEndpoint
	//
	// Adds Profiles to the ProfileGroup with GroupID {GroupID} and returns the updated ProfileGroup.
	//
	// ---
	// parameters:
	// - name: GroupID
	//   in: path
	//   description: ID of ProfileGroup to add to
	//   type: string
	//   required: false
	// - name: ProfileIDs
	//   in: body
	//   description: Profile IDs to add to the ProfileGroup
	//   required: true
	//   schema:
	//     "$ref": "#/models/ProfileIDList"
	// responses:
	//   '200':
	//     description: ProfileGroup response
	//     schema:
	//       "$ref": "#/responses/ProfileGroupResponseSwagger"
	router.HandleFunc("/api/profile/group/{GroupID}/add", endpoints.AddProfilesToGroupEndpoint).Methods("POST")

	// swagger:operation POST /api/profile/group/{GroupID}/remove ProfileGroup RemoveProfilesFromGroupEndpoint
	//
	// Removes Profiles from the ProfileGroup with GroupID {GroupID} and returns the updated ProfileGroup.
	//
	// ---
	// parameters:
	// - name: GroupID
	//   in: path
	//   description: ID of ProfileGroup to add to
	//   type: string
	//   required: false
	// - name: ProfileIDs
	//   in: body
	//   description: Profile IDs to add to the ProfileGroup
	//   required: true
	//   schema:
	//     "$ref": "#/models/ProfileIDList"
	// responses:
	//   '200':
	//     description: ProfileGroup response
	//     schema:
	//       "$ref": "#/responses/ProfileGroupResponseSwagger"
	router.HandleFunc("/api/profile/group/{GroupID}/remove", endpoints.RemoveProfilesFromGroupEndpoint).Methods("POST")

	// swagger:operation GET /api/profile Profile GetAllProfilesEndpoint
	//
	// Returns a list of all Profiles
	//
	// ---
	// responses:
	//   '200':
	//     description: Profile response
	//     schema:
	//       "$ref": "#/responses/ProfileResponseSwagger"
	router.HandleFunc("/api/profile", endpoints.GetAllProfilesEndpoint).Methods("GET")

	// swagger:operation GET /api/profile/{ID} Profile GetProfileEndpoint
	//
	// Returns the Profile with ID {ID}.
	//
	// ---
	// parameters:
	// - name: ID
	//   in: path
	//   description: ID of Profile to retrieve
	//   type: string
	//   required: true
	// responses:
	//   '200':
	//     description: Profile response
	//     schema:
	//       "$ref": "#/responses/ProfileResponseSwagger"
	router.HandleFunc("/api/profile/{ID}", endpoints.GetProfileEndpoint).Methods("GET")

	// swagger:operation POST /api/profile Profile CreateProfileEndpoint
	//
	// Creates a Profile in the database
	//
	// ---
	// parameters:
	// - name: ProfileDetails
	//   in: body
	//   description: Profile details
	//   required: true
	//   schema:
	//     "$ref": "#/models/CreateOrUpdateProfileRequest"
	// responses:
	//   '200':
	//     description: Profile response
	//     schema:
	//       "$ref": "#/responses/ProfileResponseSwagger"
	router.HandleFunc("/api/profile", endpoints.CreateProfileEndpoint).Methods("POST")

	// swagger:operation DELETE /api/profile/{ID} Profile RemoveProfileEndpoint
	//
	// Deletes and returns the Profile with ID {ID}.
	//
	// ---
	// parameters:
	// - name: ID
	//   in: path
	//   description: ID of Profile to retrieve
	//   type: string
	//   required: true
	// responses:
	//   '200':
	//     description: Profile response
	//     schema:
	//       "$ref": "#/responses/ProfileResponseSwagger"
	router.HandleFunc("/api/profile/{ID}", endpoints.RemoveProfileEndpoint).Methods("DELETE")

	// swagger:operation PUT /api/profile/{ID} Profile UpdateProfileEndpoint
	//
	// Updates and returns the Profile with ID {ID}.
	//
	// ---
	// parameters:
	// - name: ID
	//   in: path
	//   description: ID of Profile to update
	//   type: string
	//   required: false
	// - name: ProfileDetails
	//   in: body
	//   description: Details to update Profile with
	//   required: true
	//   schema:
	//     "$ref": "#/models/CreateOrUpdateProfileRequest"
	// responses:
	//   '200':
	//     description: Profile response
	//     schema:
	//       "$ref": "#/responses/ProfileResponseSwagger"
	router.HandleFunc("/api/profile/{ID}", endpoints.UpdateProfileEndpoint).Methods("PUT")

	// swagger:operation POST /api/profile/{ID}/clone Profile CloneProfileEndpoint
	//
	// Clones the Profile with ID {ID} and returns the clone.
	//
	// ---
	// parameters:
	// - name: ID
	//   in: path
	//   description: ID of Profile to clone
	//   type: string
	//   required: false
	// responses:
	//   '200':
	//     description: Profile response
	//     schema:
	//       "$ref": "#/responses/ProfileResponseSwagger"
	router.HandleFunc("/api/profile/{ID}/clone", endpoints.CloneProfileEndpoint).Methods("POST")
}
