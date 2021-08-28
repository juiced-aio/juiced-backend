package entities

// ProfileGroup is a class that holds a list of Profile IDs
type ProfileGroup struct {
	GroupID          string     `json:"groupID" db:"groupID"`
	Name             string     `json:"name" db:"name"`
	ProfileIDs       []string   `json:"-"`
	ProfileIDsJoined string     `json:"-" db:"profileIDsJoined"`
	CreationDate     int64      `json:"creationDate" db:"creationDate"`
	Profiles         []*Profile `json:"profiles"`
}

// AddProfileIDsToGroup adds the given ProfileIDs to the ProfileGroup
func (profileGroup *ProfileGroup) AddProfileIDsToGroup(profileIDsToAdd []string) {
	profileIDs := profileGroup.ProfileIDs
	for i := 0; i < len(profileIDsToAdd); i++ {
		alreadyInProfiles := false
		for j := 0; j < len(profileIDs); j++ {
			if profileIDs[j] == profileIDsToAdd[i] {
				alreadyInProfiles = true
			}
		}
		if !alreadyInProfiles {
			profileIDs = append(profileIDs, profileIDsToAdd[i])
		}
	}
	profileGroup.ProfileIDs = profileIDs
}

// RemoveProfileIDsFromGroup removes the given Profiles from the ProfileGroup
func (profileGroup *ProfileGroup) RemoveProfileIDsFromGroup(profileIDsToRemove []string) {
	profileIDs := []string{}
	for i := 0; i < len(profileGroup.ProfileIDs); i++ {
		inProfilesToRemove := false
		for j := 0; j < len(profileIDsToRemove); j++ {
			if profileIDsToRemove[j] == profileGroup.ProfileIDs[i] {
				inProfilesToRemove = true
			}
		}
		if !inProfilesToRemove {
			profileIDs = append(profileIDs, profileGroup.ProfileIDs[i])
		}
	}
	profileGroup.ProfileIDs = profileIDs
}
