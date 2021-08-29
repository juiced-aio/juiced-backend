package routes

import (
	"github.com/gorilla/mux"
)

// RouteProfilesEndpoints routes endpoints that handle profiless and profiles groups
func RouteProfilesEndpoints(router *mux.Router) {
	// router.HandleFunc("/api/profile/group", endpoints.GetAllProfileGroupsEndpoint).Methods("GET")
	// router.HandleFunc("/api/profile/group/{GroupID}", endpoints.GetProfileGroupEndpoint).Methods("GET")
	// router.HandleFunc("/api/profile/group", endpoints.CreateProfileGroupEndpoint).Methods("POST")
	// router.HandleFunc("/api/profile/group/{GroupID}", endpoints.RemoveProfileGroupEndpoint).Methods("DELETE")
	// router.HandleFunc("/api/profile/group/{GroupID}", endpoints.UpdateProfileGroupEndpoint).Methods("PUT")
	// router.HandleFunc("/api/profile/group/{GroupID}/clone", endpoints.CloneProfileGroupEndpoint).Methods("POST")
	// router.HandleFunc("/api/profile/group/{GroupID}/add", endpoints.AddProfilesToGroupEndpoint).Methods("POST")
	// router.HandleFunc("/api/profile/group/{GroupID}/remove", endpoints.RemoveProfilesFromGroupEndpoint).Methods("POST")

	// router.HandleFunc("/api/profile", endpoints.GetAllProfilesEndpoint).Methods("GET")
	// router.HandleFunc("/api/profile/{ID}", endpoints.GetProfileEndpoint).Methods("GET")
	// router.HandleFunc("/api/profile", endpoints.CreateProfileEndpoint).Methods("POST")
	// router.HandleFunc("/api/profile/{ID}", endpoints.RemoveProfileEndpoint).Methods("DELETE")
	// router.HandleFunc("/api/profile/{ID}", endpoints.UpdateProfileEndpoint).Methods("PUT")
	// router.HandleFunc("/api/profile/{ID}/clone", endpoints.CloneProfileEndpoint).Methods("POST")
	// router.HandleFunc("/api/profile/import", endpoints.ImportProfilesEndpoint).Methods("POST")
}
