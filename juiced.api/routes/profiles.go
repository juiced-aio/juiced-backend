package routes

import (
	"backend.juicedbot.io/juiced.api/endpoints"

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

	// swagger:operation POST /api/profile/group/remove RemoveProfileGroupsEndpoint
	//
	// Deletes and returns the ProfileGroups with the given GroupIDs.
	//
	// ---
	// parameters:
	// - name: GroupIDs
	//   in: body
	//   description: IDs of ProfileGroups to retrieve
	//   required: true
	// responses:
	//   '200':
	//     description: ProfileGroups response
	//     schema:
	//       "$ref": "#/responses/ProfileGroupResponseSwagger"
	router.HandleFunc("/api/profile/group/remove", endpoints.RemoveProfileGroupsEndpoint).Methods("POST")

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

	// swagger:operation POST /api/profile/group/clone CloneProfileGroupsEndpoint
	//
	// Clones the ProfileGroups with the given GroupIDs and returns the clones.
	//
	// ---
	// parameters:
	// - name: GroupIDs
	//   in: body
	//   description: IDs of ProfileGroups to clone
	//   required: false
	// responses:
	//   '200':
	//     description: ProfileGroups response
	//     schema:
	//       "$ref": "#/responses/ProfileGroupResponseSwagger"
	router.HandleFunc("/api/profile/group/clone", endpoints.CloneProfileGroupsEndpoint).Methods("POST")

	// swagger:operation POST /api/profile/group/add AddProfilesToGroupsEndpoint
	//
	// Adds Profiles to the ProfileGroups with the given GroupIDs and returns the updated ProfileGroups.
	//
	// ---
	// parameters:
	// - name: GroupIDs
	//   in: body
	//   description: IDs of ProfileGroups to add to
	//   required: true
	// - name: ProfileIDs
	//   in: body
	//   description: Profile IDs to add to the ProfileGroups
	//   required: true
	//   schema:
	//     "$ref": "#/models/ProfileIDList"
	// responses:
	//   '200':
	//     description: ProfileGroups response
	//     schema:
	//       "$ref": "#/responses/ProfileGroupResponseSwagger"
	router.HandleFunc("/api/profile/group/add", endpoints.AddProfilesToGroupsEndpoint).Methods("POST")

	// swagger:operation POST /api/profile/group/remove RemoveProfilesFromGroupsEndpoint
	//
	// Removes Profiles from the ProfileGroups with given GroupIDs and returns the updated ProfileGroups.
	//
	// ---
	// parameters:
	// - name: GroupIDs
	//   in: body
	//   description: IDs of ProfileGroups to remove from
	//   required: true
	// - name: ProfileIDs
	//   in: body
	//   description: Profile IDs to remove from the ProfileGroups
	//   required: true
	//   schema:
	//     "$ref": "#/models/ProfileIDList"
	// responses:
	//   '200':
	//     description: ProfileGroups response
	//     schema:
	//       "$ref": "#/responses/ProfileGroupResponseSwagger"
	router.HandleFunc("/api/profile/group/remove", endpoints.RemoveProfilesFromGroupsEndpoint).Methods("POST")

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

	// swagger:operation DELETE /api/profile/remove RemoveProfilesEndpoint
	//
	// Deletes and returns the Profiles with given IDs.
	//
	// ---
	// parameters:
	// - name: IDs
	//   in: body
	//   description: IDs of Profiles to remove
	//   required: true
	// responses:
	//   '200':
	//     description: Profiles response
	//     schema:
	//       "$ref": "#/responses/ProfileResponseSwagger"
	router.HandleFunc("/api/profile/remove", endpoints.RemoveProfilesEndpoint).Methods("POST")

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

	// swagger:operation POST /api/profile/clone CloneProfilesEndpoint
	//
	// Clones the Profiles with given IDs and returns the clones.
	//
	// ---
	// parameters:
	// - name: IDs
	//   in: body
	//   description: IDs of Profiles to clone
	//   required: false
	// responses:
	//   '200':
	//     description: Profiles response
	//     schema:
	//       "$ref": "#/responses/ProfileResponseSwagger"
	router.HandleFunc("/api/profile/clone", endpoints.CloneProfilesEndpoint).Methods("POST")

	router.HandleFunc("/api/profile/import", endpoints.ImportProfilesEndpoint).Methods("POST")
}
